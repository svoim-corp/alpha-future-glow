package company

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type okvedDTO struct {
	Code string `json:"code" example:"10.71"`
	Name string `json:"name" example:"Производство хлеба"`
}

type personRefDTO struct {
	PersonID string `json:"personId" example:"p-1"`
	FullName string `json:"fullName" example:"Смирнов Олег Викторович"`
}

type founderDTO struct {
	PersonID string  `json:"personId" example:"p-1"`
	FullName string  `json:"fullName" example:"Смирнов Олег Викторович"`
	Share    float64 `json:"share" example:"100"`
}

type candidateDTO struct {
	PersonID string              `json:"personId" example:"p-1"`
	FullName string              `json:"fullName" example:"Смирнов Олег Викторович"`
	Roles    []shared.PersonRole `json:"roles" example:"HEAD,FOUNDER"`
}

type turnoverDTO struct {
	Source           shared.CheckSource `json:"source" example:"BANK_ACCOUNTS"`
	AvgMonthlyIncome int64              `json:"avgMonthlyIncome" example:"1840000"`
	PeriodMonths     int                `json:"periodMonths" example:"12"`
}

type girboDTO struct {
	Accounted bool `json:"accounted" example:"true"`
}

type profileResponse struct {
	Source              []shared.CheckSource `json:"source" example:"BANK,EGRUL"`
	RegAddress          string               `json:"regAddress" example:"г. Москва, ул. Пекарская, 7"`
	ActualAddress       string               `json:"actualAddress" example:"г. Москва, ул. Складская, 3"`
	OKVED               okvedDTO             `json:"okved"`
	Head                *personRefDTO        `json:"head"`
	Founders            []founderDTO         `json:"founders"`
	GuarantorCandidates []candidateDTO       `json:"guarantorCandidates"`
	TaxRegime           string               `json:"taxRegime" example:"USN_INCOME_MINUS_EXPENSES"`
	Turnover            turnoverDTO          `json:"turnover"`
	GIRBO               *girboDTO            `json:"girbo"` // null у ИП: ГИР БО ведётся только по юрлицам
}

func toProfileResponse(p *shared.CompanyProfile) profileResponse {
	resp := profileResponse{
		Source:              p.Sources,
		RegAddress:          p.RegAddress,
		ActualAddress:       p.ActualAddress,
		OKVED:               okvedDTO{Code: p.OKVED.Code, Name: p.OKVED.Name},
		Founders:            make([]founderDTO, len(p.Founders)),
		GuarantorCandidates: make([]candidateDTO, len(p.GuarantorCandidates)),
		TaxRegime:           p.TaxRegime,
		Turnover:            turnoverDTO{Source: shared.SourceBankAccounts, AvgMonthlyIncome: p.Turnover.AvgMonthlyIncome, PeriodMonths: p.Turnover.PeriodMonths},
	}
	if p.Head.PersonID != "" {
		resp.Head = &personRefDTO{PersonID: p.Head.PersonID, FullName: p.Head.FullName}
	}
	for i, f := range p.Founders {
		resp.Founders[i] = founderDTO{PersonID: f.PersonID, FullName: f.FullName, Share: f.Share}
	}
	for i, g := range p.GuarantorCandidates {
		resp.GuarantorCandidates[i] = candidateDTO{PersonID: g.PersonID, FullName: g.FullName, Roles: g.Roles}
	}
	if p.Type == shared.CompanyLLC {
		resp.GIRBO = &girboDTO{Accounted: p.GIRBOAccounted}
	}
	return resp
}

// companyProfile godoc
//
//	@Summary		Данные о компании для заявки
//	@Description	Профиль из клиентской базы банка, сверенный с реестрами. Только для чтения.
//	@Tags			applications
//	@Produce		json
//	@Param			id	path		string	true	"ID заявки"
//	@Success		200	{object}	profileResponse
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Failure		404	{object}	httpx.ErrorBody
//	@Failure		502	{object}	httpx.ErrorBody
//	@Router			/applications/{id}/company-profile [get]
func (m *Module) companyProfile(c *gin.Context) {
	ctx := c.Request.Context()
	app, err := m.ports.Applications.Application(ctx, c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if app == nil {
		httpx.Fail(c, httpx.NotFound("Заявка не найдена"))
		return
	}
	if u := auth.UserFrom(c); u.CompanyID == nil || *u.CompanyID != app.CompanyID {
		httpx.Fail(c, httpx.Forbidden("Заявка принадлежит другой компании"))
		return
	}
	p, err := m.profiles.Profile(ctx, app.CompanyID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if p == nil {
		httpx.Fail(c, httpx.NotFound("Компания не найдена"))
		return
	}
	httpx.JSON(c, http.StatusOK, toProfileResponse(p))
}
