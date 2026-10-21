package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Products отдаёт реализацию shared.Products другим зонам.
func (m *Module) Products() shared.Products { return productsStub{} }

// productsStub — заглушка этапа 0: продуктов нет. Заменить чтением из своей схемы (A).
type productsStub struct{}

func (productsStub) Product(context.Context, string) (*shared.ProductView, error) { return nil, nil }
