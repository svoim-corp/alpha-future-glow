package servicing

import (
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
)

// Стратегии частичного досрочного погашения.
const (
	strategyReduceTerm    = "REDUCE_TERM"    // платёж прежний, срок короче
	strategyReducePayment = "REDUCE_PAYMENT" // срок прежний, платёж меньше
)

// reschedule — будущие строки после частичного досрочного погашения на дату paidOn (docs/servicing.md).
// future — заменяемые строки (даты позже paidOn): новый график идёт по их датам. Проценты первой строки — с paidOn
// по ставке / 365 (по paidOn они уже оплачены), дальше — по ставке / 12, как в buildSchedule. Строки с датой не
// позже graceEnd (конец отсрочки; нулевое время — отсрочки нет) остаются только с процентами; последняя строка гасит
// остаток. Номера — с firstN по порядку: после досрочного погашения номер строки уже не равен номеру месяца.
func reschedule(future []payment, balance int64, ratePct float64, graceEnd time.Time, monthly int64, strategy string,
	paidOn time.Time, firstN int) (monthlyAfter int64, out []payment) {
	if strategy == strategyReducePayment {
		amortizing := 0
		for _, r := range future {
			if r.DueDate.After(graceEnd) {
				amortizing++
			}
		}
		monthly = finance.Annuity(balance, ratePct, amortizing)
	}
	prev := paidOn
	for i, r := range future {
		if balance <= 0 {
			break
		}
		var interest int64
		if i == 0 {
			interest = roundRub(float64(balance) * ratePct / 100 / daysInYear * float64(daysBetween(prev, r.DueDate)))
		} else {
			interest = roundRub(float64(balance) * ratePct / 100 / 12)
		}
		var principal int64
		switch {
		case i == len(future)-1:
			principal = balance
		case !r.DueDate.After(graceEnd):
		default:
			principal = min(max(monthly-interest, 0), balance)
		}
		balance -= principal
		out = append(out, payment{
			N: firstN + i, DueDate: r.DueDate, Principal: principal, Interest: interest, Total: principal + interest, Status: paymentPlanned,
		})
		prev = r.DueDate
	}
	return monthly, out
}

// graceEnd — дата последнего платежа отсрочки; нулевое время — отсрочки нет.
func graceEnd(l *loanRow) time.Time {
	if l.GraceMonths == 0 {
		return time.Time{}
	}
	return finance.PaymentDate(l.IssuedAt, l.GraceMonths)
}
