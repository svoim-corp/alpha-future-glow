package underwriting

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestFilterSQL(t *testing.T) {
	tests := []struct {
		f            queueFilter
		where, order string
		args         int
	}{
		{queueFilter{}, "", " ORDER BY submitted_at DESC NULLS LAST, app_id", 0},
		{queueFilter{Status: "ON_DECISION", Sort: "amount"}, " WHERE status = $1", " ORDER BY amount ASC NULLS LAST, app_id", 1},
		{queueFilter{Status: "ON_DECISION", Product: "WORKING", Risk: "MEDIUM", DecisionMode: "MANUAL", Sort: "-score"},
			" WHERE status = $1 AND product_code = $2 AND risk_band = $3 AND decision_mode = $4", " ORDER BY score DESC NULLS LAST, app_id", 4},
	}
	for _, tt := range tests {
		where, order, args := tt.f.sql()
		if where != tt.where || order != tt.order || len(args) != tt.args {
			t.Errorf("%+v: %q %q %v", tt.f, where, order, args)
		}
	}
}

func TestShortName(t *testing.T) {
	for in, want := range map[string]string{
		"Иванова Елена Петровна": "Иванова Е. П.",
		"Сидоров Максим":         "Сидоров М.",
		"":                       "",
	} {
		if got := shortName(in); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
}

func TestQueueAccessAndFilters(t *testing.T) {
	e := newEnv(t, false)
	tests := []struct {
		path, login string
		code        int
		field       string
	}{
		{"/bank/applications", "", http.StatusUnauthorized, ""},
		{"/bank/applications", "smirnov", http.StatusForbidden, ""},
		{"/bank/applications?status=DRAFTY", "ivanova", http.StatusBadRequest, "status"},
		{"/bank/applications?risk=EXTREME", "ivanova", http.StatusBadRequest, "risk"},
		{"/bank/applications?decisionMode=SOMETIMES", "ivanova", http.StatusBadRequest, "decisionMode"},
		{"/bank/applications?sort=-company_name", "ivanova", http.StatusBadRequest, "sort"},
	}
	for _, tt := range tests {
		w := e.do("GET", tt.path, tt.login, "")
		var b httpx.ErrorBody
		_ = json.Unmarshal(w.Body.Bytes(), &b)
		if w.Code != tt.code || (tt.field != "" && (len(b.Fields) != 1 || b.Fields[0].Field != tt.field)) {
			t.Errorf("%s: %d %+v", tt.path, w.Code, b)
		}
	}
}

func TestQueue(t *testing.T) {
	e := newEnv(t, true)
	t0 := today.Add(-3 * time.Hour)
	e.apps.add(submittedApp("a-1", "c-1", "WORKING", 1_500_000, t0))
	e.apps.add(submittedApp("a-2", "c-8", "WORKING", 3_000_000, t0.Add(time.Hour)))
	e.apps.add(submittedApp("a-3", "c-1", "EXPRESS", 500_000, t0.Add(2*time.Hour)))
	draft := submittedApp("a-4", "c-1", "WORKING", 1_000_000, t0)
	draft.SubmittedAt = nil
	e.apps.add(draft)

	e.move(t, "a-4", shared.StatusClosed) // удалённый черновик в очередь не попадает
	e.move(t, "a-1", shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusScoring)
	e.scores["a-1"] = &shared.ScoreView{Score: 740, RiskBand: shared.RiskLow, DecisionMode: shared.DecisionAuto}
	e.move(t, "a-1", shared.StatusOnDecision, shared.StatusApproved)
	e.move(t, "a-2", shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusScoring)
	e.scores["a-2"] = &shared.ScoreView{Score: 660, RiskBand: shared.RiskMedium, DecisionMode: shared.DecisionManual}
	e.move(t, "a-2", shared.StatusOnDecision)
	e.move(t, "a-3", shared.StatusSubmitted)

	var stats statsResponse
	w := e.do("GET", "/bank/applications/stats", "ivanova", "")
	_ = json.Unmarshal(w.Body.Bytes(), &stats)
	if w.Code != http.StatusOK || stats != (statsResponse{Submitted: 1, OnDecision: 1, DecidedToday: 1}) {
		t.Fatalf("stats: %d %+v", w.Code, stats)
	}

	list := func(query string) (ids []string, page httpx.Page[queueItemDTO]) {
		t.Helper()
		w := e.do("GET", "/bank/applications"+query, "ivanova", "")
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", query, w.Code, w.Body)
		}
		_ = json.Unmarshal(w.Body.Bytes(), &page)
		for _, it := range page.Items {
			ids = append(ids, it.ID)
		}
		return ids, page
	}
	tests := []struct {
		query string
		ids   []string
	}{
		{"", []string{"a-3", "a-2", "a-1"}},
		{"?sort=submittedAt", []string{"a-1", "a-2", "a-3"}},
		{"?sort=-score", []string{"a-1", "a-2", "a-3"}}, // без балла — в конце
		{"?status=ON_DECISION&decisionMode=MANUAL&risk=MEDIUM", []string{"a-2"}},
		{"?product=EXPRESS", []string{"a-3"}},
		{"?status=REJECTED", nil},
	}
	for _, tt := range tests {
		if ids, _ := list(tt.query); !slices.Equal(ids, tt.ids) {
			t.Errorf("%q: %v, want %v", tt.query, ids, tt.ids)
		}
	}
	ids, page := list("?page=2&size=2")
	if !slices.Equal(ids, []string{"a-1"}) || page.Total != 3 || page.Page != 2 {
		t.Fatalf("пагинация: %v %+v", ids, page)
	}
	_, page = list("?status=ON_DECISION")
	it := page.Items[0]
	if it.Company.INN != "7707789050" || it.Product.Name != "Оборотный кредит" || *it.Score != 660 || *it.DecisionMode != shared.DecisionManual || it.Assignee != nil {
		t.Fatalf("строка очереди: %+v", it)
	}

	// Взять в работу: повтор своим — ок, другой сотрудник — 409, чужой заявки нет — 404.
	for _, tt := range []struct {
		login, id string
		code      int
	}{
		{"ivanova", "a-2", http.StatusOK}, {"ivanova", "a-2", http.StatusOK},
		{"sidorov", "a-2", http.StatusConflict}, {"sidorov", "a-404", http.StatusNotFound},
	} {
		if w := e.do("POST", "/bank/applications/"+tt.id+"/assign", tt.login, ""); w.Code != tt.code {
			t.Errorf("%s assign %s: %d %s", tt.login, tt.id, w.Code, w.Body)
		}
	}
	_, page = list("?status=ON_DECISION")
	if a := page.Items[0].Assignee; a == nil || a.ID != "u-9" || a.FullName != "Иванова Е. П." {
		t.Fatalf("assignee: %+v", a)
	}
	// Смена статуса не сбрасывает назначение и известный балл.
	e.scores["a-2"] = nil
	e.move(t, "a-2", shared.StatusInfoRequested)
	_, page = list("?status=INFO_REQUESTED")
	if it := page.Items[0]; it.Assignee == nil || it.Score == nil || *it.Score != 660 {
		t.Fatalf("после смены статуса: %+v", it)
	}
}
