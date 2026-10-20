// Package shared — общие типы и интерфейсы между зонами.
// Зоны не импортируют друг друга, а договариваются через этот пакет.
package shared

type Role string

const (
	RoleClient      Role = "CLIENT"
	RoleBankOfficer Role = "BANK_OFFICER"
)

// Status — статус заявки (API_SPEC.md, раздел 2).
type Status string

const (
	StatusDraft         Status = "DRAFT"
	StatusSubmitted     Status = "SUBMITTED"
	StatusDataCheck     Status = "DATA_CHECK"
	StatusInfoRequested Status = "INFO_REQUESTED"
	StatusScoring       Status = "SCORING"
	StatusOnDecision    Status = "ON_DECISION"
	StatusApproved      Status = "APPROVED"
	StatusCounterOffer  Status = "COUNTER_OFFER"
	StatusRejected      Status = "REJECTED"
	StatusSigning       Status = "SIGNING"
	StatusIssued        Status = "ISSUED"
	StatusClosed        Status = "CLOSED"
)

var statusLabels = map[Status]string{
	StatusDraft:         "Черновик",
	StatusSubmitted:     "Подана",
	StatusDataCheck:     "Проверка данных",
	StatusInfoRequested: "Нужна доп. информация",
	StatusScoring:       "Обогащение и скоринг",
	StatusOnDecision:    "На решении",
	StatusApproved:      "Одобрено",
	StatusCounterOffer:  "Иные условия",
	StatusRejected:      "Отказ",
	StatusSigning:       "Подписание",
	StatusIssued:        "Выдано",
	StatusClosed:        "Закрыта",
}

// Label — русская подпись статуса.
func (s Status) Label() string {
	if l, ok := statusLabels[s]; ok {
		return l
	}
	return string(s)
}

type ActorType string

const (
	ActorClient  ActorType = "CLIENT"
	ActorOfficer ActorType = "OFFICER"
	ActorSystem  ActorType = "SYSTEM"
)

// Actor — кто выполнил действие (для истории и аудита).
type Actor struct {
	Type   ActorType
	UserID string
	Name   string
}

type ProductRef struct {
	Code string
	Name string
}
