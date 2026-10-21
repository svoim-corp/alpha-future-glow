package origination

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var allStatuses = []shared.Status{
	shared.StatusDraft, shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusInfoRequested, shared.StatusScoring,
	shared.StatusOnDecision, shared.StatusApproved, shared.StatusCounterOffer, shared.StatusRejected, shared.StatusSigning,
	shared.StatusIssued, shared.StatusClosed,
}

// TestCanTransition проверяет все пары статусов: разрешены ровно переходы из списка, остальные — нет.
func TestCanTransition(t *testing.T) {
	allowed := []struct{ from, to shared.Status }{
		{shared.StatusDraft, shared.StatusSubmitted},          // клиент отправил
		{shared.StatusDraft, shared.StatusClosed},             // клиент удалил черновик
		{shared.StatusSubmitted, shared.StatusDataCheck},      // конвейер: проверки собраны
		{shared.StatusDataCheck, shared.StatusScoring},        // конвейер: скоринг
		{shared.StatusDataCheck, shared.StatusInfoRequested},  // сотрудник запросил пояснения
		{shared.StatusInfoRequested, shared.StatusDataCheck},  // клиент ответил
		{shared.StatusScoring, shared.StatusOnDecision},       // конвейер: балл посчитан
		{shared.StatusOnDecision, shared.StatusApproved},      // решение
		{shared.StatusOnDecision, shared.StatusCounterOffer},  // решение: иные условия
		{shared.StatusOnDecision, shared.StatusRejected},      // решение: отказ
		{shared.StatusOnDecision, shared.StatusInfoRequested}, // сотрудник запросил пояснения на решении
		{shared.StatusApproved, shared.StatusSigning},         // клиент принял оффер
		{shared.StatusApproved, shared.StatusClosed},          // клиент отказался от оффера
		{shared.StatusCounterOffer, shared.StatusSigning},     // клиент принял иные условия
		{shared.StatusCounterOffer, shared.StatusClosed},      // клиент отказался от иных условий
		{shared.StatusSigning, shared.StatusIssued},           // выдача
	}
	for _, from := range allStatuses {
		for _, to := range allStatuses {
			want := slices.ContainsFunc(allowed, func(p struct{ from, to shared.Status }) bool { return p.from == from && p.to == to })
			if got := canTransition(from, to); got != want {
				t.Errorf("%s → %s: %v, want %v", from, to, got, want)
			}
		}
	}
}

// recorder запоминает опубликованные ApplicationStatusChanged.
type recorder struct {
	mu  sync.Mutex
	got []shared.ApplicationStatusChanged
}

func (r *recorder) listen(bus *events.Bus) {
	bus.Subscribe(shared.ApplicationStatusChanged{}.Name(), func(_ context.Context, e events.Event) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.got = append(r.got, e.(shared.ApplicationStatusChanged))
		return nil
	})
}

func TestTransition(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	rec := &recorder{}
	rec.listen(e.bus)
	id := e.draft(t, "smirnov").ID
	officer := shared.Actor{Type: shared.ActorOfficer, UserID: "u-9", Name: "Иванова Елена Петровна"}

	a, err := e.m.transition(ctx, id, shared.StatusClosed, officer, "TEST", "комментарий", nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != string(shared.StatusClosed) || a.Version != 2 {
		t.Fatalf("после перехода: %+v", a)
	}
	evs, err := e.m.store.Events(ctx, id)
	if err != nil || len(evs) != 1 {
		t.Fatalf("история: %+v %v", evs, err)
	}
	if ev := evs[0]; ev.From != "DRAFT" || ev.To != "CLOSED" || ev.Action != "TEST" || ev.ActorType != "OFFICER" ||
		*ev.ActorID != "u-9" || *ev.Comment != "комментарий" || !ev.At.Equal(today) {
		t.Fatalf("событие истории: %+v", ev)
	}
	if len(rec.got) != 1 || rec.got[0].AppID != id || rec.got[0].From != shared.StatusDraft || rec.got[0].To != shared.StatusClosed {
		t.Fatalf("ApplicationStatusChanged: %+v", rec.got)
	}

	// Недопустимый переход — 409, ничего не меняется и не публикуется.
	_, err = e.m.transition(ctx, id, shared.StatusSubmitted, officer, "TEST", "", nil)
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != http.StatusConflict || he.Code != httpx.CodeInvalidTransition || !strings.Contains(he.Message, "«Закрыта»") {
		t.Fatalf("недопустимый переход: %v", err)
	}
	if evs, _ := e.m.store.Events(ctx, id); len(evs) != 1 || len(rec.got) != 1 {
		t.Fatalf("недопустимый переход не пишет историю: %d событий, %d публикаций", len(evs), len(rec.got))
	}
	if _, err := e.m.transition(ctx, "a-0", shared.StatusClosed, officer, "TEST", "", nil); code(err) != httpx.CodeNotFound {
		t.Fatalf("нет заявки: %v", err)
	}
}

// TestTransitionConcurrent — из двух одновременных переходов из одного статуса проходит ровно один.
func TestTransitionConcurrent(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	var wg sync.WaitGroup
	errs := make([]error, 5)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = e.m.transition(context.Background(), id, shared.StatusClosed, shared.Actor{Type: shared.ActorSystem}, "TEST", "", nil)
		}()
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
		} else if code(err) != httpx.CodeInvalidTransition {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
	}
	if evs, _ := e.m.store.Events(context.Background(), id); ok != 1 || len(evs) != 1 {
		t.Fatalf("успешных переходов %d, событий %d", ok, len(evs))
	}
}
