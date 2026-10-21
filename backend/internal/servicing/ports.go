package servicing

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const loanActive = "ACTIVE"

// Loans отдаёт реализацию shared.Loans другим зонам.
func (m *Module) Loans() shared.Loans { return loans{m: m} }

type loans struct{ m *Module }

// Open заводит кредит и аннуитетный график по условиям договора. Название продукта сохраняется в кредите: список
// кредитов и календарь не должны зависеть от того, что потом станет с каталогом.
func (p loans) Open(ctx context.Context, in shared.LoanOpening) (string, error) {
	product, err := p.m.ports.Products.Product(ctx, in.ProductCode)
	if err != nil {
		return "", fmt.Errorf("load product %s: %w", in.ProductCode, err)
	}
	name := in.ProductCode
	if product != nil {
		name = product.Name
	}
	monthly, schedule := buildSchedule(in.Amount, in.Rate, in.Term, in.GraceMonths, in.IssuedAt)
	l := &loanRow{
		AppID: in.AppID, CompanyID: in.CompanyID, ApplicantID: in.ApplicantID, ContractNumber: in.ContractNumber,
		ProductCode: in.ProductCode, ProductName: name, Amount: in.Amount, Term: in.Term, Rate: in.Rate,
		GraceMonths: in.GraceMonths, MonthlyPayment: monthly, IssuedAt: in.IssuedAt.UTC(), Status: loanActive,
	}
	err = db.WithTx(ctx, p.m.deps.DB, func(tx *sqlx.Tx) error { return p.m.store.InsertLoan(ctx, tx, l, schedule) })
	if err != nil {
		return "", fmt.Errorf("open loan for %s: %w", in.AppID, err)
	}
	return l.ID, nil
}
