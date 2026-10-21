package offer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"slices"
	"strconv"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

//go:embed templates/document.html
var templates embed.FS

var documentTmpl = template.Must(template.ParseFS(templates, "templates/document.html"))

const humanDate = "02.01.2006"

// documentData — поля печатной формы. html/template экранирует их сам: названия компаний и ФИО приходят из данных банка.
type documentData struct {
	Title, Number, Date              string
	Surety                           bool
	Company, INN, Guarantor, Product string
	Amount, Payment, Rate            string
	Term                             int
	GraceMonths                      int
	FirstPayment, Signed             string
}

// renderDocument — печатная форма документа договора. Документа нет в составе договора — 404.
func (m *Module) renderDocument(ctx context.Context, appID, companyID, docID string) ([]byte, error) {
	k, err := m.contractOf(ctx, appID, companyID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(k.docs, docID) {
		return nil, httpx.NotFound("Документ не найден")
	}
	prof, err := m.ports.Companies.Profile(ctx, k.app.CompanyID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, fmt.Errorf("company %s not found", k.app.CompanyID)
	}
	o := k.o
	d := documentData{
		Title: docTitles[docID], Number: *o.ContractNumber, Date: o.AcceptedAt.UTC().Format(humanDate),
		Surety: docID == docSurety, Company: prof.Name, INN: prof.INN, Product: o.ProductCode,
		Amount: finance.Rub(o.Amount), Payment: finance.Rub(o.MonthlyPayment), Rate: strconv.FormatFloat(o.Rate, 'f', -1, 64),
		Term: o.Term, FirstPayment: k.firstPaymentDate(m.now()).Format(humanDate),
	}
	if o.GraceMonths != nil {
		d.GraceMonths = *o.GraceMonths
	}
	if o.SignedAt != nil {
		d.Signed = o.SignedAt.UTC().Format(humanDate) + ", подпись " + *o.SignatureID
	}
	if g := k.app.GuarantorPersonID; g != nil {
		d.Guarantor = *g
		for _, c := range prof.GuarantorCandidates {
			if c.PersonID == *g {
				d.Guarantor = c.FullName
			}
		}
	}
	var buf bytes.Buffer
	if err := documentTmpl.Execute(&buf, d); err != nil {
		return nil, fmt.Errorf("render document %s/%s: %w", appID, docID, err)
	}
	return buf.Bytes(), nil
}
