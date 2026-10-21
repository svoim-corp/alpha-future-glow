package catalog

import (
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
)

func product(code string, goals []string, amountMin, amountMax int64, termMin, termMax int) productRow {
	return productRow{
		Code: code, Goals: db.JSON[[]string]{V: goals},
		AmountMin: amountMin, AmountMax: amountMax, TermMin: termMin, TermMax: termMax,
	}
}

var testLine = []productRow{
	product("WORKING", []string{"WORKING_CAPITAL"}, 300_000, 10_000_000, 3, 36),
	product("INVEST", []string{"EQUIPMENT"}, 1_000_000, 30_000_000, 12, 84),
	product("EXPRESS", []string{"WORKING_CAPITAL", "EQUIPMENT"}, 100_000, 1_000_000, 3, 24),
	product("OVERDRAFT", []string{"CASH_GAP"}, 100_000, 5_000_000, 12, 12),
}

func TestMismatch(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		goal   string
		amount int64
		term   int
		want   string
	}{
		{"подходит", "WORKING", "WORKING_CAPITAL", 1_500_000, 24, ""},
		{"другая цель", "OVERDRAFT", "WORKING_CAPITAL", 1_500_000, 24, "Не подходит для цели «Цель»"},
		{"сумма меньше", "INVEST", "EQUIPMENT", 500_000, 24, "Сумма меньше 1 000 000 ₽"},
		{"сумма больше", "EXPRESS", "WORKING_CAPITAL", 1_500_000, 12, "Сумма больше 1 000 000 ₽"},
		{"срок длиннее", "EXPRESS", "WORKING_CAPITAL", 500_000, 36, "Срок от 3 до 24 мес."},
		{"срок короче", "INVEST", "EQUIPMENT", 2_000_000, 6, "Срок от 12 до 84 мес."},
		{"фиксированный срок", "OVERDRAFT", "CASH_GAP", 500_000, 6, "Срок только 12 мес."},
		{"срок не указан — не проверяется", "INVEST", "EQUIPMENT", 2_000_000, 0, ""},
		{"границы включаются", "WORKING", "WORKING_CAPITAL", 10_000_000, 36, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p productRow
			for _, x := range testLine {
				if x.Code == tt.code {
					p = x
				}
			}
			if got := mismatch(p, tt.goal, "Цель", tt.amount, tt.term); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchOrder(t *testing.T) {
	got := match(testLine, "WORKING_CAPITAL", "Пополнить оборотные средства", 500_000, 12)
	want := []struct {
		code string
		fits bool
	}{{"WORKING", true}, {"EXPRESS", true}, {"INVEST", false}, {"OVERDRAFT", false}}
	for i, w := range want {
		if got[i].code != w.code || got[i].fits != w.fits {
			t.Fatalf("позиция %d: %+v, want %+v (всё: %+v)", i, got[i], w, got)
		}
	}
	if got[0].reason != reasonFits {
		t.Fatalf("причина подходящего: %q", got[0].reason)
	}
}

func TestRub(t *testing.T) {
	for in, want := range map[int64]string{0: "0 ₽", 999: "999 ₽", 1000: "1 000 ₽", 300_000: "300 000 ₽", 10_000_000: "10 000 000 ₽"} {
		if got := rub(in); got != want {
			t.Errorf("rub(%d) = %q, want %q", in, got, want)
		}
	}
}
