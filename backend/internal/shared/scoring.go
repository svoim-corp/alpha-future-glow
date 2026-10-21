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

// Scores — результат полного скоринга. Реализует: разработчик B (Scoring). Использует: Underwriting (B) для очереди и разбора.
type Scores interface {
	// Score — последний результат скоринга; (nil, nil) — скоринга ещё не было.
	Score(ctx context.Context, appID string) (*ScoreView, error)
}

// DecisionMaker — принятие решения по заявке. Реализует: разработчик B (Scoring). Использует: Underwriting (B) для ручного решения.
// Проверяет рамки продукта (httpx.BusinessRule), переводит заявку в статус решения и публикует DecisionMade.
type DecisionMaker interface {
	Decide(ctx context.Context, appID string, in DecisionInput, actor Actor) (*DecisionView, error)
}

type RiskBand string

const (
	RiskLow    RiskBand = "LOW"
	RiskMedium RiskBand = "MEDIUM"
	RiskHigh   RiskBand = "HIGH"
)

type ScoreView struct {
	Score                int
	RiskBand             RiskBand
	AutoDecisionEligible bool
	AutoDecisionReason   string
	DecisionMode         DecisionMode // AUTO — решено автоматически, MANUAL — ждёт сотрудника
	Factors              []ScoreFactor
	Suggested            *SuggestedDecision
	ModelVersion         string
	ScoredAt             time.Time
}

// ScoreFactor — вклад признака в балл (стр. 14 ТЗ: каждому признаку — своё влияние).
type ScoreFactor struct {
	Code   string
	Label  string
	Impact int
}

type SuggestedDecision struct {
	Decision Status // StatusApproved, StatusCounterOffer или StatusRejected
	Amount   int64
	Rate     float64
}

// DecisionInput — решение. Для отказа нужен только Reason; для одобрения и иных условий — все параметры.
type DecisionInput struct {
	Result       Status // StatusApproved, StatusCounterOffer или StatusRejected
	ProductCode  string
	Amount       int64
	Term         int
	Rate         float64
	Reason       string
	Alternatives []Alternative
}
