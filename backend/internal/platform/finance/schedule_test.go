package finance

import (
	"testing"
	"time"
)

func TestPaymentDate(t *testing.T) {
	date := func(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }
	tests := []struct {
		from string
		n    int
		want string
	}{
		{"2026-10-21", 1, "2026-11-21"},
		{"2026-10-21", 3, "2027-01-21"},
		{"2027-01-31", 1, "2027-02-28"},
		{"2028-01-31", 1, "2028-02-29"},
		{"2027-01-31", 2, "2027-03-31"},
		{"2026-08-31", 1, "2026-09-30"},
		{"2026-10-21", 0, "2026-10-21"},
	}
	for _, tt := range tests {
		if got := PaymentDate(date(tt.from), tt.n).Format("2006-01-02"); got != tt.want {
			t.Errorf("%s + %d: %s, want %s", tt.from, tt.n, got, tt.want)
		}
	}
	// Время выдачи не влияет на дату: важен календарный день в UTC.
	if got := PaymentDate(time.Date(2026, 10, 21, 23, 59, 0, 0, time.UTC), 1); !got.Equal(date("2026-11-21")) {
		t.Errorf("время отбрасывается: %s", got)
	}
}
