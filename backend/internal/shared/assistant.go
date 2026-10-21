package shared

import "context"

// AssistantCatalog — данные и штатный расчёт каталога для поддержки.
type AssistantCatalog interface {
	Search(ctx context.Context, goal string, amount int64, term int) ([]ProductView, error)
	Calculate(ctx context.Context, productCode string, amount int64, term int) (*PaymentEstimate, error)
}

// PaymentEstimate — предварительный диапазон платежа, не оффер.
type PaymentEstimate struct {
	Amount  int64   `json:"amount"`
	Term    int     `json:"term"`
	Min     int64   `json:"min"`
	Max     int64   `json:"max"`
	RateMin float64 `json:"rateMin"`
	RateMax float64 `json:"rateMax"`
}

// AssistantApplications — чтение своих заявок и идемпотентное создание черновика.
type AssistantApplications interface {
	List(ctx context.Context, companyID string) ([]ApplicationView, error)
	CreateDraft(ctx context.Context, who DraftApplicant, in AssistantDraft, commandID string) (*ApplicationView, error)
}

// DraftApplicant — доверенный контекст сессии, модель его не задаёт.
type DraftApplicant struct{ UserID, CompanyID, PersonINN, Name string }

// AssistantDraft — параметры черновика с повторной доменной валидацией.
type AssistantDraft struct {
	ProductCode string `json:"productCode" binding:"required,max=50"`
	Amount      int64  `json:"amount" binding:"required,gt=0,lte=1000000000000"`
	Term        int    `json:"term" binding:"required,gt=0,lte=1200"`
	Goal        string `json:"goal" binding:"required,max=100"`
}

// AssistantPayments — ближайшие платежи только компании из сессии.
type AssistantPayments interface {
	Upcoming(ctx context.Context, companyID string) ([]UpcomingPayment, error)
}

// UpcomingPayment — карточка ближайшего платежа.
type UpcomingPayment struct {
	LoanID      string `json:"loanId"`
	ProductName string `json:"productName"`
	Date        string `json:"date"`
	Amount      int64  `json:"amount"`
}
