package shared

import (
	"context"
	"time"
)

// Prelims — предварительное решение этапа 1. Реализует: разработчик B. Использует: Origination (A) для GET /applications/{id}.
type Prelims interface {
	// Prelim — последний результат этапа 1; (nil, nil) — этап 1 не пройден.
	Prelim(ctx context.Context, appID string) (*PrelimView, error)
}

type PrelimOutcome string

const (
	PrelimLimit PrelimOutcome = "LIMIT"
	PrelimStop  PrelimOutcome = "STOP"
)

type PrelimChoice string

const (
	ChoiceReduceAmount PrelimChoice = "REDUCE_AMOUNT"
	ChoiceKeepAmount   PrelimChoice = "KEEP_AMOUNT"
)

type PrelimView struct {
	Outcome    PrelimOutcome
	Limit      int64
	PrelimRate float64
	AmountFits bool
	StopReason *string
	Choice     *PrelimChoice
	// Параметры, по которым посчитан результат. Origination сравнивает их с текущими и сам выставляет prelimInvalidated.
	ProductCode string
	Amount      int64
	Term        int
}

// Decisions — решение по заявке (авто или сотрудником). Реализует: разработчик B. Использует: Offer (A) по событию DecisionMade.
type Decisions interface {
	// LatestDecision — действующее решение; (nil, nil) — решения ещё нет.
	LatestDecision(ctx context.Context, appID string) (*DecisionView, error)
}

type DecisionMode string

const (
	DecisionAuto   DecisionMode = "AUTO"
	DecisionManual DecisionMode = "MANUAL"
)

type DecisionView struct {
	ID           string
	Result       Status // StatusApproved, StatusCounterOffer или StatusRejected
	Mode         DecisionMode
	ProductCode  string
	Amount       int64
	Term         int
	Rate         float64
	Reason       string
	Alternatives []Alternative // только для отказа
	DecidedBy    *string       // id сотрудника; nil — авторешение
	DecidedAt    time.Time
}

type Alternative struct {
	ProductCode string
	MaxAmount   int64
}
