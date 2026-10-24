package planner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Search — ограниченная выдача своей зоны; поиск не зависит от лимита обычного списка.
func (m *Module) Search(ctx context.Context, scope shared.SearchScope, query shared.SearchQuery) ([]shared.SearchHit, error) {

	if scope.CompanyID == "" || scope.UserID == "" {
		return nil, fmt.Errorf("missing search scope")
	}

	if m.ports.SearchLabels == nil {
		return nil, fmt.Errorf("search labels unavailable")
	}
	labels, err := m.ports.SearchLabels(ctx)
	if err != nil {
		return nil, fmt.Errorf("load search labels: %w", err)
	}
	products, err := json.Marshal(labels.Products)
	if err != nil {
		return nil, err
	}
	goals, err := json.Marshal(labels.Goals)
	if err != nil {
		return nil, err
	}

	return m.store.search(ctx, scope, query, string(products), string(goals))
}
