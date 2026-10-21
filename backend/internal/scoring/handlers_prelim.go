package scoring

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// prelimResponse — лимит (outcome LIMIT) или стоп-отказ (outcome STOP); поля другого исхода опускаются.
type prelimResponse struct {
	Outcome         shared.PrelimOutcome `json:"outcome" example:"LIMIT"`
	RequestedAmount *int64               `json:"requestedAmount,omitempty" example:"1500000"`
	Limit           *int64               `json:"limit,omitempty" example:"3000000"`
	PrelimRate      *float64             `json:"prelimRate,omitempty" example:"22"`
	PrelimPayment   *int64               `json:"prelimPayment,omitempty" example:"77800"`
	AmountFits      *bool                `json:"amountFits,omitempty" example:"true"`
	Checks          []prelimCheck        `json:"checks,omitempty"`
	StopReason      *string              `json:"stopReason,omitempty" example:"BUSINESS_TOO_YOUNG"`
	StopText        string               `json:"stopText,omitempty" example:"Бизнес зарегистрирован меньше 6 месяцев назад"`
	WhatToDo        []string             `json:"whatToDo,omitempty"`
}

func toPrelimResponse(p *prelimRow, res prelimResult) prelimResponse {
	if p.StopReason != nil {
		text, todo := stopTexts(*p.StopReason, res.minAge)
		return prelimResponse{Outcome: shared.PrelimStop, StopReason: p.StopReason, StopText: text, WhatToDo: todo}
	}
	return prelimResponse{
		Outcome: shared.PrelimLimit, RequestedAmount: &p.Amount, Limit: &p.LimitAmount, PrelimRate: &p.PrelimRate,
		PrelimPayment: &p.PrelimPayment, AmountFits: &p.AmountFits, Checks: p.Checks.V,
	}
}

// companyOf — компания клиента; у сотрудника банка её нет, и чужие заявки он здесь не видит.
func companyOf(c *gin.Context) string {
	if u := auth.UserFrom(c); u.CompanyID != nil {
		return *u.CompanyID
	}
	return ""
}

// prelim godoc
//
//	@Summary		Предварительное решение (этап 1)
//	@Description	Проверки ЕГРЮЛ/ЕГРИП, БКИ, ФНС, суды, ГИР БО, обороты по счетам в банке. Лимит или стоп-отказ. Требует согласие на БКИ.
//	@Tags			applications
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	prelimResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		409	{object}	httpx.ErrorBody
//	@Failure		422	{object}	httpx.ErrorBody
//	@Failure		502	{object}	httpx.ErrorBody
//	@Router			/applications/{id}/prelim [post]
func (m *Module) prelim(c *gin.Context) {
	p, res, err := m.runPrelim(c.Request.Context(), c.Param("id"), companyOf(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toPrelimResponse(p, res))
}

type choiceRequest struct {
	Choice shared.PrelimChoice `json:"choice" binding:"required,oneof=REDUCE_AMOUNT KEEP_AMOUNT" example:"REDUCE_AMOUNT"`
	Amount int64               `json:"amount" binding:"required_if=Choice REDUCE_AMOUNT,gte=0" example:"3000000"`
}

type choiceParams struct {
	Amount int64 `json:"amount" example:"3000000"`
	Term   int   `json:"term" example:"24"`
}

type choiceResponse struct {
	Stage   int          `json:"stage" example:"2"`
	Params  choiceParams `json:"params"`
	Version int          `json:"version" example:"9"`
}

// stageFull — после выбора открывается этап 2.
const stageFull = 2

// prelimChoice godoc
//
//	@Summary		Выбор при сумме выше лимита
//	@Description	REDUCE_AMOUNT — уменьшить сумму до лимита, KEEP_AMOUNT — оставить запрошенную (решит сотрудник банка).
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"ID заявки"
//	@Param			body	body		choiceRequest	true	"Выбор клиента"
//	@Success		200		{object}	choiceResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Failure		409		{object}	httpx.ErrorBody
//	@Failure		422		{object}	httpx.ErrorBody
//	@Router			/applications/{id}/prelim/choice [post]
func (m *Module) prelimChoice(c *gin.Context) {
	var req choiceRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	u := auth.UserFrom(c)
	actor := shared.Actor{Type: shared.ActorClient, UserID: u.ID, Name: u.FullName}
	app, amount, version, err := m.choosePrelim(c.Request.Context(), c.Param("id"), companyOf(c), req.Choice, req.Amount, actor)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, choiceResponse{Stage: stageFull, Params: choiceParams{Amount: amount, Term: app.Term}, Version: version})
}
