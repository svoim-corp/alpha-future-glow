// Package integrations — зона Integrations: адаптеры к мокам внешних источников (АБС банка, ЕГРЮЛ, БКИ, ФНС и др.), снимки ответов. Владелец: разработчик B.
// Единственная зона, которая ходит в MOCKS_BASE_URL. Свои адаптеры отдаёт Company и Scoring через интерфейсы в shared.
package integrations

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
	deps    module.Deps
	client  *client
	store   *Store
	checker *checker
}

func New(deps module.Deps, mocksBaseURL string) *Module {
	c := newClient(mocksBaseURL, requestTimeout)
	s := &Store{db: deps.DB}
	return &Module{deps: deps, client: c, store: s, checker: &checker{client: c, store: s, now: time.Now}}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "integrations", migrations)
}

// Register — своих эндпоинтов у зоны нет: проверки показывает Underwriting.
func (m *Module) Register(r httpx.Routes) {}

// ResetDemo удаляет снимки проверок: они относятся к заявкам демо.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.Reset(ctx) }
