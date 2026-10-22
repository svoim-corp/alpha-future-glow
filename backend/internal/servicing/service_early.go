package servicing

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Типы досрочного погашения.
const (
	earlyPartial = "PARTIAL"
	earlyFull    = "FULL"
)

// quoteTTL — сколько действует расчёт. Он также истекает в конце операционного дня: проценты считаются по дату.
const quoteTTL = 15 * time.Minute

type quoteCmd struct {
	Type          string  `json:"type"`
	Amount        *int64  `json:"amount"`
	Strategy      *string `json:"strategy"`
	EffectiveDate *string `json:"effectiveDate"`
}

// Quote рассчитывает досрочное погашение на сегодня и сохраняет расчёт. Денег не списывает.
func (m *Module) Quote(ctx context.Context, companyID, loanID, key string, cmd quoteCmd) (*quoteRow, error) {
	if _, err := m.ownLoan(ctx, companyID, loanID); err != nil {
		return nil, err
	}
	fp := fingerprint(cmd)
	prev, err := m.store.QuoteByKey(ctx, m.deps.DB, loanID, key)
	if err != nil {
		return nil, err
	}
	if prev != nil {
		if prev.Fingerprint != fp {
			return nil, errKeyReused
		}
		return prev, nil
	}
	today, now := m.today(), m.now().UTC()
	if cmd.EffectiveDate != nil && *cmd.EffectiveDate != today.Format(dateLayout) {
		return nil, httpx.BusinessRule("Досрочное погашение возможно только сегодняшней датой")
	}

	var q *quoteRow
	err = db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		l, err := m.store.LockLoan(ctx, tx, loanID)
		if err != nil {
			return err
		}
		if q, err = m.store.QuoteByKey(ctx, tx, loanID, key); err != nil || q != nil {
			return err
		}
		switch open, err := m.store.HasOpenOperation(ctx, tx, loanID); {
		case err != nil:
			return err
		case l.Status != loanActive:
			return errLoanClosed
		case open:
			return errOpInFlight
		}
		rows, posts, err := m.loanState(ctx, tx, loanID)
		if err != nil {
			return err
		}
		if q, err = calcQuote(l, rows, debtOf(l, rows, posts, today), cmd, today); err != nil {
			return err
		}
		q.IdemKey, q.Fingerprint, q.CalculatedAt = key, fp, now
		q.ValidUntil = now.Add(quoteTTL)
		if dayEnd := today.AddDate(0, 0, 1); q.ValidUntil.After(dayEnd) {
			q.ValidUntil = dayEnd
		}
		return m.store.InsertQuote(ctx, tx, q)
	})
	if err != nil {
		return nil, err
	}
	if q.Fingerprint != fp {
		return nil, errKeyReused
	}
	return q, nil
}

// calcQuote — расчёт по policy-1: сначала проценты текущего периода, затем основной долг. Проценты текущей строки,
// оплаченные заранее сверх начисленных по сегодня, засчитываются в основной долг: за непрошедшие дни клиент проценты
// не платит. Досрочно гасить можно, только когда строки с датой сегодня и раньше оплачены и пеней нет.
func calcQuote(l *loanRow, rows []payment, d debt, cmd quoteCmd, today time.Time) (*quoteRow, error) {
	switch {
	case d.unpaid || d.overdueTotal() > 0:
		return nil, httpx.BusinessRule("Сначала внесите платёж по графику: досрочное погашение доступно без просрочки и неоплаченного платежа на сегодня")
	case d.current == nil || d.principal == 0:
		return nil, httpx.BusinessRule("Основной долг погашен, досрочно гасить нечего")
	}
	q := &quoteRow{
		LoanID: l.ID, Type: cmd.Type, LoanVersion: l.Version, ScheduleVersion: l.ScheduleVersion, EffectiveDate: today,
		AllocInterest: d.accrued, InterestCredit: d.prepaid, MonthlyAfter: l.MonthlyPayment, Preview: db.JSON[[]previewRow]{V: []previewRow{}},
	}
	if cmd.Type == earlyFull {
		q.AllocPrincipal = d.principal - d.prepaid
		q.AmountToDebit = q.AllocPrincipal + d.accrued
		return q, nil
	}

	amount := *cmd.Amount
	switch {
	case amount <= d.accrued:
		return nil, httpx.BusinessRule("Сумма должна быть больше начисленных процентов: " + finance.Rub(d.accrued))
	case amount-d.accrued+d.prepaid >= d.principal:
		return nil, httpx.BusinessRule("Сумма покрывает весь долг — выберите полное погашение")
	}
	q.Strategy, q.AmountToDebit, q.AllocPrincipal = cmd.Strategy, amount, amount-d.accrued
	var future []payment
	for _, r := range rows {
		if r.N >= d.current.N {
			future = append(future, r)
		}
	}
	monthly, next := reschedule(future, d.principal-q.AllocPrincipal-q.InterestCredit, l.Rate, graceEnd(l), l.MonthlyPayment, *cmd.Strategy, today, d.current.N+1)
	q.MonthlyAfter = monthly
	for _, r := range next {
		q.Preview.V = append(q.Preview.V, previewRow{N: r.N, Date: r.DueDate.Format(dateLayout), Principal: r.Principal, Interest: r.Interest})
	}
	return q, nil
}

type earlyCmd struct {
	QuoteID   string `json:"quoteId"`
	AccountID string `json:"accountId"`
}

// ConfirmEarly подтверждает расчёт и запрашивает списание. До списания заново проверяются полномочия, счёт,
// версия кредита, незавершённые операции и сам расчёт.
func (m *Module) ConfirmEarly(ctx context.Context, a actor, loanID string, version int, key string, cmd earlyCmd) (*opRow, error) {
	if _, err := m.ownLoan(ctx, a.companyID, loanID); err != nil {
		return nil, err
	}
	if err := m.authorize(ctx, a); err != nil {
		return nil, err
	}
	fp := fingerprint(cmd)
	prev, err := m.store.OperationByKey(ctx, m.deps.DB, loanID, key)
	if err != nil {
		return nil, err
	}
	if prev != nil {
		return repeated(prev, opEarly, fp)
	}
	if q, err := m.store.Quote(ctx, m.deps.DB, cmd.QuoteID); err != nil || q == nil || q.LoanID != loanID {
		if err != nil {
			return nil, err
		}
		return nil, httpx.NotFound("Расчёт досрочного погашения не найден")
	}
	acc, err := m.account(ctx, a.companyID, cmd.AccountID)
	if err != nil {
		return nil, err
	}

	var op *opRow
	err = db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		l, prev, err := m.reserve(ctx, tx, loanID, version, key, opEarly, fp)
		if err != nil || prev != nil {
			op = prev
			return err
		}
		q, err := m.store.Quote(ctx, tx, cmd.QuoteID)
		if err != nil {
			return err
		}
		now := m.now().UTC()
		switch {
		case q.UsedAt != nil:
			return httpx.BusinessRule("Расчёт уже использован, рассчитайте досрочное погашение заново")
		case !now.Before(q.ValidUntil):
			return httpx.BusinessRule("Срок действия расчёта истёк, рассчитайте досрочное погашение заново")
		case q.LoanVersion != l.Version:
			return httpx.BusinessRule("Кредит изменился после расчёта, рассчитайте досрочное погашение заново")
		}
		if err := m.store.UseQuote(ctx, tx, q.ID, now); err != nil {
			return err
		}
		op = &opRow{
			Type: opEarly, AccountID: acc.ID, AccountNumber: acc.Number, RequestedAmount: q.AmountToDebit, QuoteID: &q.ID,
			IdemKey: key, Fingerprint: fp,
		}
		return m.newOperation(ctx, tx, l, op)
	})
	if err != nil {
		return nil, err
	}
	if op.Status == opPending && op.ExternalID == nil {
		m.submit(ctx, op, a.companyID)
	}
	return op, nil
}

// applyQuote применяет подтверждённый расчёт: строки позже даты исполнения уходят в архив версии графика, их место
// занимают оплаченная строка досрочного погашения и новый график. Переплаченные проценты переносятся в основной долг
// парой проводок. Если с расчёта кредит изменился или банк провёл другую сумму, график не трогаем: деньги остаются
// нераспределёнными до сверки банком.
func (m *Module) applyQuote(ctx context.Context, tx *sqlx.Tx, l *loanRow, rows []payment, op *opRow, res *shared.DebitResult) (allocation, error) {
	q, err := m.store.Quote(ctx, tx, *op.QuoteID)
	if err != nil {
		return allocation{}, err
	}
	i := -1
	for j, r := range rows {
		if r.DueDate.After(q.EffectiveDate) {
			i = j
			break
		}
	}
	if res.PostedAmount != q.AmountToDebit || l.Version != op.LoanVersion || l.ScheduleVersion != q.ScheduleVersion || i < 0 {
		slog.Warn("early repayment not applied", "operationId", op.ID, "posted", res.PostedAmount, "quoted", q.AmountToDebit)
		return allocation{unapplied: res.PostedAmount}, nil
	}

	cur := rows[i]
	if err := m.store.ArchiveFrom(ctx, tx, l.ID, l.ScheduleVersion, cur.N); err != nil {
		return allocation{}, err
	}
	early := payment{
		N: cur.N, DueDate: q.EffectiveDate, Principal: cur.PaidPrincipal + q.AllocPrincipal + q.InterestCredit,
		Interest: cur.PaidInterest + q.AllocInterest - q.InterestCredit,
		Early:    true, Status: paymentPaid,
	}
	early.Total, early.PaidPrincipal, early.PaidInterest = early.Principal+early.Interest, early.Principal, early.Interest
	next := []payment{early}
	for _, p := range q.Preview.V {
		due, err := time.Parse(dateLayout, p.Date)
		if err != nil {
			return allocation{}, err
		}
		next = append(next, payment{N: p.N, DueDate: due, Principal: p.Principal, Interest: p.Interest, Total: p.Principal + p.Interest, Status: paymentPlanned})
	}
	if err := m.store.InsertPayments(ctx, tx, l.ID, next); err != nil {
		return allocation{}, err
	}
	l.ScheduleVersion++
	l.MonthlyPayment = q.MonthlyAfter

	a := allocation{principal: q.AllocPrincipal, interest: q.AllocInterest}
	n := cur.N
	for _, p := range []posting{
		{N: &n, Component: componentInterest, Amount: q.AllocInterest},
		{N: &n, Component: componentPrincipal, Amount: q.AllocPrincipal},
		{N: &n, Component: componentInterest, Amount: -q.InterestCredit},
		{N: &n, Component: componentPrincipal, Amount: q.InterestCredit},
	} {
		if p.Amount != 0 {
			p.PostedAt = res.PostedAt
			a.posts = append(a.posts, p)
		}
	}
	return a, nil
}

// CloseLimit закрывает овердрафт с нулевой задолженностью: нулевой долг сам лимит не закрывает, решение — клиента.
func (m *Module) CloseLimit(ctx context.Context, a actor, loanID string, version int) error {
	if _, err := m.ownLoan(ctx, a.companyID, loanID); err != nil {
		return err
	}
	if err := m.authorize(ctx, a); err != nil {
		return err
	}
	return db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		l, err := m.store.LockLoan(ctx, tx, loanID)
		if err != nil {
			return err
		}
		switch open, err := m.store.HasOpenOperation(ctx, tx, loanID); {
		case err != nil:
			return err
		case l.Status != loanActive:
			return httpx.InvalidTransition("Кредит уже закрыт")
		case l.Version != version:
			return errLoanVersion
		case !l.Revolving:
			return httpx.BusinessRule("Кредит закрывается автоматически после полного погашения")
		case open:
			return errOpInFlight
		}
		rows, posts, err := m.loanState(ctx, tx, loanID)
		if err != nil {
			return err
		}
		if total := debtOf(l, rows, posts, m.today()).total(); total > 0 {
			return httpx.BusinessRule("Закрыть лимит можно после полного погашения задолженности: осталось " + finance.Rub(total))
		}
		now, reason := m.now().UTC(), closeLimit
		l.Status, l.ClosedAt, l.CloseReason = loanClosed, &now, &reason
		l.Version++
		return m.store.UpdateLoan(ctx, tx, l)
	})
}
