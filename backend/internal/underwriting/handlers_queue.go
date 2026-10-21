package underwriting

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// statsResponse — счётчики над таблицей очереди (API_SPEC.md, 4.11). Ключи — коды статусов, как в спецификации.
type statsResponse struct {
	Submitted     int `json:"SUBMITTED" example:"3"`
	DataCheck     int `json:"DATA_CHECK" example:"2"`
	InfoRequested int `json:"INFO_REQUESTED" example:"4"`
	Scoring       int `json:"SCORING" example:"1"`
	OnDecision    int `json:"ON_DECISION" example:"5"`
	DecidedToday  int `json:"decidedToday" example:"7"`
}

// bankStats godoc
//
//	@Summary	Счётчики очереди
//	@Tags		bank
//	@Produce	json
//	@Success	200	{object}	statsResponse
//	@Failure	401	{object}	httpx.ErrorBody
//	@Failure	403	{object}	httpx.ErrorBody
//	@Router		/bank/applications/stats [get]
func (m *Module) bankStats(c *gin.Context) {
	now := m.now().UTC()
	byStatus, decided, err := m.store.Stats(c.Request.Context(), time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, statsResponse{
		Submitted:     byStatus[string(shared.StatusSubmitted)],
		DataCheck:     byStatus[string(shared.StatusDataCheck)],
		InfoRequested: byStatus[string(shared.StatusInfoRequested)],
		Scoring:       byStatus[string(shared.StatusScoring)],
		OnDecision:    byStatus[string(shared.StatusOnDecision)],
		DecidedToday:  decided,
	})
}

type companyRefDTO struct {
	INN  string `json:"inn" example:"7712345678"`
	Name string `json:"name" example:"ООО «Северная пекарня»"`
}

type productRefDTO struct {
	Code string `json:"code" example:"WORKING"`
	Name string `json:"name" example:"Оборотный кредит"`
}

type assigneeDTO struct {
	ID       string `json:"id" example:"u-9"`
	FullName string `json:"fullName" example:"Иванова Е. П."`
}

type queueItemDTO struct {
	ID           string               `json:"id" example:"a-1024"`
	Number       *string              `json:"number" example:"КЗ-2026-001024"`
	Company      companyRefDTO        `json:"company"`
	Product      productRefDTO        `json:"product"`
	Amount       int64                `json:"amount" example:"1500000"`
	Term         int                  `json:"term" example:"24"`
	Status       shared.Status        `json:"status" example:"ON_DECISION"`
	Score        *int                 `json:"score" example:"640"`
	RiskBand     *shared.RiskBand     `json:"riskBand" example:"MEDIUM"`
	DecisionMode *shared.DecisionMode `json:"decisionMode" example:"MANUAL"`
	Assignee     *assigneeDTO         `json:"assignee"`
	SubmittedAt  time.Time            `json:"submittedAt" example:"2026-10-20T10:25:00Z"`
}

func toQueueItem(r queueRow) queueItemDTO {
	it := queueItemDTO{
		ID: r.AppID, Number: r.Number,
		Company: companyRefDTO{INN: r.CompanyINN, Name: r.CompanyName},
		Product: productRefDTO{Code: r.ProductCode, Name: r.ProductName},
		Amount:  r.Amount, Term: r.Term, Status: shared.Status(r.Status), Score: r.Score, SubmittedAt: r.SubmittedAt,
	}
	if r.RiskBand != nil {
		b := shared.RiskBand(*r.RiskBand)
		it.RiskBand = &b
	}
	if r.DecisionMode != nil {
		d := shared.DecisionMode(*r.DecisionMode)
		it.DecisionMode = &d
	}
	if r.AssigneeID != nil {
		it.Assignee = &assigneeDTO{ID: *r.AssigneeID, FullName: *r.AssigneeName}
	}
	return it
}

// queuePage — схема ответа для swag: обобщённый httpx.Page он без --parseDependency не разбирает.
// Отдаётся по-прежнему httpx.NewPage, JSON тот же.
type queuePage struct {
	Items []queueItemDTO `json:"items"`
	Page  int            `json:"page" example:"1"`
	Size  int            `json:"size" example:"20"`
	Total int            `json:"total" example:"5"`
}

// Допустимые значения фильтров очереди. Черновиков в очереди нет.
var (
	queueStatuses = []shared.Status{
		shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusInfoRequested, shared.StatusScoring, shared.StatusOnDecision,
		shared.StatusApproved, shared.StatusCounterOffer, shared.StatusRejected, shared.StatusSigning, shared.StatusIssued, shared.StatusClosed,
	}
	riskBands     = []shared.RiskBand{shared.RiskLow, shared.RiskMedium, shared.RiskHigh}
	decisionModes = []shared.DecisionMode{shared.DecisionAuto, shared.DecisionManual}
)

func oneOf[T ~string](v string, allowed []T) bool {
	for _, a := range allowed {
		if string(a) == v {
			return true
		}
	}
	return false
}

// parseFilter читает фильтры из query и проверяет перечисления.
func parseFilter(c *gin.Context) (queueFilter, error) {
	f := queueFilter{
		Status: c.Query("status"), Product: c.Query("product"), Risk: c.Query("risk"),
		DecisionMode: c.Query("decisionMode"), Sort: c.Query("sort"),
	}
	var fields []httpx.FieldError
	bad := func(field string) {
		fields = append(fields, httpx.FieldError{Field: field, Code: "INVALID", Message: "Недопустимое значение"})
	}
	if f.Status != "" && !oneOf(f.Status, queueStatuses) {
		bad("status")
	}
	if f.Risk != "" && !oneOf(f.Risk, riskBands) {
		bad("risk")
	}
	if f.DecisionMode != "" && !oneOf(f.DecisionMode, decisionModes) {
		bad("decisionMode")
	}
	if _, ok := sortColumns[strings.TrimPrefix(f.Sort, "-")]; f.Sort != "" && !ok {
		bad("sort")
	}
	if len(fields) > 0 {
		return f, httpx.Validation("Проверьте фильтры", fields...)
	}
	return f, nil
}

// bankQueue godoc
//
//	@Summary		Очередь заявок
//	@Description	Таблица заявок с фильтрами. sort: submittedAt, updatedAt, amount, score; «-» в начале — по убыванию.
//	@Tags			bank
//	@Produce		json
//	@Param			status			query		string	false	"Статус"
//	@Param			product			query		string	false	"Код продукта"
//	@Param			risk			query		string	false	"LOW, MEDIUM, HIGH"
//	@Param			decisionMode	query		string	false	"AUTO, MANUAL"
//	@Param			sort			query		string	false	"Сортировка"	default(-submittedAt)
//	@Param			page			query		int		false	"Страница"		default(1)
//	@Param			size			query		int		false	"Размер страницы"	default(20)
//	@Success		200				{object}	queuePage
//	@Failure		400				{object}	httpx.ErrorBody
//	@Failure		401				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Router			/bank/applications [get]
func (m *Module) bankQueue(c *gin.Context) {
	f, err := parseFilter(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	page, size := httpx.ParsePage(c)
	rows, total, err := m.store.ListQueue(c.Request.Context(), f, page, size)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	items := make([]queueItemDTO, len(rows))
	for i, r := range rows {
		items[i] = toQueueItem(r)
	}
	httpx.JSON(c, http.StatusOK, httpx.NewPage(items, page, size, total))
}

type assignResponse struct {
	Assignee assigneeDTO `json:"assignee"`
}

// bankAssign godoc
//
//	@Summary	Взять заявку в работу
//	@Tags		bank
//	@Produce	json
//	@Param		id	path		string	true	"ID заявки"
//	@Success	200	{object}	assignResponse
//	@Failure	401	{object}	httpx.ErrorBody
//	@Failure	403	{object}	httpx.ErrorBody
//	@Failure	404	{object}	httpx.ErrorBody
//	@Failure	409	{object}	httpx.ErrorBody
//	@Router		/bank/applications/{id}/assign [post]
func (m *Module) bankAssign(c *gin.Context) {
	u := auth.UserFrom(c)
	name, err := m.assign(c.Request.Context(), c.Param("id"), u.ID, u.FullName)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, assignResponse{Assignee: assigneeDTO{ID: u.ID, FullName: name}})
}
