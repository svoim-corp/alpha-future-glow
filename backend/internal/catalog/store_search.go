package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/searchtext"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func (s *Store) search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {
	prefix, contains := searchtext.Patterns(query.Text)
	hits := []shared.SearchHit{}
	err := s.db.SelectContext(ctx, &hits, `WITH documents AS (
    SELECT code AS id, name AS title, left(short_text || ' ' || description, 240) AS snippet,
           '' AS state, '' AS step,
           replace(lower(code || ' ' || name || ' ' || short_text || ' ' || description), 'ё', 'е') AS document,
           CASE WHEN replace(lower(code), 'ё', 'е') = $3 THEN 4
                WHEN replace(lower(name), 'ё', 'е') = $3 THEN 3
                WHEN replace(lower(name), 'ё', 'е') LIKE $4 THEN 2 ELSE 1 END AS score
    FROM catalog.products
    WHERE $1 = ''
      AND (replace(lower(code || ' ' || name || ' ' || short_text || ' ' || description), 'ё', 'е') LIKE $5
           OR to_tsvector('russian', replace(lower(code || ' ' || name || ' ' || short_text || ' ' || description), 'ё', 'е')) @@ plainto_tsquery('russian', $3) OR EXISTS (SELECT 1 FROM catalog.goals AS g WHERE products.goals @> jsonb_build_array(g.code) AND (replace(lower(g.code || ' ' || g.name), 'ё', 'е') LIKE $5 OR to_tsvector('russian', replace(lower(g.code || ' ' || g.name), 'ё', 'е')) @@ plainto_tsquery('russian', $3))))
)
SELECT id, title, snippet, state, step, score FROM documents
WHERE ($6::integer = 0 OR score < $6 OR (score = $6 AND id > $7))
ORDER BY score DESC, id ASC LIMIT $2`, "", query.Limit, query.Text, prefix, contains, query.After.Score, query.After.ID)
	return hits, err
}
