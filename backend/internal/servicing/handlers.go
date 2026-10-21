package servicing

import (
	"net/http"
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
