package events

import (
	"context"
	"errors"
	"testing"
)

type ev struct{}

func (ev) Name() string { return "test" }

func TestPublish(t *testing.T) {
	b := New()
	var order []int
	e1, e2 := errors.New("e1"), errors.New("e2")
	b.Subscribe("test", func(context.Context, Event) error { order = append(order, 1); return e1 })
	b.Subscribe("test", func(context.Context, Event) error { order = append(order, 2); return nil })
	b.Subscribe("test", func(context.Context, Event) error { order = append(order, 3); return e2 })
	b.Subscribe("other", func(context.Context, Event) error { t.Fatal("чужая подписка"); return nil })

	err := b.Publish(context.Background(), ev{})
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("порядок: %v", order)
	}
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("ошибки не объединены: %v", err)
	}
	if err := New().Publish(context.Background(), ev{}); err != nil {
		t.Fatal(err)
	}
}
