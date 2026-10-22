package support

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

func chatID() string { return rand.Text() }

const conversationCols = `id,user_id,company_id,application_id,created_at`
const messageCols = `id,conversation_id,role,status,text,payload,created_at`

func (s *Store) conversations(ctx context.Context, user, company string) ([]conversation, error) {
	out := []conversation{}
	err := s.db.SelectContext(ctx, &out, `SELECT `+conversationCols+`, COALESCE((SELECT left(m.text,80) FROM support.messages m WHERE m.conversation_id=support.conversations.id AND m.role='user' ORDER BY m.seq LIMIT 1),'Новый диалог') AS title FROM support.conversations WHERE user_id=$1 AND company_id=$2 ORDER BY created_at DESC,id LIMIT 100`, user, company)
	return out, err
}

func (s *Store) ownConversation(ctx context.Context, id, user, company string) (*conversation, error) {
	var out conversation
	err := s.db.GetContext(ctx, &out, `SELECT `+conversationCols+` FROM support.conversations WHERE id=$1 AND user_id=$2 AND company_id=$3`, id, user, company)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.NotFound("Диалог не найден")
	}
	return &out, err
}

func (s *Store) messages(ctx context.Context, id string) ([]chatMessage, error) {
	out := []chatMessage{}
	err := s.db.SelectContext(ctx, &out, `SELECT `+messageCols+` FROM (SELECT `+messageCols+`,seq FROM support.messages WHERE conversation_id=$1 ORDER BY seq DESC LIMIT 100) recent ORDER BY seq`, id)
	return out, err
}

func (s *Store) enqueue(ctx context.Context, id, text, key string, viewer ...viewerContext) (*chatMessage, error) {
	var out chatMessage
	payload := messagePayload{}
	if len(viewer) > 0 {
		payload.Viewer = &viewer[0]
	}
	err := db.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		var locked string
		if err := tx.GetContext(ctx, &locked, `SELECT id FROM support.conversations WHERE id=$1 FOR UPDATE`, id); err != nil {
			return err
		}
		err := tx.GetContext(ctx, &out, `SELECT `+messageCols+` FROM support.messages WHERE conversation_id=$1 AND request_key=$2`, id, key)
		if err == nil {
			if out.Text != text {
				return httpx.Validation("Ключ уже использован для другого сообщения")
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var pending int
		if err := tx.GetContext(ctx, &pending, `SELECT count(id) FROM support.messages WHERE conversation_id=$1 AND status='PENDING'`, id); err != nil {
			return err
		}
		if pending > 0 {
			return httpx.InvalidTransition("Дождитесь ответа на предыдущий вопрос")
		}
		if err := tx.GetContext(ctx, &out, `INSERT INTO support.messages(id,conversation_id,role,status,text,request_key,payload) VALUES($1,$2,'user','COMPLETED',$3,$4,$5) RETURNING `+messageCols, chatID(), id, text, key, db.JSON[messagePayload]{V: payload}); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO support.messages(id,conversation_id,role,status) VALUES($1,$2,'assistant','PENDING')`, chatID(), id)
		return err
	})
	return &out, err
}

func (s *Store) complete(ctx context.Context, msg chatMessage, text, status string, payload messagePayload) error {
	return db.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		for _, b := range payload.Blocks {
			if b.Kind == "action" {
				if _, err := tx.ExecContext(ctx, `INSERT INTO support.actions(id,conversation_id,params) VALUES($1,$2,$3)`, b.ActionID, msg.ConversationID, db.JSON[actionParams]{V: *b.Action}); err != nil {
					return err
				}
			}
		}
		_, err := tx.ExecContext(ctx, `UPDATE support.messages SET text=$2,status=$3,payload=$4 WHERE id=$1 AND status='PENDING'`, msg.ID, text, status, db.JSON[messagePayload]{V: payload})
		return err
	})
}

// RecoverChat помечает незавершённые ответы после остановки сервера; никаких автоповторов действий.
func (m *Module) RecoverChat(ctx context.Context) error {
	_, err := m.deps.DB.ExecContext(ctx, `UPDATE support.messages SET status='FAILED',text='Ответ прерван перезапуском сервера. Отправьте вопрос ещё раз.' WHERE status='PENDING'`)
	return err
}
