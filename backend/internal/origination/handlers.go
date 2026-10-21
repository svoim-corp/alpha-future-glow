package origination

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

// clientTurn — статусы, в которых ход за клиентом: ответить банку, принять оффер, подписать договор.
var clientTurn = map[shared.Status]bool{
	shared.StatusInfoRequested: true, shared.StatusApproved: true, shared.StatusCounterOffer: true, shared.StatusSigning: true,
}

type productRefDTO struct {
	Code string `json:"code" example:"WORKING"`
	Name string `json:"name" example:"Оборотный кредит"`
}

type applicationShortDTO struct {
	ID             string        `json:"id" example:"a-1024"`
	Number         *string       `json:"number" example:"КЗ-2026-001024"` // null у черновика
	Product        productRefDTO `json:"product"`
	Amount         int64         `json:"amount" example:"1500000"`
	Term           int           `json:"term" example:"24"`
	Status         shared.Status `json:"status" example:"INFO_REQUESTED"`
	StatusLabel    string        `json:"statusLabel" example:"Нужна доп. информация"`
	ActionRequired bool          `json:"actionRequired" example:"true"`
	UpdatedAt      time.Time     `json:"updatedAt" example:"2026-10-20T12:40:00Z"`
}

// applicationsPage — схема ответа для swag: обобщённый httpx.Page он без --parseDependency не разбирает.
// Отдаётся по-прежнему httpx.NewPage, JSON тот же.
type applicationsPage struct {
	Items []applicationShortDTO `json:"items"`
	Page  int                   `json:"page" example:"1"`
	Size  int                   `json:"size" example:"20"`
	Total int                   `json:"total" example:"2"`
}

// productNames — названия продуктов для списка. Продуктов единицы, поэтому каждый читается один раз на запрос.
type productNames struct {
	products shared.Products
	names    map[string]string
}

func (p *productNames) ref(ctx context.Context, code string) (productRefDTO, error) {
	if name, ok := p.names[code]; ok {
		return productRefDTO{Code: code, Name: name}, nil
	}
	v, err := p.products.Product(ctx, code)
	if err != nil {
		return productRefDTO{}, fmt.Errorf("load product %s: %w", code, err)
	}
	name := code
	if v != nil {
		name = v.Name
	}
	p.names[code] = name
	return productRefDTO{Code: code, Name: name}, nil
}

// list godoc
//
//	@Summary		Заявки компании
//	@Description	Блок «Мои заявки»: черновики и поданные заявки компании, недавно изменённые первыми.
//	@Tags			applications
//	@Produce		json
//	@Param			mine	query		bool	false	"Заявки своей компании (клиент видит только их)"
//	@Param			page	query		int		false	"Страница"	default(1)
//	@Param			size	query		int		false	"Размер страницы (до 100)"	default(20)
//	@Success		200		{object}	applicationsPage
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/applications [get]
func (m *Module) list(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	page, size := httpx.ParsePage(c)
	rows, total, err := m.store.CompanyApplications(ctx, who.companyID, size, (page-1)*size)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	names := &productNames{products: m.ports.Products, names: map[string]string{}}
	items := make([]applicationShortDTO, len(rows))
	for i, r := range rows {
		ref, err := names.ref(ctx, r.ProductCode)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		st := shared.Status(r.Status)
		items[i] = applicationShortDTO{
			ID: r.ID, Number: r.Number, Product: ref, Amount: r.Amount, Term: r.Term,
			Status: st, StatusLabel: st.Label(), ActionRequired: clientTurn[st], UpdatedAt: r.UpdatedAt,
		}
	}
	httpx.JSON(c, http.StatusOK, httpx.NewPage(items, page, size, total))
}
