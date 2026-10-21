package offer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
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

// ownApplication — заявка компании клиента: нет — 404, чужая — 403.
func (m *Module) ownApplication(ctx context.Context, appID, companyID string) (*shared.ApplicationView, error) {
	app, err := m.ports.Applications.Application(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("load application %s: %w", appID, err)
	}
	if app == nil {
		return nil, httpx.NotFound("Заявка не найдена")
	}
	if app.CompanyID != companyID {
		return nil, httpx.Forbidden("Заявка принадлежит другой компании")
	}
	return app, nil
}

// offerOf — предложение по заявке своей компании; решения ещё нет — 404.
func (m *Module) offerOf(ctx context.Context, appID, companyID string) (*shared.ApplicationView, *offerRow, error) {
	app, err := m.ownApplication(ctx, appID, companyID)
	if err != nil {
		return nil, nil, err
	}
	o, err := m.store.Offer(ctx, app.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("load offer %s: %w", app.ID, err)
	}
	if o == nil {
		return nil, nil, httpx.NotFound("Решения по заявке ещё нет")
	}
	return app, o, nil
}

const actionOfferAccept = "OFFER_ACCEPT"

// accept — клиент принимает одобрение или иные условия: договор получает номер, заявка — статус «Подписание».
// Принятие записывается раньше перехода: страница подписания сразу видит договор. Переход не удался — откат.
func (m *Module) accept(ctx context.Context, appID, companyID string, actor shared.Actor) error {
	_, o, err := m.offerOf(ctx, appID, companyID)
	if err != nil {
		return err
	}
	now := m.now().UTC()
	switch {
	case o.Decision == string(shared.StatusRejected):
		return httpx.InvalidTransition("Банк отказал — принимать нечего")
	case o.AcceptedAt != nil:
		return httpx.InvalidTransition("Предложение уже принято")
	case o.DeclinedAt != nil:
		return httpx.InvalidTransition("От предложения уже отказались")
	case !now.Before(o.ValidUntil.AddDate(0, 0, 1)): // действует по дату validUntil включительно
		return httpx.BusinessRule("Срок действия предложения истёк. Оформите новую заявку")
	}
	seq, err := m.store.NextContractNumber(ctx)
	if err != nil {
		return fmt.Errorf("next contract number: %w", err)
	}
	ok, err := m.store.Accept(ctx, appID, fmt.Sprintf("КД-%d-%04d", now.Year(), seq), now)
	if err != nil {
		return fmt.Errorf("accept offer %s: %w", appID, err)
	}
	if !ok {
		return httpx.InvalidTransition("Предложение уже принято")
	}
	if err := m.ports.Applications.Transition(ctx, appID, shared.StatusSigning, actor, actionOfferAccept, ""); err != nil {
		if uerr := m.store.Unaccept(ctx, appID); uerr != nil {
			slog.Error("rollback offer acceptance", "appId", appID, "err", uerr)
		}
		return err
	}
	return nil
}

const actionOfferDecline = "OFFER_DECLINE"

// decline — клиент отказывается от предложения: заявка закрывается, причина попадает в историю для банка.
func (m *Module) decline(ctx context.Context, appID, companyID, reason string, actor shared.Actor) error {
	_, o, err := m.offerOf(ctx, appID, companyID)
	if err != nil {
		return err
	}
	switch {
	case o.Decision == string(shared.StatusRejected):
		return httpx.InvalidTransition("Банк отказал — отказываться не от чего")
	case o.AcceptedAt != nil:
		return httpx.InvalidTransition("Предложение уже принято")
	}
	ok, err := m.store.Decline(ctx, appID, reason, m.now().UTC())
	if err != nil {
		return fmt.Errorf("decline offer %s: %w", appID, err)
	}
	if !ok {
		return httpx.InvalidTransition("На предложение уже ответили")
	}
	if err := m.ports.Applications.Transition(ctx, appID, shared.StatusClosed, actor, actionOfferDecline, reason); err != nil {
		if uerr := m.store.Undecline(ctx, appID); uerr != nil {
			slog.Error("rollback offer decline", "appId", appID, "err", uerr)
		}
		return err
	}
	return nil
}
