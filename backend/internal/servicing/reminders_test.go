package servicing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestRemindDue(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	e.open(t, "a-1", "c-1", time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC))  // платёж 23.10 — через 2 дня
	e.open(t, "a-2", "c-1", time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC))  // 25.10 — через 4 дня, рано
	e.open(t, "a-3", "c-1", time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC))  // 20.10 — вчера, уже просрочен
	e.open(t, "a-4", "c-1", time.Date(2026, 9, 21, 15, 0, 0, 0, time.UTC)) // 21.10 — сегодня

	if err := e.m.remindDue(ctx); err != nil {
		t.Fatal(err)
	}
	if len(e.notifier.sent) != 2 || e.notifier.sent[0] != "u-1:PAYMENT_DUE:Платёж по кредиту 21.10.2026" ||
		e.notifier.sent[1] != "u-1:PAYMENT_DUE:Платёж по кредиту 23.10.2026" {
		t.Fatalf("напоминания: %v", e.notifier.sent)
	}
	// Повторный проход не дублирует напоминания.
	if err := e.m.remindDue(ctx); err != nil || len(e.notifier.sent) != 2 {
		t.Fatalf("повтор: %v %v", e.notifier.sent, err)
	}
}

type failingNotifier struct{}

func (failingNotifier) Notify(context.Context, string, shared.Notification) error {
	return errors.New("notifications unavailable")
}

func TestRemindDueRetriesFailed(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	e.open(t, "a-1", "c-1", time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	e.m.ports.Notifier = failingNotifier{}
	if err := e.m.remindDue(ctx); err == nil || !strings.Contains(err.Error(), "notify") {
		t.Fatalf("ожидалась ошибка уведомления: %v", err)
	}
	e.m.ports.Notifier = e.notifier
	if err := e.m.remindDue(ctx); err != nil || len(e.notifier.sent) != 1 {
		t.Fatalf("неотправленное напоминание повторяется: %v %v", e.notifier.sent, err)
	}
}

func TestRunStops(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		e.m.Run(ctx)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run не остановился после отмены контекста")
	}
}
