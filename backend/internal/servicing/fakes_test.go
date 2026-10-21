package servicing

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
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func init() { gin.SetMode(gin.TestMode) }

// today — «сейчас» в тестах; кредиты выдаются в этот день.
var today = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }

type fakeProducts map[string]shared.ProductView

func (f fakeProducts) Product(_ context.Context, code string) (*shared.ProductView, error) {
	if p, ok := f[code]; ok {
		return &p, nil
	}
	return nil, nil
}

type fakeNotifier struct {
	mu   sync.Mutex
	sent []string // "userID:TYPE:title"
}

func (f *fakeNotifier) Notify(_ context.Context, userID string, n shared.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, userID+":"+string(n.Type)+":"+n.Title)
	return nil
}

var users = map[string]*auth.User{
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1")},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3")},
	"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer},
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

type env struct {
	m        *Module
	notifier *fakeNotifier
	router   *gin.Engine
}

// newEnv — модуль на тестовой БД без кредитов. Без TEST_DATABASE_URL тест пропускается:
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
	e := &env{notifier: &fakeNotifier{}}
	e.m = New(module.Deps{DB: conn})
	e.m.now = func() time.Time { return today }
	e.m.Wire(Ports{Notifier: e.notifier, Products: fakeProducts{"WORKING": {Code: "WORKING", Name: "Оборотный кредит"}}})
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

// open выдаёт кредит компании company по заявке appID на 1 200 000 ₽, 24 мес., 21,5%.
func (e *env) open(t *testing.T, appID, company string, issued time.Time) string {
	t.Helper()
	id, err := e.m.Loans().Open(context.Background(), shared.LoanOpening{
		AppID: appID, CompanyID: company, ApplicantID: "u-1", ContractNumber: "КД-2026-" + appID, ProductCode: "WORKING",
		Amount: 1_200_000, Term: 24, Rate: 21.5, IssuedAt: issued,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (e *env) do(method, path, user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(""))
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
