package scoring

import (
	"fmt"
	"math"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Модель rules-v1: база плюс вклад каждого признака (стр. 14 ТЗ). Веса подобраны под демо-данные моков
// (backend/SCENARIOS.md): smirnov и kuznetsov — авторешение «одобрено», morozov и kozlov — к сотруднику, polyakov — автоотказ.
const (
	modelVersion = "rules-v1"
	baseScore    = 500

	autoApproveScore = 700 // не ниже — одобрение без сотрудника
	autoRejectScore  = 450 // ниже — отказ без сотрудника
	mediumRiskScore  = 550

	mismatchTolerance = 0.3 // обороты по счетам и выручка ГИР БО могут расходиться на 30%

	counterAmountShare = 0.8 // иные условия: сумма снижается на 20%
	counterRateMarkup  = 1.5

	courtDefendant = "DEFENDANT"
)

// Коды признаков (API_SPEC.md, 4.12).
const (
	factorBusinessAge       = "BUSINESS_AGE"
	factorBKI               = "BKI_DISCIPLINE"
	factorDebtLoad          = "DEBT_LOAD"
	factorTurnoverStability = "TURNOVER_STABILITY"
	factorCourts            = "COURTS"
	factorFNSDebt           = "FNS_DEBT"
	factorTransparent       = "TRANSPARENT_BUSINESS"
	factorTurnoverMismatch  = "TURNOVER_MISMATCH"
	factorProfitability     = "PROFITABILITY"
)

type scoreInput struct {
	report  *shared.CheckReport
	product shared.ProductView
	amount  int64
	term    int
	prelim  *prelimRow // nil — этап 1 не проходили
	now     time.Time
}

type scoreResult struct {
	score     int
	band      shared.RiskBand
	factors   []shared.ScoreFactor
	eligible  bool
	reason    string
	suggested shared.SuggestedDecision
}

// score — полный скоринг и рекомендация решения. Без ввода-вывода.
func score(in scoreInput) scoreResult {
	f := in.report.Facts
	var factors []shared.ScoreFactor
	add := func(code string, impact int, label string, args ...any) {
		factors = append(factors, shared.ScoreFactor{Code: code, Label: fmt.Sprintf(label, args...), Impact: impact})
	}

	if f.Registration != nil {
		months := monthsBetween(f.Registration.RegDate, in.now)
		switch {
		case months < 12:
			add(factorBusinessAge, -60, "Бизнесу меньше года")
		case months < 36:
			add(factorBusinessAge, 10, "Возраст бизнеса %d мес.", months)
		case months < 84:
			add(factorBusinessAge, 40, "Возраст бизнеса %d лет", months/12)
		default:
			add(factorBusinessAge, 60, "Возраст бизнеса %d лет", months/12)
		}
	}

	var debt int64
	if f.BKI != nil {
		debt = f.BKI.TotalDebt
		switch {
		case f.BKI.MaxOverdueDays > maxOverdueDays:
			add(factorBKI, -150, "Просрочка %d дн.", f.BKI.MaxOverdueDays)
		case f.BKI.MaxOverdueDays > 0:
			add(factorBKI, -60, "Просрочка до 30 дней")
		case len(f.BKI.Loans) == 0:
			add(factorBKI, 0, "Кредитной истории нет")
		default:
			add(factorBKI, 80, "Без просрочек")
		}
	}

	var annual int64
	if f.Turnover != nil {
		annual = f.Turnover.AvgMonthlyIncome * 12
	}
	if annual > 0 {
		load := float64(debt+in.amount) / float64(annual)
		pct := int(math.Round(load * 100))
		switch {
		case load < 0.25:
			add(factorDebtLoad, 40, "Долговая нагрузка %d%% от оборота", pct)
		case load < 0.5:
			add(factorDebtLoad, 0, "Долговая нагрузка %d%% от оборота", pct)
		case load < 0.75:
			add(factorDebtLoad, -70, "Долговая нагрузка %d%% от оборота", pct)
		default:
			add(factorDebtLoad, -150, "Долговая нагрузка %d%% от оборота", pct)
		}
	}

	if f.Turnover != nil && len(f.Turnover.Months) > 0 {
		n := len(f.Turnover.Months)
		cv := variation(f.Turnover.Months)
		switch {
		case n < 6:
			add(factorTurnoverStability, -30, "Обороты по счетам в банке только за %d мес.", n)
		case cv < 0.15:
			add(factorTurnoverStability, 40, "Обороты по счетам в банке стабильны %d мес.", n)
		case cv < 0.35:
			add(factorTurnoverStability, 0, "Обороты по счетам в банке колеблются")
		default:
			add(factorTurnoverStability, -60, "Обороты по счетам в банке нестабильны")
		}
	}

	if f.Courts != nil {
		var sued int
		for _, a := range f.Courts.Arbitration {
			if a.Role == courtDefendant {
				sued++
			}
		}
		switch {
		case sued == 0 && len(f.Courts.Enforcement) == 0:
			add(factorCourts, 10, "Судебных дел нет")
		default:
			add(factorCourts, -20*sued-30*len(f.Courts.Enforcement), "Арбитражных дел: %d, исполнительных производств: %d", sued, len(f.Courts.Enforcement))
		}
	}

	if f.FNS != nil && f.FNS.HasDebt {
		add(factorFNSDebt, -40, "Налоговая задолженность")
	}
	if f.Transparent != nil && (f.Transparent.MassAddress || f.Transparent.MassHead) {
		add(factorTransparent, -40, "Признаки массового адреса или руководителя")
	}

	mismatch := false
	if f.GIRBO != nil && f.GIRBO.Revenue > 0 && annual > 0 {
		ratio := float64(annual)/float64(f.GIRBO.Revenue) - 1
		if math.Abs(ratio) > mismatchTolerance {
			mismatch = true
			add(factorTurnoverMismatch, -60, "Обороты по счетам расходятся с выручкой в ГИР БО на %d%%", int(math.Round(math.Abs(ratio)*100)))
		} else {
			add(factorTurnoverMismatch, 20, "Обороты по счетам подтверждены отчётностью")
		}
		if f.GIRBO.NetProfit < 0 {
			add(factorProfitability, -40, "Убыток по итогам %d года", f.GIRBO.Year)
		} else {
			add(factorProfitability, 20, "Прибыль по итогам %d года", f.GIRBO.Year)
		}
	}

	total := baseScore
	for _, fc := range factors {
		total += fc.Impact
	}
	r := scoreResult{score: total, band: band(total), factors: factors}
	r.suggested = suggest(in, total, factors)
	r.eligible, r.reason = eligibility(in, total, mismatch)
	return r
}

func band(score int) shared.RiskBand {
	switch {
	case score >= autoApproveScore:
		return shared.RiskLow
	case score >= mediumRiskScore:
		return shared.RiskMedium
	}
	return shared.RiskHigh
}

// eligibility — можно ли решить без сотрудника. Сотрудник нужен, если данных не хватает, клиент сам попросил
// сумму выше лимита, обороты не сходятся с отчётностью или балл пограничный.
func eligibility(in scoreInput, total int, mismatch bool) (bool, string) {
	for _, r := range in.report.Results {
		if r.Status == shared.CheckUnavailable {
			return false, "Не все источники ответили"
		}
	}
	switch {
	case in.prelim != nil && in.prelim.Choice != nil && shared.PrelimChoice(*in.prelim.Choice) == shared.ChoiceKeepAmount:
		return false, "Клиент запросил сумму выше предварительного лимита"
	case mismatch:
		return false, "Обороты по счетам расходятся с отчётностью — нужны пояснения клиента"
	case total >= autoApproveScore:
		return true, "Высокий балл"
	case total < autoRejectScore:
		return true, "Низкий балл"
	}
	return false, "Пограничный балл"
}

// suggest — рекомендация сотруднику (и решение, если оно автоматическое).
func suggest(in scoreInput, total int, factors []shared.ScoreFactor) shared.SuggestedDecision {
	rate := in.product.RateMin
	if in.prelim != nil && in.prelim.PrelimRate > 0 {
		rate = in.prelim.PrelimRate
	}
	switch {
	case total < autoRejectScore:
		return shared.SuggestedDecision{Decision: shared.StatusRejected}
	case total >= autoApproveScore:
		return shared.SuggestedDecision{Decision: shared.StatusApproved, Amount: in.amount, Rate: rate}
	}
	for _, fc := range factors {
		if fc.Code == factorDebtLoad && fc.Impact < 0 {
			amount := int64(float64(in.amount)*counterAmountShare) / limitStep * limitStep
			return shared.SuggestedDecision{
				Decision: shared.StatusCounterOffer,
				Amount:   max(amount, in.product.AmountMin),
				Rate:     math.Min(rate+counterRateMarkup, max(in.product.RateMax, rate)),
			}
		}
	}
	return shared.SuggestedDecision{Decision: shared.StatusApproved, Amount: in.amount, Rate: rate}
}

// variation — коэффициент вариации месячных поступлений: насколько обороты «рваные».
func variation(ms []shared.TurnoverMonth) float64 {
	var sum float64
	for _, m := range ms {
		sum += float64(m.Income)
	}
	mean := sum / float64(len(ms))
	if mean == 0 {
		return math.Inf(1)
	}
	var sq float64
	for _, m := range ms {
		d := float64(m.Income) - mean
		sq += d * d
	}
	return math.Sqrt(sq/float64(len(ms))) / mean
}
