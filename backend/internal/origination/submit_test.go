package origination

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

func TestTimeline(t *testing.T) {
	e := newEnv(t)
	id, version := e.ready(t, "smirnov")
	if w := e.do("POST", "/api/v1/applications/"+id+"/submit", "smirnov", `{"truthConfirmed":true}`, "If-Match", fmt.Sprint(version)); w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	w := e.do("GET", "/api/v1/applications/"+id+"/timeline", "smirnov", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[timelineResponse](t, w)
	if got.Current.Status != shared.StatusSubmitted || got.Current.Label != "Подана" || got.Current.Actor != turnBank || got.Current.WhatsHappening == "" {
		t.Fatalf("текущий статус: %+v", got.Current)
	}
	if len(got.Steps) != 4 || got.Steps[0].State != stepCurrent || got.Steps[0].At == nil || !got.Steps[0].At.Equal(today) || got.Steps[1].At != nil {
		t.Fatalf("шаги: %+v", got.Steps)
	}
	if w := e.do("GET", "/api/v1/applications/"+id+"/timeline", "orlova", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужая заявка: %d", w.Code)
	}
}

func TestBankHistory(t *testing.T) {
	e := newEnv(t)
	id, version := e.ready(t, "petrova")
	if w := e.do("POST", "/api/v1/applications/"+id+"/submit", "petrova", `{"truthConfirmed":true}`, "If-Match", fmt.Sprint(version)); w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	apps := e.m.Applications()
	ctx := context.Background()
	if err := apps.Transition(ctx, id, shared.StatusDataCheck, shared.Actor{Type: shared.ActorSystem}, "CHECKS_DONE", ""); err != nil {
		t.Fatal(err)
	}
	officer := shared.Actor{Type: shared.ActorOfficer, UserID: "u-9", Name: "Иванова Елена Петровна"}
	if err := apps.Transition(ctx, id, shared.StatusInfoRequested, officer, "INFO_REQUEST", "Уточните адрес склада"); err != nil {
		t.Fatal(err)
	}

	w := e.do("GET", "/api/v1/bank/applications/"+id+"/history", "ivanova", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[[]historyEntryDTO](t, w)
	if len(got) != 3 {
		t.Fatalf("история: %+v", got)
	}
	if h := got[0]; h.Action != actionSubmit || h.From != shared.StatusDraft || h.To != shared.StatusSubmitted ||
		h.Actor.Type != shared.ActorClient || h.Actor.Name == nil || *h.Actor.Name != "Петрова А. С." {
		t.Fatalf("отправка: %+v", h)
	}
	if h := got[1]; h.Actor.Type != shared.ActorSystem || h.Actor.Name != nil || h.Comment != nil {
		t.Fatalf("система: %+v", h)
	}
	if h := got[2]; h.Actor.Name == nil || *h.Actor.Name != "Иванова Е. П." || h.Comment == nil || *h.Comment != "Уточните адрес склада" {
		t.Fatalf("сотрудник: %+v", h)
	}
	if !strings.Contains(w.Body.String(), `"actor":{"type":"SYSTEM"}`) {
		t.Fatalf("у системы нет имени: %s", w.Body.String())
	}

	if w := e.do("GET", "/api/v1/bank/applications/"+id+"/history", "smirnov", ""); w.Code != http.StatusForbidden {
		t.Fatalf("клиенту история банка недоступна: %d", w.Code)
	}
	if w := e.do("GET", "/api/v1/bank/applications/a-0/history", "ivanova", ""); w.Code != http.StatusNotFound {
		t.Fatalf("нет заявки: %d", w.Code)
	}
}

func TestShortName(t *testing.T) {
	for in, want := range map[string]string{"Иванова Елена Петровна": "Иванова Е. П.", "Петрова": "Петрова", "": ""} {
		if got := shortName(in); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
}
