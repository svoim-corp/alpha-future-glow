package offer

import (
	"context"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestBuildOffer(t *testing.T) {
	app := &shared.ApplicationView{ID: "a-1", ProductCode: "WORKING", Amount: 1_500_000, Term: 24}
	decided := time.Date(2026, 10, 20, 23, 30, 0, 0, time.UTC)
	counter := shared.DecisionView{ID: "dec-1", Result: shared.StatusCounterOffer, Mode: shared.DecisionManual, ProductCode: "WORKING",
		Amount: 1_200_000, Term: 24, Rate: 21.5, Reason: "Сумма снижена по уровню долговой нагрузки", DecidedAt: decided}

	o := buildOffer(app, &counter)
	if o.Amount != 1_200_000 || o.RequestedAmount != 1_500_000 || o.MonthlyPayment != 61_958 || o.Explanation == nil ||
		*o.Explanation != counter.Reason || o.ValidUntil.Format("2006-01-02") != "2026-10-27" || o.Alternatives.V == nil {
		t.Fatalf("иные условия: %+v", o)
	}

	invest := &shared.ApplicationView{ID: "a-2", ProductCode: "INVEST", Amount: 6_000_000, Term: 60, GraceMonths: ptr(6)}
	withGrace := buildOffer(invest, &shared.DecisionView{Result: shared.StatusApproved, Mode: shared.DecisionAuto, Amount: 6_000_000, Term: 60, Rate: 18})
	if withGrace.MonthlyPayment != 162_908 || withGrace.GraceMonths == nil || withGrace.Explanation != nil {
		t.Fatalf("платёж после отсрочки и без пояснения к одобрению: %+v", withGrace)
	}

	rejected := buildOffer(app, &shared.DecisionView{Result: shared.StatusRejected, Mode: shared.DecisionAuto, Reason: "Автоматический отказ: балл 220",
		Alternatives: []shared.Alternative{{ProductCode: "EXPRESS", MaxAmount: 600_000}}})
	if rejected.MonthlyPayment != 0 || *rejected.Explanation != autoRejectText || len(rejected.Alternatives.V) != 1 {
		t.Fatalf("отказ без балла в тексте: %+v", rejected)
	}
}

func TestOnDecisionMade(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)
	o, err := e.m.store.Offer(context.Background(), "a-1")
	if err != nil || o == nil {
		t.Fatalf("предложение: %v", err)
	}
	if o.Decision != "APPROVED" || o.DecisionID != "dec-a-1" || o.Rate != 21.5 || o.MonthlyPayment == 0 || o.Explanation != nil {
		t.Fatalf("предложение: %+v", o)
	}

	// Новое решение по той же заявке заменяет предложение.
	e.decide(t, "a-1", "c-1", shared.DecisionView{ID: "dec-2", Result: shared.StatusCounterOffer, Mode: shared.DecisionManual,
		ProductCode: "WORKING", Amount: 1_000_000, Term: 24, Rate: 23, Reason: "Сумма снижена"})
	if o, _ := e.m.store.Offer(context.Background(), "a-1"); o.DecisionID != "dec-2" || o.Amount != 1_000_000 {
		t.Fatalf("после нового решения: %+v", o)
	}
}
