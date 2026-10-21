package scoring

import (
	"slices"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var (
	stable   = []float64{0.97, 1.02, 0.95, 1.04, 0.99, 1.01, 1.03, 0.98, 1.00, 1.02, 0.97, 1.02}
	unstable = []float64{1.6, 0.3, 1.4, 0.2, 1.5, 0.6, 1.7, 0.4, 1.3, 0.5, 1.8, 0.7}
)

func turnover(base int64, pattern []float64) *shared.TurnoverReport {
	t := &shared.TurnoverReport{AvgMonthlyIncome: base}
	for _, k := range pattern {
		t.Months = append(t.Months, shared.TurnoverMonth{Income: int64(float64(base) * k)})
	}
	return t
}

func loans(debt int64, overdue int) *shared.BKIReport {
	return &shared.BKIReport{Loans: []shared.BKILoan{{Balance: debt, OverdueDays: overdue}}, TotalDebt: debt, MaxOverdueDays: overdue}
}

// company — ответы источников, как у компаний из моков (backend/SCENARIOS.md).
type company struct {
	reg      string
	turnover *shared.TurnoverReport
	bki      *shared.BKIReport
	girbo    *shared.GIRBOReport
	courts   *shared.CourtRecords
	fns      *shared.TaxDebt
	mass     bool
}

func (c company) report() *shared.CheckReport {
	courts := c.courts
	if courts == nil {
		courts = &shared.CourtRecords{}
	}
	fns := c.fns
	if fns == nil {
		fns = &shared.TaxDebt{}
	}
	return &shared.CheckReport{Facts: shared.CheckFacts{
		Registration: &shared.Registration{State: "ACTIVE", RegDate: date(c.reg)},
		Turnover:     c.turnover, BKI: c.bki, GIRBO: c.girbo, Courts: courts, FNS: fns,
		Transparent: &shared.TransparentBusiness{MassAddress: c.mass},
	}}
}

var (
	smirnov = company{reg: "2015-03-12", turnover: turnover(1_840_000, stable), bki: loans(900_000, 0),
		girbo:  &shared.GIRBOReport{Year: 2025, Revenue: 21_400_000, NetProfit: 2_100_000},
		courts: &shared.CourtRecords{Arbitration: []shared.ArbitrationCase{{Role: "DEFENDANT", Amount: 120_000}}}}
	kuznetsov = company{reg: "2018-05-14", turnover: turnover(900_000, stable), bki: loans(300_000, 0)}
	morozov   = company{reg: "2013-08-14", turnover: turnover(2_200_000, stable), bki: loans(14_000_000, 0),
		girbo: &shared.GIRBOReport{Year: 2025, Revenue: 25_900_000, NetProfit: 1_100_000}}
	kozlov = company{reg: "2011-02-07", turnover: turnover(2_000_000, stable), bki: loans(600_000, 0),
		girbo: &shared.GIRBOReport{Year: 2025, Revenue: 15_200_000, NetProfit: 900_000}}
	polyakov = company{reg: "2024-05-16", turnover: turnover(700_000, unstable), bki: loans(1_900_000, 12),
		girbo:  &shared.GIRBOReport{Year: 2025, Revenue: 7_900_000, NetProfit: -350_000},
		courts: &shared.CourtRecords{Arbitration: []shared.ArbitrationCase{{Role: "DEFENDANT"}, {Role: "DEFENDANT"}}, Enforcement: []shared.Enforcement{{}}},
		fns:    &shared.TaxDebt{HasDebt: true, Amount: 85_000}, mass: true}
	sokolova = company{reg: "2019-03-25", turnover: turnover(520_000, stable), bki: &shared.BKIReport{},
		girbo: &shared.GIRBOReport{Year: 2025, Revenue: 6_100_000, NetProfit: 700_000}}
)

func TestScore(t *testing.T) {
	keep := string(shared.ChoiceKeepAmount)
	tests := []struct {
		name      string
		c         company
		amount    int64
		prelim    *prelimRow
		band      shared.RiskBand
		eligible  bool
		suggested shared.Status
		factor    string // признак, который обязан быть в разборе
	}{
		{"smirnov: авторешение «одобрено»", smirnov, 1_500_000, &prelimRow{PrelimRate: 20.5}, shared.RiskLow, true, shared.StatusApproved, factorTurnoverMismatch},
		{"kuznetsov, ИП: авторешение «одобрено» без ГИР БО", kuznetsov, 1_000_000, nil, shared.RiskLow, true, shared.StatusApproved, factorBusinessAge},
		{"morozov: долговая нагрузка — иные условия у сотрудника", morozov, 3_000_000, nil, shared.RiskMedium, false, shared.StatusCounterOffer, factorDebtLoad},
		{"kozlov: обороты не сходятся с отчётностью — к сотруднику", kozlov, 2_000_000, nil, shared.RiskMedium, false, shared.StatusApproved, factorTurnoverMismatch},
		{"polyakov: автоотказ", polyakov, 1_000_000, nil, shared.RiskHigh, true, shared.StatusRejected, factorTransparent},
		{"sokolova: оставила сумму выше лимита — к сотруднику", sokolova, 1_500_000, &prelimRow{Choice: &keep}, shared.RiskMedium, false, shared.StatusApproved, factorBKI},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := score(scoreInput{report: tt.c.report(), product: working, amount: tt.amount, term: 24, prelim: tt.prelim, now: today})
			sum := baseScore
			for _, f := range r.factors {
				sum += f.Impact
			}
			if r.score != sum {
				t.Fatalf("балл %d не равен сумме вкладов %d", r.score, sum)
			}
			if r.band != tt.band || r.eligible != tt.eligible || r.suggested.Decision != tt.suggested {
				t.Fatalf("score %d band %s eligible %v (%s) suggested %s; factors %+v", r.score, r.band, r.eligible, r.reason, r.suggested.Decision, r.factors)
			}
			if !slices.ContainsFunc(r.factors, func(f shared.ScoreFactor) bool { return f.Code == tt.factor }) {
				t.Fatalf("нет признака %s: %+v", tt.factor, r.factors)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	r := score(scoreInput{report: morozov.report(), product: working, amount: 1_500_000, term: 24, now: today})
	if s := r.suggested; s.Amount != 1_200_000 || s.Rate != working.RateMin+counterRateMarkup {
		t.Fatalf("иные условия: %+v", s)
	}
	r = score(scoreInput{report: smirnov.report(), product: working, amount: 1_500_000, term: 24, prelim: &prelimRow{PrelimRate: 20.5}, now: today})
	if s := r.suggested; s.Amount != 1_500_000 || s.Rate != 20.5 {
		t.Fatalf("одобрение по ставке этапа 1: %+v", s)
	}
}

func TestScoreUnavailableSource(t *testing.T) {
	rep := smirnov.report()
	rep.Results = []shared.CheckResult{{Source: shared.SourceBKI, Status: shared.CheckUnavailable}}
	r := score(scoreInput{report: rep, product: working, amount: 1_500_000, term: 24, now: today})
	if r.eligible || r.reason != "Не все источники ответили" {
		t.Fatalf("%v %q", r.eligible, r.reason)
	}
}
