package servicing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const (
	remindBefore   = 3 // за сколько дней до платежа напоминать
	remindInterval = time.Hour
	humanDate      = "02.01.2006"
	calendarLink   = "/credits/calendar"
)

// Run — фоновые задачи зоны, один воркер на процесс: напоминания о платежах (сразу и затем раз в час) и сверка
// незавершённых операций с банком (сразу и затем каждые reconcileInterval). Возвращается после отмены ctx.
// Сбой одного прохода логируется, следующий проход его повторит.
func (m *Module) Run(ctx context.Context) {
	remind, reconcile := time.NewTicker(remindInterval), time.NewTicker(reconcileInterval)
	defer remind.Stop()
	defer reconcile.Stop()
	m.remindPass(ctx)
	m.reconcilePass(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-remind.C:
			m.remindPass(ctx)
		case <-reconcile.C:
			m.reconcilePass(ctx)
		}
	}
}

func (m *Module) remindPass(ctx context.Context) {
	if err := m.remindDue(ctx); err != nil && ctx.Err() == nil {
		slog.Error("payment reminders", "err", err)
	}
}

func (m *Module) reconcilePass(ctx context.Context) {
	if err := m.reconcileOpen(ctx); err != nil && ctx.Err() == nil {
		slog.Error("reconcile operations", "err", err)
	}
}

// remindDue отправляет PAYMENT_DUE заявителю по каждому неоплаченному платежу в ближайшие remindBefore дней
// (включая сегодня), один раз на платёж. Платёж сначала забирается отметкой, затем уведомление; если уведомить
// не удалось, отметка снимается, и следующий проход повторит.
func (m *Module) remindDue(ctx context.Context) error {
	today := m.today()
	due, err := m.store.DueForReminder(ctx, today, today.AddDate(0, 0, remindBefore))
	if err != nil {
		return fmt.Errorf("load due payments: %w", err)
	}
	for _, p := range due {
		claimed, err := m.store.ClaimReminder(ctx, p.LoanID, p.N, m.now().UTC())
		if err != nil {
			return fmt.Errorf("claim reminder %s/%d: %w", p.LoanID, p.N, err)
		}
		if !claimed {
			continue
		}
		n := shared.Notification{
			Type:  shared.NotifyPaymentDue,
			Title: "Платёж по кредиту " + p.DueDate.Format(humanDate),
			Text:  finance.Rub(p.remaining()) + " по договору " + p.ContractNumber + " («" + p.ProductName + "»)",
			Link:  calendarLink,
		}
		if err := m.ports.Notifier.Notify(ctx, p.ApplicantID, n); err != nil {
			if rerr := m.store.UnclaimReminder(ctx, p.LoanID, p.N); rerr != nil {
				slog.Error("unclaim reminder", "loanId", p.LoanID, "n", p.N, "err", rerr)
			}
			return fmt.Errorf("notify payment due %s/%d: %w", p.LoanID, p.N, err)
		}
	}
	return nil
}
