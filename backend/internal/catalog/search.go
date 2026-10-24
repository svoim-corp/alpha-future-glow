package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Search — ограниченная выдача своей зоны; поиск не зависит от лимита обычного списка.
func (m *Module) Search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {

	return m.store.search(ctx, scope, query)
}
