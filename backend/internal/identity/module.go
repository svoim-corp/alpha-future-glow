// Package identity — зона Identity: вход, сессия, роли. Владелец: разработчик A.
package identity

import (
	"context"
	"embed"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Ports — данные других зон, нужные Identity.
type Ports struct {
	Companies shared.Companies // /auth/me
}

type Module struct {
	deps  module.Deps
	store *Store
	ports Ports
}

func New(deps module.Deps) *Module {
	return &Module{deps: deps, store: &Store{db: deps.DB}}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "identity", migrations)
}

func (m *Module) Wire(p Ports) { m.ports = p }

func (m *Module) Register(r httpx.Routes) {
	r.Public.POST("/auth/login", m.login)
	r.Client.POST("/auth/logout", m.logout)
	r.Client.GET("/auth/me", m.me)
}

// Resolve реализует auth.SessionResolver.
func (m *Module) Resolve(ctx context.Context, token string) (*auth.User, error) {
	u, err := m.store.userByToken(ctx, token)
	if err != nil || u == nil {
		return nil, err
	}
	return &auth.User{ID: u.ID, Role: shared.Role(u.Role), CompanyID: u.CompanyID, PersonINN: u.PersonINN, FullName: u.FullName, Position: u.Position}, nil
}
