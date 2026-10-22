package servicing

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
)

// Snapshot выполняет чтения fn в одном снимке БД: список и счётчик, график, проводки и версии согласованы между
// собой, даже если параллельно проводится оплата.
func (s *Store) Snapshot(ctx context.Context, fn func(q sqlx.ExtContext) error) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	return fn(tx)
}

// LockLoan — кредит под блокировкой до конца транзакции; (nil, nil) — кредита нет.
func (s *Store) LockLoan(ctx context.Context, tx *sqlx.Tx, id string) (*loanRow, error) {
	var l loanRow
	err := tx.GetContext(ctx, &l, `SELECT `+loanCols+` FROM servicing.loans WHERE id = $1 FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// LoanIn — кредит по id в транзакции или снимке; (nil, nil) — кредита нет.
func (s *Store) LoanIn(ctx context.Context, q sqlx.QueryerContext, id string) (*loanRow, error) {
	var l loanRow
	err := sqlx.GetContext(ctx, q, &l, `SELECT `+loanCols+` FROM servicing.loans WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateLoan сохраняет финансовое состояние кредита: версии, оплаченные пени, платёж и закрытие.
func (s *Store) UpdateLoan(ctx context.Context, tx *sqlx.Tx, l *loanRow) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE servicing.loans SET version = $2, schedule_version = $3, penalties_paid = $4, monthly_payment = $5,
			status = $6, closed_at = $7, close_reason = $8
		WHERE id = $1`,
		l.ID, l.Version, l.ScheduleVersion, l.PenaltiesPaid, l.MonthlyPayment, l.Status, l.ClosedAt, l.CloseReason)
	return err
}

// UpdatePaid сохраняет оплаченные суммы и статус строк.
func (s *Store) UpdatePaid(ctx context.Context, tx *sqlx.Tx, loanID string, rows []payment) error {
	for _, r := range rows {
		if _, err := tx.ExecContext(ctx, `
			UPDATE servicing.payments SET paid_principal = $3, paid_interest = $4, status = $5 WHERE loan_id = $1 AND n = $2`,
			loanID, r.N, r.PaidPrincipal, r.PaidInterest, r.Status); err != nil {
			return err
		}
	}
	return nil
}

// ArchiveFrom переносит строки с номера fromN в архив версии графика version и удаляет их из действующего графика.
func (s *Store) ArchiveFrom(ctx context.Context, tx *sqlx.Tx, loanID string, version, fromN int) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO servicing.payments_archive (loan_id, schedule_version, n, due_date, principal, interest, total,
			paid_principal, paid_interest)
		SELECT loan_id, $2, n, due_date, principal, interest, total, paid_principal, paid_interest
		FROM servicing.payments WHERE loan_id = $1 AND n >= $3`, loanID, version, fromN); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM servicing.payments WHERE loan_id = $1 AND n >= $2`, loanID, fromN)
	return err
}

// LoanPostings — проводки кредита по времени.
func (s *Store) LoanPostings(ctx context.Context, q sqlx.QueryerContext, loanID string) ([]posting, error) {
	var out []posting
	err := sqlx.SelectContext(ctx, q, &out, `
		SELECT n, component, amount, posted_at FROM servicing.postings WHERE loan_id = $1 ORDER BY posted_at, id`, loanID)
	return out, err
}

// companyPosting — проводка с кредитом, к которому она относится.
type companyPosting struct {
	posting
	LoanID string `db:"loan_id"`
}

// CompanyPostings — проводки всех кредитов компании по времени.
func (s *Store) CompanyPostings(ctx context.Context, q sqlx.QueryerContext, companyID string) ([]companyPosting, error) {
	var out []companyPosting
	err := sqlx.SelectContext(ctx, q, &out, `
		SELECT t.loan_id, t.n, t.component, t.amount, t.posted_at
		FROM servicing.postings t JOIN servicing.loans l ON l.id = t.loan_id
		WHERE l.company_id = $1 ORDER BY t.posted_at, t.id`, companyID)
	return out, err
}

func (s *Store) InsertPostings(ctx context.Context, tx *sqlx.Tx, opID, loanID string, posts []posting) error {
	for _, p := range posts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO servicing.postings (op_id, loan_id, n, component, amount, posted_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			opID, loanID, p.N, p.Component, p.Amount, p.PostedAt); err != nil {
			return err
		}
	}
	return nil
}

// opRow — денежная операция по кредиту: обычная оплата или досрочное погашение.
type opRow struct {
	ID              string     `db:"id"`
	LoanID          string     `db:"loan_id"`
	Type            string     `db:"type"`
	Status          string     `db:"status"`
	AccountID       string     `db:"account_id"`
	AccountNumber   string     `db:"account_number"`
	RequestedAmount int64      `db:"requested_amount"`
	QuoteID         *string    `db:"quote_id"`
	IdemKey         string     `db:"idem_key"`
	Fingerprint     string     `db:"fingerprint"`
	ExternalID      *string    `db:"external_id"`
	PostedAmount    *int64     `db:"posted_amount"`
	PostedAt        *time.Time `db:"posted_at"`
	FailureReason   *string    `db:"failure_reason"`
	AllocPrincipal  *int64     `db:"alloc_principal"`
	AllocInterest   *int64     `db:"alloc_interest"`
	AllocPenalties  *int64     `db:"alloc_penalties"`
	AllocFees       *int64     `db:"alloc_fees"`
	Unapplied       *int64     `db:"unapplied"`
	LoanVersion     int        `db:"loan_version_after"`
	ScheduleVersion int        `db:"schedule_version_after"`
	LoanStatus      string     `db:"loan_status_after"`
	RequestedAt     time.Time  `db:"requested_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

const opCols = `id, loan_id, type, status, account_id, account_number, requested_amount, quote_id, idem_key, fingerprint,
	external_id, posted_amount, posted_at, failure_reason, alloc_principal, alloc_interest, alloc_penalties, alloc_fees,
	unapplied, loan_version_after, schedule_version_after, loan_status_after, requested_at, updated_at`

// Префиксы id операций по типу: rp — обычная оплата, er — досрочное погашение.
var opPrefix = map[string]string{opRegular: "rp-", opEarly: "er-"}

// InsertOperation заводит операцию; id присваивает БД.
func (s *Store) InsertOperation(ctx context.Context, tx *sqlx.Tx, op *opRow) error {
	return tx.GetContext(ctx, &op.ID, `
		INSERT INTO servicing.operations (id, loan_id, type, status, account_id, account_number, requested_amount, quote_id,
			idem_key, fingerprint, loan_version_after, schedule_version_after, loan_status_after, requested_at, updated_at)
		VALUES ($1 || nextval('servicing.operation_seq'), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
		RETURNING id`,
		opPrefix[op.Type], op.LoanID, op.Type, op.Status, op.AccountID, op.AccountNumber, op.RequestedAmount, op.QuoteID,
		op.IdemKey, op.Fingerprint, op.LoanVersion, op.ScheduleVersion, op.LoanStatus, op.RequestedAt)
}

func getOp(ctx context.Context, q sqlx.QueryerContext, query string, args ...any) (*opRow, error) {
	var op opRow
	err := sqlx.GetContext(ctx, q, &op, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &op, nil
}

// OperationByKey — операция кредита по ключу идемпотентности; (nil, nil) — такой команды не было.
func (s *Store) OperationByKey(ctx context.Context, q sqlx.QueryerContext, loanID, key string) (*opRow, error) {
	return getOp(ctx, q, `SELECT `+opCols+` FROM servicing.operations WHERE loan_id = $1 AND idem_key = $2`, loanID, key)
}

// Operation — операция по id; (nil, nil) — операции нет.
func (s *Store) Operation(ctx context.Context, q sqlx.QueryerContext, id string) (*opRow, error) {
	return getOp(ctx, q, `SELECT `+opCols+` FROM servicing.operations WHERE id = $1`, id)
}

// LockOperation — операция под блокировкой до конца транзакции; (nil, nil) — операции нет.
func (s *Store) LockOperation(ctx context.Context, tx *sqlx.Tx, id string) (*opRow, error) {
	return getOp(ctx, tx, `SELECT `+opCols+` FROM servicing.operations WHERE id = $1 FOR UPDATE`, id)
}

// HasOpenOperation — есть ли у кредита незавершённая операция.
func (s *Store) HasOpenOperation(ctx context.Context, q sqlx.QueryerContext, loanID string) (bool, error) {
	var open bool
	err := sqlx.GetContext(ctx, q, &open, `
		SELECT EXISTS (SELECT 1 FROM servicing.operations WHERE loan_id = $1 AND status IN ('PENDING', 'PROCESSING'))`, loanID)
	return open, err
}

// OpenOperations — незавершённые операции всех кредитов, старые первыми: их дообрабатывает сверщик.
func (s *Store) OpenOperations(ctx context.Context, limit int) ([]opRow, error) {
	var out []opRow
	err := s.db.SelectContext(ctx, &out, `SELECT `+opCols+` FROM servicing.operations
		WHERE status IN ('PENDING', 'PROCESSING') ORDER BY requested_at, id LIMIT $1`, limit)
	return out, err
}

// RecentOperations — последние операции кредита, новые первыми.
func (s *Store) RecentOperations(ctx context.Context, q sqlx.QueryerContext, loanID string, limit int) ([]opRow, error) {
	var out []opRow
	err := sqlx.SelectContext(ctx, q, &out, `SELECT `+opCols+` FROM servicing.operations WHERE loan_id = $1
		ORDER BY requested_at DESC, id DESC LIMIT $2`, loanID, limit)
	return out, err
}

// LoanOperations — страница истории операций кредита и их общее число, из одного снимка.
func (s *Store) LoanOperations(ctx context.Context, loanID string, limit, offset int) (ops []opRow, total int, err error) {
	err = s.Snapshot(ctx, func(q sqlx.ExtContext) error {
		if err := sqlx.GetContext(ctx, q, &total, `SELECT count(*) FROM servicing.operations WHERE loan_id = $1`, loanID); err != nil {
			return err
		}
		return sqlx.SelectContext(ctx, q, &ops, `SELECT `+opCols+` FROM servicing.operations WHERE loan_id = $1
			ORDER BY requested_at DESC, id DESC LIMIT $2 OFFSET $3`, loanID, limit, offset)
	})
	return ops, total, err
}

// AcceptOperation отмечает, что банк принял поручение: PENDING → PROCESSING с id поручения.
func (s *Store) AcceptOperation(ctx context.Context, id, externalID string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE servicing.operations SET status = 'PROCESSING', external_id = $2, updated_at = $3
		WHERE id = $1 AND status IN ('PENDING', 'PROCESSING')`, id, externalID, at)
	return err
}

// FinishOperation сохраняет итог операции.
func (s *Store) FinishOperation(ctx context.Context, tx *sqlx.Tx, op *opRow) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE servicing.operations SET status = $2, posted_amount = $3, posted_at = $4, failure_reason = $5,
			alloc_principal = $6, alloc_interest = $7, alloc_penalties = $8, alloc_fees = $9, unapplied = $10,
			loan_version_after = $11, schedule_version_after = $12, loan_status_after = $13, updated_at = $14
		WHERE id = $1`,
		op.ID, op.Status, op.PostedAmount, op.PostedAt, op.FailureReason, op.AllocPrincipal, op.AllocInterest,
		op.AllocPenalties, op.AllocFees, op.Unapplied, op.LoanVersion, op.ScheduleVersion, op.LoanStatus, op.UpdatedAt)
	return err
}

// LoanUnapplied — нераспределённые средства кредита.
func (s *Store) LoanUnapplied(ctx context.Context, q sqlx.QueryerContext, loanID string) (int64, error) {
	var v int64
	err := sqlx.GetContext(ctx, q, &v, `SELECT coalesce(sum(unapplied), 0) FROM servicing.operations WHERE loan_id = $1`, loanID)
	return v, err
}

// quoteRow — сохранённый расчёт досрочного погашения.
type quoteRow struct {
	ID              string                `db:"id"`
	LoanID          string                `db:"loan_id"`
	Type            string                `db:"type"`
	Strategy        *string               `db:"strategy"`
	AmountToDebit   int64                 `db:"amount_to_debit"`
	AllocPrincipal  int64                 `db:"alloc_principal"`
	AllocInterest   int64                 `db:"alloc_interest"`
	InterestCredit  int64                 `db:"interest_credit"` // переплаченные проценты, засчитанные в основной долг
	MonthlyAfter    int64                 `db:"monthly_after"`
	Preview         db.JSON[[]previewRow] `db:"preview"`
	LoanVersion     int                   `db:"loan_version"`
	ScheduleVersion int                   `db:"schedule_version"`
	EffectiveDate   time.Time             `db:"effective_date"`
	CalculatedAt    time.Time             `db:"calculated_at"`
	ValidUntil      time.Time             `db:"valid_until"`
	UsedAt          *time.Time            `db:"used_at"`
	IdemKey         string                `db:"idem_key"`
	Fingerprint     string                `db:"fingerprint"`
}

// previewRow — строка нового графика в расчёте: применяется как есть после подтверждённого списания.
type previewRow struct {
	N         int    `json:"n"`
	Date      string `json:"date"`
	Principal int64  `json:"principal"`
	Interest  int64  `json:"interest"`
}

const quoteCols = `id, loan_id, type, strategy, amount_to_debit, alloc_principal, alloc_interest, interest_credit, monthly_after, preview,
	loan_version, schedule_version, effective_date, calculated_at, valid_until, used_at, idem_key, fingerprint`

func (s *Store) InsertQuote(ctx context.Context, tx *sqlx.Tx, q *quoteRow) error {
	return tx.GetContext(ctx, &q.ID, `
		INSERT INTO servicing.quotes (id, loan_id, type, strategy, amount_to_debit, alloc_principal, alloc_interest,
			interest_credit, monthly_after, preview, loan_version, schedule_version, effective_date, calculated_at, valid_until,
			idem_key, fingerprint)
		VALUES ('eq-' || nextval('servicing.quote_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`,
		q.LoanID, q.Type, q.Strategy, q.AmountToDebit, q.AllocPrincipal, q.AllocInterest, q.InterestCredit, q.MonthlyAfter, q.Preview,
		q.LoanVersion, q.ScheduleVersion, q.EffectiveDate, q.CalculatedAt, q.ValidUntil, q.IdemKey, q.Fingerprint)
}

func getQuote(ctx context.Context, q sqlx.QueryerContext, query string, args ...any) (*quoteRow, error) {
	var r quoteRow
	err := sqlx.GetContext(ctx, q, &r, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// QuoteByKey — расчёт кредита по ключу идемпотентности; (nil, nil) — не было.
func (s *Store) QuoteByKey(ctx context.Context, q sqlx.QueryerContext, loanID, key string) (*quoteRow, error) {
	return getQuote(ctx, q, `SELECT `+quoteCols+` FROM servicing.quotes WHERE loan_id = $1 AND idem_key = $2`, loanID, key)
}

// Quote — расчёт по id; (nil, nil) — расчёта нет.
func (s *Store) Quote(ctx context.Context, q sqlx.QueryerContext, id string) (*quoteRow, error) {
	return getQuote(ctx, q, `SELECT `+quoteCols+` FROM servicing.quotes WHERE id = $1`, id)
}

// UseQuote отмечает расчёт использованным.
func (s *Store) UseQuote(ctx context.Context, tx *sqlx.Tx, id string, at time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE servicing.quotes SET used_at = $2 WHERE id = $1`, id, at)
	return err
}

// CompanyLoansPage — страница кредитов компании и их общее число с учётом фильтра; пустой status — все.
func (s *Store) CompanyLoansPage(ctx context.Context, q sqlx.QueryerContext, companyID, status string, limit, offset int) ([]loanRow, int, error) {
	const where = ` WHERE company_id = $1 AND ($2 = '' OR status = $2)`
	var total int
	if err := sqlx.GetContext(ctx, q, &total, `SELECT count(*) FROM servicing.loans`+where, companyID, status); err != nil {
		return nil, 0, err
	}
	var out []loanRow
	err := sqlx.SelectContext(ctx, q, &out, `SELECT `+loanCols+` FROM servicing.loans`+where+`
		ORDER BY issued_at DESC, id DESC LIMIT $3 OFFSET $4`, companyID, status, limit, offset)
	return out, total, err
}

// AllCompanyLoans — все кредиты компании, без ограничения: для сводки по всей компании.
func (s *Store) AllCompanyLoans(ctx context.Context, q sqlx.QueryerContext, companyID string) ([]loanRow, error) {
	var out []loanRow
	err := sqlx.SelectContext(ctx, q, &out, `SELECT `+loanCols+` FROM servicing.loans
		WHERE company_id = $1 ORDER BY issued_at DESC, id DESC`, companyID)
	return out, err
}
