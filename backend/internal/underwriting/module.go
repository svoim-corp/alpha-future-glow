// Package underwriting — зона Underwriting: рабочее место банка, очередь, разбор заявки, ручное решение. Владелец: разработчик B.
// Очередь — своя таблица, которую ведёт подписка на ApplicationStatusChanged.
package underwriting

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
	Applications  shared.Applications  // анкета, переходы статусов
	Companies     shared.Companies     // данные компании в очереди и анкете
	Products      shared.Products      // название продукта в очереди
	Notifier      shared.Notifier      // уведомление о запросе информации и решении
	Scores        shared.Scores        // балл, риск и режим решения в очереди
	Prelims       shared.Prelims       // предварительный результат в анкете
	Checks        shared.Checks        // внешние проверки и их перезапрос
	DecisionMaker shared.DecisionMaker // ручное решение (Scoring)
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
	return db.Migrate(ctx, m.deps.DB, "underwriting", migrations)
}

func (m *Module) Wire(p Ports) {
	m.ports = p
	m.deps.Bus.Subscribe(shared.ApplicationStatusChanged{}.Name(), m.onStatusChanged)
	m.deps.Bus.Subscribe(shared.DecisionMade{}.Name(), m.onDecisionMade)
}

func (m *Module) Register(r httpx.Routes) {
	r.Bank.GET("/applications/stats", m.bankStats)
	r.Bank.GET("/applications", m.bankQueue)
	r.Bank.POST("/applications/:id/assign", m.bankAssign)
	r.Bank.GET("/applications/:id", m.bankApplication)
	r.Bank.GET("/applications/:id/completeness", m.bankCompleteness)
	r.Bank.GET("/applications/:id/checks", m.bankChecks)
	r.Bank.POST("/applications/:id/checks/refresh", m.bankRefreshChecks)
	r.Bank.GET("/applications/:id/info-requests", m.bankInfoRequests)
	r.Bank.POST("/applications/:id/info-requests", m.bankRequestInfo)
	r.Bank.POST("/applications/:id/decision", m.deps.Idem.Middleware(), m.bankDecision)

	r.Client.GET("/applications/:id/info-requests", m.clientInfoRequests)
	r.Client.POST("/applications/:id/info-requests/:reqId/answer", m.clientAnswerInfo)
}

// ResetDemo очищает очередь и запросы информации: они относятся к заявкам демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
