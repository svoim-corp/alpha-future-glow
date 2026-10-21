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

// guarantorConsentText — согласие поручителя в демо фиксирует заявитель: поручитель — человек, известный банку,
// отдельного входа у него нет.
const guarantorConsentText = "Поручитель согласен на запрос его кредитной истории в бюро кредитных историй"

// giveGuarantorConsent фиксирует согласие выбранного поручителя. Нужен этап 2 и выбранный поручитель.
func (m *Module) giveGuarantorConsent(ctx context.Context, id string, who applicant) (*consentRow, error) {
	a, err := m.draftOf(ctx, id, who.companyID)
	if err != nil {
		return nil, err
	}
	prelim, err := m.ports.Prelims.Prelim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load prelim %s: %w", id, err)
	}
	if stageOf(a, prelim).stage != stageFull {
		return nil, httpx.InvalidTransition("Сначала пройдите предварительную проверку")
	}
	if a.GuarantorPersonID == nil {
		return nil, httpx.Validation("", httpx.FieldError{Field: "stage2.guarantor.personId", Code: "REQUIRED", Message: "Сначала выберите поручителя"})
	}
	c := &consentRow{
		AppID: a.ID, Subject: string(shared.ConsentGuarantor), PersonID: a.GuarantorPersonID, Text: guarantorConsentText,
		GivenBy: who.actor.UserID, GivenAt: m.now().UTC(),
	}
	if err := m.store.SaveConsent(ctx, c); err != nil {
		return nil, fmt.Errorf("save guarantor consent %s: %w", id, err)
	}
	return c, nil
}
