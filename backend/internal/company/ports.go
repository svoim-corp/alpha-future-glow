package company

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Companies отдаёт реализацию shared.Companies другим зонам.
func (m *Module) Companies() shared.Companies { return companies{store: m.store} }

// Authorities отдаёт реализацию shared.Authorities другим зонам.
func (m *Module) Authorities() shared.Authorities { return authoritiesStub{} }

type companies struct{ store *Store }

func (c companies) GetByID(ctx context.Context, id string) (*shared.CompanyView, error) {
	return c.store.GetByID(ctx, id)
}

// Profile — заглушка этапа 0: только данные из таблицы компаний, без АБС и реестров. Заменить реализацией (B).
func (c companies) Profile(ctx context.Context, companyID string) (*shared.CompanyProfile, error) {
	cv, err := c.store.GetByID(ctx, companyID)
	if err != nil || cv == nil {
		return nil, err
	}
	return &shared.CompanyProfile{CompanyView: *cv}, nil
}

// authoritiesStub — заглушка этапа 0: любой заявитель считается руководителем, чтобы A мог создавать черновики.
// Заменить проверкой по АБС и реестру МЧД (B).
type authoritiesStub struct{}

func (authoritiesStub) Resolve(context.Context, string, string) (shared.AuthorityView, error) {
	return shared.AuthorityView{
		Type:   shared.AuthorityHead,
		Powers: []string{shared.PowerCreditApplication, shared.PowerBKIConsent},
	}, nil
}
