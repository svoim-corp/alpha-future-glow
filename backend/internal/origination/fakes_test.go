package origination

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// users — токен сессии в тестах равен логину.
var users = map[string]*auth.User{
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("771234567890"), FullName: "Смирнов Олег Викторович"},
	"petrova": {ID: "u-2", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("771298765432"), FullName: "Петрова Анна Сергеевна"},
	"volkov":  {ID: "u-3", Role: shared.RoleClient, CompanyID: ptr("c-2"), PersonINN: ptr("770411223344"), FullName: "Волков Игорь Николаевич"},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3"), PersonINN: ptr("772512345601"), FullName: "Орлова Мария Игоревна"},
	"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer, FullName: "Иванова Елена Петровна"},
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

type env struct {
	m      *Module
	bus    *events.Bus
	router *gin.Engine
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
	m := New(module.Deps{DB: conn, Bus: bus})
	m.now = func() time.Time { return today }
	m.Wire(Ports{
		Products: fakeProducts{"WORKING": working, "EXPRESS": express},
		Authorities: fakeAuthorities{
			"771234567890": {Type: shared.AuthorityHead, Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent}},
			"771298765432": {Type: shared.AuthorityMCHD, MCHDNumber: ptr("a1b2"), ValidUntil: ptr(time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)),
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
	return &env{m: m, bus: bus, router: r}
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

// draft создаёт черновик WORKING на 1 500 000 ₽ и 24 мес. от имени user.
func (e *env) draft(t *testing.T, user string) createResponse {
	t.Helper()
	w := e.do("POST", "/api/v1/applications", user, `{"productCode":"WORKING","amount":1500000,"term":24}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("черновик %s: %d %s", user, w.Code, w.Body.String())
	}
	return decode[createResponse](t, w)
}
