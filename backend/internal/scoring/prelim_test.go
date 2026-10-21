package scoring

import (
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var today = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

var working = shared.ProductView{
	Code: "WORKING", AmountMin: 300_000, AmountMax: 10_000_000, TermMin: 6, TermMax: 36, RateMin: 19, RateMax: 26,
	MinBusinessAgeMonths: 6, GuarantorForLLC: true,
}

func date(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }

// facts — «чистая» компания: действует с 2015 года, оборот 1,84 млн, кредитная история без просрочек.
func facts(mod func(*shared.CheckFacts)) *shared.CheckReport {
	f := shared.CheckFacts{
		Registration: &shared.Registration{State: "ACTIVE", RegDate: date("2015-03-12")},
		Turnover:     &shared.TurnoverReport{AvgMonthlyIncome: 1_840_000},
		BKI:          &shared.BKIReport{},
		FNS:          &shared.TaxDebt{},
		Courts:       &shared.CourtRecords{},
	}
	if mod != nil {
		mod(&f)
	}
	return &shared.CheckReport{Facts: f}
}

func warns(n int) []shared.CheckResult {
	out := make([]shared.CheckResult, n)
	for i := range out {
		out[i].Status = shared.CheckWarn
	}
	return out
}

func TestEvaluatePrelim(t *testing.T) {
	tests := []struct {
		name    string
		product shared.ProductView
		amount  int64
		report  *shared.CheckReport
		outcome shared.PrelimOutcome
		stop    string
		limit   int64
		fits    bool
		rate    float64
	}{
		{name: "smirnov: сумма в лимите", product: working, amount: 1_500_000, report: facts(nil),
			outcome: shared.PrelimLimit, limit: 3_300_000, fits: true, rate: 19},
		{name: "sokolova: сумма выше лимита", product: working, amount: 1_500_000,
			report:  facts(func(f *shared.CheckFacts) { f.Turnover.AvgMonthlyIncome = 519_916 }),
			outcome: shared.PrelimLimit, limit: 900_000, fits: false, rate: 19},
		{name: "лимит не выше рамок продукта", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.Turnover.AvgMonthlyIncome = 100_000_000 }),
			outcome: shared.PrelimLimit, limit: 10_000_000, fits: true, rate: 19},
		{name: "morozov: большой долг лимит не режет", product: working, amount: 1_500_000,
			report:  facts(func(f *shared.CheckFacts) { f.BKI.TotalDebt = 14_000_000; f.Turnover.AvgMonthlyIncome = 2_200_000 }),
			outcome: shared.PrelimLimit, limit: 3_900_000, fits: true, rate: 19},
		{name: "предупреждения повышают ставку", product: working, amount: 1_000_000,
			report:  func() *shared.CheckReport { r := facts(nil); r.Results = warns(2); return r }(),
			outcome: shared.PrelimLimit, limit: 3_300_000, fits: true, rate: 22},
		{name: "ставка не выше рамок продукта", product: working, amount: 1_000_000,
			report:  func() *shared.CheckReport { r := facts(nil); r.Results = warns(10); return r }(),
			outcome: shared.PrelimLimit, limit: 3_300_000, fits: true, rate: 26},
		{name: "polyakov: мелкая просрочка и налоговый долг — не стоп", product: working, amount: 1_000_000,
			report: facts(func(f *shared.CheckFacts) {
				f.BKI.MaxOverdueDays = 12
				f.FNS = &shared.TaxDebt{HasDebt: true, Amount: 85_000}
				f.Registration.RegDate = date("2024-05-16")
			}),
			outcome: shared.PrelimLimit, limit: 3_300_000, fits: true, rate: 19},

		{name: "gromov: ликвидация", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.Registration.State = "LIQUIDATING" }),
			outcome: shared.PrelimStop, stop: stopCompanyInactive},
		{name: "нет записи в реестре", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.Registration = nil }),
			outcome: shared.PrelimStop, stop: stopCompanyInactive},
		{name: "zaitsev: банкротство", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.Courts.Bankruptcy = true; f.BKI.MaxOverdueDays = 20 }),
			outcome: shared.PrelimStop, stop: stopCriticalRisk},
		{name: "крупный налоговый долг", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.FNS = &shared.TaxDebt{HasDebt: true, Amount: 2_000_000} }),
			outcome: shared.PrelimStop, stop: stopCriticalRisk},
		{name: "belov: просрочка 95 дней", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.BKI.MaxOverdueDays = 95 }),
			outcome: shared.PrelimStop, stop: stopBKIOverdue},
		{name: "просрочка ровно 30 дней — не стоп", product: working, amount: 1_000_000,
			report:  facts(func(f *shared.CheckFacts) { f.BKI.MaxOverdueDays = 30 }),
			outcome: shared.PrelimLimit, limit: 3_300_000, fits: true, rate: 19},
		{name: "orlova: бизнесу 3 месяца", product: working, amount: 500_000,
			report:  facts(func(f *shared.CheckFacts) { f.Registration.RegDate = date("2026-07-01") }),
			outcome: shared.PrelimStop, stop: stopBusinessTooYoung},
		{name: "продукт требует 24 месяца", product: func() shared.ProductView { p := working; p.MinBusinessAgeMonths = 24; return p }(), amount: 500_000,
			report:  facts(func(f *shared.CheckFacts) { f.Registration.RegDate = date("2025-01-10") }),
			outcome: shared.PrelimStop, stop: stopBusinessTooYoung},
		{name: "ликвидация важнее банкротства и возраста", product: working, amount: 500_000,
			report: facts(func(f *shared.CheckFacts) {
				f.Registration.State, f.Registration.RegDate, f.Courts.Bankruptcy = "LIQUIDATING", date("2026-09-01"), true
			}),
			outcome: shared.PrelimStop, stop: stopCompanyInactive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluatePrelim(prelimInput{product: tt.product, amount: tt.amount, term: 24, report: tt.report, now: today})
			if got.outcome != tt.outcome || got.stopReason != tt.stop {
				t.Fatalf("outcome %s %q, want %s %q", got.outcome, got.stopReason, tt.outcome, tt.stop)
			}
			if tt.outcome == shared.PrelimStop {
				return
			}
			if got.limit != tt.limit || got.amountFits != tt.fits || got.rate != tt.rate || got.payment <= 0 || got.payment%paymentStep != 0 {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestStopTexts(t *testing.T) {
	for _, r := range []string{stopCompanyInactive, stopBusinessTooYoung, stopBKIOverdue, stopCriticalRisk} {
		text, todo := stopTexts(r, 6)
		if text == "" || len(todo) == 0 {
			t.Errorf("%s: нет текста", r)
		}
	}
	if text, _ := stopTexts(stopBusinessTooYoung, 24); text != "Бизнес зарегистрирован меньше 24 месяцев назад" {
		t.Errorf("%q", text)
	}
}

func TestMonthsBetween(t *testing.T) {
	tests := []struct {
		from string
		want int
	}{
		{"2026-07-01", 3}, {"2026-04-21", 6}, {"2026-04-22", 5}, {"2015-03-12", 139},
	}
	for _, tt := range tests {
		if got := monthsBetween(date(tt.from), today); got != tt.want {
			t.Errorf("%s: %d, want %d", tt.from, got, tt.want)
		}
	}
}
