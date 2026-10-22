package servicing

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	reconcileInterval = 10 * time.Second
	reconcileBatch    = 50 // операций за проход: следующий проход возьмёт остальные
)

// reconcileOpen доводит незавершённые операции по ответу банка: после сбоя или перезапуска они не теряются.
// Сбой одной операции не останавливает проход, следующий проход её повторит.
func (m *Module) reconcileOpen(ctx context.Context) error {
	ops, err := m.store.OpenOperations(ctx, reconcileBatch)
	if err != nil {
		return fmt.Errorf("load open operations: %w", err)
	}
	for _, op := range ops {
		if _, err := m.Reconcile(ctx, op.ID); err != nil && ctx.Err() == nil {
			slog.Warn("reconcile operation", "operationId", op.ID, "err", err)
		}
	}
	return nil
}
