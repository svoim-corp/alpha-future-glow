package scoring

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var (
	express = shared.ProductView{Code: "EXPRESS", AmountMin: 100_000, AmountMax: 3_000_000, TermMin: 3, TermMax: 24, RateMin: 21, RateMax: 29}
	officer = shared.Actor{Type: shared.ActorOfficer, UserID: "u-9", Name: "Иванова Е. П."}
)

func (f *fakeApps) add(a *shared.ApplicationView) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apps[a.ID] = a
	f.consents[a.ID] = &shared.ConsentView{ID: "cons-" + a.ID}
}

func (f *fakeApps) statusOf(id string) (shared.Status, []shared.Status) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.apps[id].Status, slices.Clone(f.history[id])
}

func (f *fakeApps) setStatus(id string, s shared.Status) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apps[id].Status = s
}

func onDecision(id, company string, amount int64) *shared.ApplicationView {
	a := draftApp(id, company, amount)
	a.Status = shared.StatusOnDecision
	return a
}

// TestDecideRejects — отказы до записи решения: БД не нужна.
func TestDecideRejects(t *testing.T) {
	m, apps, _ := newTestModule(nil)
	apps.add(onDecision("d-1", "c-1", 1_500_000))
	ctx := context.Background()
	approve := func(amount int64, term int, rate float64) shared.DecisionInput {
		return shared.DecisionInput{Result: shared.StatusApproved, ProductCode: "WORKING", Amount: amount, Term: term, Rate: rate, Reason: "ok"}
	}
	tests := []struct {
		name, app string
		in        shared.DecisionInput
		code      string
	}{
		{"нет заявки", "d-404", approve(1_000_000, 24, 20), httpx.CodeNotFound},
		{"заявка не на решении", "a-1", approve(1_000_000, 24, 20), httpx.CodeInvalidTransition},
		{"сумма выше рамок", "d-1", approve(20_000_000, 24, 20), httpx.CodeBusinessRule},
		{"срок вне рамок", "d-1", approve(1_000_000, 60, 20), httpx.CodeBusinessRule},
		{"ставка ниже минимума", "d-1", approve(1_000_000, 24, 10), httpx.CodeBusinessRule},
		{"неизвестный продукт", "d-1", shared.DecisionInput{Result: shared.StatusCounterOffer, ProductCode: "NOPE", Amount: 1, Term: 1, Rate: 1}, httpx.CodeBusinessRule},
		{"отказ без причины", "d-1", shared.DecisionInput{Result: shared.StatusRejected}, httpx.CodeValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := m.Decide(ctx, tt.app, tt.in, officer); errCode(err) != tt.code {
				t.Fatalf("err = %v, want %s", err, tt.code)
			}
		})
	}
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("не дождались: %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestPipeline(t *testing.T) {
	deps := testDeps(t)
	m, apps, checks := newTestModule(deps)
	m.ports.Products = fakeProducts{"WORKING": &working, "EXPRESS": &express}
	ctx := context.Background()

	var decided []string
	decidedCh := make(chan string, 10)
	deps.Bus.Subscribe(shared.DecisionMade{}.Name(), func(_ context.Context, e events.Event) error {
		decidedCh <- e.(shared.DecisionMade).AppID
		return nil
	})

	submitted := func(id, company string, amount int64, c company) {
		a := draftApp(id, company, amount)
		a.Status = shared.StatusSubmitted
		apps.add(a)
		checks.mu.Lock()
		checks.reports["inn-"+company] = c.report()
		checks.mu.Unlock()
	}
	submitted("s-1", "c-s", 1_500_000, smirnov)
	submitted("s-2", "c-m", 3_000_000, morozov)
	submitted("s-3", "c-p", 1_000_000, polyakov)
	submitted("s-4", "c-k", 2_000_000, kozlov)
	if err := m.store.InsertPrelim(ctx, &prelimRow{AppID: "s-3", Outcome: "LIMIT", LimitAmount: 1_200_000, ProductCode: "WORKING",
		Amount: 1_000_000, Term: 24, Checks: db.JSON[[]prelimCheck]{V: []prelimCheck{}}, CreatedAt: today}); err != nil {
		t.Fatal(err)
	}

	runCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); m.Run(runCtx) }()
	defer func() {
		stop()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("воркер не остановился")
		}
	}()

	for _, id := range []string{"s-1", "s-2", "s-3", "s-4"} {
		if err := deps.Bus.Publish(ctx, shared.ApplicationSubmitted{AppID: id}); err != nil {
			t.Fatal(err)
		}
	}
	jobsDone := func() bool {
		var n int
		_ = deps.DB.GetContext(ctx, &n, `SELECT count(*) FROM scoring.jobs WHERE state <> 'DONE'`)
		return n == 0
	}
	eventually(t, "конвейер по четырём заявкам", jobsDone)
	for range 2 {
		select {
		case id := <-decidedCh:
			decided = append(decided, id)
		case <-time.After(time.Second):
			t.Fatalf("DecisionMade: %v", decided)
		}
	}
	slices.Sort(decided)
	if !slices.Equal(decided, []string{"s-1", "s-3"}) {
		t.Fatalf("авторешения: %v", decided)
	}

	pipeline := []shared.Status{shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision}
	tests := []struct {
		id       string
		status   shared.Status
		mode     shared.DecisionMode
		decision *shared.DecisionMode
	}{
		{"s-1", shared.StatusApproved, shared.DecisionAuto, ptr(shared.DecisionAuto)},
		{"s-2", shared.StatusOnDecision, shared.DecisionManual, nil},
		{"s-3", shared.StatusRejected, shared.DecisionAuto, ptr(shared.DecisionAuto)},
		{"s-4", shared.StatusOnDecision, shared.DecisionManual, nil},
	}
	for _, tt := range tests {
		status, hist := apps.statusOf(tt.id)
		want := pipeline
		if tt.status != shared.StatusOnDecision {
			want = append(slices.Clone(pipeline), tt.status)
		}
		if status != tt.status || !slices.Equal(hist, want) {
			t.Errorf("%s: %s %v", tt.id, status, hist)
		}
		sc, err := m.Scores().Score(ctx, tt.id)
		if err != nil || sc == nil || sc.DecisionMode != tt.mode || sc.ModelVersion != modelVersion || len(sc.Factors) == 0 {
			t.Errorf("%s score: %+v %v", tt.id, sc, err)
		}
		d, err := m.Decisions().LatestDecision(ctx, tt.id)
		if err != nil || (d == nil) != (tt.decision == nil) || (d != nil && (d.Mode != *tt.decision || d.DecidedBy != nil)) {
			t.Errorf("%s decision: %+v %v", tt.id, d, err)
		}
	}
	if d, _ := m.Decisions().LatestDecision(ctx, "s-3"); len(d.Alternatives) != 1 || d.Alternatives[0] != (shared.Alternative{ProductCode: "EXPRESS", MaxAmount: 600_000}) {
		t.Errorf("альтернатива при отказе: %+v", d.Alternatives)
	}
	if sc, _ := m.Scores().Score(ctx, "s-2"); sc.Suggested == nil || sc.Suggested.Decision != shared.StatusCounterOffer {
		t.Errorf("morozov: рекомендация %+v", sc.Suggested)
	}

	// Сотрудник: иные условия для morozov.
	if _, err := m.Decide(ctx, "s-2", shared.DecisionInput{Result: shared.StatusCounterOffer, ProductCode: "WORKING",
		Amount: 2_400_000, Term: 24, Rate: 30, Reason: "Снижение суммы"}, officer); errCode(err) != httpx.CodeBusinessRule {
		t.Fatalf("ставка вне рамок: %v", err)
	}
	d, err := m.Decide(ctx, "s-2", shared.DecisionInput{Result: shared.StatusCounterOffer, ProductCode: "WORKING",
		Amount: 2_400_000, Term: 24, Rate: 20.5, Reason: "Снижение суммы"}, officer)
	if err != nil || d.Mode != shared.DecisionManual || d.DecidedBy == nil || *d.DecidedBy != "u-9" || d.Result != shared.StatusCounterOffer {
		t.Fatalf("ручное решение: %+v %v", d, err)
	}
	if status, _ := apps.statusOf("s-2"); status != shared.StatusCounterOffer {
		t.Fatalf("статус после решения: %s", status)
	}
	if id := <-decidedCh; id != "s-2" {
		t.Fatalf("DecisionMade: %s", id)
	}

	// kozlov: запрос информации, ответ клиента — заявка снова проходит скоринг и ждёт сотрудника.
	apps.setStatus("s-4", shared.StatusDataCheck)
	if err := deps.Bus.Publish(ctx, shared.ApplicationStatusChanged{AppID: "s-4", From: shared.StatusInfoRequested, To: shared.StatusDataCheck}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "повторный скоринг kozlov", func() bool {
		status, _ := apps.statusOf("s-4")
		return status == shared.StatusOnDecision && jobsDone()
	})
	var scores int
	_ = deps.DB.GetContext(ctx, &scores, `SELECT count(*) FROM scoring.score_results WHERE app_id = 's-4'`)
	if scores != 2 {
		t.Fatalf("скорингов kozlov: %d", scores)
	}

	// Переход не удался — решение не остаётся висеть.
	apps.add(onDecision("s-5", "c-s", 1_000_000))
	apps.mu.Lock()
	apps.failTo = shared.StatusApproved
	apps.mu.Unlock()
	if _, err := m.Decide(ctx, "s-5", shared.DecisionInput{Result: shared.StatusApproved, ProductCode: "WORKING", Amount: 1_000_000, Term: 24, Rate: 20, Reason: "ok"}, officer); err == nil {
		t.Fatal("ожидалась ошибка перехода")
	}
	if d, _ := m.Decisions().LatestDecision(ctx, "s-5"); d != nil {
		t.Fatalf("решение без перехода: %+v", d)
	}
}

func ptr[T any](v T) *T { return &v }
