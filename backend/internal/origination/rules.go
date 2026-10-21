package origination

import (
	"fmt"
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// params — параметры кредита, которые выбирает клиент.
type params struct {
	Amount int64
	Term   int
	Goal   *string
}

// checkParams — параметры в рамках продукта (docs/origination.md): нарушение — 422 BUSINESS_RULE с понятным текстом.
func checkParams(p shared.ProductView, prm params) error {
	switch {
	case prm.Amount < p.AmountMin || prm.Amount > p.AmountMax:
		return httpx.BusinessRule(fmt.Sprintf("Сумма вне рамок продукта: от %s до %s", finance.Rub(p.AmountMin), finance.Rub(p.AmountMax)))
	case prm.Term < p.TermMin || prm.Term > p.TermMax:
		return httpx.BusinessRule(fmt.Sprintf("Срок вне рамок продукта: от %d до %d мес.", p.TermMin, p.TermMax))
	case prm.Goal != nil && !slices.Contains(p.Goals, *prm.Goal):
		return httpx.BusinessRule("Продукт «" + p.Name + "» не подходит для этой цели")
	}
	return nil
}

// defaultTerm — срок, если клиент его не указал: у продукта с единственным сроком (овердрафт) он подставляется сам.
func defaultTerm(p shared.ProductView, term *int) (int, error) {
	switch {
	case term != nil:
		return *term, nil
	case p.TermMin == p.TermMax:
		return p.TermMin, nil
	}
	return 0, httpx.Validation("", httpx.FieldError{Field: "term", Code: "REQUIRED", Message: "Укажите срок кредита"})
}

// defaultGoal — цель, если клиент её не указал: у продукта с единственной целью она очевидна.
func defaultGoal(p shared.ProductView, goal *string) *string {
	if goal == nil && len(p.Goals) == 1 {
		return &p.Goals[0]
	}
	return goal
}
