package planner

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/searchtext"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func (s *Store) search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery, products, goals string) ([]shared.SearchHit, error) {
	prefix, contains := searchtext.Patterns(query.Text)
	hits := []shared.SearchHit{}
	err := s.db.SelectContext(ctx, &hits, `WITH documents AS (
    SELECT id AS id, title AS title, left(coalesce(note, '') || ' · ' || coalesce($9::jsonb ->> goal, goal), 240) AS snippet,
           state AS state, '' AS step,
           replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е') AS document,
           CASE WHEN id = $3 THEN 4
                WHEN replace(lower(title), 'ё', 'е') = $3 THEN 3
                WHEN replace(lower(title), 'ё', 'е') LIKE $4 THEN 2 ELSE 1 END AS score
    FROM planner.plans
    WHERE company_id = $1 AND $8::jsonb IS NOT NULL
      AND (replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е') LIKE $5
           OR to_tsvector('russian', replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е')) @@ plainto_tsquery('russian', $3) OR replace(lower(coalesce($9::jsonb ->> goal, '')), 'ё', 'е') LIKE $5 OR to_tsvector('russian', replace(lower(coalesce($9::jsonb ->> goal, '')), 'ё', 'е')) @@ plainto_tsquery('russian', $3))
)
SELECT id, title, snippet, state, step, score FROM documents
WHERE ($6::integer = 0 OR score < $6 OR (score = $6 AND id > $7))
ORDER BY score DESC, id ASC LIMIT $2`, scope.CompanyID, query.Limit, query.Text, prefix, contains, query.After.Score, query.After.ID, products, goals)
	return hits, err
}
