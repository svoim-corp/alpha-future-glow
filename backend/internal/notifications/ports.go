package notifications

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Notifier отдаёт реализацию shared.Notifier другим зонам.
func (m *Module) Notifier() shared.Notifier { return notifierStub{} }

// notifierStub — заглушка этапа 0: уведомления никуда не сохраняются. Заменить реализацией (A).
type notifierStub struct{}

func (notifierStub) Notify(context.Context, string, shared.Notification) error { return nil }
