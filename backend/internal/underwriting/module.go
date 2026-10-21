// Package underwriting — зона Underwriting: рабочее место банка, очередь, разбор заявки, ручное решение. Владелец: разработчик B.
package underwriting

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Ports — данные и действия других зон, нужные зоне. Подставляются в cmd/api/main.go.
type Ports struct {
	Applications shared.Applications // анкета, переходы статусов
	Companies    shared.Companies    // данные компании в очереди и анкете
	Products     shared.Products     // 422 для иных условий
	Notifier     shared.Notifier     // уведомление о запросе информации и решении
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
