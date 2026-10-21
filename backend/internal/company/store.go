package company

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type Store struct{ db *sqlx.DB }

type companyRow struct {
	ID   string `db:"id"`
	INN  string `db:"inn"`
	Name string `db:"name"`
	Type string `db:"type"`
}

// GetByID — компания по id; (nil, nil) — компании нет.
func (s *Store) GetByID(ctx context.Context, id string) (*shared.CompanyView, error) {
	var r companyRow
	err := s.db.GetContext(ctx, &r, `SELECT id, inn, name, type FROM company.companies WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shared.CompanyView{ID: r.ID, INN: r.INN, Name: r.Name, Type: shared.CompanyType(r.Type)}, nil
}

type profileRow struct {
	Profile   db.JSON[shared.CompanyProfile] `db:"profile"`
	FetchedAt time.Time                      `db:"fetched_at"`
}

// CachedProfile — профиль из кэша и время, когда его собрали; (nil, …, nil) — в кэше нет.
func (s *Store) CachedProfile(ctx context.Context, companyID string) (*shared.CompanyProfile, time.Time, error) {
	var r profileRow
	err := s.db.GetContext(ctx, &r, `SELECT profile, fetched_at FROM company.profiles WHERE company_id = $1`, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	return &r.Profile.V, r.FetchedAt, nil
}

func (s *Store) SaveProfile(ctx context.Context, p *shared.CompanyProfile, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO company.profiles (company_id, profile, fetched_at) VALUES ($1, $2, $3)
		ON CONFLICT (company_id) DO UPDATE SET profile = EXCLUDED.profile, fetched_at = EXCLUDED.fetched_at`,
		p.ID, db.JSON[shared.CompanyProfile]{V: *p}, at)
	return err
}

func (s *Store) ResetProfiles(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE company.profiles`)
	return err
}
