package origination

import (
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestStageOf(t *testing.T) {
	draft := &appRow{ProductCode: "WORKING", Amount: 1_500_000, Term: 24, Status: string(shared.StatusDraft)}
	submitted := &appRow{ProductCode: "WORKING", Amount: 1_500_000, Term: 24, Status: string(shared.StatusSubmitted)}
	limit := func(fits bool, choice *shared.PrelimChoice) *shared.PrelimView {
		return &shared.PrelimView{Outcome: shared.PrelimLimit, Limit: 3_300_000, AmountFits: fits, Choice: choice,
			ProductCode: "WORKING", Amount: 1_500_000, Term: 24}
	}
	stop := &shared.PrelimView{Outcome: shared.PrelimStop, ProductCode: "WORKING", Amount: 1_500_000, Term: 24}
	changed := func(f func(*shared.PrelimView)) *shared.PrelimView { p := limit(true, nil); f(p); return p }
	keep := shared.ChoiceKeepAmount

	tests := []struct {
		name        string
		app         *appRow
		prelim      *shared.PrelimView
		stage       int
		hasPrelim   bool
		invalidated bool
	}{
		{"этап 1 не пройден", draft, nil, stagePrelim, false, false},
		{"сумма в лимите", draft, limit(true, nil), stageFull, true, false},
		{"сумма выше лимита, выбора нет", draft, limit(false, nil), stagePrelim, true, false},
		{"сумма выше лимита, выбор сделан", draft, limit(false, &keep), stageFull, true, false},
		{"стоп-отказ", draft, stop, stagePrelim, true, false},
		{"изменилась сумма", draft, changed(func(p *shared.PrelimView) { p.Amount = 1_000_000 }), stagePrelim, false, true},
		{"изменился срок", draft, changed(func(p *shared.PrelimView) { p.Term = 12 }), stagePrelim, false, true},
		{"изменился продукт", draft, changed(func(p *shared.PrelimView) { p.ProductCode = "EXPRESS" }), stagePrelim, false, true},
		{"поданная заявка", submitted, limit(true, nil), stageFull, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stageOf(tt.app, tt.prelim)
			if got.stage != tt.stage || (got.prelim != nil) != tt.hasPrelim || got.invalidated != tt.invalidated {
				t.Errorf("got %+v", got)
			}
		})
	}
}
