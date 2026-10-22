package servicing

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

// maxIdemKey — ключ идемпотентности длиннее не принимаем: он хранится в БД.
const maxIdemKey = 128

// actorOf — клиент из сессии: компания и ИНН человека для проверки полномочий.
func actorOf(c *gin.Context) (actor, error) {
	company, err := clientCompany(c)
	if err != nil {
		return actor{}, err
	}
	a := actor{companyID: company}
	if u := auth.UserFrom(c); u.PersonINN != nil {
		a.personINN = *u.PersonINN
	}
	return a, nil
}

// loanVersion — финансовая версия кредита из If-Match (допускаются кавычки).
func loanVersion(c *gin.Context) (int, error) {
	v, err := strconv.Atoi(strings.Trim(strings.TrimSpace(c.GetHeader("If-Match")), `"`))
	if err != nil || v < 0 {
		return 0, httpx.Validation("Нужен заголовок If-Match с версией кредита")
	}
	return v, nil
}

// idemKey — обязательный Idempotency-Key денежной команды.
func idemKey(c *gin.Context) (string, error) {
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if key == "" || len(key) > maxIdemKey {
		return "", httpx.Validation("Нужен заголовок Idempotency-Key — уникальный ключ команды")
	}
	return key, nil
}

// utc — момент в UTC для ответа: БД отдаёт время в часовом поясе соединения.
func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.UTC()
	return &v
}

// maskAccount — номер счёта для клиента: только последние четыре цифры.
func maskAccount(number string) string {
	const visible = 4
	if len(number) <= visible {
		return number
	}
	return "•••• " + number[len(number)-visible:]
}

type repaymentAccountDTO struct {
	ID               string `json:"id" example:"acc-7712345678-1"`
	MaskedNumber     string `json:"maskedNumber" example:"•••• 0001"`
	Currency         string `json:"currency" example:"RUB"`
	Available        bool   `json:"available" example:"true"`
	AvailableBalance *int64 `json:"availableBalance" example:"4800000"` // null — АБС остаток не отдаёт
}

// repaymentAccounts godoc
//
//	@Summary		Счета для оплаты кредита
//	@Description	Счета компании в АБС банка. available — активный рублёвый счёт, с которого можно списать.
//	@Tags			loans
//	@Produce		json
//	@Param			id	path		string	true	"ID кредита"
//	@Success		200	{array}		repaymentAccountDTO
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		502	{object}	httpx.ErrorBody
//	@Router			/loans/{id}/repayment-accounts [get]
func (m *Module) repaymentAccounts(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	if _, err := m.ownLoan(ctx, company, c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	inn, err := m.companyINN(ctx, company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	accs, err := m.ports.Debits.Accounts(ctx, inn)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]repaymentAccountDTO, len(accs))
	for i, a := range accs {
		out[i] = repaymentAccountDTO{
			ID: a.ID, MaskedNumber: maskAccount(a.Number), Currency: a.Currency, Available: a.Active && a.Currency == currencyRUB,
			AvailableBalance: a.Balance,
		}
	}
	httpx.JSON(c, http.StatusOK, out)
}

type accountRefDTO struct {
	ID           string `json:"id" example:"acc-7712345678-1"`
	MaskedNumber string `json:"maskedNumber" example:"•••• 0001"`
}

type allocationDTO struct {
	Principal int64 `json:"principal" example:"0"`
	Interest  int64 `json:"interest" example:"20000"`
	Penalties int64 `json:"penalties" example:"0"`
	Fees      int64 `json:"fees" example:"0"`
}

// operationDTO — денежная операция по кредиту. Суммы проводки и распределение — только у SUCCEEDED.
type operationDTO struct {
	ID                    string         `json:"id" example:"rp-501"`
	LoanID                string         `json:"loanId" example:"l-55"`
	Type                  string         `json:"type" example:"REGULAR"`   // REGULAR, EARLY
	Status                string         `json:"status" example:"PENDING"` // PENDING, PROCESSING, SUCCEEDED, FAILED
	Account               accountRefDTO  `json:"account"`
	RequestedAmount       int64          `json:"requestedAmount" example:"20000"`
	PostedAmount          *int64         `json:"postedAmount" example:"20000"`
	Allocation            *allocationDTO `json:"allocation"`
	UnappliedAmount       *int64         `json:"unappliedAmount" example:"0"`
	ExternalTransactionID *string        `json:"externalTransactionId" example:"dbt-rp-501"`
	FailureReason         *string        `json:"failureReason" example:"Недостаточно средств на счёте"`
	QuoteID               *string        `json:"quoteId" example:"eq-101"`
	RequestedAt           time.Time      `json:"requestedAt" example:"2026-10-21T10:00:00Z"`
	PostedAt              *time.Time     `json:"postedAt" example:"2026-10-21T10:00:05Z"`
	LoanVersion           int            `json:"loanVersion" example:"2"`
	ScheduleVersion       int            `json:"scheduleVersion" example:"1"`
	LoanStatus            string         `json:"loanStatus" example:"ACTIVE"`
}

func toOperationDTO(op *opRow) operationDTO {
	out := operationDTO{
		ID: op.ID, LoanID: op.LoanID, Type: op.Type, Status: op.Status,
		Account:         accountRefDTO{ID: op.AccountID, MaskedNumber: maskAccount(op.AccountNumber)},
		RequestedAmount: op.RequestedAmount, PostedAmount: op.PostedAmount, UnappliedAmount: op.Unapplied,
		ExternalTransactionID: op.ExternalID, FailureReason: op.FailureReason, QuoteID: op.QuoteID,
		RequestedAt: op.RequestedAt.UTC(), PostedAt: utc(op.PostedAt), LoanVersion: op.LoanVersion, ScheduleVersion: op.ScheduleVersion,
		LoanStatus: op.LoanStatus,
	}
	if op.Status == opSucceeded {
		out.Allocation = &allocationDTO{Principal: *op.AllocPrincipal, Interest: *op.AllocInterest, Penalties: *op.AllocPenalties, Fees: *op.AllocFees}
	}
	return out
}

type repaymentRequest struct {
	AccountID string `json:"accountId" binding:"required,max=64" example:"acc-7712345678-1"`
	Amount    int64  `json:"amount" binding:"required,gt=0" example:"20000"`
}

// createRepayment godoc
//
//	@Summary		Оплатить кредит
//	@Description	Запрос на списание со счёта компании. Долг уменьшится после подтверждения банком: статус — через GET операции.
//	@Description	Сумма — не больше просрочки, пеней и остатка ближайшего платежа.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string				true	"ID кредита"
//	@Param			If-Match		header		string				true	"Версия кредита"
//	@Param			Idempotency-Key	header		string				true	"Ключ команды"
//	@Param			body			body		repaymentRequest	true	"Счёт и сумма"
//	@Success		202				{object}	operationDTO
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		412				{object}	httpx.ErrorBody
//	@Failure		422				{object}	httpx.ErrorBody
//	@Failure		502				{object}	httpx.ErrorBody
//	@Router			/loans/{id}/repayments [post]
func (m *Module) createRepayment(c *gin.Context) {
	a, version, key, err := commandHeaders(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req repaymentRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	op, err := m.CreateRepayment(c.Request.Context(), a, c.Param("id"), version, key, repaymentCmd{AccountID: req.AccountID, Amount: req.Amount})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusAccepted, toOperationDTO(op))
}

// commandHeaders — клиент, If-Match и Idempotency-Key денежной команды.
func commandHeaders(c *gin.Context) (a actor, version int, key string, err error) {
	if a, err = actorOf(c); err != nil {
		return
	}
	if version, err = loanVersion(c); err != nil {
		return
	}
	key, err = idemKey(c)
	return
}

// operationsPage — схема ответа для swag, как loansPage.
type operationsPage struct {
	Items []operationDTO `json:"items"`
	Page  int            `json:"page" example:"1"`
	Size  int            `json:"size" example:"20"`
	Total int            `json:"total" example:"2"`
}

// repayments godoc
//
//	@Summary		История операций по кредиту
//	@Description	Обычные и досрочные операции, новые первыми.
//	@Tags			loans
//	@Produce		json
//	@Param			id		path		string	true	"ID кредита"
//	@Param			page	query		int		false	"Страница с 1"	default(1)
//	@Param			size	query		int		false	"Размер 1–100"	default(20)
//	@Success		200		{object}	operationsPage
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/loans/{id}/repayments [get]
func (m *Module) repayments(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	l, err := m.ownLoan(ctx, company, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	page, size := httpx.ParsePage(c)
	ops, total, err := m.store.LoanOperations(ctx, l.ID, size, (page-1)*size)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]operationDTO, len(ops))
	for i := range ops {
		out[i] = toOperationDTO(&ops[i])
	}
	httpx.JSON(c, http.StatusOK, httpx.NewPage(out, page, size, total))
}

// repayment godoc
//
//	@Summary		Операция по кредиту
//	@Description	Статус операции. Незавершённую сервер сначала сверяет с банком.
//	@Tags			loans
//	@Produce		json
//	@Param			id		path		string	true	"ID кредита"
//	@Param			opId	path		string	true	"ID операции"
//	@Success		200		{object}	operationDTO
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/loans/{id}/repayments/{opId} [get]
func (m *Module) repayment(c *gin.Context) { m.operation(c, "") }

// operation отдаёт операцию кредита типа typ (пусто — любого), предварительно сверив незавершённую с банком.
// Сбой сверки не ломает ответ: клиент получает сохранённое состояние, сверщик повторит.
func (m *Module) operation(c *gin.Context, typ string) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	l, err := m.ownLoan(ctx, company, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	op, err := m.store.Operation(ctx, m.deps.DB, c.Param("opId"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if op == nil || op.LoanID != l.ID || typ != "" && op.Type != typ {
		httpx.Fail(c, httpx.NotFound("Операция не найдена"))
		return
	}
	if op.Status == opPending || op.Status == opProcessing {
		switch fresh, err := m.Reconcile(ctx, op.ID); {
		case err != nil:
			slog.Warn("reconcile on read", "operationId", op.ID, "err", err)
		case fresh != nil:
			op = fresh
		}
	}
	httpx.JSON(c, http.StatusOK, toOperationDTO(op))
}
