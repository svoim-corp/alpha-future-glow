package catalog

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

// newTestModule — модуль на тестовой БД с seed-линейкой. Без TEST_DATABASE_URL тест пропускается:
// TEST_DATABASE_URL='postgres://cabinet:cabinet@localhost:5432/cabinet?sslmode=disable' go test ./...
func newTestModule(t *testing.T) (*Module, *gin.Engine) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	conn, err := db.Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	m := New(module.Deps{DB: conn})
	if err := m.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(httpx.TraceID())
	api := r.Group("/api/v1")
	m.Register(httpx.Routes{Public: api, Client: api, Bank: api.Group("/bank")})
	return m, r
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
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

func TestListProducts(t *testing.T) {
	_, r := newTestModule(t)
	w := do(r, "GET", "/api/v1/products", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	got := decode[[]productSummaryDTO](t, w)
	codes := make([]string, len(got))
	for i, p := range got {
		codes[i] = p.Code
	}
	if strings.Join(codes, ",") != "WORKING,INVEST,EXPRESS,OVERDRAFT" {
		t.Fatalf("порядок линейки: %v", codes)
	}
	if w0 := got[0]; w0.Amount.Min != 300_000 || w0.Term.Max != 36 || w0.Rate.Min != 20 || len(w0.Goals) == 0 {
		t.Fatalf("рамки WORKING: %+v", w0)
	}
}
