package servicing

import (
	"net/http"
	"strings"
	"testing"
	"time"
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

func TestGet(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today.AddDate(0, -2, 0)) // выдан два месяца назад: два платежа пропущены
	w := e.do("GET", "/api/v1/loans/"+id, "smirnov")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[loanDetailsDTO](t, w)
	if got.ApplicationID != "a-1" || got.ContractNumber != "КД-2026-a-1" || got.Terms.MonthlyPayment != 61_958 || len(got.Payments) != 24 {
		t.Fatalf("кредит: %+v", got)
	}
	p := got.Payments
	if p[0].Date != "2026-09-21" || p[0].Status != "OVERDUE" || p[1].Status != "PLANNED" || p[0].Interest != 21_500 || p[0].Total != 61_958 {
		t.Fatalf("график: %+v %+v", p[0], p[1])
	}
	if w := e.do("GET", "/api/v1/loans/"+id, "orlova"); w.Code != http.StatusNotFound {
		t.Fatalf("чужой кредит: %d", w.Code)
	}
	if w := e.do("GET", "/api/v1/loans/l-0", "smirnov"); w.Code != http.StatusNotFound {
		t.Fatalf("нет кредита: %d", w.Code)
	}
}

func TestSchedule(t *testing.T) {
	e := newEnv(t)
	a := e.open(t, "a-1", "c-1", today.AddDate(0, -1, 0))
	b := e.open(t, "a-2", "c-1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))
	e.open(t, "a-3", "c-3", today)

	got := decode[[]scheduleItemDTO](t, e.do("GET", "/api/v1/loans/schedule?from=2026-10-01&to=2026-11-30", "smirnov"))
	want := []string{"2026-10-21 " + a + " PLANNED", "2026-11-05 " + b + " PLANNED", "2026-11-21 " + a + " PLANNED"}
	if len(got) != len(want) {
		t.Fatalf("календарь: %+v", got)
	}
	for i, w := range want {
		if g := got[i].Date + " " + got[i].LoanID + " " + got[i].Status; g != w || got[i].ProductName != "Оборотный кредит" || got[i].Amount != 61_958 {
			t.Errorf("строка %d: %+v, want %s", i, got[i], w)
		}
	}
	for _, q := range []string{"", "?from=2026-11-01", "?from=2026-12-01&to=2026-11-01", "?from=2026-01-01&to=2027-06-01", "?from=01.11.2026&to=2026-12-01"} {
		if w := e.do("GET", "/api/v1/loans/schedule"+q, "smirnov"); w.Code != http.StatusBadRequest {
			t.Errorf("%q: %d", q, w.Code)
		}
	}
}

func TestUpcoming(t *testing.T) {
	e := newEnv(t)
	a := e.open(t, "a-1", "c-1", today.AddDate(0, -2, 0)) // 21.09 просрочен, 21.10 — сегодня
	b := e.open(t, "a-2", "c-1", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))

	got := decode[[]scheduleItemDTO](t, e.do("GET", "/api/v1/loans/schedule/upcoming", "smirnov"))
	want := []string{"2026-10-21 " + a, "2026-11-05 " + b, "2026-11-21 " + a}
	if len(got) != 3 {
		t.Fatalf("по умолчанию — три: %+v", got)
	}
	for i, w := range want {
		if g := got[i].Date + " " + got[i].LoanID; g != w || got[i].Status != "PLANNED" {
			t.Errorf("строка %d: %+v, want %s", i, got[i], w)
		}
	}
	if got := decode[[]scheduleItemDTO](t, e.do("GET", "/api/v1/loans/schedule/upcoming?limit=1", "smirnov")); len(got) != 1 {
		t.Fatalf("limit=1: %+v", got)
	}
	for _, q := range []string{"0", "13", "x"} {
		if w := e.do("GET", "/api/v1/loans/schedule/upcoming?limit="+q, "smirnov"); w.Code != http.StatusBadRequest {
			t.Errorf("limit=%s: %d", q, w.Code)
		}
	}
	if w := e.do("GET", "/api/v1/loans/schedule/upcoming", "orlova"); w.Body.String() != "[]" {
		t.Fatalf("без кредитов: %s", w.Body.String())
	}
}
