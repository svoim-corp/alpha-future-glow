package planner

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// PlanOwners отдаёт реализацию shared.PlanOwners другим зонам.
func (m *Module) PlanOwners() shared.PlanOwners { return planOwners{m: m} }

type planOwners struct{ m *Module }

func (p planOwners) PlanCompany(ctx context.Context, id string) (string, error) {
	pl, err := p.m.store.Plan(ctx, id)
	if err != nil || pl == nil {
		return "", err
	}
	return pl.CompanyID, nil
}
