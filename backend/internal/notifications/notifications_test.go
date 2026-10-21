package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

var now = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

// users — токен сессии в тестах равен id пользователя.
type users struct{}

func (users) Resolve(_ context.Context, token string) (*auth.User, error) {
	return &auth.User{ID: token, Role: shared.RoleClient}, nil
}

// newTestModule — модуль на тестовой БД с пустой таблицей уведомлений. Без TEST_DATABASE_URL тест пропускается:
// TEST_DATABASE_URL='postgres://cabinet:cabinet@localhost:5432/cabinet?sslmode=disable' go test ./...
func newTestModule(t *testing.T) (*Module, *gin.Engine) {
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
	m.now = func() time.Time { return now }
	if err := m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.ResetDemo(ctx); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(httpx.TraceID(), auth.Middleware(users{}))
	api := r.Group("/api/v1")
	m.Register(httpx.Routes{Public: api, Client: api.Group("", auth.RequireAuth()), Bank: api.Group("/bank")})
	return m, r
}

func do(r http.Handler, method, path, user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(""))
	if user != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: user})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestNotify(t *testing.T) {
	m, _ := newTestModule(t)
	ctx := context.Background()
	n := shared.Notification{Type: shared.NotifyDecision, Title: "Решение по заявке", Text: "Одобрено", Link: "/applications/a-1"}
	if err := m.Notifier().Notify(ctx, "u-1", n); err != nil {
		t.Fatal(err)
	}
	var got notificationRow
	if err := m.deps.DB.GetContext(ctx, &got, `SELECT id, user_id, type, title, text, link, read, created_at
		FROM notifications.notifications`); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got.ID, "n-") || got.UserID != "u-1" || got.Type != "DECISION" || got.Read || !got.CreatedAt.Equal(now) {
		t.Fatalf("уведомление: %+v", got)
	}
}
