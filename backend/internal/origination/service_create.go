package origination

import (
	"context"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// applicant — кто действует от имени компании: пользователь-клиент и его ИНН в данных банка.
type applicant struct {
	actor     shared.Actor
	companyID string
	personINN string
}

type draftInput struct {
	CommandID   string // только внутренняя команда Support, не поле HTTP-запроса
	ProductCode string
	Amount      int64
	Term        *int
	Goal        *string
}

// product — продукт заявки; неизвестный — 422: клиент выбирает продукт из каталога.
func (m *Module) product(ctx context.Context, code string) (*shared.ProductView, error) {
	p, err := m.ports.Products.Product(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("load product %s: %w", code, err)
	}
	if p == nil {
		return nil, httpx.BusinessRule("Продукт не найден, выберите продукт в каталоге")
	}
	return p, nil
}

// createDraft создаёт черновик. Компания — из пользователя, полномочия — по данным банка (Company): без полномочий
// черновика нет (403 с причиной), а снимок полномочий хранится в заявке.
func (m *Module) createDraft(ctx context.Context, who applicant, in draftInput) (*appRow, error) {
	p, err := m.product(ctx, in.ProductCode)
	if err != nil {
		return nil, err
	}
	term, err := defaultTerm(*p, in.Term)
	if err != nil {
		return nil, err
	}
	prm := params{Amount: in.Amount, Term: term, Goal: defaultGoal(*p, in.Goal)}
	if err := checkParams(*p, prm); err != nil {
		return nil, err
	}

	auth, err := m.ports.Authorities.Resolve(ctx, who.companyID, who.personINN)
	if err != nil {
		return nil, err
	}
	if auth.Type == shared.AuthorityNone {
		return nil, httpx.Forbidden(auth.Reason)
	}

	now := m.now().UTC()
	row := &appRow{
		ID:        in.CommandID,
		CompanyID: who.companyID, ApplicantID: who.actor.UserID, ProductCode: p.Code, Goal: prm.Goal,
		Amount: prm.Amount, Term: prm.Term, Status: string(shared.StatusDraft), Version: 1,
		Authority: db.JSON[authorityRecord]{V: authorityRecord{
			Type: auth.Type, MCHDNumber: auth.MCHDNumber, ValidUntil: auth.ValidUntil, Powers: auth.Powers,
		}},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := m.store.Insert(ctx, row); err != nil {
		return nil, fmt.Errorf("save draft: %w", err)
	}
	return row, nil
}
