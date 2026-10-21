package underwriting

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func init() { gin.SetMode(gin.TestMode) }

var today = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

// fakeApps — заявки Origination с переходами из API_SPEC.md, раздел 2. Каждый переход публикует ApplicationStatusChanged,
// как настоящая реализация.
type fakeApps struct {
	mu       sync.Mutex
	bus      *events.Bus
	apps     map[string]*shared.ApplicationView
	consents map[string]*shared.ConsentView // ключ — appID + "/" + subject
	history  []string                       // "id:FROM->TO:action"
}

var transitions = map[shared.Status][]shared.Status{
	shared.StatusDraft:         {shared.StatusSubmitted, shared.StatusClosed},
	shared.StatusSubmitted:     {shared.StatusDataCheck},
	shared.StatusDataCheck:     {shared.StatusScoring, shared.StatusInfoRequested},
	shared.StatusInfoRequested: {shared.StatusDataCheck},
	shared.StatusScoring:       {shared.StatusOnDecision},
	shared.StatusOnDecision:    {shared.StatusApproved, shared.StatusCounterOffer, shared.StatusRejected, shared.StatusInfoRequested},
}

func newFakeApps(bus *events.Bus) *fakeApps {
	return &fakeApps{bus: bus, apps: map[string]*shared.ApplicationView{}, consents: map[string]*shared.ConsentView{}}
}

func (f *fakeApps) add(a *shared.ApplicationView) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apps[a.ID] = a
}

func (f *fakeApps) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if a, ok := f.apps[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, nil
}

func (f *fakeApps) Transition(ctx context.Context, id string, to shared.Status, _ shared.Actor, action, _ string) error {
	f.mu.Lock()
	a, ok := f.apps[id]
	if !ok {
		f.mu.Unlock()
		return httpx.NotFound("")
	}
	from := a.Status
	if !slices.Contains(transitions[from], to) {
		f.mu.Unlock()
		return httpx.InvalidTransition("")
	}
	a.Status = to
	f.history = append(f.history, id+":"+string(from)+"->"+string(to)+":"+action)
	f.mu.Unlock()
	return f.bus.Publish(ctx, shared.ApplicationStatusChanged{AppID: id, From: from, To: to, At: today})
}

func (f *fakeApps) SetAmount(context.Context, string, int64, shared.Actor) (int, error) {
	return 0, errors.New("not used")
}

func (f *fakeApps) BKIConsent(_ context.Context, appID string, s shared.ConsentSubject) (*shared.ConsentView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.consents[appID+"/"+string(s)], nil
}

func (f *fakeApps) statusOf(id string) shared.Status {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.apps[id].Status
}

type fakeCompanies map[string]*shared.CompanyProfile

func (f fakeCompanies) GetByID(_ context.Context, id string) (*shared.CompanyView, error) {
	if p, ok := f[id]; ok {
		return &p.CompanyView, nil
	}
	return nil, nil
}

func (f fakeCompanies) Profile(_ context.Context, id string) (*shared.CompanyProfile, error) {
	return f[id], nil
}

type fakeProducts map[string]*shared.ProductView

func (f fakeProducts) Product(_ context.Context, code string) (*shared.ProductView, error) {
	return f[code], nil
}

type fakePrelims map[string]*shared.PrelimView

func (f fakePrelims) Prelim(_ context.Context, appID string) (*shared.PrelimView, error) {
	return f[appID], nil
}

// fakeChecks — снимки проверок по заявке; Run запоминает запрос и отмечает перезапрошенные источники.
type fakeChecks struct {
	mu      sync.Mutex
	reports map[string]*shared.CheckReport
	last    *shared.CheckRequest
}

func (f *fakeChecks) Run(_ context.Context, req shared.CheckRequest) (*shared.CheckReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.last = &req
	rep := f.reports[req.AppID]
	if rep == nil {
		rep = &shared.CheckReport{}
		f.reports[req.AppID] = rep
	}
	for _, s := range req.Sources {
		rep.Results = append(rep.Results, shared.CheckResult{Source: s, Status: shared.CheckOK, RequestedAt: today, Summary: "перезапрошено"})
	}
	return rep, nil
}

func (f *fakeChecks) Latest(_ context.Context, appID string) (*shared.CheckReport, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.reports[appID]; ok {
		return r, nil
	}
	return &shared.CheckReport{}, nil
}

type fakeScores map[string]*shared.ScoreView

func (f fakeScores) Score(_ context.Context, appID string) (*shared.ScoreView, error) {
	return f[appID], nil
}

var (
	working = &shared.ProductView{Code: "WORKING", Name: "Оборотный кредит", AmountMin: 300_000, AmountMax: 10_000_000,
		TermMin: 6, TermMax: 36, RateMin: 19, RateMax: 26, GuarantorForLLC: true}
	express = &shared.ProductView{Code: "EXPRESS", Name: "Экспресс-кредит", AmountMin: 100_000, AmountMax: 3_000_000,
		TermMin: 3, TermMax: 24, RateMin: 21, RateMax: 29}
	bakery = &shared.CompanyProfile{CompanyView: shared.CompanyView{ID: "c-1", INN: "7712345678", Name: "ООО «Северная пекарня»", Type: shared.CompanyLLC}}
	trade  = &shared.CompanyProfile{CompanyView: shared.CompanyView{ID: "c-8", INN: "7707789050", Name: "ООО «Альфа-Трейд»", Type: shared.CompanyLLC}}
)

func submittedApp(id, company, product string, amount int64, at time.Time) *shared.ApplicationView {
	n := "КЗ-2026-" + id
	return &shared.ApplicationView{ID: id, Number: &n, CompanyID: company, ApplicantID: "u-1", ProductCode: product,
		Amount: amount, Term: 24, Status: shared.StatusDraft, SubmittedAt: &at}
}

// env — модуль с фейками других зон. С БД — если задан TEST_DATABASE_URL.
type env struct {
	m       *Module
	bus     *events.Bus
	apps    *fakeApps
	scores  fakeScores
	prelims fakePrelims
	checks  *fakeChecks
	router  *gin.Engine
}

func newEnv(t *testing.T, withDB bool) *env {
	t.Helper()
	deps := module.Deps{Bus: events.New()}
	if withDB {
		url := os.Getenv("TEST_DATABASE_URL")
		if url == "" {
			t.Skip("TEST_DATABASE_URL не задан")
		}
		conn, err := db.Open(context.Background(), url)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		deps.DB = conn
	}
	e := &env{bus: deps.Bus, apps: newFakeApps(deps.Bus), scores: fakeScores{}, prelims: fakePrelims{},
		checks: &fakeChecks{reports: map[string]*shared.CheckReport{}}}
	e.m = New(deps)
	e.m.now = func() time.Time { return today }
	if withDB {
		if err := e.m.Migrate(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := e.m.ResetDemo(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	e.m.Wire(Ports{
		Applications: e.apps, Scores: e.scores, Prelims: e.prelims, Checks: e.checks,
		Companies: fakeCompanies{"c-1": bakery, "c-8": trade},
		Products:  fakeProducts{"WORKING": working, "EXPRESS": express},
	})

	e.router = gin.New()
	e.router.Use(httpx.TraceID(), auth.Middleware(users))
	e.m.Register(httpx.Routes{
		Client: e.router.Group("/api/v1", auth.RequireAuth()),
		Bank:   e.router.Group("/api/v1/bank", auth.RequireAuth(), auth.RequireRole(shared.RoleBankOfficer)),
	})
	return e
}

var (
	companyC1 = "c-1"
	users     = fakeUsers{
		"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer, FullName: "Иванова Елена Петровна"},
		"sidorov": {ID: "u-14", Role: shared.RoleBankOfficer, FullName: "Сидоров Максим Андреевич"},
		"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: &companyC1, FullName: "Смирнов Олег Викторович"},
	}
)

type fakeUsers map[string]*auth.User

func (f fakeUsers) Resolve(_ context.Context, token string) (*auth.User, error) { return f[token], nil }

// do выполняет запрос от имени пользователя (логин — токен сессии).
func (e *env) do(method, path, login, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if login != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: login})
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

// move проводит заявку по статусам через фейковый Origination (с событиями).
func (e *env) move(t *testing.T, id string, to ...shared.Status) {
	t.Helper()
	for _, s := range to {
		if err := e.apps.Transition(context.Background(), id, s, shared.Actor{Type: shared.ActorSystem}, "TEST", ""); err != nil {
			t.Fatalf("%s → %s: %v", id, s, err)
		}
	}
}
