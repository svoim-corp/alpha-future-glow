package planner

import (
	"context"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestAssistantPlanConfirmation(t *testing.T) {
	e := newEnv(t, true)
	ctx := context.Background()
	a := e.m.Assistant()
	who := shared.DraftApplicant{UserID: "u-1", CompanyID: "c-1"}
	title, goal, mode, date := "Печь", "EQUIPMENT", "DAY", "2026-11-10"
	budget := int64(500000)
	in := shared.AssistantPlanChange{Title: &title, Goal: &goal, BudgetRub: &budget, DateMode: &mode, StartDate: &date}
	preview, err := a.Preview(ctx, who.CompanyID, in)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := a.List(ctx, who.CompanyID, "")
	if err != nil || len(rows) != 0 {
		t.Fatal("подготовка сохранила план", rows, err)
	}
	if preview.StartDate == nil || *preview.StartDate != date {
		t.Fatal("потеряна дата", preview)
	}
	saved, err := a.Apply(ctx, who, in, "create")
	if err != nil {
		t.Fatal(err)
	}
	replay, err := a.Apply(ctx, who, in, "create")
	if err != nil || replay.ID != saved.ID {
		t.Fatal("неидемпотентное создание", replay, err)
	}
	rows, _ = a.List(ctx, who.CompanyID, "")
	if len(rows) != 1 {
		t.Fatal("дубликат плана", rows)
	}
	other := shared.DraftApplicant{UserID: "u-10", CompanyID: "c-8"}
	newBudget := int64(600000)
	change := shared.AssistantPlanChange{ID: saved.ID, BudgetRub: &newBudget, ExpectedUpdatedAt: &saved.UpdatedAt}
	if _, err := a.Preview(ctx, other.CompanyID, change); err == nil {
		t.Fatal("подготовлено изменение чужого плана")
	}
	if _, err := a.Apply(ctx, other, change, "foreign"); err == nil {
		t.Fatal("изменён чужой план")
	}
	e.m.now = func() time.Time { return now.Add(time.Minute) }
	updated, err := a.Apply(ctx, who, change, "budget")
	if err != nil {
		t.Fatal(err)
	}
	if updated.BudgetRub != newBudget || updated.StartDate == nil || *updated.StartDate != date {
		t.Fatal("изменены непереданные поля", updated)
	}
	if _, err := a.Apply(ctx, who, change, "stale"); err == nil {
		t.Fatal("принято устаревшее предложение")
	}
	replay, err = a.Apply(ctx, who, change, "budget")
	if err != nil || replay.BudgetRub != newBudget {
		t.Fatal("повтор после изменения версии", replay, err)
	}
	state := "ARCHIVED"
	archive := shared.AssistantPlanChange{ID: saved.ID, State: &state, ExpectedUpdatedAt: &updated.UpdatedAt}
	if _, err := a.Apply(ctx, who, archive, "archive"); err != nil {
		t.Fatal(err)
	}
	rows, _ = a.List(ctx, who.CompanyID, "")
	if len(rows) != 0 {
		t.Fatal("архив остался в активных")
	}
	rows, _ = a.List(ctx, who.CompanyID, "ARCHIVED")
	if len(rows) != 1 || rows[0].State != "ARCHIVED" {
		t.Fatal("архив не сохранён", rows)
	}
	if _, err := a.Apply(ctx, shared.DraftApplicant{UserID: who.UserID, CompanyID: other.CompanyID}, in, "create"); err == nil {
		t.Fatal("квитанция раскрыла план другой компании")
	}
}
