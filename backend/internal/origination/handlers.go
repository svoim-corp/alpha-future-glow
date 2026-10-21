package origination

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

type consentRequest struct {
	Given *bool  `json:"given" binding:"required" example:"true"`
	Text  string `json:"text" binding:"required,max=1000" example:"Согласен на запрос моей кредитной истории в бюро кредитных историй"`
}

type consentResponse struct {
	Given bool      `json:"given" example:"true"`
	By    string    `json:"by" example:"u-1"`
	At    time.Time `json:"at" example:"2026-10-20T10:05:00Z"`
}

// mustGive — согласие нельзя «не дать»: без него нет этапа 1, а отказ клиента — просто отсутствие запроса.
func mustGive(given *bool) error {
	if !*given {
		return httpx.Validation("", httpx.FieldError{Field: "given", Code: "INVALID", Message: "Без согласия банк не сможет запросить кредитную историю"})
	}
	return nil
}

// consentBKI godoc
//
//	@Summary		Согласие на запрос в БКИ
//	@Description	Фиксирует согласие заявителя: кто, когда, по какой заявке. Нужно для этапа 1. Повтор возвращает уже данное.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"ID заявки"
//	@Param			body	body		consentRequest	true	"Согласие"
//	@Success		201		{object}	consentResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody	"Нет полномочия BKI_CONSENT по доверенности"
//	@Failure		404		{object}	httpx.ErrorBody
//	@Failure		409		{object}	httpx.ErrorBody
//	@Router			/applications/{id}/consents/bki [post]
func (m *Module) consentBKI(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req consentRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := mustGive(req.Given); err != nil {
		httpx.Fail(c, err)
		return
	}
	cs, err := m.giveBKIConsent(c.Request.Context(), c.Param("id"), who, req.Text)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, consentResponse{Given: true, By: cs.GivenBy, At: cs.GivenAt})
}

// delete godoc
//
//	@Summary		Удалить черновик
//	@Description	Только черновик. Заявка закрывается (DRAFT → CLOSED) и пропадает из списка клиента.
//	@Tags			applications
//	@Param			id	path	string	true	"ID заявки"
//	@Success		204
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		409	{object}	httpx.ErrorBody
//	@Router			/applications/{id} [delete]
func (m *Module) delete(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := m.deleteDraft(c.Request.Context(), c.Param("id"), who); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type guarantorConsentRequest struct {
	Given *bool `json:"given" binding:"required" example:"true"`
}

type guarantorConsentResponse struct {
	Given bool      `json:"given" example:"true"`
	At    time.Time `json:"at" example:"2026-10-20T10:20:00Z"`
}

// consentGuarantor godoc
//
//	@Summary		Согласие поручителя на запрос в БКИ
//	@Description	Для выбранного поручителя, на этапе 2. При смене поручителя согласие нужно дать заново.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID заявки"
//	@Param			body	body		guarantorConsentRequest	true	"Согласие"
//	@Success		201		{object}	guarantorConsentResponse
//	@Failure		400		{object}	httpx.ErrorBody	"Поручитель не выбран"
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Failure		409		{object}	httpx.ErrorBody	"Не черновик или этап 1 не пройден"
//	@Router			/applications/{id}/guarantor/consent-bki [post]
func (m *Module) consentGuarantor(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req guarantorConsentRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := mustGive(req.Given); err != nil {
		httpx.Fail(c, err)
		return
	}
	cs, err := m.giveGuarantorConsent(c.Request.Context(), c.Param("id"), who)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, guarantorConsentResponse{Given: true, At: cs.GivenAt})
}

type missingDTO struct {
	Field   string `json:"field" example:"stage2.guarantor.personId"`
	Message string `json:"message" example:"Выберите поручителя"`
}

type validateResponse struct {
	Valid   bool         `json:"valid" example:"false"`
	Missing []missingDTO `json:"missing"`
}

// validate godoc
//
//	@Summary		Проверка полноты перед отправкой
//	@Description	Что ещё нужно для отправки. Подтверждение достоверности приходит при отправке, здесь оно всегда в списке.
//	@Tags			applications
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	validateResponse
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		409	{object}	httpx.ErrorBody
//	@Failure		502	{object}	httpx.ErrorBody
//	@Router			/applications/{id}/validate [post]
func (m *Module) validate(c *gin.Context) {
	who, err := client(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	a, err := m.draftOf(ctx, c.Param("id"), who.companyID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	missing, err := m.checkDraft(ctx, a, a.TruthConfirmed)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	resp := validateResponse{Valid: len(missing) == 0, Missing: make([]missingDTO, len(missing))}
	for i, f := range missing {
		resp.Missing[i] = missingDTO{Field: f.Field, Message: f.Message}
	}
	httpx.JSON(c, http.StatusOK, resp)
}

type submitRequest struct {
	TruthConfirmed *bool `json:"truthConfirmed" binding:"required" example:"true"`
}

type submitResponse struct {
	ID       string        `json:"id" example:"a-1031"`
	Number   string        `json:"number" example:"КЗ-2026-001031"`
	Status   shared.Status `json:"status" example:"SUBMITTED"`
	Redirect string        `json:"redirect" example:"/applications/a-1031"`
}

// submitApplication godoc
//
//	@Summary		Отправить заявку в банк
//	@Description	Проверка полноты, номер заявки, статус SUBMITTED и запуск банковского конвейера в фоне.
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"ID заявки"
//	@Param			If-Match		header		string			true	"Версия заявки"
//	@Param			Idempotency-Key	header		string			false	"Ключ идемпотентности"
//	@Param			body			body		submitRequest	true	"Подтверждение достоверности"
//	@Success		200				{object}	submitResponse
//	@Failure		400				{object}	httpx.ErrorBody	"Заявка заполнена не полностью — поля как в validate"
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		404				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody
//	@Failure		412				{object}	httpx.ErrorBody
//	@Failure		502				{object}	httpx.ErrorBody
//	@Router			/applications/{id}/submit [post]
func (m *Module) submitApplication(c *gin.Context) {
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
	var req submitRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	a, err := m.submit(c.Request.Context(), c.Param("id"), who, version, *req.TruthConfirmed)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, submitResponse{ID: a.ID, Number: *a.Number, Status: shared.Status(a.Status), Redirect: "/applications/" + a.ID})
}

type timelineCurrentDTO struct {
	Status         shared.Status `json:"status" example:"INFO_REQUESTED"`
	Label          string        `json:"label" example:"Нужна доп. информация"`
	WhatsHappening string        `json:"whatsHappening" example:"Банк проверил заявку и просит уточнить данные"`
	NextStep       string        `json:"nextStep" example:"Ответьте на запрос банка ниже"`
	Actor          string        `json:"actor" example:"CLIENT"` // чей ход: CLIENT или BANK
}

type timelineStepDTO struct {
	Status shared.Status `json:"status" example:"SUBMITTED"`
	Label  string        `json:"label" example:"Подана"`
	At     *time.Time    `json:"at" example:"2026-10-20T10:25:00Z"`
	State  string        `json:"state" example:"DONE"` // DONE, CURRENT, PENDING
}

type timelineResponse struct {
	Current timelineCurrentDTO `json:"current"`
	Steps   []timelineStepDTO  `json:"steps"`
}

// timeline godoc
//
//	@Summary		Таймлайн заявки для клиента
//	@Description	Что сейчас происходит, что делать дальше, чей ход, и шаги: пройденные, текущий и ожидаемые.
//	@Tags			applications
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	timelineResponse
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/applications/{id}/timeline [get]
func (m *Module) timeline(c *gin.Context) {
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
	events, err := m.store.Events(ctx, a.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	st := shared.Status(a.Status)
	txt := statusTexts[st]
	resp := timelineResponse{
		Current: timelineCurrentDTO{Status: st, Label: st.Label(), WhatsHappening: txt.whatsHappening, NextStep: txt.nextStep, Actor: txt.turn},
	}
	for _, s := range steps(st, events) {
		resp.Steps = append(resp.Steps, timelineStepDTO{Status: s.status, Label: s.status.Label(), At: s.at, State: s.state})
	}
	if resp.Steps == nil {
		resp.Steps = []timelineStepDTO{}
	}
	httpx.JSON(c, http.StatusOK, resp)
}

type historyActorDTO struct {
	Type shared.ActorType `json:"type" example:"OFFICER"`
	Name *string          `json:"name,omitempty" example:"Иванова Е. П."` // у системы нет
}

type historyEntryDTO struct {
	At      time.Time       `json:"at" example:"2026-10-20T13:10:00Z"`
	Actor   historyActorDTO `json:"actor"`
	Action  string          `json:"action" example:"DECISION"`
	From    shared.Status   `json:"from" example:"ON_DECISION"`
	To      shared.Status   `json:"to" example:"COUNTER_OFFER"`
	Comment *string         `json:"comment,omitempty" example:"Сумма снижена по уровню долговой нагрузки"`
}

// shortName — «Иванова Елена Петровна» → «Иванова Е. П.»: так подписывают действия в истории.
func shortName(full string) string {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " " + string([]rune(p)[0]) + "."
	}
	return out
}

// bankHistory godoc
//
//	@Summary		История статусов и действий по заявке
//	@Description	Кто, когда и каким действием менял статус заявки, по порядку. Только для сотрудника банка.
//	@Tags			bank
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{array}		historyEntryDTO
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/history [get]
func (m *Module) bankHistory(c *gin.Context) {
	ctx := c.Request.Context()
	a, err := m.store.Application(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if a == nil {
		httpx.Fail(c, httpx.NotFound("Заявка не найдена"))
		return
	}
	events, err := m.store.Events(ctx, a.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]historyEntryDTO, len(events))
	for i, ev := range events {
		out[i] = historyEntryDTO{
			At: ev.At, Actor: historyActorDTO{Type: shared.ActorType(ev.ActorType)}, Action: ev.Action,
			From: shared.Status(ev.From), To: shared.Status(ev.To), Comment: ev.Comment,
		}
		if ev.ActorName != nil {
			out[i].Actor.Name = optional(shortName(*ev.ActorName))
		}
	}
	httpx.JSON(c, http.StatusOK, out)
}
