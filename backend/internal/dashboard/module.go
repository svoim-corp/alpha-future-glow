// Package dashboard — зона Dashboard: сохранённая настройка главной клиента (docs/dashboard.md).
// Владелец: разработчик B. Содержимое виджетов зона не загружает: их данные клиент берёт из API бизнес-зон.
package dashboard

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
	Applications shared.Applications // привязка виджета к заявке — только своей компании
	Loans        shared.LoanOwners   // привязка к кредиту
	Plans        shared.PlanOwners   // привязка к плану
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
	return db.Migrate(ctx, m.deps.DB, "dashboard", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

// Register — Idempotency-Key сохранения зона проверяет сама: общий middleware не сравнивает тело запроса.
func (m *Module) Register(r httpx.Routes) {
	r.Client.GET("/me/dashboard", m.getDashboard)
	r.Client.PUT("/me/dashboard", m.putDashboard)
}

// ResetDemo удаляет настройки главной: их меняют в ходе демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
