package dashboard

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

var (
	errConflict = &httpx.Error{Status: http.StatusPreconditionFailed, Code: httpx.CodeDashboardConflict,
		Message: "Настройку главной уже изменили в другом окне или на другом устройстве. Загрузите сохранённую версию"}
	errKeyReused = &httpx.Error{Status: http.StatusConflict, Code: httpx.CodeIdempotencyReused,
		Message: "Ключ идемпотентности уже использован для другой настройки"}
	// errKeyTaken — тот же ключ параллельно сохранил другой запрос; его результат отдаётся как повтор.
	errKeyTaken = errors.New("idempotency key taken")
)

// scope — чья настройка: пользователь и компания из сессии, не из тела запроса.
type scope struct{ userID, companyID string }

// precondition — ожидание клиента о записи: create — её ещё нет (If-None-Match: *), иначе её ревизия (If-Match).
type precondition struct {
	create   bool
	revision string
}

func (p precondition) holds(cur *dashboardRow) bool {
	if p.create {
		return cur == nil
	}
	return cur != nil && cur.Revision == p.revision
}

// saveInput — проверенная команда PUT. key пустой — клиент не прислал Idempotency-Key.
type saveInput struct {
	cfg      *config
	pre      precondition
	key      string
	bodyHash string
}

// save атомарно заменяет документ. Повтор по ключу проверяется раньше версии: если исходное сохранение прошло,
// клиент получает его результат, а не 412 из-за собственной же записи. Конфликт перепроверяется по ключу ещё раз:
// параллельный повтор мог увидеть запись исходной команды, закоммиченную после первой проверки.
func (m *Module) save(ctx context.Context, s scope, in saveInput) (*dashboardRow, error) {
	if in.key == "" {
		return m.write(ctx, s, in)
	}
	if d, err := m.replay(ctx, s, in); d != nil || err != nil {
		return d, err
	}
	d, err := m.write(ctx, s, in)
	if errors.Is(err, errConflict) || errors.Is(err, errKeyTaken) {
		if prev, rerr := m.replay(ctx, s, in); prev != nil || rerr != nil {
			return prev, rerr
		}
		return nil, errConflict
	}
	return d, err
}

// write проверяет версию и привязки и сохраняет документ вместе с ключом в одной транзакции.
func (m *Module) write(ctx context.Context, s scope, in saveInput) (*dashboardRow, error) {
	cur, err := m.store.Dashboard(ctx, s.userID, s.companyID)
	if err != nil {
		return nil, fmt.Errorf("load dashboard: %w", err)
	}
	if !in.pre.holds(cur) {
		return nil, errConflict
	}
	if err := m.checkSources(ctx, s.companyID, in.cfg, cur); err != nil {
		return nil, err
	}

	d := &dashboardRow{
		UserID: s.userID, CompanyID: s.companyID, Revision: newRevision(),
		Config: db.JSON[config]{V: *in.cfg}, UpdatedAt: m.now().UTC().Truncate(time.Microsecond),
	}
	err = db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		if in.key != "" {
			ok, err := m.store.InsertSaveKey(ctx, tx, &saveKeyRow{
				UserID: s.userID, CompanyID: s.companyID, Key: in.key, BodyHash: in.bodyHash,
				Revision: d.Revision, Config: d.Config, UpdatedAt: d.UpdatedAt,
			})
			if err != nil {
				return fmt.Errorf("save idempotency key: %w", err)
			}
			if !ok {
				return errKeyTaken
			}
		}
		var ok bool
		var err error
		if in.pre.create {
			ok, err = m.store.InsertDashboard(ctx, tx, d)
		} else {
			ok, err = m.store.UpdateDashboard(ctx, tx, d, in.pre.revision)
		}
		if err != nil {
			return fmt.Errorf("save dashboard: %w", err)
		}
		if !ok {
			return errConflict
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

// replay — результат прошлого сохранения с тем же ключом; (nil, nil) — ключ новый.
func (m *Module) replay(ctx context.Context, s scope, in saveInput) (*dashboardRow, error) {
	k, err := m.store.SaveKey(ctx, s.userID, s.companyID, in.key)
	if err != nil {
		return nil, fmt.Errorf("load idempotency key: %w", err)
	}
	if k == nil {
		return nil, nil
	}
	if k.BodyHash != in.bodyHash {
		return nil, errKeyReused
	}
	return &dashboardRow{UserID: k.UserID, CompanyID: k.CompanyID, Revision: k.Revision, Config: k.Config, UpdatedAt: k.UpdatedAt}, nil
}

// checkSources проверяет новые привязки к договорам, заявкам и планам. Привязки, которые уже есть в сохранённой
// настройке, не перепроверяются: потерянный доступ к старому договору не должен блокировать остальные изменения,
// а сам виджет покажет 404 доменного GET. Отсутствующий и чужой объект неразличимы для клиента.
func (m *Module) checkSources(ctx context.Context, companyID string, cfg *config, cur *dashboardRow) error {
	type key struct {
		kind refKind
		id   string
	}
	known := map[key]bool{}
	if cur != nil {
		for _, r := range refs(cur.Config.V.Widgets) {
			known[key{r.kind, r.id}] = true
		}
	}
	checked := map[key]bool{}
	var fields []httpx.FieldError
	for _, r := range refs(cfg.Widgets) {
		k := key{r.kind, r.id}
		if known[k] {
			continue
		}
		ok, seen := checked[k]
		if !seen {
			owner, err := m.owner(ctx, r)
			if err != nil {
				return fmt.Errorf("check source %s: %w", r.id, err)
			}
			ok = owner == companyID
			checked[k] = ok
		}
		if !ok {
			fields = append(fields, invalid(r.field, "Источник не найден или недоступен"))
		}
	}
	if len(fields) > 0 {
		return httpx.Validation("Выберите другой источник для виджета", fields...)
	}
	return nil
}

// owner — компания объекта; "" — объекта нет.
func (m *Module) owner(ctx context.Context, r ref) (string, error) {
	switch r.kind {
	case refLoan:
		return m.ports.Loans.LoanCompany(ctx, r.id)
	case refPlan:
		return m.ports.Plans.PlanCompany(ctx, r.id)
	case refApplication:
		app, err := m.ports.Applications.Application(ctx, r.id)
		if err != nil || app == nil {
			return "", err
		}
		return app.CompanyID, nil
	}
	return "", fmt.Errorf("unknown source kind %d", r.kind)
}

// newRevision — непрозрачный токен ревизии в виде UUID v4.
func newRevision() string {
	var b [16]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read не возвращает ошибку
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
