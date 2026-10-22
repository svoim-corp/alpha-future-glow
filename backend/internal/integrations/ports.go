package integrations

import "github.com/svoim-corp/alpha-future-glow/backend/internal/shared"

// BankData отдаёт реализацию shared.BankData другим зонам.
func (m *Module) BankData() shared.BankData { return m.client }

// Registries отдаёт реализацию shared.Registries другим зонам.
func (m *Module) Registries() shared.Registries { return m.client }

// Checks отдаёт реализацию shared.Checks другим зонам.
func (m *Module) Checks() shared.Checks { return m.checker }

// Debits отдаёт реализацию shared.Debits другим зонам.
func (m *Module) Debits() shared.Debits { return m.client }
