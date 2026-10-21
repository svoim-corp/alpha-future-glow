package origination

import (
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const dateLayout = "2006-01-02"

type paramsDTO struct {
	Amount         int64   `json:"amount" example:"1500000"`
	Term           int     `json:"term" example:"24"`
	Goal           *string `json:"goal" example:"WORKING_CAPITAL"`
	PurposeText    *string `json:"purposeText" example:"Закупка муки и упаковки к сезону"`
	GraceMonths    *int    `json:"graceMonths" example:"6"`
	OverdraftLimit *int64  `json:"overdraftLimit" example:"500000"`
}

type companyDTO struct {
	INN     string             `json:"inn" example:"7712345678"`
	Type    shared.CompanyType `json:"type" example:"LLC"`
	Name    string             `json:"name" example:"ООО «Северная пекарня»"`
	OGRN    string             `json:"ogrn" example:"1157746000000"`
	RegDate *string            `json:"regDate" example:"2015-03-12"`
	State   string             `json:"state" example:"ACTIVE"`
}

type authorityDTO struct {
	Type       shared.AuthorityType `json:"type" example:"MCHD"`
	Verified   bool                 `json:"verified" example:"true"` // снимок проверен по данным банка и реестру МЧД
	ValidUntil *string              `json:"validUntil" example:"2027-03-01"`
}

type consentDTO struct {
	Given bool      `json:"given" example:"true"`
	At    time.Time `json:"at" example:"2026-10-20T10:05:00Z"`
}

type consentsDTO struct {
	BKI *consentDTO `json:"bki,omitempty"`
}

type prelimDTO struct {
	Outcome    shared.PrelimOutcome `json:"outcome" example:"LIMIT"`
	Limit      *int64               `json:"limit" example:"3300000"`
	PrelimRate *float64             `json:"prelimRate" example:"22"`
	AmountFits *bool                `json:"amountFits" example:"true"`
	Choice     *shared.PrelimChoice `json:"choice" example:"KEEP_AMOUNT"`
	StopReason *string              `json:"stopReason" example:"BUSINESS_TOO_YOUNG"`
}

type turnoverDTO struct {
	Source           shared.CheckSource `json:"source" example:"BANK_ACCOUNTS"`
	AvgMonthlyIncome int64              `json:"avgMonthlyIncome" example:"1840000"`
	PeriodMonths     int                `json:"periodMonths" example:"12"`
}

type guarantorDTO struct {
	PersonID   string `json:"personId" example:"p-1"`
	FullName   string `json:"fullName" example:"Смирнов Олег Викторович"`
	BKIConsent bool   `json:"bkiConsent" example:"true"`
}

type stage2DTO struct {
	ActualAddress  *string       `json:"actualAddress" example:"г. Москва, ул. Складская, 3"`
	Turnover       *turnoverDTO  `json:"turnover"`
	Guarantor      *guarantorDTO `json:"guarantor"`
	TruthConfirmed bool          `json:"truthConfirmed" example:"false"`
}

type locksDTO struct {
	ParamsLocked    bool `json:"paramsLocked" example:"false"`
	Stage2Available bool `json:"stage2Available" example:"true"`
}

type applicationDTO struct {
	ID                string        `json:"id" example:"a-1031"`
	Number            *string       `json:"number" example:"КЗ-2026-001031"`
	Status            shared.Status `json:"status" example:"DRAFT"`
	StatusLabel       string        `json:"statusLabel" example:"Черновик"`
	Stage             int           `json:"stage" example:"2"`
	Version           int           `json:"version" example:"7"`
	Product           productRefDTO `json:"product"`
	Params            paramsDTO     `json:"params"`
	Company           *companyDTO   `json:"company"` // null — данные банка сейчас недоступны
	Authority         authorityDTO  `json:"authority"`
	Consents          consentsDTO   `json:"consents"`
	Prelim            *prelimDTO    `json:"prelim"`            // null — этап 1 не пройден или результат устарел
	PrelimInvalidated bool          `json:"prelimInvalidated"` // этап 1 нужно пройти заново: изменились сумма, срок или продукт
	Stage2            stage2DTO     `json:"stage2"`
	Locks             locksDTO      `json:"locks"`
	SavedAt           time.Time     `json:"savedAt" example:"2026-10-20T10:12:00Z"`
}

func dateOrNil(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	s := t.Format(dateLayout)
	return &s
}

func toPrelimDTO(p *shared.PrelimView) *prelimDTO {
	if p == nil {
		return nil
	}
	out := &prelimDTO{Outcome: p.Outcome, Choice: p.Choice, StopReason: p.StopReason}
	if p.Outcome == shared.PrelimLimit {
		limit, rate, fits := p.Limit, p.PrelimRate, p.AmountFits
		out.Limit, out.PrelimRate, out.AmountFits = &limit, &rate, &fits
	}
	return out
}

func toApplicationDTO(d *dossier) applicationDTO {
	a := d.app
	st := shared.Status(a.Status)
	product := productRefDTO{Code: a.ProductCode, Name: a.ProductCode}
	if d.product != nil {
		product.Name = d.product.Name
	}
	resp := applicationDTO{
		ID: a.ID, Number: a.Number, Status: st, StatusLabel: st.Label(), Stage: d.stage.stage, Version: a.Version,
		Product: product,
		Params: paramsDTO{
			Amount: a.Amount, Term: a.Term, Goal: a.Goal, PurposeText: a.PurposeText, GraceMonths: a.GraceMonths, OverdraftLimit: a.OverdraftLimit,
		},
		Authority:         authorityDTO{Type: a.Authority.V.Type, Verified: true},
		Prelim:            toPrelimDTO(d.stage.prelim),
		PrelimInvalidated: d.stage.invalidated,
		Stage2:            stage2DTO{TruthConfirmed: a.TruthConfirmed},
		Locks:             locksDTO{ParamsLocked: st != shared.StatusDraft, Stage2Available: d.stage.stage == stageFull},
		SavedAt:           a.UpdatedAt.UTC(),
	}
	if until := a.Authority.V.ValidUntil; until != nil {
		resp.Authority.ValidUntil = dateOrNil(*until)
	}
	if p := d.profile; p != nil {
		resp.Company = &companyDTO{INN: p.INN, Type: p.Type, Name: p.Name, OGRN: p.OGRN, RegDate: dateOrNil(p.RegDate), State: p.State}
		resp.Stage2.ActualAddress = &p.ActualAddress
		resp.Stage2.Turnover = &turnoverDTO{Source: shared.SourceBankAccounts, AvgMonthlyIncome: p.Turnover.AvgMonthlyIncome, PeriodMonths: p.Turnover.PeriodMonths}
	}
	if c := d.consent; c != nil {
		resp.Consents.BKI = &consentDTO{Given: true, At: c.GivenAt.UTC()}
	}
	if id := a.GuarantorPersonID; id != nil {
		resp.Stage2.Guarantor = &guarantorDTO{PersonID: *id, FullName: d.guarantorName(*id), BKIConsent: d.guarantorAgreed()}
	}
	return resp
}
