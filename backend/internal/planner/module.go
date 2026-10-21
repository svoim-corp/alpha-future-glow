// Package planner — зона Planner: планы развития и подбор финансирования по ним (API_SPEC_ADDITIONAL.md). Владелец: разработчик B.
package planner

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Ports — данные и действия других зон, нужные зоне. Подставляются в cmd/api/main.go.
type Ports struct {
	Products shared.Products // подбор продукта по плану
	Notifier shared.Notifier // уведомления PLAN_MATCH
}

type Module struct {
	deps  module.Deps
	ports Ports
}

func New(deps module.Deps) *Module { return &Module{deps: deps} }

// Migrate — миграций пока нет. Когда появятся таблицы, добавьте migrations/ и db.Migrate (образец — identity).
func (m *Module) Migrate(ctx context.Context) error { return nil }

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {}
