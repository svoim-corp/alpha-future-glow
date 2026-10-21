package shared

import "context"

// Notifier — уведомления в кабинете (API_SPEC.md, 4.2). Реализует: разработчик A. Используют: Underwriting, Planner (B), Offer, Servicing (A).
type Notifier interface {
	Notify(ctx context.Context, userID string, n Notification) error
}

type NotificationType string

const (
	NotifyInfoRequested NotificationType = "INFO_REQUESTED"
	NotifyDecision      NotificationType = "DECISION"
	NotifyPaymentDue    NotificationType = "PAYMENT_DUE"
	NotifyTicketReply   NotificationType = "TICKET_REPLY"
	NotifyPlanMatch     NotificationType = "PLAN_MATCH" // API_SPEC_ADDITIONAL.md, Планер
)

type Notification struct {
	Type  NotificationType
	Title string
	Text  string
	Link  string
}
