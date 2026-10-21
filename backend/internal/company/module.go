// Package company — зона Company: профиль клиента банка из АБС и реестров, полномочия заявителя. Владелец: разработчик B.
// Сейчас — таблица компаний (связь id ↔ ИНН) для /auth/me и заглушки Profile и Authorities (ports.go).
package company

import (
	"context"
	"embed"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Module struct {
	deps  module.Deps
	store *Store
}

func New(deps module.Deps) *Module {
	return &Module{deps: deps, store: &Store{db: deps.DB}}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "company", migrations)
}

func (m *Module) Register(r httpx.Routes) {}
