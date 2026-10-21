package catalog

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
)

type Store struct{ db *sqlx.DB }

type productRow struct {
	Code                 string            `db:"code"`
	Name                 string            `db:"name"`
	ShortText            string            `db:"short_text"`
	Description          string            `db:"description"`
	Goals                db.JSON[[]string] `db:"goals"`
	AmountMin            int64             `db:"amount_min"`
	AmountMax            int64             `db:"amount_max"`
	TermMin              int               `db:"term_min"`
	TermMax              int               `db:"term_max"`
	RateMin              float64           `db:"rate_min"`
	RateMax              float64           `db:"rate_max"`
	GraceMaxMonths       int               `db:"grace_max_months"`
	MinBusinessAgeMonths int               `db:"min_business_age_months"`
	GuarantorForLLC      bool              `db:"guarantor_for_llc"`
	Revolving            bool              `db:"revolving"`
	LimitRule            string            `db:"limit_rule"`
	WhatYouNeed          db.JSON[[]string] `db:"what_you_need"`
}

const productCols = `code, name, short_text, description, goals, amount_min, amount_max, term_min, term_max,
	rate_min, rate_max, grace_max_months, min_business_age_months, guarantor_for_llc, revolving, limit_rule, what_you_need`

// Products — вся линейка в порядке показа в каталоге. Продуктов единицы, поэтому без пагинации.
func (s *Store) Products(ctx context.Context) ([]productRow, error) {
	var out []productRow
	err := s.db.SelectContext(ctx, &out, `SELECT `+productCols+` FROM catalog.products ORDER BY sort, code`)
	return out, err
}
