package scoring

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Prelims отдаёт реализацию shared.Prelims другим зонам.
func (m *Module) Prelims() shared.Prelims { return prelimsStub{} }

// Decisions отдаёт реализацию shared.Decisions другим зонам.
func (m *Module) Decisions() shared.Decisions { return decisionsStub{} }

// prelimsStub — заглушка этапа 0: этап 1 ни у кого не пройден. Заменить реализацией (B).
type prelimsStub struct{}

func (prelimsStub) Prelim(context.Context, string) (*shared.PrelimView, error) { return nil, nil }

// decisionsStub — заглушка этапа 0: решений нет. Заменить реализацией (B).
type decisionsStub struct{}

func (decisionsStub) LatestDecision(context.Context, string) (*shared.DecisionView, error) {
	return nil, nil
}
