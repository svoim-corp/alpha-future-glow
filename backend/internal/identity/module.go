// Package identity — зона Identity: вход, сессия, роли. Владелец: разработчик A.
package identity

import (
	"context"
	"embed"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Module struct {
	store     *Store
	companies shared.Companies
}

// New применяет миграции зоны. Данные компании читаются через shared.Companies.
func New(ctx context.Context, d *sqlx.DB, companies shared.Companies) (*Module, error) {
	if err := db.Migrate(ctx, d, "identity", migrations); err != nil {
		return nil, err
	}
	return &Module{store: &Store{db: d}, companies: companies}, nil
}

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
	return &auth.User{ID: u.ID, Role: shared.Role(u.Role), CompanyID: u.CompanyID, FullName: u.FullName, Position: u.Position}, nil
}
