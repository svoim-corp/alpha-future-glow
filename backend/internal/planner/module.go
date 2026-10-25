// Package planner — зона Planner: планы развития и подбор финансирования по ним (API_SPEC_ADDITIONAL.md). Владелец: разработчик B.
// Раздел — «Доп.»: у организаторов открыт вопрос, нужен ли он.
package planner

import (
	"context"
	"embed"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Ports — данные и действия других зон, нужные зоне. Подставляются в cmd/api/main.go.
type Ports struct {
	SearchLabels shared.SearchLabelsProvider
	Matcher      shared.ProductMatcher // подбор продукта по цели и бюджету плана
	Notifier     shared.Notifier       // уведомления PLAN_MATCH
}

type Module struct {
	deps  module.Deps
	store *Store
	ports Ports
	now   func() time.Time
}

func New(deps module.Deps) *Module {
	return &Module{deps: deps, store: &Store{db: deps.DB}, now: time.Now}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "planner", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {
	r.Client.GET("/plans", m.listPlans)
	r.Client.POST("/plans", m.createPlan)
	r.Client.GET("/plans/recommendations", m.planRecommendations)
	r.Client.GET("/plans/:id", m.getPlan)
	r.Client.PATCH("/plans/:id", m.patchPlan)
	r.Client.DELETE("/plans/:id", m.deletePlan)
}

// ResetDemo удаляет планы: их создают в ходе демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
