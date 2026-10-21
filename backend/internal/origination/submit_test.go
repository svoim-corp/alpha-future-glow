package origination

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestSubmit(t *testing.T) {
	e := newEnv(t)
	rec := &recorder{}
	rec.listen(e.bus)
	var submitted []string
	e.bus.Subscribe(shared.ApplicationSubmitted{}.Name(), func(_ context.Context, ev events.Event) error {
		submitted = append(submitted, ev.(shared.ApplicationSubmitted).AppID)
		return nil
	})
	id, version := e.ready(t, "smirnov")
	path := "/api/v1/applications/" + id + "/submit"

	w := e.do("POST", path, "smirnov", `{"truthConfirmed":true}`, "If-Match", fmt.Sprint(version), "Idempotency-Key", "submit-"+id)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[submitResponse](t, w)
	if got.ID != id || got.Status != shared.StatusSubmitted || got.Redirect != "/applications/"+id || len(got.Number) != len("КЗ-2026-000001") || got.Number[:len("КЗ-2026-")] != "КЗ-2026-" {
		t.Fatalf("ответ: %+v", got)
	}
	if len(submitted) != 1 || submitted[0] != id || len(rec.got) != 1 || rec.got[0].To != shared.StatusSubmitted {
		t.Fatalf("события: submitted=%v changed=%+v", submitted, rec.got)
	}
	v, _ := e.m.Applications().Application(context.Background(), id)
	if v.Number == nil || *v.Number != got.Number || v.SubmittedAt == nil || v.Status != shared.StatusSubmitted {
		t.Fatalf("заявка после отправки: %+v", v)
	}
	a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if !a.Locks.ParamsLocked || !a.Stage2.TruthConfirmed {
		t.Fatalf("после отправки параметры заблокированы и достоверность подтверждена: %+v", a)
	}

	// Повтор с тем же ключом — тот же ответ, без второй отправки.
	again := e.do("POST", path, "smirnov", `{"truthConfirmed":true}`, "If-Match", fmt.Sprint(version), "Idempotency-Key", "submit-"+id)
	if again.Code != http.StatusOK || again.Body.String() != w.Body.String() || len(submitted) != 1 {
		t.Fatalf("повтор по ключу: %d %s", again.Code, again.Body.String())
	}
	// Без ключа — заявка уже не черновик.
	if w := e.do("POST", path, "smirnov", `{"truthConfirmed":true}`, "If-Match", fmt.Sprint(version+1)); w.Code != http.StatusConflict {
		t.Fatalf("повторная отправка: %d", w.Code)
	}
}

func TestSubmitErrors(t *testing.T) {
	e := newEnv(t)
	id, version := e.ready(t, "smirnov")
	path := "/api/v1/applications/" + id + "/submit"
	empty := e.draft(t, "smirnov").ID

	tests := []struct {
		name, path, body, version string
		status                    int
		field                     string
	}{
		{"не подтверждена достоверность", path, `{"truthConfirmed":false}`, fmt.Sprint(version), http.StatusBadRequest, "stage2.truthConfirmed"},
		{"без подтверждения", path, `{}`, fmt.Sprint(version), http.StatusBadRequest, "truthConfirmed"},
		{"старая версия", path, `{"truthConfirmed":true}`, "1", http.StatusPreconditionFailed, ""},
		{"неполный черновик", "/api/v1/applications/" + empty + "/submit", `{"truthConfirmed":true}`, "1", http.StatusBadRequest, "prelim"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := e.do("POST", tt.path, "smirnov", tt.body, "If-Match", tt.version)
			b := decode[httpx.ErrorBody](t, w)
			if w.Code != tt.status || tt.field != "" && (len(b.Fields) == 0 || b.Fields[0].Field != tt.field) {
				t.Fatalf("%d %+v", w.Code, b)
			}
		})
	}
	if v, _ := e.m.Applications().Application(context.Background(), id); v.Status != shared.StatusDraft {
		t.Fatalf("после ошибок заявка — черновик: %s", v.Status)
	}
}
