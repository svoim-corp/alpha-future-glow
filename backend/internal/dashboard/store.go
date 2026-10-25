package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
)

type Store struct{ db *sqlx.DB }

type dashboardRow struct {
	UserID    string          `db:"user_id"`
	CompanyID string          `db:"company_id"`
	Revision  string          `db:"revision"`
	Config    db.JSON[config] `db:"config"`
	UpdatedAt time.Time       `db:"updated_at"`
}

// Dashboard — настройка пользователя в компании; (nil, nil) — ещё не сохранялась.
func (s *Store) Dashboard(ctx context.Context, userID, companyID string) (*dashboardRow, error) {
	var d dashboardRow
	err := s.db.GetContext(ctx, &d, `
		SELECT user_id, company_id, revision, config, updated_at FROM dashboard.dashboards
		WHERE user_id = $1 AND company_id = $2`, userID, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// InsertDashboard — первое сохранение; false — запись уже появилась.
func (s *Store) InsertDashboard(ctx context.Context, q sqlx.ExtContext, d *dashboardRow) (bool, error) {
	res, err := q.ExecContext(ctx, `
		INSERT INTO dashboard.dashboards (user_id, company_id, revision, config, updated_at)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
		d.UserID, d.CompanyID, d.Revision, d.Config, d.UpdatedAt)
	return affected(res, err)
}

// UpdateDashboard заменяет документ, если ревизия всё ещё expected; false — не совпала или записи нет.
func (s *Store) UpdateDashboard(ctx context.Context, q sqlx.ExtContext, d *dashboardRow, expected string) (bool, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE dashboard.dashboards SET revision = $3, config = $4, updated_at = $5
		WHERE user_id = $1 AND company_id = $2 AND revision = $6`,
		d.UserID, d.CompanyID, d.Revision, d.Config, d.UpdatedAt, expected)
	return affected(res, err)
}

type saveKeyRow struct {
	UserID    string          `db:"user_id"`
	CompanyID string          `db:"company_id"`
	Key       string          `db:"key"`
	BodyHash  string          `db:"body_hash"`
	Revision  string          `db:"revision"`
	Config    db.JSON[config] `db:"config"`
	UpdatedAt time.Time       `db:"updated_at"`
}

// SaveKey — успешное сохранение по ключу; (nil, nil) — ключ не использовался.
func (s *Store) SaveKey(ctx context.Context, userID, companyID, key string) (*saveKeyRow, error) {
	var k saveKeyRow
	err := s.db.GetContext(ctx, &k, `
		SELECT user_id, company_id, key, body_hash, revision, config, updated_at FROM dashboard.save_keys
		WHERE user_id = $1 AND company_id = $2 AND key = $3`, userID, companyID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// InsertSaveKey запоминает результат сохранения; false — ключ уже занят параллельным запросом.
func (s *Store) InsertSaveKey(ctx context.Context, q sqlx.ExtContext, k *saveKeyRow) (bool, error) {
	res, err := q.ExecContext(ctx, `
		INSERT INTO dashboard.save_keys (user_id, company_id, key, body_hash, revision, config, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
		k.UserID, k.CompanyID, k.Key, k.BodyHash, k.Revision, k.Config, k.UpdatedAt)
	return affected(res, err)
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE dashboard.dashboards, dashboard.save_keys`)
	return err
}

func affected(res sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}
