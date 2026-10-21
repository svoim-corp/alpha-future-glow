package origination

import (
	"net/http"
	"strings"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestGet(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "petrova").ID

	got := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if got.ID != id || got.Status != shared.StatusDraft || got.Stage != stagePrelim || got.Version != 1 ||
		got.Product.Name != "Оборотный кредит" || got.Params.Amount != 1_500_000 || got.Params.Term != 24 ||
		got.Params.Goal == nil || *got.Params.Goal != "WORKING_CAPITAL" || got.Prelim != nil || got.PrelimInvalidated {
		t.Fatalf("черновик: %+v", got)
	}
	if got.Company == nil || got.Company.INN != "7712345678" || got.Company.RegDate == nil || *got.Company.RegDate != "2015-03-12" ||
		got.Stage2.ActualAddress == nil || got.Stage2.Turnover == nil || got.Stage2.Turnover.AvgMonthlyIncome != 1_840_000 {
		t.Fatalf("данные банка: %+v %+v", got.Company, got.Stage2)
	}
	if got.Authority.Type != shared.AuthorityMCHD || !got.Authority.Verified || got.Authority.ValidUntil == nil || *got.Authority.ValidUntil != "2027-03-01" {
		t.Fatalf("полномочия: %+v", got.Authority)
	}
	if got.Locks.ParamsLocked || got.Locks.Stage2Available || got.Stage2.Guarantor != nil {
		t.Fatalf("блокировки: %+v", got.Locks)
	}

	e.passPrelim(t, id)
	got = decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if got.Stage != stageFull || !got.Locks.Stage2Available || got.Prelim == nil || got.Prelim.Limit == nil || *got.Prelim.Limit != 3_300_000 {
		t.Fatalf("после этапа 1: %+v", got)
	}

	if _, err := e.m.deps.DB.Exec(`UPDATE origination.applications SET amount = 1000000 WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	got = decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if got.Stage != stagePrelim || got.Prelim != nil || !got.PrelimInvalidated {
		t.Fatalf("после смены суммы этап 1 устарел: %+v", got)
	}

	// Данные банка недоступны — заявка открывается без них.
	e.companies.upstream = true
	w := e.do("GET", "/api/v1/applications/"+id, "smirnov", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"company":null`) {
		t.Fatalf("без данных банка: %d %s", w.Code, w.Body.String())
	}
}

func TestGetAccess(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	for _, tt := range []struct {
		user, path string
		status     int
	}{
		{"orlova", "/api/v1/applications/" + id, http.StatusForbidden},
		{"ivanova", "/api/v1/applications/" + id, http.StatusForbidden},
		{"smirnov", "/api/v1/applications/a-1", http.StatusNotFound},
		{"", "/api/v1/applications/" + id, http.StatusUnauthorized},
	} {
		if w := e.do("GET", tt.path, tt.user, ""); w.Code != tt.status {
			t.Errorf("%s %s: %d, want %d", tt.user, tt.path, w.Code, tt.status)
		}
	}
}
