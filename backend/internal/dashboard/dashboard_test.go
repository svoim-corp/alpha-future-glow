package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
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

// owners — компания объекта по id; общий фейк для кредитов, планов и заявок. Под мьютексом: тест меняет его.
type owners struct {
	mu sync.Mutex
	m  map[string]string
}

func (o *owners) company(id string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.m[id]
}

func (o *owners) LoanCompany(_ context.Context, id string) (string, error) { return o.company(id), nil }
func (o *owners) PlanCompany(_ context.Context, id string) (string, error) { return o.company(id), nil }

func (o *owners) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	if c := o.company(id); c != "" {
		return &shared.ApplicationView{ID: id, CompanyID: c}, nil
	}
	return nil, nil
}

func (*owners) Transition(context.Context, string, shared.Status, shared.Actor, string, string) error {
	return nil
}

func (*owners) SetAmount(context.Context, string, int64, shared.Actor) (int, error) { return 0, nil }

func (*owners) BKIConsent(context.Context, string, shared.ConsentSubject) (*shared.ConsentView, error) {
	return nil, nil
}

func init() { gin.SetMode(gin.TestMode) }

type env struct {
	m      *Module
	router *gin.Engine
	owners *owners
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
	o := &owners{m: map[string]string{"l-1": "c-1", "l-2": "c-3", "a-1": "c-1", "a-2": "c-3", "p-1": "c-1", "p-2": "c-3"}}
	m := New(module.Deps{DB: conn})
	m.Wire(Ports{Applications: o, Loans: o, Plans: o})
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
	return &env{m: m, router: r, owners: o}
}

// req — запрос от пользователя с заголовками вида "If-Match", `"rev"`.
func (e *env) req(user, method, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1/me/dashboard", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
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

func errCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	return decode[httpx.ErrorBody](t, w).Code
}

const (
	loanWidget = `{"schemaVersion":1,"widgets":[{"id":"w1","type":"loan","x":0,"y":0,"w":4,"h":3,"settings":{"loanId":"l-1"}}]}`
	emptyBoard = `{"schemaVersion":1,"widgets":[]}`
)

func TestCreateAndRead(t *testing.T) {
	e := newEnv(t)
	if w := e.req("smirnov", "GET", ""); w.Code != http.StatusNoContent || w.Header().Get("Cache-Control") != cacheControl {
		t.Fatalf("до сохранения: %d %v", w.Code, w.Header())
	}
	if w := e.req("smirnov", "PUT", loanWidget); w.Code != http.StatusPreconditionRequired || errCode(t, w) != httpx.CodePreconditionReq {
		t.Fatalf("без проверки версии: %d %s", w.Code, w.Body.String())
	}

	w := e.req("smirnov", "PUT", loanWidget, "If-None-Match", "*")
	if w.Code != http.StatusOK {
		t.Fatalf("первое сохранение: %d %s", w.Code, w.Body.String())
	}
	saved := decode[documentDTO](t, w)
	if saved.Revision == "" || w.Header().Get("ETag") != `"`+saved.Revision+`"` || saved.UpdatedAt.IsZero() ||
		len(saved.Config.Widgets) != 1 || saved.Config.Widgets[0].Type != "loan" {
		t.Fatalf("документ: %+v, ETag %s", saved, w.Header().Get("ETag"))
	}

	w = e.req("smirnov", "GET", "")
	if got := decode[documentDTO](t, w); w.Code != http.StatusOK || got.Revision != saved.Revision ||
		w.Header().Get("ETag") != `"`+saved.Revision+`"` || string(got.Config.Widgets[0].Settings["loanId"]) != `"l-1"` {
		t.Fatalf("чтение: %d %s", w.Code, w.Body.String())
	}

	// Пустой список — тоже сохранённая настройка, а не её отсутствие.
	w = e.req("smirnov", "PUT", emptyBoard, "If-Match", `"`+saved.Revision+`"`)
	if w.Code != http.StatusOK {
		t.Fatalf("сохранение пустой: %d %s", w.Code, w.Body.String())
	}
	w = e.req("smirnov", "GET", "")
	if got := decode[documentDTO](t, w); w.Code != http.StatusOK || got.Config.Widgets == nil || len(got.Config.Widgets) != 0 {
		t.Fatalf("пустая настройка: %d %s", w.Code, w.Body.String())
	}
}

func TestVersionConflict(t *testing.T) {
	e := newEnv(t)
	first := decode[documentDTO](t, e.req("smirnov", "PUT", emptyBoard, "If-None-Match", "*"))

	tests := []struct {
		name    string
		headers []string
		status  int
		code    string
	}{
		{"запись уже создана", []string{"If-None-Match", "*"}, http.StatusPreconditionFailed, httpx.CodeDashboardConflict},
		{"чужая ревизия", []string{"If-Match", `"8efc6500-8373-4e9d-9f0a-a5d9efae1175"`}, http.StatusPreconditionFailed, httpx.CodeDashboardConflict},
		{"оба заголовка", []string{"If-Match", `"` + first.Revision + `"`, "If-None-Match", "*"}, http.StatusBadRequest, httpx.CodeValidation},
		{"If-None-Match не звёздочка", []string{"If-None-Match", `"x"`}, http.StatusBadRequest, httpx.CodeValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := e.req("smirnov", "PUT", loanWidget, tt.headers...)
			if w.Code != tt.status || errCode(t, w) != tt.code {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}

	second := e.req("smirnov", "PUT", loanWidget, "If-Match", `"`+first.Revision+`"`)
	if second.Code != http.StatusOK {
		t.Fatalf("обновление: %d %s", second.Code, second.Body.String())
	}
	// Второе окно со старой ревизией не перезаписывает новую настройку.
	if w := e.req("smirnov", "PUT", emptyBoard, "If-Match", `"`+first.Revision+`"`); w.Code != http.StatusPreconditionFailed {
		t.Fatalf("устаревшая ревизия: %d %s", w.Code, w.Body.String())
	}
	if got := decode[documentDTO](t, e.req("smirnov", "GET", "")); len(got.Config.Widgets) != 1 {
		t.Fatalf("настройка перезаписана: %+v", got)
	}
}

func TestIdempotentSave(t *testing.T) {
	e := newEnv(t)
	key := []string{"If-None-Match", "*", "Idempotency-Key", "k-1"}
	first := e.req("smirnov", "PUT", loanWidget, key...)
	if first.Code != http.StatusOK {
		t.Fatalf("сохранение: %d %s", first.Code, first.Body.String())
	}
	// Ответ потерян, клиент повторяет ту же команду: запись уже есть, но результат — исходный, а не 412.
	again := e.req("smirnov", "PUT", loanWidget, key...)
	if again.Code != http.StatusOK || again.Body.String() != first.Body.String() || again.Header().Get("ETag") != first.Header().Get("ETag") {
		t.Fatalf("повтор: %d %s", again.Code, again.Body.String())
	}
	if w := e.req("smirnov", "PUT", emptyBoard, key...); w.Code != http.StatusConflict || errCode(t, w) != httpx.CodeIdempotencyReused {
		t.Fatalf("тот же ключ с другим документом: %d %s", w.Code, w.Body.String())
	}
	// Ключи изолированы по пользователю: коллега с тем же ключом сохраняет своё.
	if w := e.req("petrova", "PUT", emptyBoard, key...); w.Code != http.StatusOK {
		t.Fatalf("ключ коллеги: %d %s", w.Code, w.Body.String())
	}
}

func TestSources(t *testing.T) {
	e := newEnv(t)
	tests := []struct {
		name, body string
		status     int
	}{
		{"свой кредит", loanWidget, http.StatusOK},
		{"чужой кредит", `{"schemaVersion":1,"widgets":[{"id":"w","type":"payment","x":0,"y":0,"w":4,"h":2,"settings":{"loanId":"l-2"}}]}`, http.StatusBadRequest},
		{"нет кредита", `{"schemaVersion":1,"widgets":[{"id":"w","type":"payments","x":0,"y":0,"w":8,"h":3,"settings":{"days":7,"loanId":"l-9"}}]}`, http.StatusBadRequest},
		{"своя заявка", `{"schemaVersion":1,"widgets":[{"id":"w","type":"application","x":0,"y":0,"w":4,"h":3,"settings":{"applicationId":"a-1"}}]}`, http.StatusOK},
		{"чужая заявка", `{"schemaVersion":1,"widgets":[{"id":"w","type":"application","x":0,"y":0,"w":4,"h":3,"settings":{"applicationId":"a-2"}}]}`, http.StatusBadRequest},
		{"свой план", `{"schemaVersion":1,"widgets":[{"id":"w","type":"planFunding","x":0,"y":0,"w":6,"h":4,"settings":{"planId":"p-1"}}]}`, http.StatusOK},
		{"чужой план", `{"schemaVersion":1,"widgets":[{"id":"w","type":"planFunding","x":0,"y":0,"w":6,"h":4,"settings":{"planId":"p-2"}}]}`, http.StatusBadRequest},
		{"источник не выбран", `{"schemaVersion":1,"widgets":[{"id":"w","type":"planFunding","x":0,"y":0,"w":6,"h":4,"settings":{"planId":null}}]}`, http.StatusOK},
	}
	messages := map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := e.m.ResetDemo(context.Background()); err != nil {
				t.Fatal(err)
			}
			w := e.req("smirnov", "PUT", tt.body, "If-None-Match", "*")
			if w.Code != tt.status {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if w.Code == http.StatusBadRequest {
				messages[tt.name] = decode[httpx.ErrorBody](t, w).Fields[0].Message
			}
		})
	}
	if messages["чужой кредит"] == "" || messages["чужой кредит"] != messages["нет кредита"] {
		t.Fatalf("чужой и отсутствующий кредит различимы: %v", messages)
	}
}

// Потеря доступа к уже сохранённому договору не блокирует остальные изменения и не подменяет привязку.
func TestKeptSource(t *testing.T) {
	e := newEnv(t)
	saved := decode[documentDTO](t, e.req("smirnov", "PUT", loanWidget, "If-None-Match", "*"))
	e.owners.mu.Lock()
	delete(e.owners.m, "l-1")
	e.owners.mu.Unlock()

	moved := strings.Replace(loanWidget, `"y":0`, `"y":3`, 1)
	w := e.req("smirnov", "PUT", moved, "If-Match", `"`+saved.Revision+`"`)
	if w.Code != http.StatusOK {
		t.Fatalf("перемещение виджета с недоступным кредитом: %d %s", w.Code, w.Body.String())
	}
	if got := decode[documentDTO](t, w); string(got.Config.Widgets[0].Settings["loanId"]) != `"l-1"` || got.Config.Widgets[0].Y != 3 {
		t.Fatalf("привязка: %+v", got.Config.Widgets[0])
	}
}

func TestAccess(t *testing.T) {
	e := newEnv(t)
	if w := e.req("smirnov", "PUT", loanWidget, "If-None-Match", "*"); w.Code != http.StatusOK {
		t.Fatalf("сохранение: %d %s", w.Code, w.Body.String())
	}
	// Настройка коллеги из той же компании — своя, не общая.
	if w := e.req("petrova", "GET", ""); w.Code != http.StatusNoContent {
		t.Fatalf("коллега видит чужую настройку: %d %s", w.Code, w.Body.String())
	}
	tests := []struct {
		name, user, method string
		status             int
	}{
		{"сотрудник банка читает", "ivanova", "GET", http.StatusForbidden},
		{"сотрудник банка сохраняет", "ivanova", "PUT", http.StatusForbidden},
		{"без входа", "", "GET", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := e.req(tt.user, tt.method, emptyBoard, "If-None-Match", "*"); w.Code != tt.status {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestUnsupportedSchema(t *testing.T) {
	e := newEnv(t)
	w := e.req("smirnov", "PUT", `{"schemaVersion":2,"widgets":[]}`, "If-None-Match", "*")
	if w.Code != http.StatusBadRequest || errCode(t, w) != httpx.CodeDashboardSchema {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}

// Параллельные повторы одной команды: запись одна, все получают её результат.
func TestConcurrentRetries(t *testing.T) {
	e := newEnv(t)
	const n = 8
	codes, etags := make([]int, n), make([]string, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := e.req("smirnov", "PUT", loanWidget, "If-None-Match", "*", "Idempotency-Key", "k-race")
			codes[i], etags[i] = w.Code, w.Header().Get("ETag")
		}()
	}
	wg.Wait()
	for i := range n {
		if codes[i] != http.StatusOK || etags[i] != etags[0] {
			t.Fatalf("повтор %d: %d %s, первый %s", i, codes[i], etags[i], etags[0])
		}
	}
}
