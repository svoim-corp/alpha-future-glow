package shared

import "context"

// Origination (разработчик A) → используют другие зоны.
type Applications interface {
	Get(ctx context.Context, id string) (ApplicationView, error)
	Transition(ctx context.Context, id string, to Status, actor Actor, action, comment string) error
	SetAmount(ctx context.Context, id string, amount int64) (version int, err error)
}

type ApplicationView struct {
	ID, CompanyID, ApplicantID, ProductCode, Goal string
	Number                                        *string
	Amount                                        int64
	Term                                          int
	Status                                        Status
	Stage, Version                                int
}

// Company (разработчик B) → Identity (/auth/me), Origination.
type Companies interface {
	GetByID(ctx context.Context, id string) (*CompanyView, error) // nil, nil — нет компании
}

type CompanyView struct {
	ID, INN, Name, Type string
	IsBankClient        bool
}

// Scoring (B) → Origination (GET /applications/{id}).
type Prelims interface {
	Get(ctx context.Context, appID string) (*PrelimView, error)
}

type PrelimView struct {
	Outcome                              string
	LimitNoStatement, LimitWithStatement int64
	PrelimRate                           float64
	Choice                               *string
}

// Notifications (A) → Underwriting, Scoring.
type Notifier interface {
	Notify(ctx context.Context, userID string, n NotificationInput) error
}

type NotificationInput struct {
	Type, Title, Text, Link string
}
