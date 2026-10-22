package servicing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Типы и статусы денежных операций (docs/servicing.md).
const (
	opRegular = "REGULAR"
	opEarly   = "EARLY"

	opPending    = "PENDING"    // создана, банк поручение ещё не принял
	opProcessing = "PROCESSING" // банк принял поручение
	opSucceeded  = "SUCCEEDED"  // деньги подтверждены и учтены в кредите
	opFailed     = "FAILED"     // банк отказал, долг не изменился
)

// Закрытие кредита.
const (
	loanClosed             = "CLOSED"
	closeScheduleCompleted = "SCHEDULE_COMPLETED"
	closeEarlyRepayment    = "EARLY_REPAYMENT"
	closeLimit             = "LIMIT_CLOSED"
)

const currencyRUB = "RUB"

// actor — клиент, от имени которого выполняется команда.
type actor struct{ companyID, personINN string }

// Ошибки команд по кредиту. Общие httpx.IfMatch и VersionConflict говорят о заявке, поэтому тексты свои.
var (
	errLoanNotFound = httpx.NotFound("Кредит не найден")
	errKeyReused    = httpx.Validation("Ключ идемпотентности уже использован для другого запроса")
	errLoanClosed   = httpx.InvalidTransition("Кредит закрыт")
	errOpInFlight   = httpx.InvalidTransition("По кредиту уже выполняется операция, дождитесь её результата")
	errNoAuthority  = httpx.Forbidden("Списывать деньги со счёта может руководитель компании или представитель по действующей МЧД")
	errLoanVersion  = &httpx.Error{Status: http.StatusPreconditionFailed, Code: httpx.CodeVersionConflict,
		Message: "Кредит изменился, обновите данные и повторите"}
)

// ownLoan — кредит компании. Чужой — 404: существование кредитов других компаний не раскрывается.
func (m *Module) ownLoan(ctx context.Context, companyID, id string) (*loanRow, error) {
	l, err := m.store.Loan(ctx, id)
	if err != nil {
		return nil, err
	}
	if l == nil || l.CompanyID != companyID {
		return nil, errLoanNotFound
	}
	return l, nil
}

// authorize — списание разрешено руководителю или представителю с действующей МЧД.
func (m *Module) authorize(ctx context.Context, a actor) error {
	v, err := m.ports.Authorities.Resolve(ctx, a.companyID, a.personINN)
	if err != nil {
		return err
	}
	if v.Type == shared.AuthorityNone {
		return errNoAuthority
	}
	return nil
}

// companyINN — ИНН компании: по нему АБС знает счета клиента.
func (m *Module) companyINN(ctx context.Context, companyID string) (string, error) {
	c, err := m.ports.Companies.GetByID(ctx, companyID)
	if err != nil {
		return "", err
	}
	if c == nil {
		return "", fmt.Errorf("company %s not found", companyID)
	}
	return c.INN, nil
}

// account — счёт компании для списания. Номер счёта клиент не передаёт: только id из списка АБС.
func (m *Module) account(ctx context.Context, companyID, accountID string) (*shared.BankAccount, error) {
	inn, err := m.companyINN(ctx, companyID)
	if err != nil {
		return nil, err
	}
	accs, err := m.ports.Debits.Accounts(ctx, inn)
	if err != nil {
		return nil, err
	}
	for _, a := range accs {
		if a.ID != accountID {
			continue
		}
		if !a.Active || a.Currency != currencyRUB {
			return nil, httpx.BusinessRule("С этого счёта списать нельзя: он заблокирован или не в рублях")
		}
		return &a, nil
	}
	return nil, httpx.Validation("", httpx.FieldError{Field: "accountId", Code: "INVALID", Message: "Счёт не найден среди счетов компании"})
}

// fingerprint — отпечаток тела команды: тот же ключ с другим телом отклоняется.
func fingerprint(v any) string {
	b, _ := json.Marshal(v) // структуры из строк и чисел маршалятся всегда
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// repeated — сохранённый результат той же команды; распознаётся до проверки версии, которую команда уже изменила.
func repeated(prev *opRow, typ, fp string) (*opRow, error) {
	if prev.Type != typ || prev.Fingerprint != fp {
		return nil, errKeyReused
	}
	return prev, nil
}

// plain — строки графика без данных кредита: для расчётов policy-1.
func plain(rows []paymentRow) []payment {
	out := make([]payment, len(rows))
	for i, r := range rows {
		out[i] = r.payment
	}
	return out
}

// loanState — график и проводки кредита в транзакции или снимке.
func (m *Module) loanState(ctx context.Context, q sqlx.QueryerContext, loanID string) ([]payment, []posting, error) {
	rows, err := m.store.LoanPayments(ctx, q, loanID)
	if err != nil {
		return nil, nil, err
	}
	posts, err := m.store.LoanPostings(ctx, q, loanID)
	if err != nil {
		return nil, nil, err
	}
	return plain(rows), posts, nil
}

// reserve — общие проверки денежной команды под блокировкой кредита. (nil, prev) — команда уже выполнялась.
func (m *Module) reserve(ctx context.Context, tx *sqlx.Tx, loanID string, version int, key, typ, fp string) (*loanRow, *opRow, error) {
	l, err := m.store.LockLoan(ctx, tx, loanID)
	if err != nil {
		return nil, nil, err
	}
	// Параллельный повтор той же команды ждал блокировку и теперь видит результат первой.
	prev, err := m.store.OperationByKey(ctx, tx, loanID, key)
	if err != nil {
		return nil, nil, err
	}
	if prev != nil {
		op, err := repeated(prev, typ, fp)
		return nil, op, err
	}
	switch open, err := m.store.HasOpenOperation(ctx, tx, loanID); {
	case err != nil:
		return nil, nil, err
	case l.Status != loanActive:
		return nil, nil, errLoanClosed
	case l.Version != version:
		return nil, nil, errLoanVersion
	case open:
		return nil, nil, errOpInFlight
	}
	return l, nil, nil
}

// newOperation — операция PENDING; версия кредита растёт: незавершённая операция меняет его финансовое состояние.
func (m *Module) newOperation(ctx context.Context, tx *sqlx.Tx, l *loanRow, op *opRow) error {
	l.Version++
	if err := m.store.UpdateLoan(ctx, tx, l); err != nil {
		return err
	}
	op.LoanID, op.Status, op.LoanVersion, op.ScheduleVersion, op.LoanStatus = l.ID, opPending, l.Version, l.ScheduleVersion, l.Status
	op.RequestedAt = m.now().UTC()
	return m.store.InsertOperation(ctx, tx, op)
}

type repaymentCmd struct {
	AccountID string `json:"accountId"`
	Amount    int64  `json:"amount"`
}

// CreateRepayment — обычная оплата: операция PENDING и поручение в банк. Долг уменьшится после подтверждения.
func (m *Module) CreateRepayment(ctx context.Context, a actor, loanID string, version int, key string, cmd repaymentCmd) (*opRow, error) {
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
		return repeated(prev, opRegular, fp)
	}
	acc, err := m.account(ctx, a.companyID, cmd.AccountID)
	if err != nil {
		return nil, err
	}

	var op *opRow
	err = db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		l, prev, err := m.reserve(ctx, tx, loanID, version, key, opRegular, fp)
		if err != nil || prev != nil {
			op = prev
			return err
		}
		rows, posts, err := m.loanState(ctx, tx, loanID)
		if err != nil {
			return err
		}
		switch payable := debtOf(l, rows, posts, m.today()).payable(); {
		case payable == 0:
			return httpx.BusinessRule("Сейчас оплачивать нечего: ближайший платёж уже погашен")
		case cmd.Amount > payable:
			return httpx.BusinessRule("Сумма больше задолженности к оплате: не больше " + finance.Rub(payable) +
				". Будущие платежи можно погасить досрочно")
		}
		op = &opRow{Type: opRegular, AccountID: acc.ID, AccountNumber: acc.Number, RequestedAmount: cmd.Amount, IdemKey: key, Fingerprint: fp}
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

// submit передаёт поручение в банк. Сбой не означает отказ: операция остаётся PENDING, сверщик передаст её заново
// с тем же id, и банк не спишет дважды.
func (m *Module) submit(ctx context.Context, op *opRow, companyID string) {
	inn, err := m.companyINN(ctx, companyID)
	if err == nil {
		var ext string
		if ext, err = m.ports.Debits.Debit(ctx, shared.DebitRequest{RequestID: op.ID, INN: inn, AccountID: op.AccountID, Amount: op.RequestedAmount}); err == nil {
			now := m.now().UTC()
			if err = m.store.AcceptOperation(ctx, op.ID, ext, now); err == nil {
				op.Status, op.ExternalID, op.UpdatedAt = opProcessing, &ext, now
				return
			}
		}
	}
	slog.Warn("submit debit", "operationId", op.ID, "err", err)
}

// Reconcile доводит незавершённую операцию по ответу банка и возвращает её текущее состояние. Подтверждённое
// списание учитывается в кредите один раз: повторная обработка видит итоговый статус под блокировкой.
func (m *Module) Reconcile(ctx context.Context, opID string) (*opRow, error) {
	op, err := m.store.Operation(ctx, m.deps.DB, opID)
	if err != nil || op == nil || op.Status == opSucceeded || op.Status == opFailed {
		return op, err
	}
	l, err := m.store.Loan(ctx, op.LoanID)
	if err != nil {
		return nil, err
	}
	if op.ExternalID == nil {
		m.submit(ctx, op, l.CompanyID)
		return op, nil
	}
	res, err := m.ports.Debits.DebitStatus(ctx, *op.ExternalID)
	switch {
	case err != nil:
		return nil, err
	case res == nil: // банк поручение забыл — передаём заново с тем же id
		m.submit(ctx, op, l.CompanyID)
		return op, nil
	case res.Status == shared.DebitFailed:
		return m.finish(ctx, op.ID, func(tx *sqlx.Tx, l *loanRow, op *opRow) error {
			op.Status, op.FailureReason = opFailed, &res.Reason
			return nil
		})
	case res.Status == shared.DebitSucceeded:
		return m.finish(ctx, op.ID, func(tx *sqlx.Tx, l *loanRow, op *opRow) error { return m.post(ctx, tx, l, op, res) })
	}
	return op, nil
}

// finish завершает операцию в транзакции: fn меняет кредит и операцию, версия кредита растёт.
func (m *Module) finish(ctx context.Context, opID string, fn func(tx *sqlx.Tx, l *loanRow, op *opRow) error) (*opRow, error) {
	var out *opRow
	err := db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		op, err := m.store.Operation(ctx, tx, opID)
		if err != nil {
			return err
		}
		if op == nil {
			return fmt.Errorf("operation %s not found", opID)
		}
		// Блокировки — в порядке «кредит, операция», как у команд: иначе возможна взаимоблокировка.
		l, err := m.store.LockLoan(ctx, tx, op.LoanID)
		if err != nil {
			return err
		}
		if op, err = m.store.LockOperation(ctx, tx, opID); err != nil {
			return err
		}
		out = op
		if op.Status == opSucceeded || op.Status == opFailed {
			return nil
		}
		if err := fn(tx, l, op); err != nil {
			return err
		}
		l.Version++
		if err := m.store.UpdateLoan(ctx, tx, l); err != nil {
			return err
		}
		op.LoanVersion, op.ScheduleVersion, op.LoanStatus, op.UpdatedAt = l.Version, l.ScheduleVersion, l.Status, m.now().UTC()
		return m.store.FinishOperation(ctx, tx, op)
	})
	return out, err
}

// post учитывает подтверждённые деньги: распределение по policy-1 или применение расчёта досрочного погашения,
// проводки, оплаченные суммы строк и закрытие полностью погашенного кредита.
func (m *Module) post(ctx context.Context, tx *sqlx.Tx, l *loanRow, op *opRow, res *shared.DebitResult) error {
	rows, posts, err := m.loanState(ctx, tx, l.ID)
	if err != nil {
		return err
	}
	var a allocation
	reason := closeScheduleCompleted
	if op.Type == opEarly {
		reason = closeEarlyRepayment
		if a, err = m.applyQuote(ctx, tx, l, rows, op, res); err != nil {
			return err
		}
	} else {
		a = allocate(res.PostedAmount, rows, debtOf(l, rows, posts, dateOf(res.PostedAt)), res.PostedAt)
		if err := m.store.UpdatePaid(ctx, tx, l.ID, a.apply(rows)); err != nil {
			return err
		}
	}
	if err := m.store.InsertPostings(ctx, tx, op.ID, l.ID, a.posts); err != nil {
		return err
	}
	l.PenaltiesPaid += a.penalties
	if err := m.closeIfRepaid(ctx, tx, l, res.PostedAt, reason); err != nil {
		return err
	}
	var fees int64
	op.Status, op.PostedAmount, op.PostedAt = opSucceeded, &res.PostedAmount, &res.PostedAt
	op.AllocPrincipal, op.AllocInterest, op.AllocPenalties, op.AllocFees, op.Unapplied = &a.principal, &a.interest, &a.penalties, &fees, &a.unapplied
	return nil
}

// closeIfRepaid закрывает невозобновляемый кредит, когда погашены все строки и пени. Овердрафт остаётся
// действующим: лимит закрывает клиент (POST /loans/{id}/close).
func (m *Module) closeIfRepaid(ctx context.Context, tx *sqlx.Tx, l *loanRow, at time.Time, reason string) error {
	if l.Revolving {
		return nil
	}
	rows, posts, err := m.loanState(ctx, tx, l.ID)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.remaining() > 0 {
			return nil
		}
	}
	if debtOf(l, rows, posts, dateOf(at)).penalties > 0 {
		return nil
	}
	at = at.UTC()
	l.Status, l.ClosedAt, l.CloseReason = loanClosed, &at, &reason
	return nil
}
