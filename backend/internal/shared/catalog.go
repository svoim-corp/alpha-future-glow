package shared

import "context"

// Products — продукты зоны Catalog. Реализует: разработчик A. Используют: Scoring, Underwriting, Planner (B), Origination, Offer (A).
type Products interface {
	// Product возвращает продукт по коду; (nil, nil) — такого продукта нет.
	Product(ctx context.Context, code string) (*ProductView, error)
}

// ProductView — рамки продукта (API_SPEC.md, 4.4 и раздел 5).
type ProductView struct {
	Code                 string
	Name                 string
	Goals                []string
	AmountMin, AmountMax int64
	TermMin, TermMax     int
	RateMin, RateMax     float64
	GraceMaxMonths       int // 0 — отсрочки нет
	MinBusinessAgeMonths int
	GuarantorForLLC      bool
}
