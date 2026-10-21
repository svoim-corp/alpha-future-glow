package servicing

import (
	"testing"
	"time"
)

func TestBuildSchedule(t *testing.T) {
	issued := time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		amount      int64
		rate        float64
		term, grace int
		monthly     int64
	}{
		{"оборотный", 1_500_000, 21.5, 24, 0, 77_447},
		{"с отсрочкой", 6_000_000, 18, 60, 6, 162_908},
		{"один месяц", 1_200_000, 12, 1, 0, 1_212_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monthly, ps := buildSchedule(tt.amount, tt.rate, tt.term, tt.grace, issued)
			if monthly != tt.monthly || len(ps) != tt.term {
				t.Fatalf("платёж %d, строк %d", monthly, len(ps))
			}
			var principal int64
			for i, p := range ps {
				principal += p.Principal
				switch {
				case p.Total != p.Principal+p.Interest:
					t.Fatalf("строка %d: итог не сходится %+v", i+1, p)
				case i < tt.grace && p.Principal != 0:
					t.Fatalf("строка %d: в отсрочку только проценты %+v", i+1, p)
				case i >= tt.grace && i < tt.term-1 && p.Total != monthly:
					t.Fatalf("строка %d: платёж %d, want %d", i+1, p.Total, monthly)
				case i == tt.term-1 && abs(p.Total-monthly) > int64(tt.term):
					// Последний платёж собирает округление процентов — не больше рубля за месяц.
					t.Fatalf("последний платёж %d слишком далёк от аннуитета %d", p.Total, monthly)
				}
			}
			if principal != tt.amount {
				t.Fatalf("основной долг %d, want %d", principal, tt.amount)
			}
		})
	}
	_, ps := buildSchedule(1_000_000, 20, 3, 0, issued)
	for i, want := range []string{"2026-02-28", "2026-03-31", "2026-04-30"} {
		if got := ps[i].DueDate.Format("2006-01-02"); got != want {
			t.Errorf("дата %d: %s, want %s", i+1, got, want)
		}
	}
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
