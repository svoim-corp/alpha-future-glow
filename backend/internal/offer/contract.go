package offer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
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
	if k.o.DisbursedAt != nil {
		now = *k.o.DisbursedAt
	}
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

const actionDisburse = "DISBURSE"

// disbursement — итог выдачи.
type disbursement struct {
	loanID        string
	at            time.Time
	accountMasked string
}

// disburse — имитация выдачи: кредит и график в Servicing, заявка — «Выдано». Кредит заводится идемпотентно,
// поэтому сбой после него безопасно повторить. Отметка о выдаче пишется до перехода и откатывается при его сбое.
func (m *Module) disburse(ctx context.Context, appID, companyID string, actor shared.Actor) (*disbursement, error) {
	k, err := m.contractOf(ctx, appID, companyID)
	if err != nil {
		return nil, err
	}
	o := k.o
	switch {
	case k.app.Status != shared.StatusSigning:
		return nil, httpx.InvalidTransition("Кредит выдаётся в статусе «Подписание»")
	case o.SignedAt == nil:
		return nil, httpx.InvalidTransition("Сначала подпишите договор")
	case o.DisbursedAt != nil:
		return nil, httpx.InvalidTransition("Кредит уже выдан")
	}
	cv, err := m.ports.Companies.GetByID(ctx, k.app.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("load company %s: %w", k.app.CompanyID, err)
	}
	if cv == nil {
		return nil, fmt.Errorf("company %s of application %s not found", k.app.CompanyID, appID)
	}

	now := m.now().UTC()
	in := shared.LoanOpening{
		AppID: appID, CompanyID: k.app.CompanyID, ApplicantID: k.app.ApplicantID, ContractNumber: *o.ContractNumber,
		ProductCode: o.ProductCode, Amount: o.Amount, Term: o.Term, Rate: o.Rate, IssuedAt: now,
	}
	if o.GraceMonths != nil {
		in.GraceMonths = *o.GraceMonths
	}
	loanID, err := m.ports.Loans.Open(ctx, in)
	if err != nil {
		return nil, err
	}
	ok, err := m.store.Disburse(ctx, appID, loanID, now)
	if err != nil {
		return nil, fmt.Errorf("save disbursement %s: %w", appID, err)
	}
	if !ok {
		return nil, httpx.InvalidTransition("Кредит уже выдан")
	}
	if err := m.ports.Applications.Transition(ctx, appID, shared.StatusIssued, actor, actionDisburse, ""); err != nil {
		if uerr := m.store.Undisburse(ctx, appID); uerr != nil {
			slog.Error("rollback disbursement", "appId", appID, "err", uerr)
		}
		return nil, err
	}
	return &disbursement{loanID: loanID, at: now, accountMasked: maskedAccount(cv.INN)}, nil
}

// maskedAccount — счёт зачисления в виде «•••• 5678». Счетов в демо-данных банка нет, поэтому последние цифры
// берутся из ИНН клиента: так номер стабилен для компании.
func maskedAccount(inn string) string {
	return "•••• " + inn[max(0, len(inn)-4):]
}
