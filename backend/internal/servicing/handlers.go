package servicing

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

const dateLayout = "2006-01-02"

// clientCompany — компания клиента; у сотрудника банка её нет, и кредиты клиентов ему здесь недоступны.
func clientCompany(c *gin.Context) (string, error) {
	u := auth.UserFrom(c)
	if u.CompanyID == nil {
		return "", httpx.Forbidden("Раздел доступен клиенту банка")
	}
	return *u.CompanyID, nil
}

// today — начало текущих суток в UTC: платёж сегодня ещё не просрочен.
func (m *Module) today() time.Time {
	y, mo, d := m.now().UTC().Date()
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}

type paymentRefDTO struct {
	Date   string `json:"date" example:"2026-11-20"`
	Amount int64  `json:"amount" example:"61958"`
}

type summaryResponse struct {
	ActiveLoans int            `json:"activeLoans" example:"1"`
	NextPayment *paymentRefDTO `json:"nextPayment"` // null — платежей впереди нет
}

// summary godoc
//
//	@Summary		Кредиты на главной
//	@Description	Сколько действующих кредитов и ближайший платёж: сумма всех платежей компании в ближайшую дату.
//	@Tags			loans
//	@Produce		json
//	@Success		200	{object}	summaryResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Router			/loans/summary [get]
func (m *Module) summary(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	active, err := m.store.ActiveLoans(ctx, company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	resp := summaryResponse{ActiveLoans: active}
	next, err := m.store.UpcomingPayments(ctx, company, m.today(), active)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	// Ближайших платежей — не больше одного на кредит; в одну дату они складываются.
	for _, p := range next {
		if resp.NextPayment == nil {
			resp.NextPayment = &paymentRefDTO{Date: p.DueDate.Format(dateLayout)}
		}
		if p.DueDate.Format(dateLayout) == resp.NextPayment.Date {
			resp.NextPayment.Amount += p.Total
		}
	}
	httpx.JSON(c, http.StatusOK, resp)
}

type productRefDTO struct {
	Code string `json:"code" example:"WORKING"`
	Name string `json:"name" example:"Оборотный кредит"`
}

type loanDTO struct {
	ID          string         `json:"id" example:"l-55"`
	Product     productRefDTO  `json:"product"`
	Amount      int64          `json:"amount" example:"1200000"`
	Balance     int64          `json:"balance" example:"1200000"`
	Rate        float64        `json:"rate" example:"21.5"`
	Term        int            `json:"term" example:"24"`
	IssuedAt    string         `json:"issuedAt" example:"2026-10-20"`
	NextPayment *paymentRefDTO `json:"nextPayment"` // null — платежей впереди нет
	Status      string         `json:"status" example:"ACTIVE"`
}

// byLoan — строки графиков по id кредита.
func byLoan(rows []paymentRow) map[string][]paymentRow {
	out := map[string][]paymentRow{}
	for _, p := range rows {
		out[p.LoanID] = append(out[p.LoanID], p)
	}
	return out
}

// list godoc
//
//	@Summary		Мои кредиты
//	@Description	Выданные кредиты компании, новые первыми: остаток долга и ближайший платёж.
//	@Tags			loans
//	@Produce		json
//	@Success		200	{array}		loanDTO
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Router			/loans [get]
func (m *Module) list(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	loans, err := m.store.CompanyLoans(ctx, company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	payments, err := m.store.CompanyPayments(ctx, company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	schedules, today := byLoan(payments), m.today()
	out := make([]loanDTO, len(loans))
	for i, l := range loans {
		st := stateOf(l.Amount, schedules[l.ID], today)
		out[i] = loanDTO{
			ID: l.ID, Product: productRefDTO{Code: l.ProductCode, Name: l.ProductName}, Amount: l.Amount, Balance: st.balance,
			Rate: l.Rate, Term: l.Term, IssuedAt: l.IssuedAt.UTC().Format(dateLayout), Status: l.Status,
		}
		if st.next != nil {
			out[i].NextPayment = &paymentRefDTO{Date: st.next.DueDate.Format(dateLayout), Amount: st.next.Total}
		}
	}
	httpx.JSON(c, http.StatusOK, out)
}

type loanTermsDTO struct {
	Amount         int64   `json:"amount" example:"1200000"`
	Term           int     `json:"term" example:"24"`
	Rate           float64 `json:"rate" example:"21.5"`
	GraceMonths    int     `json:"graceMonths" example:"0"`
	MonthlyPayment int64   `json:"monthlyPayment" example:"61958"`
}

type paymentDTO struct {
	Date      string `json:"date" example:"2026-11-20"`
	Principal int64  `json:"principal" example:"40458"`
	Interest  int64  `json:"interest" example:"21500"`
	Total     int64  `json:"total" example:"61958"`
	Status    string `json:"status" example:"PLANNED"` // PLANNED, PAID, OVERDUE
}

type loanDetailsDTO struct {
	ID             string        `json:"id" example:"l-55"`
	ApplicationID  string        `json:"applicationId" example:"a-1031"`
	ContractNumber string        `json:"contractNumber" example:"КД-2026-0055"`
	Product        productRefDTO `json:"product"`
	Balance        int64         `json:"balance" example:"1200000"`
	IssuedAt       string        `json:"issuedAt" example:"2026-10-20"`
	Status         string        `json:"status" example:"ACTIVE"`
	Terms          loanTermsDTO  `json:"terms"`
	Payments       []paymentDTO  `json:"payments"`
}

// get godoc
//
//	@Summary		Кредит: условия и график платежей
//	@Tags			loans
//	@Produce		json
//	@Param			id	path		string	true	"ID кредита"
//	@Success		200	{object}	loanDetailsDTO
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/loans/{id} [get]
func (m *Module) get(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	l, err := m.store.Loan(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	// Чужой кредит — 404: существование кредитов других компаний не раскрывается.
	if l == nil || l.CompanyID != company {
		httpx.Fail(c, httpx.NotFound("Кредит не найден"))
		return
	}
	schedule, err := m.store.LoanPayments(ctx, l.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	today := m.today()
	resp := loanDetailsDTO{
		ID: l.ID, ApplicationID: l.AppID, ContractNumber: l.ContractNumber,
		Product: productRefDTO{Code: l.ProductCode, Name: l.ProductName}, Balance: stateOf(l.Amount, schedule, today).balance,
		IssuedAt: l.IssuedAt.UTC().Format(dateLayout), Status: l.Status,
		Terms:    loanTermsDTO{Amount: l.Amount, Term: l.Term, Rate: l.Rate, GraceMonths: l.GraceMonths, MonthlyPayment: l.MonthlyPayment},
		Payments: make([]paymentDTO, len(schedule)),
	}
	for i, p := range schedule {
		resp.Payments[i] = paymentDTO{
			Date: p.DueDate.Format(dateLayout), Principal: p.Principal, Interest: p.Interest, Total: p.Total, Status: paymentStatus(p, today),
		}
	}
	httpx.JSON(c, http.StatusOK, resp)
}

// maxPeriodDays — календарь показывает месяц-два; год — с запасом, но без выгрузки всей истории.
const maxPeriodDays = 366

type scheduleItemDTO struct {
	Date        string `json:"date" example:"2026-11-20"`
	LoanID      string `json:"loanId" example:"l-55"`
	ProductName string `json:"productName" example:"Оборотный кредит"`
	Amount      int64  `json:"amount" example:"61958"`
	Status      string `json:"status" example:"PLANNED"` // PLANNED, PAID, OVERDUE
}

func toScheduleItems(rows []paymentRow, today time.Time) []scheduleItemDTO {
	out := make([]scheduleItemDTO, len(rows))
	for i, p := range rows {
		out[i] = scheduleItemDTO{
			Date: p.DueDate.Format(dateLayout), LoanID: p.LoanID, ProductName: p.ProductName, Amount: p.Total, Status: paymentStatus(p, today),
		}
	}
	return out
}

// period — ?from и ?to: даты YYYY-MM-DD, from не позже to, не больше года.
func period(c *gin.Context) (from, to time.Time, err error) {
	field := func(name, msg string) error {
		return httpx.Validation("", httpx.FieldError{Field: name, Code: "INVALID", Message: msg})
	}
	if from, err = time.Parse(dateLayout, c.Query("from")); err != nil {
		return from, to, field("from", "Укажите дату в формате ГГГГ-ММ-ДД")
	}
	if to, err = time.Parse(dateLayout, c.Query("to")); err != nil {
		return from, to, field("to", "Укажите дату в формате ГГГГ-ММ-ДД")
	}
	switch {
	case to.Before(from):
		return from, to, field("to", "Дата окончания раньше даты начала")
	case to.Sub(from) > maxPeriodDays*24*time.Hour:
		return from, to, field("to", "Период — не больше года")
	}
	return from, to, nil
}

// schedule godoc
//
//	@Summary		Платёжный календарь
//	@Description	Платежи всех кредитов компании за период, по датам. Период — не больше года.
//	@Tags			loans
//	@Produce		json
//	@Param			from	query		string	true	"С даты"	example(2026-11-01)
//	@Param			to		query		string	true	"По дату включительно"	example(2026-12-31)
//	@Success		200		{array}		scheduleItemDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/loans/schedule [get]
func (m *Module) schedule(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	from, to, err := period(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := m.store.PeriodPayments(c.Request.Context(), company, from, to)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toScheduleItems(rows, m.today()))
}

// Сколько ближайших платежей показывать: по умолчанию три, как в блоке на главной, и не больше года по одному кредиту.
const (
	defaultUpcoming = 3
	maxUpcoming     = 12
)

// upcoming godoc
//
//	@Summary		Ближайшие платежи
//	@Description	Неоплаченные платежи всех кредитов компании сегодня и позже, ближайшие первыми.
//	@Tags			loans
//	@Produce		json
//	@Param			limit	query		int	false	"Сколько платежей (1–12)"	default(3)
//	@Success		200		{array}		scheduleItemDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/loans/schedule/upcoming [get]
func (m *Module) upcoming(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	limit := defaultUpcoming
	if q := c.Query("limit"); q != "" {
		if limit, err = strconv.Atoi(q); err != nil || limit < 1 || limit > maxUpcoming {
			httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "limit", Code: "INVALID", Message: "Значение вне допустимого диапазона"}))
			return
		}
	}
	today := m.today()
	rows, err := m.store.UpcomingPayments(c.Request.Context(), company, today, limit)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toScheduleItems(rows, today))
}
