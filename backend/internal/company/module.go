// Package company — зона Company: профиль компании/ИП. Владелец: разработчик B.
// Сейчас только минимальная таблица для /auth/me; зона расширит её сама.
package company

import (
	"context"
	"embed"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Module struct{ store *Store }

func New(ctx context.Context, d *sqlx.DB) (*Module, error) {
	if err := db.Migrate(ctx, d, "company", migrations); err != nil {
		return nil, err
	}
	return &Module{store: &Store{db: d}}, nil
}

func (m *Module) Register(r httpx.Routes) {}

// Companies отдаёт реализацию shared.Companies другим зонам.
func (m *Module) Companies() shared.Companies { return m.store }
