package offer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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

// sign — имитация подписи без SMS: подтверждение клиента и идентификатор подписи. Статус заявки не меняется —
// подписание заканчивается выдачей.
func (m *Module) sign(ctx context.Context, appID, companyID string) (*offerRow, error) {
	k, err := m.contractOf(ctx, appID, companyID)
	if err != nil {
		return nil, err
	}
	switch {
	case k.app.Status != shared.StatusSigning:
		return nil, httpx.InvalidTransition("Договор подписывается в статусе «Подписание»")
	case k.o.SignedAt != nil:
		return nil, httpx.InvalidTransition("Договор уже подписан")
	}
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("signature id: %w", err)
	}
	sig, now := "SIM-"+strings.ToUpper(hex.EncodeToString(b)), m.now().UTC()
	ok, err := m.store.Sign(ctx, appID, sig, now)
	if err != nil {
		return nil, fmt.Errorf("sign contract %s: %w", appID, err)
	}
	if !ok {
		return nil, httpx.InvalidTransition("Договор уже подписан")
	}
	k.o.SignedAt, k.o.SignatureID = &now, &sig
	return k.o, nil
}
