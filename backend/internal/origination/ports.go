package origination

import (
	"context"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Applications отдаёт реализацию shared.Applications другим зонам.
func (m *Module) Applications() shared.Applications { return applications{m: m} }

type applications struct{ m *Module }

// Application — заявка по id, в том числе удалённый черновик (он в статусе CLOSED); (nil, nil) — заявки нет.
func (p applications) Application(ctx context.Context, id string) (*shared.ApplicationView, error) {
	a, err := p.m.store.Application(ctx, id)
	if err != nil || a == nil {
		return nil, err
	}
	prelim, err := p.m.ports.Prelims.Prelim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load prelim %s: %w", id, err)
	}
	v := a.view()
	v.Stage = stageOf(a, prelim).stage
	return &v, nil
}

func (p applications) Transition(ctx context.Context, id string, to shared.Status, actor shared.Actor, action, comment string) error {
	_, err := p.m.transition(ctx, id, to, actor, action, comment, nil)
	return err
}

// SetAmount — сумма черновика по выбору клиента на этапе 1. Рамки продукта и лимит проверяет Scoring.
func (p applications) SetAmount(ctx context.Context, id string, amount int64, _ shared.Actor) (int, error) {
	if amount <= 0 {
		return 0, httpx.BusinessRule("Сумма должна быть больше нуля")
	}
	version, ok, err := p.m.store.SetDraftAmount(ctx, id, amount, p.m.now().UTC())
	if err != nil {
		return 0, fmt.Errorf("set amount %s: %w", id, err)
	}
	if !ok {
		return 0, p.m.conflict(ctx, id)
	}
	return version, nil
}

func (p applications) BKIConsent(ctx context.Context, appID string, subject shared.ConsentSubject) (*shared.ConsentView, error) {
	c, err := p.m.store.Consent(ctx, appID, subject)
	if err != nil || c == nil {
		return nil, err
	}
	return &shared.ConsentView{ID: c.ID, GivenBy: c.GivenBy, GivenAt: c.GivenAt}, nil
}

func (a *appRow) view() shared.ApplicationView {
	auth := a.Authority.V
	return shared.ApplicationView{
		ID: a.ID, Number: a.Number, CompanyID: a.CompanyID, ApplicantID: a.ApplicantID, ProductCode: a.ProductCode,
		Goal: deref(a.Goal), PurposeText: deref(a.PurposeText), Amount: a.Amount, Term: a.Term,
		GraceMonths: a.GraceMonths, OverdraftLimit: a.OverdraftLimit, GuarantorPersonID: a.GuarantorPersonID,
		Authority: shared.AuthorityView{Type: auth.Type, MCHDNumber: auth.MCHDNumber, ValidUntil: auth.ValidUntil, Powers: auth.Powers},
		Status:    shared.Status(a.Status), Version: a.Version, CreatedAt: a.CreatedAt, SubmittedAt: a.SubmittedAt,
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
