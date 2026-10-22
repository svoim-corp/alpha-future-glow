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

	Version         int        `db:"version"`          // финансовая версия: If-Match денежных команд
	ScheduleVersion int        `db:"schedule_version"` // новая при перерасчёте графика
	Revolving       bool       `db:"revolving"`        // возобновляемый лимит (овердрафт)
	PenaltiesPaid   int64      `db:"penalties_paid"`
	ClosedAt        *time.Time `db:"closed_at"`
	CloseReason     *string    `db:"close_reason"`
}

const loanCols = `id, app_id, company_id, applicant_id, contract_number, product_code, product_name, amount, term, rate,
	grace_months, monthly_payment, issued_at, status, version, schedule_version, revolving, penalties_paid, closed_at, close_reason`

// InsertLoan заводит кредит с графиком в одной транзакции. Если кредит по заявке уже есть, ничего не меняет:
// в l.ID — id существующего кредита.
func (s *Store) InsertLoan(ctx context.Context, tx *sqlx.Tx, l *loanRow, schedule []payment) error {
	err := tx.GetContext(ctx, &l.ID, `
		INSERT INTO servicing.loans (id, app_id, company_id, applicant_id, contract_number, product_code, product_name,
			amount, term, rate, grace_months, monthly_payment, issued_at, status, revolving)
		VALUES ('l-' || nextval('servicing.loan_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (app_id) DO NOTHING
		RETURNING id`,
		l.AppID, l.CompanyID, l.ApplicantID, l.ContractNumber, l.ProductCode, l.ProductName,
		l.Amount, l.Term, l.Rate, l.GraceMonths, l.MonthlyPayment, l.IssuedAt, l.Status, l.Revolving)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.GetContext(ctx, &l.ID, `SELECT id FROM servicing.loans WHERE app_id = $1`, l.AppID)
	}
	if err != nil {
		return err
	}
	return s.InsertPayments(ctx, tx, l.ID, schedule)
}

// InsertPayments добавляет строки графика кредита.
func (s *Store) InsertPayments(ctx context.Context, q sqlx.ExecerContext, loanID string, rows []payment) error {
	for _, p := range rows {
		if _, err := q.ExecContext(ctx, `
			INSERT INTO servicing.payments (loan_id, n, due_date, principal, interest, total, paid_principal, paid_interest, early, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			loanID, p.N, p.DueDate, p.Principal, p.Interest, p.Total, p.PaidPrincipal, p.PaidInterest, p.Early, p.Status); err != nil {
			return err
		}
	}
	return nil
}

// paymentRow — строка графика с кредитом, к которому она относится.
type paymentRow struct {
	payment
	LoanID      string `db:"loan_id"`
	ProductName string `db:"product_name"`
}

const paymentCols = `p.loan_id, l.product_name, p.n, p.due_date, p.principal, p.interest, p.total, p.paid_principal,
	p.paid_interest, p.early, p.status`

// UpcomingPayments — неоплаченные платежи компании с датой не раньше from, ближайшие первыми.
func (s *Store) UpcomingPayments(ctx context.Context, companyID string, from time.Time, limit int) ([]paymentRow, error) {
	var out []paymentRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT `+paymentCols+` FROM servicing.payments p JOIN servicing.loans l ON l.id = p.loan_id
		WHERE l.company_id = $1 AND l.status = 'ACTIVE' AND p.status = 'PLANNED' AND p.due_date >= $2
		ORDER BY p.due_date, p.loan_id LIMIT $3`, companyID, from, limit)
	return out, err
}

// maxLoans — у компании единицы кредитов; ограничение — защита списка, а не пагинация.
const maxLoans = 100

// CompanyLoans — кредиты компании, новые первыми.
func (s *Store) CompanyLoans(ctx context.Context, companyID string) ([]loanRow, error) {
	var out []loanRow
	err := s.db.SelectContext(ctx, &out, `SELECT `+loanCols+` FROM servicing.loans
		WHERE company_id = $1 ORDER BY issued_at DESC, id DESC LIMIT $2`, companyID, maxLoans)
	return out, err
}

// CompanyPayments — все строки графиков кредитов компании по порядку: одним запросом для всего списка кредитов.
func (s *Store) CompanyPayments(ctx context.Context, q sqlx.QueryerContext, companyID string) ([]paymentRow, error) {
	var out []paymentRow
	err := sqlx.SelectContext(ctx, q, &out, `
		SELECT `+paymentCols+` FROM servicing.payments p JOIN servicing.loans l ON l.id = p.loan_id
		WHERE l.company_id = $1 ORDER BY p.loan_id, p.n`, companyID)
	return out, err
}

// Loan — кредит по id; (nil, nil) — кредита нет.
func (s *Store) Loan(ctx context.Context, id string) (*loanRow, error) {
	var l loanRow
	err := s.db.GetContext(ctx, &l, `SELECT `+loanCols+` FROM servicing.loans WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// LoanPayments — действующий график кредита по порядку.
func (s *Store) LoanPayments(ctx context.Context, q sqlx.QueryerContext, loanID string) ([]paymentRow, error) {
	var out []paymentRow
	err := sqlx.SelectContext(ctx, q, &out, `
		SELECT `+paymentCols+` FROM servicing.payments p JOIN servicing.loans l ON l.id = p.loan_id
		WHERE p.loan_id = $1 ORDER BY p.n`, loanID)
	return out, err
}

// PeriodPayments — платежи компании с from по to включительно по датам.
func (s *Store) PeriodPayments(ctx context.Context, companyID string, from, to time.Time) ([]paymentRow, error) {
	var out []paymentRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT `+paymentCols+` FROM servicing.payments p JOIN servicing.loans l ON l.id = p.loan_id
		WHERE l.company_id = $1 AND p.due_date BETWEEN $2 AND $3
		ORDER BY p.due_date, p.loan_id`, companyID, from, to)
	return out, err
}

// reminderRow — платёж, о котором пора напомнить, и кому.
type reminderRow struct {
	paymentRow
	ApplicantID    string `db:"applicant_id"`
	ContractNumber string `db:"contract_number"`
}

// DueForReminder — неоплаченные платежи действующих кредитов с from по to, о которых ещё не напоминали.
func (s *Store) DueForReminder(ctx context.Context, from, to time.Time) ([]reminderRow, error) {
	var out []reminderRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT `+paymentCols+`, l.applicant_id, l.contract_number
		FROM servicing.payments p JOIN servicing.loans l ON l.id = p.loan_id
		WHERE l.status = 'ACTIVE' AND p.status = 'PLANNED' AND p.reminded_at IS NULL AND p.due_date BETWEEN $1 AND $2
		ORDER BY p.due_date, p.loan_id`, from, to)
	return out, err
}

// ClaimReminder отмечает, что о платеже напомнили; false — уже напомнили (например, параллельный проход).
func (s *Store) ClaimReminder(ctx context.Context, loanID string, n int, at time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE servicing.payments SET reminded_at = $3 WHERE loan_id = $1 AND n = $2 AND reminded_at IS NULL`, loanID, n, at)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	return rows > 0, err
}

// UnclaimReminder снимает отметку, если уведомление не ушло.
func (s *Store) UnclaimReminder(ctx context.Context, loanID string, n int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE servicing.payments SET reminded_at = NULL WHERE loan_id = $1 AND n = $2`, loanID, n)
	return err
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE servicing.loans CASCADE`)
	return err
}
