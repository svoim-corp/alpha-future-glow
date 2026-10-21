package servicing

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type loanRow struct {
	ID             string    `db:"id"`
	AppID          string    `db:"app_id"`
	CompanyID      string    `db:"company_id"`
	ApplicantID    string    `db:"applicant_id"`
	ContractNumber string    `db:"contract_number"`
	ProductCode    string    `db:"product_code"`
	ProductName    string    `db:"product_name"`
	Amount         int64     `db:"amount"`
	Term           int       `db:"term"`
	Rate           float64   `db:"rate"`
	GraceMonths    int       `db:"grace_months"`
	MonthlyPayment int64     `db:"monthly_payment"`
	IssuedAt       time.Time `db:"issued_at"`
	Status         string    `db:"status"`
}

const loanCols = `id, app_id, company_id, applicant_id, contract_number, product_code, product_name, amount, term, rate,
	grace_months, monthly_payment, issued_at, status`

// InsertLoan заводит кредит с графиком в одной транзакции. Если кредит по заявке уже есть, ничего не меняет:
// в l.ID — id существующего кредита.
func (s *Store) InsertLoan(ctx context.Context, tx *sqlx.Tx, l *loanRow, schedule []payment) error {
	err := tx.GetContext(ctx, &l.ID, `
		INSERT INTO servicing.loans (id, app_id, company_id, applicant_id, contract_number, product_code, product_name,
			amount, term, rate, grace_months, monthly_payment, issued_at, status)
		VALUES ('l-' || nextval('servicing.loan_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (app_id) DO NOTHING
		RETURNING id`,
		l.AppID, l.CompanyID, l.ApplicantID, l.ContractNumber, l.ProductCode, l.ProductName,
		l.Amount, l.Term, l.Rate, l.GraceMonths, l.MonthlyPayment, l.IssuedAt, l.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.GetContext(ctx, &l.ID, `SELECT id FROM servicing.loans WHERE app_id = $1`, l.AppID)
	}
	if err != nil {
		return err
	}
	for _, p := range schedule {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO servicing.payments (loan_id, n, due_date, principal, interest, total, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'PLANNED')`, l.ID, p.N, p.DueDate, p.Principal, p.Interest, p.Total); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE servicing.loans CASCADE`)
	return err
}
