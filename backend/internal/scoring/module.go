// Package scoring — зона Scoring & Decision: предварительный лимит, полный скоринг, авторешение. Владелец: разработчик B.
// Конвейер после отправки заявки работает в фоне: Run запускается из cmd/api.
package scoring

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
	Applications shared.Applications // параметры заявки, переходы статусов, prelim/choice
	Products     shared.Products     // рамки продукта
	Companies    shared.Companies    // ИНН и тип компании для проверок
	Checks       shared.Checks       // внешние проверки (Integrations)
}

type Module struct {
	deps  module.Deps
	store *Store
	ports Ports
	now   func() time.Time
	wake  chan struct{} // сигнал воркеру конвейера: появилась задача
}

func New(deps module.Deps) *Module {
	return &Module{deps: deps, store: &Store{db: deps.DB}, now: time.Now, wake: make(chan struct{}, 1)}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "scoring", migrations)
}

func (m *Module) Wire(p Ports) {
	m.ports = p
	m.subscribe(m.deps.Bus)
}

func (m *Module) Register(r httpx.Routes) {
	r.Client.POST("/applications/:id/prelim", m.prelim)
	r.Client.POST("/applications/:id/prelim/choice", m.prelimChoice)
	r.Bank.GET("/applications/:id/score", m.bankScore)
}

// ResetDemo удаляет результаты этапа 1, скоринг, решения и задачи конвейера: всё это относится к заявкам демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
