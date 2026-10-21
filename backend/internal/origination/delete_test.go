package origination

import (
	"context"
	"net/http"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestDelete(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	keep := e.draft(t, "smirnov").ID

	if w := e.do("DELETE", "/api/v1/applications/"+id, "orlova", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужой черновик: %d", w.Code)
	}
	if w := e.do("DELETE", "/api/v1/applications/"+id, "smirnov", ""); w.Code != http.StatusNoContent {
		t.Fatalf("удаление: %d %s", w.Code, w.Body.String())
	}
	if w := e.do("GET", "/api/v1/applications/"+id, "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("удалённый черновик для клиента: %d", w.Code)
	}
	if w := e.do("DELETE", "/api/v1/applications/"+id, "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("повторное удаление: %d", w.Code)
	}
	if p := decode[httpx.Page[applicationShortDTO]](t, e.do("GET", "/api/v1/applications", "smirnov", "")); p.Total != 1 || p.Items[0].ID != keep {
		t.Fatalf("список без удалённого: %+v", p)
	}
	// Для других зон заявка есть, в статусе CLOSED, а в истории — удаление.
	if v, _ := e.m.Applications().Application(context.Background(), id); v == nil || v.Status != shared.StatusClosed {
		t.Fatalf("удалённый черновик для зон: %+v", v)
	}
	if evs, _ := e.m.store.Events(context.Background(), id); len(evs) != 1 || evs[0].Action != actionDeleteDraft {
		t.Fatalf("история удаления: %+v", evs)
	}
}

func TestDeleteNotDraft(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	if _, err := e.m.deps.DB.Exec(`UPDATE origination.applications SET status = 'SUBMITTED', number = 'КЗ-2026-000001', submitted_at = now() WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if w := e.do("DELETE", "/api/v1/applications/"+id, "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("поданная заявка: %d", w.Code)
	}
}
