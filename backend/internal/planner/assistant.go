package planner

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Assistant предоставляет те же правила планера без вызова HTTP из Support.
func (m *Module) Assistant() shared.AssistantPlanner { return assistant{m} }

type assistant struct{ m *Module }

func planCard(p planRow) shared.AssistantPlan {
	return shared.AssistantPlan{ID: p.ID, Title: p.Title, Goal: p.Goal, Note: p.Note, BudgetRub: p.BudgetRub,
		DateMode: p.DateMode, StartDate: dateStr(p.StartDate), EndDate: dateStr(p.EndDate), State: p.State, UpdatedAt: p.UpdatedAt}
}

func (a assistant) List(ctx context.Context, companyID, state string) ([]shared.AssistantPlan, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	if state != "" && !validStates[state] {
		return nil, httpx.Validation("Недопустимое состояние плана")
	}
	rows, err := a.m.store.Plans(ctx, companyID, state)
	if err != nil {
		return nil, err
	}
	out := make([]shared.AssistantPlan, 0, len(rows))
	for _, p := range rows {
		out = append(out, planCard(p))
	}
	return out, nil
}

func preparedPlan(p planRow, in shared.AssistantPlanChange) (*planRow, error) {
	if err := binding.Validator.ValidateStruct(in); err != nil {
		return nil, httpx.Validation("Проверьте параметры плана")
	}
	apply(&p, planInput{Title: in.Title, Goal: in.Goal, Note: in.Note, BudgetRub: in.BudgetRub,
		DateMode: in.DateMode, StartDate: parseDate(in.StartDate), EndDate: parseDate(in.EndDate), State: in.State})
	if strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.Goal) == "" || p.BudgetRub <= 0 {
		return nil, httpx.Validation("Укажите название, цель и бюджет плана")
	}
	if fields := validate(&p); len(fields) > 0 {
		return nil, httpx.Validation("Проверьте даты плана", fields...)
	}
	return &p, nil
}

func (a assistant) Preview(ctx context.Context, companyID string, in shared.AssistantPlanChange) (*shared.AssistantPlan, error) {
	if err := client(companyID); err != nil {
		return nil, err
	}
	p := planRow{DateMode: modeUnscheduled, State: stateActive, NotifyInApp: true}
	if in.ID != "" {
		own, err := a.m.own(ctx, companyID, in.ID)
		if err != nil {
			return nil, err
		}
		p = *own
	}
	prepared, err := preparedPlan(p, in)
	if err != nil {
		return nil, err
	}
	out := planCard(*prepared)
	return &out, nil
}

func (a assistant) Apply(ctx context.Context, who shared.DraftApplicant, in shared.AssistantPlanChange, commandID string) (*shared.AssistantPlan, error) {
	if err := client(who.CompanyID); err != nil {
		return nil, err
	}
	if who.UserID == "" || commandID == "" {
		return nil, httpx.Validation("Не указан контекст подтверждения")
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(append([]byte(who.CompanyID+":"), raw...))
	fingerprint := hex.EncodeToString(sum[:])
	var result shared.AssistantPlan
	err = db.WithTx(ctx, a.m.deps.DB, func(tx *sqlx.Tx) error {
		// Один actionId сериализуется до чтения квитанции, включая повтор после сбоя Support.
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, who.UserID+":"+commandID); err != nil {
			return err
		}
		var receipt struct {
			Fingerprint string                        `db:"fingerprint"`
			Result      db.JSON[shared.AssistantPlan] `db:"result"`
		}
		err := tx.GetContext(ctx, &receipt, `SELECT fingerprint,result FROM planner.assistant_commands WHERE user_id=$1 AND command_id=$2`, who.UserID, commandID)
		if err == nil {
			if receipt.Fingerprint != fingerprint {
				return httpx.Validation("Подтверждение уже использовано для другого изменения")
			}
			result = receipt.Result.V
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		now := a.m.now().UTC().Truncate(time.Microsecond)
		p := planRow{CompanyID: who.CompanyID, CreatedBy: who.UserID, DateMode: modeUnscheduled, State: stateActive, NotifyInApp: true, CreatedAt: now}
		if in.ID != "" {
			if err := tx.GetContext(ctx, &p, `SELECT `+planCols+` FROM planner.plans WHERE id=$1 FOR UPDATE`, in.ID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return httpx.NotFound("План не найден")
				}
				return err
			}
			if p.CompanyID != who.CompanyID {
				return httpx.Forbidden("План принадлежит другой компании")
			}
			if in.ExpectedUpdatedAt == nil || !p.UpdatedAt.Equal(*in.ExpectedUpdatedAt) {
				return httpx.BusinessRule("План изменился. Попросите ассистента подготовить предложение заново.")
			}
		}
		prepared, err := preparedPlan(p, in)
		if err != nil {
			return err
		}
		prepared.UpdatedAt = now
		if in.ID == "" {
			err = a.m.store.insertPlan(ctx, tx, prepared)
		} else {
			err = a.m.store.updatePlan(ctx, tx, prepared)
		}
		if err != nil {
			return err
		}
		result = planCard(*prepared)
		_, err = tx.ExecContext(ctx, `INSERT INTO planner.assistant_commands(user_id,command_id,fingerprint,result) VALUES($1,$2,$3,$4)`, who.UserID, commandID, fingerprint, db.JSON[shared.AssistantPlan]{V: result})
		return err
	})
	return &result, err
}
