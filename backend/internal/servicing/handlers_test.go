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
