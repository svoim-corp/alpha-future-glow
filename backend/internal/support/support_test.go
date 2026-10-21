package support

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func ptr[T any](v T) *T { return &v }

// users — токен сессии в тестах равен логину.
var users = map[string]*auth.User{
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1")},
	"petrova": {ID: "u-2", Role: shared.RoleClient, CompanyID: ptr("c-1")},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3")},
	"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer},
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

// fakeApps — заявки по id: только компания важна для проверки доступа.
type fakeApps map[string]string

func (f fakeApps) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	if company, ok := f[id]; ok {
		return &shared.ApplicationView{ID: id, CompanyID: company}, nil
	}
	return nil, nil
}

func (fakeApps) Transition(context.Context, string, shared.Status, shared.Actor, string, string) error {
	return nil
}

func (fakeApps) SetAmount(context.Context, string, int64, shared.Actor) (int, error) { return 0, nil }

func (fakeApps) BKIConsent(context.Context, string, shared.ConsentSubject) (*shared.ConsentView, error) {
	return nil, nil
}

func init() { gin.SetMode(gin.TestMode) }

type env struct {
	m      *Module
	router *gin.Engine
}

// newEnv — модуль на тестовой БД. Без TEST_DATABASE_URL тест пропускается:
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
	m := New(module.Deps{DB: conn})
	m.Wire(Ports{Applications: fakeApps{"a-1": "c-1", "a-2": "c-3"}})
	if err := m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.ResetDemo(ctx); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(httpx.TraceID(), auth.Middleware(resolver{}))
	api := r.Group("/api/v1")
	m.Register(httpx.Routes{Public: api, Client: api.Group("", auth.RequireAuth()), Bank: api.Group("/bank")})
	return &env{m: m, router: r}
}

func (e *env) do(method, path, body string) *httptest.ResponseRecorder {
	return e.as("", method, path, body)
}

func (e *env) as(user, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

func TestFAQ(t *testing.T) {
	e := newEnv(t)
	got := decode[[]faqDTO](t, e.do("GET", "/api/v1/support/faq?step=APPLICATION", ""))
	if len(got) == 0 || got[0].ID != "f-4" {
		t.Fatalf("вопросы по заявке: %+v", got)
	}
	for _, f := range got {
		if f.Step != "APPLICATION" || f.Question == "" || f.Answer == "" {
			t.Fatalf("вопрос другого шага или пустой: %+v", f)
		}
	}
	all := decode[[]faqDTO](t, e.do("GET", "/api/v1/support/faq", ""))
	if len(all) <= len(got) {
		t.Fatalf("без шага — все вопросы: %d", len(all))
	}
	if w := e.do("GET", "/api/v1/support/faq?step=NOPE", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("неизвестный шаг: %d", w.Code)
	}
}

func TestCreateTicket(t *testing.T) {
	e := newEnv(t)
	w := e.as("smirnov", "POST", "/api/v1/support/tickets", `{"subject":"Устаревший адрес","text":"Мы переехали","applicationId":"a-1"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if got := decode[ticketCreatedDTO](t, w); !strings.HasPrefix(got.ID, "t-") || got.Status != "OPEN" || got.CreatedAt.IsZero() {
		t.Fatalf("обращение: %+v", got)
	}
	if w := e.as("smirnov", "POST", "/api/v1/support/tickets", `{"subject":"Вопрос","text":"Без заявки"}`); w.Code != http.StatusCreated {
		t.Fatalf("без заявки: %d", w.Code)
	}

	tests := []struct {
		name, user, body string
		status           int
	}{
		{"чужая заявка", "smirnov", `{"subject":"x","text":"y","applicationId":"a-2"}`, http.StatusBadRequest},
		{"нет заявки", "smirnov", `{"subject":"x","text":"y","applicationId":"a-9"}`, http.StatusBadRequest},
		{"без темы", "smirnov", `{"text":"y"}`, http.StatusBadRequest},
		{"сотрудник банка", "ivanova", `{"subject":"x","text":"y"}`, http.StatusForbidden},
		{"без входа", "", `{"subject":"x","text":"y"}`, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := e.as(tt.user, "POST", "/api/v1/support/tickets", tt.body); w.Code != tt.status {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
}
