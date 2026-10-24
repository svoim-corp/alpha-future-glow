package support

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Store) faqByID(ctx context.Context, id string) (*faqRow, error) {
	var row faqRow
	err := s.db.GetContext(ctx, &row, `SELECT id, step, question, answer FROM support.faq WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &row, err
}

func (s *Store) ticketByID(ctx context.Context, id, userID string) (*ticketRow, error) {
	var row ticketRow
	err := s.db.GetContext(ctx, &row, `
		SELECT id, user_id, company_id, application_id, subject, text, status, reply, created_at, updated_at
		FROM support.tickets WHERE id = $1 AND user_id = $2`, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &row, err
}
