package servicing

import (
	"context"
	"testing"
)

func TestAssistantRemainingPayment(t *testing.T) {
	e := newEnv(t)
	loan := e.open(t, "assistant-own", "c-1", today)
	other := e.open(t, "assistant-other", "c-3", today)
	ctx := context.Background()
	if _, err := e.m.deps.DB.ExecContext(ctx, `UPDATE servicing.payments SET paid_principal=1000,paid_interest=500 WHERE loan_id=$1 AND n=1`, loan); err != nil {
		t.Fatal(err)
	}
	rows, err := e.m.Assistant().Upcoming(ctx, "c-1")
	if err != nil || len(rows) == 0 {
		t.Fatalf("платежи: %+v %v", rows, err)
	}
	if rows[0].Amount != 61_958-1500 {
		t.Fatalf("чат показал исходную сумму вместо остатка: %d", rows[0].Amount)
	}
	for _, row := range rows {
		if row.LoanID == other {
			t.Fatal("платёж другой компании попал в чат")
		}
	}
}
