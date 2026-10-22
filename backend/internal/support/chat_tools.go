package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin/binding"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func strictArgs(raw []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return httpx.Validation("Проверьте параметры действия")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return httpx.Validation("Некорректные параметры")
	}
	if err := binding.Validator.ValidateStruct(out); err != nil {
		return httpx.Validation("Уточните обязательные параметры действия")
	}
	return nil
}

func tool(name, description string, properties map[string]any, required ...string) toolDefinition {
	if required == nil {
		required = []string{}
	}
	return toolDefinition{Type: "function", Function: toolSpec{Name: name, Description: description, Parameters: map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}}}
}

func tools() []toolDefinition {
	s := func() any { return map[string]any{"type": "string"} }
	n := func() any { return map[string]any{"type": "integer", "minimum": 1} }
	return []toolDefinition{
		tool("products", "Каталог; для подбора нужны goal, amount, term. Без параметров — все продукты. Цели: WORKING_CAPITAL, EQUIPMENT, REAL_ESTATE, EXPANSION, CASH_GAP. Используй реальные goals из каталога.", map[string]any{"goal": s(), "amount": n(), "term": n()}),
		tool("calculate", "Предварительный диапазон платежа по продукту, сумма в рублях, срок в месяцах", map[string]any{"productCode": s(), "amount": n(), "term": n()}, "productCode", "amount", "term"),
		tool("applications", "Заявки своей компании и следующие действия. При нескольких заявках уточни выбор.", map[string]any{}),
		tool("payments", "Ближайшие платежи своей компании", map[string]any{}),
		tool("propose_draft", "Только предложение создать черновик, без исполнения. Запроси недостающие параметры, не выдумывай их.", map[string]any{"productCode": s(), "amount": n(), "term": n(), "goal": s()}, "productCode", "amount", "term", "goal"),
		tool("propose_ticket", "Только предложение обращения. Тема и текст отражают проблему пользователя; не выдумывай её.", map[string]any{"subject": s(), "text": s()}, "subject", "text"),
	}
}

func appCard(a shared.ApplicationView) *applicationCard {
	next := "Откройте заявку, чтобы посмотреть детали"
	switch a.Status {
	case shared.StatusDraft:
		next = "Продолжите заполнение черновика"
	case shared.StatusInfoRequested:
		next = "Ответьте на запрос банка в карточке заявки"
	case shared.StatusApproved, shared.StatusCounterOffer:
		next = "Ознакомьтесь с предложением банка"
	case shared.StatusSigning:
		next = "Перейдите к документам для подписания"
	case shared.StatusIssued:
		next = "Кредит выдан; график доступен в разделе кредитов"
	case shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision:
		next = "Заявка рассматривается; следите за уведомлениями"
	}
	return &applicationCard{ID: a.ID, Number: a.Number, ProductCode: a.ProductCode, Amount: a.Amount, Term: a.Term, Status: a.Status.Label(), NextStep: next, CanDelete: a.Status == shared.StatusDraft}
}

func (m *Module) executeTool(ctx context.Context, conv conversation, call toolCall) ([]block, error) {
	raw := call.Function.Arguments
	switch call.Function.Name {
	case "products":
		var in struct {
			Goal   string `json:"goal" binding:"max=100"`
			Amount int64  `json:"amount" binding:"gte=0,lte=1000000000000"`
			Term   int    `json:"term" binding:"gte=0,lte=1200"`
		}
		if err := strictArgs(raw, &in); err != nil {
			return nil, err
		}
		if in.Goal == "" && (in.Amount != 0 || in.Term != 0) {
			return nil, httpx.Validation("Уточните цель финансирования")
		}
		ps, err := m.ports.Catalog.Search(ctx, in.Goal, in.Amount, in.Term)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, p := range ps {
			out = append(out, block{Kind: "product", Title: p.Name, Product: &productCard{p.Code, p.Name, p.Goals, p.AmountMin, p.AmountMax, p.TermMin, p.TermMax, p.RateMin, p.RateMax}})
		}
		return out, nil
	case "calculate":
		var in struct {
			ProductCode string `json:"productCode" binding:"required,max=50"`
			Amount      int64  `json:"amount" binding:"required,gt=0,lte=1000000000000"`
			Term        int    `json:"term" binding:"required,gt=0,lte=1200"`
		}
		if err := strictArgs(raw, &in); err != nil {
			return nil, err
		}
		v, err := m.ports.Catalog.Calculate(ctx, in.ProductCode, in.Amount, in.Term)
		if err != nil {
			return nil, err
		}
		return []block{{Kind: "calculation", Title: "Предварительный платёж", Calculation: v}}, nil
	case "applications":
		if err := strictArgs(raw, &struct{}{}); err != nil {
			return nil, err
		}
		rows, err := m.ports.Drafts.List(ctx, conv.CompanyID)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, a := range rows {
			if a.CompanyID != conv.CompanyID {
				return nil, httpx.Forbidden("")
			}
			if conv.ApplicationID != nil && a.ID != *conv.ApplicationID {
				continue
			}
			out = append(out, block{Kind: "application", Title: "Заявка", Application: appCard(a)})
			if len(out) == 10 {
				break
			}
		}
		return out, nil
	case "payments":
		if err := strictArgs(raw, &struct{}{}); err != nil {
			return nil, err
		}
		rows, err := m.ports.Payments.Upcoming(ctx, conv.CompanyID)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, p := range rows {
			out = append(out, block{Kind: "payment", Title: "Ближайший платёж", Payment: &p})
		}
		return out, nil
	case "propose_draft":
		var in shared.AssistantDraft
		if err := strictArgs(raw, &in); err != nil {
			return nil, err
		}
		if _, err := m.ports.Catalog.Calculate(ctx, in.ProductCode, in.Amount, in.Term); err != nil {
			return nil, err
		}
		ps, err := m.ports.Catalog.Search(ctx, in.Goal, in.Amount, in.Term)
		if err != nil {
			return nil, err
		}
		found := false
		productName := ""
		for _, p := range ps {
			if p.Code == in.ProductCode {
				found = true
				productName = p.Name
			}
		}
		if !found {
			return nil, httpx.BusinessRule("Продукт не подходит под указанные параметры")
		}
		goals := map[string]string{"WORKING_CAPITAL": "Пополнить оборотные средства", "EQUIPMENT": "Купить оборудование", "CASH_GAP": "Закрыть кассовый разрыв", "REAL_ESTATE": "Купить или отремонтировать помещение", "EXPANSION": "Расширить бизнес"}
		return []block{{Kind: "action", Title: "Создать черновик заявки?", ActionID: chatID(), Action: &actionParams{Kind: "draft", Draft: &in, ProductName: productName, GoalLabel: goals[in.Goal]}}}, nil
	case "propose_ticket":
		var in struct {
			Subject string `json:"subject" binding:"required,max=200"`
			Text    string `json:"text" binding:"required,max=4000"`
		}
		if err := strictArgs(raw, &in); err != nil {
			return nil, err
		}
		if strings.TrimSpace(in.Subject) == "" || strings.TrimSpace(in.Text) == "" {
			return nil, httpx.Validation("Опишите проблему")
		}
		return []block{{Kind: "action", Title: "Отправить обращение?", ActionID: chatID(), Action: &actionParams{Kind: "ticket", Subject: in.Subject, Text: in.Text}}}, nil
	default:
		return m.executeCabinetTool(ctx, conv, call)
	}
}

func toolError(err error) string {
	var e *httpx.Error
	if errors.As(err, &e) {
		return e.Message
	}
	return "Данные сейчас недоступны. Предложи открыть нужный раздел или повторить позже."
}

func toolResult(blocks []block) string {
	data, err := json.Marshal(blocks)
	if err != nil {
		return fmt.Sprint("Результат недоступен")
	}
	return string(data)
}
