package support

import (
	"context"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// SearchTickets — ограниченная выдача своей зоны; поиск не зависит от лимита обычного списка.
func (m *Module) SearchTickets(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {

	if scope.CompanyID == "" || scope.UserID == "" {
		return nil, fmt.Errorf("missing search scope")
	}

	return m.store.searchTickets(ctx, scope, query)
}
