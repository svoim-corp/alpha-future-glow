package underwriting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const dateLayout = "2006-01-02"

type applicantDTO struct {
	FullName       string               `json:"fullName" example:"Петрова Анна Сергеевна"`
	Authority      shared.AuthorityType `json:"authority" example:"MCHD"`
	MCHDNumber     *string              `json:"mchdNumber" example:"a1b2c3d4-5e6f-4a7b-8c9d-0e1f2a3b4c5d"`
	MCHDValidUntil *string              `json:"mchdValidUntil" example:"2027-03-01"`
}

type reviewCompanyDTO struct {
	INN           string             `json:"inn" example:"7712345678"`
	Name          string             `json:"name" example:"ООО «Северная пекарня»"`
	Type          shared.CompanyType `json:"type" example:"LLC"`
	RegDate       string             `json:"regDate" example:"2015-03-12"`
	OKVED         string             `json:"okved" example:"10.71"`
	TaxRegime     string             `json:"taxRegime" example:"USN_INCOME_MINUS_EXPENSES"`
	ActualAddress string             `json:"actualAddress" example:"г. Москва, ул. Складская, 3"`
	ClientSince   string             `json:"clientSince" example:"2016-02-01"`
}

type requestDTO struct {
	ProductCode string `json:"productCode" example:"WORKING"`
	Amount      int64  `json:"amount" example:"1500000"`
	Term        int    `json:"term" example:"24"`
	Goal        string `json:"goal" example:"WORKING_CAPITAL"`
	PurposeText string `json:"purposeText" example:"Закупка муки и упаковки к сезону"`
}

type reviewTurnoverDTO struct {
	Source           shared.CheckSource `json:"source" example:"BANK_ACCOUNTS"`
	AvgMonthlyIncome int64              `json:"avgMonthlyIncome" example:"1840000"`
	PeriodMonths     int                `json:"periodMonths" example:"12"`
}

// reviewGIRBODTO — цифры ГИР БО видит только сотрудник (клиенту — «учтена»).
type reviewGIRBODTO struct {
	Year      int   `json:"year" example:"2025"`
	Revenue   int64 `json:"revenue" example:"21400000"`
	NetProfit int64 `json:"netProfit" example:"2100000"`
}

type guarantorDTO struct {
	PersonID   string `json:"personId" example:"p-1"`
	FullName   string `json:"fullName" example:"Смирнов Олег Викторович"`
	INN        string `json:"inn" example:"771234567890"`
	BKIConsent bool   `json:"bkiConsent" example:"true"`
}

type reviewPrelimDTO struct {
	Limit      int64                `json:"limit" example:"3000000"`
	PrelimRate float64              `json:"prelimRate" example:"22"`
	AmountFits bool                 `json:"amountFits" example:"true"`
	Choice     *shared.PrelimChoice `json:"choice"`
}

type reviewResponse struct {
	ID        string            `json:"id" example:"a-1024"`
	Number    *string           `json:"number" example:"КЗ-2026-001024"`
	Status    shared.Status     `json:"status" example:"ON_DECISION"`
	Applicant applicantDTO      `json:"applicant"`
	Company   reviewCompanyDTO  `json:"company"`
	Request   requestDTO        `json:"request"`
	Turnover  reviewTurnoverDTO `json:"turnover"`
	GIRBO     *reviewGIRBODTO   `json:"girbo"`
	Guarantor *guarantorDTO     `json:"guarantor"`
	Prelim    *reviewPrelimDTO  `json:"prelim"`
}

func dateOrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dateLayout)
}

func toReviewResponse(d *dossier) reviewResponse {
	app, co := d.app, d.company
	resp := reviewResponse{
		ID: app.ID, Number: app.Number, Status: app.Status,
		Applicant: applicantDTO{Authority: app.Authority.Type, MCHDNumber: app.Authority.MCHDNumber},
		Company: reviewCompanyDTO{
			INN: co.INN, Name: co.Name, Type: co.Type, RegDate: dateOrEmpty(co.RegDate), OKVED: co.OKVED.Code,
			TaxRegime: co.TaxRegime, ActualAddress: co.ActualAddress, ClientSince: dateOrEmpty(co.ClientSince),
		},
		Request:  requestDTO{ProductCode: app.ProductCode, Amount: app.Amount, Term: app.Term, Goal: app.Goal, PurposeText: app.PurposeText},
		Turnover: reviewTurnoverDTO{Source: shared.SourceBankAccounts, AvgMonthlyIncome: co.Turnover.AvgMonthlyIncome, PeriodMonths: co.Turnover.PeriodMonths},
	}
	// Имя есть в данных банка только у руководителя. Имя представителя по МЧД появится, когда Origination
	// отдаст его в ApplicationView (вопрос к разработчику A).
	if app.Authority.Type == shared.AuthorityHead {
		resp.Applicant.FullName = co.Head.FullName
	}
	if until := app.Authority.ValidUntil; until != nil {
		s := until.Format(dateLayout)
		resp.Applicant.MCHDValidUntil = &s
	}
	if d.checks != nil && d.checks.Facts.GIRBO != nil {
		g := d.checks.Facts.GIRBO
		resp.GIRBO = &reviewGIRBODTO{Year: g.Year, Revenue: g.Revenue, NetProfit: g.NetProfit}
	}
	if d.guarantor != nil {
		resp.Guarantor = &guarantorDTO{PersonID: d.guarantor.PersonID, FullName: d.guarantor.FullName, INN: d.guarantor.INN, BKIConsent: d.guarantorBKI != nil}
	}
	if p := d.prelim; p != nil && p.Outcome == shared.PrelimLimit {
		resp.Prelim = &reviewPrelimDTO{Limit: p.Limit, PrelimRate: p.PrelimRate, AmountFits: p.AmountFits, Choice: p.Choice}
	}
	return resp
}

// bankApplication godoc
//
//	@Summary		Анкета заявки для сотрудника
//	@Description	Отличие от клиентской версии — цифры ГИР БО, предварительный результат и заявитель.
//	@Tags			bank
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	reviewResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		502	{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id} [get]
func (m *Module) bankApplication(c *gin.Context) {
	d, err := m.dossier(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toReviewResponse(d))
}

type completenessResponse struct {
	Complete bool    `json:"complete" example:"true"`
	Issues   []issue `json:"issues"`
}

// bankCompleteness godoc
//
//	@Summary		Полнота и корректность заявки
//	@Description	severity: INFO, WARN, BLOCKER. complete = нет блокеров.
//	@Tags			bank
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	completenessResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/completeness [get]
func (m *Module) bankCompleteness(c *gin.Context) {
	d, err := m.dossier(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	complete, issues := completeness(d)
	if issues == nil {
		issues = []issue{}
	}
	httpx.JSON(c, http.StatusOK, completenessResponse{Complete: complete, Issues: issues})
}

type checkDTO struct {
	Source      shared.CheckSource `json:"source" example:"BKI"`
	Status      shared.CheckStatus `json:"status" example:"OK"`
	RequestedAt time.Time          `json:"requestedAt" example:"2026-10-20T10:26:00Z"`
	Summary     string             `json:"summary" example:"2 кредита, просрочек нет"`
	Details     map[string]any     `json:"details,omitempty"`
}

func toChecks(rep *shared.CheckReport) []checkDTO {
	out := make([]checkDTO, 0, len(rep.Results))
	for _, r := range rep.Results {
		out = append(out, checkDTO{Source: r.Source, Status: r.Status, RequestedAt: r.RequestedAt, Summary: r.Summary, Details: r.Details})
	}
	return out
}

// bankChecks godoc
//
//	@Summary	Результаты внешних проверок
//	@Tags		bank
//	@Produce	json
//	@Param		id	path	string	true	"ID заявки"
//	@Success	200	{array}	checkDTO
//	@Failure	401	{object}	httpx.ErrorBody
//	@Failure	403	{object}	httpx.ErrorBody
//	@Failure	404	{object}	httpx.ErrorBody
//	@Router		/bank/applications/{id}/checks [get]
func (m *Module) bankChecks(c *gin.Context) {
	ctx := c.Request.Context()
	app, err := m.load(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rep, err := m.ports.Checks.Latest(ctx, app.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toChecks(rep))
}

type refreshRequest struct {
	Sources []shared.CheckSource `json:"sources" binding:"omitempty,dive,oneof=BANK BANK_ACCOUNTS EGRUL TRANSPARENT_BUSINESS BKI FNS GIRBO COURTS" example:"BKI,COURTS"`
}

// bankRefreshChecks godoc
//
//	@Summary		Перезапросить проверки
//	@Description	Тело необязательно: без sources перезапрашиваются все источники.
//	@Tags			bank
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string			true	"ID заявки"
//	@Param			body	body	refreshRequest	false	"Источники"
//	@Success		200		{array}	checkDTO
//	@Failure		400		{object}	httpx.ErrorBody
//	@Failure		401		{object}	httpx.ErrorBody
//	@Failure		403		{object}	httpx.ErrorBody
//	@Failure		404		{object}	httpx.ErrorBody
//	@Router			/bank/applications/{id}/checks/refresh [post]
func (m *Module) bankRefreshChecks(c *gin.Context) {
	var req refreshRequest
	if c.Request.ContentLength != 0 {
		if err := httpx.Bind(c, &req); err != nil {
			httpx.Fail(c, err)
			return
		}
	}
	rep, err := m.refreshChecks(c.Request.Context(), c.Param("id"), req.Sources)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, toChecks(rep))
}
