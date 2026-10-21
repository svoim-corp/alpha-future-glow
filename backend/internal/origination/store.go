package origination

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

const appCols = `id, number, company_id, applicant_id, product_code, goal, purpose_text, amount, term, grace_months,
	overdraft_limit, guarantor_person_id, truth_confirmed, authority, status, version, created_at, updated_at, submitted_at, deleted_at`

func (s *Store) Insert(ctx context.Context, a *appRow) error {
	return s.db.GetContext(ctx, &a.ID, `
		INSERT INTO origination.applications (id, company_id, applicant_id, product_code, goal, amount, term, authority,
			status, version, created_at, updated_at)
		VALUES ('a-' || nextval('origination.application_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id`,
		a.CompanyID, a.ApplicantID, a.ProductCode, a.Goal, a.Amount, a.Term, a.Authority, a.Status, a.Version, a.CreatedAt)
}

// Application — заявка по id; (nil, nil) — заявки нет.
func (s *Store) Application(ctx context.Context, id string) (*appRow, error) {
	var a appRow
	err := s.db.GetContext(ctx, &a, `SELECT `+appCols+` FROM origination.applications WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateDraft сохраняет параметры черновика и увеличивает версию; false — версия уже другая или это не черновик.
// В a записываются новые версия и время сохранения.
func (s *Store) UpdateDraft(ctx context.Context, a *appRow, version int, at time.Time) (bool, error) {
	err := s.db.GetContext(ctx, &a.Version, `
		UPDATE origination.applications SET amount = $3, term = $4, goal = $5, purpose_text = $6, grace_months = $7,
			overdraft_limit = $8, guarantor_person_id = $9, version = version + 1, updated_at = $10
		WHERE id = $1 AND version = $2 AND status = 'DRAFT'
		RETURNING version`,
		a.ID, version, a.Amount, a.Term, a.Goal, a.PurposeText, a.GraceMonths, a.OverdraftLimit, a.GuarantorPersonID, at)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	a.UpdatedAt = at
	return true, nil
}

// CompanyApplications — заявки компании без удалённых черновиков, недавно изменённые первыми, и их общее число.
func (s *Store) CompanyApplications(ctx context.Context, companyID string, limit, offset int) ([]appRow, int, error) {
	var total int
	if err := s.db.GetContext(ctx, &total, `
		SELECT count(*) FROM origination.applications WHERE company_id = $1 AND deleted_at IS NULL`, companyID); err != nil {
		return nil, 0, err
	}
	var out []appRow
	err := s.db.SelectContext(ctx, &out, `SELECT `+appCols+` FROM origination.applications
		WHERE company_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC LIMIT $2 OFFSET $3`, companyID, limit, offset)
	return out, total, err
}

// Reset удаляет все заявки; связанные таблицы схемы очищаются каскадом.
func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE origination.applications CASCADE`)
	return err
}
