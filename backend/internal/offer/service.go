package offer

import (
	"context"
	"fmt"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// offerValidity — сколько дней действует предложение банка (docs/offer.md).
const offerValidity = 7

// Тексты для клиента вместо причины авторешения: она содержит балл скоринга, а он клиенту не показывается.
const (
	autoRejectText  = "Заявка не прошла автоматическую оценку кредитоспособности"
	autoCounterText = "Условия подобраны по результатам оценки кредитоспособности"
)

// onDecisionMade готовит предложение по решению. Решение уже записано: ошибка здесь только логируется вызывающей
// зоной, поэтому подписчик идемпотентен — повторное событие перезаписывает то же предложение.
func (m *Module) onDecisionMade(ctx context.Context, e events.Event) error {
	ev := e.(shared.DecisionMade)
	d, err := m.ports.Decisions.LatestDecision(ctx, ev.AppID)
	if err != nil {
		return fmt.Errorf("load decision %s: %w", ev.AppID, err)
	}
	if d == nil {
		return nil
	}
	app, err := m.ports.Applications.Application(ctx, ev.AppID)
	if err != nil {
		return fmt.Errorf("load application %s: %w", ev.AppID, err)
	}
	if app == nil {
		return nil
	}
	o := buildOffer(app, d)
	if err := m.store.SaveOffer(ctx, o); err != nil {
		return fmt.Errorf("save offer %s: %w", ev.AppID, err)
	}
	return nil
}

// buildOffer — предложение из решения банка и запрошенных клиентом параметров. Платёж — аннуитет после отсрочки,
// которую клиент выбрал в заявке. Срок действия — неделя с даты решения.
func buildOffer(app *shared.ApplicationView, d *shared.DecisionView) *offerRow {
	o := &offerRow{
		AppID: app.ID, DecisionID: d.ID, Decision: string(d.Result),
		ProductCode: d.ProductCode, Amount: d.Amount, Term: d.Term, Rate: d.Rate,
		RequestedProductCode: app.ProductCode, RequestedAmount: app.Amount, RequestedTerm: app.Term,
		Explanation: clientText(d), Alternatives: db.JSON[[]shared.Alternative]{V: d.Alternatives},
		ValidUntil: dateOf(d.DecidedAt).AddDate(0, 0, offerValidity), CreatedAt: d.DecidedAt.UTC(),
	}
	if o.Alternatives.V == nil {
		o.Alternatives.V = []shared.Alternative{}
	}
	if d.Result != shared.StatusRejected {
		o.GraceMonths = app.GraceMonths
		grace := 0
		if app.GraceMonths != nil {
			grace = *app.GraceMonths
		}
		o.MonthlyPayment = finance.Annuity(d.Amount, d.Rate, d.Term-grace)
	}
	return o
}

// clientText — пояснение к решению для клиента. Причину пишет сотрудник для клиента, а у авторешения она
// техническая. Одобрение на запрошенных условиях пояснений не требует.
func clientText(d *shared.DecisionView) *string {
	text := d.Reason
	if d.Mode == shared.DecisionAuto {
		switch d.Result {
		case shared.StatusRejected:
			text = autoRejectText
		case shared.StatusCounterOffer:
			text = autoCounterText
		default:
			text = ""
		}
	}
	if text == "" {
		return nil
	}
	return &text
}

// dateOf — дата без времени в UTC.
func dateOf(t time.Time) time.Time {
	y, mo, d := t.UTC().Date()
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}
