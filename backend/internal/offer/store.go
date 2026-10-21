package offer

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

type offerRow struct {
	AppID                string                        `db:"app_id"`
	DecisionID           string                        `db:"decision_id"`
	Decision             string                        `db:"decision"`
	ProductCode          string                        `db:"product_code"`
	Amount               int64                         `db:"amount"`
	Term                 int                           `db:"term"`
	Rate                 float64                       `db:"rate"`
	GraceMonths          *int                          `db:"grace_months"`
	MonthlyPayment       int64                         `db:"monthly_payment"`
	RequestedProductCode string                        `db:"requested_product_code"`
	RequestedAmount      int64                         `db:"requested_amount"`
	RequestedTerm        int                           `db:"requested_term"`
	Explanation          *string                       `db:"explanation"`
	Alternatives         db.JSON[[]shared.Alternative] `db:"alternatives"`
	ValidUntil           time.Time                     `db:"valid_until"`
	CreatedAt            time.Time                     `db:"created_at"`
}

const offerCols = `app_id, decision_id, decision, product_code, amount, term, rate, grace_months, monthly_payment,
	requested_product_code, requested_amount, requested_term, explanation, alternatives, valid_until, created_at`

// SaveOffer записывает предложение по заявке; новое решение заменяет прежнее предложение.
func (s *Store) SaveOffer(ctx context.Context, o *offerRow) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO offer.offers (`+offerCols+`)
		VALUES (:app_id, :decision_id, :decision, :product_code, :amount, :term, :rate, :grace_months, :monthly_payment,
			:requested_product_code, :requested_amount, :requested_term, :explanation, :alternatives, :valid_until, :created_at)
		ON CONFLICT (app_id) DO UPDATE SET decision_id = EXCLUDED.decision_id, decision = EXCLUDED.decision,
			product_code = EXCLUDED.product_code, amount = EXCLUDED.amount, term = EXCLUDED.term, rate = EXCLUDED.rate,
			grace_months = EXCLUDED.grace_months, monthly_payment = EXCLUDED.monthly_payment,
			requested_product_code = EXCLUDED.requested_product_code, requested_amount = EXCLUDED.requested_amount,
			requested_term = EXCLUDED.requested_term, explanation = EXCLUDED.explanation,
			alternatives = EXCLUDED.alternatives, valid_until = EXCLUDED.valid_until, created_at = EXCLUDED.created_at`, o)
	return err
}

// Offer — предложение по заявке; (nil, nil) — решения ещё нет.
func (s *Store) Offer(ctx context.Context, appID string) (*offerRow, error) {
	var o offerRow
	err := s.db.GetContext(ctx, &o, `SELECT `+offerCols+` FROM offer.offers WHERE app_id = $1`, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE offer.offers`)
	return err
}
