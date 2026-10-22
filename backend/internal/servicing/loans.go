package servicing

import "time"

// Статусы строки графика.
const (
	paymentPlanned = "PLANNED"
	paymentPaid    = "PAID"
	paymentOverdue = "OVERDUE" // не хранится: неоплаченный платёж с прошедшей датой
)

// paymentStatus — статус платежа для клиента на дату today.
func paymentStatus(p paymentRow, today time.Time) string {
	if p.Status == paymentPlanned && p.DueDate.Before(today) {
		return paymentOverdue
	}
	return p.Status
}

// loanState — остаток долга и ближайший неоплаченный платёж кредита по его графику.
type loanState struct {
	balance int64
	next    *paymentRow // nil — платежей впереди нет
}

// stateOf считает состояние кредита по строкам графика. Остаток — сумма минус весь оплаченный основной долг,
// включая частичные оплаты; ближайший платёж — первый неоплаченный не раньше today (просроченные показывает график).
func stateOf(amount int64, schedule []paymentRow, today time.Time) loanState {
	st := loanState{balance: amount}
	for i := range schedule {
		p := &schedule[i]
		st.balance -= p.PaidPrincipal
		if st.next == nil && p.Status != paymentPaid && !p.DueDate.Before(today) {
			st.next = p
		}
	}
	return st
}
