package support

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// SearchFAQ — ограниченная выдача своей зоны; поиск не зависит от лимита обычного списка.
func (m *Module) SearchFAQ(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {

	return m.store.searchFAQ(ctx, scope, query)
}
