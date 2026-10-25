package servicing

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Assistant отдаёт ближайшие платежи компании из сессии.
func (m *Module) Assistant() shared.AssistantPayments { return assistant{m} }

type assistant struct{ m *Module }

// AssistantLoans возвращает список и остатки по тем же расчётам, что обычный кабинет.
func (m *Module) AssistantLoans() shared.AssistantLoans { return assistant{m} }

func (p assistant) List(ctx context.Context, companyID string) ([]shared.AssistantLoan, error) {
	loans, err := p.m.store.CompanyLoans(ctx, companyID)
	if err != nil {
		return nil, err
	}
	payments, err := p.m.store.CompanyPayments(ctx, p.m.deps.DB, companyID)
	if err != nil {
		return nil, err
	}
	rows := toLoanDTOs(loans, byLoan(payments), p.m.today())
	out := make([]shared.AssistantLoan, 0, len(rows))
	for _, l := range rows {
		out = append(out, shared.AssistantLoan{ID: l.ID, ProductName: l.Product.Name, Amount: l.Amount, Balance: l.Balance, Rate: l.Rate, Term: l.Term, Status: l.Status})
	}
	return out, nil
}

func (p assistant) Upcoming(ctx context.Context, companyID string) ([]shared.UpcomingPayment, error) {
	rows, err := p.m.store.UpcomingPayments(ctx, companyID, p.m.today(), 10)
	if err != nil {
		return nil, err
	}
	out := make([]shared.UpcomingPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, shared.UpcomingPayment{LoanID: r.LoanID, ProductName: r.ProductName, Date: r.DueDate.Format(dateLayout), Amount: r.remaining()})
	}
	return out, nil
}
