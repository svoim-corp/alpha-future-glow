package origination

import (
	"net/http"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestList(t *testing.T) {
	e := newEnv(t)
	first := e.draft(t, "smirnov")
	second := e.draft(t, "petrova") // та же компания c-1
	e.draft(t, "orlova")            // другая компания
	if _, err := e.m.deps.DB.Exec(`UPDATE origination.applications SET updated_at = updated_at + interval '1 hour', status = 'INFO_REQUESTED',
		number = 'КЗ-2026-001001', submitted_at = now() WHERE id = $1`, first.ID); err != nil {
		t.Fatal(err)
	}

	w := e.do("GET", "/api/v1/applications?mine=true", "smirnov", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[httpx.Page[applicationShortDTO]](t, w)
	if got.Total != 2 || len(got.Items) != 2 || got.Items[0].ID != first.ID || got.Items[1].ID != second.ID {
		t.Fatalf("заявки компании: %+v", got)
	}
	a, d := got.Items[0], got.Items[1]
	if a.Status != shared.StatusInfoRequested || a.StatusLabel != "Нужна доп. информация" || !a.ActionRequired || a.Number == nil {
		t.Fatalf("поданная: %+v", a)
	}
	if d.Number != nil || d.ActionRequired || d.Product.Name != "Оборотный кредит" || d.Amount != 1_500_000 || d.Term != 24 {
		t.Fatalf("черновик: %+v", d)
	}

	if p := decode[httpx.Page[applicationShortDTO]](t, e.do("GET", "/api/v1/applications?page=2&size=1", "smirnov", "")); p.Total != 2 || len(p.Items) != 1 || p.Items[0].ID != second.ID {
		t.Fatalf("пагинация: %+v", p)
	}
	if w := e.do("GET", "/api/v1/applications", "ivanova", ""); w.Code != http.StatusForbidden {
		t.Fatalf("сотрудник банка: %d", w.Code)
	}
}
