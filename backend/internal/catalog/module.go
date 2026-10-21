// Package catalog — зона Catalog: линейка продуктов, их рамки, подбор по задаче, калькулятор платежа. Владелец: разработчик A.
package catalog

import (
	"context"
	"embed"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Module struct {
	deps  module.Deps
	store *Store
}

func New(deps module.Deps) *Module { return &Module{deps: deps, store: &Store{db: deps.DB}} }

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "catalog", migrations)
}

// Register — каталог открыт без входа: его смотрят до решения оформить заявку (docs/catalog.md).
func (m *Module) Register(r httpx.Routes) {
	r.Public.GET("/products", m.listProducts)
	r.Public.GET("/products/goals", m.listGoals)
	r.Public.GET("/products/:code", m.productDetails)
}
