package origination

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const actionSubmit = "SUBMIT"

// checkDraft — полнота черновика для validate и submit. Без данных банка поручителя не проверить: 502.
func (m *Module) checkDraft(ctx context.Context, a *appRow, truthConfirmed bool) ([]missingField, error) {
	d, err := m.dossier(ctx, a)
	if err != nil {
		return nil, err
	}
	if d.profile == nil {
		return nil, httpx.Upstream("Данные банка о компании сейчас недоступны, повторите проверку позже")
	}
	return completeness(d, truthConfirmed), nil
}

// submit отправляет заявку в банк: проверка полноты, номер КЗ-<год>-<номер>, переход в SUBMITTED. После коммита —
// ApplicationSubmitted: по нему Scoring ставит заявку в фоновый конвейер, поэтому ответ не ждёт моков.
func (m *Module) submit(ctx context.Context, id string, who applicant, version int, truthConfirmed bool) (*appRow, error) {
	a, err := m.editableDraft(ctx, id, who.companyID, version)
	if err != nil {
		return nil, err
	}
	missing, err := m.checkDraft(ctx, a, truthConfirmed)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		fields := make([]httpx.FieldError, len(missing))
		for i, f := range missing {
			fields[i] = httpx.FieldError{Field: f.Field, Code: "REQUIRED", Message: f.Message}
		}
		return nil, httpx.Validation("Заявка заполнена не полностью", fields...)
	}
	seq, err := m.store.NextNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("next application number: %w", err)
	}

	a, err = m.transition(ctx, id, shared.StatusSubmitted, who.actor, actionSubmit, "", func(a *appRow) error {
		// Под блокировкой: если черновик успели изменить после проверки, отправлять нечего проверенного.
		if a.Version != version {
			return httpx.VersionConflict()
		}
		number := fmt.Sprintf("КЗ-%d-%06d", a.UpdatedAt.Year(), seq)
		a.Number, a.SubmittedAt, a.TruthConfirmed = &number, &a.UpdatedAt, true
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := m.deps.Bus.Publish(ctx, shared.ApplicationSubmitted{AppID: id}); err != nil {
		slog.Error("publish application submitted", "appId", id, "err", err)
	}
	return a, nil
}
