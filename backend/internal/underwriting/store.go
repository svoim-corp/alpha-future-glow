package underwriting

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type queueRow struct {
	AppID        string     `db:"app_id"`
	Number       *string    `db:"number"`
	CompanyID    string     `db:"company_id"`
	CompanyINN   string     `db:"company_inn"`
	CompanyName  string     `db:"company_name"`
	ProductCode  string     `db:"product_code"`
	ProductName  string     `db:"product_name"`
	Amount       int64      `db:"amount"`
	Term         int        `db:"term"`
	Status       string     `db:"status"`
	Score        *int       `db:"score"`
	RiskBand     *string    `db:"risk_band"`
	DecisionMode *string    `db:"decision_mode"`
	AssigneeID   *string    `db:"assignee_id"`
	AssigneeName *string    `db:"assignee_name"`
	SubmittedAt  time.Time  `db:"submitted_at"`
	DecidedAt    *time.Time `db:"decided_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

const queueCols = `app_id, number, company_id, company_inn, company_name, product_code, product_name, amount, term, status,
	score, risk_band, decision_mode, assignee_id, assignee_name, submitted_at, decided_at, updated_at`

// UpsertQueue обновляет строку очереди. Назначение сотрудника не трогается, время решения ставится один раз.
// Пустой скоринг не затирает уже известный: событие о смене статуса может прийти раньше пересчёта.
func (s *Store) UpsertQueue(ctx context.Context, r *queueRow) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO underwriting.queue (app_id, number, company_id, company_inn, company_name, product_code, product_name,
			amount, term, status, score, risk_band, decision_mode, submitted_at, decided_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (app_id) DO UPDATE SET
			number = EXCLUDED.number, company_inn = EXCLUDED.company_inn, company_name = EXCLUDED.company_name,
			product_code = EXCLUDED.product_code, product_name = EXCLUDED.product_name,
			amount = EXCLUDED.amount, term = EXCLUDED.term, status = EXCLUDED.status,
			score = COALESCE(EXCLUDED.score, underwriting.queue.score),
			risk_band = COALESCE(EXCLUDED.risk_band, underwriting.queue.risk_band),
			decision_mode = COALESCE(EXCLUDED.decision_mode, underwriting.queue.decision_mode),
			decided_at = COALESCE(underwriting.queue.decided_at, EXCLUDED.decided_at),
			updated_at = EXCLUDED.updated_at`,
		r.AppID, r.Number, r.CompanyID, r.CompanyINN, r.CompanyName, r.ProductCode, r.ProductName,
		r.Amount, r.Term, r.Status, r.Score, r.RiskBand, r.DecisionMode, r.SubmittedAt, r.DecidedAt, r.UpdatedAt)
	return err
}

// QueueItem — строка очереди; (nil, nil) — заявки в очереди нет.
func (s *Store) QueueItem(ctx context.Context, appID string) (*queueRow, error) {
	var r queueRow
	err := s.db.GetContext(ctx, &r, `SELECT `+queueCols+` FROM underwriting.queue WHERE app_id = $1`, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type statusCount struct {
	Status string `db:"status"`
	N      int    `db:"n"`
}

// Stats — заявок по статусам и решений с начала дня since.
func (s *Store) Stats(ctx context.Context, since time.Time) (map[string]int, int, error) {
	var rows []statusCount
	if err := s.db.SelectContext(ctx, &rows, `SELECT status, count(*) AS n FROM underwriting.queue GROUP BY status`); err != nil {
		return nil, 0, err
	}
	byStatus := make(map[string]int, len(rows))
	for _, r := range rows {
		byStatus[r.Status] = r.N
	}
	var decided int
	err := s.db.GetContext(ctx, &decided, `SELECT count(*) FROM underwriting.queue WHERE decided_at >= $1`, since)
	return byStatus, decided, err
}

// queueFilter — фильтры таблицы очереди. Пустое поле — без фильтра; sort — ключ из sortColumns, «-» в начале — по убыванию.
type queueFilter struct {
	Status, Product, Risk, DecisionMode string
	Sort                                string
}

// sortColumns — белый список сортировок: имя колонки попадает в SQL, поэтому только из констант.
var sortColumns = map[string]string{
	"submittedAt": "submitted_at",
	"updatedAt":   "updated_at",
	"amount":      "amount",
	"score":       "score",
}

const defaultSort = "-submittedAt"

// sql собирает WHERE и ORDER BY. Значения фильтров идут плейсхолдерами, в текст запроса — только константы.
func (f queueFilter) sql() (where, order string, args []any) {
	var conds []string
	add := func(col, v string) {
		if v != "" {
			args = append(args, v)
			conds = append(conds, col+" = $"+strconv.Itoa(len(args)))
		}
	}
	add("status", f.Status)
	add("product_code", f.Product)
	add("risk_band", f.Risk)
	add("decision_mode", f.DecisionMode)
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	sort := f.Sort
	if sort == "" {
		sort = defaultSort
	}
	dir := "ASC"
	if strings.HasPrefix(sort, "-") {
		sort, dir = sort[1:], "DESC"
	}
	order = " ORDER BY " + sortColumns[sort] + " " + dir + " NULLS LAST, app_id"
	return where, order, args
}

func (s *Store) ListQueue(ctx context.Context, f queueFilter, page, size int) ([]queueRow, int, error) {
	where, order, args := f.sql()
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT count(*) FROM underwriting.queue`+where, args...); err != nil {
		return nil, 0, err
	}
	n := len(args)
	args = append(args, size, (page-1)*size)
	var rows []queueRow
	err := s.db.SelectContext(ctx, &rows, `SELECT `+queueCols+` FROM underwriting.queue`+where+order+
		` LIMIT $`+strconv.Itoa(n+1)+` OFFSET $`+strconv.Itoa(n+2), args...)
	return rows, total, err
}

// Assign назначает заявку сотруднику, если она свободна или уже его. false — заявку взял другой (или её нет).
func (s *Store) Assign(ctx context.Context, appID, userID, name string, at time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE underwriting.queue SET assignee_id = $2, assignee_name = $3, updated_at = $4
		WHERE app_id = $1 AND (assignee_id IS NULL OR assignee_id = $2)`, appID, userID, name, at)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE underwriting.queue`)
	return err
}
