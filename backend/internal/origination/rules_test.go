package origination

import (
	"errors"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var working = shared.ProductView{
	Code: "WORKING", Name: "Оборотный кредит", Goals: []string{"WORKING_CAPITAL"},
	AmountMin: 300_000, AmountMax: 10_000_000, TermMin: 3, TermMax: 36, RateMin: 20, RateMax: 28, GuarantorForLLC: true,
}

func ptr[T any](v T) *T { return &v }

// code — код прикладной ошибки или "" для nil.
func code(err error) string {
	var e *httpx.Error
	if errors.As(err, &e) {
		return e.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestCheckParams(t *testing.T) {
	tests := []struct {
		name string
		prm  params
		want string
	}{
		{"в рамках", params{Amount: 1_500_000, Term: 24, Goal: ptr("WORKING_CAPITAL")}, ""},
		{"границы включаются", params{Amount: 300_000, Term: 36}, ""},
		{"сумма меньше", params{Amount: 299_999, Term: 24}, httpx.CodeBusinessRule},
		{"сумма больше", params{Amount: 10_000_001, Term: 24}, httpx.CodeBusinessRule},
		{"срок короче", params{Amount: 1_500_000, Term: 2}, httpx.CodeBusinessRule},
		{"срок длиннее", params{Amount: 1_500_000, Term: 37}, httpx.CodeBusinessRule},
		{"чужая цель", params{Amount: 1_500_000, Term: 24, Goal: ptr("EQUIPMENT")}, httpx.CodeBusinessRule},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := code(checkParams(working, tt.prm)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	overdraft := shared.ProductView{Goals: []string{"CASH_GAP"}, TermMin: 12, TermMax: 12}
	if term, err := defaultTerm(overdraft, nil); err != nil || term != 12 {
		t.Fatalf("срок овердрафта: %d %v", term, err)
	}
	if _, err := defaultTerm(working, nil); code(err) != httpx.CodeValidation {
		t.Fatalf("без срока: %v", err)
	}
	if term, _ := defaultTerm(working, ptr(24)); term != 24 {
		t.Fatalf("указанный срок: %d", term)
	}
	if g := defaultGoal(working, nil); g == nil || *g != "WORKING_CAPITAL" {
		t.Fatalf("единственная цель: %v", g)
	}
	if g := defaultGoal(shared.ProductView{Goals: []string{"A", "B"}}, nil); g != nil {
		t.Fatalf("целей несколько — не угадываем: %v", *g)
	}
}
