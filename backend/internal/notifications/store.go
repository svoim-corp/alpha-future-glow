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

// List — уведомления пользователя, новые первыми; onlyUnread — только непрочитанные.
func (s *Store) List(ctx context.Context, userID string, onlyUnread bool, limit int) ([]notificationRow, error) {
	var out []notificationRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT id, user_id, type, title, text, link, read, created_at FROM notifications.notifications
		WHERE user_id = $1 AND (NOT $2 OR NOT read)
		ORDER BY created_at DESC, id DESC LIMIT $3`, userID, onlyUnread, limit)
	return out, err
}

func (s *Store) UnreadCount(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.db.GetContext(ctx, &n, `SELECT count(*) FROM notifications.notifications WHERE user_id = $1 AND NOT read`, userID)
	return n, err
}
