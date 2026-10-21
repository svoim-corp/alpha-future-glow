package scoring

import (
	"fmt"
	"math"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Стоп-причины этапа 1 (API_SPEC.md, 4.5).
const (
	stopCompanyInactive  = "COMPANY_INACTIVE"
	stopBusinessTooYoung = "BUSINESS_TOO_YOUNG"
	stopBKIOverdue       = "BKI_OVERDUE"
	stopCriticalRisk     = "CRITICAL_RISK"
)

// Правила этапа 1. Значения подобраны под демо-данные моков (backend/SCENARIOS.md).
const (
	stateActive = "ACTIVE"

	minBusinessAgeMonths = 6         // меньше — стоп, даже если продукт допускает моложе
	maxOverdueDays       = 30        // просрочка дольше — стоп
	criticalTaxDebt      = 1_000_000 // налоговый долг больше — критический риск

	// Лимит — доля годового оборота по счетам в банке, без выписок (Уточн.). Долг в лимите не учитывается:
	// его вес — фактор DEBT_LOAD полного скоринга, иначе клиент с большим, но обслуживаемым долгом не дошёл бы до банка.
	limitTurnoverShare = 0.15
	limitStep          = 100_000

	warnRateMarkup = 1.5 // надбавка к минимальной ставке продукта за каждое предупреждение проверок
	paymentStep    = 100 // предварительный платёж — ориентир, округляем
)

type prelimInput struct {
	product shared.ProductView
	amount  int64
	term    int
	report  *shared.CheckReport
	now     time.Time
}

type prelimResult struct {
	outcome    shared.PrelimOutcome
	stopReason string
	minAge     int // для текста BUSINESS_TOO_YOUNG
	limit      int64
	rate       float64
	payment    int64
	amountFits bool
}

// evaluatePrelim — предварительное решение по ответам источников. Порядок стоп-причин — от самой серьёзной:
// недействующей компании и банкротству нет смысла сообщать про возраст бизнеса.
func evaluatePrelim(in prelimInput) prelimResult {
	f := in.report.Facts
	minAge := max(minBusinessAgeMonths, in.product.MinBusinessAgeMonths)

	switch {
	case f.Registration == nil || f.Registration.State != stateActive:
		return prelimResult{outcome: shared.PrelimStop, stopReason: stopCompanyInactive}
	case f.Courts != nil && f.Courts.Bankruptcy, f.FNS != nil && f.FNS.Amount > criticalTaxDebt:
		return prelimResult{outcome: shared.PrelimStop, stopReason: stopCriticalRisk}
	case f.BKI != nil && f.BKI.MaxOverdueDays > maxOverdueDays:
		return prelimResult{outcome: shared.PrelimStop, stopReason: stopBKIOverdue}
	case monthsBetween(f.Registration.RegDate, in.now) < minAge:
		return prelimResult{outcome: shared.PrelimStop, stopReason: stopBusinessTooYoung, minAge: minAge}
	}

	var income int64
	if f.Turnover != nil {
		income = f.Turnover.AvgMonthlyIncome
	}
	limit := int64(float64(income)*12*limitTurnoverShare) / limitStep * limitStep
	if in.product.AmountMax > 0 {
		limit = min(limit, in.product.AmountMax)
	}

	var warns int
	for _, r := range in.report.Results {
		if r.Status == shared.CheckWarn {
			warns++
		}
	}
	rate := in.product.RateMin + float64(warns)*warnRateMarkup
	if in.product.RateMax > 0 {
		rate = math.Min(rate, in.product.RateMax)
	}

	payment := finance.Annuity(in.amount, rate, in.term)
	return prelimResult{
		outcome:    shared.PrelimLimit,
		limit:      limit,
		rate:       rate,
		payment:    int64(math.Round(float64(payment)/paymentStep)) * paymentStep,
		amountFits: in.amount <= limit,
	}
}

// monthsBetween — полных месяцев от from до to.
func monthsBetween(from, to time.Time) int {
	m := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month())
	if to.Day() < from.Day() {
		m--
	}
	return m
}

// stopTexts — объяснение стоп-отказа и что делать клиенту.
func stopTexts(reason string, minAge int) (text string, whatToDo []string) {
	switch reason {
	case stopCompanyInactive:
		return "По данным реестра компания не действует или ликвидируется",
			[]string{"Если данные устарели, обратитесь в банк, чтобы обновить сведения о компании"}
	case stopBusinessTooYoung:
		return fmt.Sprintf("Бизнес зарегистрирован меньше %d месяцев назад", minAge),
			[]string{fmt.Sprintf("Вернуться, когда бизнесу будет %d месяцев", minAge), "Посмотреть экспресс-кредит после этого срока"}
	case stopBKIOverdue:
		return "В кредитной истории есть просрочка больше 30 дней",
			[]string{"Погасить просроченную задолженность", "Подать заявку снова после обновления кредитной истории"}
	case stopCriticalRisk:
		return "По данным реестров у компании критический риск: банкротство или крупный налоговый долг",
			[]string{"Обсудить варианты с менеджером банка"}
	}
	return "", nil
}
