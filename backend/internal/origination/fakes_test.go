package origination

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

type fakeProducts map[string]shared.ProductView

func (f fakeProducts) Product(_ context.Context, code string) (*shared.ProductView, error) {
	if p, ok := f[code]; ok {
		return &p, nil
	}
	return nil, nil
}

var express = shared.ProductView{
	Code: "EXPRESS", Name: "Экспресс-кредит", Goals: []string{"WORKING_CAPITAL", "EQUIPMENT"},
	AmountMin: 100_000, AmountMax: 1_000_000, TermMin: 3, TermMax: 24, RateMin: 24, RateMax: 30,
}

// fakeAuthorities — полномочия по ИНН человека; нет записи — полномочий нет.
type fakeAuthorities map[string]shared.AuthorityView

func (f fakeAuthorities) Resolve(_ context.Context, _, personINN string) (shared.AuthorityView, error) {
	if a, ok := f[personINN]; ok {
		return a, nil
	}
	return shared.AuthorityView{Type: shared.AuthorityNone, Reason: "Срок действия доверенности истёк 30.09.2026"}, nil
}

// fakeCompanies — профили компаний; upstream — данные банка недоступны.
type fakeCompanies struct {
	mu       sync.Mutex
	profiles map[string]*shared.CompanyProfile
	upstream bool
}

func (f *fakeCompanies) GetByID(_ context.Context, id string) (*shared.CompanyView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p, ok := f.profiles[id]; ok {
		cv := p.CompanyView
		return &cv, nil
	}
	return nil, nil
}

func (f *fakeCompanies) Profile(_ context.Context, id string) (*shared.CompanyProfile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.upstream {
		return nil, httpx.Upstream("")
	}
	return f.profiles[id], nil
}

var smirnovPerson = shared.Person{PersonID: "p-1", FullName: "Смирнов Олег Викторович", INN: "771234567890", Roles: []shared.PersonRole{shared.PersonHead, shared.PersonFounder}}

func newFakeCompanies() *fakeCompanies {
	return &fakeCompanies{profiles: map[string]*shared.CompanyProfile{
		"c-1": {
			CompanyView: shared.CompanyView{ID: "c-1", INN: "7712345678", Name: "ООО «Северная пекарня»", Type: shared.CompanyLLC},
			OGRN:        "1157746000000", State: "ACTIVE", RegDate: time.Date(2015, 3, 12, 0, 0, 0, 0, time.UTC),
			ActualAddress: "г. Москва, ул. Складская, 3", Head: smirnovPerson,
			GuarantorCandidates: []shared.Person{smirnovPerson, {PersonID: "p-3", FullName: "Кравцов Илья Андреевич", Roles: []shared.PersonRole{shared.PersonFounder}}},
			Turnover:            shared.Turnover{AvgMonthlyIncome: 1_840_000, PeriodMonths: 12},
		},
		"c-3": {
			CompanyView: shared.CompanyView{ID: "c-3", INN: "772512345601", Name: "ИП Орлова Мария Игоревна", Type: shared.CompanyIE},
			State:       "ACTIVE", ActualAddress: "г. Москва, ул. Садовая, 1", GuarantorCandidates: []shared.Person{},
		},
	}}
}

// fakePrelims — результаты этапа 1 по id заявки.
type fakePrelims struct {
	mu sync.Mutex
	m  map[string]*shared.PrelimView
}

func (f *fakePrelims) Prelim(_ context.Context, appID string) (*shared.PrelimView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.m[appID], nil
}

// passPrelim — этап 1 пройден с лимитом для текущих параметров заявки.
func (e *env) passPrelim(t *testing.T, appID string) {
	t.Helper()
	a, err := e.m.store.Application(context.Background(), appID)
	if err != nil || a == nil {
		t.Fatalf("заявка %s: %v", appID, err)
	}
	e.prelims.mu.Lock()
	defer e.prelims.mu.Unlock()
	e.prelims.m[appID] = &shared.PrelimView{Outcome: shared.PrelimLimit, Limit: 3_300_000, PrelimRate: 22, AmountFits: true,
		ProductCode: a.ProductCode, Amount: a.Amount, Term: a.Term}
}

// users — токен сессии в тестах равен логину.
var users = map[string]*auth.User{
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("771234567890"), FullName: "Смирнов Олег Викторович"},
	"petrova": {ID: "u-2", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("771298765432"), FullName: "Петрова Анна Сергеевна"},
	"lebedev": {ID: "u-90", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("770000000001"), FullName: "Лебедев Пётр Ильич"}, // МЧД без BKI_CONSENT
	"volkov":  {ID: "u-3", Role: shared.RoleClient, CompanyID: ptr("c-2"), PersonINN: ptr("770411223344"), FullName: "Волков Игорь Николаевич"},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3"), PersonINN: ptr("772512345601"), FullName: "Орлова Мария Игоревна"},
	"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer, FullName: "Иванова Елена Петровна"},
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

type env struct {
	m         *Module
	bus       *events.Bus
	companies *fakeCompanies
	prelims   *fakePrelims
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

	bus := events.New()
	companies, prelims := newFakeCompanies(), &fakePrelims{m: map[string]*shared.PrelimView{}}
	if err := idempotency.Migrate(ctx, conn); err != nil {
		t.Fatal(err)
	}
	m := New(module.Deps{DB: conn, Bus: bus, Idem: idempotency.New(conn)})
	m.now = func() time.Time { return today }
	m.Wire(Ports{
		Products:  fakeProducts{"WORKING": working, "EXPRESS": express},
		Companies: companies,
		Prelims:   prelims,
		Authorities: fakeAuthorities{
			"771234567890": {Type: shared.AuthorityHead, Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent}},
			"771298765432": {Type: shared.AuthorityMCHD, MCHDNumber: ptr("a1b2"), ValidUntil: ptr(time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)),
				Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent}},
			"770000000001": {Type: shared.AuthorityMCHD, MCHDNumber: ptr("c3d4"), ValidUntil: ptr(time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)),
				Powers: []string{shared.PowerCreditApplication}},
			"772512345601": {Type: shared.AuthorityHead, Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent}},
		},
	})
	if err := m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.ResetDemo(ctx); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.Use(httpx.TraceID(), auth.Middleware(resolver{}))
	api := r.Group("/api/v1")
	m.Register(httpx.Routes{
		Public: api,
		Client: api.Group("", auth.RequireAuth()),
		Bank:   api.Group("/bank", auth.RequireAuth(), auth.RequireRole(shared.RoleBankOfficer)),
	})
	return &env{m: m, bus: bus, companies: companies, prelims: prelims, router: r}
}

// do выполняет запрос от имени user; hdr — дополнительные заголовки.
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

// decodeBody — разбор без *testing.T для хелперов, которые сами возвращают результат проверки.
func decodeBody[T any](b []byte) T {
	var v T
	_ = json.Unmarshal(b, &v)
	return v
}

// draft создаёт черновик WORKING на 1 500 000 ₽ и 24 мес. от имени user.
func (e *env) draft(t *testing.T, user string) createResponse {
	t.Helper()
	w := e.do("POST", "/api/v1/applications", user, `{"productCode":"WORKING","amount":1500000,"term":24}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("черновик %s: %d %s", user, w.Code, w.Body.String())
	}
	return decode[createResponse](t, w)
}

// ready — черновик ООО, готовый к отправке: этап 1, согласия и поручитель. Возвращает id и версию.
func (e *env) ready(t *testing.T, user string) (string, int) {
	t.Helper()
	id := e.draft(t, user).ID
	e.passPrelim(t, id)
	if w := e.do("POST", "/api/v1/applications/"+id+"/consents/bki", user, consentBody); w.Code != http.StatusCreated {
		t.Fatalf("согласие: %s", w.Body.String())
	}
	if st, _, fail := e.patch(id, user, "1", `{"stage2":{"guarantor":{"personId":"p-1"}}}`); st != http.StatusOK {
		t.Fatalf("поручитель: %d %+v", st, fail)
	}
	if w := e.do("POST", "/api/v1/applications/"+id+"/guarantor/consent-bki", user, `{"given":true}`); w.Code != http.StatusCreated {
		t.Fatalf("согласие поручителя: %s", w.Body.String())
	}
	return id, 2
}
