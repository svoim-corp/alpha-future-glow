package origination

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// path — история заявки как цепочка статусов, начиная с черновика.
func path(statuses ...shared.Status) []eventRow {
	var out []eventRow
	from := shared.StatusDraft
	for i, to := range statuses {
		out = append(out, eventRow{From: string(from), To: string(to), At: today.Add(time.Duration(i) * time.Minute)})
		from = to
	}
	return out
}

// render — шаги в виде «STATUS:STATE» для сравнения.
func render(ss []timelineStep) string {
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = fmt.Sprintf("%s:%s", s.status, s.state)
		if (s.at != nil) != (s.state != stepPending) {
			return "время должно быть только у пройденных шагов: " + parts[i]
		}
	}
	return strings.Join(parts, " ")
}

func TestSteps(t *testing.T) {
	s := shared.StatusSubmitted
	tests := []struct {
		name    string
		current shared.Status
		events  []eventRow
		want    string
	}{
		{"черновик", shared.StatusDraft, nil,
			"SUBMITTED:PENDING DATA_CHECK:PENDING SCORING:PENDING ON_DECISION:PENDING"},
		{"подана", s, path(s),
			"SUBMITTED:CURRENT DATA_CHECK:PENDING SCORING:PENDING ON_DECISION:PENDING"},
		{"запрос информации", shared.StatusInfoRequested, path(s, shared.StatusDataCheck, shared.StatusInfoRequested),
			"SUBMITTED:DONE DATA_CHECK:DONE INFO_REQUESTED:CURRENT DATA_CHECK:PENDING SCORING:PENDING ON_DECISION:PENDING"},
		{"после ответа — снова проверка", shared.StatusDataCheck,
			path(s, shared.StatusDataCheck, shared.StatusInfoRequested, shared.StatusDataCheck),
			"SUBMITTED:DONE DATA_CHECK:DONE INFO_REQUESTED:DONE DATA_CHECK:CURRENT SCORING:PENDING ON_DECISION:PENDING"},
		{"на решении", shared.StatusOnDecision, path(s, shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision),
			"SUBMITTED:DONE DATA_CHECK:DONE SCORING:DONE ON_DECISION:CURRENT"},
		{"одобрено", shared.StatusApproved, path(s, shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision, shared.StatusApproved),
			"SUBMITTED:DONE DATA_CHECK:DONE SCORING:DONE ON_DECISION:DONE APPROVED:CURRENT SIGNING:PENDING ISSUED:PENDING"},
		{"выдано", shared.StatusIssued,
			path(s, shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision, shared.StatusApproved, shared.StatusSigning, shared.StatusIssued),
			"SUBMITTED:DONE DATA_CHECK:DONE SCORING:DONE ON_DECISION:DONE APPROVED:DONE SIGNING:DONE ISSUED:CURRENT"},
		{"удалённый черновик", shared.StatusClosed, path(shared.StatusClosed), "CLOSED:CURRENT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(steps(tt.current, tt.events)); got != tt.want {
				t.Errorf("\n got %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestStatusTexts(t *testing.T) {
	for _, s := range allStatuses {
		if txt, ok := statusTexts[s]; !ok || txt.whatsHappening == "" || txt.nextStep == "" || txt.turn == "" {
			t.Errorf("нет текста для %s", s)
		}
	}
}
