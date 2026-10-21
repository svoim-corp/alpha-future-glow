package underwriting

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// mismatchTolerance — на сколько обороты по счетам могут расходиться с выручкой ГИР БО (то же правило в скоринге).
const mismatchTolerance = 0.3

type severity string

const (
	sevInfo    severity = "INFO"
	sevWarn    severity = "WARN"
	sevBlocker severity = "BLOCKER"
)

type issue struct {
	Severity severity `json:"severity" example:"WARN"`
	Code     string   `json:"code" example:"TURNOVER_MISMATCH"`
	Text     string   `json:"text" example:"Обороты по счетам в банке выше выручки в ГИР БО на 7%"`
}

// dossier — всё, что видит сотрудник при разборе заявки. Собирается из других зон через Ports.
type dossier struct {
	app          *shared.ApplicationView
	company      *shared.CompanyProfile
	product      *shared.ProductView // nil — продукт сняли с линейки
	prelim       *shared.PrelimView
	checks       *shared.CheckReport
	consent      *shared.ConsentView // согласие заявителя на БКИ
	guarantor    *shared.Person      // выбранный поручитель из кандидатов банка
	guarantorBKI *shared.ConsentView
}

// load — заявка для сотрудника: черновики банк не видит.
func (m *Module) load(ctx context.Context, appID string) (*shared.ApplicationView, error) {
	app, err := m.ports.Applications.Application(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("load application %s: %w", appID, err)
	}
	if app == nil || app.Status == shared.StatusDraft {
		return nil, httpx.NotFound("Заявка не найдена")
	}
	return app, nil
}

func (m *Module) dossier(ctx context.Context, appID string) (*dossier, error) {
	app, err := m.load(ctx, appID)
	if err != nil {
		return nil, err
	}
	d := &dossier{app: app}
	if d.company, err = m.ports.Companies.Profile(ctx, app.CompanyID); err != nil {
		return nil, err
	}
	if d.company == nil {
		return nil, fmt.Errorf("company %s of application %s not found", app.CompanyID, app.ID)
	}
	if d.product, err = m.ports.Products.Product(ctx, app.ProductCode); err != nil {
		return nil, fmt.Errorf("load product %s: %w", app.ProductCode, err)
	}
	if d.prelim, err = m.ports.Prelims.Prelim(ctx, app.ID); err != nil {
		return nil, fmt.Errorf("load prelim %s: %w", app.ID, err)
	}
	if d.checks, err = m.ports.Checks.Latest(ctx, app.ID); err != nil {
		return nil, fmt.Errorf("load checks %s: %w", app.ID, err)
	}
	if d.consent, err = m.ports.Applications.BKIConsent(ctx, app.ID, shared.ConsentApplicant); err != nil {
		return nil, fmt.Errorf("load bki consent %s: %w", app.ID, err)
	}
	if app.GuarantorPersonID != nil {
		i := slices.IndexFunc(d.company.GuarantorCandidates, func(p shared.Person) bool { return p.PersonID == *app.GuarantorPersonID })
		if i >= 0 {
			d.guarantor = &d.company.GuarantorCandidates[i]
		}
		if d.guarantorBKI, err = m.ports.Applications.BKIConsent(ctx, app.ID, shared.ConsentGuarantor); err != nil {
			return nil, fmt.Errorf("load guarantor consent %s: %w", app.ID, err)
		}
	}
	return d, nil
}

// completeness — проверка полноты и корректности (API_SPEC.md, 4.12). Блокер не даёт принять решение без доработки.
func completeness(d *dossier) (complete bool, issues []issue) {
	add := func(s severity, code, text string) { issues = append(issues, issue{s, code, text}) }
	app, co := d.app, d.company

	if d.consent == nil {
		add(sevBlocker, "BKI_CONSENT_MISSING", "Нет согласия заявителя на запрос в БКИ")
	}
	needGuarantor := co.Type == shared.CompanyLLC && (d.product == nil || d.product.GuarantorForLLC)
	switch {
	case app.GuarantorPersonID == nil && needGuarantor:
		add(sevBlocker, "GUARANTOR_MISSING", "Не выбран поручитель")
	case app.GuarantorPersonID != nil && d.guarantor == nil:
		add(sevBlocker, "GUARANTOR_UNKNOWN", "Поручитель не найден среди руководителя и учредителей по данным банка")
	case app.GuarantorPersonID != nil && d.guarantorBKI == nil:
		add(sevBlocker, "GUARANTOR_CONSENT_MISSING", "Нет согласия поручителя на запрос в БКИ")
	}

	var facts shared.CheckFacts
	if d.checks != nil {
		facts = d.checks.Facts
		for _, r := range d.checks.Results {
			switch r.Status {
			case shared.CheckUnavailable:
				add(sevWarn, "SOURCE_UNAVAILABLE", fmt.Sprintf("Источник %s не ответил — перезапросите проверку", r.Source))
			case shared.CheckFail:
				add(sevWarn, "CHECK_FAILED", fmt.Sprintf("%s: %s", r.Source, r.Summary))
			}
		}
	}
	if g := facts.GIRBO; g != nil && g.Revenue > 0 && co.Turnover.AvgMonthlyIncome > 0 {
		annual := co.Turnover.AvgMonthlyIncome * 12
		diff := float64(annual)/float64(g.Revenue) - 1
		if math.Abs(diff) > mismatchTolerance {
			dir := "выше"
			if diff < 0 {
				dir = "ниже"
			}
			add(sevWarn, "TURNOVER_MISMATCH", fmt.Sprintf("Обороты по счетам в банке %s выручки в ГИР БО на %d%%", dir, int(math.Round(math.Abs(diff)*100))))
		}
	}

	if p := d.prelim; p != nil && p.Choice != nil && *p.Choice == shared.ChoiceKeepAmount {
		add(sevInfo, "AMOUNT_ABOVE_LIMIT", fmt.Sprintf("Клиент оставил сумму выше предварительного лимита %d ₽", p.Limit))
	}
	if a := app.Authority; a.Type == shared.AuthorityMCHD && a.ValidUntil != nil {
		add(sevInfo, "MCHD_APPLICANT", "Заявитель действует по МЧД до "+a.ValidUntil.Format("02.01.2006"))
	}

	for _, is := range issues {
		if is.Severity == sevBlocker {
			return false, issues
		}
	}
	return true, issues
}

// refreshChecks перезапрашивает источники. БКИ — только по действующему согласию заявителя.
func (m *Module) refreshChecks(ctx context.Context, appID string, sources []shared.CheckSource) (*shared.CheckReport, error) {
	app, err := m.load(ctx, appID)
	if err != nil {
		return nil, err
	}
	cv, err := m.ports.Companies.GetByID(ctx, app.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("load company %s: %w", app.CompanyID, err)
	}
	if cv == nil {
		return nil, fmt.Errorf("company %s of application %s not found", app.CompanyID, app.ID)
	}
	req := shared.CheckRequest{AppID: app.ID, INN: cv.INN, CompanyType: cv.Type, Sources: sources}
	consent, err := m.ports.Applications.BKIConsent(ctx, app.ID, shared.ConsentApplicant)
	if err != nil {
		return nil, fmt.Errorf("load bki consent %s: %w", app.ID, err)
	}
	if consent != nil {
		req.ConsentID = consent.ID
	}
	if _, err := m.ports.Checks.Run(ctx, req); err != nil {
		return nil, err
	}
	// Отдаём последний снимок по всем источникам, а не только перезапрошенным: экран показывает полный список.
	return m.ports.Checks.Latest(ctx, app.ID)
}
