package servicing

import (
	"context"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Search — ограниченная выдача своей зоны; поиск не зависит от лимита обычного списка.
func (m *Module) Search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {

	if scope.CompanyID == "" || scope.UserID == "" {
		return nil, fmt.Errorf("missing search scope")
	}

	return m.store.search(ctx, scope, query)
}
