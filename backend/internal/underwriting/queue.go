package underwriting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// decidedStatuses — статусы решения: с ними заявка попадает в счётчик decidedToday.
var decidedStatuses = map[shared.Status]bool{
	shared.StatusApproved: true, shared.StatusCounterOffer: true, shared.StatusRejected: true,
}

// onStatusChanged обновляет очередь по событию Origination. Черновики в очередь не попадают.
func (m *Module) onStatusChanged(ctx context.Context, e events.Event) error {
	ev := e.(shared.ApplicationStatusChanged)
	if ev.To == shared.StatusDraft || ev.From == shared.StatusDraft && ev.To == shared.StatusClosed {
		return nil
	}
	return m.refreshQueue(ctx, ev.AppID, ev.At)
}

// refreshQueue перечитывает заявку, компанию, продукт и скоринг и записывает строку очереди.
func (m *Module) refreshQueue(ctx context.Context, appID string, at time.Time) error {
	app, err := m.ports.Applications.Application(ctx, appID)
	if err != nil {
		return fmt.Errorf("load application %s: %w", appID, err)
	}
	if app == nil || app.SubmittedAt == nil {
		return nil
	}
	cv, err := m.ports.Companies.GetByID(ctx, app.CompanyID)
	if err != nil {
		return fmt.Errorf("load company %s: %w", app.CompanyID, err)
	}
	if cv == nil {
		return fmt.Errorf("company %s of application %s not found", app.CompanyID, app.ID)
	}
	product, err := m.ports.Products.Product(ctx, app.ProductCode)
	if err != nil {
		return fmt.Errorf("load product %s: %w", app.ProductCode, err)
	}
	productName := app.ProductCode
	if product != nil {
		productName = product.Name
	}
	sc, err := m.ports.Scores.Score(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("load score %s: %w", app.ID, err)
	}

	row := &queueRow{
		AppID: app.ID, Number: app.Number, CompanyID: cv.ID, CompanyINN: cv.INN, CompanyName: cv.Name,
		ProductCode: app.ProductCode, ProductName: productName, Amount: app.Amount, Term: app.Term,
		Status: string(app.Status), SubmittedAt: app.SubmittedAt.UTC(), UpdatedAt: at.UTC(),
	}
	if sc != nil {
		band, mode := string(sc.RiskBand), string(sc.DecisionMode)
		row.Score, row.RiskBand, row.DecisionMode = &sc.Score, &band, &mode
	}
	if decidedStatuses[app.Status] {
		t := at.UTC()
		row.DecidedAt = &t
	}
	if err := m.store.UpsertQueue(ctx, row); err != nil {
		return fmt.Errorf("upsert queue %s: %w", app.ID, err)
	}
	return nil
}

// assign берёт заявку в работу. Повторное назначение на себя — не ошибка.
func (m *Module) assign(ctx context.Context, appID, userID, fullName string) (string, error) {
	name := shortName(fullName)
	ok, err := m.store.Assign(ctx, appID, userID, name, m.now().UTC())
	if err != nil {
		return "", fmt.Errorf("assign %s: %w", appID, err)
	}
	if ok {
		return name, nil
	}
	item, err := m.store.QueueItem(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("load queue item %s: %w", appID, err)
	}
	if item == nil {
		return "", httpx.NotFound("Заявка не найдена в очереди")
	}
	return "", httpx.InvalidTransition("Заявку уже взял в работу другой сотрудник")
}

// shortName — «Иванова Елена Петровна» → «Иванова Е. П.».
func shortName(full string) string {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " " + string([]rune(p)[0]) + "."
	}
	return out
}
