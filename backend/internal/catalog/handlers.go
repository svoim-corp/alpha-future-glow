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

type requirementsDTO struct {
	MinBusinessAgeMonths *int `json:"minBusinessAgeMonths" example:"12"` // null — требования к возрасту бизнеса нет
	GuarantorForLLC      bool `json:"guarantorForLlc" example:"true"`
}

type productDetailsDTO struct {
	productSummaryDTO
	Description    string          `json:"description" example:"Оборудование, помещение, расширение бизнеса"`
	GraceMaxMonths *int            `json:"graceMaxMonths" example:"6"` // null — отсрочки нет
	Requirements   requirementsDTO `json:"requirements"`
	LimitRule      string          `json:"limitRule" example:"До 15% годового оборота по счетам в банке"`
	WhatYouNeed    []string        `json:"whatYouNeed" example:"Согласие на запрос в БКИ"`
}

// positive — 0 в рамках продукта значит «нет», в API это null.
func positive(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

// productDetails godoc
//
//	@Summary		Условия продукта
//	@Description	Все условия для панели «Условия». whatYouNeed — только действия клиента: данные компании банк берёт сам.
//	@Tags			catalog
//	@Produce		json
//	@Param			code	path		string	true	"Код продукта"	example(INVEST)
//	@Success		200		{object}	productDetailsDTO
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/products/{code} [get]
func (m *Module) productDetails(c *gin.Context) {
	p, err := m.store.Product(c.Request.Context(), c.Param("code"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if p == nil {
		httpx.Fail(c, httpx.NotFound("Продукт не найден"))
		return
	}
	httpx.JSON(c, http.StatusOK, productDetailsDTO{
		productSummaryDTO: toSummary(*p),
		Description:       p.Description,
		GraceMaxMonths:    positive(p.GraceMaxMonths),
		Requirements:      requirementsDTO{MinBusinessAgeMonths: positive(p.MinBusinessAgeMonths), GuarantorForLLC: p.GuarantorForLLC},
		LimitRule:         p.LimitRule,
		WhatYouNeed:       p.WhatYouNeed.V,
	})
}

type goalDTO struct {
	Code string `json:"code" example:"WORKING_CAPITAL"`
	Name string `json:"name" example:"Пополнить оборотные средства"`
}

// listGoals godoc
//
//	@Summary	Цели кредита для подбора по задаче
//	@Tags		catalog
//	@Produce	json
//	@Success	200	{array}	goalDTO
//	@Router		/products/goals [get]
func (m *Module) listGoals(c *gin.Context) {
	goals, err := m.store.Goals(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]goalDTO, len(goals))
	for i, g := range goals {
		out[i] = goalDTO{Code: g.Code, Name: g.Name}
	}
	httpx.JSON(c, http.StatusOK, out)
}
