package shared

import (
	"context"
	"time"
)

// AssistantPlanner — подготовка и подтверждение изменений планера через его правила.
type AssistantPlanner interface {
	List(ctx context.Context, companyID, state string) ([]AssistantPlan, error)
	Preview(ctx context.Context, companyID string, in AssistantPlanChange) (*AssistantPlan, error)
	Apply(ctx context.Context, who DraftApplicant, in AssistantPlanChange, commandID string) (*AssistantPlan, error)
}

// AssistantPlanChange содержит только явно переданные изменения. Версию задаёт сервер.
type AssistantPlanChange struct {
	ID                string     `json:"id,omitempty" binding:"max=100"`
	Title             *string    `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Goal              *string    `json:"goal,omitempty" binding:"omitempty,min=1,max=64"`
	Note              *string    `json:"note,omitempty" binding:"omitempty,max=2000"`
	BudgetRub         *int64     `json:"budgetRub,omitempty" binding:"omitempty,gt=0,lte=1000000000000"`
	DateMode          *string    `json:"dateMode,omitempty" binding:"omitempty,oneof=DAY PERIOD UNSCHEDULED"`
	StartDate         *string    `json:"startDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
	EndDate           *string    `json:"endDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
	State             *string    `json:"state,omitempty" binding:"omitempty,oneof=ACTIVE PAUSED DONE ARCHIVED"`
	ExpectedUpdatedAt *time.Time `json:"expectedUpdatedAt,omitempty"`
}

// AssistantPlan — проверенная карточка плана без данных другой компании.
type AssistantPlan struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Goal      string    `json:"goal"`
	Note      *string   `json:"note,omitempty"`
	BudgetRub int64     `json:"budgetRub"`
	DateMode  string    `json:"dateMode"`
	StartDate *string   `json:"startDate,omitempty"`
	EndDate   *string   `json:"endDate,omitempty"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AssistantLoans — чтение собственных кредитов без денежных команд.
type AssistantLoans interface {
	List(ctx context.Context, companyID string) ([]AssistantLoan, error)
}

// AssistantLoan — сумма, остаток и условия собственного кредита для карточки чата.
type AssistantLoan struct {
	ID          string  `json:"id"`
	ProductName string  `json:"productName"`
	Amount      int64   `json:"amount"`
	Balance     int64   `json:"balance"`
	Rate        float64 `json:"rate"`
	Term        int     `json:"term"`
	Status      string  `json:"status"`
}

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
	Find(ctx context.Context, companyID string, ids []string) ([]ApplicationView, error)
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
