package servicing

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

const formatPDF = "pdf"

// document godoc
//
//	@Summary		PDF полного графика
//	@Description	Условия договора, все строки действующего графика с оплаченными и оставшимися суммами, итоги и версии.
//	@Description	Ошибки — JSON.
//	@Tags			loans
//	@Produce		application/pdf
//	@Param			id		path		string	true	"ID кредита"
//	@Param			format	query		string	false	"Только pdf"	default(pdf)
//	@Success		200		{file}		file
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/loans/{id}/schedule/document [get]
func (m *Module) document(c *gin.Context) {
	company, err := clientCompany(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if f := c.DefaultQuery("format", formatPDF); f != formatPDF {
		httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "format", Code: "INVALID", Message: "Доступен только формат pdf"}))
		return
	}
	ctx := c.Request.Context()
	l, err := m.ownLoan(ctx, company, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := m.snapshot(ctx, l.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	pdf, err := schedulePDF(s, m.today())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	// id кредита — латиница, цифры и дефис: имя файла безопасно без экранирования.
	c.Header("Content-Disposition", `attachment; filename="schedule-`+l.ID+`.pdf"`)
	c.Header("Cache-Control", "private, no-store")
	c.Data(http.StatusOK, "application/pdf", pdf)
}
