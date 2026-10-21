package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Products отдаёт реализацию shared.Products другим зонам.
func (m *Module) Products() shared.Products { return products{store: m.store} }

// ProductMatcher отдаёт реализацию shared.ProductMatcher другим зонам.
func (m *Module) ProductMatcher() shared.ProductMatcher { return products{store: m.store} }

type products struct{ store *Store }

func (p products) Product(ctx context.Context, code string) (*shared.ProductView, error) {
	r, err := p.store.Product(ctx, code)
	if err != nil || r == nil {
		return nil, err
	}
	v := r.view()
	return &v, nil
}

// Match — продукты под цель и бюджет Планера, самые дешёвые первыми.
func (p products) Match(ctx context.Context, goal string, amount int64) ([]shared.ProductView, error) {
	all, err := p.store.Products(ctx)
	if err != nil {
		return nil, err
	}
	return suitable(all, goal, amount), nil
}

func (r productRow) view() shared.ProductView {
	return shared.ProductView{
		Code: r.Code, Name: r.Name, Goals: r.Goals.V,
		AmountMin: r.AmountMin, AmountMax: r.AmountMax, TermMin: r.TermMin, TermMax: r.TermMax,
		RateMin: r.RateMin, RateMax: r.RateMax, GraceMaxMonths: r.GraceMaxMonths,
		MinBusinessAgeMonths: r.MinBusinessAgeMonths, GuarantorForLLC: r.GuarantorForLLC,
	}
}
