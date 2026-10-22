package support

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func navigation(section string) (*navigationCard, error) {
	sections := map[string]navigationCard{
		"home":         {Section: "home", Label: "Главная", Href: "/"},
		"planner":      {Section: "planner", Label: "Открыть планер", Href: "/planner"},
		"applications": {Section: "applications", Label: "Мои заявки", Href: "/applications"},
		"products":     {Section: "products", Label: "Каталог продуктов", Href: "/products"},
		"credits":      {Section: "credits", Label: "Мои кредиты", Href: "/credits"},
		"calendar":     {Section: "calendar", Label: "Платёжный календарь", Href: "/credits/calendar"},
		"faq":          {Section: "faq", Label: "Частые вопросы", Href: "/support?tab=faq"},
		"tickets":      {Section: "tickets", Label: "Мои обращения", Href: "/support?tab=tickets"},
		"assistant":    {Section: "assistant", Label: "Ассистент", Href: "/support"},
	}
	n, ok := sections[section]
	if !ok {
		return nil, httpx.Validation("Укажите раздел: планер, заявки, кредиты, календарь, каталог или поддержка")
	}
	return &n, nil
}

func accountQuestion(question string) bool {
	q := strings.ToLower(question)
	return strings.Contains(q, "кто я") || strings.Contains(q, "как меня зовут") || strings.Contains(q, "моя должность") || (strings.Contains(q, "компани") && (strings.Contains(q, "называ") || strings.Contains(q, "название моей")))
}

// Прямые короткие команды навигации не требуют генерации модели.
func navigationRequest(question string, rows []chatMessage) string {
	q := strings.Trim(strings.ToLower(strings.TrimSpace(question)), ".!?")
	aliases := map[string]string{"планер": "planner", "каталог": "products", "мои заявки": "applications", "мои кредиты": "credits", "календарь платежей": "calendar", "платёжный календарь": "calendar", "частые вопросы": "faq", "мои обращения": "tickets", "главную": "home"}
	for alias, section := range aliases {
		for _, prefix := range []string{"открой ", "открыть ", "перейди в ", "дай ссылку на ", "где "} {
			if q == prefix+alias {
				return section
			}
		}
	}
	if (strings.Contains(q, "ссыл") || strings.Contains(q, "открой")) && (strings.Contains(q, "на него") || strings.Contains(q, "на неё") || strings.Contains(q, "на нее") || q == "дай ссылку" || q == "ссылку дай") {
		for i := len(rows) - 2; i >= 0; i-- {
			for _, b := range rows[i].Payload.V.Blocks {
				if b.Navigation != nil {
					return b.Navigation.Section
				}
			}
			if rows[i].Role == "user" {
				for alias, section := range aliases {
					if strings.Contains(strings.ToLower(rows[i].Text), alias) {
						return section
					}
				}
			}
		}
	}
	return ""
}

func (m *Module) executeCabinetTool(ctx context.Context, conv conversation, call toolCall) ([]block, error) {
	switch call.Function.Name {
	case "navigate":
		var in struct {
			Section string `json:"section" binding:"required,max=50"`
		}
		if err := strictArgs(call.Function.Arguments, &in); err != nil {
			return nil, err
		}
		n, err := navigation(in.Section)
		if err != nil {
			return nil, err
		}
		return []block{{Kind: "navigation", Title: n.Label, Navigation: n}}, nil
	case "plans":
		var in struct {
			State string `json:"state" binding:"omitempty,oneof=ACTIVE PAUSED DONE ARCHIVED"`
		}
		if err := strictArgs(call.Function.Arguments, &in); err != nil {
			return nil, err
		}
		if m.ports.Planner == nil {
			return nil, httpx.Upstream("Планер сейчас недоступен")
		}
		ps, err := m.ports.Planner.List(ctx, conv.CompanyID, in.State)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, p := range ps {
			out = append(out, block{Kind: "plan", Title: p.Title, Plan: &p})
			if len(out) == 20 {
				break
			}
		}
		return out, nil
	case "propose_plan":
		var in shared.AssistantPlanChange
		if err := strictArgs(call.Function.Arguments, &in); err != nil {
			return nil, err
		}
		if in.ExpectedUpdatedAt != nil {
			return nil, httpx.Validation("Версию плана определяет сервер")
		}
		if m.ports.Planner == nil {
			return nil, httpx.Upstream("Планер сейчас недоступен")
		}
		preview, err := m.ports.Planner.Preview(ctx, conv.CompanyID, in)
		if err != nil {
			return nil, err
		}
		if in.ID != "" {
			in.ExpectedUpdatedAt = &preview.UpdatedAt
		}
		return []block{{Kind: "action", Title: "Сохранить план?", ActionID: chatID(), Action: &actionParams{Kind: "plan", PlanChange: &in, PlanPreview: preview}}}, nil
	case "loans":
		if err := strictArgs(call.Function.Arguments, &struct{}{}); err != nil {
			return nil, err
		}
		if m.ports.Loans == nil {
			return nil, httpx.Upstream("Кредиты сейчас недоступны")
		}
		ls, err := m.ports.Loans.List(ctx, conv.CompanyID)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, l := range ls {
			out = append(out, block{Kind: "loan", Title: l.ProductName, Loan: &l})
			if len(out) == 10 {
				break
			}
		}
		return out, nil
	case "company":
		if err := strictArgs(call.Function.Arguments, &struct{}{}); err != nil {
			return nil, err
		}
		if m.ports.Companies == nil {
			return nil, httpx.Upstream("Данные компании сейчас недоступны")
		}
		p, err := m.ports.Companies.Profile(ctx, conv.CompanyID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, httpx.NotFound("Компания не найдена")
		}
		if p.ID != conv.CompanyID {
			return nil, httpx.Forbidden("")
		}
		card := &companyCard{Name: p.Name, INN: p.INN, RegAddress: p.RegAddress, ActualAddress: p.ActualAddress, MonthlyIncome: p.Turnover.AvgMonthlyIncome}
		if conv.Viewer != nil {
			card.ViewerName = conv.Viewer.Name
			card.ViewerPosition = conv.Viewer.Position
		}
		return []block{{Kind: "company", Title: p.Name, Company: card}}, nil
	case "tickets":
		if err := strictArgs(call.Function.Arguments, &struct{}{}); err != nil {
			return nil, err
		}
		ts, err := m.store.Tickets(ctx, conv.UserID, 10)
		if err != nil {
			return nil, err
		}
		out := []block{}
		for _, t := range ts {
			out = append(out, block{Kind: "ticket", Title: t.Subject, Ticket: &ticketCard{ID: t.ID, Subject: t.Subject, Text: t.Text, Status: t.Status, Reply: t.Reply}})
		}
		return out, nil
	}
	return nil, httpx.Validation("Такое действие недоступно")
}

func toolWithParams(name string, params any) toolCall {
	raw, _ := json.Marshal(params) // Все параметры ниже — простые JSON-структуры без пользовательского кода.
	return toolCall{Function: toolFunction{Name: name, Arguments: raw}}
}

func missingFinance(p intentParams, product bool) string {
	missing := []string{}
	if p.Goal == "" && !product {
		missing = append(missing, "цель финансирования")
	}
	if p.Amount == 0 {
		missing = append(missing, "сумму в рублях")
	}
	if p.Term == 0 {
		missing = append(missing, "срок в месяцах")
	}
	if len(missing) == 0 {
		return ""
	}
	return "Уточните " + strings.Join(missing, ", ") + ". Уже названные параметры повторять не нужно."
}

func (m *Module) structuredAnswer(ctx context.Context, conv conversation, question string, plan chatPlan) (string, messagePayload, error) {
	p := *plan.Params
	// Назначение устройств не теряется, даже если модель оставила код цели пустым.
	// Личное/деловое использование неоднозначной покупки уточняется до подбора.
	q := strings.ToLower(question)
	if strings.Contains(q, "телефон") || strings.Contains(q, "смартфон") {
		if p.Goal == "" {
			p.Goal = "EQUIPMENT"
		}
		if p.BusinessUse == "" {
			p.BusinessUse = "unknown"
		}
	}
	payload := messagePayload{Sources: []source{}, Blocks: []block{}}
	completed := chatPlan{}
	notes := []string{}
	for _, name := range plan.Tools {
		var args any = struct{}{}
		financial := slices.Contains([]string{"products", "calculate", "propose_draft"}, name)
		if financial {
			payload.Finance = &financeContext{Amount: p.Amount, Term: p.Term, Goal: p.Goal, ProductCode: p.ProductCode, BusinessUse: p.BusinessUse, PendingTool: name}
		}
		if financial && p.BusinessUse == "personal" {
			notes = append(notes, "Этот кабинет предназначен для финансирования бизнеса. Потребительский кредит для личных покупок здесь оформить нельзя.")
			continue
		}
		if financial && name != "calculate" && p.BusinessUse == "unknown" {
			note := "Телефоны или другая покупка нужны для бизнеса или для личного использования?"
			if p.Amount > 0 && p.Term > 0 {
				note += fmt.Sprintf(" Сумму %d ₽ и срок %d мес. я уже учёл.", p.Amount, p.Term)
			}
			notes = append(notes, note)
			continue
		}
		switch name {
		case "products", "calculate", "propose_draft":
			listing := name == "products" && p.Goal == "" && p.Amount == 0 && p.Term == 0 && strings.Contains(strings.ToLower(question), "продукт")
			if !listing {
				if text := missingFinance(p, name == "calculate"); text != "" {
					notes = append(notes, text)
					continue
				}
			}
			if name != "products" && p.ProductCode == "" {
				blocks, err := m.executeTool(ctx, conv, toolWithParams("products", map[string]any{"goal": p.Goal, "amount": p.Amount, "term": p.Term}))
				if err != nil {
					notes = append(notes, toolError(err))
					continue
				}
				payload.Blocks = append(payload.Blocks, blocks...)
				notes = append(notes, "Какой продукт выбираете? Напишите его название — остальные параметры уже учтены.")
				continue
			}
			args = map[string]any{"amount": p.Amount, "term": p.Term}
			if name == "products" {
				args = map[string]any{"goal": p.Goal, "amount": p.Amount, "term": p.Term}
			}
			if name == "calculate" {
				args = map[string]any{"productCode": p.ProductCode, "amount": p.Amount, "term": p.Term}
			}
			if name == "propose_draft" {
				args = shared.AssistantDraft{ProductCode: p.ProductCode, Amount: p.Amount, Term: p.Term, Goal: p.Goal}
			}
		case "navigate":
			args = map[string]string{"section": p.Section}
		case "plans":
			args = map[string]string{"state": p.State}
		case "propose_ticket":
			if p.Text == "" || p.Subject == "" {
				notes = append(notes, "Опишите проблему — подготовлю тему и текст обращения.")
				continue
			}
			args = map[string]string{"subject": p.Subject, "text": p.Text}
		case "propose_plan":
			if p.Plan == nil {
				notes = append(notes, "Укажите название, цель и бюджет нового плана или назовите существующий план и нужные изменения.")
				continue
			}
			change := *p.Plan
			if p.PlanOperation == "update" && change.ID == "" {
				matches, err := m.resolvePlan(ctx, conv.CompanyID, p.Query)
				if err != nil {
					notes = append(notes, toolError(err))
					continue
				}
				if len(matches) != 1 || p.Query == "" {
					for _, v := range matches {
						payload.Blocks = append(payload.Blocks, block{Kind: "plan", Title: v.Title, Plan: &v})
					}
					notes = append(notes, "Уточните, какой план изменить: напишите точное название или ID из карточки.")
					continue
				}
				change.ID = matches[0].ID
			}
			if p.PlanOperation != "create" && p.PlanOperation != "update" {
				notes = append(notes, "Создать новый план или изменить существующий?")
				continue
			}
			if p.PlanOperation == "create" && change.ID != "" {
				// Идентификатор нового объекта всегда назначает Planner.
				change.ID = ""
			}
			args = change
		}
		blocks, err := m.executeTool(ctx, conv, toolWithParams(name, args))
		if err != nil {
			notes = append(notes, toolError(err))
			continue
		}
		payload.Blocks = append(payload.Blocks, blocks...)
		completed.Tools = append(completed.Tools, name)
		if financial {
			payload.Finance.PendingTool = ""
		}
	}
	text := cardAnswer(completed, payload.Blocks)
	for _, note := range slices.Compact(notes) {
		if text != "" {
			text += "\n\n"
		}
		text += note
	}
	return text, payload, nil
}

func (m *Module) resolvePlan(ctx context.Context, companyID, query string) ([]shared.AssistantPlan, error) {
	if m.ports.Planner == nil {
		return nil, httpx.Upstream("Планер сейчас недоступен")
	}
	all, err := m.ports.Planner.List(ctx, companyID, "")
	if err != nil {
		return nil, err
	}
	archived, err := m.ports.Planner.List(ctx, companyID, "ARCHIVED")
	if err != nil {
		return nil, err
	}
	all = append(all, archived...)
	out := []shared.AssistantPlan{}
	for _, p := range all {
		if strings.EqualFold(p.ID, query) || strings.EqualFold(p.Title, query) {
			out = append(out, p)
		}
	}
	if len(out) == 0 && query == "" {
		return all, nil
	}
	return out, nil
}
