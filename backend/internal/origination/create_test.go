package origination

import (
	"net/http"
	"strings"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestCreate(t *testing.T) {
	e := newEnv(t)
	got := e.draft(t, "smirnov")
	if !strings.HasPrefix(got.ID, "a-") || got.Status != shared.StatusDraft || got.Version != 1 || got.Redirect != "/applications/"+got.ID+"/edit" {
		t.Fatalf("ответ: %+v", got)
	}
	var row appRow
	if err := e.m.deps.DB.Get(&row, `SELECT company_id, applicant_id, goal, authority FROM origination.applications WHERE id = $1`, got.ID); err != nil {
		t.Fatal(err)
	}
	if row.CompanyID != "c-1" || row.ApplicantID != "u-1" || row.Goal == nil || *row.Goal != "WORKING_CAPITAL" ||
		row.Authority.V.Type != shared.AuthorityHead || len(row.Authority.V.Powers) != 2 {
		t.Fatalf("черновик: %+v", row)
	}

	// Представитель по МЧД: снимок с номером и сроком доверенности.
	mchd := e.draft(t, "petrova")
	if err := e.m.deps.DB.Get(&row, `SELECT authority FROM origination.applications WHERE id = $1`, mchd.ID); err != nil {
		t.Fatal(err)
	}
	if a := row.Authority.V; a.Type != shared.AuthorityMCHD || a.MCHDNumber == nil || a.ValidUntil == nil {
		t.Fatalf("снимок МЧД: %+v", a)
	}
}

func TestCreateErrors(t *testing.T) {
	e := newEnv(t)
	tests := []struct {
		name, user, body string
		status           int
		code             string
	}{
		{"нет полномочий", "volkov", `{"productCode":"WORKING","amount":1500000,"term":24}`, http.StatusForbidden, httpx.CodeForbidden},
		{"сотрудник банка", "ivanova", `{"productCode":"WORKING","amount":1500000,"term":24}`, http.StatusForbidden, httpx.CodeForbidden},
		{"без входа", "", `{"productCode":"WORKING","amount":1500000,"term":24}`, http.StatusUnauthorized, httpx.CodeUnauthorized},
		{"сумма вне рамок", "smirnov", `{"productCode":"WORKING","amount":100,"term":24}`, http.StatusUnprocessableEntity, httpx.CodeBusinessRule},
		{"чужая цель", "smirnov", `{"productCode":"WORKING","amount":1500000,"term":24,"goal":"EQUIPMENT"}`, http.StatusUnprocessableEntity, httpx.CodeBusinessRule},
		{"неизвестный продукт", "smirnov", `{"productCode":"NOPE","amount":1500000,"term":24}`, http.StatusUnprocessableEntity, httpx.CodeBusinessRule},
		{"без срока", "smirnov", `{"productCode":"WORKING","amount":1500000}`, http.StatusBadRequest, httpx.CodeValidation},
		{"без суммы", "smirnov", `{"productCode":"WORKING","term":24}`, http.StatusBadRequest, httpx.CodeValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := e.do("POST", "/api/v1/applications", tt.user, tt.body)
			if b := decode[httpx.ErrorBody](t, w); w.Code != tt.status || b.Code != tt.code {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
	w := e.do("POST", "/api/v1/applications", "volkov", `{"productCode":"WORKING","amount":1500000,"term":24}`)
	if b := decode[httpx.ErrorBody](t, w); !strings.Contains(b.Message, "доверенности истёк") {
		t.Fatalf("причина отказа в полномочиях: %q", b.Message)
	}
}
