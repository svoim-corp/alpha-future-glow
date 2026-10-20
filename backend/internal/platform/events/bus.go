// Package events — простая синхронная шина доменных событий внутри процесса.
package events

import (
	"context"
	"errors"
	"sync"
)

type Event interface{ Name() string }

type Handler func(ctx context.Context, e Event) error

type Bus struct {
	mu   sync.RWMutex
	subs map[string][]Handler
}

func New() *Bus { return &Bus{subs: map[string][]Handler{}} }

// Subscribe подписывает обработчик на событие с данным именем.
func (b *Bus) Subscribe(name string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[name] = append(b.subs[name], h)
}

// Publish вызывает обработчики синхронно в порядке подписки; ошибки объединяются.
// Публиковать нужно после коммита транзакции, иначе подписчики увидят неактуальные данные.
func (b *Bus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	hs := append([]Handler(nil), b.subs[e.Name()]...)
	b.mu.RUnlock()
	var errs []error
	for _, h := range hs {
		if err := h(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
