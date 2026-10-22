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
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/idempotency"
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
	"smirnov": {ID: "u-1", Role: shared.RoleClient, CompanyID: ptr("c-1"), PersonINN: ptr("771234567890")},
	"orlova":  {ID: "u-4", Role: shared.RoleClient, CompanyID: ptr("c-3"), PersonINN: ptr("772512345601")},
	"guest":   {ID: "u-7", Role: shared.RoleClient, CompanyID: ptr("c-1")}, // в АБС не связан с компанией
	"ivanova": {ID: "u-9", Role: shared.RoleBankOfficer},
}

// fakeCompanies — ИНН компании c-N — «inn-c-N».
type fakeCompanies struct{}

func (fakeCompanies) GetByID(_ context.Context, id string) (*shared.CompanyView, error) {
	return &shared.CompanyView{ID: id, INN: "inn-" + id}, nil
}

func (fakeCompanies) Profile(context.Context, string) (*shared.CompanyProfile, error) {
	return nil, nil
}

// fakeAuthorities — без ИНН человека полномочий нет, с ИНН — руководитель.
type fakeAuthorities struct{}

func (fakeAuthorities) Resolve(_ context.Context, _, personINN string) (shared.AuthorityView, error) {
	if personINN == "" {
		return shared.AuthorityView{Type: shared.AuthorityNone, Reason: "нет полномочий"}, nil
	}
	return shared.AuthorityView{Type: shared.AuthorityHead}, nil
}

// fakeDebits — АБС: у каждой компании счета acc-1 (активный), acc-2 (заблокирован), acc-usd. Списание с acc-poor
// банк принимает и отклоняет при опросе; down — АБС не принимает поручения и не отвечает о них.
type fakeDebits struct {
	mu     sync.Mutex
	down   bool
	debits map[string]shared.DebitRequest // по id поручения
}

var errBankDown = httpx.Upstream("Сервис «АБС банка» не ответил, повторите попытку позже")

func (f *fakeDebits) Accounts(context.Context, string) ([]shared.BankAccount, error) {
	return []shared.BankAccount{
		{ID: "acc-1", Number: "40702810000000010001", Currency: currencyRUB, Active: true, Balance: ptr[int64](5_000_000)},
		{ID: "acc-poor", Number: "40702810000000010002", Currency: currencyRUB, Active: true},
		{ID: "acc-2", Number: "40702810000000010003", Currency: currencyRUB},
		{ID: "acc-usd", Number: "40702840000000010004", Currency: "USD", Active: true},
	}, nil
}

func (f *fakeDebits) Debit(_ context.Context, req shared.DebitRequest) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		return "", errBankDown
	}
	if f.debits == nil {
		f.debits = map[string]shared.DebitRequest{}
	}
	id := "dbt-" + req.RequestID
	f.debits[id] = req
	return id, nil
}

func (f *fakeDebits) DebitStatus(_ context.Context, id string) (*shared.DebitResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		return nil, errBankDown
	}
	req, ok := f.debits[id]
	if !ok {
		return nil, nil
	}
	if req.AccountID == "acc-poor" {
		return &shared.DebitResult{Status: shared.DebitFailed, Reason: "Недостаточно средств на счёте"}, nil
	}
	return &shared.DebitResult{Status: shared.DebitSucceeded, PostedAmount: req.Amount, PostedAt: today.Add(time.Minute)}, nil
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return users[token], nil
}

type env struct {
	m        *Module
	notifier *fakeNotifier
	debits   *fakeDebits
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
	e := &env{notifier: &fakeNotifier{}, debits: &fakeDebits{}}
	e.m = New(module.Deps{DB: conn, Idem: idempotency.New(conn)})
	e.m.now = func() time.Time { return today }
	e.m.Wire(Ports{
		Notifier: e.notifier, Debits: e.debits, Companies: fakeCompanies{}, Authorities: fakeAuthorities{},
		Products: fakeProducts{
			"WORKING":   {Code: "WORKING", Name: "Оборотный кредит"},
			"OVERDRAFT": {Code: "OVERDRAFT", Name: "Овердрафт", Revolving: true},
		},
	})
	if err := idempotency.Migrate(ctx, conn); err != nil {
		t.Fatal(err)
	}
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
	return e.openProduct(t, appID, company, "WORKING", issued)
}

func (e *env) openProduct(t *testing.T, appID, company, product string, issued time.Time) string {
	t.Helper()
	id, err := e.m.Loans().Open(context.Background(), shared.LoanOpening{
		AppID: appID, CompanyID: company, ApplicantID: "u-1", ContractNumber: "КД-2026-" + appID, ProductCode: product,
		Amount: 1_200_000, Term: 24, Rate: 21.5, IssuedAt: issued,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (e *env) do(method, path, user string) *httptest.ResponseRecorder {
	return e.send(method, path, user, "")
}

// send — запрос с телом и заголовками парами «имя, значение».
func (e *env) send(method, path, user, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
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
