package support

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type faqRow struct {
	ID       string `db:"id"`
	Step     string `db:"step"`
	Question string `db:"question"`
	Answer   string `db:"answer"`
}

// FAQ — частые вопросы шага пути по порядку; step = "" — все шаги.
func (s *Store) FAQ(ctx context.Context, step string) ([]faqRow, error) {
	var out []faqRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT id, step, question, answer FROM support.faq WHERE $1 = '' OR step = $1 ORDER BY step, sort, id`, step)
	return out, err
}
