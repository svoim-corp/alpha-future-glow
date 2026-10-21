// Package origination — зона Origination: заявка, этапы 1 и 2, проверка полноты, отправка, машина состояний. Владелец: разработчик A.
package origination

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
	Products    shared.Products    // рамки продукта: 422 при создании и PATCH
	Companies   shared.Companies   // карточка, адрес, обороты, кандидаты в поручители
	Authorities shared.Authorities // полномочия заявителя при создании черновика
	Prelims     shared.Prelims     // блок prelim и prelimInvalidated
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
	return db.Migrate(ctx, m.deps.DB, "origination", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {
	r.Client.POST("/applications", m.create)
	r.Client.GET("/applications", m.list)
	r.Client.GET("/applications/:id", m.get)
	r.Client.PATCH("/applications/:id", m.patch)
	r.Client.DELETE("/applications/:id", m.delete)
	r.Client.POST("/applications/:id/consents/bki", m.consentBKI)
	r.Client.POST("/applications/:id/guarantor/consent-bki", m.consentGuarantor)
	r.Client.POST("/applications/:id/validate", m.validate)
	r.Client.POST("/applications/:id/submit", m.deps.Idem.Middleware(), m.submitApplication)
	r.Client.GET("/applications/:id/timeline", m.timeline)
}

// ResetDemo удаляет заявки: в демо их создают заново.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
