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

// get godoc
//
//	@Summary		Заявка для мастера и карточки
//	@Description	Полное состояние заявки: этап, параметры, данные банка (только чтение), результат этапа 1, блокировки.
//	@Tags			applications
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	applicationDTO
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/applications/{id} [get]
func (m *Module) get(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	a, err := m.ownApplication(ctx, c.Param("id"), who.companyID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	d, err := m.dossier(ctx, a)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toApplicationDTO(d))
}

type patchParamsDTO struct {
	Amount         *int64  `json:"amount" binding:"omitempty,gt=0" example:"1200000"`
	Term           *int    `json:"term" binding:"omitempty,gt=0" example:"24"`
	Goal           *string `json:"goal" example:"WORKING_CAPITAL"`
	PurposeText    *string `json:"purposeText" binding:"omitempty,max=1000" example:"Закупка муки"` // "" — убрать
	GraceMonths    *int    `json:"graceMonths" binding:"omitempty,gte=0" example:"6"`               // 0 — без отсрочки
	OverdraftLimit *int64  `json:"overdraftLimit" binding:"omitempty,gte=0" example:"500000"`       // 0 — убрать
}

type patchGuarantorDTO struct {
	PersonID string `json:"personId" binding:"required" example:"p-1"`
}

type patchStage2DTO struct {
	Guarantor *patchGuarantorDTO `json:"guarantor"`
}

// patchRequest — только то, что выбирает клиент. Данных банка здесь нет: попытка их передать — 400.
type patchRequest struct {
	Params *patchParamsDTO `json:"params"`
	Stage2 *patchStage2DTO `json:"stage2"`
}

type patchResponse struct {
	Version           int       `json:"version" example:"8"`
	SavedAt           time.Time `json:"savedAt" example:"2026-10-20T10:13:00Z"`
	PrelimInvalidated bool      `json:"prelimInvalidated" example:"true"`
}

func (r patchRequest) patch() draftPatch {
	var p draftPatch
	if r.Params != nil {
		p = draftPatch{
			Amount: r.Params.Amount, Term: r.Params.Term, Goal: r.Params.Goal, PurposeText: r.Params.PurposeText,
			GraceMonths: r.Params.GraceMonths, OverdraftLimit: r.Params.OverdraftLimit,
		}
	}
	if r.Stage2 != nil && r.Stage2.Guarantor != nil {
		p.GuarantorPersonID = &r.Stage2.Guarantor.PersonID
	}
	return p
}

// patch godoc
//
//	@Summary		Автосохранение черновика
//	@Description	Передаются только изменённые поля: params и stage2.guarantor.personId. Данные банка не меняются — 400.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"ID заявки"
//	@Param			If-Match	header		string			true	"Версия заявки"
//	@Param			body		body		patchRequest	true	"Изменённые поля"
//	@Success		200			{object}	patchResponse
//	@Failure		400			{object}	httpx.ErrorBody
//	@Failure		403			{object}	httpx.ErrorBody
//	@Failure		404			{object}	httpx.ErrorBody
//	@Failure		409			{object}	httpx.ErrorBody
//	@Failure		412			{object}	httpx.ErrorBody
//	@Failure		422			{object}	httpx.ErrorBody
//	@Router			/applications/{id} [patch]
func (m *Module) patch(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	version, err := httpx.IfMatch(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req patchRequest
	if err := httpx.BindStrict(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	a, invalidated, err := m.patchDraft(c.Request.Context(), c.Param("id"), who.companyID, version, req.patch())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, patchResponse{Version: a.Version, SavedAt: a.UpdatedAt, PrelimInvalidated: invalidated})
}
