package offer

import (
	"context"
	"encoding/json"
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
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/idempotency"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func init() { gin.SetMode(gin.TestMode) }

var today = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }

// fakeApps — заявки Origination с переходами docs/origination.md.
type fakeApps struct {
	mu      sync.Mutex
	apps    map[string]*shared.ApplicationView
	history []string // "id:TO:action:comment"
}

var transitions = map[shared.Status][]shared.Status{
	shared.StatusOnDecision:   {shared.StatusApproved, shared.StatusCounterOffer, shared.StatusRejected},
	shared.StatusApproved:     {shared.StatusSigning, shared.StatusClosed},
	shared.StatusCounterOffer: {shared.StatusSigning, shared.StatusClosed},
	shared.StatusSigning:      {shared.StatusIssued},
}

func (f *fakeApps) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if a, ok := f.apps[id]; ok {
		v := *a
		return &v, nil
	}
	return nil, nil
}

func (f *fakeApps) Transition(_ context.Context, id string, to shared.Status, _ shared.Actor, action, comment string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.apps[id]
	if !ok {
		return httpx.NotFound("")
	}
	if !slices.Contains(transitions[a.Status], to) {
		return httpx.InvalidTransition("")
	}
	a.Status = to
	f.history = append(f.history, id+":"+string(to)+":"+action+":"+comment)
	return nil
}

func (f *fakeApps) SetAmount(context.Context, string, int64, shared.Actor) (int, error) {
	return 0, nil
}

func (f *fakeApps) BKIConsent(context.Context, string, shared.ConsentSubject) (*shared.ConsentView, error) {
	return nil, nil
}

func (f *fakeApps) status(id string) shared.Status {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.apps[id].Status
}

type fakeDecisions map[string]*shared.DecisionView

func (f fakeDecisions) LatestDecision(_ context.Context, appID string) (*shared.DecisionView, error) {
	return f[appID], nil
}

type fakeCompanies map[string]*shared.CompanyProfile

func (f fakeCompanies) GetByID(_ context.Context, id string) (*shared.CompanyView, error) {
	if p, ok := f[id]; ok {
		cv := p.CompanyView
		return &cv, nil
	}
	return nil, nil
}

func (f fakeCompanies) Profile(_ context.Context, id string) (*shared.CompanyProfile, error) {
	return f[id], nil
}

var users = map[string]*auth.User{
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1"), FullName: "Смирнов Олег Викторович"},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3"), FullName: "Орлова Мария Игоревна"},
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

type env struct {
	m         *Module
	bus       *events.Bus
	apps      *fakeApps
	decisions fakeDecisions
	router    *gin.Engine
}

// newEnv — модуль на тестовой БД с фейками других зон. Без TEST_DATABASE_URL тест пропускается:
// TEST_DATABASE_URL='postgres://cabinet:cabinet@localhost:5432/cabinet?sslmode=disable' go test ./...
func newEnv(t *testing.T) *env {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := idempotency.Migrate(ctx, conn); err != nil {
		t.Fatal(err)
	}

	e := &env{bus: events.New(), apps: &fakeApps{apps: map[string]*shared.ApplicationView{}}, decisions: fakeDecisions{}}
	e.m = New(module.Deps{DB: conn, Bus: e.bus, Idem: idempotency.New(conn)})
	e.m.now = func() time.Time { return today }
	e.m.Wire(Ports{
		Applications: e.apps,
		Decisions:    e.decisions,
		Companies: fakeCompanies{
			"c-1": {CompanyView: shared.CompanyView{ID: "c-1", INN: "7712345678", Name: "ООО «Северная пекарня»", Type: shared.CompanyLLC},
				GuarantorCandidates: []shared.Person{{PersonID: "p-1", FullName: "Смирнов Олег Викторович"}}},
			"c-3": {CompanyView: shared.CompanyView{ID: "c-3", INN: "772512345601", Name: "ИП Орлова Мария Игоревна", Type: shared.CompanyIE}},
		},
	})
	if err := e.m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := e.m.ResetDemo(ctx); err != nil {
		t.Fatal(err)
	}
	e.router = gin.New()
	e.router.Use(httpx.TraceID(), auth.Middleware(resolver{}))
	api := e.router.Group("/api/v1")
	e.m.Register(httpx.Routes{Public: api, Client: api.Group("", auth.RequireAuth()), Bank: api.Group("/bank")})
	return e
}

// decide — заявка компании company получила решение d: как если бы его принял Scoring и опубликовал DecisionMade.
func (e *env) decide(t *testing.T, id, company string, d shared.DecisionView) {
	t.Helper()
	app := &shared.ApplicationView{
		ID: id, CompanyID: company, ApplicantID: "u-1", ProductCode: "WORKING", Amount: 1_500_000, Term: 24,
		Status: d.Result, Number: ptr("КЗ-2026-000001"),
	}
	if company == "c-1" {
		app.GuarantorPersonID = ptr("p-1")
	}
	e.apps.mu.Lock()
	e.apps.apps[id] = app
	e.apps.mu.Unlock()
	if d.ID == "" {
		d.ID = "dec-" + id
	}
	if d.DecidedAt.IsZero() {
		d.DecidedAt = today
	}
	e.decisions[id] = &d
	if err := e.bus.Publish(context.Background(), shared.DecisionMade{AppID: id, DecisionID: d.ID}); err != nil {
		t.Fatal(err)
	}
}

var approved = shared.DecisionView{Result: shared.StatusApproved, Mode: shared.DecisionAuto, ProductCode: "WORKING",
	Amount: 1_500_000, Term: 24, Rate: 21.5, Reason: "Автоматическое одобрение: балл 740"}

func (e *env) do(method, path, user, body string, hdr ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	if user != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: user})
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
	return v
}
