package notifications

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type notificationRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Type      string    `db:"type"`
	Title     string    `db:"title"`
	Text      string    `db:"text"`
	Link      string    `db:"link"`
	Read      bool      `db:"read"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *Store) Insert(ctx context.Context, n *notificationRow) error {
	return s.db.GetContext(ctx, &n.ID, `
		INSERT INTO notifications.notifications (id, user_id, type, title, text, link, created_at)
		VALUES ('n-' || nextval('notifications.notification_seq'), $1, $2, $3, $4, $5, $6)
		RETURNING id`, n.UserID, n.Type, n.Title, n.Text, n.Link, n.CreatedAt)
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE notifications.notifications`)
	return err
}
