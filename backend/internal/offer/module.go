// Package offer — зона Offer & Contract: оффер, принятие условий, имитация подписи и выдачи. Владелец: разработчик A.
// Предложение появляется по событию DecisionMade из решения Scoring.
package offer

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
	Applications shared.Applications // переходы SIGNING, ISSUED, CLOSED
	Decisions    shared.Decisions    // условия оффера по событию DecisionMade
	Companies    shared.Companies    // договор поручительства только для ООО
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
	return db.Migrate(ctx, m.deps.DB, "offer", migrations)
}

func (m *Module) Wire(p Ports) {
	m.ports = p
	m.deps.Bus.Subscribe(shared.DecisionMade{}.Name(), m.onDecisionMade)
}

func (m *Module) Register(r httpx.Routes) {
	r.Client.GET("/applications/:id/offer", m.getOffer)
	r.Client.POST("/applications/:id/offer/accept", m.acceptOffer)
	r.Client.POST("/applications/:id/offer/decline", m.declineOffer)
	r.Client.GET("/applications/:id/contract", m.getContract)
}

// ResetDemo удаляет предложения и договоры: они относятся к заявкам демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
