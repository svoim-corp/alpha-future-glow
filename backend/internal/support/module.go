// Package support — зона Support: частые вопросы и обращения. Владелец: разработчик A.
package support

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

func New(deps module.Deps) *Module { return &Module{deps: deps, store: &Store{db: deps.DB}} }

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "support", migrations)
}

// Register — частые вопросы открыты без входа, как и каталог.
func (m *Module) Register(r httpx.Routes) {
	r.Public.GET("/support/faq", m.faq)
}
