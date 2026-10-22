package servicing

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// loanSnapshot — кредит, график, проводки и операции из одного снимка БД: суммы, статусы и версии в ответе
// относятся к одному состоянию, даже если параллельно проводится оплата.
type loanSnapshot struct {
	loan      *loanRow
	rows      []paymentRow
	posts     []posting
	ops       []opRow // последние операции, новые первыми
	unapplied int64
	at        time.Time
}

// snapshotOps — сколько последних операций попадает в снимок: для PDF истории досрочных погашений хватает.
const snapshotOps = 100

func (m *Module) snapshot(ctx context.Context, loanID string) (*loanSnapshot, error) {
	s := &loanSnapshot{at: m.now().UTC()}
	err := m.store.Snapshot(ctx, func(q sqlx.ExtContext) error {
		var err error
		if s.loan, err = m.store.LoanIn(ctx, q, loanID); err != nil {
			return err
		}
		if s.loan == nil {
			return fmt.Errorf("loan %s not found", loanID)
		}
		if s.rows, err = m.store.LoanPayments(ctx, q, loanID); err != nil {
			return err
		}
		if s.posts, err = m.store.LoanPostings(ctx, q, loanID); err != nil {
			return err
		}
		if s.ops, err = m.store.RecentOperations(ctx, q, loanID, snapshotOps); err != nil {
			return err
		}
		s.unapplied, err = m.store.LoanUnapplied(ctx, q, loanID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}
