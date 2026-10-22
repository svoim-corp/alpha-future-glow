package servicing

import (
	"math"
	"time"
)

// Политика расчётов policy-1 (docs/servicing.md): демо-правила, согласованные с пользователем, а не тарифы банка.
// Функции здесь чистые: состояние кредита на дату today считается по строкам графика и журналу проводок.
const (
	policyVersion     = "policy-1"
	penaltyRatePerDay = 0.001 // 0,1% в день от просроченных основного долга и процентов
	daysInYear        = 365
	day               = 24 * time.Hour
)

// Компоненты проводки.
const (
	componentPrincipal = "PRINCIPAL"
	componentInterest  = "INTEREST"
	componentPenalty   = "PENALTY"
)

// posting — проводка по строке графика n (у пеней строки нет).
type posting struct {
	N         *int      `db:"n"`
	Component string    `db:"component"`
	Amount    int64     `db:"amount"`
	PostedAt  time.Time `db:"posted_at"`
}

func (p payment) remainingPrincipal() int64 { return p.Principal - p.PaidPrincipal }
func (p payment) remainingInterest() int64  { return p.Interest - p.PaidInterest }
func (p payment) remaining() int64          { return p.remainingPrincipal() + p.remainingInterest() }

// dateOf — операционный день момента t (UTC).
func dateOf(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int { return int(dateOf(to).Sub(dateOf(from)) / day) }

func roundRub(v float64) int64 { return int64(math.Round(v)) }

// penaltiesAccrued — пени, начисленные по строкам на конец дня today. Просрочка строки начинается на следующий день
// после даты платежа; пеня за день — на остаток строки на начало дня (проводки раньше этого дня его уменьшают),
// округление до рубля на каждом дневном начислении.
func penaltiesAccrued(rows []payment, posts []posting, today time.Time) int64 {
	paid := map[int][]posting{} // проводки основного долга и процентов по строкам, по времени
	for _, p := range posts {
		if p.N != nil && p.Component != componentPenalty {
			paid[*p.N] = append(paid[*p.N], p)
		}
	}
	var total int64
	for _, r := range rows {
		if !r.DueDate.Before(today) {
			continue
		}
		rem, ps := r.Total, paid[r.N]
		for d := r.DueDate.AddDate(0, 0, 1); !d.After(today); d = d.AddDate(0, 0, 1) {
			for len(ps) > 0 && dateOf(ps[0].PostedAt).Before(d) {
				rem -= ps[0].Amount
				ps = ps[1:]
			}
			if rem <= 0 {
				break
			}
			total += roundRub(float64(rem) * penaltyRatePerDay)
		}
	}
	return total
}

// debt — задолженность кредита на дату today по policy-1.
type debt struct {
	principal int64 // остаток основного долга
	interest  int64 // неоплаченные проценты строк с датой не позже today и проценты текущего периода
	penalties int64 // начисленные минус оплаченные пени
	accrued   int64 // проценты текущего периода (входят в interest)
	prepaid   int64 // проценты текущей строки, оплаченные заранее сверх начисленных по today

	overduePrincipal, overdueInterest int64
	overduePayments                   int
	overdueSince                      *time.Time // первый день самой ранней просрочки

	next    *payment // ближайшая неоплаченная строка с датой не раньше today
	current *payment // первая строка с датой позже today: её период идёт сейчас
	unpaid  bool     // есть неоплаченные строки с датой не позже today
}

func (d debt) total() int64        { return d.principal + d.interest + d.penalties }
func (d debt) overdueTotal() int64 { return d.overduePrincipal + d.overdueInterest + d.penalties }

// payable — сколько можно внести обычной оплатой: просрочка, пени и остаток ближайшей строки.
func (d debt) payable() int64 {
	v := d.overdueTotal()
	if d.next != nil {
		v += d.next.remaining()
	}
	return v
}

// debtOf считает задолженность. Проценты текущего периода — остаток основного долга на начало периода × ставка /
// 365 × дни с даты предыдущей строки (или выдачи), не больше процентов строки графика, минус уже оплаченные.
// Оплаченное сверх начисленного — prepaid: при досрочном погашении оно засчитывается в основной долг.
func debtOf(l *loanRow, rows []payment, posts []posting, today time.Time) debt {
	d := debt{principal: l.Amount, penalties: max(penaltiesAccrued(rows, posts, today)-l.PenaltiesPaid, 0)}
	periodStart, base := dateOf(l.IssuedAt), l.Amount
	for i := range rows {
		r := &rows[i]
		d.principal -= r.PaidPrincipal
		if d.next == nil && r.remaining() > 0 && !r.DueDate.Before(today) {
			d.next = r
		}
		if r.DueDate.After(today) {
			if d.current == nil {
				d.current = r
			}
			continue
		}
		periodStart, base = r.DueDate, base-r.PaidPrincipal
		d.interest += r.remainingInterest()
		if r.remaining() == 0 {
			continue
		}
		d.unpaid = true
		if r.DueDate.Before(today) {
			d.overduePrincipal += r.remainingPrincipal()
			d.overdueInterest += r.remainingInterest()
			d.overduePayments++
			if d.overdueSince == nil {
				since := r.DueDate.AddDate(0, 0, 1)
				d.overdueSince = &since
			}
		}
	}
	if c := d.current; c != nil {
		accrued := min(roundRub(float64(base)*l.Rate/100/daysInYear*float64(daysBetween(periodStart, today))), c.Interest)
		d.accrued, d.prepaid = max(accrued-c.PaidInterest, 0), max(c.PaidInterest-accrued, 0)
		d.interest += d.accrued
	}
	return d
}

// allocation — распределение поступившей суммы.
type allocation struct {
	principal, interest, penalties, unapplied int64
	posts                                     []posting
}

// allocate распределяет amount по порядку policy-1: просроченные проценты, просроченный основной долг (от ранних строк
// к поздним), пени, проценты и основной долг ближайшей строки. Остаток — нераспределённые средства.
func allocate(amount int64, rows []payment, d debt, at time.Time) allocation {
	a := allocation{}
	left := amount
	take := func(r *payment, component string, due int64) {
		v := min(left, due)
		if v <= 0 {
			return
		}
		left -= v
		n := r.N
		a.posts = append(a.posts, posting{N: &n, Component: component, Amount: v, PostedAt: at})
		if component == componentPrincipal {
			a.principal += v
		} else {
			a.interest += v
		}
	}
	today := dateOf(at)
	var overdue []*payment
	for i := range rows {
		if r := &rows[i]; r.DueDate.Before(today) && r.remaining() > 0 {
			overdue = append(overdue, r)
		}
	}
	for _, r := range overdue {
		take(r, componentInterest, r.remainingInterest())
	}
	for _, r := range overdue {
		take(r, componentPrincipal, r.remainingPrincipal())
	}
	if v := min(left, d.penalties); v > 0 {
		left -= v
		a.penalties = v
		a.posts = append(a.posts, posting{Component: componentPenalty, Amount: v, PostedAt: at})
	}
	if d.next != nil {
		take(d.next, componentInterest, d.next.remainingInterest())
		take(d.next, componentPrincipal, d.next.remainingPrincipal())
	}
	a.unapplied = left
	return a
}

// apply переносит проводки на строки графика: оплаченные суммы и статус PAID полностью погашенных строк. Строки rows
// меняются на месте; возвращаются изменённые.
func (a allocation) apply(rows []payment) []payment {
	byN := map[int]int{}
	for i, r := range rows {
		byN[r.N] = i
	}
	changed := map[int]bool{}
	for _, p := range a.posts {
		if p.N == nil {
			continue
		}
		r := &rows[byN[*p.N]]
		if p.Component == componentPrincipal {
			r.PaidPrincipal += p.Amount
		} else {
			r.PaidInterest += p.Amount
		}
		if r.remaining() == 0 {
			r.Status = paymentPaid
		}
		changed[r.N] = true
	}
	var out []payment
	for _, r := range rows {
		if changed[r.N] {
			out = append(out, r)
		}
	}
	return out
}
