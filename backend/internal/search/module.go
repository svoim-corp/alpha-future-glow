// Package search — общий поиск клиента; источники предоставляют доменные зоны (docs/search.md).
package search

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Ports — ограниченные поисковые чтения; координатор не получает доступ к чужим таблицам.
type Ports struct {
	Applications shared.SearchProvider
	Loans        shared.SearchProvider
	Plans        shared.SearchProvider
	Products     shared.SearchProvider
	FAQ          shared.SearchProvider
	Tickets      shared.SearchProvider
}

type Module struct{ ports Ports }

func New(ports Ports) *Module { return &Module{ports: ports} }

// Migrate — собственных таблиц нет, индексы принадлежат источникам.
func (m *Module) Migrate(_ context.Context) error { return nil }

func (m *Module) Register(r httpx.Routes) { r.Client.GET("/search", m.search) }
