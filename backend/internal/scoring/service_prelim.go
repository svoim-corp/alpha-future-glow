package scoring

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// prelimSources — источники этапа 1 (API_SPEC.md, 4.5); ГИР БО Integrations сама не запросит для ИП.
var prelimSources = []shared.CheckSource{
	shared.SourceEGRUL, shared.SourceBankAccounts, shared.SourceBKI, shared.SourceFNS, shared.SourceCourts, shared.SourceGIRBO,
}

var sourceNames = map[shared.CheckSource]string{
	shared.SourceEGRUL:        "ЕГРЮЛ/ЕГРИП",
	shared.SourceBankAccounts: "обороты по счетам",
	shared.SourceBKI:          "БКИ",
	shared.SourceFNS:          "ФНС",
	shared.SourceCourts:       "КАД/ФССП/ЕФРСБ",
	shared.SourceGIRBO:        "ГИР БО",
}

// draft — черновик своей компании: нет — 404, чужой — 403, уже подан — 409.
func (m *Module) draft(ctx context.Context, appID, companyID string) (*shared.ApplicationView, error) {
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
	if app.Status != shared.StatusDraft {
		return nil, httpx.InvalidTransition("Предварительная проверка доступна только для черновика")
	}
	return app, nil
}

// runPrelim проходит этап 1: проверки источников и предварительное решение. Результат сохраняется даже при стопе:
// его показывает GET /applications/{id} и разбор заявки.
func (m *Module) runPrelim(ctx context.Context, appID, companyID string) (*prelimRow, prelimResult, error) {
	app, err := m.draft(ctx, appID, companyID)
	if err != nil {
		return nil, prelimResult{}, err
	}
	if !slices.Contains(app.Authority.Powers, shared.PowerCreditApplication) {
		return nil, prelimResult{}, httpx.Forbidden("Нет полномочий подавать кредитную заявку от имени компании")
	}
	consent, err := m.ports.Applications.BKIConsent(ctx, app.ID, shared.ConsentApplicant)
	if err != nil {
		return nil, prelimResult{}, fmt.Errorf("load bki consent %s: %w", app.ID, err)
	}
	if consent == nil {
		return nil, prelimResult{}, httpx.Forbidden("Дайте согласие на запрос кредитной истории в БКИ")
	}
	product, err := m.ports.Products.Product(ctx, app.ProductCode)
	if err != nil {
		return nil, prelimResult{}, fmt.Errorf("load product %s: %w", app.ProductCode, err)
	}
	if product == nil {
		return nil, prelimResult{}, httpx.BusinessRule("Продукт заявки не найден, выберите продукт заново")
	}
	cv, err := m.ports.Companies.GetByID(ctx, app.CompanyID)
	if err != nil {
		return nil, prelimResult{}, fmt.Errorf("load company %s: %w", app.CompanyID, err)
	}
	if cv == nil {
		return nil, prelimResult{}, fmt.Errorf("company %s of application %s not found", app.CompanyID, app.ID)
	}

	rep, err := m.ports.Checks.Run(ctx, shared.CheckRequest{
		AppID: app.ID, INN: cv.INN, CompanyType: cv.Type, ConsentID: consent.ID, Sources: prelimSources,
	})
	if err != nil {
		return nil, prelimResult{}, err
	}
	// Без всех ответов решение было бы случайным: лучше попросить повторить, чем выдать неверный стоп или лимит.
	if missing := unavailable(rep); len(missing) > 0 {
		return nil, prelimResult{}, httpx.Upstream("Не удалось получить данные: " + strings.Join(missing, ", ") + ". Повторите проверку позже")
	}

	res := evaluatePrelim(prelimInput{product: *product, amount: app.Amount, term: app.Term, report: rep, now: m.now().UTC()})
	row := &prelimRow{
		AppID: app.ID, Outcome: string(res.outcome), LimitAmount: res.limit, PrelimRate: res.rate, PrelimPayment: res.payment,
		AmountFits: res.amountFits, ProductCode: app.ProductCode, Amount: app.Amount, Term: app.Term,
		Checks: db.JSON[[]prelimCheck]{V: clientChecks(rep)}, CreatedAt: m.now().UTC(),
	}
	if res.outcome == shared.PrelimStop {
		row.StopReason = &res.stopReason
	}
	if err := m.store.InsertPrelim(ctx, row); err != nil {
		return nil, prelimResult{}, fmt.Errorf("save prelim %s: %w", app.ID, err)
	}
	return row, res, nil
}

func unavailable(rep *shared.CheckReport) []string {
	var out []string
	for _, r := range rep.Results {
		if r.Status == shared.CheckUnavailable {
			out = append(out, sourceNames[r.Source])
		}
	}
	return out
}

// clientChecks — проверки для клиента. Integrations пишет сводки без цифр ГИР БО, поэтому их можно показывать как есть.
func clientChecks(rep *shared.CheckReport) []prelimCheck {
	out := make([]prelimCheck, len(rep.Results))
	for i, r := range rep.Results {
		out[i] = prelimCheck{Source: r.Source, Status: r.Status, Label: r.Summary}
	}
	return out
}

// choosePrelim фиксирует выбор клиента, когда сумма больше лимита (API_SPEC.md, prelim/choice).
// Возвращает итоговую сумму и версию заявки.
func (m *Module) choosePrelim(ctx context.Context, appID, companyID string, choice shared.PrelimChoice, amount int64, actor shared.Actor) (*shared.ApplicationView, int64, int, error) {
	app, err := m.draft(ctx, appID, companyID)
	if err != nil {
		return nil, 0, 0, err
	}
	p, err := m.store.LatestPrelim(ctx, app.ID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("load prelim %s: %w", app.ID, err)
	}
	switch {
	case p == nil || p.Outcome != string(shared.PrelimLimit):
		return nil, 0, 0, httpx.InvalidTransition("Сначала пройдите предварительную проверку")
	case p.ProductCode != app.ProductCode || p.Amount != app.Amount || p.Term != app.Term:
		return nil, 0, 0, httpx.InvalidTransition("Параметры заявки изменились, пройдите предварительную проверку заново")
	case p.AmountFits:
		return nil, 0, 0, httpx.InvalidTransition("Сумма в пределах лимита, выбор не нужен")
	}

	if choice == shared.ChoiceKeepAmount {
		if err := m.store.SetChoice(ctx, p.ID, choice, p.Amount, false); err != nil {
			return nil, 0, 0, fmt.Errorf("save prelim choice %s: %w", app.ID, err)
		}
		return app, app.Amount, app.Version, nil
	}

	if amount > p.LimitAmount {
		return nil, 0, 0, httpx.BusinessRule("Сумма больше предварительного лимита")
	}
	product, err := m.ports.Products.Product(ctx, app.ProductCode)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("load product %s: %w", app.ProductCode, err)
	}
	if product == nil || amount < product.AmountMin || amount > product.AmountMax {
		return nil, 0, 0, httpx.BusinessRule("Сумма вне рамок продукта")
	}
	version, err := m.ports.Applications.SetAmount(ctx, app.ID, amount, actor)
	if err != nil {
		return nil, 0, 0, err
	}
	// Результат этапа 1 переносится на новую сумму, чтобы Origination не счёл его устаревшим.
	if err := m.store.SetChoice(ctx, p.ID, choice, amount, true); err != nil {
		return nil, 0, 0, fmt.Errorf("save prelim choice %s: %w", app.ID, err)
	}
	return app, amount, version, nil
}
