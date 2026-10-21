package scoring

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const actionDecision = "DECISION"

// Decide реализует shared.DecisionMaker: и авторешение конвейера, и ручное решение сотрудника идут здесь,
// чтобы рамки продукта, история и событие DecisionMade были одинаковыми.
func (m *Module) Decide(ctx context.Context, appID string, in shared.DecisionInput, actor shared.Actor) (*shared.DecisionView, error) {
	app, err := m.ports.Applications.Application(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("load application %s: %w", appID, err)
	}
	if app == nil {
		return nil, httpx.NotFound("Заявка не найдена")
	}
	if app.Status != shared.StatusOnDecision {
		return nil, httpx.InvalidTransition("Решение принимается по заявке в статусе «На решении»")
	}

	row := &decisionRow{
		AppID: app.ID, Result: string(in.Result), Mode: string(shared.DecisionManual), Reason: in.Reason,
		ProductCode: in.ProductCode, Amount: in.Amount, Term: in.Term, Rate: in.Rate,
		Alternatives: db.JSON[[]shared.Alternative]{V: in.Alternatives}, DecidedAt: m.now().UTC(),
	}
	switch in.Result {
	case shared.StatusApproved, shared.StatusCounterOffer:
		if err := m.checkTerms(ctx, in); err != nil {
			return nil, err
		}
	case shared.StatusRejected:
		if in.Reason == "" {
			return nil, httpx.Validation("", httpx.FieldError{Field: "reason", Code: "REQUIRED", Message: "Укажите причину отказа"})
		}
		// Условия отказа — те, что просил клиент: по ним видно, от чего отказали.
		row.ProductCode, row.Amount, row.Term, row.Rate = app.ProductCode, app.Amount, app.Term, 0
	default:
		return nil, fmt.Errorf("decide %s: unexpected result %q", appID, in.Result)
	}
	if row.Alternatives.V == nil {
		row.Alternatives.V = []shared.Alternative{}
	}
	if actor.Type == shared.ActorSystem {
		row.Mode = string(shared.DecisionAuto)
	} else {
		row.DecidedBy = &actor.UserID
	}

	// Сначала решение, потом статус: Offer по событию сразу читает LatestDecision. Переход не удался — решение убираем.
	if err := m.store.InsertDecision(ctx, row); err != nil {
		return nil, fmt.Errorf("save decision %s: %w", app.ID, err)
	}
	if err := m.ports.Applications.Transition(ctx, app.ID, in.Result, actor, actionDecision, in.Reason); err != nil {
		if derr := m.store.DeleteDecision(ctx, row.ID); derr != nil {
			slog.Error("rollback decision", "decisionId", row.ID, "err", derr)
		}
		return nil, err
	}
	// Решение уже принято и записано: сбой подписчика не должен превращать его в ошибку для сотрудника.
	if err := m.deps.Bus.Publish(ctx, shared.DecisionMade{AppID: app.ID, DecisionID: row.ID}); err != nil {
		slog.Error("publish decision made", "appId", app.ID, "decisionId", row.ID, "err", err)
	}
	return row.view(), nil
}

// checkTerms — условия в рамках продукта (API_SPEC.md, 4.12: 422 вне рамок).
func (m *Module) checkTerms(ctx context.Context, in shared.DecisionInput) error {
	p, err := m.ports.Products.Product(ctx, in.ProductCode)
	if err != nil {
		return fmt.Errorf("load product %s: %w", in.ProductCode, err)
	}
	switch {
	case p == nil:
		return httpx.BusinessRule("Продукт не найден")
	case in.Amount < p.AmountMin || in.Amount > p.AmountMax:
		return httpx.BusinessRule(fmt.Sprintf("Сумма вне рамок продукта: от %d до %d ₽", p.AmountMin, p.AmountMax))
	case in.Term < p.TermMin || in.Term > p.TermMax:
		return httpx.BusinessRule(fmt.Sprintf("Срок вне рамок продукта: от %d до %d мес.", p.TermMin, p.TermMax))
	case in.Rate < p.RateMin || in.Rate > p.RateMax:
		return httpx.BusinessRule(fmt.Sprintf("Ставка вне рамок продукта: от %g до %g%%", p.RateMin, p.RateMax))
	}
	return nil
}

func (r *decisionRow) view() *shared.DecisionView {
	return &shared.DecisionView{
		ID: r.ID, Result: shared.Status(r.Result), Mode: shared.DecisionMode(r.Mode), ProductCode: r.ProductCode,
		Amount: r.Amount, Term: r.Term, Rate: r.Rate, Reason: r.Reason, Alternatives: r.Alternatives.V,
		DecidedBy: r.DecidedBy, DecidedAt: r.DecidedAt,
	}
}
