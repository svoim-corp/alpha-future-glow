package support

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type faqRow struct {
	ID       string `db:"id"`
	Step     string `db:"step"`
	Question string `db:"question"`
	Answer   string `db:"answer"`
}

// FAQ — частые вопросы шага пути по порядку; step = "" — все шаги.
func (s *Store) FAQ(ctx context.Context, step string) ([]faqRow, error) {
	var out []faqRow
	err := s.db.SelectContext(ctx, &out, `
		SELECT id, step, question, answer FROM support.faq WHERE $1 = '' OR step = $1 ORDER BY step, sort, id`, step)
	return out, err
}

type ticketRow struct {
	ID            string    `db:"id"`
	UserID        string    `db:"user_id"`
	CompanyID     string    `db:"company_id"`
	ApplicationID *string   `db:"application_id"`
	Subject       string    `db:"subject"`
	Text          string    `db:"text"`
	Status        string    `db:"status"`
	Reply         *string   `db:"reply"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (s *Store) InsertTicket(ctx context.Context, t *ticketRow) error {
	return s.db.GetContext(ctx, &t.ID, `
		INSERT INTO support.tickets (id, user_id, company_id, application_id, subject, text, status, created_at, updated_at)
		VALUES ('t-' || nextval('support.ticket_seq'), $1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id`, t.UserID, t.CompanyID, t.ApplicationID, t.Subject, t.Text, t.Status, t.CreatedAt)
}

func (s *Store) ResetTickets(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE support.tickets`)
	return err
}
