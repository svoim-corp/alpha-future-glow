package integrations

import (
	"context"
	"errors"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// BankData отдаёт реализацию shared.BankData другим зонам.
func (m *Module) BankData() shared.BankData { return stub{} }

// Registries отдаёт реализацию shared.Registries другим зонам.
func (m *Module) Registries() shared.Registries { return stub{} }

// Checks отдаёт реализацию shared.Checks другим зонам.
func (m *Module) Checks() shared.Checks { return stub{} }

var errNotImplemented = errors.New("integrations: not implemented")

// stub — заглушка: клиентов и записей в реестрах нет, проверки не выполняются. Заменить реализацией (B).
type stub struct{}

func (stub) BankClient(context.Context, string) (*shared.BankClient, error)   { return nil, nil }
func (stub) Turnover(context.Context, string) (*shared.TurnoverReport, error) { return nil, nil }
func (stub) MCHD(context.Context, string) (*shared.MCHDRecord, error)         { return nil, nil }
func (stub) Latest(context.Context, string) (*shared.CheckReport, error) {
	return &shared.CheckReport{}, nil
}
func (stub) Run(context.Context, shared.CheckRequest) (*shared.CheckReport, error) {
	return nil, errNotImplemented
}
func (stub) Registration(context.Context, string, shared.CompanyType) (*shared.Registration, error) {
	return nil, nil
}
