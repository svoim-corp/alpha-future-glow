// Package support — зона Support: частые вопросы и обращения. Владелец: разработчик A.
package support

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

// Ports — данные других зон, нужные зоне. Подставляются в cmd/api/main.go.
type Ports struct {
	Applications shared.Applications // заявка в обращении — своей компании
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
	return db.Migrate(ctx, m.deps.DB, "support", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

// Register — частые вопросы открыты без входа, как и каталог.
func (m *Module) Register(r httpx.Routes) {
	r.Public.GET("/support/faq", m.faq)
	r.Client.POST("/support/tickets", m.createTicket)
	r.Client.GET("/support/tickets", m.listTickets)
}

// ResetDemo удаляет обращения; частые вопросы — seed и остаются.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.ResetTickets(ctx) }
