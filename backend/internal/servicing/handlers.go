package servicing

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

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

type overdueSummaryDTO struct {
	Loans    int   `json:"loans" example:"0"`
	Payments int   `json:"payments" example:"0"`
	Amount   int64 `json:"amount" example:"0"` // просроченная задолженность с пенями, как overdue.total в /debt
}

type summaryResponse struct {
	ActiveLoans           int               `json:"activeLoans" example:"1"`
	ClosedLoans           int               `json:"closedLoans" example:"0"`
	TotalLoans            int               `json:"totalLoans" example:"1"`
	TotalPrincipalBalance int64             `json:"totalPrincipalBalance" example:"1200000"`
	NextPayment           *paymentRefDTO    `json:"nextPayment"` // null — платежей впереди нет
	Overdue               overdueSummaryDTO `json:"overdue"`
}

// summary godoc
//
//	@Summary		Кредиты на главной
//	@Description	Сводка по всей компании: число кредитов, остаток долга, ближайший платёж (сумма остатков всех кредитов
//	@Description	в ближайшую дату) и просрочка с пенями.
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
	var (
		loans    []loanRow
		payments []paymentRow
		posts    []companyPosting
	)
	err = m.store.Snapshot(ctx, func(q sqlx.ExtContext) error {
		if loans, err = m.store.AllCompanyLoans(ctx, q, company); err != nil {
			return err
		}
		if payments, err = m.store.CompanyPayments(ctx, q, company); err != nil {
			return err
		}
		posts, err = m.store.CompanyPostings(ctx, q, company)
		return err
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, summarize(loans, byLoan(payments), postingsByLoan(posts), m.today()))
}

func postingsByLoan(rows []companyPosting) map[string][]posting {
	out := map[string][]posting{}
	for _, p := range rows {
		out[p.LoanID] = append(out[p.LoanID], p.posting)
	}
	return out
}

// summarize считает сводку по всем кредитам компании. Ближайших платежей — не больше одного на кредит; в одну дату
// их остатки складываются. Просрочка в nextPayment не входит.
func summarize(loans []loanRow, schedules map[string][]paymentRow, posts map[string][]posting, today time.Time) summaryResponse {
	resp := summaryResponse{TotalLoans: len(loans)}
	var next *time.Time
	for i := range loans {
		l := &loans[i]
		if l.Status != loanActive {
			resp.ClosedLoans++
			continue
		}
		resp.ActiveLoans++
		rows := schedules[l.ID]
		st := stateOf(l.Amount, rows, today)
		resp.TotalPrincipalBalance += st.balance
		if d := debtOf(l, plain(rows), posts[l.ID], today); d.overdueTotal() > 0 {
			resp.Overdue.Loans++
			resp.Overdue.Payments += d.overduePayments
			resp.Overdue.Amount += d.overdueTotal()
		}
		switch {
		case st.next == nil:
		case next == nil || st.next.DueDate.Before(*next):
			next = &st.next.DueDate
			resp.NextPayment = &paymentRefDTO{Date: next.Format(dateLayout), Amount: st.next.remaining()}
		case st.next.DueDate.Equal(*next):
			resp.NextPayment.Amount += st.next.remaining()
		}
	}
	return resp
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
	payments, err := m.store.CompanyPayments(ctx, m.deps.DB, company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toLoanDTOs(loans, byLoan(payments), m.today()))
}

// toLoanDTOs — элементы списка кредитов; nextPayment.amount — остаток к оплате ближайшей строки.
func toLoanDTOs(loans []loanRow, schedules map[string][]paymentRow, today time.Time) []loanDTO {
	out := make([]loanDTO, len(loans))
	for i, l := range loans {
		st := stateOf(l.Amount, schedules[l.ID], today)
		out[i] = loanDTO{
			ID: l.ID, Product: productRefDTO{Code: l.ProductCode, Name: l.ProductName}, Amount: l.Amount, Balance: st.balance,
			Rate: l.Rate, Term: l.Term, IssuedAt: l.IssuedAt.UTC().Format(dateLayout), Status: l.Status,
		}
		if st.next != nil {
			out[i].NextPayment = &paymentRefDTO{Date: st.next.DueDate.Format(dateLayout), Amount: st.next.remaining()}
		}
	}
	return out
}

// loansPage — схема ответа для swag: обобщённый httpx.Page он без --parseDependency не разбирает.
// Отдаётся по-прежнему httpx.NewPage, JSON тот же.
type loansPage struct {
	Items []loanDTO `json:"items"`
	Page  int       `json:"page" example:"1"`
	Size  int       `json:"size" example:"20"`
	Total int       `json:"total" example:"135"`
}

// Фильтр страницы кредитов по статусу; пусто — все.
var loanStatuses = map[string]bool{"": true, loanActive: true, loanClosed: true}

// page godoc
//
//	@Summary		Мои кредиты — страницами
//	@Description	Выданные кредиты компании страницами, новые первыми. total — с учётом фильтра.
//	@Tags			loans
//	@Produce		json
//	@Param			page	query		int		false	"Страница с 1"	default(1)
//	@Param			size	query		int		false	"Размер 1–100"	default(20)
//	@Param			status	query		string	false	"ACTIVE или CLOSED; без фильтра — все"
//	@Success		200		{object}	loansPage
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/loans/page [get]
func (m *Module) page(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	status := c.Query("status")
	if !loanStatuses[status] {
		httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "status", Code: "INVALID", Message: "Недопустимое значение"}))
		return
	}
	page, size := httpx.ParsePage(c)
	ctx := c.Request.Context()
	var (
		loans    []loanRow
		total    int
		payments []paymentRow
	)
	err = m.store.Snapshot(ctx, func(q sqlx.ExtContext) error {
		if loans, total, err = m.store.CompanyLoansPage(ctx, q, company, status, size, (page-1)*size); err != nil {
			return err
		}
		payments, err = m.store.CompanyPayments(ctx, q, company)
		return err
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, httpx.NewPage(toLoanDTOs(loans, byLoan(payments), m.today()), page, size, total))
}

type loanTermsDTO struct {
	Amount         int64   `json:"amount" example:"1200000"`
	Term           int     `json:"term" example:"24"`
	Rate           float64 `json:"rate" example:"21.5"`
	GraceMonths    int     `json:"graceMonths" example:"0"`
	MonthlyPayment int64   `json:"monthlyPayment" example:"61958"`
}

// paymentDTO — строка графика: плановые суммы, оплаченное и остаток.
type paymentDTO struct {
	ID                 string `json:"id" example:"sp-55-1"`
	Date               string `json:"date" example:"2026-11-20"`
	Principal          int64  `json:"principal" example:"40458"`
	Interest           int64  `json:"interest" example:"21500"`
	Total              int64  `json:"total" example:"61958"`
	PaidPrincipal      int64  `json:"paidPrincipal" example:"0"`
	PaidInterest       int64  `json:"paidInterest" example:"20000"`
	RemainingPrincipal int64  `json:"remainingPrincipal" example:"40458"`
	RemainingInterest  int64  `json:"remainingInterest" example:"1500"`
	RemainingTotal     int64  `json:"remainingTotal" example:"41958"`
	Early              bool   `json:"early" example:"false"`    // строка досрочного погашения
	Status             string `json:"status" example:"PLANNED"` // PLANNED, PAID, OVERDUE
}

// paymentID — устойчивый id строки: номер кредита и номер строки (l-55, 1 → sp-55-1).
func paymentID(loanID string, n int) string {
	return "sp-" + strings.TrimPrefix(loanID, "l-") + "-" + strconv.Itoa(n)
}

func toPaymentDTO(p paymentRow, today time.Time) paymentDTO {
	return paymentDTO{
		ID: paymentID(p.LoanID, p.N), Date: p.DueDate.Format(dateLayout), Principal: p.Principal, Interest: p.Interest, Total: p.Total,
		PaidPrincipal: p.PaidPrincipal, PaidInterest: p.PaidInterest, RemainingPrincipal: p.remainingPrincipal(),
		RemainingInterest: p.remainingInterest(), RemainingTotal: p.remaining(), Early: p.Early, Status: paymentStatus(p, today),
	}
}

type loanDetailsDTO struct {
	ID              string        `json:"id" example:"l-55"`
	ApplicationID   string        `json:"applicationId" example:"a-1031"`
	ContractNumber  string        `json:"contractNumber" example:"КД-2026-0055"`
	Product         productRefDTO `json:"product"`
	Balance         int64         `json:"balance" example:"1200000"`
	IssuedAt        string        `json:"issuedAt" example:"2026-10-20"`
	Status          string        `json:"status" example:"ACTIVE"`
	Revolving       bool          `json:"revolving" example:"false"`
	ClosedAt        *time.Time    `json:"closedAt"`
	CloseReason     *string       `json:"closeReason" example:"EARLY_REPAYMENT"` // SCHEDULE_COMPLETED, EARLY_REPAYMENT, LIMIT_CLOSED
	Version         int           `json:"version" example:"1"`
	ScheduleVersion int           `json:"scheduleVersion" example:"1"`
	Terms           loanTermsDTO  `json:"terms"`
	Payments        []paymentDTO  `json:"payments"`
}

func toLoanDetails(l *loanRow, schedule []paymentRow, today time.Time) loanDetailsDTO {
	resp := loanDetailsDTO{
		ID: l.ID, ApplicationID: l.AppID, ContractNumber: l.ContractNumber,
		Product: productRefDTO{Code: l.ProductCode, Name: l.ProductName}, Balance: stateOf(l.Amount, schedule, today).balance,
		IssuedAt: l.IssuedAt.UTC().Format(dateLayout), Status: l.Status, Revolving: l.Revolving, ClosedAt: utc(l.ClosedAt),
		CloseReason: l.CloseReason, Version: l.Version, ScheduleVersion: l.ScheduleVersion,
		Terms:    loanTermsDTO{Amount: l.Amount, Term: l.Term, Rate: l.Rate, GraceMonths: l.GraceMonths, MonthlyPayment: l.MonthlyPayment},
		Payments: make([]paymentDTO, len(schedule)),
	}
	for i, p := range schedule {
		resp.Payments[i] = toPaymentDTO(p, today)
	}
	return resp
}

// get godoc
//
//	@Summary		Кредит: условия и график платежей
//	@Description	Условия договора, версии и действующий график с оплаченными и оставшимися суммами строк.
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
	schedule, err := m.store.LoanPayments(ctx, m.deps.DB, l.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toLoanDetails(l, schedule, m.today()))
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

// toScheduleItems — платежи календаря: amount — остаток к оплате, у оплаченной строки — её сумма.
func toScheduleItems(rows []paymentRow, today time.Time) []scheduleItemDTO {
	out := make([]scheduleItemDTO, len(rows))
	for i, p := range rows {
		amount := p.remaining()
		if p.Status == paymentPaid {
			amount = p.Total
		}
		out[i] = scheduleItemDTO{
			Date: p.DueDate.Format(dateLayout), LoanID: p.LoanID, ProductName: p.ProductName, Amount: amount, Status: paymentStatus(p, today),
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
