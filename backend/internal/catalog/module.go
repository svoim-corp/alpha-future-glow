// Package catalog — зона Catalog: линейка продуктов, их рамки, подбор по задаче, калькулятор платежа. Владелец: разработчик A.
package catalog

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

type Module struct{ deps module.Deps }

func New(deps module.Deps) *Module { return &Module{deps: deps} }

// Migrate — миграций пока нет. Когда появятся таблицы, добавьте migrations/ и db.Migrate (образец — identity).
func (m *Module) Migrate(ctx context.Context) error { return nil }

func (m *Module) Register(r httpx.Routes) {}
