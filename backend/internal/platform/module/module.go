// Package module — общие зависимости, которые cmd/api передаёт каждой зоне.
package module

import (
	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/idempotency"
)

// Deps — инфраструктура зоны. Данные и действия других зон сюда не входят: они приходят через Ports зоны и метод Wire.
type Deps struct {
	DB   *sqlx.DB
	Bus  *events.Bus
	Idem *idempotency.Store // Idem.Middleware() — на роуты-команды с Idempotency-Key
}
