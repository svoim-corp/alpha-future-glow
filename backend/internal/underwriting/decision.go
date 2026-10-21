package underwriting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const (
	actionInfoRequest  = "INFO_REQUEST"
	actionInfoAnswered = "INFO_ANSWERED"
)

func appLink(appID string) string { return "/applications/" + appID }

// notify — уведомление не должно отменять действие, которое уже записано: сбой только логируется.
func (m *Module) notify(ctx context.Context, userID string, n shared.Notification) {
	if err := m.ports.Notifier.Notify(ctx, userID, n); err != nil {
		slog.Error("notify", "userId", userID, "type", n.Type, "err", err)
	}
}

// requestInfo — запрос пояснений у клиента (API_SPEC.md, 4.12): заявка → INFO_REQUESTED, клиенту — уведомление.
func (m *Module) requestInfo(ctx context.Context, appID, comment string, items []infoItem, officer shared.Actor) (*infoRequestRow, error) {
	app, err := m.load(ctx, appID)
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(items))
	for i, it := range items {
		keys[i] = it.Key
	}
	slices.Sort(keys)
	if len(slices.Compact(keys)) != len(items) {
		return nil, httpx.Validation("", httpx.FieldError{Field: "items", Code: "INVALID", Message: "Ключи вопросов должны быть уникальными"})
	}

	r := &infoRequestRow{AppID: app.ID, Comment: comment, Items: db.JSON[[]infoItem]{V: items}, CreatedBy: officer.UserID, CreatedAt: m.now().UTC()}
	if err := m.store.InsertInfoRequest(ctx, r); err != nil {
		if errors.Is(err, errOpenRequestExists) {
			return nil, httpx.InvalidTransition("По заявке уже есть запрос, на который клиент ещё не ответил")
		}
		return nil, fmt.Errorf("save info request %s: %w", app.ID, err)
	}
	// Запрос записан раньше перехода: клиент, открыв карточку по событию, сразу его увидит. Переход не удался — убираем.
	if err := m.ports.Applications.Transition(ctx, app.ID, shared.StatusInfoRequested, officer, actionInfoRequest, comment); err != nil {
		if derr := m.store.DeleteInfoRequest(ctx, r.ID); derr != nil {
			slog.Error("rollback info request", "requestId", r.ID, "err", derr)
		}
		return nil, err
	}
	m.notify(ctx, app.ApplicantID, shared.Notification{
		Type: shared.NotifyInfoRequested, Title: "Банку нужны пояснения по заявке", Text: comment, Link: appLink(app.ID),
	})
	r.Status = requestOpen
	return r, nil
}

// ownApplication — заявка своей компании для клиента: нет — 404, чужая — 403.
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

// answerInfo — ответ клиента. Ключи ответа — ровно ключи вопросов. Заявка возвращается в DATA_CHECK,
// дальше её подхватит конвейер Scoring.
func (m *Module) answerInfo(ctx context.Context, appID, companyID, reqID string, answers []infoAnswer, client shared.Actor) error {
	app, err := m.ownApplication(ctx, appID, companyID)
	if err != nil {
		return err
	}
	r, err := m.store.InfoRequest(ctx, app.ID, reqID)
	if err != nil {
		return fmt.Errorf("load info request %s: %w", reqID, err)
	}
	if r == nil {
		return httpx.NotFound("Запрос не найден")
	}
	if r.Status != requestOpen {
		return httpx.InvalidTransition("На этот запрос уже ответили")
	}
	if fields := answerErrors(r.Items.V, answers); len(fields) > 0 {
		return httpx.Validation("Ответьте на все вопросы банка", fields...)
	}

	ok, err := m.store.AnswerInfoRequest(ctx, r.ID, answers, m.now().UTC())
	if err != nil {
		return fmt.Errorf("save answer %s: %w", r.ID, err)
	}
	if !ok {
		return httpx.InvalidTransition("На этот запрос уже ответили")
	}
	if err := m.ports.Applications.Transition(ctx, app.ID, shared.StatusDataCheck, client, actionInfoAnswered, ""); err != nil {
		if rerr := m.store.ReopenInfoRequest(ctx, r.ID); rerr != nil {
			slog.Error("reopen info request", "requestId", r.ID, "err", rerr)
		}
		return err
	}
	return nil
}

// answerErrors сверяет ответы с вопросами: каждый вопрос отвечен один раз, лишних ключей нет.
func answerErrors(items []infoItem, answers []infoAnswer) []httpx.FieldError {
	var fields []httpx.FieldError
	seen := map[string]bool{}
	for _, a := range answers {
		if seen[a.Key] || !slices.ContainsFunc(items, func(it infoItem) bool { return it.Key == a.Key }) {
			fields = append(fields, httpx.FieldError{Field: "answers." + a.Key, Code: "INVALID", Message: "Такого вопроса нет в запросе"})
		}
		seen[a.Key] = true
	}
	for _, it := range items {
		if !seen[it.Key] {
			fields = append(fields, httpx.FieldError{Field: "answers." + it.Key, Code: "REQUIRED", Message: "Нужен ответ: " + it.Label})
		}
	}
	return fields
}

// Типы ручного решения (API_SPEC.md, 4.12).
const (
	decisionApprove = "APPROVE"
	decisionCounter = "COUNTER"
	decisionReject  = "REJECT"
)

var decisionResults = map[string]shared.Status{
	decisionApprove: shared.StatusApproved,
	decisionCounter: shared.StatusCounterOffer,
	decisionReject:  shared.StatusRejected,
}

type decisionCommand struct {
	Type        string
	ProductCode string
	Amount      int64
	Term        int
	Rate        float64
	Reason      string
}

// decide — ручное решение. Одобрение без параметров — на условиях заявки; рамки продукта проверяет Scoring.
// Решивший сотрудник становится исполнителем заявки, если её никто не взял.
func (m *Module) decide(ctx context.Context, appID string, cmd decisionCommand, officer shared.Actor) (*shared.DecisionView, error) {
	app, err := m.load(ctx, appID)
	if err != nil {
		return nil, err
	}
	in := shared.DecisionInput{
		Result: decisionResults[cmd.Type], ProductCode: cmd.ProductCode, Amount: cmd.Amount, Term: cmd.Term, Rate: cmd.Rate, Reason: cmd.Reason,
	}
	if cmd.Type == decisionApprove {
		if in.ProductCode == "" {
			in.ProductCode = app.ProductCode
		}
		if in.Amount == 0 {
			in.Amount = app.Amount
		}
		if in.Term == 0 {
			in.Term = app.Term
		}
	}
	d, err := m.ports.DecisionMaker.Decide(ctx, app.ID, in, officer)
	if err != nil {
		return nil, err
	}
	if _, err := m.store.Assign(ctx, app.ID, officer.UserID, shortName(officer.Name), m.now().UTC()); err != nil {
		slog.Error("assign after decision", "appId", app.ID, "err", err)
	}
	return d, nil
}

// onDecisionMade — уведомление клиенту о решении, и об авторешении, и о решении сотрудника.
func (m *Module) onDecisionMade(ctx context.Context, e events.Event) error {
	ev := e.(shared.DecisionMade)
	app, err := m.ports.Applications.Application(ctx, ev.AppID)
	if err != nil || app == nil {
		return err
	}
	number := app.ID
	if app.Number != nil {
		number = *app.Number
	}
	return m.ports.Notifier.Notify(ctx, app.ApplicantID, shared.Notification{
		Type: shared.NotifyDecision, Title: "Решение по заявке " + number, Text: "Статус заявки: " + app.Status.Label(), Link: appLink(app.ID),
	})
}
