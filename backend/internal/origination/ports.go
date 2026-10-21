package origination

import (
	"context"
	"errors"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Applications отдаёт реализацию shared.Applications другим зонам.
func (m *Module) Applications() shared.Applications { return applicationsStub{} }

var errNotImplemented = errors.New("origination: not implemented")

// applicationsStub — заглушка этапа 0: заявок нет, команды не выполняются. Заменить реализацией (A).
type applicationsStub struct{}

func (applicationsStub) Application(context.Context, string) (*shared.ApplicationView, error) {
	return nil, nil
}

func (applicationsStub) Transition(context.Context, string, shared.Status, shared.Actor, string, string) error {
	return errNotImplemented
}

func (applicationsStub) SetAmount(context.Context, string, int64, shared.Actor) (int, error) {
	return 0, errNotImplemented
}

func (applicationsStub) BKIConsent(context.Context, string, shared.ConsentSubject) (*shared.ConsentView, error) {
	return nil, nil
}
