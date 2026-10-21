package scoring

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Конвейер после отправки заявки (API_SPEC.md, раздел 2): проверки → DATA_CHECK → SCORING → ON_DECISION → авторешение
// или ожидание сотрудника. Работает в фоне, чтобы submit не ждал моки.
const (
	pollInterval   = 5 * time.Second  // повтор неудачных задач и подстраховка, если сигнал потерялся
	jobTimeout     = 30 * time.Second // одна заявка целиком, включая все моки
	jobBatch       = 20
	maxJobAttempts = 5
	maxSteps       = 8 // защита от зацикливания, если статус меняют параллельно

	actionChecksDone = "CHECKS_DONE"
	actionScoring    = "SCORING_STARTED"
	actionScored     = "SCORED"
)

// alternativeProducts — что предложить при отказе (API_SPEC.md, 4.6: «отказ с альтернативой»).
var alternativeProducts = []string{"EXPRESS"}

var systemActor = shared.Actor{Type: shared.ActorSystem}

// subscribe — заявка попадает в конвейер после отправки и после ответа клиента на запрос информации.
// Обработчики только ставят задачу: шина синхронная, и submit ждал бы весь конвейер.
func (m *Module) subscribe(bus *events.Bus) {
	bus.Subscribe(shared.ApplicationSubmitted{}.Name(), func(ctx context.Context, e events.Event) error {
		return m.enqueue(ctx, e.(shared.ApplicationSubmitted).AppID)
	})
	bus.Subscribe(shared.ApplicationStatusChanged{}.Name(), func(ctx context.Context, e events.Event) error {
		ev := e.(shared.ApplicationStatusChanged)
		if ev.From == shared.StatusInfoRequested && ev.To == shared.StatusDataCheck {
			return m.enqueue(ctx, ev.AppID)
		}
		return nil
	})
}

func (m *Module) enqueue(ctx context.Context, appID string) error {
	if err := m.store.EnqueueJob(ctx, appID, m.now().UTC()); err != nil {
		return fmt.Errorf("enqueue pipeline %s: %w", appID, err)
	}
	select {
	case m.wake <- struct{}{}:
	default: // воркер уже разбужен
	}
	return nil
}

// Run — воркер конвейера, один на процесс. Возвращается после отмены ctx; прерванная задача остаётся PENDING
// и продолжится с текущего статуса заявки после рестарта.
func (m *Module) Run(ctx context.Context) {
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		m.drain(ctx)
		select {
		case <-ctx.Done():
			return
		case <-m.wake:
		case <-t.C:
		}
	}
}

// drain обрабатывает одну пачку задач. Неудачная задача повторится не раньше следующего тика.
func (m *Module) drain(ctx context.Context) {
	jobs, err := m.store.PendingJobs(ctx, jobBatch)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("load pipeline jobs", "err", err)
		}
		return
	}
	for _, j := range jobs {
		jctx, cancel := context.WithTimeout(ctx, jobTimeout)
		err := m.process(jctx, j.AppID)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			err = m.store.FinishJob(ctx, j.AppID, m.now().UTC())
		} else {
			slog.Warn("pipeline step failed", "appId", j.AppID, "attempt", j.Attempts+1, "err", err)
			err = m.store.RetryJob(ctx, j.AppID, err.Error(), maxJobAttempts, m.now().UTC())
		}
		if err != nil {
			slog.Error("save pipeline job", "appId", j.AppID, "err", err)
		}
	}
}

// process ведёт заявку по статусам, пока не дойдёт до решения или до сотрудника. Каждый шаг смотрит на текущий
// статус, поэтому повтор после сбоя безопасен.
func (m *Module) process(ctx context.Context, appID string) error {
	for range maxSteps {
		app, err := m.ports.Applications.Application(ctx, appID)
		if err != nil {
			return fmt.Errorf("load application %s: %w", appID, err)
		}
		if app == nil {
			return nil
		}
		done, err := m.step(ctx, app)
		var he *httpx.Error
		if errors.As(err, &he) && he.Code == httpx.CodeInvalidTransition {
			continue // статус сменил сотрудник — смотрим заново
		}
		if err != nil || done {
			return err
		}
	}
	return fmt.Errorf("pipeline %s: no progress after %d steps", appID, maxSteps)
}

func (m *Module) step(ctx context.Context, app *shared.ApplicationView) (done bool, err error) {
	switch app.Status {
	case shared.StatusSubmitted:
		if _, err := m.runChecks(ctx, app); err != nil {
			return false, err
		}
		return false, m.ports.Applications.Transition(ctx, app.ID, shared.StatusDataCheck, systemActor, actionChecksDone, "")
	case shared.StatusDataCheck:
		return false, m.ports.Applications.Transition(ctx, app.ID, shared.StatusScoring, systemActor, actionScoring, "")
	case shared.StatusScoring:
		sc, err := m.scoreApplication(ctx, app)
		if err != nil {
			return false, err
		}
		return false, m.ports.Applications.Transition(ctx, app.ID, shared.StatusOnDecision, systemActor, actionScored, fmt.Sprintf("Балл %d", sc.Score))
	case shared.StatusOnDecision:
		return true, m.autoDecide(ctx, app)
	}
	return true, nil
}

func (m *Module) runChecks(ctx context.Context, app *shared.ApplicationView) (*shared.CheckReport, error) {
	cv, err := m.ports.Companies.GetByID(ctx, app.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("load company %s: %w", app.CompanyID, err)
	}
	if cv == nil {
		return nil, fmt.Errorf("company %s of application %s not found", app.CompanyID, app.ID)
	}
	req := shared.CheckRequest{AppID: app.ID, INN: cv.INN, CompanyType: cv.Type}
	consent, err := m.ports.Applications.BKIConsent(ctx, app.ID, shared.ConsentApplicant)
	if err != nil {
		return nil, fmt.Errorf("load bki consent %s: %w", app.ID, err)
	}
	if consent != nil {
		req.ConsentID = consent.ID
	}
	return m.ports.Checks.Run(ctx, req)
}

func (m *Module) scoreApplication(ctx context.Context, app *shared.ApplicationView) (*scoreRow, error) {
	rep, err := m.ports.Checks.Latest(ctx, app.ID)
	if err != nil {
		return nil, fmt.Errorf("load checks %s: %w", app.ID, err)
	}
	if len(rep.Results) == 0 {
		if rep, err = m.runChecks(ctx, app); err != nil {
			return nil, err
		}
	}
	product, err := m.ports.Products.Product(ctx, app.ProductCode)
	if err != nil {
		return nil, fmt.Errorf("load product %s: %w", app.ProductCode, err)
	}
	if product == nil {
		return nil, fmt.Errorf("product %s of application %s not found", app.ProductCode, app.ID)
	}
	prelim, err := m.store.LatestPrelim(ctx, app.ID)
	if err != nil {
		return nil, fmt.Errorf("load prelim %s: %w", app.ID, err)
	}

	res := score(scoreInput{report: rep, product: *product, amount: app.Amount, term: app.Term, prelim: prelim, now: m.now().UTC()})
	row := &scoreRow{
		AppID: app.ID, Score: res.score, RiskBand: string(res.band), Factors: db.JSON[[]shared.ScoreFactor]{V: res.factors},
		AutoEligible: res.eligible, AutoReason: res.reason, Suggested: db.JSON[shared.SuggestedDecision]{V: res.suggested},
		ModelVersion: modelVersion, CreatedAt: m.now().UTC(),
	}
	if err := m.store.InsertScore(ctx, row); err != nil {
		return nil, fmt.Errorf("save score %s: %w", app.ID, err)
	}
	return row, nil
}

// autoDecide принимает решение, если скоринг это разрешает; иначе заявка ждёт сотрудника в статусе ON_DECISION.
func (m *Module) autoDecide(ctx context.Context, app *shared.ApplicationView) error {
	if d, err := m.store.LatestDecision(ctx, app.ID); err != nil || d != nil {
		return err
	}
	sc, err := m.store.LatestScore(ctx, app.ID)
	if err != nil || sc == nil || !sc.AutoEligible {
		return err
	}
	s := sc.Suggested.V
	in := shared.DecisionInput{Result: s.Decision, ProductCode: app.ProductCode, Amount: s.Amount, Term: app.Term, Rate: s.Rate}
	switch s.Decision {
	case shared.StatusApproved:
		in.Reason = fmt.Sprintf("Автоматическое одобрение: балл %d", sc.Score)
	case shared.StatusRejected:
		in.Reason = fmt.Sprintf("Автоматический отказ: балл %d", sc.Score)
		if in.Alternatives, err = m.alternatives(ctx, app); err != nil {
			return err
		}
	default:
		return nil // иные условия автоматически не предлагаем: их подбирает сотрудник
	}
	_, err = m.Decide(ctx, app.ID, in, systemActor)
	return err
}

// alternatives — другие продукты на сумму не больше половины предварительного лимита.
func (m *Module) alternatives(ctx context.Context, app *shared.ApplicationView) ([]shared.Alternative, error) {
	prelim, err := m.store.LatestPrelim(ctx, app.ID)
	if err != nil || prelim == nil || prelim.LimitAmount == 0 {
		return nil, err
	}
	var out []shared.Alternative
	for _, code := range alternativeProducts {
		if code == app.ProductCode {
			continue
		}
		p, err := m.ports.Products.Product(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("load product %s: %w", code, err)
		}
		if p == nil {
			continue
		}
		maxAmount := min(p.AmountMax, prelim.LimitAmount/2/limitStep*limitStep)
		if maxAmount >= p.AmountMin {
			out = append(out, shared.Alternative{ProductCode: code, MaxAmount: maxAmount})
		}
	}
	return out, nil
}

func (r *scoreRow) view() *shared.ScoreView {
	mode := shared.DecisionManual
	if r.AutoEligible {
		mode = shared.DecisionAuto
	}
	v := &shared.ScoreView{
		Score: r.Score, RiskBand: shared.RiskBand(r.RiskBand), AutoDecisionEligible: r.AutoEligible,
		AutoDecisionReason: r.AutoReason, DecisionMode: mode, Factors: r.Factors.V, ModelVersion: r.ModelVersion,
		ScoredAt: r.CreatedAt,
	}
	if r.Suggested.V.Decision != "" {
		s := r.Suggested.V
		v.Suggested = &s
	}
	return v
}
