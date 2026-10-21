package finance

import "testing"

func TestAnnuity(t *testing.T) {
	tests := []struct {
		name      string
		principal int64
		rate      float64
		months    int
		want      int64
	}{
		{"пример из API_SPEC (там округлено до 77 800)", 1_500_000, 22, 24, 77_817},
		{"без ставки — равные доли", 1_000_000, 0, 3, 333_334},
		{"один месяц — долг с процентами", 1_200_000, 12, 1, 1_212_000},
		{"нулевой срок", 1_000_000, 20, 0, 0},
		{"нулевая сумма", 0, 20, 12, 0},
	}
	for _, tt := range tests {
		if got := Annuity(tt.principal, tt.rate, tt.months); got != tt.want {
			t.Errorf("%s: %d, want %d", tt.name, got, tt.want)
		}
	}
}
