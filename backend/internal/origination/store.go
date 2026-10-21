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
func (s *Store) UpdateDraft(ctx context.Context, q sqlx.ExtContext, a *appRow, version int, at time.Time) (bool, error) {
	err := sqlx.GetContext(ctx, q, &a.Version, `
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

// DeleteGuarantorConsent убирает согласие поручителя: оно дано конкретным человеком и при смене поручителя не действует.
func (s *Store) DeleteGuarantorConsent(ctx context.Context, q sqlx.ExtContext, appID string) error {
	_, err := q.ExecContext(ctx, `DELETE FROM origination.consents WHERE app_id = $1 AND subject = 'GUARANTOR'`, appID)
	return err
}

// LockApplication — заявка под блокировкой строки до конца транзакции; (nil, nil) — заявки нет.
func (s *Store) LockApplication(ctx context.Context, q sqlx.ExtContext, id string) (*appRow, error) {
	var a appRow
	err := sqlx.GetContext(ctx, q, &a, `SELECT `+appCols+` FROM origination.applications WHERE id = $1 FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SaveTransition записывает статус и поля, которые меняются вместе с ним, и увеличивает версию (в a — новая).
func (s *Store) SaveTransition(ctx context.Context, q sqlx.ExtContext, a *appRow) error {
	return sqlx.GetContext(ctx, q, &a.Version, `
		UPDATE origination.applications SET status = $2, number = $3, submitted_at = $4, truth_confirmed = $5,
			deleted_at = $6, updated_at = $7, version = version + 1
		WHERE id = $1
		RETURNING version`,
		a.ID, a.Status, a.Number, a.SubmittedAt, a.TruthConfirmed, a.DeletedAt, a.UpdatedAt)
}

type eventRow struct {
	AppID     string    `db:"app_id"`
	From      string    `db:"from_status"`
	To        string    `db:"to_status"`
	Action    string    `db:"action"`
	ActorType string    `db:"actor_type"`
	ActorID   *string   `db:"actor_id"`
	ActorName *string   `db:"actor_name"`
	Comment   *string   `db:"comment"`
	At        time.Time `db:"at"`
}

func (s *Store) InsertEvent(ctx context.Context, q sqlx.ExtContext, e *eventRow) error {
	_, err := sqlx.NamedExecContext(ctx, q, `
		INSERT INTO origination.status_events (app_id, from_status, to_status, action, actor_type, actor_id, actor_name, comment, at)
		VALUES (:app_id, :from_status, :to_status, :action, :actor_type, :actor_id, :actor_name, :comment, :at)`, e)
	return err
}

// Events — история заявки по порядку.
func (s *Store) Events(ctx context.Context, appID string) ([]eventRow, error) {
	var out []eventRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT app_id, from_status, to_status, action, actor_type, actor_id, actor_name, comment, at
		FROM origination.status_events WHERE app_id = $1 ORDER BY at, id`, appID)
	return out, err
}

type consentRow struct {
	ID       string    `db:"id"`
	AppID    string    `db:"app_id"`
	Subject  string    `db:"subject"`
	PersonID *string   `db:"person_id"`
	Text     string    `db:"text"`
	GivenBy  string    `db:"given_by"`
	GivenAt  time.Time `db:"given_at"`
}

const consentCols = `id, app_id, subject, person_id, text, given_by, given_at`

// SaveConsent сохраняет согласие, если по этому субъекту его ещё нет. В c — согласие, которое действует.
func (s *Store) SaveConsent(ctx context.Context, c *consentRow) error {
	return s.db.GetContext(ctx, c, `
		WITH ins AS (
			INSERT INTO origination.consents (id, app_id, subject, person_id, text, given_by, given_at)
			VALUES ('cs-' || nextval('origination.consent_seq'), $1, $2, $3, $4, $5, $6)
			ON CONFLICT (app_id, subject) DO NOTHING
			RETURNING `+consentCols+`)
		SELECT `+consentCols+` FROM ins
		UNION ALL
		SELECT `+consentCols+` FROM origination.consents WHERE app_id = $1 AND subject = $2 AND NOT EXISTS (SELECT 1 FROM ins)`,
		c.AppID, c.Subject, c.PersonID, c.Text, c.GivenBy, c.GivenAt)
}

// Consent — согласие по заявке от субъекта; (nil, nil) — согласия нет.
func (s *Store) Consent(ctx context.Context, appID string, subject shared.ConsentSubject) (*consentRow, error) {
	var c consentRow
	err := s.db.GetContext(ctx, &c, `SELECT `+consentCols+` FROM origination.consents WHERE app_id = $1 AND subject = $2`, appID, subject)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SetDraftAmount меняет сумму черновика и увеличивает версию; false — заявки нет или это не черновик.
func (s *Store) SetDraftAmount(ctx context.Context, id string, amount int64, at time.Time) (int, bool, error) {
	var version int
	err := s.db.GetContext(ctx, &version, `
		UPDATE origination.applications SET amount = $2, version = version + 1, updated_at = $3
		WHERE id = $1 AND status = 'DRAFT'
		RETURNING version`, id, amount, at)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return version, err == nil, err
}

// NextNumber — порядковый номер для номера заявки. Пропуски допустимы: номер берётся до перехода.
func (s *Store) NextNumber(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.GetContext(ctx, &n, `SELECT nextval('origination.number_seq')`)
	return n, err
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
