package company

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Companies отдаёт реализацию shared.Companies другим зонам.
func (m *Module) Companies() shared.Companies { return companies{store: m.store, profiles: m.profiles} }

// Authorities отдаёт реализацию shared.Authorities другим зонам.
func (m *Module) Authorities() shared.Authorities { return m.authorities }

type companies struct {
	store    *Store
	profiles *profiles
}

func (c companies) GetByID(ctx context.Context, id string) (*shared.CompanyView, error) {
	return c.store.GetByID(ctx, id)
}

func (c companies) Profile(ctx context.Context, companyID string) (*shared.CompanyProfile, error) {
	return c.profiles.Profile(ctx, companyID)
}
