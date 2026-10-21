// Package notifications — зона Notifications: in-app уведомления (сквозная зона). Владелец: разработчик A.
package notifications

import (
	"context"
	"embed"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Module struct {
	deps  module.Deps
	store *Store
	now   func() time.Time
}

func New(deps module.Deps) *Module {
	return &Module{deps: deps, store: &Store{db: deps.DB}, now: time.Now}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "notifications", migrations)
}

func (m *Module) Register(r httpx.Routes) {}

// ResetDemo удаляет уведомления: они относятся к заявкам и кредитам демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
