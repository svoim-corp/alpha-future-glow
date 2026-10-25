package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// SearchLabels — небольшой словарь кодов без выдачи каталожной БД другим зонам.
func (m *Module) SearchLabels(ctx context.Context) (shared.SearchLabels, error) {
	labels := shared.SearchLabels{Products: map[string]string{}, Goals: map[string]string{}}
	products, err := m.store.Products(ctx)
	if err != nil {
		return labels, err
	}
	goals, err := m.store.Goals(ctx)
	if err != nil {
		return labels, err
	}
	for _, product := range products {
		labels.Products[product.Code] = product.Name
	}
	for _, goal := range goals {
		labels.Goals[goal.Code] = goal.Name
	}
	return labels, nil
}
