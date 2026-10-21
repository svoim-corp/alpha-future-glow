package shared

import "time"

// ApplicationSubmitted — клиент отправил заявку. Публикует Origination (A), слушает Scoring (B) и запускает конвейер.
type ApplicationSubmitted struct{ AppID string }

func (ApplicationSubmitted) Name() string { return "application.submitted" }

// ApplicationStatusChanged — у заявки сменился статус. Публикует Origination (A) при каждом переходе.
// Слушает Underwriting (B), чтобы вести свою таблицу очереди: JOIN между схемами зон запрещён.
type ApplicationStatusChanged struct {
	AppID    string
	From, To Status
	At       time.Time
}

func (ApplicationStatusChanged) Name() string { return "application.status_changed" }

// DecisionMade — по заявке принято решение (авто или сотрудником). Публикует B, слушает Offer (A).
// Само решение читается через Decisions.LatestDecision.
type DecisionMade struct{ AppID, DecisionID string }

func (DecisionMade) Name() string { return "decision.made" }
