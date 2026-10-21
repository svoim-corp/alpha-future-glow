package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Products отдаёт реализацию shared.Products другим зонам.
func (m *Module) Products() shared.Products { return products{store: m.store} }

// ProductMatcher отдаёт реализацию shared.ProductMatcher другим зонам.
func (m *Module) ProductMatcher() shared.ProductMatcher { return productsStub{} }

type products struct{ store *Store }

func (p products) Product(ctx context.Context, code string) (*shared.ProductView, error) {
	r, err := p.store.Product(ctx, code)
	if err != nil || r == nil {
		return nil, err
	}
	v := r.view()
	return &v, nil
}

func (r productRow) view() shared.ProductView {
	return shared.ProductView{
		Code: r.Code, Name: r.Name, Goals: r.Goals.V,
		AmountMin: r.AmountMin, AmountMax: r.AmountMax, TermMin: r.TermMin, TermMax: r.TermMax,
		RateMin: r.RateMin, RateMax: r.RateMax, GraceMaxMonths: r.GraceMaxMonths,
		MinBusinessAgeMonths: r.MinBusinessAgeMonths, GuarantorForLLC: r.GuarantorForLLC,
	}
}

// productsStub — заглушка этапа 0 для подбора: продуктов нет. Заменить чтением из своей схемы (A).
type productsStub struct{}

func (productsStub) Match(context.Context, string, int64) ([]shared.ProductView, error) {
	return nil, nil
}
