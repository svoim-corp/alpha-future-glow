package planner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const planLink = "/planner"

// planInput — поля плана из запроса; nil — поле не передано (в PATCH — не меняется).
type planInput struct {
	Title, Note, Goal  *string
	BudgetRub          *int64
	DateMode, State    *string
	StartDate, EndDate *time.Time
	InApp, Email       *bool
}

// apply сливает переданные поля в план. Пустое описание — описания нет.
func apply(p *planRow, in planInput) {
	set := func(dst *string, v *string) {
		if v != nil {
			*dst = *v
		}
	}
	set(&p.Title, in.Title)
	set(&p.Goal, in.Goal)
	set(&p.DateMode, in.DateMode)
	set(&p.State, in.State)
	if in.Note != nil {
		p.Note = in.Note
		if *in.Note == "" {
			p.Note = nil
		}
	}
	if in.BudgetRub != nil {
		p.BudgetRub = *in.BudgetRub
	}
	if in.StartDate != nil {
		p.StartDate = in.StartDate
	}
	if in.EndDate != nil {
		p.EndDate = in.EndDate
	}
	if in.InApp != nil {
		p.NotifyInApp = *in.InApp
	}
	if in.Email != nil {
		p.NotifyEmail = *in.Email
	}
	normalize(p)
}

// planView — план и подбор по нему на сегодня.
type planView struct {
	plan planRow
	rec  *recommendation
}

func client(companyID string) error {
	if companyID == "" {
		return httpx.Forbidden("Планер доступен только клиенту")
	}
	return nil
}

// own — план своей компании: нет — 404, чужой — 403.
func (m *Module) own(ctx context.Context, companyID, id string) (*planRow, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	p, err := m.store.Plan(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load plan %s: %w", id, err)
	}
	if p == nil {
		return nil, httpx.NotFound("План не найден")
	}
	if p.CompanyID != companyID {
		return nil, httpx.Forbidden("План принадлежит другой компании")
	}
	return p, nil
}

func (m *Module) create(ctx context.Context, companyID, userID string, in planInput) (*planView, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	now := m.now().UTC()
	p := &planRow{CompanyID: companyID, CreatedBy: userID, State: stateActive, CreatedAt: now, UpdatedAt: now}
	apply(p, in)
	if fields := validate(p); len(fields) > 0 {
		return nil, httpx.Validation("", fields...)
	}
	if err := m.store.InsertPlan(ctx, p); err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}
	return m.withRecommendation(ctx, *p)
}

func (m *Module) update(ctx context.Context, companyID, id string, in planInput) (*planView, error) {
	p, err := m.own(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	apply(p, in)
	if fields := validate(p); len(fields) > 0 {
		return nil, httpx.Validation("", fields...)
	}
	p.UpdatedAt = m.now().UTC()
	if err := m.store.UpdatePlan(ctx, p); err != nil {
		return nil, fmt.Errorf("update plan %s: %w", id, err)
	}
	return m.withRecommendation(ctx, *p)
}

func (m *Module) remove(ctx context.Context, companyID, id string) error {
	if _, err := m.own(ctx, companyID, id); err != nil {
		return err
	}
	return m.store.DeletePlan(ctx, id)
}

func (m *Module) list(ctx context.Context, companyID, state string) ([]planView, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	rows, err := m.store.Plans(ctx, companyID, state)
	if err != nil {
		return nil, fmt.Errorf("list plans %s: %w", companyID, err)
	}
	return m.withRecommendations(ctx, rows)
}

// recommendations — планы, по которым подходящий продукт есть уже сегодня (виджет главной).
func (m *Module) recommendations(ctx context.Context, companyID string) ([]planView, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	rows, err := m.store.Plans(ctx, companyID, stateActive)
	if err != nil {
		return nil, fmt.Errorf("list plans %s: %w", companyID, err)
	}
	views, err := m.withRecommendations(ctx, rows)
	if err != nil {
		return nil, err
	}
	today := day(m.now())
	out := views[:0]
	for _, v := range views {
		if v.rec != nil && v.rec.started(today) {
			out = append(out, v)
		}
	}
	return out, nil
}

func (m *Module) withRecommendation(ctx context.Context, p planRow) (*planView, error) {
	views, err := m.withRecommendations(ctx, []planRow{p})
	if err != nil {
		return nil, err
	}
	return &views[0], nil
}

// withRecommendations подбирает продукт каждому плану. Одинаковые цель и бюджет запрашиваются у Catalog один раз.
// Подбор считается при чтении, поэтому здесь же отправляется PLAN_MATCH, если вариант новый.
func (m *Module) withRecommendations(ctx context.Context, rows []planRow) ([]planView, error) {
	type key struct {
		goal   string
		budget int64
	}
	cache := map[key][]shared.ProductView{}
	today := day(m.now())
	views := make([]planView, len(rows))
	for i, p := range rows {
		views[i].plan = p
		if !matchable(&p, today) {
			continue
		}
		k := key{p.Goal, p.BudgetRub}
		products, ok := cache[k]
		if !ok {
			var err error
			if products, err = m.ports.Matcher.Match(ctx, p.Goal, p.BudgetRub); err != nil {
				return nil, fmt.Errorf("match products for plan %s: %w", p.ID, err)
			}
			cache[k] = products
		}
		views[i].rec = recommend(&p, products, today)
		if r := views[i].rec; r != nil && r.started(today) && p.NotifyInApp {
			m.notifyMatch(ctx, &p, r)
		}
	}
	return views, nil
}

// notifyMatch — PLAN_MATCH не чаще одного раза на вариант. Сбой уведомления не мешает показать план.
func (m *Module) notifyMatch(ctx context.Context, p *planRow, r *recommendation) {
	fresh, err := m.store.RecordMatch(ctx, p.ID, r.product.Code, m.now().UTC())
	if err != nil {
		slog.Error("record plan match", "planId", p.ID, "err", err)
		return
	}
	if !fresh {
		return
	}
	err = m.ports.Notifier.Notify(ctx, p.CreatedBy, shared.Notification{
		Type:  shared.NotifyPlanMatch,
		Title: "Нашли финансирование для плана «" + p.Title + "»",
		Text:  r.product.Name + ": " + matchReason + ". Это не решение банка",
		Link:  planLink,
	})
	if err != nil {
		slog.Error("notify plan match", "planId", p.ID, "err", err)
	}
}
