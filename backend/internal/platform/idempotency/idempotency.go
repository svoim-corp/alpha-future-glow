// Package idempotency — middleware для команд с заголовком Idempotency-Key
// (submit, sign, disburse, decision). Вешается на конкретные роуты, а не глобально:
//
//	r.Client.POST("/applications/:id/submit", m.deps.Idem.Middleware(), m.submit)
package idempotency

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

//go:embed migrations/*.sql
var migrations embed.FS

const header = "Idempotency-Key"

type Store struct{ db *sqlx.DB }

func New(d *sqlx.DB) *Store { return &Store{db: d} }

// Migrate создаёт таблицу ключей (схема platform).
func Migrate(ctx context.Context, d *sqlx.DB) error {
	return db.Migrate(ctx, d, "platform", migrations)
}

type saved struct {
	Method      string         `db:"method"`
	Path        string         `db:"path"`
	Status      int            `db:"status"`
	Body        []byte         `db:"body"`
	ContentType sql.NullString `db:"content_type"`
}

// Middleware повторяет сохранённый ответ для известного ключа и запоминает успешные (2xx) ответы.
func (s *Store) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(header)
		u := auth.UserFrom(c)
		if key == "" || u == nil {
			c.Next()
			return
		}
		ctx := c.Request.Context()

		var prev saved
		err := s.db.GetContext(ctx, &prev,
			`SELECT method, path, status, body, content_type FROM platform.idempotency_keys WHERE user_id = $1 AND key = $2`,
			u.ID, key)
		switch {
		case err == nil:
			if prev.Method != c.Request.Method || prev.Path != c.Request.URL.Path {
				httpx.Fail(c, httpx.Validation("Ключ идемпотентности уже использован для другого запроса"))
				return
			}
			ct := prev.ContentType.String
			if ct == "" {
				ct = "application/json; charset=utf-8"
			}
			c.Data(prev.Status, ct, prev.Body)
			c.Abort()
			return
		case !errors.Is(err, sql.ErrNoRows):
			httpx.Fail(c, err)
			return
		}

		w := &capture{ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()

		if st := c.Writer.Status(); st >= 200 && st < 300 {
			_, err := s.db.ExecContext(ctx,
				`INSERT INTO platform.idempotency_keys (user_id, key, method, path, status, body, content_type)
				 VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
				u.ID, key, c.Request.Method, c.Request.URL.Path, st, w.buf.Bytes(), c.Writer.Header().Get("Content-Type"))
			if err != nil {
				slog.Error("idempotency save", "err", err, "traceId", httpx.TraceIDFrom(c))
			}
		}
	}
}

// capture копирует тело ответа в буфер.
type capture struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *capture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *capture) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
