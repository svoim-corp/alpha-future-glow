// Package servicing — зона Servicing: выданные кредиты, график, платёжный календарь. Владелец: разработчик A.
package servicing

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
	Notifier shared.Notifier // напоминания PAYMENT_DUE
	Products shared.Products // название продукта кредита
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
	return db.Migrate(ctx, m.deps.DB, "servicing", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {
	r.Client.GET("/loans/summary", m.summary)
	r.Client.GET("/loans", m.list)
}

// ResetDemo удаляет кредиты и графики: в демо их выдают заново.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
