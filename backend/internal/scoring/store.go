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
	_, err := s.db.ExecContext(ctx, `TRUNCATE scoring.prelims`)
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
