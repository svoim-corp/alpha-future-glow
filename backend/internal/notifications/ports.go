package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Notifier отдаёт реализацию shared.Notifier другим зонам.
func (m *Module) Notifier() shared.Notifier { return notifier{store: m.store, now: m.now} }

type notifier struct {
	store *Store
	now   func() time.Time
}

func (n notifier) Notify(ctx context.Context, userID string, msg shared.Notification) error {
	row := &notificationRow{
		UserID: userID, Type: string(msg.Type), Title: msg.Title, Text: msg.Text, Link: msg.Link, CreatedAt: n.now().UTC(),
	}
	if err := n.store.Insert(ctx, row); err != nil {
		return fmt.Errorf("save notification for %s: %w", userID, err)
	}
	return nil
}
