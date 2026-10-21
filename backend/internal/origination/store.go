package origination

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type Store struct{ db *sqlx.DB }

type appRow struct {
	ID                string                   `db:"id"`
	Number            *string                  `db:"number"`
	CompanyID         string                   `db:"company_id"`
	ApplicantID       string                   `db:"applicant_id"`
	ProductCode       string                   `db:"product_code"`
	Goal              *string                  `db:"goal"`
	PurposeText       *string                  `db:"purpose_text"`
	Amount            int64                    `db:"amount"`
	Term              int                      `db:"term"`
	GraceMonths       *int                     `db:"grace_months"`
	OverdraftLimit    *int64                   `db:"overdraft_limit"`
	GuarantorPersonID *string                  `db:"guarantor_person_id"`
	TruthConfirmed    bool                     `db:"truth_confirmed"`
	Authority         db.JSON[authorityRecord] `db:"authority"`
	Status            string                   `db:"status"`
	Version           int                      `db:"version"`
	CreatedAt         time.Time                `db:"created_at"`
	UpdatedAt         time.Time                `db:"updated_at"`
	SubmittedAt       *time.Time               `db:"submitted_at"`
	DeletedAt         *time.Time               `db:"deleted_at"`
}

// authorityRecord — снимок полномочий в jsonb. Свои json-теги, чтобы формат хранения не зависел от shared.
type authorityRecord struct {
	Type       shared.AuthorityType `json:"type"`
	MCHDNumber *string              `json:"mchdNumber,omitempty"`
	ValidUntil *time.Time           `json:"validUntil,omitempty"`
	Powers     []string             `json:"powers"`
}

func (s *Store) Insert(ctx context.Context, a *appRow) error {
	return s.db.GetContext(ctx, &a.ID, `
		INSERT INTO origination.applications (id, company_id, applicant_id, product_code, goal, amount, term, authority,
			status, version, created_at, updated_at)
		VALUES ('a-' || nextval('origination.application_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id`,
		a.CompanyID, a.ApplicantID, a.ProductCode, a.Goal, a.Amount, a.Term, a.Authority, a.Status, a.Version, a.CreatedAt)
}

// Reset удаляет все заявки; связанные таблицы схемы очищаются каскадом.
func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE origination.applications CASCADE`)
	return err
}
