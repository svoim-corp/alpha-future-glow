package servicing

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/searchtext"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func (s *Store) search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {
	prefix, contains := searchtext.Patterns(query.Text)
	hits := []shared.SearchHit{}
	err := s.db.SelectContext(ctx, &hits, `WITH documents AS (
    SELECT id AS id, product_name || ' · ' || contract_number AS title, left(contract_number || ' · ' || id, 240) AS snippet,
           status AS state, '' AS step,
           replace(lower(id || ' ' || contract_number || ' ' || product_code || ' ' || product_name), 'ё', 'е') AS document,
           CASE WHEN id = $3 OR replace(lower(contract_number), 'ё', 'е') = $3 THEN 4
                WHEN replace(lower(product_name), 'ё', 'е') = $3 THEN 3
                WHEN replace(lower(product_name), 'ё', 'е') LIKE $4 THEN 2 ELSE 1 END AS score
    FROM servicing.loans
    WHERE company_id = $1
      AND (replace(lower(id || ' ' || contract_number || ' ' || product_code || ' ' || product_name), 'ё', 'е') LIKE $5
           OR to_tsvector('russian', replace(lower(id || ' ' || contract_number || ' ' || product_code || ' ' || product_name), 'ё', 'е')) @@ plainto_tsquery('russian', $3))
)
SELECT id, title, snippet, state, step, score FROM documents
WHERE ($6::integer = 0 OR score < $6 OR (score = $6 AND id > $7))
ORDER BY score DESC, id ASC LIMIT $2`, scope.CompanyID, query.Limit, query.Text, prefix, contains, query.After.Score, query.After.ID)
	return hits, err
}
