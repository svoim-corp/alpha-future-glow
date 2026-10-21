package catalog

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const reasonFits = "Подходит по цели, сумме и сроку"

type matchResult struct {
	code   string
	fits   bool
	reason string
}

// match сверяет задачу клиента с каждым продуктом и объясняет, почему продукт подходит или нет.
// Подходящие — первыми, внутри групп сохраняется порядок каталога. term = 0 — срок не указан и не проверяется.
func match(products []productRow, goal, goalName string, amount int64, term int) []matchResult {
	out := make([]matchResult, 0, len(products))
	for _, p := range products {
		reason := mismatch(p, goal, goalName, amount, term)
		out = append(out, matchResult{code: p.Code, fits: reason == "", reason: cmp.Or(reason, reasonFits)})
	}
	slices.SortStableFunc(out, func(a, b matchResult) int {
		switch {
		case a.fits == b.fits:
			return 0
		case a.fits:
			return -1
		}
		return 1
	})
	return out
}

// mismatch — первая причина, по которой продукт не подходит; "" — подходит.
func mismatch(p productRow, goal, goalName string, amount int64, term int) string {
	switch {
	case !slices.Contains(p.Goals.V, goal):
		return "Не подходит для цели «" + goalName + "»"
	case amount < p.AmountMin:
		return "Сумма меньше " + finance.Rub(p.AmountMin)
	case amount > p.AmountMax:
		return "Сумма больше " + finance.Rub(p.AmountMax)
	case term > 0 && p.TermMin == p.TermMax && term != p.TermMin:
		return fmt.Sprintf("Срок только %d мес.", p.TermMin)
	case term > 0 && (term < p.TermMin || term > p.TermMax):
		return fmt.Sprintf("Срок от %d до %d мес.", p.TermMin, p.TermMax)
	}
	return ""
}

// suitable — продукты с целью goal, в рамки которых укладывается amount. «Лучший» для клиента — с меньшей минимальной
// ставкой; при равной ставке сохраняется порядок каталога. Срок Планер не задаёт, поэтому он не проверяется.
func suitable(products []productRow, goal string, amount int64) []shared.ProductView {
	var out []shared.ProductView
	for _, p := range products {
		if mismatch(p, goal, "", amount, 0) == "" {
			out = append(out, p.view())
		}
	}
	slices.SortStableFunc(out, func(a, b shared.ProductView) int { return cmp.Compare(a.RateMin, b.RateMin) })
	return out
}
