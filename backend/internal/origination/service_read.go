package origination

import (
	"context"
	"errors"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// ownApplication — заявка компании клиента: нет или черновик удалён — 404, чужая — 403.
func (m *Module) ownApplication(ctx context.Context, id, companyID string) (*appRow, error) {
	a, err := m.store.Application(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load application %s: %w", id, err)
	}
	if a == nil || a.DeletedAt != nil {
		return nil, httpx.NotFound("Заявка не найдена")
	}
	if a.CompanyID != companyID {
		return nil, httpx.Forbidden("Заявка принадлежит другой компании")
	}
	return a, nil
}

// dossier — всё, что мастер показывает по заявке: сама заявка, продукт, данные банка и этап.
type dossier struct {
	app              *appRow
	product          *shared.ProductView    // nil — продукт убрали из каталога
	profile          *shared.CompanyProfile // nil — данные банка сейчас недоступны
	stage            stageState
	consent          *consentRow // согласие заявителя на БКИ; nil — не дано
	guarantorConsent *consentRow // согласие поручителя на БКИ; nil — не дано
}

// dossier собирает состояние заявки. Недоступность данных банка не ломает мастер: заявка показывается без них,
// а ошибки этапа 2 подскажут обновить страницу.
func (m *Module) dossier(ctx context.Context, a *appRow) (*dossier, error) {
	d := &dossier{app: a}
	var err error
	if d.product, err = m.ports.Products.Product(ctx, a.ProductCode); err != nil {
		return nil, fmt.Errorf("load product %s: %w", a.ProductCode, err)
	}
	if d.profile, err = m.ports.Companies.Profile(ctx, a.CompanyID); err != nil && !isUpstream(err) {
		return nil, fmt.Errorf("load company profile %s: %w", a.CompanyID, err)
	}
	prelim, err := m.ports.Prelims.Prelim(ctx, a.ID)
	if err != nil {
		return nil, fmt.Errorf("load prelim %s: %w", a.ID, err)
	}
	d.stage = stageOf(a, prelim)
	if d.consent, err = m.store.Consent(ctx, a.ID, shared.ConsentApplicant); err != nil {
		return nil, fmt.Errorf("load bki consent %s: %w", a.ID, err)
	}
	if d.guarantorConsent, err = m.store.Consent(ctx, a.ID, shared.ConsentGuarantor); err != nil {
		return nil, fmt.Errorf("load guarantor consent %s: %w", a.ID, err)
	}
	return d, nil
}

func isUpstream(err error) bool {
	var e *httpx.Error
	return errors.As(err, &e) && e.Code == httpx.CodeUpstreamUnavail
}

// guarantorName — ФИО поручителя из кандидатов банка; "" — данных банка нет или человек больше не кандидат.
func (d *dossier) guarantorName(personID string) string {
	if d.profile == nil {
		return ""
	}
	for _, c := range d.profile.GuarantorCandidates {
		if c.PersonID == personID {
			return c.FullName
		}
	}
	return ""
}

// guarantorAgreed — согласие на БКИ дал именно выбранный сейчас поручитель.
func (d *dossier) guarantorAgreed() bool {
	g, c := d.app.GuarantorPersonID, d.guarantorConsent
	return g != nil && c != nil && c.PersonID != nil && *c.PersonID == *g
}
