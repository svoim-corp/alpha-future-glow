package shared

// ApplicationSubmitted — клиент отправил заявку.
type ApplicationSubmitted struct{ AppID string }

func (ApplicationSubmitted) Name() string { return "application.submitted" }

// DecisionMade — по заявке принято решение.
type DecisionMade struct{ AppID, DecisionID string }

func (DecisionMade) Name() string { return "decision.made" }
