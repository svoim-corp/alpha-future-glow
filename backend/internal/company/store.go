package company

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type Store struct{ db *sqlx.DB }

type companyRow struct {
	ID           string `db:"id"`
	INN          string `db:"inn"`
	Name         string `db:"name"`
	Type         string `db:"type"`
	IsBankClient bool   `db:"is_bank_client"`
}

// GetByID реализует shared.Companies; (nil, nil) — компании нет.
func (s *Store) GetByID(ctx context.Context, id string) (*shared.CompanyView, error) {
	var r companyRow
	err := s.db.GetContext(ctx, &r, `SELECT id, inn, name, type, is_bank_client FROM company.companies WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shared.CompanyView{ID: r.ID, INN: r.INN, Name: r.Name, Type: r.Type, IsBankClient: r.IsBankClient}, nil
}
