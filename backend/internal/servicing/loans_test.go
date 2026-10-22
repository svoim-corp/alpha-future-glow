package servicing

import (
	"testing"
	"time"
)

func TestStateOf(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	row := func(d int, principal int64, status string) paymentRow {
		p := paymentRow{payment: payment{DueDate: day(d), Principal: principal, Interest: 100, Total: principal + 100, Status: status}}
		if status == paymentPaid {
			p.PaidPrincipal, p.PaidInterest = p.Principal, p.Interest
		}
		return p
	}
	schedule := []paymentRow{row(1, 300, paymentPaid), row(10, 300, paymentPlanned), row(20, 300, paymentPlanned), row(30, 100, paymentPlanned)}

	tests := []struct {
		name     string
		today    time.Time
		nextDay  int // 0 — платежей впереди нет
		statuses []string
	}{
		{"до платежа", day(5), 10, []string{paymentPaid, paymentPlanned, paymentPlanned, paymentPlanned}},
		{"в день платежа он не просрочен", day(10), 10, []string{paymentPaid, paymentPlanned, paymentPlanned, paymentPlanned}},
		{"пропущенный платёж просрочен", day(15), 20, []string{paymentPaid, paymentOverdue, paymentPlanned, paymentPlanned}},
		{"график закончился", day(30).AddDate(0, 0, 1), 0, []string{paymentPaid, paymentOverdue, paymentOverdue, paymentOverdue}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stateOf(1000, schedule, tt.today)
			if st.balance != 700 {
				t.Errorf("остаток %d, want 700", st.balance)
			}
			if (st.next == nil) != (tt.nextDay == 0) || st.next != nil && !st.next.DueDate.Equal(day(tt.nextDay)) {
				t.Errorf("ближайший платёж: %+v, want день %d", st.next, tt.nextDay)
			}
			for i, want := range tt.statuses {
				if got := paymentStatus(schedule[i], tt.today); got != want {
					t.Errorf("платёж %d: %s, want %s", i+1, got, want)
				}
			}
		})
	}
}
