package servicing

import (
	"net/http"
	"strings"
	"testing"
)

func TestSummary(t *testing.T) {
	e := newEnv(t)
	if w := e.do("GET", "/api/v1/loans/summary", "smirnov"); !strings.Contains(w.Body.String(), `"activeLoans":0,"nextPayment":null`) {
		t.Fatalf("без кредитов: %s", w.Body.String())
	}
	e.open(t, "a-1", "c-1", today)
	e.open(t, "a-2", "c-1", today)
	e.open(t, "a-3", "c-1", today.AddDate(0, 0, 5))
	e.open(t, "a-4", "c-3", today)

	got := decode[summaryResponse](t, e.do("GET", "/api/v1/loans/summary", "smirnov"))
	if got.ActiveLoans != 3 || got.NextPayment == nil || got.NextPayment.Date != "2026-11-21" || got.NextPayment.Amount != 2*61_958 {
		t.Fatalf("сводка: %+v %+v", got, got.NextPayment)
	}
	if w := e.do("GET", "/api/v1/loans/summary", "ivanova"); w.Code != http.StatusForbidden {
		t.Fatalf("сотрудник банка: %d", w.Code)
	}
}

func TestList(t *testing.T) {
	e := newEnv(t)
	if w := e.do("GET", "/api/v1/loans", "smirnov"); w.Body.String() != "[]" {
		t.Fatalf("без кредитов — пустой массив: %s", w.Body.String())
	}
	first := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0))
	second := e.open(t, "a-2", "c-1", today)
	e.open(t, "a-3", "c-3", today)

	got := decode[[]loanDTO](t, e.do("GET", "/api/v1/loans", "smirnov"))
	if len(got) != 2 || got[0].ID != second || got[1].ID != first {
		t.Fatalf("кредиты компании: %+v", got)
	}
	l := got[1]
	if l.Product.Name != "Оборотный кредит" || l.Amount != 1_200_000 || l.Balance != 1_200_000 || l.Rate != 21.5 || l.Term != 24 ||
		l.IssuedAt != "2026-09-21" || l.Status != "ACTIVE" || l.NextPayment == nil || l.NextPayment.Date != "2026-10-21" || l.NextPayment.Amount != 61_958 {
		t.Fatalf("кредит: %+v %+v", l, l.NextPayment)
	}
}
