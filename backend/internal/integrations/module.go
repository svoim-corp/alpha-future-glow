// Package integrations — зона Integrations: адаптеры к мокам внешних источников (АБС банка, ЕГРЮЛ, БКИ, ФНС и др.), снимки ответов. Владелец: разработчик B.
// Единственная зона, которая ходит в MOCKS_BASE_URL. Свои адаптеры отдаёт Company и Scoring через интерфейсы в shared.
package integrations

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

type Module struct {
	deps         module.Deps
	mocksBaseURL string
}

func New(deps module.Deps, mocksBaseURL string) *Module {
	return &Module{deps: deps, mocksBaseURL: mocksBaseURL}
}

// Migrate — миграций пока нет. Когда появятся таблицы, добавьте migrations/ и db.Migrate (образец — identity).
func (m *Module) Migrate(ctx context.Context) error { return nil }

func (m *Module) Register(r httpx.Routes) {}
