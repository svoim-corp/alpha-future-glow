package catalog

import (
	"context"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Assistant отдаёт поддержке существующий подбор и калькулятор.
func (m *Module) Assistant() shared.AssistantCatalog { return assistant{m} }

type assistant struct{ m *Module }

func (p assistant) Search(ctx context.Context, goal string, amount int64, term int) ([]shared.ProductView, error) {
	rows, err := p.m.store.Products(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]shared.ProductView, 0, len(rows))
	if goal == "" {
		for _, r := range rows {
			out = append(out, r.view())
		}
		return out, nil
	}
	goals, err := p.m.store.Goals(ctx)
	if err != nil {
		return nil, err
	}
	name := ""
	for _, g := range goals {
		if g.Code == goal {
			name = g.Name
		}
	}
	if name == "" || amount <= 0 || term < 0 {
		return nil, httpx.Validation("Уточните цель, сумму и срок кредита")
	}
	for _, result := range match(rows, goal, name, amount, term) {
		if result.fits {
			for _, r := range rows {
				if r.Code == result.code {
					out = append(out, r.view())
				}
			}
		}
	}
	return out, nil
}

func (p assistant) Calculate(ctx context.Context, code string, amount int64, term int) (*shared.PaymentEstimate, error) {
	r, err := p.m.store.Product(ctx, code)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, httpx.NotFound("Продукт не найден")
	}
	if amount <= 0 || amount > 1000000000000 || term <= 0 || term > 1200 {
		return nil, httpx.Validation("Проверьте сумму и срок")
	}
	c := calculate(*r, amount, term, 0)
	if !c.withinLimits {
		return nil, httpx.BusinessRule("Сумма или срок вне условий продукта")
	}
	return &shared.PaymentEstimate{Amount: amount, Term: term, Min: c.paymentMin, Max: c.paymentMax, RateMin: r.RateMin, RateMax: r.RateMax}, nil
}
