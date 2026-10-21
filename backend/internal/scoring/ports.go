package scoring

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Prelims отдаёт реализацию shared.Prelims другим зонам.
func (m *Module) Prelims() shared.Prelims { return prelims{store: m.store} }

// Decisions отдаёт реализацию shared.Decisions другим зонам.
func (m *Module) Decisions() shared.Decisions { return decisionsStub{} }

type prelims struct{ store *Store }

func (p prelims) Prelim(ctx context.Context, appID string) (*shared.PrelimView, error) {
	r, err := p.store.LatestPrelim(ctx, appID)
	if err != nil || r == nil {
		return nil, err
	}
	return r.view(), nil
}

// decisionsStub — заглушка этапа 0: решений нет. Заменить реализацией (B).
type decisionsStub struct{}

func (decisionsStub) LatestDecision(context.Context, string) (*shared.DecisionView, error) {
	return nil, nil
}
