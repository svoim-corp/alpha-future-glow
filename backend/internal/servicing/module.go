// Package servicing — зона Servicing: выданные кредиты, график, платёжный календарь, оплаты, досрочное погашение и
// закрытие (docs/servicing.md). Владелец: разработчик A.
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
	Notifier    shared.Notifier    // напоминания PAYMENT_DUE
	Products    shared.Products    // название продукта и признак овердрафта
	Debits      shared.Debits      // счета и списания в АБС (Integrations)
	Companies   shared.Companies   // ИНН компании для АБС
	Authorities shared.Authorities // кто может списывать деньги
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
	r.Client.GET("/loans/page", m.page)
	r.Client.GET("/loans/schedule", m.schedule)
	r.Client.GET("/loans/schedule/upcoming", m.upcoming)
	r.Client.GET("/loans/:id", m.get)
	r.Client.GET("/loans/:id/debt", m.debt)
	r.Client.GET("/loans/:id/schedule/document", m.document)
	r.Client.GET("/loans/:id/repayment-accounts", m.repaymentAccounts)
	// Денежные команды проверяют Idempotency-Key сами: повтор сравнивается по телу и распознаётся под блокировкой
	// кредита, общего middleware для денег недостаточно.
	r.Client.POST("/loans/:id/repayments", m.createRepayment)
	r.Client.GET("/loans/:id/repayments", m.repayments)
	r.Client.GET("/loans/:id/repayments/:opId", m.repayment)
	r.Client.POST("/loans/:id/early-repayment-quotes", m.createQuote)
	r.Client.POST("/loans/:id/early-repayments", m.createEarly)
	r.Client.GET("/loans/:id/early-repayments/:opId", m.earlyRepayment)
	r.Client.POST("/loans/:id/close", m.deps.Idem.Middleware(), m.closeLoan)
}

// ResetDemo удаляет кредиты и графики: в демо их выдают заново.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
