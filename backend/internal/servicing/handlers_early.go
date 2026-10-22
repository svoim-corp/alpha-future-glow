package servicing

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

type quoteRequest struct {
	Type          string  `json:"type" binding:"required,oneof=PARTIAL FULL" example:"PARTIAL"`
	Amount        *int64  `json:"amount" binding:"required_if=Type PARTIAL,excluded_if=Type FULL,omitempty,gt=0" example:"200000"`
	Strategy      *string `json:"strategy" binding:"required_if=Type PARTIAL,excluded_if=Type FULL,omitempty,oneof=REDUCE_TERM REDUCE_PAYMENT" example:"REDUCE_TERM"`
	EffectiveDate *string `json:"effectiveDate" binding:"omitempty,datetime=2006-01-02" example:"2026-10-21"`
}

type previewDTO struct {
	Date      string `json:"date" example:"2026-11-20"`
	Principal int64  `json:"principal" example:"44045"`
	Interest  int64  `json:"interest" example:"17913"`
	Total     int64  `json:"total" example:"61958"`
}

type quoteDTO struct {
	QuoteID                  string        `json:"quoteId" example:"eq-101"`
	LoanID                   string        `json:"loanId" example:"l-55"`
	Type                     string        `json:"type" example:"PARTIAL"`
	Strategy                 *string       `json:"strategy" example:"REDUCE_TERM"`
	LoanVersion              int           `json:"loanVersion" example:"2"`
	ScheduleVersion          int           `json:"scheduleVersion" example:"1"`
	CalculatedAt             time.Time     `json:"calculatedAt" example:"2026-10-21T10:00:00Z"`
	EffectiveDate            string        `json:"effectiveDate" example:"2026-10-21"`
	ValidUntil               time.Time     `json:"validUntil" example:"2026-10-21T10:15:00Z"`
	AmountToDebit            int64         `json:"amountToDebit" example:"200000"`
	Allocation               allocationDTO `json:"allocation"`
	PrepaidInterestCredit    int64         `json:"prepaidInterestCredit" example:"0"` // переплаченные проценты, засчитанные в основной долг
	PrincipalAfter           int64         `json:"principalAfter" example:"1000707"`
	MonthlyPaymentAfter      int64         `json:"monthlyPaymentAfter" example:"61958"`
	RemainingPaymentsAfter   int           `json:"remainingPaymentsAfter" example:"20"`
	ExpectedCloseDate        string        `json:"expectedCloseDate" example:"2028-06-20"`
	SchedulePreview          []previewDTO  `json:"schedulePreview"`
	Warnings                 []string      `json:"warnings"`
	CalculationPolicyVersion string        `json:"calculationPolicyVersion" example:"policy-1"`
}

const warnLimitOpen = "Лимит овердрафта останется открытым — после погашения его можно закрыть в кабинете"

func toQuoteDTO(q *quoteRow, revolving bool) quoteDTO {
	out := quoteDTO{
		QuoteID: q.ID, LoanID: q.LoanID, Type: q.Type, Strategy: q.Strategy, LoanVersion: q.LoanVersion, ScheduleVersion: q.ScheduleVersion,
		CalculatedAt: q.CalculatedAt.UTC(), EffectiveDate: q.EffectiveDate.Format(dateLayout), ValidUntil: q.ValidUntil.UTC(),
		AmountToDebit: q.AmountToDebit, Allocation: allocationDTO{Principal: q.AllocPrincipal, Interest: q.AllocInterest},
		PrepaidInterestCredit: q.InterestCredit,
		ExpectedCloseDate:     q.EffectiveDate.Format(dateLayout), SchedulePreview: make([]previewDTO, len(q.Preview.V)),
		Warnings: []string{}, CalculationPolicyVersion: policyVersion,
	}
	for i, p := range q.Preview.V {
		out.SchedulePreview[i] = previewDTO{Date: p.Date, Principal: p.Principal, Interest: p.Interest, Total: p.Principal + p.Interest}
		out.PrincipalAfter += p.Principal
		out.ExpectedCloseDate = p.Date
	}
	if q.Type == earlyPartial {
		out.MonthlyPaymentAfter, out.RemainingPaymentsAfter = q.MonthlyAfter, len(q.Preview.V)
	} else if revolving {
		out.Warnings = append(out.Warnings, warnLimitOpen)
	}
	return out
}

// createQuote godoc
//
//	@Summary		Рассчитать досрочное погашение
//	@Description	Расчёт на сегодня по policy-1: сумма списания, распределение и новый график. Денег не списывает.
//	@Description	Действует 15 минут и до изменения кредита.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"ID кредита"
//	@Param			Idempotency-Key	header		string			true	"Ключ расчёта"
//	@Param			body			body		quoteRequest	true	"Тип, сумма и стратегия"
//	@Success		201				{object}	quoteDTO
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		422				{object}	httpx.ErrorBody
//	@Router			/loans/{id}/early-repayment-quotes [post]
func (m *Module) createQuote(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	key, err := idemKey(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req quoteRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	cmd := quoteCmd{Type: req.Type, Amount: req.Amount, Strategy: req.Strategy, EffectiveDate: req.EffectiveDate}
	q, err := m.Quote(ctx, company, c.Param("id"), key, cmd)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	l, err := m.ownLoan(ctx, company, q.LoanID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, toQuoteDTO(q, l.Revolving))
}

type earlyRequest struct {
	QuoteID   string `json:"quoteId" binding:"required,max=64" example:"eq-101"`
	AccountID string `json:"accountId" binding:"required,max=64" example:"acc-7712345678-1"`
}

// createEarly godoc
//
//	@Summary		Подтвердить досрочное погашение
//	@Description	Подтверждение расчёта и запрос списания. Новый график применяется после подтверждения банком.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"ID кредита"
//	@Param			If-Match		header		string			true	"Версия кредита"
//	@Param			Idempotency-Key	header		string			true	"Ключ команды"
//	@Param			body			body		earlyRequest	true	"Расчёт и счёт"
//	@Success		202				{object}	operationDTO
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		412				{object}	httpx.ErrorBody
//	@Failure		422				{object}	httpx.ErrorBody
//	@Failure		502				{object}	httpx.ErrorBody
//	@Router			/loans/{id}/early-repayments [post]
func (m *Module) createEarly(c *gin.Context) {
	a, version, key, err := commandHeaders(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req earlyRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	op, err := m.ConfirmEarly(c.Request.Context(), a, c.Param("id"), version, key, earlyCmd{QuoteID: req.QuoteID, AccountID: req.AccountID})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusAccepted, toOperationDTO(op))
}

// earlyRepayment godoc
//
//	@Summary		Досрочное погашение: статус
//	@Description	Статус досрочного погашения и итоговые версии кредита и графика.
//	@Tags			loans
//	@Produce		json
//	@Param			id		path		string	true	"ID кредита"
//	@Param			opId	path		string	true	"ID операции"
//	@Success		200		{object}	operationDTO
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/loans/{id}/early-repayments/{opId} [get]
func (m *Module) earlyRepayment(c *gin.Context) { m.operation(c, opEarly) }

// closeLoan godoc
//
//	@Summary		Закрыть овердрафт
//	@Description	Закрытие возобновляемого лимита с нулевой задолженностью. Обычный кредит закрывается сам после погашения.
//	@Tags			loans
//	@Produce		json
//	@Param			id				path		string	true	"ID кредита"
//	@Param			If-Match		header		string	true	"Версия кредита"
//	@Param			Idempotency-Key	header		string	true	"Ключ команды"
//	@Success		200				{object}	loanDetailsDTO
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		412				{object}	httpx.ErrorBody
//	@Failure		422				{object}	httpx.ErrorBody
//	@Router			/loans/{id}/close [post]
func (m *Module) closeLoan(c *gin.Context) {
	a, version, _, err := commandHeaders(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	if err := m.CloseLimit(ctx, a, c.Param("id"), version); err != nil {
		httpx.Fail(c, err)
		return
	}
	l, err := m.ownLoan(ctx, a.companyID, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	schedule, err := m.store.LoanPayments(ctx, m.deps.DB, l.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toLoanDetails(l, schedule, m.today()))
}

type overdueDTO struct {
	Since     *string `json:"since" example:"2026-11-21"` // null — просрочки нет
	Days      int     `json:"days" example:"35"`
	Principal int64   `json:"principal" example:"40458"`
	Interest  int64   `json:"interest" example:"21500"`
	Penalties int64   `json:"penalties" example:"310"`
	Fees      int64   `json:"fees" example:"0"`
	Total     int64   `json:"total" example:"62268"`
}

type debtDTO struct {
	LoanID                   string     `json:"loanId" example:"l-55"`
	LoanVersion              int        `json:"loanVersion" example:"4"`
	ScheduleVersion          int        `json:"scheduleVersion" example:"1"`
	AsOf                     time.Time  `json:"asOf" example:"2026-12-25T10:00:00Z"`
	BusinessDate             string     `json:"businessDate" example:"2026-12-25"`
	TimeZone                 string     `json:"timeZone" example:"UTC"`
	PrincipalOutstanding     int64      `json:"principalOutstanding" example:"1200000"`
	InterestOutstanding      int64      `json:"interestOutstanding" example:"24004"`
	PenaltiesOutstanding     int64      `json:"penaltiesOutstanding" example:"310"`
	FeesOutstanding          int64      `json:"feesOutstanding" example:"0"`
	TotalOutstanding         int64      `json:"totalOutstanding" example:"1224314"`
	PrepaidInterest          int64      `json:"prepaidInterest" example:"0"` // оплачено заранее сверх начисленного; учитывается при досрочном погашении
	Overdue                  overdueDTO `json:"overdue"`
	UnappliedFunds           int64      `json:"unappliedFunds" example:"0"`
	CalculationPolicyVersion string     `json:"calculationPolicyVersion" example:"policy-1"`
}

// debt godoc
//
//	@Summary		Полная задолженность
//	@Description	Основной долг, проценты, пени, комиссии и просрочка на момент расчёта по policy-1. Просрочка — часть итога.
//	@Tags			loans
//	@Produce		json
//	@Param			id	path		string	true	"ID кредита"
//	@Success		200	{object}	debtDTO
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/loans/{id}/debt [get]
func (m *Module) debt(c *gin.Context) {
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
	s, err := m.snapshot(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	today := m.today()
	d := debtOf(s.loan, plain(s.rows), s.posts, today)
	resp := debtDTO{
		LoanID: s.loan.ID, LoanVersion: s.loan.Version, ScheduleVersion: s.loan.ScheduleVersion, AsOf: s.at, BusinessDate: today.Format(dateLayout),
		TimeZone: "UTC", PrincipalOutstanding: d.principal, InterestOutstanding: d.interest, PenaltiesOutstanding: d.penalties,
		TotalOutstanding: d.total(), PrepaidInterest: d.prepaid, UnappliedFunds: s.unapplied, CalculationPolicyVersion: policyVersion,
		Overdue: overdueDTO{Principal: d.overduePrincipal, Interest: d.overdueInterest, Penalties: d.penalties, Total: d.overdueTotal()},
	}
	if d.overdueSince != nil {
		since := d.overdueSince.Format(dateLayout)
		resp.Overdue.Since, resp.Overdue.Days = &since, daysBetween(*d.overdueSince, today)+1
	}
	httpx.JSON(c, http.StatusOK, resp)
}
