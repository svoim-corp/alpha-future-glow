package support

import (
	"context"
	"fmt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

const ticketOpen = "OPEN"

// openTicket создаёт обращение. Заявка в обращении — только своей компании: иначе через поддержку можно было бы
// узнать о чужой заявке.
func (m *Module) openTicket(ctx context.Context, t *ticketRow) error {
	if id := t.ApplicationID; id != nil {
		app, err := m.ports.Applications.Application(ctx, *id)
		if err != nil {
			return fmt.Errorf("load application %s: %w", *id, err)
		}
		if app == nil || app.CompanyID != t.CompanyID {
			return httpx.Validation("", httpx.FieldError{Field: "applicationId", Code: "INVALID", Message: "Заявка не найдена"})
		}
	}
	t.Status, t.CreatedAt = ticketOpen, m.now().UTC()
	t.UpdatedAt = t.CreatedAt
	if err := m.store.InsertTicket(ctx, t); err != nil {
		return fmt.Errorf("save ticket: %w", err)
	}
	return nil
}
