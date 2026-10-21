package identity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct{ db *sqlx.DB }

type userRow struct {
	ID           string  `db:"id"`
	Login        string  `db:"login"`
	PasswordHash string  `db:"password_hash"`
	Role         string  `db:"role"`
	CompanyID    *string `db:"company_id"`
	PersonINN    *string `db:"person_inn"`
	FullName     string  `db:"full_name"`
	Position     string  `db:"position"`
}

const userCols = `u.id, u.login, u.password_hash, u.role, u.company_id, u.person_inn, u.full_name, u.position`

// FindByLogin возвращает пользователя или (nil, nil).
func (s *Store) FindByLogin(ctx context.Context, login string) (*userRow, error) {
	var u userRow
	err := s.db.GetContext(ctx, &u, `SELECT `+userCols+` FROM identity.users u WHERE u.login = $1`, login)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

// CreateSession создаёт сессию с токеном из 32 случайных байт.
func (s *Store) CreateSession(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO identity.sessions (token, user_id, expires_at) VALUES ($1, $2, now() + $3 * interval '1 second')`,
		token, userID, int64(ttl.Seconds()))
	return token, err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM identity.sessions WHERE token = $1`, token)
	return err
}

// userByToken возвращает владельца действующей сессии или (nil, nil).
func (s *Store) userByToken(ctx context.Context, token string) (*userRow, error) {
	var u userRow
	err := s.db.GetContext(ctx, &u, `SELECT `+userCols+`
		FROM identity.sessions s JOIN identity.users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > now()`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}
