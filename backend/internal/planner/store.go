package planner

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type planRow struct {
	ID          string     `db:"id"`
	CompanyID   string     `db:"company_id"`
	CreatedBy   string     `db:"created_by"`
	Title       string     `db:"title"`
	Note        *string    `db:"note"`
	Goal        string     `db:"goal"`
	BudgetRub   int64      `db:"budget_rub"`
	DateMode    string     `db:"date_mode"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	State       string     `db:"state"`
	NotifyInApp bool       `db:"notify_in_app"`
	NotifyEmail bool       `db:"notify_email"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

const planCols = `id, company_id, created_by, title, note, goal, budget_rub, date_mode, start_date, end_date, state,
	notify_in_app, notify_email, created_at, updated_at`

// maxPlans — сколько планов отдаёт список: у компании их единицы, страницы в контракте нет.
const maxPlans = 200

// InsertPlan присваивает плану номер p-N.
func (s *Store) InsertPlan(ctx context.Context, p *planRow) error {
	return s.insertPlan(ctx, s.db, p)
}

func (s *Store) insertPlan(ctx context.Context, q sqlx.QueryerContext, p *planRow) error {
	return sqlx.GetContext(ctx, q, &p.ID, `
		INSERT INTO planner.plans (id, company_id, created_by, title, note, goal, budget_rub, date_mode, start_date, end_date,
			state, notify_in_app, notify_email, created_at, updated_at)
		VALUES ('p-' || nextval('planner.plan_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
		RETURNING id`,
		p.CompanyID, p.CreatedBy, p.Title, p.Note, p.Goal, p.BudgetRub, p.DateMode, p.StartDate, p.EndDate,
		p.State, p.NotifyInApp, p.NotifyEmail, p.CreatedAt)
}

// Plan — план по id; (nil, nil) — плана нет.
func (s *Store) Plan(ctx context.Context, id string) (*planRow, error) {
	var p planRow
	err := s.db.GetContext(ctx, &p, `SELECT `+planCols+` FROM planner.plans WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Plans — планы компании; state пустой — все, кроме архивных.
func (s *Store) Plans(ctx context.Context, companyID, state string) ([]planRow, error) {
	var rows []planRow
	err := s.db.SelectContext(ctx, &rows, `SELECT `+planCols+` FROM planner.plans
		WHERE company_id = $1 AND (($2 = '' AND state <> 'ARCHIVED') OR state = $2)
		ORDER BY created_at DESC, id LIMIT $3`, companyID, state, maxPlans)
	return rows, err
}

func (s *Store) UpdatePlan(ctx context.Context, p *planRow) error {
	return s.updatePlan(ctx, s.db, p)
}

func (s *Store) updatePlan(ctx context.Context, q sqlx.ExecerContext, p *planRow) error {
	_, err := q.ExecContext(ctx, `
		UPDATE planner.plans SET title = $2, note = $3, goal = $4, budget_rub = $5, date_mode = $6, start_date = $7,
			end_date = $8, state = $9, notify_in_app = $10, notify_email = $11, updated_at = $12
		WHERE id = $1`,
		p.ID, p.Title, p.Note, p.Goal, p.BudgetRub, p.DateMode, p.StartDate, p.EndDate, p.State, p.NotifyInApp, p.NotifyEmail, p.UpdatedAt)
	return err
}

func (s *Store) DeletePlan(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM planner.plans WHERE id = $1`, id)
	return err
}

// RecordMatch запоминает предложенный вариант; false — по этому плану его уже предлагали.
func (s *Store) RecordMatch(ctx context.Context, planID, productCode string, at time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO planner.plan_matches (plan_id, product_code, notified_at) VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, planID, productCode, at)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE planner.plans, planner.plan_matches, planner.assistant_commands`)
	return err
}
