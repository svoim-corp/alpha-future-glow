package servicing

import (
	"slices"
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// testLoan — кредит 1 200 000 ₽ на 24 мес. под 21,5%, выданный 20.10.2026, и его график.
func testLoan() (*loanRow, []payment) {
	issued := date(2026, 10, 20)
	monthly, rows := buildSchedule(1_200_000, 21.5, 24, 0, issued)
	return &loanRow{ID: "l-1", Amount: 1_200_000, Rate: 21.5, Term: 24, MonthlyPayment: monthly, IssuedAt: issued, Status: loanActive}, rows
}

func at(t time.Time) time.Time { return t.Add(10 * time.Hour) }

func TestPenalties(t *testing.T) {
	_, rows := testLoan()
	first := rows[0] // 20.11.2026, 61 958 ₽
	perDay := roundRub(float64(first.Total) * penaltyRatePerDay)

	tests := []struct {
		name  string
		posts []posting
		today time.Time
		want  int64
	}{
		{"в день платежа пеней нет", nil, first.DueDate, 0},
		{"три дня просрочки", nil, first.DueDate.AddDate(0, 0, 3), 3 * perDay},
		// Половина оплачена 22.11: пени за 21 и 22 — на всю сумму, с 23 — на остаток.
		{"частичная оплата посередине", []posting{
			{N: ptr(1), Component: componentInterest, Amount: first.Interest, PostedAt: at(date(2026, 11, 22))},
		}, date(2026, 11, 25), 2*perDay + 3*roundRub(float64(first.Principal)*penaltyRatePerDay)},
		{"после полной оплаты пени не растут", []posting{
			{N: ptr(1), Component: componentInterest, Amount: first.Interest, PostedAt: at(date(2026, 11, 22))},
			{N: ptr(1), Component: componentPrincipal, Amount: first.Principal, PostedAt: at(date(2026, 11, 22))},
		}, date(2026, 12, 15), 2 * perDay},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := penaltiesAccrued(rows[:1], tt.posts, tt.today); got != tt.want {
				t.Fatalf("пени %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDebtOf(t *testing.T) {
	l, rows := testLoan()

	t.Run("до первого платежа", func(t *testing.T) {
		d := debtOf(l, rows, nil, date(2026, 11, 4)) // 15 дней с выдачи
		want := roundRub(1_200_000 * 21.5 / 100 / 365 * 15)
		if d.principal != 1_200_000 || d.accrued != want || d.interest != want || d.penalties != 0 || d.overdueTotal() != 0 {
			t.Fatalf("%+v, проценты want %d", d, want)
		}
		if d.next == nil || d.next.N != 1 || d.payable() != rows[0].Total {
			t.Fatalf("ближайший платёж %+v, к оплате %d", d.next, d.payable())
		}
	})

	t.Run("просрочка входит в итог один раз", func(t *testing.T) {
		today := date(2026, 11, 25)
		d := debtOf(l, rows, nil, today)
		penalties := penaltiesAccrued(rows, nil, today)
		if d.overduePrincipal != rows[0].Principal || d.overdueInterest != rows[0].Interest || d.penalties != penalties ||
			d.overduePayments != 1 || !d.overdueSince.Equal(date(2026, 11, 21)) {
			t.Fatalf("просрочка: %+v", d)
		}
		accrued := roundRub(float64(1_200_000) * 21.5 / 100 / 365 * 5)
		if d.total() != 1_200_000+rows[0].Interest+accrued+penalties {
			t.Fatalf("итог %d: основной долг, проценты строки, текущие проценты и пени", d.total())
		}
		if d.payable() != d.overdueTotal()+rows[1].Total {
			t.Fatalf("к оплате %d, want просрочка %d + следующая строка %d", d.payable(), d.overdueTotal(), rows[1].Total)
		}
	})

	t.Run("проценты периода не больше процентов строки", func(t *testing.T) {
		d := debtOf(l, rows, nil, date(2026, 11, 19))
		if d.accrued > rows[0].Interest {
			t.Fatalf("начислено %d больше процентов строки %d", d.accrued, rows[0].Interest)
		}
	})
}

func TestAllocate(t *testing.T) {
	l, rows := testLoan()
	today := date(2026, 12, 23) // строки 20.11 и 20.12 просрочены
	d := debtOf(l, rows, nil, today)

	t.Run("очередь policy-1", func(t *testing.T) {
		amount := rows[0].Interest + rows[1].Interest + 1000
		a := allocate(amount, rows, d, at(today))
		if a.interest != rows[0].Interest+rows[1].Interest || a.principal != 1000 || a.penalties != 0 || a.unapplied != 0 {
			t.Fatalf("сначала все просроченные проценты, затем основной долг ранней строки: %+v", a)
		}
		if p := a.posts[2]; *p.N != 1 || p.Component != componentPrincipal {
			t.Fatalf("основной долг — с ранней строки: %+v", p)
		}
	})

	t.Run("пени после просрочки, затем ближайшая строка", func(t *testing.T) {
		amount := d.overdueTotal() + 500
		a := allocate(amount, rows, d, at(today))
		if a.penalties != d.penalties || a.unapplied != 0 {
			t.Fatalf("%+v", a)
		}
		last := a.posts[len(a.posts)-1]
		if *last.N != 3 || last.Component != componentInterest || last.Amount != 500 {
			t.Fatalf("остаток — в проценты ближайшей строки: %+v", last)
		}
		changed := a.apply(slices.Clone(rows))
		if len(changed) != 3 || changed[0].Status != paymentPaid || changed[1].Status != paymentPaid || changed[2].Status != paymentPlanned {
			t.Fatalf("статусы строк: %+v", changed)
		}
	})

	t.Run("сумма сходится до рубля", func(t *testing.T) {
		amount := d.payable() + 777
		a := allocate(amount, rows, d, at(today))
		if a.principal+a.interest+a.penalties+a.unapplied != amount || a.unapplied != 777 {
			t.Fatalf("%+v", a)
		}
	})
}

func TestReschedule(t *testing.T) {
	l, rows := testLoan()
	paidOn := date(2026, 11, 4)
	future := rows // ни один платёж ещё не наступил
	const balance = 1_000_000

	for _, strategy := range []string{strategyReduceTerm, strategyReducePayment} {
		t.Run(strategy, func(t *testing.T) {
			monthly, out := reschedule(future, balance, l.Rate, graceEnd(l), l.MonthlyPayment, strategy, paidOn, 2)
			var principal int64
			for i, r := range out {
				principal += r.Principal
				if r.N != 2+i || !r.DueDate.Equal(future[i].DueDate) || r.Total != r.Principal+r.Interest {
					t.Fatalf("строка %d: %+v", i, r)
				}
			}
			if principal != balance {
				t.Fatalf("основной долг %d, want %d", principal, balance)
			}
			first := roundRub(balance * l.Rate / 100 / daysInYear * float64(daysBetween(paidOn, future[0].DueDate)))
			if out[0].Interest != first {
				t.Fatalf("проценты первой строки %d, want %d — с даты погашения", out[0].Interest, first)
			}
			switch strategy {
			case strategyReduceTerm:
				if monthly != l.MonthlyPayment || len(out) >= len(future) {
					t.Fatalf("платёж %d, строк %d: срок должен сократиться", monthly, len(out))
				}
			case strategyReducePayment:
				if monthly >= l.MonthlyPayment || len(out) != len(future) {
					t.Fatalf("платёж %d, строк %d: платёж должен уменьшиться", monthly, len(out))
				}
			}
		})
	}

	t.Run("отсрочка сохраняется", func(t *testing.T) {
		g := &loanRow{IssuedAt: l.IssuedAt, GraceMonths: 3}
		_, out := reschedule(future, balance, l.Rate, graceEnd(g), l.MonthlyPayment, strategyReducePayment, paidOn, 2)
		for i := range 3 {
			if out[i].Principal != 0 {
				t.Fatalf("строка %d в отсрочку: %+v", i, out[i])
			}
		}
		if out[3].Principal == 0 {
			t.Fatalf("после отсрочки гасится долг: %+v", out[3])
		}
	})
}
