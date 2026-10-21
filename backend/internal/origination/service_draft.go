package origination

import (
	"context"
	"fmt"
	"slices"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// draftPatch — изменённые клиентом поля черновика; nil — поле не меняется. Нулевое значение снимает
// необязательный параметр: отсрочку, лимит овердрафта, пояснение к цели.
type draftPatch struct {
	Amount            *int64
	Term              *int
	Goal              *string
	PurposeText       *string
	GraceMonths       *int
	OverdraftLimit    *int64
	GuarantorPersonID *string
}

// editableDraft — черновик своей компании той версии, которую видел клиент.
func (m *Module) editableDraft(ctx context.Context, id, companyID string, version int) (*appRow, error) {
	a, err := m.ownApplication(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if a.Status != string(shared.StatusDraft) {
		return nil, httpx.InvalidTransition("Заявку можно менять только в статусе «Черновик»")
	}
	if a.Version != version {
		return nil, httpx.VersionConflict()
	}
	return a, nil
}

// patchDraft — автосохранение черновика. Данные банка (компания, полномочия, адрес, обороты) не меняются: их
// в запросе быть не может. Возвращает сохранённую заявку и признак, что результат этапа 1 устарел.
func (m *Module) patchDraft(ctx context.Context, id, companyID string, version int, in draftPatch) (*appRow, bool, error) {
	a, err := m.editableDraft(ctx, id, companyID, version)
	if err != nil {
		return nil, false, err
	}
	p, err := m.product(ctx, a.ProductCode)
	if err != nil {
		return nil, false, err
	}
	next := *a
	applyPatch(&next, in)
	prm := params{Amount: next.Amount, Term: next.Term, Goal: next.Goal, GraceMonths: next.GraceMonths, OverdraftLimit: next.OverdraftLimit}
	if err := checkParams(*p, prm); err != nil {
		return nil, false, err
	}
	if in.GuarantorPersonID != nil {
		if err := m.checkGuarantor(ctx, a.CompanyID, *in.GuarantorPersonID); err != nil {
			return nil, false, err
		}
	}

	guarantorChanged := in.GuarantorPersonID != nil && (a.GuarantorPersonID == nil || *a.GuarantorPersonID != *in.GuarantorPersonID)
	var ok bool
	err = db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		if ok, err = m.store.UpdateDraft(ctx, tx, &next, version, m.now().UTC()); err != nil || !ok {
			return err
		}
		if guarantorChanged {
			return m.store.DeleteGuarantorConsent(ctx, tx, id)
		}
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("save draft %s: %w", id, err)
	}
	if !ok {
		// Между чтением и записью заявку изменили или отправили.
		return nil, false, m.conflict(ctx, id)
	}
	prelim, err := m.ports.Prelims.Prelim(ctx, id)
	if err != nil {
		return nil, false, fmt.Errorf("load prelim %s: %w", id, err)
	}
	return &next, stageOf(&next, prelim).invalidated, nil
}

func applyPatch(a *appRow, in draftPatch) {
	if in.Amount != nil {
		a.Amount = *in.Amount
	}
	if in.Term != nil {
		a.Term = *in.Term
	}
	if in.Goal != nil {
		a.Goal = in.Goal
	}
	if in.PurposeText != nil {
		a.PurposeText = in.PurposeText
		if *in.PurposeText == "" {
			a.PurposeText = nil
		}
	}
	if in.GraceMonths != nil {
		a.GraceMonths = in.GraceMonths
		if *in.GraceMonths == 0 {
			a.GraceMonths = nil
		}
	}
	if in.OverdraftLimit != nil {
		a.OverdraftLimit = in.OverdraftLimit
		if *in.OverdraftLimit == 0 {
			a.OverdraftLimit = nil
		}
	}
	if in.GuarantorPersonID != nil {
		a.GuarantorPersonID = in.GuarantorPersonID
	}
}

// checkGuarantor — поручитель только у ООО и только из кандидатов банка: руководителя и учредителей-физлиц.
func (m *Module) checkGuarantor(ctx context.Context, companyID, personID string) error {
	prof, err := m.ports.Companies.Profile(ctx, companyID)
	if err != nil {
		return err
	}
	field := func(msg string) error {
		return httpx.Validation("", httpx.FieldError{Field: "stage2.guarantor.personId", Code: "INVALID", Message: msg})
	}
	switch {
	case prof == nil:
		return fmt.Errorf("company %s not found", companyID)
	case prof.Type != shared.CompanyLLC:
		return field("У ИП поручитель не нужен")
	case !slices.ContainsFunc(prof.GuarantorCandidates, func(p shared.Person) bool { return p.PersonID == personID }):
		return field("Поручителем может быть руководитель или учредитель компании")
	}
	return nil
}

// conflict — почему не удалась запись черновика: его отправили (409) или изменили параллельно (412).
func (m *Module) conflict(ctx context.Context, id string) error {
	a, err := m.store.Application(ctx, id)
	if err != nil {
		return fmt.Errorf("load application %s: %w", id, err)
	}
	if a == nil || a.Status != string(shared.StatusDraft) {
		return httpx.InvalidTransition("Заявку можно менять только в статусе «Черновик»")
	}
	return httpx.VersionConflict()
}

const actionDeleteDraft = "DELETE_DRAFT"

// deleteDraft — черновик не стирается, а закрывается: переход DRAFT → CLOSED остаётся в истории, а клиенту
// удалённый черновик больше не виден.
func (m *Module) deleteDraft(ctx context.Context, id string, who applicant) error {
	if _, err := m.draftOf(ctx, id, who.companyID); err != nil {
		return err
	}
	_, err := m.transition(ctx, id, shared.StatusClosed, who.actor, actionDeleteDraft, "", func(a *appRow) {
		a.DeletedAt = &a.UpdatedAt
	})
	return err
}
