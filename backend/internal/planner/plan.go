package planner

import (
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const (
	modeDay         = "DAY"
	modePeriod      = "PERIOD"
	modeUnscheduled = "UNSCHEDULED"

	stateActive   = "ACTIVE"
	stateArchived = "ARCHIVED"

	dateLayout = "2006-01-02"

	matchReason = "Цель и сумма укладываются в рамки продукта"
)

// msk — часовой пояс, по которому определяется «сегодня» для начала и конца подбора (API_SPEC_ADDITIONAL.md):
// пояс компании, а все демо-компании — в Москве.
var msk = time.FixedZone("MSK", 3*60*60)

// day — календарная дата now в часовом поясе компании, в том же виде, в каком даты читаются из колонок date (полночь UTC).
func day(now time.Time) time.Time {
	y, m, d := now.In(msk).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// normalize приводит даты к режиму: у DAY конец совпадает с началом, у UNSCHEDULED дат нет.
func normalize(p *planRow) {
	switch p.DateMode {
	case modeDay:
		p.EndDate = p.StartDate
	case modeUnscheduled:
		p.StartDate, p.EndDate = nil, nil
	}
}

// validate — правила для дат: они зависят от нескольких полей сразу, поэтому не выражаются тегами binding.
// Проверяется уже слитый план — и после POST, и после PATCH.
func validate(p *planRow) []httpx.FieldError {
	var fields []httpx.FieldError
	bad := func(field, code, msg string) {
		fields = append(fields, httpx.FieldError{Field: field, Code: code, Message: msg})
	}
	if p.DateMode != modeUnscheduled && p.StartDate == nil {
		bad("startDate", "REQUIRED", "Укажите дату начала")
	}
	if p.DateMode == modePeriod && p.EndDate == nil {
		bad("endDate", "REQUIRED", "Укажите дату окончания")
	}
	if p.StartDate != nil && p.EndDate != nil && p.EndDate.Before(*p.StartDate) {
		bad("endDate", "INVALID", "Дата окончания не может быть раньше даты начала")
	}
	return fields
}

type recommendation struct {
	product       shared.ProductView
	availableFrom time.Time
}

// matchable — участвует ли план в подборе сегодня: только активный и только пока период не закончился.
func matchable(p *planRow, today time.Time) bool {
	return p.State == stateActive && (p.EndDate == nil || !today.After(*p.EndDate))
}

// recommend — первый подходящий продукт. Подбор начинается с даты начала плана, поэтому в будущем плане
// вариант показывается с availableFrom = startDate.
func recommend(p *planRow, products []shared.ProductView, today time.Time) *recommendation {
	if !matchable(p, today) || len(products) == 0 {
		return nil
	}
	from := today
	if p.StartDate != nil && p.StartDate.After(today) {
		from = *p.StartDate
	}
	return &recommendation{product: products[0], availableFrom: from}
}

// started — подбор уже идёт: только тогда о варианте сообщается уведомлением.
func (r *recommendation) started(today time.Time) bool { return !r.availableFrom.After(today) }
