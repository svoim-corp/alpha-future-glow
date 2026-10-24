package planner

import (
	"context"
	"fmt"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

// planByID не раскрывает существование чужого плана и не зависит от списка последних 200.
func (m *Module) planByID(ctx context.Context, companyID, id string) (*planView, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	p, err := m.store.Plan(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load plan: %w", err)
	}
	if p == nil || p.CompanyID != companyID {
		return nil, httpx.NotFound("План не найден")
	}
	return m.withRecommendation(ctx, *p)
}
