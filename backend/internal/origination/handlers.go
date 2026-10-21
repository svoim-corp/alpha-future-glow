package origination

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// client — пользователь-клиент со своей компанией. У сотрудника банка компании нет: заявки клиента ему недоступны.
func client(c *gin.Context) (applicant, error) {
	u := auth.UserFrom(c)
	if u.CompanyID == nil {
		return applicant{}, httpx.Forbidden("Заявки подаёт клиент банка от имени своей компании")
	}
	who := applicant{actor: shared.Actor{Type: shared.ActorClient, UserID: u.ID, Name: u.FullName}, companyID: *u.CompanyID}
	if u.PersonINN != nil {
		who.personINN = *u.PersonINN
	}
	return who, nil
}

type createRequest struct {
	ProductCode string  `json:"productCode" binding:"required" example:"WORKING"`
	Amount      int64   `json:"amount" binding:"required,gt=0" example:"1500000"`
	Term        *int    `json:"term" binding:"omitempty,gt=0" example:"24"`         // null — только у продукта с единственным сроком
	Goal        *string `json:"goal" binding:"omitempty" example:"WORKING_CAPITAL"` // null — цель продукта, если она одна
}

type createResponse struct {
	ID       string        `json:"id" example:"a-1031"`
	Status   shared.Status `json:"status" example:"DRAFT"`
	Version  int           `json:"version" example:"1"`
	Redirect string        `json:"redirect" example:"/applications/a-1031/edit"`
}

// create godoc
//
//	@Summary		Создать черновик заявки
//	@Description	Кнопка «Оформить». Компания берётся из пользователя, полномочия заявителя — из данных банка.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createRequest	true	"Параметры из подбора"
//	@Success		201		{object}	createResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody	"Нет полномочий: не руководитель и нет действующей МЧД"
//	@Failure		422		{object}	httpx.ErrorBody
//	@Failure		502		{object}	httpx.ErrorBody
//	@Router			/applications [post]
func (m *Module) create(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req createRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	row, err := m.createDraft(c.Request.Context(), who, draftInput{ProductCode: req.ProductCode, Amount: req.Amount, Term: req.Term, Goal: req.Goal})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, createResponse{
		ID: row.ID, Status: shared.Status(row.Status), Version: row.Version, Redirect: "/applications/" + row.ID + "/edit",
	})
}
