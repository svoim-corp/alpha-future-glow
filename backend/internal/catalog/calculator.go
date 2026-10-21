package catalog

import (
	"math"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
)

// paymentStep — платёж до проверки только ориентир, поэтому округляем до сотен, как и предварительный платёж этапа 1.
const paymentStep = 100

type calculation struct {
	paymentMin, paymentMax int64
	graceMin, graceMax     int64 // проценты без основного долга в отсрочку; 0 — отсрочки нет
	withinLimits           bool
}

// calculate — диапазон платежа по ставкам продукта: ставку до проверки банк не знает. В отсрочку клиент платит только
// проценты, после неё долг гасится аннуитетом за оставшиеся месяцы.
func calculate(p productRow, amount int64, term, grace int) calculation {
	c := calculation{
		paymentMin: roundPayment(finance.Annuity(amount, p.RateMin, term-grace)),
		paymentMax: roundPayment(finance.Annuity(amount, p.RateMax, term-grace)),
		withinLimits: amount >= p.AmountMin && amount <= p.AmountMax && term >= p.TermMin && term <= p.TermMax &&
			grace <= p.GraceMaxMonths && grace < term,
	}
	if grace > 0 {
		c.graceMin, c.graceMax = roundPayment(interest(amount, p.RateMin)), roundPayment(interest(amount, p.RateMax))
	}
	return c
}

// interest — проценты за месяц на сумму долга.
func interest(amount int64, annualRatePct float64) int64 {
	return int64(math.Round(float64(amount) * annualRatePct / 100 / 12))
}

func roundPayment(n int64) int64 {
	return int64(math.Round(float64(n)/paymentStep)) * paymentStep
}
