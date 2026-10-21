package offer

import (
	"context"
	"fmt"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Документы к подписанию. Банк формирует их сам по условиям принятого предложения.
const (
	docCredit = "d-1" // кредитный договор
	docSurety = "d-2" // договор поручительства — только у ООО с поручителем
)

var docTitles = map[string]string{docCredit: "Кредитный договор", docSurety: "Договор поручительства"}

// contract — принятое предложение, заявка и состав документов.
type contract struct {
	app  *shared.ApplicationView
	o    *offerRow
	docs []string
}

// contractOf — договор по заявке своей компании. Он есть только после принятия предложения.
func (m *Module) contractOf(ctx context.Context, appID, companyID string) (*contract, error) {
	app, o, err := m.offerOf(ctx, appID, companyID)
	if err != nil {
		return nil, err
	}
	if o.AcceptedAt == nil {
		return nil, httpx.NotFound("Договор появится после принятия предложения")
	}
	k := &contract{app: app, o: o, docs: []string{docCredit}}
	if app.GuarantorPersonID != nil {
		cv, err := m.ports.Companies.GetByID(ctx, app.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("load company %s: %w", app.CompanyID, err)
		}
		if cv != nil && cv.Type == shared.CompanyLLC {
			k.docs = append(k.docs, docSurety)
		}
	}
	return k, nil
}

// firstPaymentDate — через месяц после выдачи. До выдачи — через месяц от сегодня: клиент видит ориентир.
func (k *contract) firstPaymentDate(now time.Time) time.Time {
	return finance.PaymentDate(now, 1)
}
