package planner

import (
	"slices"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func date(s string) *time.Time {
	t, _ := time.Parse(dateLayout, s)
	return &t
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		p      planRow
		fields []string
	}{
		{"без дат", planRow{DateMode: modeUnscheduled}, nil},
		{"день", planRow{DateMode: modeDay, StartDate: date("2027-06-01")}, nil},
		{"период", planRow{DateMode: modePeriod, StartDate: date("2027-06-01"), EndDate: date("2027-08-31")}, nil},
		{"день без даты", planRow{DateMode: modeDay}, []string{"startDate"}},
		{"период без конца", planRow{DateMode: modePeriod, StartDate: date("2027-06-01")}, []string{"endDate"}},
		{"конец раньше начала", planRow{DateMode: modePeriod, StartDate: date("2027-06-01"), EndDate: date("2027-05-31")}, []string{"endDate"}},
		{"период без дат", planRow{DateMode: modePeriod}, []string{"startDate", "endDate"}},
	}
	for _, tt := range tests {
		normalize(&tt.p)
		var got []string
		for _, f := range validate(&tt.p) {
			got = append(got, f.Field)
		}
		if !slices.Equal(got, tt.fields) {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.fields)
		}
	}
}

func TestNormalize(t *testing.T) {
	p := planRow{DateMode: modeDay, StartDate: date("2027-06-01"), EndDate: date("2027-12-31")}
	normalize(&p)
	if !p.EndDate.Equal(*p.StartDate) {
		t.Fatalf("DAY: конец %v", p.EndDate)
	}
	p = planRow{DateMode: modeUnscheduled, StartDate: date("2027-06-01"), EndDate: date("2027-12-31")}
	normalize(&p)
	if p.StartDate != nil || p.EndDate != nil {
		t.Fatalf("UNSCHEDULED: даты %v %v", p.StartDate, p.EndDate)
	}
}

func TestRecommend(t *testing.T) {
	today := *date("2026-10-21")
	invest := []shared.ProductView{{Code: "INVEST", Name: "Инвестиционный кредит"}, {Code: "WORKING"}}
	tests := []struct {
		name    string
		p       planRow
		product string
		from    string
		started bool
	}{
		{"без дат — сразу", planRow{State: stateActive, DateMode: modeUnscheduled}, "INVEST", "2026-10-21", true},
		{"период идёт", planRow{State: stateActive, StartDate: date("2026-10-01"), EndDate: date("2026-12-31")}, "INVEST", "2026-10-21", true},
		{"последний день периода", planRow{State: stateActive, StartDate: date("2026-10-01"), EndDate: date("2026-10-21")}, "INVEST", "2026-10-21", true},
		{"период в будущем", planRow{State: stateActive, StartDate: date("2027-06-01"), EndDate: date("2027-08-31")}, "INVEST", "2027-06-01", false},
		{"период закончился", planRow{State: stateActive, StartDate: date("2026-09-01"), EndDate: date("2026-10-20")}, "", "", false},
		{"на паузе", planRow{State: "PAUSED", DateMode: modeUnscheduled}, "", "", false},
	}
	for _, tt := range tests {
		r := recommend(&tt.p, invest, today)
		if tt.product == "" {
			if r != nil {
				t.Errorf("%s: ожидался null, получено %+v", tt.name, r)
			}
			continue
		}
		if r == nil || r.product.Code != tt.product || r.availableFrom.Format(dateLayout) != tt.from || r.started(today) != tt.started {
			t.Errorf("%s: %+v", tt.name, r)
		}
	}
	if r := recommend(&planRow{State: stateActive}, nil, today); r != nil {
		t.Errorf("нет продуктов — нет рекомендации: %+v", r)
	}
}

// TestDay — «сегодня» по Москве: в 22:30 UTC в Москве уже следующий день.
func TestDay(t *testing.T) {
	if got := day(time.Date(2026, 10, 21, 22, 30, 0, 0, time.UTC)); got.Format(dateLayout) != "2026-10-22" {
		t.Fatalf("%s", got)
	}
	if got := day(time.Date(2026, 10, 21, 20, 59, 0, 0, time.UTC)); got.Format(dateLayout) != "2026-10-21" {
		t.Fatalf("%s", got)
	}
}
