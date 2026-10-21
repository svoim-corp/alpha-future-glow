package origination

import (
	"context"
	"fmt"
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// draftOf — черновик своей компании: согласия и проверки этапов имеют смысл только до отправки.
func (m *Module) draftOf(ctx context.Context, id, companyID string) (*appRow, error) {
	a, err := m.ownApplication(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if a.Status != string(shared.StatusDraft) {
		return nil, httpx.InvalidTransition("Действие доступно только для черновика")
	}
	return a, nil
}

// giveBKIConsent фиксирует согласие заявителя на запрос в БКИ. Право — по снимку полномочий: представителю по МЧД
// нужно полномочие BKI_CONSENT. Повторное согласие не создаёт новое: возвращается уже данное.
func (m *Module) giveBKIConsent(ctx context.Context, id string, who applicant, text string) (*consentRow, error) {
	a, err := m.draftOf(ctx, id, who.companyID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(a.Authority.V.Powers, shared.PowerBKIConsent) {
		return nil, httpx.Forbidden("Доверенность не даёт права давать согласие на запрос в БКИ. Согласие даёт руководитель")
	}
	c := &consentRow{AppID: a.ID, Subject: string(shared.ConsentApplicant), Text: text, GivenBy: who.actor.UserID, GivenAt: m.now().UTC()}
	if err := m.store.SaveConsent(ctx, c); err != nil {
		return nil, fmt.Errorf("save bki consent %s: %w", id, err)
	}
	return c, nil
}
