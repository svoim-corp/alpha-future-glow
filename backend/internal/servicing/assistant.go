package servicing

import (
	"context"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Assistant отдаёт ближайшие платежи компании из сессии.
func (m *Module) Assistant() shared.AssistantPayments { return assistant{m} }

type assistant struct{ m *Module }

func (p assistant) Upcoming(ctx context.Context, companyID string) ([]shared.UpcomingPayment, error) {
	rows, err := p.m.store.UpcomingPayments(ctx, companyID, p.m.today(), 10)
	if err != nil {
		return nil, err
	}
	out := make([]shared.UpcomingPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, shared.UpcomingPayment{LoanID: r.LoanID, ProductName: r.ProductName, Date: r.DueDate.Format(dateLayout), Amount: r.Total})
	}
	return out, nil
}
