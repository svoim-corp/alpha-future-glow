package scoring

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

// prelimCheck — проверка в том виде, в каком её видит клиент (API_SPEC.md, 4.5, ответ prelim).
type prelimCheck struct {
	Source shared.CheckSource `json:"source" example:"EGRUL"`
	Status shared.CheckStatus `json:"status" example:"OK"`
	Label  string             `json:"label" example:"Действует, 11 лет"`
}

type prelimRow struct {
	ID            int64                  `db:"id"`
	AppID         string                 `db:"app_id"`
	Outcome       string                 `db:"outcome"`
	LimitAmount   int64                  `db:"limit_amount"`
	PrelimRate    float64                `db:"prelim_rate"`
	PrelimPayment int64                  `db:"prelim_payment"`
	AmountFits    bool                   `db:"amount_fits"`
	StopReason    *string                `db:"stop_reason"`
	Choice        *string                `db:"choice"`
	ProductCode   string                 `db:"product_code"`
	Amount        int64                  `db:"amount"`
	Term          int                    `db:"term"`
	Checks        db.JSON[[]prelimCheck] `db:"checks"`
	CreatedAt     time.Time              `db:"created_at"`
}

const prelimCols = `id, app_id, outcome, limit_amount, prelim_rate, prelim_payment, amount_fits, stop_reason, choice,
	product_code, amount, term, checks, created_at`

func (s *Store) InsertPrelim(ctx context.Context, r *prelimRow) error {
	return s.db.GetContext(ctx, &r.ID, `
		INSERT INTO scoring.prelims (app_id, outcome, limit_amount, prelim_rate, prelim_payment, amount_fits, stop_reason,
			product_code, amount, term, checks, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		r.AppID, r.Outcome, r.LimitAmount, r.PrelimRate, r.PrelimPayment, r.AmountFits, r.StopReason,
		r.ProductCode, r.Amount, r.Term, r.Checks, r.CreatedAt)
}

// LatestPrelim — последний результат этапа 1; (nil, nil) — этап 1 не проходили.
func (s *Store) LatestPrelim(ctx context.Context, appID string) (*prelimRow, error) {
	var r prelimRow
	err := s.db.GetContext(ctx, &r, `SELECT `+prelimCols+` FROM scoring.prelims
		WHERE app_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// SetChoice фиксирует выбор клиента; при уменьшении суммы результат пересчитан на новую сумму, и она в лимите.
func (s *Store) SetChoice(ctx context.Context, id int64, choice shared.PrelimChoice, amount int64, amountFits bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE scoring.prelims SET choice = $2, amount = $3, amount_fits = $4 WHERE id = $1`,
		id, choice, amount, amountFits)
	return err
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE scoring.prelims, scoring.score_results, scoring.decisions, scoring.jobs`)
	return err
}

func (r *prelimRow) view() *shared.PrelimView {
	v := &shared.PrelimView{
		Outcome: shared.PrelimOutcome(r.Outcome), Limit: r.LimitAmount, PrelimRate: r.PrelimRate, AmountFits: r.AmountFits,
		StopReason: r.StopReason, ProductCode: r.ProductCode, Amount: r.Amount, Term: r.Term,
	}
	if r.Choice != nil {
		c := shared.PrelimChoice(*r.Choice)
		v.Choice = &c
	}
	return v
}

type scoreRow struct {
	ID           int64                             `db:"id"`
	AppID        string                            `db:"app_id"`
	Score        int                               `db:"score"`
	RiskBand     string                            `db:"risk_band"`
	Factors      db.JSON[[]shared.ScoreFactor]     `db:"factors"`
	AutoEligible bool                              `db:"auto_eligible"`
	AutoReason   string                            `db:"auto_reason"`
	Suggested    db.JSON[shared.SuggestedDecision] `db:"suggested"`
	ModelVersion string                            `db:"model_version"`
	CreatedAt    time.Time                         `db:"created_at"`
}

func (s *Store) InsertScore(ctx context.Context, r *scoreRow) error {
	return s.db.GetContext(ctx, &r.ID, `
		INSERT INTO scoring.score_results (app_id, score, risk_band, factors, auto_eligible, auto_reason, suggested, model_version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		r.AppID, r.Score, r.RiskBand, r.Factors, r.AutoEligible, r.AutoReason, r.Suggested, r.ModelVersion, r.CreatedAt)
}

// LatestScore — последний скоринг; (nil, nil) — скоринга не было.
func (s *Store) LatestScore(ctx context.Context, appID string) (*scoreRow, error) {
	var r scoreRow
	err := s.db.GetContext(ctx, &r, `
		SELECT id, app_id, score, risk_band, factors, auto_eligible, auto_reason, suggested, model_version, created_at
		FROM scoring.score_results WHERE app_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type decisionRow struct {
	ID           string                        `db:"id"`
	AppID        string                        `db:"app_id"`
	Result       string                        `db:"result"`
	Mode         string                        `db:"mode"`
	ProductCode  string                        `db:"product_code"`
	Amount       int64                         `db:"amount"`
	Term         int                           `db:"term"`
	Rate         float64                       `db:"rate"`
	Reason       string                        `db:"reason"`
	Alternatives db.JSON[[]shared.Alternative] `db:"alternatives"`
	DecidedBy    *string                       `db:"decided_by"`
	DecidedAt    time.Time                     `db:"decided_at"`
}

const decisionCols = `id, app_id, result, mode, product_code, amount, term, rate, reason, alternatives, decided_by, decided_at`

// InsertDecision присваивает решению номер dec-N.
func (s *Store) InsertDecision(ctx context.Context, r *decisionRow) error {
	return s.db.GetContext(ctx, &r.ID, `
		INSERT INTO scoring.decisions (id, app_id, result, mode, product_code, amount, term, rate, reason, alternatives, decided_by, decided_at)
		VALUES ('dec-' || nextval('scoring.decision_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`,
		r.AppID, r.Result, r.Mode, r.ProductCode, r.Amount, r.Term, r.Rate, r.Reason, r.Alternatives, r.DecidedBy, r.DecidedAt)
}

func (s *Store) DeleteDecision(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM scoring.decisions WHERE id = $1`, id)
	return err
}

// LatestDecision — действующее решение; (nil, nil) — решения нет.
func (s *Store) LatestDecision(ctx context.Context, appID string) (*decisionRow, error) {
	var r decisionRow
	err := s.db.GetContext(ctx, &r, `SELECT `+decisionCols+` FROM scoring.decisions
		WHERE app_id = $1 ORDER BY decided_at DESC, id DESC LIMIT 1`, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// EnqueueJob ставит заявку в конвейер; повторная постановка сбрасывает счётчик попыток.
func (s *Store) EnqueueJob(ctx context.Context, appID string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scoring.jobs (app_id, state, attempts, last_error, updated_at) VALUES ($1, 'PENDING', 0, NULL, $2)
		ON CONFLICT (app_id) DO UPDATE SET state = 'PENDING', attempts = 0, last_error = NULL, updated_at = EXCLUDED.updated_at`,
		appID, at)
	return err
}

type jobRow struct {
	AppID    string `db:"app_id"`
	Attempts int    `db:"attempts"`
}

func (s *Store) PendingJobs(ctx context.Context, limit int) ([]jobRow, error) {
	var rows []jobRow
	err := s.db.SelectContext(ctx, &rows, `SELECT app_id, attempts FROM scoring.jobs
		WHERE state = 'PENDING' ORDER BY updated_at, app_id LIMIT $1`, limit)
	return rows, err
}

func (s *Store) FinishJob(ctx context.Context, appID string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE scoring.jobs SET state = 'DONE', last_error = NULL, updated_at = $2 WHERE app_id = $1`, appID, at)
	return err
}

// RetryJob записывает неудачную попытку; после maxAttempts задача остаётся FAILED для разбора.
func (s *Store) RetryJob(ctx context.Context, appID, cause string, maxAttempts int, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE scoring.jobs SET attempts = attempts + 1, last_error = $2, updated_at = $4,
			state = CASE WHEN attempts + 1 >= $3 THEN 'FAILED' ELSE 'PENDING' END
		WHERE app_id = $1`, appID, cause, maxAttempts, at)
	return err
}
