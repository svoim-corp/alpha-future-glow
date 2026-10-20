// Package underwriting — зона Underwriting: рабочее место банка, очередь, разбор заявки, ручное решение. Владелец: разработчик B.
package underwriting

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

type Module struct{ db *sqlx.DB }

// New — заготовка. Когда появятся таблицы, добавьте migrations/ и вызов db.Migrate.
func New(ctx context.Context, db *sqlx.DB) (*Module, error) { return &Module{db: db}, nil }

func (m *Module) Register(r httpx.Routes) {}
