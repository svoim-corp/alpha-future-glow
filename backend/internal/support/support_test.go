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

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

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
	if err := m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(httpx.TraceID())
	api := r.Group("/api/v1")
	m.Register(httpx.Routes{Public: api, Client: api, Bank: api.Group("/bank")})
	return &env{m: m, router: r}
}

func (e *env) do(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
