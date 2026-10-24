package planner

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// getPlan godoc
// @Summary План по ID, включая архив
// @Tags planner
// @Produce json
// @Param id path string true "ID плана"
// @Success 200 {object} planDTO
// @Failure 401 {object} httpx.ErrorBody
// @Failure 403 {object} httpx.ErrorBody
// @Failure 404 {object} httpx.ErrorBody
// @Router /plans/{id} [get]
func (m *Module) getPlan(c *gin.Context) {
	if auth.UserFrom(c).Role != shared.RoleClient {
		httpx.Fail(c, httpx.Forbidden("Планер доступен клиенту банка"))
		return
	}
	view, err := m.planByID(c.Request.Context(), companyOf(c), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toPlanDTO(*view))
}
