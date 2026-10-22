package integrations

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var now = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

// Ответы мока для ООО «Северная пекарня» в формате API_SPEC.md, 4.14.
var fixtures = map[string]string{
	"/mock/core-banking/clients/7712345678": `{"inn":"7712345678","name":"ООО «Северная пекарня»","type":"LLC","clientSince":"2016-02-01",
		"regAddress":"г. Москва","actualAddress":"г. Москва","taxRegime":"USN_INCOME_MINUS_EXPENSES",
		"persons":[{"personId":"p-1","fullName":"Смирнов Олег Викторович","inn":"771234567890","roles":["HEAD","FOUNDER"],"share":100},
		{"personId":"p-2","fullName":"Петрова Анна Сергеевна","inn":"771298765432","roles":["REPRESENTATIVE"],"mchdNumber":"m-1"}]}`,
	"/mock/core-banking/turnover/7712345678": `{"avgMonthlyIncome":1840000,"months":[{"month":"2026-09","income":1840000,"outcome":1600000}]}`,
	"/mock/egrul/7712345678": `{"inn":"7712345678","name":"ООО «Северная пекарня»","ogrn":"1157746000000","state":"ACTIVE","regDate":"2015-03-12",
		"regAddress":"г. Москва","okved":{"code":"10.71","name":"Производство хлеба"},"head":{"fullName":"Смирнов Олег Викторович","inn":"771234567890"},
		"founders":[{"fullName":"Смирнов Олег Викторович","share":100}]}`,
	"/mock/egrip/772512345601":                 `{"inn":"772512345601","ogrnip":"326774600123456","fullName":"Орлова Мария Игоревна","state":"ACTIVE","regDate":"2026-07-01","okved":{"code":"47.91","name":"Торговля"}}`,
	"/mock/transparent-business/7712345678":    `{"massAddress":false,"massHead":false,"taxRegime":"USN_INCOME_MINUS_EXPENSES","employees":34}`,
	"/mock/bki/report":                         `{"loans":[{"amount":1500000,"balance":600000,"overdueDays":0},{"amount":500000,"balance":300000,"overdueDays":0}],"totalDebt":900000,"maxOverdueDays":0}`,
	"/mock/fns/debts/7712345678":               `{"hasDebt":false,"amount":0}`,
	"/mock/girbo/7712345678":                   `{"year":2025,"revenue":21400000,"netProfit":2100000}`,
	"/mock/courts/7712345678":                  `{"arbitration":[{"caseNumber":"А40-1/2026","role":"DEFENDANT","amount":120000}],"enforcement":[],"bankruptcy":false}`,
	"/mock/mchd/m-1":                           `{"valid":true,"principalInn":"7712345678","representative":{"fullName":"Петрова Анна Сергеевна","inn":"771298765432"},"validUntil":"2027-03-01","powers":["CREDIT_APPLICATION","BKI_CONSENT"]}`,
	"/mock/courts/broken":                      `{"arbitration":`,
	"/mock/core-banking/clients/bad-date-0000": `{"inn":"0","clientSince":"01.02.2016"}`,
	"/mock/core-banking/accounts/7712345678": `[{"id":"acc-1","number":"40702810123456780001","currency":"RUB","status":"ACTIVE","balance":5520000},
		{"id":"acc-2","number":"40702810123456780002","currency":"RUB","status":"BLOCKED"}]`,
	"/mock/core-banking/debits":         `{"debitId":"dbt-1","status":"PROCESSING"}`,
	"/mock/core-banking/debits/dbt-1":   `{"debitId":"dbt-1","requestId":"rp-1","status":"SUCCEEDED","amount":20000,"postedAmount":20000,"postedAt":"2026-10-21T10:00:05Z"}`,
	"/mock/core-banking/debits/dbt-2":   `{"debitId":"dbt-2","requestId":"rp-2","status":"FAILED","amount":20000,"reason":"Недостаточно средств на счёте"}`,
	"/mock/core-banking/debits/dbt-bad": `{"debitId":"dbt-bad","status":"SUCCEEDED"}`,
}

// mockServer отдаёт fixtures; failing — пути, на которых мок «падает» с 500. hits считает запросы к пути.
func mockServer(t *testing.T, failing ...string) (*httptest.Server, map[string]*atomic.Int32) {
	t.Helper()
	hits := map[string]*atomic.Int32{}
	for p := range fixtures {
		hits[p] = &atomic.Int32{}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, f := range failing {
			if r.URL.Path == f {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		body, ok := fixtures[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		hits[r.URL.Path].Add(1)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, hits
}

func isUpstream(err error) bool {
	var e *httpx.Error
	return errors.As(err, &e) && e.Code == httpx.CodeUpstreamUnavail
}

func TestClient(t *testing.T) {
	srv, _ := mockServer(t, "/mock/fns/debts/7712345678")
	c := newClient(srv.URL, time.Second)
	ctx := context.Background()

	bc, err := c.BankClient(ctx, "7712345678")
	if err != nil || bc == nil {
		t.Fatalf("BankClient: %v %v", bc, err)
	}
	if bc.Type != shared.CompanyLLC || bc.ClientSince.Format(dateLayout) != "2016-02-01" || len(bc.Persons) != 2 ||
		bc.Persons[0].Roles[0] != shared.PersonHead || bc.Persons[0].Share != 100 || *bc.Persons[1].MCHDNumber != "m-1" {
		t.Fatalf("BankClient: %+v", bc)
	}

	if bc, err := c.BankClient(ctx, "0000000000"); bc != nil || err != nil {
		t.Fatalf("404 должен давать (nil, nil): %v %v", bc, err)
	}
	if _, err := c.fns(ctx, "7712345678"); !isUpstream(err) {
		t.Fatalf("500 должен давать UPSTREAM_UNAVAILABLE: %v", err)
	}
	if _, err := c.courts(ctx, "broken"); !isUpstream(err) {
		t.Fatalf("битый JSON должен давать UPSTREAM_UNAVAILABLE: %v", err)
	}
	if _, err := c.BankClient(ctx, "bad-date-0000"); !isUpstream(err) {
		t.Fatalf("некорректная дата должна давать UPSTREAM_UNAVAILABLE: %v", err)
	}

	m, err := c.MCHD(ctx, "m-1")
	if err != nil || !m.Valid || m.ValidUntil.Format(dateLayout) != "2027-03-01" || m.Representative.INN != "771298765432" {
		t.Fatalf("MCHD: %+v %v", m, err)
	}
}

func TestDebits(t *testing.T) {
	srv, _ := mockServer(t)
	c := newClient(srv.URL, time.Second)
	ctx := context.Background()

	accs, err := c.Accounts(ctx, "7712345678")
	if err != nil || len(accs) != 2 || !accs[0].Active || *accs[0].Balance != 5_520_000 || accs[1].Active || accs[1].Balance != nil {
		t.Fatalf("Accounts: %+v %v", accs, err)
	}
	if accs, err := c.Accounts(ctx, "0000000000"); err != nil || accs == nil || len(accs) != 0 {
		t.Fatalf("404 должен давать пустой список: %v %v", accs, err)
	}

	id, err := c.Debit(ctx, shared.DebitRequest{RequestID: "rp-1", INN: "7712345678", AccountID: "acc-1", Amount: 20_000})
	if err != nil || id != "dbt-1" {
		t.Fatalf("Debit: %q %v", id, err)
	}
	r, err := c.DebitStatus(ctx, "dbt-1")
	if err != nil || r.Status != shared.DebitSucceeded || r.PostedAmount != 20_000 || !r.PostedAt.Equal(now.Add(5*time.Second)) {
		t.Fatalf("DebitStatus SUCCEEDED: %+v %v", r, err)
	}
	if r, err := c.DebitStatus(ctx, "dbt-2"); err != nil || r.Status != shared.DebitFailed || r.Reason == "" {
		t.Fatalf("DebitStatus FAILED: %+v %v", r, err)
	}
	if r, err := c.DebitStatus(ctx, "dbt-404"); r != nil || err != nil {
		t.Fatalf("неизвестное поручение должно давать (nil, nil): %+v %v", r, err)
	}
	if _, err := c.DebitStatus(ctx, "dbt-bad"); !isUpstream(err) {
		t.Fatalf("SUCCEEDED без проводки должен давать UPSTREAM_UNAVAILABLE: %v", err)
	}
}

func TestClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	_, err := newClient(srv.URL, 20*time.Millisecond).Turnover(context.Background(), "7712345678")
	if !isUpstream(err) {
		t.Fatalf("таймаут должен давать UPSTREAM_UNAVAILABLE: %v", err)
	}
	var e *httpx.Error
	errors.As(err, &e)
	if strings.Contains(e.Message, "mock") || !strings.Contains(e.Message, nameBank) {
		t.Fatalf("сообщение клиенту: %q", e.Message)
	}
}

func TestRegistration(t *testing.T) {
	srv, _ := mockServer(t)
	c := newClient(srv.URL, time.Second)
	ctx := context.Background()

	llc, err := c.Registration(ctx, "7712345678", shared.CompanyLLC)
	if err != nil || llc.OGRN != "1157746000000" || llc.Head.INN != "771234567890" || len(llc.Founders) != 1 {
		t.Fatalf("ЕГРЮЛ: %+v %v", llc, err)
	}
	ie, err := c.Registration(ctx, "772512345601", shared.CompanyIE)
	if err != nil || ie.Name != "Орлова Мария Игоревна" || ie.OGRN != "326774600123456" || ie.Head != nil {
		t.Fatalf("ЕГРИП: %+v %v", ie, err)
	}
	if r, err := c.Registration(ctx, "7712345678", shared.CompanyIE); r != nil || err != nil {
		t.Fatalf("ООО нет в ЕГРИП: %+v %v", r, err)
	}
}

func TestCollect(t *testing.T) {
	srv, hits := mockServer(t, "/mock/courts/7712345678")
	ch := &checker{client: newClient(srv.URL, time.Second), now: func() time.Time { return now }}

	rep := ch.collect(context.Background(), shared.CheckRequest{AppID: "a-1", INN: "7712345678", CompanyType: shared.CompanyLLC, ConsentID: "cons-1"})
	if len(rep.Results) != len(shared.AllSources) {
		t.Fatalf("results: %d", len(rep.Results))
	}
	for i, r := range rep.Results {
		if r.Source != shared.AllSources[i] {
			t.Fatalf("порядок источников: %d %s", i, r.Source)
		}
		want := shared.CheckOK
		if r.Source == shared.SourceCourts {
			want = shared.CheckUnavailable
		}
		if r.Status != want || !r.RequestedAt.Equal(now) {
			t.Errorf("%s: %s %q", r.Source, r.Status, r.Summary)
		}
	}
	f := rep.Facts
	if f.Client == nil || f.Turnover == nil || f.Registration == nil || f.Transparent == nil || f.BKI == nil || f.FNS == nil || f.GIRBO == nil {
		t.Fatalf("facts: %+v", f)
	}
	if f.Courts != nil {
		t.Fatal("у не ответившего источника фактов быть не должно")
	}

	// ИП: без ГИР БО; без согласия БКИ не запрашивается.
	before := hits["/mock/bki/report"].Load()
	rep = ch.collect(context.Background(), shared.CheckRequest{AppID: "a-2", INN: "772512345601", CompanyType: shared.CompanyIE,
		Sources: []shared.CheckSource{shared.SourceEGRUL, shared.SourceBKI, shared.SourceGIRBO}})
	if len(rep.Results) != 2 || rep.Results[0].Source != shared.SourceEGRUL || rep.Results[1].Status != shared.CheckUnavailable {
		t.Fatalf("ИП: %+v", rep.Results)
	}
	if hits["/mock/bki/report"].Load() != before {
		t.Fatal("БКИ запрошено без согласия")
	}
	if rep.Results[0].Summary != "Действует, 3 мес." {
		t.Fatalf("возраст ИП: %q", rep.Results[0].Summary)
	}
}

func TestAssess(t *testing.T) {
	tests := []struct {
		name    string
		src     shared.CheckSource
		fact    any
		status  shared.CheckStatus
		summary string
	}{
		{"БКИ без просрочек", shared.SourceBKI, &shared.BKIReport{Loans: make([]shared.BKILoan, 2)}, shared.CheckOK, "2 кредита, просрочек нет"},
		{"БКИ без кредитов", shared.SourceBKI, &shared.BKIReport{}, shared.CheckOK, "Кредитов нет"},
		{"БКИ просрочка 12 дней", shared.SourceBKI, &shared.BKIReport{MaxOverdueDays: 12}, shared.CheckWarn, "Просрочка до 30 дней (12 дн.)"},
		{"БКИ просрочка 95 дней", shared.SourceBKI, &shared.BKIReport{MaxOverdueDays: 95}, shared.CheckFail, "Просрочка 95 дн."},
		{"ликвидация", shared.SourceEGRUL, &shared.Registration{State: "LIQUIDATING"}, shared.CheckFail, "Не действует: LIQUIDATING"},
		{"действует 11 лет", shared.SourceEGRUL, &shared.Registration{State: "ACTIVE", RegDate: time.Date(2015, 3, 12, 0, 0, 0, 0, time.UTC)}, shared.CheckOK, "Действует, 11 лет"},
		{"банкротство", shared.SourceCourts, &shared.CourtRecords{Bankruptcy: true}, shared.CheckFail, "Сведения о банкротстве в ЕФРСБ"},
		{"иск к компании", shared.SourceCourts, &shared.CourtRecords{Arbitration: []shared.ArbitrationCase{{Role: "DEFENDANT", Amount: 120000}, {Role: "PLAINTIFF", Amount: 5}}}, shared.CheckWarn, "1 арбитражное дело, 120 000 ₽"},
		{"иск от компании", shared.SourceCourts, &shared.CourtRecords{Arbitration: []shared.ArbitrationCase{{Role: "PLAINTIFF", Amount: 5}}}, shared.CheckOK, "Судебных дел нет"},
		{"налоговый долг", shared.SourceFNS, &shared.TaxDebt{HasDebt: true, Amount: 85000}, shared.CheckWarn, "Налоговая задолженность 85 000 ₽"},
		{"массовый адрес", shared.SourceTransparentBusiness, &shared.TransparentBusiness{MassAddress: true}, shared.CheckWarn, "Массовый адрес регистрации"},
		{"нет в АБС", shared.SourceBank, nil, shared.CheckFail, "Клиент не найден в АБС банка"},
		{"нет отчётности", shared.SourceGIRBO, nil, shared.CheckWarn, "Отчётность в ГИР БО не найдена"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := assess(tt.src, tt.fact, now)
			if r.Status != tt.status || r.Summary != tt.summary {
				t.Fatalf("%s %q, want %s %q", r.Status, r.Summary, tt.status, tt.summary)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct{ got, want string }{
		{rub(0), "0 ₽"},
		{rub(120000), "120 000 ₽"},
		{rub(1840000), "1 840 000 ₽"},
		{rub(-350000), "-350 000 ₽"},
		{plural(1, "год", "года", "лет"), "1 год"},
		{plural(3, "год", "года", "лет"), "3 года"},
		{plural(11, "год", "года", "лет"), "11 лет"},
		{plural(21, "год", "года", "лет"), "21 год"},
		{plural(112, "год", "года", "лет"), "112 лет"},
		{age(time.Date(2025, 10, 22, 0, 0, 0, 0, time.UTC), now), "11 мес."},
		{age(time.Date(2025, 10, 21, 0, 0, 0, 0, time.UTC), now), "1 год"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%q, want %q", tt.got, tt.want)
		}
	}
}

// TestFactRoundTrip — снимок, сохранённый в payload, читается обратно в тот же тип.
func TestFactRoundTrip(t *testing.T) {
	facts := shared.CheckFacts{
		BKI:          &shared.BKIReport{TotalDebt: 900000, Loans: []shared.BKILoan{{Amount: 1, Balance: 2}}},
		Registration: &shared.Registration{State: "ACTIVE", RegDate: time.Date(2015, 3, 12, 0, 0, 0, 0, time.UTC)},
	}
	var got shared.CheckFacts
	for _, s := range shared.AllSources {
		p, err := encodeFact(factOf(facts, s))
		if err != nil {
			t.Fatal(err)
		}
		if p == nil {
			continue
		}
		f, err := decodeFact(s, []byte(*p))
		if err != nil {
			t.Fatal(err)
		}
		setFact(&got, f)
	}
	if got.BKI == nil || got.BKI.TotalDebt != 900000 || len(got.BKI.Loans) != 1 || !got.Registration.RegDate.Equal(facts.Registration.RegDate) || got.Client != nil {
		t.Fatalf("%+v", got)
	}
}
