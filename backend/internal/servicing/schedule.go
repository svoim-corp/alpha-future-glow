package servicing

import (
	"math"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
)

// payment — строка графика: плановые суммы и фактически оплаченное по ней.
type payment struct {
	N             int       `db:"n"`
	DueDate       time.Time `db:"due_date"`
	Principal     int64     `db:"principal"`
	Interest      int64     `db:"interest"`
	Total         int64     `db:"total"`
	PaidPrincipal int64     `db:"paid_principal"`
	PaidInterest  int64     `db:"paid_interest"`
	Early         bool      `db:"early"` // строка досрочного погашения в дату исполнения
	Status        string    `db:"status"`
}

// buildSchedule — аннуитетный график. Проценты месяца — на остаток долга. В отсрочку — только проценты, затем
// равные платежи; последний платёж гасит остаток, чтобы сумма основного долга сходилась с кредитом до рубля.
func buildSchedule(amount int64, ratePct float64, term, grace int, issued time.Time) (monthly int64, out []payment) {
	monthly = finance.Annuity(amount, ratePct, term-grace)
	balance := amount
	for n := 1; n <= term; n++ {
		interest := int64(math.Round(float64(balance) * ratePct / 100 / 12))
		var principal int64
		switch {
		case n <= grace:
		case n == term:
			principal = balance
		default:
			principal = min(monthly-interest, balance)
		}
		balance -= principal
		out = append(out, payment{
			N: n, DueDate: finance.PaymentDate(issued, n), Principal: principal, Interest: interest, Total: principal + interest,
			Status: paymentPlanned,
		})
	}
	return monthly, out
}
