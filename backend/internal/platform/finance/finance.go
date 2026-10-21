// Package finance — кредитная арифметика без бизнес-правил: аннуитетный платёж, даты платежей, формат суммы в рублях.
// Общая для зон, которые считают платёж: Catalog (калькулятор), Scoring (этап 1), Offer (договор), Servicing (график).
package finance

import "math"

// Annuity — ежемесячный аннуитетный платёж в рублях: principal под annualRatePct процентов годовых на months месяцев.
// Ставка 0 — равные доли долга. months <= 0 — платежа нет.
func Annuity(principal int64, annualRatePct float64, months int) int64 {
	if months <= 0 || principal <= 0 {
		return 0
	}
	if annualRatePct <= 0 {
		return int64(math.Ceil(float64(principal) / float64(months)))
	}
	r := annualRatePct / 100 / 12
	return int64(math.Round(float64(principal) * r / (1 - math.Pow(1+r, -float64(months)))))
}
