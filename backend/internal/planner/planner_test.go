package planner

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
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeMatcher — линейка Catalog: продукт подходит, если у него нужная цель и бюджет в его рамках.
type fakeMatcher struct {
	mu    sync.Mutex
	calls int
}

var catalog = []shared.ProductView{
	{Code: "INVEST", Name: "Инвестиционный кредит", Goals: []string{"EQUIPMENT", "REAL_ESTATE"}, AmountMin: 1_000_000, AmountMax: 50_000_000},
	{Code: "EXPRESS", Name: "Экспресс-кредит", Goals: []string{"EQUIPMENT", "WORKING_CAPITAL"}, AmountMin: 100_000, AmountMax: 3_000_000},
}

func (f *fakeMatcher) Match(_ context.Context, goal string, amount int64) ([]shared.ProductView, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	var out []shared.ProductView
	for _, p := range catalog {
		if slices.Contains(p.Goals, goal) && amount >= p.AmountMin && amount <= p.AmountMax {
			out = append(out, p)
		}
	}
	return out, nil
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

type fakeUsers map[string]*auth.User

func (f fakeUsers) Resolve(_ context.Context, token string) (*auth.User, error) { return f[token], nil }

var (
	c1, c8 = "c-1", "c-8"
	users  = fakeUsers{
		"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: &c1},
		"morozov": {ID: "u-10", Role: shared.RoleClient, CompanyID: &c8},
		"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer},
	}
	now = time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
)

type env struct {
	m        *Module
	matcher  *fakeMatcher
	notifier *fakeNotifier
	router   *gin.Engine
}

func newEnv(t *testing.T, withDB bool) *env {
	t.Helper()
	deps := module.Deps{}
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
	e := &env{m: New(deps), matcher: &fakeMatcher{}, notifier: &fakeNotifier{}}
	e.m.now = func() time.Time { return now }
	if withDB {
		if err := e.m.Migrate(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := e.m.ResetDemo(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	e.m.Wire(Ports{Matcher: e.matcher, Notifier: e.notifier})
	e.router = gin.New()
	e.router.Use(httpx.TraceID(), auth.Middleware(users))
	e.m.Register(httpx.Routes{Client: e.router.Group("", auth.RequireAuth())})
	return e
}

func (e *env) do(method, path, login, body string, out any) int {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: login})
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	if out != nil {
		_ = json.Unmarshal(w.Body.Bytes(), out)
	}
	return w.Code
}

// TestValidation — ошибки, которые отсекаются до БД.
func TestValidation(t *testing.T) {
	e := newEnv(t, false)
	tests := []struct {
		body  string
		field string
	}{
		{`{"goal":"EQUIPMENT","budgetRub":1,"dateMode":"UNSCHEDULED"}`, "title"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":0,"dateMode":"UNSCHEDULED"}`, "budgetRub"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":-5,"dateMode":"UNSCHEDULED"}`, "budgetRub"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":1,"dateMode":"SOMEDAY"}`, "dateMode"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":1,"dateMode":"DAY","startDate":"01.06.2027"}`, "startDate"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":1,"dateMode":"DAY"}`, "startDate"},
		{`{"title":"Печь","goal":"EQUIPMENT","budgetRub":1,"dateMode":"PERIOD","startDate":"2027-06-01","endDate":"2027-05-01"}`, "endDate"},
	}
	for _, tt := range tests {
		var b httpx.ErrorBody
		if code := e.do("POST", "/plans", "smirnov", tt.body, &b); code != http.StatusBadRequest || len(b.Fields) != 1 || b.Fields[0].Field != tt.field {
			t.Errorf("%s: %d %+v", tt.body, code, b)
		}
	}
	if code := e.do("GET", "/plans?state=LOST", "smirnov", "", nil); code != http.StatusBadRequest {
		t.Errorf("state: %d", code)
	}
	if code := e.do("GET", "/plans", "ivanova", "", nil); code != http.StatusForbidden {
		t.Errorf("сотрудник банка: %d", code)
	}
}

func TestPlans(t *testing.T) {
	e := newEnv(t, true)
	clock := now
	e.m.now = func() time.Time { return clock }
	create := func(login, body string) planDTO {
		t.Helper()
		clock = clock.Add(time.Minute) // у каждого плана своё время создания: список — новые первыми
		var p planDTO
		if code := e.do("POST", "/plans", login, body, &p); code != http.StatusCreated {
			t.Fatalf("create %s: %d", body, code)
		}
		return p
	}

	oven := create("smirnov", `{"title":"Новая печь","goal":"EQUIPMENT","budgetRub":1200000,"dateMode":"UNSCHEDULED","startDate":"2027-01-01"}`)
	if oven.StartDate != nil || oven.Recommendation == nil || oven.Recommendation.ProductCode != "INVEST" ||
		oven.Recommendation.AvailableFrom != "2026-10-21" || !oven.Notifications.InApp || oven.State != stateActive || !strings.HasPrefix(oven.ID, "p-") {
		t.Fatalf("печь: %+v", oven)
	}
	repair := create("smirnov", `{"title":"Ремонт","goal":"REAL_ESTATE","budgetRub":2400000,"dateMode":"PERIOD","startDate":"2027-06-01","endDate":"2027-08-31"}`)
	if repair.Recommendation == nil || repair.Recommendation.AvailableFrom != "2027-06-01" {
		t.Fatalf("ремонт: %+v", repair.Recommendation)
	}
	day := create("smirnov", `{"title":"Выставка","note":"Стенд","goal":"MARKETING","budgetRub":300000,"dateMode":"DAY","startDate":"2026-11-15","endDate":"2026-12-01","notifications":{"inApp":false,"email":true}}`)
	if day.EndDate == nil || *day.EndDate != "2026-11-15" || day.Recommendation != nil || !day.Notifications.Email {
		t.Fatalf("выставка: %+v", day)
	}
	quiet := create("smirnov", `{"title":"Склад","goal":"EQUIPMENT","budgetRub":500000,"dateMode":"UNSCHEDULED","notifications":{"inApp":false}}`)
	if quiet.Recommendation == nil || quiet.Recommendation.ProductCode != "EXPRESS" {
		t.Fatalf("склад: %+v", quiet.Recommendation)
	}
	create("morozov", `{"title":"Чужой","goal":"EQUIPMENT","budgetRub":1200000,"dateMode":"UNSCHEDULED"}`)

	// PLAN_MATCH: только у начавшегося подбора, только при inApp и один раз на вариант.
	wantSent := []string{"u-1:PLAN_MATCH:Нашли финансирование для плана «Новая печь»", "u-10:PLAN_MATCH:Нашли финансирование для плана «Чужой»"}
	var list plansResponse
	if code := e.do("GET", "/plans", "smirnov", "", &list); code != http.StatusOK {
		t.Fatalf("list: %d", code)
	}
	if !slices.Equal(e.notifier.sent, wantSent) {
		t.Fatalf("уведомления: %v", e.notifier.sent)
	}
	if len(list.Items) != 4 || list.Items[0].ID != quiet.ID || list.Summary != (summaryDTO{Total: 4, WithPeriod: 2, Unscheduled: 2, BudgetRub: 4_400_000}) {
		t.Fatalf("список: %+v", list)
	}

	var recs []planRecommendationDTO
	e.do("GET", "/plans/recommendations", "smirnov", "", &recs)
	if len(recs) != 2 || recs[0].PlanID != quiet.ID || recs[1].PlanID != oven.ID {
		t.Fatalf("рекомендации: %+v", recs)
	}

	// PATCH: пауза снимает рекомендацию, пустая строка очищает описание, смена режима — даты.
	var p planDTO
	if code := e.do("PATCH", "/plans/"+day.ID, "smirnov", `{"note":"","dateMode":"UNSCHEDULED","state":"PAUSED"}`, &p); code != http.StatusOK ||
		p.Note != nil || p.StartDate != nil || p.EndDate != nil || p.State != "PAUSED" || p.Title != "Выставка" {
		t.Fatalf("patch: %d %+v", code, p)
	}
	if code := e.do("PATCH", "/plans/"+repair.ID, "smirnov", `{"dateMode":"DAY"}`, &p); code != http.StatusOK || *p.EndDate != "2027-06-01" {
		t.Fatalf("patch DAY: %d %+v", code, p)
	}
	if code := e.do("PATCH", "/plans/"+repair.ID, "smirnov", `{"dateMode":"PERIOD","endDate":"2027-01-01"}`, nil); code != http.StatusBadRequest {
		t.Fatalf("patch конец раньше начала: %d", code)
	}
	if code := e.do("PATCH", "/plans/"+oven.ID, "smirnov", `{"title":""}`, nil); code != http.StatusBadRequest {
		t.Fatalf("patch пустое название: %d", code)
	}
	if code := e.do("PATCH", "/plans/"+oven.ID, "smirnov", `{"state":"ARCHIVED"}`, nil); code != http.StatusOK {
		t.Fatalf("архив: %d", code)
	}
	list = plansResponse{}
	e.do("GET", "/plans", "smirnov", "", &list)
	var archived plansResponse
	e.do("GET", "/plans?state=ARCHIVED", "smirnov", "", &archived)
	if len(list.Items) != 3 || len(archived.Items) != 1 || archived.Items[0].ID != oven.ID || archived.Items[0].Recommendation != nil {
		t.Fatalf("архив скрыт из списка: %d %+v", len(list.Items), archived)
	}

	// Доступ и удаление.
	for _, tt := range []struct {
		method, login string
		code          int
	}{
		{"PATCH", "morozov", http.StatusForbidden}, {"DELETE", "morozov", http.StatusForbidden},
		{"DELETE", "smirnov", http.StatusNoContent}, {"DELETE", "smirnov", http.StatusNotFound},
	} {
		if code := e.do(tt.method, "/plans/"+quiet.ID, tt.login, `{"title":"x"}`, nil); code != tt.code {
			t.Errorf("%s %s: %d, want %d", tt.login, tt.method, code, tt.code)
		}
	}
	if !slices.Equal(e.notifier.sent, wantSent) {
		t.Fatalf("повторных уведомлений быть не должно: %v", e.notifier.sent)
	}
}
