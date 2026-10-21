package support

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

// steps — шаги пути клиента, по которым сгруппированы частые вопросы.
var steps = []string{"PRODUCT", "APPLICATION", "DOCUMENTS", "STATUS", "DECISION"}

type faqDTO struct {
	ID       string `json:"id" example:"f-4"`
	Step     string `json:"step" example:"APPLICATION"`
	Question string `json:"question" example:"Нужно ли загружать документы или выписки?"`
	Answer   string `json:"answer" example:"Нет. Данные о компании и оборотах банк берёт из ваших счетов и реестров."`
}

// faq godoc
//
//	@Summary		Частые вопросы
//	@Description	Вопросы по шагу пути клиента. Без step — по всем шагам.
//	@Tags			support
//	@Produce		json
//	@Param			step	query		string	false	"Шаг пути"	Enums(PRODUCT, APPLICATION, DOCUMENTS, STATUS, DECISION)
//	@Success		200		{array}		faqDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Router			/support/faq [get]
func (m *Module) faq(c *gin.Context) {
	step := c.Query("step")
	if step != "" && !slices.Contains(steps, step) {
		httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "step", Code: "INVALID", Message: "Недопустимое значение"}))
		return
	}
	rows, err := m.store.FAQ(c.Request.Context(), step)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]faqDTO, len(rows))
	for i, r := range rows {
		out[i] = faqDTO{ID: r.ID, Step: r.Step, Question: r.Question, Answer: r.Answer}
	}
	httpx.JSON(c, http.StatusOK, out)
}

type ticketRequest struct {
	Subject       string  `json:"subject" binding:"required,max=200" example:"Устаревший фактический адрес"`
	Text          string  `json:"text" binding:"required,max=4000" example:"В заявке старый адрес склада, мы переехали"`
	ApplicationID *string `json:"applicationId" example:"a-1031"` // со страницы заявки подставляется сам
}

type ticketCreatedDTO struct {
	ID        string    `json:"id" example:"t-12"`
	Status    string    `json:"status" example:"OPEN"`
	CreatedAt time.Time `json:"createdAt" example:"2026-10-20T11:00:00Z"`
}

// createTicket godoc
//
//	@Summary		Новое обращение в поддержку
//	@Description	applicationId — необязательно, только заявка своей компании.
//	@Tags			support
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ticketRequest	true	"Обращение"
//	@Success		201		{object}	ticketCreatedDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/support/tickets [post]
func (m *Module) createTicket(c *gin.Context) {
	u := auth.UserFrom(c)
	if u.CompanyID == nil {
		httpx.Fail(c, httpx.Forbidden("Обращения пишет клиент банка"))
		return
	}
	var req ticketRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.ApplicationID != nil && *req.ApplicationID == "" {
		req.ApplicationID = nil
	}
	t := &ticketRow{UserID: u.ID, CompanyID: *u.CompanyID, ApplicationID: req.ApplicationID, Subject: req.Subject, Text: req.Text}
	if err := m.openTicket(c.Request.Context(), t); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, ticketCreatedDTO{ID: t.ID, Status: t.Status, CreatedAt: t.CreatedAt})
}
