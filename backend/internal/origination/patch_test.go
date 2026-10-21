package origination

import (
	"net/http"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

func (e *env) patch(id, user, version, body string) (int, patchResponse, httpx.ErrorBody) {
	w := e.do("PATCH", "/api/v1/applications/"+id, user, body, "If-Match", version)
	var ok patchResponse
	var fail httpx.ErrorBody
	if w.Code == http.StatusOK {
		ok = decodeBody[patchResponse](w.Body.Bytes())
	} else {
		fail = decodeBody[httpx.ErrorBody](w.Body.Bytes())
	}
	return w.Code, ok, fail
}

func TestPatch(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	e.passPrelim(t, id)

	st, got, _ := e.patch(id, "smirnov", "1", `{"params":{"purposeText":"Закупка муки"}}`)
	if st != http.StatusOK || got.Version != 2 || got.PrelimInvalidated || !got.SavedAt.Equal(today) {
		t.Fatalf("пояснение: %d %+v", st, got)
	}
	st, got, _ = e.patch(id, "smirnov", `"2"`, `{"params":{"amount":1200000},"stage2":{"guarantor":{"personId":"p-1"}}}`)
	if st != http.StatusOK || got.Version != 3 || !got.PrelimInvalidated {
		t.Fatalf("сумма и поручитель: %d %+v", st, got)
	}
	a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if a.Params.Amount != 1_200_000 || a.Params.PurposeText == nil || a.Stage2.Guarantor == nil ||
		a.Stage2.Guarantor.FullName != "Смирнов Олег Викторович" || a.Stage != stagePrelim || !a.PrelimInvalidated {
		t.Fatalf("после сохранения: %+v", a)
	}
	if st, _, _ = e.patch(id, "smirnov", "3", `{"params":{"purposeText":""}}`); st != http.StatusOK {
		t.Fatalf("убрать пояснение: %d", st)
	}
	if a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", "")); a.Params.PurposeText != nil {
		t.Fatalf("пояснение не убрано: %v", *a.Params.PurposeText)
	}
}

func TestPatchErrors(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	ip := e.draft(t, "orlova").ID
	sent := e.draft(t, "smirnov").ID
	if _, err := e.m.deps.DB.Exec(`UPDATE origination.applications SET status = 'SUBMITTED', number = 'КЗ-2026-000001', submitted_at = now() WHERE id = $1`, sent); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, id, user, version, body string
		status                        int
		field                         string
	}{
		{"данные банка", id, "smirnov", "1", `{"company":{"inn":"7700000000"}}`, http.StatusBadRequest, "company"},
		{"обороты банка", id, "smirnov", "1", `{"stage2":{"turnover":{"avgMonthlyIncome":1}}}`, http.StatusBadRequest, "turnover"},
		{"поручитель не из кандидатов", id, "smirnov", "1", `{"stage2":{"guarantor":{"personId":"p-9"}}}`, http.StatusBadRequest, "stage2.guarantor.personId"},
		{"поручитель у ИП", ip, "orlova", "1", `{"stage2":{"guarantor":{"personId":"p-1"}}}`, http.StatusBadRequest, "stage2.guarantor.personId"},
		{"сумма вне рамок", id, "smirnov", "1", `{"params":{"amount":50000000}}`, http.StatusUnprocessableEntity, ""},
		{"отсрочка у оборотного", id, "smirnov", "1", `{"params":{"graceMonths":3}}`, http.StatusUnprocessableEntity, ""},
		{"старая версия", id, "smirnov", "0", `{"params":{"amount":1000000}}`, http.StatusPreconditionFailed, ""},
		{"без If-Match", id, "smirnov", "", `{"params":{"amount":1000000}}`, http.StatusBadRequest, ""},
		{"не черновик", sent, "smirnov", "1", `{"params":{"amount":1000000}}`, http.StatusConflict, ""},
		{"чужая заявка", id, "orlova", "1", `{"params":{"amount":1000000}}`, http.StatusForbidden, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, _, fail := e.patch(tt.id, tt.user, tt.version, tt.body)
			if st != tt.status || tt.field != "" && (len(fail.Fields) == 0 || fail.Fields[0].Field != tt.field) {
				t.Fatalf("%d %+v", st, fail)
			}
		})
	}
	if a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", "")); a.Version != 1 {
		t.Fatalf("ошибки не меняют черновик: версия %d", a.Version)
	}
}
