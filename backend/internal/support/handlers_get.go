package support

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// getFAQ godoc
// @Summary Частый вопрос по ID
// @Tags support
// @Produce json
// @Param id path string true "ID вопроса"
// @Success 200 {object} faqDTO
// @Failure 404 {object} httpx.ErrorBody
// @Router /support/faq/{id} [get]
func (m *Module) getFAQ(c *gin.Context) {
	row, err := m.store.faqByID(c.Request.Context(), c.Param("id"))
	if err == nil && row == nil {
		err = httpx.NotFound("Вопрос не найден")
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, faqDTO{ID: row.ID, Step: row.Step, Question: row.Question, Answer: row.Answer})
}

// getTicket godoc
// @Summary Обращение текущего пользователя по ID
// @Tags support
// @Produce json
// @Param id path string true "ID обращения"
// @Success 200 {object} ticketDTO
// @Failure 401 {object} httpx.ErrorBody
// @Failure 403 {object} httpx.ErrorBody
// @Failure 404 {object} httpx.ErrorBody
// @Router /support/tickets/{id} [get]
func (m *Module) getTicket(c *gin.Context) {
	u := auth.UserFrom(c)
	if u.Role != shared.RoleClient || u.CompanyID == nil {
		httpx.Fail(c, httpx.Forbidden("Обращения доступны клиенту банка"))
		return
	}
	row, err := m.store.ticketByID(c.Request.Context(), c.Param("id"), u.ID)
	if err == nil && row == nil {
		err = httpx.NotFound("Обращение не найдено")
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, ticketDTO{ID: row.ID, Subject: row.Subject, Text: row.Text, Status: row.Status,
		Reply: row.Reply, ApplicationID: row.ApplicationID, UpdatedAt: row.UpdatedAt.UTC()})
}
