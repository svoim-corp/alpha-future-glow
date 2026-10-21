package origination

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// transitions — машина состояний заявки (docs/origination.md, «Машина состояний»). Кто инициирует переход,
// решает вызывающая зона; здесь — только какие переходы вообще возможны.
var transitions = map[shared.Status][]shared.Status{
	shared.StatusDraft:         {shared.StatusSubmitted, shared.StatusClosed},
	shared.StatusSubmitted:     {shared.StatusDataCheck},
	shared.StatusDataCheck:     {shared.StatusScoring, shared.StatusInfoRequested},
	shared.StatusInfoRequested: {shared.StatusDataCheck},
	shared.StatusScoring:       {shared.StatusOnDecision},
	shared.StatusOnDecision:    {shared.StatusApproved, shared.StatusCounterOffer, shared.StatusRejected, shared.StatusInfoRequested},
	shared.StatusApproved:      {shared.StatusSigning, shared.StatusClosed},
	shared.StatusCounterOffer:  {shared.StatusSigning, shared.StatusClosed},
	shared.StatusSigning:       {shared.StatusIssued},
}

func canTransition(from, to shared.Status) bool { return slices.Contains(transitions[from], to) }

// transition переводит заявку в статус to и пишет историю в одной транзакции под блокировкой строки, поэтому
// параллельные переходы (конвейер, сотрудник, клиент) не теряются. apply меняет заявку вместе со статусом
// (номер при отправке, отметка удаления) и может отменить переход ошибкой. После коммита публикуется
// ApplicationStatusChanged.
func (m *Module) transition(ctx context.Context, id string, to shared.Status, actor shared.Actor, action, comment string, apply func(*appRow) error) (*appRow, error) {
	var a *appRow
	var from shared.Status
	now := m.now().UTC()
	err := db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		var err error
		if a, err = m.store.LockApplication(ctx, tx, id); err != nil {
			return fmt.Errorf("lock application %s: %w", id, err)
		}
		if a == nil {
			return httpx.NotFound("Заявка не найдена")
		}
		from = shared.Status(a.Status)
		if !canTransition(from, to) {
			return httpx.InvalidTransition(fmt.Sprintf("Нельзя перевести заявку из статуса «%s» в «%s»", from.Label(), to.Label()))
		}
		a.Status, a.UpdatedAt = string(to), now
		if apply != nil {
			if err := apply(a); err != nil {
				return err
			}
		}
		if err := m.store.SaveTransition(ctx, tx, a); err != nil {
			return fmt.Errorf("save status %s: %w", id, err)
		}
		ev := &eventRow{AppID: id, From: string(from), To: string(to), Action: action, ActorType: string(actor.Type), At: now}
		ev.ActorID, ev.ActorName, ev.Comment = optional(actor.UserID), optional(actor.Name), optional(comment)
		return m.store.InsertEvent(ctx, tx, ev)
	})
	if err != nil {
		return nil, err
	}
	// Переход уже записан: сбой подписчика (очередь банка, конвейер) не должен превращать его в ошибку для того,
	// кто перевёл заявку.
	if err := m.deps.Bus.Publish(ctx, shared.ApplicationStatusChanged{AppID: id, From: from, To: to, At: now}); err != nil {
		slog.Error("publish status changed", "appId", id, "to", to, "err", err)
	}
	return a, nil
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
