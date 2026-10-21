// Package origination — зона Origination: заявка, этапы 1 и 2, проверка полноты, отправка, машина состояний. Владелец: разработчик A.
package origination

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Ports — данные и действия других зон, нужные зоне. Подставляются в cmd/api/main.go.
type Ports struct {
	Products    shared.Products    // рамки продукта: 422 при создании и PATCH
	Companies   shared.Companies   // карточка, адрес, обороты, кандидаты в поручители
	Authorities shared.Authorities // полномочия заявителя при создании черновика
	Prelims     shared.Prelims     // блок prelim и prelimInvalidated
}

type Module struct {
	deps  module.Deps
	ports Ports
}

func New(deps module.Deps) *Module { return &Module{deps: deps} }

// Migrate — миграций пока нет. Когда появятся таблицы, добавьте migrations/ и db.Migrate (образец — identity).
func (m *Module) Migrate(ctx context.Context) error { return nil }

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {}
