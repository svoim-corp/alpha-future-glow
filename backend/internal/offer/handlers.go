package offer

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const dateLayout = "2006-01-02"

// clientCompany — компания клиента; у сотрудника банка её нет, и клиентские страницы ему недоступны.
func clientCompany(c *gin.Context) (string, error) {
	u := auth.UserFrom(c)
	if u.CompanyID == nil {
		return "", httpx.Forbidden("Раздел доступен клиенту банка")
	}
	return *u.CompanyID, nil
}

type requestedDTO struct {
	ProductCode string `json:"productCode" example:"WORKING"`
	Amount      int64  `json:"amount" example:"1500000"`
	Term        int    `json:"term" example:"24"`
}

type offeredDTO struct {
	ProductCode    string  `json:"productCode" example:"WORKING"`
	Amount         int64   `json:"amount" example:"1200000"`
	Term           int     `json:"term" example:"24"`
	Rate           float64 `json:"rate" example:"21.5"`
	MonthlyPayment int64   `json:"monthlyPayment" example:"61958"`
}

type alternativeDTO struct {
	ProductCode string `json:"productCode" example:"EXPRESS"`
	MaxAmount   int64  `json:"maxAmount" example:"500000"`
}

// offerResponse — одобрение и иные условия: requested, offered, explanation, validUntil; отказ: reason, alternatives.
type offerResponse struct {
	Decision     shared.Status    `json:"decision" example:"COUNTER_OFFER"`
	Requested    *requestedDTO    `json:"requested,omitempty"`
	Offered      *offeredDTO      `json:"offered,omitempty"`
	Explanation  *string          `json:"explanation,omitempty" example:"Сумма снижена по уровню долговой нагрузки"`
	ValidUntil   *string          `json:"validUntil,omitempty" example:"2026-10-27"`
	Reason       *string          `json:"reason,omitempty" example:"Высокая долговая нагрузка по данным БКИ"`
	Alternatives []alternativeDTO `json:"alternatives,omitempty"`
}

func toOfferResponse(o *offerRow) offerResponse {
	resp := offerResponse{Decision: shared.Status(o.Decision)}
	if resp.Decision == shared.StatusRejected {
		resp.Reason = o.Explanation
		resp.Alternatives = make([]alternativeDTO, len(o.Alternatives.V))
		for i, a := range o.Alternatives.V {
			resp.Alternatives[i] = alternativeDTO{ProductCode: a.ProductCode, MaxAmount: a.MaxAmount}
		}
		return resp
	}
	until := o.ValidUntil.Format(dateLayout)
	resp.Requested = &requestedDTO{ProductCode: o.RequestedProductCode, Amount: o.RequestedAmount, Term: o.RequestedTerm}
	resp.Offered = &offeredDTO{ProductCode: o.ProductCode, Amount: o.Amount, Term: o.Term, Rate: o.Rate, MonthlyPayment: o.MonthlyPayment}
	resp.Explanation, resp.ValidUntil = o.Explanation, &until
	return resp
}

// getOffer godoc
//
//	@Summary		Решение банка и предложение
//	@Description	Появляется после решения. Одобрение и иные условия — запрошенные и предложенные условия; отказ — причина и альтернативы.
//	@Tags			offer
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	offerResponse
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody	"Заявки нет или решения ещё нет"
//	@Router			/applications/{id}/offer [get]
func (m *Module) getOffer(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	_, o, err := m.offerOf(c.Request.Context(), c.Param("id"), company)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toOfferResponse(o))
}

// clientActor — клиент как автор действия в истории заявки.
func clientActor(c *gin.Context) shared.Actor {
	u := auth.UserFrom(c)
	return shared.Actor{Type: shared.ActorClient, UserID: u.ID, Name: u.FullName}
}

type statusResponse struct {
	Status   shared.Status `json:"status" example:"SIGNING"`
	Redirect string        `json:"redirect,omitempty" example:"/applications/a-1031/sign"`
}

// acceptOffer godoc
//
//	@Summary		Принять предложение банка
//	@Description	Одобрение или иные условия. Договор получает номер, заявка — статус SIGNING. Тела нет.
//	@Tags			offer
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	statusResponse
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		409	{object}	httpx.ErrorBody	"Отказ, уже принято или заявка не в статусе решения"
//	@Failure		422	{object}	httpx.ErrorBody	"Срок действия предложения истёк"
//	@Router			/applications/{id}/offer/accept [post]
func (m *Module) acceptOffer(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id := c.Param("id")
	if err := m.accept(c.Request.Context(), id, company, clientActor(c)); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, statusResponse{Status: shared.StatusSigning, Redirect: "/applications/" + id + "/sign"})
}
