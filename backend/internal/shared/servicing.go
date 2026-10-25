package shared

import (
	"context"
	"time"
)

// Loans — выданные кредиты зоны Servicing. Реализует: разработчик A. Использует: Offer (A) при выдаче.
type Loans interface {
	// Open заводит кредит с графиком платежей и возвращает его id. Повтор по той же заявке возвращает уже
	// заведённый кредит: выдачу можно безопасно повторить.
	Open(ctx context.Context, in LoanOpening) (loanID string, err error)
}

// LoanOpening — условия подписанного договора на момент выдачи.
type LoanOpening struct {
	AppID          string
	CompanyID      string
	ApplicantID    string // получатель напоминаний о платежах
	ContractNumber string
	ProductCode    string
	Amount         int64
	Term           int
	Rate           float64
	GraceMonths    int // 0 — без отсрочки
	IssuedAt       time.Time
}

// LoanOwners — принадлежность кредитов. Реализует: Servicing (A). Использует: Dashboard (B) при привязке виджета.
type LoanOwners interface {
	// LoanCompany — компания кредита, в том числе закрытого; "" — кредита нет.
	LoanCompany(ctx context.Context, id string) (string, error)
}
