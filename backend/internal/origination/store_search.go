package origination

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/searchtext"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func (s *Store) search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery, products, goals string) ([]shared.SearchHit, error) {
	prefix, contains := searchtext.Patterns(query.Text)
	hits := []shared.SearchHit{}
	err := s.db.SelectContext(ctx, &hits, `WITH documents AS (
    SELECT id AS id, coalesce(number, 'Черновик ' || id) AS title, left(coalesce($8::jsonb ->> product_code, product_code) || ' · ' || coalesce(purpose_text, '') || ' ' || coalesce($9::jsonb ->> goal, goal, ''), 240) AS snippet,
           status AS state, '' AS step,
           replace(lower(id || ' ' || coalesce(number, '') || ' ' || product_code || ' ' || coalesce(goal, '') || ' ' || coalesce(purpose_text, '')), 'ё', 'е') AS document,
           CASE WHEN id = $3 OR replace(lower(coalesce(number, '')), 'ё', 'е') = $3 THEN 4
                WHEN replace(lower(coalesce($8::jsonb ->> product_code, product_code)), 'ё', 'е') = $3 THEN 3
                WHEN replace(lower(coalesce($8::jsonb ->> product_code, product_code)), 'ё', 'е') LIKE $4 THEN 2 ELSE 1 END AS score
    FROM origination.applications
    WHERE company_id = $1 AND deleted_at IS NULL
      AND (replace(lower(id || ' ' || coalesce(number, '') || ' ' || product_code || ' ' || coalesce(goal, '') || ' ' || coalesce(purpose_text, '')), 'ё', 'е') LIKE $5
           OR replace(lower(coalesce(number, 'Черновик ' || id)), 'ё', 'е') LIKE $5
           OR to_tsvector('russian', replace(lower(id || ' ' || coalesce(number, '') || ' ' || product_code || ' ' || coalesce(goal, '') || ' ' || coalesce(purpose_text, '')), 'ё', 'е')) @@ plainto_tsquery('russian', $3) OR replace(lower(coalesce($8::jsonb ->> product_code, '') || ' ' || coalesce($9::jsonb ->> goal, '')), 'ё', 'е') LIKE $5 OR to_tsvector('russian', replace(lower(coalesce($8::jsonb ->> product_code, '') || ' ' || coalesce($9::jsonb ->> goal, '')), 'ё', 'е')) @@ plainto_tsquery('russian', $3))
)
SELECT id, title, snippet, state, step, score FROM documents
WHERE ($6::integer = 0 OR score < $6 OR (score = $6 AND id > $7))
ORDER BY score DESC, id ASC LIMIT $2`, scope.CompanyID, query.Limit, query.Text, prefix, contains, query.After.Score, query.After.ID, products, goals)
	return hits, err
}
