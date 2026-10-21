package catalog

import (
	"net/http"
	"slices"

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

type matchRequest struct {
	Goal   string `json:"goal" binding:"required" example:"WORKING_CAPITAL"`
	Amount int64  `json:"amount" binding:"required,gt=0" example:"1500000"`
	Term   int    `json:"term" binding:"gte=0" example:"24"` // 0 или нет поля — срок не проверяется
}

type matchDTO struct {
	Code   string `json:"code" example:"WORKING"`
	Fits   bool   `json:"fits" example:"true"`
	Reason string `json:"reason" example:"Подходит по цели, сумме и сроку"`
}

// matchProducts godoc
//
//	@Summary		Подбор продукта по задаче
//	@Description	Для каждого продукта — подходит ли он под цель, сумму и срок и почему. Подходящие — первыми.
//	@Tags			catalog
//	@Accept			json
//	@Produce		json
//	@Param			body	body		matchRequest	true	"Задача клиента"
//	@Success		200		{array}		matchDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Router			/products/match [post]
func (m *Module) matchProducts(c *gin.Context) {
	var req matchRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	goals, err := m.store.Goals(ctx)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	i := slices.IndexFunc(goals, func(g goalRow) bool { return g.Code == req.Goal })
	if i < 0 {
		httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "goal", Code: "INVALID", Message: "Такой цели нет в каталоге"}))
		return
	}
	products, err := m.store.Products(ctx)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	results := match(products, req.Goal, goals[i].Name, req.Amount, req.Term)
	out := make([]matchDTO, len(results))
	for i, r := range results {
		out[i] = matchDTO{Code: r.code, Fits: r.fits, Reason: r.reason}
	}
	httpx.JSON(c, http.StatusOK, out)
}

type calculatorRequest struct {
	ProductCode string `json:"productCode" binding:"required" example:"INVEST"`
	Amount      int64  `json:"amount" binding:"required,gt=0" example:"6000000"`
	Term        int    `json:"term" binding:"required,gt=0" example:"60"`
	GraceMonths int    `json:"graceMonths" binding:"gte=0" example:"6"` // 0 или null — без отсрочки
}

type gracePaymentDTO struct {
	Min    int64 `json:"min" example:"90000"`
	Max    int64 `json:"max" example:"130000"`
	Months int   `json:"months" example:"6"`
}

type calculatorResponse struct {
	Rate                rateRangeDTO     `json:"rate"`
	Payment             moneyRangeDTO    `json:"payment"`
	GracePayment        *gracePaymentDTO `json:"gracePayment"` // null — без отсрочки
	WithinProductLimits bool             `json:"withinProductLimits" example:"true"`
}

// calculatePayment godoc
//
//	@Summary		Калькулятор платежа
//	@Description	Ставку до проверки банк не знает, поэтому платёж — диапазон по ставкам продукта. Округление — до 100 ₽.
//	@Tags			catalog
//	@Accept			json
//	@Produce		json
//	@Param			body	body		calculatorRequest	true	"Параметры кредита"
//	@Success		200		{object}	calculatorResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/calculator/payment [post]
func (m *Module) calculatePayment(c *gin.Context) {
	var req calculatorRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	p, err := m.store.Product(c.Request.Context(), req.ProductCode)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if p == nil {
		httpx.Fail(c, httpx.NotFound("Продукт не найден"))
		return
	}
	calc := calculate(*p, req.Amount, req.Term, req.GraceMonths)
	resp := calculatorResponse{
		Rate:                rateRangeDTO{Min: p.RateMin, Max: p.RateMax},
		Payment:             moneyRangeDTO{Min: calc.paymentMin, Max: calc.paymentMax},
		WithinProductLimits: calc.withinLimits,
	}
	if req.GraceMonths > 0 {
		resp.GracePayment = &gracePaymentDTO{Min: calc.graceMin, Max: calc.graceMax, Months: req.GraceMonths}
	}
	httpx.JSON(c, http.StatusOK, resp)
}
