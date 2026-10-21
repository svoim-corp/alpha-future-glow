package scoring

import (
	"context"
	"encoding/json"
	"errors"
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

type fakeApps struct {
	apps     map[string]*shared.ApplicationView
	consents map[string]*shared.ConsentView
	setCalls int
}

func (f *fakeApps) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	if a, ok := f.apps[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, nil
}

func (f *fakeApps) Transition(context.Context, string, shared.Status, shared.Actor, string, string) error {
	return errors.New("not used")
}

func (f *fakeApps) SetAmount(_ context.Context, id string, amount int64, _ shared.Actor) (int, error) {
	f.setCalls++
	a := f.apps[id]
	a.Amount, a.Version = amount, a.Version+1
	return a.Version, nil
}

func (f *fakeApps) BKIConsent(_ context.Context, appID string, _ shared.ConsentSubject) (*shared.ConsentView, error) {
	return f.consents[appID], nil
}

type fakeProducts map[string]*shared.ProductView

func (f fakeProducts) Product(_ context.Context, code string) (*shared.ProductView, error) {
	return f[code], nil
}

type fakeCompanies struct{}

func (fakeCompanies) GetByID(_ context.Context, id string) (*shared.CompanyView, error) {
	return &shared.CompanyView{ID: id, INN: "inn-" + id, Type: shared.CompanyLLC}, nil
}

func (fakeCompanies) Profile(context.Context, string) (*shared.CompanyProfile, error) {
	return nil, nil
}

// fakeChecks отдаёт отчёт по ИНН и запоминает последний запрос.
type fakeChecks struct {
	reports map[string]*shared.CheckReport
	last    shared.CheckRequest
}

func (f *fakeChecks) Run(_ context.Context, req shared.CheckRequest) (*shared.CheckReport, error) {
	f.last = req
	return f.reports[req.INN], nil
}

func (f *fakeChecks) Latest(context.Context, string) (*shared.CheckReport, error) { return nil, nil }

var allPowers = shared.AuthorityView{Type: shared.AuthorityHead, Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent}}

func draftApp(id, company string, amount int64) *shared.ApplicationView {
	return &shared.ApplicationView{ID: id, CompanyID: company, ProductCode: "WORKING", Amount: amount, Term: 24,
		Status: shared.StatusDraft, Authority: allPowers, Version: 3}
}

// fixture: a-1 — c-1 с согласием (оборот 1,84 млн), a-2 — c-7 (оборот 520 тыс., сумма выше лимита),
// a-3 — без согласия, a-4 — подана, a-5 — представитель без полномочий, a-6 — продукт не найден, a-7 — источник не ответил.
func newTestModule(conn *module.Deps) (*Module, *fakeApps, *fakeChecks) {
	apps := &fakeApps{
		apps: map[string]*shared.ApplicationView{
			"a-1": draftApp("a-1", "c-1", 1_500_000),
			"a-2": draftApp("a-2", "c-7", 1_500_000),
			"a-3": draftApp("a-3", "c-1", 1_500_000),
			"a-4": func() *shared.ApplicationView {
				a := draftApp("a-4", "c-1", 1_500_000)
				a.Status = shared.StatusSubmitted
				return a
			}(),
			"a-5": func() *shared.ApplicationView {
				a := draftApp("a-5", "c-1", 1_500_000)
				a.Authority = shared.AuthorityView{}
				return a
			}(),
			"a-6": func() *shared.ApplicationView {
				a := draftApp("a-6", "c-1", 1_500_000)
				a.ProductCode = "NOPE"
				return a
			}(),
			"a-7": draftApp("a-7", "c-9", 1_500_000),
		},
		consents: map[string]*shared.ConsentView{},
	}
	for _, id := range []string{"a-1", "a-2", "a-4", "a-5", "a-6", "a-7"} {
		apps.consents[id] = &shared.ConsentView{ID: "cons-" + id}
	}
	unavailable := facts(nil)
	unavailable.Results = []shared.CheckResult{{Source: shared.SourceBKI, Status: shared.CheckUnavailable}}
	checks := &fakeChecks{reports: map[string]*shared.CheckReport{
		"inn-c-1": facts(nil),
		"inn-c-7": facts(func(f *shared.CheckFacts) { f.Turnover.AvgMonthlyIncome = 520_000 }),
		"inn-c-9": unavailable,
	}}
	deps := module.Deps{}
	if conn != nil {
		deps = *conn
	}
	m := New(deps)
	m.now = func() time.Time { return today }
	m.Wire(Ports{Applications: apps, Products: fakeProducts{"WORKING": &working}, Companies: fakeCompanies{}, Checks: checks})
	return m, apps, checks
}

func errCode(err error) string {
	var e *httpx.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// TestRunPrelimRejects — отказы до сохранения результата: БД не нужна.
func TestRunPrelimRejects(t *testing.T) {
	m, _, _ := newTestModule(nil)
	tests := []struct {
		name, app, company, code string
	}{
		{"нет заявки", "a-404", "c-1", httpx.CodeNotFound},
		{"чужая компания", "a-1", "c-2", httpx.CodeForbidden},
		{"сотрудник банка без компании", "a-1", "", httpx.CodeForbidden},
		{"заявка уже подана", "a-4", "c-1", httpx.CodeInvalidTransition},
		{"нет полномочий", "a-5", "c-1", httpx.CodeForbidden},
		{"нет согласия на БКИ", "a-3", "c-1", httpx.CodeForbidden},
		{"чужая заявка: доступ проверяется раньше продукта", "a-6", "c-9", httpx.CodeForbidden},
		{"продукт не найден у своей заявки", "a-6", "c-1", httpx.CodeBusinessRule},
		{"источник не ответил", "a-7", "c-9", httpx.CodeUpstreamUnavail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := m.runPrelim(context.Background(), tt.app, tt.company)
			if errCode(err) != tt.code {
				t.Fatalf("err = %v, want %s", err, tt.code)
			}
		})
	}
}

func TestPrelimChoiceValidation(t *testing.T) {
	m, _, _ := newTestModule(nil)
	r := gin.New()
	r.Use(httpx.TraceID(), auth.Middleware(resolver{}))
	m.Register(httpx.Routes{Client: r.Group("", auth.RequireAuth())})

	tests := []struct {
		body  string
		field string
	}{
		{`{}`, "choice"},
		{`{"choice":"MAYBE"}`, "choice"},
		{`{"choice":"REDUCE_AMOUNT"}`, "amount"},
	}
	for _, tt := range tests {
		w := post(r, "/applications/a-1/prelim/choice", tt.body)
		var b httpx.ErrorBody
		_ = json.Unmarshal(w.Body.Bytes(), &b)
		if w.Code != http.StatusBadRequest || len(b.Fields) != 1 || b.Fields[0].Field != tt.field {
			t.Errorf("%s: %d %+v", tt.body, w.Code, b)
		}
	}
}

type resolver struct{}

func (resolver) Resolve(context.Context, string) (*auth.User, error) {
	c := "c-1"
	return &auth.User{ID: "u-1", Role: shared.RoleClient, CompanyID: &c, FullName: "Смирнов Олег Викторович"}, nil
}

func post(r http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "t"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// testDeps — БД для тестов со стором. Без TEST_DATABASE_URL тест пропускается:
// TEST_DATABASE_URL='postgres://cabinet:cabinet@localhost:5432/cabinet?sslmode=disable' go test ./...
func testDeps(t *testing.T) *module.Deps {
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
	deps := module.Deps{DB: conn}
	if err := New(deps).Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, `TRUNCATE scoring.prelims`); err != nil {
		t.Fatal(err)
	}
	return &deps
}

func TestPrelimFlow(t *testing.T) {
	m, apps, checks := newTestModule(testDeps(t))
	ctx := context.Background()
	client := shared.Actor{Type: shared.ActorClient, UserID: "u-1"}

	// Сумма в лимите: выбор не нужен.
	p, _, err := m.runPrelim(ctx, "a-1", "c-1")
	if err != nil || p.Outcome != "LIMIT" || !p.AmountFits || p.LimitAmount != 3_300_000 || len(p.Checks.V) != 0 {
		t.Fatalf("a-1: %+v %v", p, err)
	}
	if checks.last.ConsentID != "cons-a-1" || len(checks.last.Sources) != len(prelimSources) {
		t.Fatalf("запрос проверок: %+v", checks.last)
	}
	if _, _, _, err := m.choosePrelim(ctx, "a-1", "c-1", shared.ChoiceKeepAmount, 0, client); errCode(err) != httpx.CodeInvalidTransition {
		t.Fatalf("выбор при сумме в лимите: %v", err)
	}
	v, err := m.Prelims().Prelim(ctx, "a-1")
	if err != nil || v.Outcome != shared.PrelimLimit || v.Amount != 1_500_000 || v.ProductCode != "WORKING" || v.Choice != nil {
		t.Fatalf("Prelims: %+v %v", v, err)
	}

	// Сумма выше лимита (sokolova): сначала выбор без пройденного этапа, затем все варианты.
	if _, _, _, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceKeepAmount, 0, client); errCode(err) != httpx.CodeInvalidTransition {
		t.Fatalf("выбор без этапа 1: %v", err)
	}
	p, _, err = m.runPrelim(ctx, "a-2", "c-7")
	if err != nil || p.AmountFits || p.LimitAmount != 900_000 {
		t.Fatalf("a-2: %+v %v", p, err)
	}
	if _, _, _, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceReduceAmount, 1_000_000, client); errCode(err) != httpx.CodeBusinessRule {
		t.Fatalf("сумма выше лимита: %v", err)
	}
	if _, _, _, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceReduceAmount, 100_000, client); errCode(err) != httpx.CodeBusinessRule {
		t.Fatalf("сумма ниже минимума продукта: %v", err)
	}
	if _, amount, version, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceKeepAmount, 0, client); err != nil || amount != 1_500_000 || version != 3 || apps.setCalls != 0 {
		t.Fatalf("KEEP_AMOUNT: %d %d %v", amount, version, err)
	}
	app, amount, version, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceReduceAmount, 900_000, client)
	if err != nil || amount != 900_000 || version != 4 || app.Term != 24 || apps.setCalls != 1 {
		t.Fatalf("REDUCE_AMOUNT: %d %d %v", amount, version, err)
	}
	v, _ = m.Prelims().Prelim(ctx, "a-2")
	if v.Choice == nil || *v.Choice != shared.ChoiceReduceAmount || v.Amount != 900_000 || !v.AmountFits {
		t.Fatalf("после уменьшения результат должен совпадать с заявкой: %+v", v)
	}
	if _, _, _, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceKeepAmount, 0, client); errCode(err) != httpx.CodeInvalidTransition {
		t.Fatalf("повторный выбор после уменьшения: %v", err)
	}

	// Клиент поменял срок после этапа 1 — выбор по устаревшему результату запрещён.
	if _, _, err := m.runPrelim(ctx, "a-2", "c-7"); err != nil {
		t.Fatal(err)
	}
	apps.apps["a-2"].Amount, apps.apps["a-2"].Term = 1_500_000, 24
	if _, _, err := m.runPrelim(ctx, "a-2", "c-7"); err != nil {
		t.Fatal(err)
	}
	apps.apps["a-2"].Term = 36
	if _, _, _, err := m.choosePrelim(ctx, "a-2", "c-7", shared.ChoiceKeepAmount, 0, client); errCode(err) != httpx.CodeInvalidTransition {
		t.Fatalf("изменённые параметры: %v", err)
	}

	// Стоп сохраняется и виден через Prelims.
	checks.reports["inn-c-1"] = facts(func(f *shared.CheckFacts) { f.BKI.MaxOverdueDays = 95 })
	p, res, err := m.runPrelim(ctx, "a-1", "c-1")
	if err != nil || p.StopReason == nil || *p.StopReason != stopBKIOverdue {
		t.Fatalf("стоп: %+v %v", p, err)
	}
	if resp := toPrelimResponse(p, res); resp.Limit != nil || resp.StopText == "" || len(resp.WhatToDo) == 0 {
		t.Fatalf("ответ стопа: %+v", resp)
	}
	if v, _ := m.Prelims().Prelim(ctx, "a-1"); v.Outcome != shared.PrelimStop {
		t.Fatalf("Prelims после стопа: %+v", v)
	}
}
