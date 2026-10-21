// Package company — зона Company: профиль клиента банка из АБС и реестров, полномочия заявителя. Владелец: разработчик B.
// Таблица компаний (связь id ↔ ИНН) и кэш профилей; сами данные — из Integrations (shared.BankData, shared.Registries).
package company

import (
	"context"
	"embed"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Ports — данные и действия других зон, нужные Company. Подставляются в cmd/api/main.go.
type Ports struct {
	BankData     shared.BankData     // карточка клиента и обороты из АБС
	Registries   shared.Registries   // ЕГРЮЛ/ЕГРИП, ГИР БО, реестр МЧД
	Applications shared.Applications // проверка доступа к company-profile
}

type Module struct {
	deps        module.Deps
	store       *Store
	ports       Ports
	profiles    *profiles
	authorities *authorities
}

func New(deps module.Deps) *Module {
	s := &Store{db: deps.DB}
	return &Module{
		deps: deps, store: s,
		profiles:    &profiles{store: s, now: time.Now},
		authorities: &authorities{store: s, now: time.Now},
	}
}

func (m *Module) Migrate(ctx context.Context) error {
	return db.Migrate(ctx, m.deps.DB, "company", migrations)
}

// Wire подставляет зависимости. Companies() и Authorities() могли отдать раньше: реализации общие по указателю.
func (m *Module) Wire(p Ports) {
	m.ports = p
	m.profiles.bank, m.profiles.reg = p.BankData, p.Registries
	m.authorities.bank, m.authorities.reg = p.BankData, p.Registries
}

func (m *Module) Register(r httpx.Routes) {
	r.Client.GET("/applications/:id/company-profile", m.companyProfile)
}

// ResetDemo сбрасывает кэш профилей, чтобы изменения моков были видны сразу.
func (m *Module) ResetDemo(ctx context.Context) error { return m.store.ResetProfiles(ctx) }
