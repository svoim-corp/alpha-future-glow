package servicing

import (
	"context"
	"strings"
	"testing"
)

func TestOpen(t *testing.T) {
	e := newEnv(t)
	id := e.open(t, "a-1", "c-1", today)
	if !strings.HasPrefix(id, "l-") {
		t.Fatalf("id кредита: %s", id)
	}
	if again := e.open(t, "a-1", "c-1", today); again != id {
		t.Fatalf("повторная выдача по заявке — тот же кредит: %s, want %s", again, id)
	}
	var l loanRow
	if err := e.m.deps.DB.Get(&l, `SELECT `+loanCols+` FROM servicing.loans WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if l.ProductName != "Оборотный кредит" || l.MonthlyPayment != 61_958 || l.Status != loanActive || l.Rate != 21.5 {
		t.Fatalf("кредит: %+v", l)
	}
	var n int
	var principal int64
	if err := e.m.deps.DB.QueryRowContext(context.Background(), `SELECT count(*), sum(principal) FROM servicing.payments WHERE loan_id = $1`, id).Scan(&n, &principal); err != nil {
		t.Fatal(err)
	}
	if n != 24 || principal != 1_200_000 {
		t.Fatalf("график: %d строк, долг %d", n, principal)
	}
}
