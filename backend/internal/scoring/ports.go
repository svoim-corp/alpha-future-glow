package scoring

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Prelims отдаёт реализацию shared.Prelims другим зонам.
func (m *Module) Prelims() shared.Prelims { return prelims{store: m.store} }

// Decisions отдаёт реализацию shared.Decisions другим зонам.
func (m *Module) Decisions() shared.Decisions { return decisions{store: m.store} }

// Scores отдаёт реализацию shared.Scores другим зонам.
func (m *Module) Scores() shared.Scores { return scores{store: m.store} }

// DecisionMaker отдаёт реализацию shared.DecisionMaker другим зонам.
func (m *Module) DecisionMaker() shared.DecisionMaker { return m }

type prelims struct{ store *Store }

func (p prelims) Prelim(ctx context.Context, appID string) (*shared.PrelimView, error) {
	r, err := p.store.LatestPrelim(ctx, appID)
	if err != nil || r == nil {
		return nil, err
	}
	return r.view(), nil
}

type decisions struct{ store *Store }

func (d decisions) LatestDecision(ctx context.Context, appID string) (*shared.DecisionView, error) {
	r, err := d.store.LatestDecision(ctx, appID)
	if err != nil || r == nil {
		return nil, err
	}
	return r.view(), nil
}

type scores struct{ store *Store }

func (s scores) Score(ctx context.Context, appID string) (*shared.ScoreView, error) {
	r, err := s.store.LatestScore(ctx, appID)
	if err != nil || r == nil {
		return nil, err
	}
	return r.view(), nil
}
