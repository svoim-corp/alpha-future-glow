package catalog

import "testing"

func TestCalculate(t *testing.T) {
	working := productRow{AmountMin: 300_000, AmountMax: 10_000_000, TermMin: 3, TermMax: 36, RateMin: 20, RateMax: 28}
	invest := productRow{AmountMin: 1_000_000, AmountMax: 30_000_000, TermMin: 12, TermMax: 84, RateMin: 18, RateMax: 26, GraceMaxMonths: 6}

	tests := []struct {
		name        string
		p           productRow
		amount      int64
		term, grace int
		want        calculation
	}{
		{"оборотный без отсрочки", working, 1_500_000, 24, 0,
			calculation{paymentMin: 76_300, paymentMax: 82_300, withinLimits: true}},
		{"инвестиционный с отсрочкой (пример API_SPEC)", invest, 6_000_000, 60, 6,
			calculation{paymentMin: 162_900, paymentMax: 189_600, graceMin: 90_000, graceMax: 130_000, withinLimits: true}},
		{"сумма вне рамок", working, 20_000_000, 24, 0,
			calculation{paymentMin: 1_017_900, paymentMax: 1_097_800}},
		{"срок вне рамок", working, 1_500_000, 48, 0,
			calculation{paymentMin: 45_600, paymentMax: 52_300}},
		{"отсрочка у продукта без отсрочки", working, 1_500_000, 24, 3,
			calculation{paymentMin: 85_200, paymentMax: 91_200, graceMin: 25_000, graceMax: 35_000}},
		{"отсрочка длиннее разрешённой", invest, 6_000_000, 60, 12,
			calculation{paymentMin: 176_300, paymentMax: 202_300, graceMin: 90_000, graceMax: 130_000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculate(tt.p, tt.amount, tt.term, tt.grace); got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
