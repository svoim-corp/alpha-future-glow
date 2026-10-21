package catalog

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

type moneyRangeDTO struct {
	Min int64 `json:"min" example:"300000"`
	Max int64 `json:"max" example:"10000000"`
}

type termRangeDTO struct {
	Min int `json:"min" example:"3"`
	Max int `json:"max" example:"36"`
}

type rateRangeDTO struct {
	Min float64 `json:"min" example:"20"`
	Max float64 `json:"max" example:"28"`
}

type productSummaryDTO struct {
	Code      string        `json:"code" example:"WORKING"`
	Name      string        `json:"name" example:"Оборотный кредит"`
	ShortText string        `json:"shortText" example:"Закупки, зарплата, налоги, аренда"`
	Goals     []string      `json:"goals" example:"WORKING_CAPITAL"`
	Amount    moneyRangeDTO `json:"amount"`
	Term      termRangeDTO  `json:"term"`
	Rate      rateRangeDTO  `json:"rate"`
}

func toSummary(p productRow) productSummaryDTO {
	return productSummaryDTO{
		Code: p.Code, Name: p.Name, ShortText: p.ShortText, Goals: p.Goals.V,
		Amount: moneyRangeDTO{Min: p.AmountMin, Max: p.AmountMax},
		Term:   termRangeDTO{Min: p.TermMin, Max: p.TermMax},
		Rate:   rateRangeDTO{Min: p.RateMin, Max: p.RateMax},
	}
}

// listProducts godoc
//
//	@Summary	Линейка продуктов для каталога
//	@Tags		catalog
//	@Produce	json
//	@Success	200	{array}	productSummaryDTO
//	@Router		/products [get]
func (m *Module) listProducts(c *gin.Context) {
	products, err := m.store.Products(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]productSummaryDTO, len(products))
	for i, p := range products {
		out[i] = toSummary(p)
	}
	httpx.JSON(c, http.StatusOK, out)
}
