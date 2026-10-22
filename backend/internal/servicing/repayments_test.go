package servicing

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func version(v int) string { return strconv.Itoa(v) }

func (e *env) details(t *testing.T, id string) loanDetailsDTO {
	t.Helper()
	return decode[loanDetailsDTO](t, e.do("GET", "/api/v1/loans/"+id, "smirnov"))
}

// pay — обычная оплата от smirnov; возвращает ответ как есть.
func (e *env) pay(id string, v int, key, body string) *httptest.ResponseRecorder {
	return e.send("POST", "/api/v1/loans/"+id+"/repayments", "smirnov", body, "If-Match", `"`+version(v)+`"`, "Idempotency-Key", key)
}

func TestRepaymentPartial(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0)) // первый платёж 61 958 ₽ — сегодня
	base := "/api/v1/loans/" + id

	accs := decode[[]repaymentAccountDTO](t, e.do("GET", base+"/repayment-accounts", "smirnov"))
	if len(accs) != 4 || accs[0].MaskedNumber != "•••• 0001" || !accs[0].Available || accs[2].Available || accs[3].Available {
		t.Fatalf("счета: %+v", accs)
	}

	body := `{"accountId":"acc-1","amount":20000}`
	for _, tt := range []struct {
		name    string
		headers []string
	}{
		{"без Idempotency-Key", []string{"If-Match", "1"}},
		{"без If-Match", []string{"Idempotency-Key", "k0"}},
	} {
		if w := e.send("POST", base+"/repayments", "smirnov", body, tt.headers...); w.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d %s", tt.name, w.Code, w.Body.String())
		}
	}

	w := e.pay(id, 1, "k1", body)
	if w.Code != http.StatusAccepted {
		t.Fatalf("оплата: %d %s", w.Code, w.Body.String())
	}
	op := decode[operationDTO](t, w)
	if op.Status != opProcessing || op.Type != opRegular || op.RequestedAmount != 20_000 || op.LoanVersion != 2 || op.Allocation != nil {
		t.Fatalf("операция: %+v", op)
	}
	// До подтверждения банком долг не меняется.
	if d := e.details(t, id); d.Payments[0].PaidInterest != 0 || d.Version != 2 {
		t.Fatalf("до подтверждения: %+v", d.Payments[0])
	}

	// Повтор той же команды — та же операция, хотя версия уже другая; тот же ключ с другим телом — 400.
	if again := decode[operationDTO](t, e.pay(id, 1, "k1", body)); again.ID != op.ID {
		t.Fatalf("повтор: %+v", again)
	}
	if w := e.pay(id, 1, "k1", `{"accountId":"acc-1","amount":30000}`); w.Code != http.StatusBadRequest {
		t.Fatalf("ключ с другим телом: %d", w.Code)
	}
	if w := e.pay(id, 2, "k2", body); w.Code != http.StatusConflict {
		t.Fatalf("вторая операция при незавершённой: %d %s", w.Code, w.Body.String())
	}

	got := decode[operationDTO](t, e.do("GET", base+"/repayments/"+op.ID, "smirnov"))
	if got.Status != opSucceeded || got.PostedAmount == nil || *got.PostedAmount != 20_000 || got.Allocation == nil ||
		got.Allocation.Interest != 20_000 || got.Allocation.Principal != 0 || *got.UnappliedAmount != 0 || got.LoanVersion != 3 {
		t.Fatalf("подтверждённая операция: %+v %+v", got, got.Allocation)
	}
	d := e.details(t, id)
	p := d.Payments[0]
	if p.PaidInterest != 20_000 || p.RemainingTotal != 41_958 || p.RemainingInterest != 1_500 || p.Status != paymentPlanned ||
		p.Total != 61_958 || d.Balance != 1_200_000 || p.ID != paymentID(id, 1) {
		t.Fatalf("частично оплаченная строка: %+v, остаток %d", p, d.Balance)
	}
	up := decode[[]scheduleItemDTO](t, e.do("GET", "/api/v1/loans/schedule/upcoming?limit=1", "smirnov"))
	if len(up) != 1 || up[0].Amount != 41_958 {
		t.Fatalf("ближайший платёж — остаток к оплате: %+v", up)
	}

	if w := e.pay(id, 1, "k3", body); w.Code != http.StatusPreconditionFailed {
		t.Fatalf("устаревшая версия: %d", w.Code)
	}
	if w := e.pay(id, 3, "k4", `{"accountId":"acc-1","amount":41959}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("сумма больше к оплате: %d %s", w.Code, w.Body.String())
	}
	e.m.now = func() time.Time { return today.Add(time.Hour) } // вторая операция позже первой
	rest := decode[operationDTO](t, e.pay(id, 3, "k5", `{"accountId":"acc-1","amount":41958}`))
	if got := decode[operationDTO](t, e.do("GET", base+"/repayments/"+rest.ID, "smirnov")); got.Allocation.Principal != 40_458 {
		t.Fatalf("остаток строки: %+v", got.Allocation)
	}
	if d := e.details(t, id); d.Payments[0].Status != paymentPaid || d.Balance != 1_200_000-40_458 {
		t.Fatalf("строка оплачена: %+v, остаток %d", d.Payments[0], d.Balance)
	}
	page := decode[struct {
		Items []operationDTO
		Total int
	}](t, e.do("GET", base+"/repayments?size=1", "smirnov"))
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != rest.ID {
		t.Fatalf("история: %+v", page)
	}
}

func TestRepaymentRejected(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0))
	body := func(acc string) string { return `{"accountId":"` + acc + `","amount":1000}` }

	for _, tt := range []struct {
		name, user, account string
		code                int
	}{
		{"сотрудник банка", "ivanova", "acc-1", http.StatusForbidden},
		{"чужой кредит", "orlova", "acc-1", http.StatusNotFound},
		{"нет полномочий", "guest", "acc-1", http.StatusForbidden},
		{"заблокированный счёт", "smirnov", "acc-2", http.StatusUnprocessableEntity},
		{"счёт не в рублях", "smirnov", "acc-usd", http.StatusUnprocessableEntity},
		{"чужой счёт", "smirnov", "acc-x", http.StatusBadRequest},
	} {
		w := e.send("POST", "/api/v1/loans/"+id+"/repayments", tt.user, body(tt.account), "If-Match", "1", "Idempotency-Key", tt.name)
		if w.Code != tt.code {
			t.Errorf("%s: %d, want %d: %s", tt.name, w.Code, tt.code, w.Body.String())
		}
	}

	// Банк отказал: долг не меняется, версия растёт.
	op := decode[operationDTO](t, e.pay(id, 1, "poor", body("acc-poor")))
	got := decode[operationDTO](t, e.do("GET", "/api/v1/loans/"+id+"/repayments/"+op.ID, "smirnov"))
	if got.Status != opFailed || got.FailureReason == nil || got.Allocation != nil {
		t.Fatalf("отказ банка: %+v", got)
	}
	if d := e.details(t, id); d.Payments[0].PaidInterest != 0 || d.Version != 3 {
		t.Fatalf("после отказа: %+v, версия %d", d.Payments[0], d.Version)
	}
	if w := e.do("GET", "/api/v1/loans/"+id+"/early-repayments/"+op.ID, "smirnov"); w.Code != http.StatusNotFound {
		t.Fatalf("обычная операция по пути досрочной: %d", w.Code)
	}
}

func TestRepaymentReconciler(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0))
	ctx := context.Background()
	e.debits.mu.Lock()
	e.debits.down = true
	e.debits.mu.Unlock()
	op := decode[operationDTO](t, e.pay(id, 1, "k1", `{"accountId":"acc-1","amount":61958}`))
	if op.Status != opPending {
		t.Fatalf("банк недоступен — операция ждёт: %+v", op)
	}
	if got := decode[operationDTO](t, e.do("GET", "/api/v1/loans/"+id+"/repayments/"+op.ID, "smirnov")); got.Status != opPending {
		t.Fatalf("сбой опроса не ломает чтение: %+v", got)
	}
	e.debits.mu.Lock()
	e.debits.down = false
	e.debits.mu.Unlock()
	for range 2 { // первый проход передаёт поручение, второй получает подтверждение
		if err := e.m.reconcileOpen(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if d := e.details(t, id); d.Payments[0].Status != paymentPaid {
		t.Fatalf("сверщик довёл операцию: %+v", d.Payments[0])
	}
	if _, err := e.m.Reconcile(ctx, op.ID); err != nil {
		t.Fatal(err)
	}
	var posts int
	if err := e.m.deps.DB.Get(&posts, `SELECT count(*) FROM servicing.postings WHERE op_id = $1`, op.ID); err != nil || posts != 2 {
		t.Fatalf("повторная обработка не дублирует проводки: %d %v", posts, err)
	}
}

func (e *env) quote(t *testing.T, id, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	return e.send("POST", "/api/v1/loans/"+id+"/early-repayment-quotes", "smirnov", body, "Idempotency-Key", key)
}

func (e *env) confirm(id string, v int, key, quoteID string) *httptest.ResponseRecorder {
	return e.send("POST", "/api/v1/loans/"+id+"/early-repayments", "smirnov", `{"quoteId":"`+quoteID+`","accountId":"acc-1"}`,
		"If-Match", version(v), "Idempotency-Key", key)
}

// accrued10 — проценты за 10 дней на 1 200 000 ₽ под 21,5%.
const accrued10 = 7_068

func TestEarlyFull(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, 0, -10))

	for _, tt := range []struct {
		name, body string
		code       int
	}{
		{"сумма для полного", `{"type":"FULL","amount":1000}`, http.StatusBadRequest},
		{"без стратегии", `{"type":"PARTIAL","amount":1000}`, http.StatusBadRequest},
		{"неизвестный тип", `{"type":"HALF"}`, http.StatusBadRequest},
		{"не сегодня", `{"type":"FULL","effectiveDate":"2026-10-22"}`, http.StatusUnprocessableEntity},
		{"меньше процентов", `{"type":"PARTIAL","amount":7068,"strategy":"REDUCE_TERM"}`, http.StatusUnprocessableEntity},
		{"весь долг", `{"type":"PARTIAL","amount":1207068,"strategy":"REDUCE_TERM"}`, http.StatusUnprocessableEntity},
	} {
		if w := e.quote(t, id, tt.name, tt.body); w.Code != tt.code {
			t.Errorf("%s: %d, want %d: %s", tt.name, w.Code, tt.code, w.Body.String())
		}
	}

	w := e.quote(t, id, "q1", `{"type":"FULL","effectiveDate":"2026-10-21"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("расчёт: %d %s", w.Code, w.Body.String())
	}
	q := decode[quoteDTO](t, w)
	if q.AmountToDebit != 1_200_000+accrued10 || q.Allocation.Interest != accrued10 || q.PrincipalAfter != 0 ||
		len(q.SchedulePreview) != 0 || q.ExpectedCloseDate != "2026-10-21" || len(q.Warnings) != 0 || q.LoanVersion != 1 {
		t.Fatalf("расчёт полного погашения: %+v", q)
	}
	if again := decode[quoteDTO](t, e.quote(t, id, "q1", `{"type":"FULL","effectiveDate":"2026-10-21"}`)); again.QuoteID != q.QuoteID {
		t.Fatalf("повтор расчёта: %+v", again)
	}

	op := decode[operationDTO](t, e.confirm(id, 1, "c1", q.QuoteID))
	got := decode[operationDTO](t, e.do("GET", "/api/v1/loans/"+id+"/early-repayments/"+op.ID, "smirnov"))
	if got.Status != opSucceeded || got.Type != opEarly || got.LoanStatus != loanClosed || *got.QuoteID != q.QuoteID ||
		got.Allocation.Principal != 1_200_000 || got.Allocation.Interest != accrued10 {
		t.Fatalf("досрочное погашение: %+v %+v", got, got.Allocation)
	}
	d := e.details(t, id)
	if d.Status != loanClosed || d.CloseReason == nil || *d.CloseReason != closeEarlyRepayment || d.Balance != 0 ||
		len(d.Payments) != 1 || !d.Payments[0].Early || d.Payments[0].Status != paymentPaid || d.ScheduleVersion != 2 {
		t.Fatalf("закрытый кредит: %+v", d)
	}
	debt := decode[debtDTO](t, e.do("GET", "/api/v1/loans/"+id+"/debt", "smirnov"))
	if debt.TotalOutstanding != 0 || debt.Overdue.Since != nil {
		t.Fatalf("задолженность закрытого: %+v", debt)
	}
	sum := decode[summaryResponse](t, e.do("GET", "/api/v1/loans/summary", "smirnov"))
	if sum.ClosedLoans != 1 || sum.ActiveLoans != 0 || sum.NextPayment != nil {
		t.Fatalf("сводка: %+v", sum)
	}
	if w := e.confirm(id, d.Version, "c2", q.QuoteID); w.Code != http.StatusConflict {
		t.Fatalf("закрытый кредит: %d", w.Code)
	}
}

func TestEarlyPartial(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, 0, -10))
	q := decode[quoteDTO](t, e.quote(t, id, "q1", `{"type":"PARTIAL","amount":200000,"strategy":"REDUCE_TERM"}`))
	principal := int64(200_000 - accrued10)
	if q.Allocation.Principal != principal || q.PrincipalAfter != 1_200_000-principal || q.MonthlyPaymentAfter != 61_958 ||
		q.RemainingPaymentsAfter >= 24 || q.RemainingPaymentsAfter != len(q.SchedulePreview) || q.SchedulePreview[0].Date != "2026-11-11" {
		t.Fatalf("расчёт частичного: %+v", q)
	}
	// Повтор подтверждения тем же ключом — та же операция; расчёт одноразовый.
	op := decode[operationDTO](t, e.confirm(id, 1, "c1", q.QuoteID))
	if again := decode[operationDTO](t, e.confirm(id, 1, "c1", q.QuoteID)); again.ID != op.ID {
		t.Fatalf("повтор подтверждения: %+v", again)
	}
	if got := decode[operationDTO](t, e.do("GET", "/api/v1/loans/"+id+"/early-repayments/"+op.ID, "smirnov")); got.Status != opSucceeded {
		t.Fatalf("исполнение: %+v", got)
	}
	d := e.details(t, id)
	var total int64
	for _, p := range d.Payments {
		total += p.Principal
	}
	if d.Status != loanActive || d.ScheduleVersion != 2 || d.Balance != 1_200_000-principal || total != 1_200_000 ||
		!d.Payments[0].Early || d.Payments[0].Principal != principal || len(d.Payments) != 1+q.RemainingPaymentsAfter ||
		d.Payments[1].Total != q.SchedulePreview[0].Total {
		t.Fatalf("новый график: %+v", d)
	}
	var archived int
	if err := e.m.deps.DB.Get(&archived, `SELECT count(*) FROM servicing.payments_archive WHERE loan_id = $1 AND schedule_version = 1`, id); err != nil || archived != 24 {
		t.Fatalf("архив прежнего графика: %d %v", archived, err)
	}
	if w := e.confirm(id, d.Version, "c2", q.QuoteID); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("использованный расчёт: %d %s", w.Code, w.Body.String())
	}
	stale := decode[quoteDTO](t, e.quote(t, id, "q2", `{"type":"FULL"}`))
	if _, err := e.m.deps.DB.Exec(`UPDATE servicing.loans SET version = version + 1 WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if w := e.confirm(id, d.Version+1, "c3", stale.QuoteID); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("расчёт до изменения кредита: %d %s", w.Code, w.Body.String())
	}
}

func TestOverdraftClose(t *testing.T) {
	e := newEnv(t)
	run := strconv.FormatInt(time.Now().UnixNano(), 36) + "-"
	loan := e.open(t, "a-1", "c-1", today.AddDate(0, 0, -10))
	limit := e.openProduct(t, "a-2", "c-1", "OVERDRAFT", today.AddDate(0, 0, -10))
	closeLoan := func(id string, v int, key string) *httptest.ResponseRecorder {
		// Ключи общего middleware живут между прогонами тестов, а id кредитов после пересоздания схемы повторяются.
		return e.send("POST", "/api/v1/loans/"+id+"/close", "smirnov", "", "If-Match", version(v), "Idempotency-Key", run+key)
	}
	if w := closeLoan(loan, 1, "x1"); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("обычный кредит закрывается сам: %d", w.Code)
	}
	if w := closeLoan(limit, 1, "x2"); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("лимит с долгом: %d", w.Code)
	}

	q := decode[quoteDTO](t, e.quote(t, limit, "q1", `{"type":"FULL"}`))
	if len(q.Warnings) != 1 {
		t.Fatalf("предупреждение об открытом лимите: %+v", q.Warnings)
	}
	op := decode[operationDTO](t, e.confirm(limit, 1, "c1", q.QuoteID))
	e.do("GET", "/api/v1/loans/"+limit+"/repayments/"+op.ID, "smirnov")
	d := e.details(t, limit)
	if d.Status != loanActive || !d.Revolving || d.Balance != 0 {
		t.Fatalf("погашенный овердрафт остаётся действующим: %+v", d)
	}
	w := closeLoan(limit, d.Version, "x3")
	if w.Code != http.StatusOK {
		t.Fatalf("закрытие лимита: %d %s", w.Code, w.Body.String())
	}
	if got := decode[loanDetailsDTO](t, w); got.Status != loanClosed || *got.CloseReason != closeLimit {
		t.Fatalf("закрытый лимит: %+v", got)
	}
	if w := closeLoan(limit, d.Version+1, "x4"); w.Code != http.StatusConflict {
		t.Fatalf("повторное закрытие: %d", w.Code)
	}
}

func TestDebtOverdue(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -2, 0)) // платёж 21.09 просрочен на 30 дней, 21.10 — сегодня
	d := decode[debtDTO](t, e.do("GET", "/api/v1/loans/"+id+"/debt", "smirnov"))
	penalties := int64(30 * 62)
	if d.Overdue.Since == nil || *d.Overdue.Since != "2026-09-22" || d.Overdue.Days != 30 || d.Overdue.Principal != 40_458 ||
		d.Overdue.Interest != 21_500 || d.Overdue.Penalties != penalties || d.Overdue.Total != 61_958+penalties {
		t.Fatalf("просрочка: %+v", d.Overdue)
	}
	if d.TotalOutstanding != d.PrincipalOutstanding+d.InterestOutstanding+d.PenaltiesOutstanding+d.FeesOutstanding ||
		d.PrincipalOutstanding != 1_200_000 || d.CalculationPolicyVersion != policyVersion || d.BusinessDate != "2026-10-21" {
		t.Fatalf("итог: %+v", d)
	}
	sum := decode[summaryResponse](t, e.do("GET", "/api/v1/loans/summary", "smirnov"))
	if sum.Overdue.Loans != 1 || sum.Overdue.Payments != 1 || sum.Overdue.Amount != d.Overdue.Total || sum.NextPayment.Date != "2026-10-21" {
		t.Fatalf("сводка с просрочкой: %+v", sum)
	}
	if w := e.quote(t, id, "q1", `{"type":"FULL"}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("досрочное при просрочке: %d", w.Code)
	}
	// Оплата гасит сначала просрочку, затем пени.
	op := decode[operationDTO](t, e.pay(id, 1, "k1", `{"accountId":"acc-1","amount":`+strconv.FormatInt(61_958+penalties, 10)+`}`))
	got := decode[operationDTO](t, e.do("GET", "/api/v1/loans/"+id+"/repayments/"+op.ID, "smirnov"))
	if got.Allocation.Penalties != penalties || got.Allocation.Principal != 40_458 || got.Allocation.Interest != 21_500 {
		t.Fatalf("распределение: %+v", got.Allocation)
	}
	if after := decode[debtDTO](t, e.do("GET", "/api/v1/loans/"+id+"/debt", "smirnov")); after.Overdue.Total != 0 || after.PenaltiesOutstanding != 0 {
		t.Fatalf("после погашения просрочки: %+v", after)
	}
}

func TestPage(t *testing.T) {
	e := newEnv(t)
	ids := []string{e.open(t, "a-1", "c-1", today.AddDate(0, 0, -2)), e.open(t, "a-2", "c-1", today.AddDate(0, 0, -1)), e.open(t, "a-3", "c-1", today)}
	e.open(t, "a-4", "c-3", today)
	type page struct {
		Items             []loanDTO
		Page, Size, Total int
	}
	p := decode[page](t, e.do("GET", "/api/v1/loans/page?page=1&size=2", "smirnov"))
	if p.Total != 3 || len(p.Items) != 2 || p.Items[0].ID != ids[2] || p.Items[1].ID != ids[1] || p.Size != 2 {
		t.Fatalf("первая страница: %+v", p)
	}
	if p := decode[page](t, e.do("GET", "/api/v1/loans/page?page=2&size=2&status=ACTIVE", "smirnov")); p.Total != 3 || len(p.Items) != 1 || p.Items[0].ID != ids[0] {
		t.Fatalf("вторая страница: %+v", p)
	}
	if p := decode[page](t, e.do("GET", "/api/v1/loans/page?status=CLOSED", "smirnov")); p.Total != 0 || p.Items == nil {
		t.Fatalf("закрытых нет: %+v", p)
	}
	if w := e.do("GET", "/api/v1/loans/page?status=DONE", "smirnov"); w.Code != http.StatusBadRequest {
		t.Fatalf("неизвестный статус: %d", w.Code)
	}
}

func TestDocument(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0))
	e.do("GET", "/api/v1/loans/"+id+"/repayments/"+decode[operationDTO](t, e.pay(id, 1, "k1", `{"accountId":"acc-1","amount":20000}`)).ID, "smirnov")

	w := e.do("GET", "/api/v1/loans/"+id+"/schedule/document?format=pdf", "smirnov")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/pdf" || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) ||
		w.Header().Get("Content-Disposition") != `attachment; filename="schedule-`+id+`.pdf"` || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("PDF: %d %v", w.Code, w.Header())
	}
	if w := e.do("GET", "/api/v1/loans/"+id+"/schedule/document?format=docx", "smirnov"); w.Code != http.StatusBadRequest {
		t.Fatalf("неизвестный формат: %d", w.Code)
	}
	if w := e.do("GET", "/api/v1/loans/"+id+"/schedule/document", "orlova"); w.Code != http.StatusNotFound {
		t.Fatalf("чужой кредит: %d", w.Code)
	}
}

func TestEarlyPrepaidInterest(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, 0, -10))
	// Ближайшая строка оплачена заранее на 20 000 ₽ процентов, а начислено по сегодня только accrued10.
	op := decode[operationDTO](t, e.pay(id, 1, "k1", `{"accountId":"acc-1","amount":20000}`))
	e.do("GET", "/api/v1/loans/"+id+"/repayments/"+op.ID, "smirnov")
	const credit = 20_000 - accrued10
	if d := decode[debtDTO](t, e.do("GET", "/api/v1/loans/"+id+"/debt", "smirnov")); d.PrepaidInterest != credit || d.InterestOutstanding != 0 {
		t.Fatalf("переплаченные проценты: %+v", d)
	}

	q := decode[quoteDTO](t, e.quote(t, id, "q1", `{"type":"FULL"}`))
	if q.PrepaidInterestCredit != credit || q.AmountToDebit != 1_200_000-credit || q.Allocation.Interest != 0 {
		t.Fatalf("переплата засчитывается в основной долг: %+v", q)
	}
	early := decode[operationDTO](t, e.confirm(id, 3, "c1", q.QuoteID))
	e.do("GET", "/api/v1/loans/"+id+"/early-repayments/"+early.ID, "smirnov")
	d := e.details(t, id)
	if d.Status != loanClosed || d.Balance != 0 || d.Payments[0].Interest != accrued10 || d.Payments[0].Principal != 1_200_000 {
		t.Fatalf("закрытие с зачётом процентов: %+v", d)
	}
}
