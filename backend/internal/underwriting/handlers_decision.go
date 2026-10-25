package underwriting

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func officerOf(c *gin.Context) shared.Actor {
	u := auth.UserFrom(c)
	return shared.Actor{Type: shared.ActorOfficer, UserID: u.ID, Name: u.FullName}
}

type infoRequestBody struct {
	Comment string     `json:"comment" binding:"required,max=2000" example:"Обороты по счетам расходятся с отчётностью"`
	Items   []infoItem `json:"items" binding:"required,min=1,max=10,dive"`
}

type infoRequestCreated struct {
	ID                string        `json:"id" example:"r-7"`
	Status            string        `json:"status" example:"OPEN"`
	ApplicationStatus shared.Status `json:"applicationStatus" example:"INFO_REQUESTED"`
}

// bankRequestInfo godoc
//
//	@Summary		Запросить информацию у клиента
//	@Description	Только пояснения: данные, которые есть у банка, и документы не запрашиваются. Заявка → INFO_REQUESTED.
//	@Tags			bank
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"ID заявки"
//	@Param			body	body		infoRequestBody	true	"Вопросы"
//	@Success		201		{object}	infoRequestCreated
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Failure		409		{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/info-requests [post]
func (m *Module) bankRequestInfo(c *gin.Context) {
	var req infoRequestBody
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	r, err := m.requestInfo(c.Request.Context(), c.Param("id"), req.Comment, req.Items, officerOf(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, infoRequestCreated{ID: r.ID, Status: r.Status, ApplicationStatus: shared.StatusInfoRequested})
}

type infoRequestDTO struct {
	ID         string       `json:"id" example:"r-7"`
	Status     string       `json:"status" example:"OPEN"`
	Comment    string       `json:"comment" example:"Обороты по счетам расходятся с отчётностью"`
	Items      []infoItem   `json:"items"`
	Answers    []infoAnswer `json:"answers,omitempty"`
	CreatedAt  time.Time    `json:"createdAt" example:"2026-10-20T12:40:00Z"`
	AnsweredAt *time.Time   `json:"answeredAt,omitempty"`
}

// clientInfoRequests godoc
//
//	@Summary	Запросы банка по заявке
//	@Tags		applications
//	@Produce	json
//	@Param		id	path	string	true	"ID заявки"
//	@Success	200	{array}	infoRequestDTO
//	@Failure	401	{object}	httpx.ErrorBody
//	@Failure	403	{object}	httpx.ErrorBody
//	@Failure	404	{object}	httpx.ErrorBody
//	@Router		/applications/{id}/info-requests [get]
func (m *Module) clientInfoRequests(c *gin.Context) {
	ctx := c.Request.Context()
	app, err := m.ownApplication(ctx, c.Param("id"), companyOf(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out, err := m.infoRequestDTOs(ctx, app.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, out)
}

// infoRequestDTOs — запросы банка с ответами клиента, от старых к новым; общий вид для клиента и сотрудника.
func (m *Module) infoRequestDTOs(ctx context.Context, appID string) ([]infoRequestDTO, error) {
	rows, err := m.store.InfoRequests(ctx, appID)
	if err != nil {
		return nil, err
	}
	out := make([]infoRequestDTO, len(rows))
	for i, r := range rows {
		out[i] = infoRequestDTO{ID: r.ID, Status: r.Status, Comment: r.Comment, Items: r.Items.V, Answers: r.Answers.V, CreatedAt: r.CreatedAt, AnsweredAt: r.AnsweredAt}
	}
	return out, nil
}

// bankInfoRequests godoc
//
//	@Summary		Запросы информации и ответы клиента
//	@Description	Все запросы по заявке с вопросами и ответами клиента: сотрудник видит, что именно ответили.
//	@Tags			bank
//	@Produce		json
//	@Param			id	path	string	true	"ID заявки"
//	@Success		200	{array}	infoRequestDTO
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/info-requests [get]
func (m *Module) bankInfoRequests(c *gin.Context) {
	ctx := c.Request.Context()
	app, err := m.load(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out, err := m.infoRequestDTOs(ctx, app.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, out)
}

// companyOf — компания клиента; у сотрудника банка её нет, и клиентские данные он здесь не видит.
func companyOf(c *gin.Context) string {
	if u := auth.UserFrom(c); u.CompanyID != nil {
		return *u.CompanyID
	}
	return ""
}

type answerBody struct {
	Answers []infoAnswer `json:"answers" binding:"required,min=1,max=10,dive"`
}

type answerResponse struct {
	RequestStatus     string        `json:"requestStatus" example:"ANSWERED"`
	ApplicationStatus shared.Status `json:"applicationStatus" example:"DATA_CHECK"`
}

// clientAnswerInfo godoc
//
//	@Summary		Ответить на запрос банка
//	@Description	Ответ на каждый вопрос запроса. Заявка возвращается в DATA_CHECK.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string		true	"ID заявки"
//	@Param			reqId	path		string		true	"ID запроса"
//	@Param			body	body		answerBody	true	"Ответы"
//	@Success		200		{object}	answerResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Failure		409		{object}	httpx.ErrorBody
//	@Router			/applications/{id}/info-requests/{reqId}/answer [post]
func (m *Module) clientAnswerInfo(c *gin.Context) {
	var req answerBody
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	u := auth.UserFrom(c)
	client := shared.Actor{Type: shared.ActorClient, UserID: u.ID, Name: u.FullName}
	if err := m.answerInfo(c.Request.Context(), c.Param("id"), companyOf(c), c.Param("reqId"), req.Answers, client); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, answerResponse{RequestStatus: requestAnswered, ApplicationStatus: shared.StatusDataCheck})
}

type decisionBody struct {
	Type        string  `json:"type" binding:"required,oneof=APPROVE COUNTER REJECT" example:"COUNTER"`
	Amount      int64   `json:"amount" binding:"required_if=Type COUNTER,gte=0" example:"1200000"`
	Term        int     `json:"term" binding:"required_if=Type COUNTER,gte=0" example:"24"`
	ProductCode string  `json:"productCode" binding:"required_if=Type COUNTER" example:"WORKING"`
	Rate        float64 `json:"rate" binding:"required_unless=Type REJECT,gte=0" example:"21.5"`
	Reason      string  `json:"reason" binding:"required_if=Type REJECT,max=2000" example:"Сумма снижена по уровню долговой нагрузки"`
}

type decisionResponse struct {
	Status     shared.Status       `json:"status" example:"COUNTER_OFFER"`
	DecisionID string              `json:"decisionId" example:"dec-31"`
	Mode       shared.DecisionMode `json:"mode" example:"MANUAL"`
	DecidedBy  *string             `json:"decidedBy" example:"u-9"`
	DecidedAt  time.Time           `json:"decidedAt" example:"2026-10-20T13:10:00Z"`
}

// bankDecision godoc
//
//	@Summary		Решение сотрудника
//	@Description	APPROVE — одобрить (без параметров — на условиях заявки), COUNTER — иные условия, REJECT — отказ (нужен только reason).
//	@Tags			bank
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"ID заявки"
//	@Param			Idempotency-Key	header		string			false	"Ключ идемпотентности"
//	@Param			body			body		decisionBody	true	"Решение"
//	@Success		200				{object}	decisionResponse
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		401				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		422				{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/decision [post]
func (m *Module) bankDecision(c *gin.Context) {
	var req decisionBody
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	cmd := decisionCommand{Type: req.Type, ProductCode: req.ProductCode, Amount: req.Amount, Term: req.Term, Rate: req.Rate, Reason: req.Reason}
	d, err := m.decide(c.Request.Context(), c.Param("id"), cmd, officerOf(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, decisionResponse{Status: d.Result, DecisionID: d.ID, Mode: d.Mode, DecidedBy: d.DecidedBy, DecidedAt: d.DecidedAt})
}
