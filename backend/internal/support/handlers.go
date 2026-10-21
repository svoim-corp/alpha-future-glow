package support

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

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
