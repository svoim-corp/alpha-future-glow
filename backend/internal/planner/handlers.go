package planner

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

type notificationsDTO struct {
	InApp bool `json:"inApp" example:"true"`
	Email bool `json:"email" example:"false"`
}

type recommendationDTO struct {
	ProductCode   string `json:"productCode" example:"INVEST"`
	ProductName   string `json:"productName" example:"Инвестиционный кредит"`
	Reason        string `json:"reason" example:"Цель и сумма укладываются в рамки продукта"`
	AvailableFrom string `json:"availableFrom" example:"2026-10-20"`
}

type planDTO struct {
	ID             string             `json:"id" example:"p-1"`
	Title          string             `json:"title" example:"Ремонт помещения"`
	Note           *string            `json:"note" example:"Обновить торговый зал и зону выпечки"`
	Goal           string             `json:"goal" example:"REAL_ESTATE"`
	BudgetRub      int64              `json:"budgetRub" example:"2400000"`
	DateMode       string             `json:"dateMode" example:"PERIOD"`
	StartDate      *string            `json:"startDate" example:"2027-06-01"`
	EndDate        *string            `json:"endDate" example:"2027-08-31"`
	State          string             `json:"state" example:"ACTIVE"`
	Notifications  notificationsDTO   `json:"notifications"`
	Recommendation *recommendationDTO `json:"recommendation"`
	CreatedAt      time.Time          `json:"createdAt" example:"2026-10-20T09:00:00Z"`
}

func dateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(dateLayout)
	return &s
}

func toPlanDTO(v planView) planDTO {
	p := v.plan
	d := planDTO{
		ID: p.ID, Title: p.Title, Note: p.Note, Goal: p.Goal, BudgetRub: p.BudgetRub, DateMode: p.DateMode,
		StartDate: dateStr(p.StartDate), EndDate: dateStr(p.EndDate), State: p.State,
		Notifications: notificationsDTO{InApp: p.NotifyInApp, Email: p.NotifyEmail}, CreatedAt: p.CreatedAt,
	}
	if r := v.rec; r != nil {
		d.Recommendation = &recommendationDTO{
			ProductCode: r.product.Code, ProductName: r.product.Name, Reason: matchReason, AvailableFrom: r.availableFrom.Format(dateLayout),
		}
	}
	return d
}

func companyOf(c *gin.Context) string {
	if u := auth.UserFrom(c); u.CompanyID != nil {
		return *u.CompanyID
	}
	return ""
}

type summaryDTO struct {
	Total       int   `json:"total" example:"3"`
	WithPeriod  int   `json:"withPeriod" example:"2"`
	Unscheduled int   `json:"unscheduled" example:"1"`
	BudgetRub   int64 `json:"budgetRub" example:"4250000"` // сумма ориентиров, а не кредитных обязательств
}

type plansResponse struct {
	Items   []planDTO  `json:"items"`
	Summary summaryDTO `json:"summary"`
}

// listPlans godoc
//
//	@Summary		Планы компании
//	@Description	По умолчанию — все, кроме ARCHIVED. recommendation — подбор на сегодня или null.
//	@Tags			planner
//	@Produce		json
//	@Param			state	query		string	false	"ACTIVE, PAUSED, DONE, ARCHIVED"
//	@Success		200		{object}	plansResponse
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Router			/plans [get]
func (m *Module) listPlans(c *gin.Context) {
	state := c.Query("state")
	if state != "" && !validStates[state] {
		httpx.Fail(c, httpx.Validation("", httpx.FieldError{Field: "state", Code: "INVALID", Message: "Недопустимое значение"}))
		return
	}
	views, err := m.list(c.Request.Context(), companyOf(c), state)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	resp := plansResponse{Items: make([]planDTO, len(views))}
	for i, v := range views {
		resp.Items[i] = toPlanDTO(v)
		resp.Summary.Total++
		resp.Summary.BudgetRub += v.plan.BudgetRub
		if v.plan.DateMode == modeUnscheduled {
			resp.Summary.Unscheduled++
		} else {
			resp.Summary.WithPeriod++
		}
	}
	httpx.JSON(c, http.StatusOK, resp)
}

var validStates = map[string]bool{stateActive: true, "PAUSED": true, "DONE": true, stateArchived: true}

type notificationsInput struct {
	InApp *bool `json:"inApp" example:"true"`
	Email *bool `json:"email" example:"false"`
}

type createPlanRequest struct {
	Title         string              `json:"title" binding:"required,max=200" example:"Новая печь для пекарни"`
	Note          *string             `json:"note" binding:"omitempty,max=2000" example:"Увеличить выпуск хлеба"`
	Goal          string              `json:"goal" binding:"required,max=64" example:"EQUIPMENT"`
	BudgetRub     int64               `json:"budgetRub" binding:"required,gt=0" example:"1200000"`
	DateMode      string              `json:"dateMode" binding:"required,oneof=DAY PERIOD UNSCHEDULED" example:"UNSCHEDULED"`
	StartDate     *string             `json:"startDate" binding:"omitempty,datetime=2006-01-02" example:"2027-06-01"`
	EndDate       *string             `json:"endDate" binding:"omitempty,datetime=2006-01-02" example:"2027-08-31"`
	Notifications *notificationsInput `json:"notifications"`
}

type patchPlanRequest struct {
	Title         *string             `json:"title" binding:"omitempty,min=1,max=200" example:"Новая печь для пекарни"`
	Note          *string             `json:"note" binding:"omitempty,max=2000" example:""`
	Goal          *string             `json:"goal" binding:"omitempty,min=1,max=64" example:"EQUIPMENT"`
	BudgetRub     *int64              `json:"budgetRub" binding:"omitempty,gt=0" example:"1500000"`
	DateMode      *string             `json:"dateMode" binding:"omitempty,oneof=DAY PERIOD UNSCHEDULED" example:"DAY"`
	StartDate     *string             `json:"startDate" binding:"omitempty,datetime=2006-01-02" example:"2027-06-01"`
	EndDate       *string             `json:"endDate" binding:"omitempty,datetime=2006-01-02" example:"2027-06-01"`
	State         *string             `json:"state" binding:"omitempty,oneof=ACTIVE PAUSED DONE ARCHIVED" example:"PAUSED"`
	Notifications *notificationsInput `json:"notifications"`
}

// parseDate — формат уже проверен тегом datetime.
func parseDate(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, _ := time.Parse(dateLayout, *s)
	return &t
}

func (n *notificationsInput) values() (inApp, email *bool) {
	if n == nil {
		return nil, nil
	}
	return n.InApp, n.Email
}

// createPlan godoc
//
//	@Summary	Создать план
//	@Tags		planner
//	@Accept		json
//	@Produce	json
//	@Param		body	body		createPlanRequest	true	"План"
//	@Success	201		{object}	planDTO
//	@Failure	400		{object}	httpx.ErrorBody
//	@Failure	401		{object}	httpx.ErrorBody
//	@Failure	403		{object}	httpx.ErrorBody
//	@Router		/plans [post]
func (m *Module) createPlan(c *gin.Context) {
	var req createPlanRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	inApp, email := req.Notifications.values()
	if inApp == nil {
		inApp = ptr(true) // уведомления в кабинете — по умолчанию включены
	}
	in := planInput{
		Title: &req.Title, Note: req.Note, Goal: &req.Goal, BudgetRub: &req.BudgetRub, DateMode: &req.DateMode,
		StartDate: parseDate(req.StartDate), EndDate: parseDate(req.EndDate), InApp: inApp, Email: email,
	}
	v, err := m.create(c.Request.Context(), companyOf(c), auth.UserFrom(c).ID, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, toPlanDTO(*v))
}

func ptr[T any](v T) *T { return &v }

// patchPlan godoc
//
//	@Summary		Изменить план
//	@Description	Только переданные поля. Пустая строка в note очищает описание; даты сбрасываются переходом на UNSCHEDULED.
//	@Tags			planner
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"ID плана"
//	@Param			body	body		patchPlanRequest	true	"Изменения"
//	@Success		200		{object}	planDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/plans/{id} [patch]
func (m *Module) patchPlan(c *gin.Context) {
	var req patchPlanRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	inApp, email := req.Notifications.values()
	in := planInput{
		Title: req.Title, Note: req.Note, Goal: req.Goal, BudgetRub: req.BudgetRub, DateMode: req.DateMode, State: req.State,
		StartDate: parseDate(req.StartDate), EndDate: parseDate(req.EndDate), InApp: inApp, Email: email,
	}
	v, err := m.update(c.Request.Context(), companyOf(c), c.Param("id"), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toPlanDTO(*v))
}

// deletePlan godoc
//
//	@Summary	Удалить план
//	@Tags		planner
//	@Param		id	path	string	true	"ID плана"
//	@Success	204
//	@Failure	401	{object}	httpx.ErrorBody
//	@Failure	403	{object}	httpx.ErrorBody
//	@Failure	404	{object}	httpx.ErrorBody
//	@Router		/plans/{id} [delete]
func (m *Module) deletePlan(c *gin.Context) {
	if err := m.remove(c.Request.Context(), companyOf(c), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type planRecommendationDTO struct {
	PlanID      string `json:"planId" example:"p-3"`
	PlanTitle   string `json:"planTitle" example:"Новая печь для пекарни"`
	BudgetRub   int64  `json:"budgetRub" example:"1200000"`
	ProductCode string `json:"productCode" example:"INVEST"`
	ProductName string `json:"productName" example:"Инвестиционный кредит"`
	Reason      string `json:"reason" example:"Сумма и цель укладываются в рамки продукта"`
}

// planRecommendations godoc
//
//	@Summary		Планы с подходящим продуктом
//	@Description	Виджет главной. Пустой массив — виджет не показывается.
//	@Tags			planner
//	@Produce		json
//	@Success		200	{array}		planRecommendationDTO
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Router			/plans/recommendations [get]
func (m *Module) planRecommendations(c *gin.Context) {
	views, err := m.recommendations(c.Request.Context(), companyOf(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]planRecommendationDTO, len(views))
	for i, v := range views {
		out[i] = planRecommendationDTO{
			PlanID: v.plan.ID, PlanTitle: v.plan.Title, BudgetRub: v.plan.BudgetRub,
			ProductCode: v.rec.product.Code, ProductName: v.rec.product.Name, Reason: matchReason,
		}
	}
	httpx.JSON(c, http.StatusOK, out)
}
