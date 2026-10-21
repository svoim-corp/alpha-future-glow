package scoring

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type factorDTO struct {
	Code   string `json:"code" example:"DEBT_LOAD"`
	Label  string `json:"label" example:"Долговая нагрузка 47% от оборота"`
	Impact int    `json:"impact" example:"-70"`
}

type suggestedDTO struct {
	Decision shared.Status `json:"decision" example:"COUNTER_OFFER"`
	Amount   int64         `json:"amount,omitempty" example:"1200000"`
	Rate     float64       `json:"rate,omitempty" example:"21.5"`
}

type scoreResponse struct {
	Score                int             `json:"score" example:"640"`
	RiskBand             shared.RiskBand `json:"riskBand" example:"MEDIUM"`
	AutoDecisionEligible bool            `json:"autoDecisionEligible" example:"false"`
	AutoDecisionReason   string          `json:"autoDecisionReason" example:"Пограничный балл"`
	Factors              []factorDTO     `json:"factors"`
	Suggested            *suggestedDTO   `json:"suggested"`
	ModelVersion         string          `json:"modelVersion" example:"rules-v1"`
}

func toScoreResponse(s *shared.ScoreView) scoreResponse {
	resp := scoreResponse{
		Score: s.Score, RiskBand: s.RiskBand, AutoDecisionEligible: s.AutoDecisionEligible,
		AutoDecisionReason: s.AutoDecisionReason, Factors: make([]factorDTO, len(s.Factors)), ModelVersion: s.ModelVersion,
	}
	for i, f := range s.Factors {
		resp.Factors[i] = factorDTO{Code: f.Code, Label: f.Label, Impact: f.Impact}
	}
	if s.Suggested != nil {
		resp.Suggested = &suggestedDTO{Decision: s.Suggested.Decision, Amount: s.Suggested.Amount, Rate: s.Suggested.Rate}
	}
	return resp
}

// bankScore godoc
//
//	@Summary		Скоринг заявки
//	@Description	Итоговый балл, вклад каждого фактора и рекомендация решения. Только для сотрудника банка.
//	@Tags			bank
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	scoreResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/score [get]
func (m *Module) bankScore(c *gin.Context) {
	r, err := m.store.LatestScore(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if r == nil {
		httpx.Fail(c, httpx.NotFound("Скоринг по заявке ещё не выполнен"))
		return
	}
	httpx.JSON(c, http.StatusOK, toScoreResponse(r.view()))
}
