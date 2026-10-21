package shared

import (
	"context"
	"time"
)

// Applications — заявка зоны Origination. Реализует: разработчик A. Используют: Scoring, Underwriting (B), Offer (A).
// Статус заявки меняется только через Transition.
type Applications interface {
	// Application — заявка по id; (nil, nil) — заявки нет.
	Application(ctx context.Context, id string) (*ApplicationView, error)
	// Transition переводит заявку в статус to, пишет историю и публикует ApplicationStatusChanged.
	// Недопустимый переход — httpx.InvalidTransition.
	Transition(ctx context.Context, id string, to Status, actor Actor, action, comment string) error
	// SetAmount меняет сумму черновика по выбору клиента на этапе 1 (prelim/choice) и возвращает новую версию.
	SetAmount(ctx context.Context, id string, amount int64, actor Actor) (version int, err error)
	// BKIConsent — согласие на запрос в БКИ; (nil, nil) — согласия нет.
	BKIConsent(ctx context.Context, appID string, subject ConsentSubject) (*ConsentView, error)
}

type ApplicationView struct {
	ID                string
	Number            *string // присваивается при отправке
	CompanyID         string
	ApplicantID       string
	ProductCode       string
	Goal              string
	PurposeText       string
	Amount            int64
	Term              int
	GraceMonths       *int
	OverdraftLimit    *int64
	GuarantorPersonID *string
	Authority         AuthorityView // снимок при создании черновика
	Status            Status
	Stage             int
	Version           int
	CreatedAt         time.Time
	SubmittedAt       *time.Time
}

type ConsentSubject string

const (
	ConsentApplicant ConsentSubject = "APPLICANT"
	ConsentGuarantor ConsentSubject = "GUARANTOR"
)

type ConsentView struct {
	ID      string // передаётся в мок БКИ как consentId
	GivenBy string
	GivenAt time.Time
}
