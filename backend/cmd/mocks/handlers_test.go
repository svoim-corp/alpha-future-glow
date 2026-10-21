package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func newRouter() *gin.Engine {
	r := gin.New()
	(&server{data: newDataset()}).register(r)
	return r
}

func do(t *testing.T, r http.Handler, method, path, body string, out any) int {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if out != nil && w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
	}
	return w.Code
}

// scenarios — строки backend/SCENARIOS.md: ИНН компании, заявитель и его роль в АБС.
var scenarios = []struct {
	login, inn, typ, personINN, role string
}{
	{"smirnov", "7712345678", "LLC", "771234567890", roleHead},
	{"petrova", "7712345678", "LLC", "771298765432", roleRepresentative},
	{"volkov", "7704211530", "LLC", "770411223344", roleRepresentative},
	{"orlova", "772512345601", "IE", "772512345601", roleHead},
	{"belov", "7728334410", "LLC", "772855667788", roleHead},
	{"gromov", "7719456720", "LLC", "771966778899", roleHead},
	{"zaitsev", "7733567830", "LLC", "773377889900", roleHead},
	{"sokolova", "7716678940", "LLC", "771688990011", roleHead},
	{"morozov", "7707789050", "LLC", "770799001122", roleHead},
	{"kozlov", "7724890160", "LLC", "772400112233", roleHead},
	{"polyakov", "7731901270", "LLC", "773111223344", roleHead},
	{"kuznetsov", "773498765432", "IE", "773498765432", roleHead},
}

func TestScenarioCompanies(t *testing.T) {
	r := newRouter()
	for _, sc := range scenarios {
		t.Run(sc.login, func(t *testing.T) {
			var cl bankClient
			if code := do(t, r, "GET", "/mock/core-banking/clients/"+sc.inn, "", &cl); code != http.StatusOK {
				t.Fatalf("clients: %d", code)
			}
			if cl.Type != sc.typ {
				t.Fatalf("type = %s, want %s", cl.Type, sc.typ)
			}
			i := slices.IndexFunc(cl.Persons, func(p bankPerson) bool { return p.INN == sc.personINN })
			if i < 0 || !slices.Contains(cl.Persons[i].Roles, sc.role) {
				t.Fatalf("person %s with role %s not found: %+v", sc.personINN, sc.role, cl.Persons)
			}
			if sc.role == roleRepresentative && cl.Persons[i].MCHDNumber == nil {
				t.Fatal("representative without mchdNumber")
			}

			var tr turnover
			if code := do(t, r, "GET", "/mock/core-banking/turnover/"+sc.inn, "", &tr); code != http.StatusOK || len(tr.Months) == 0 {
				t.Fatalf("turnover: %d %+v", code, tr)
			}
			var sum int64
			for _, m := range tr.Months {
				sum += m.Income
			}
			if tr.AvgMonthlyIncome != sum/int64(len(tr.Months)) {
				t.Fatalf("avgMonthlyIncome %d does not match months", tr.AvgMonthlyIncome)
			}

			for _, p := range []string{"/mock/transparent-business/", "/mock/fns/debts/", "/mock/courts/"} {
				if code := do(t, r, "GET", p+sc.inn, "", nil); code != http.StatusOK {
					t.Fatalf("%s: %d", p, code)
				}
			}
			if code := do(t, r, "POST", "/mock/bki/report", `{"inn":"`+sc.inn+`","consentId":"cons-1"}`, nil); code != http.StatusOK {
				t.Fatalf("bki: %d", code)
			}

			// ООО есть в ЕГРЮЛ и ГИР БО, ИП — только в ЕГРИП.
			want := map[string]int{"/mock/egrul/": http.StatusOK, "/mock/egrip/": http.StatusNotFound, "/mock/girbo/": http.StatusOK}
			if sc.typ == "IE" {
				want = map[string]int{"/mock/egrul/": http.StatusNotFound, "/mock/egrip/": http.StatusOK, "/mock/girbo/": http.StatusNotFound}
			}
			for p, code := range want {
				if got := do(t, r, "GET", p+sc.inn, "", nil); got != code {
					t.Fatalf("%s: %d, want %d", p, got, code)
				}
			}
		})
	}
}

// TestDecisionBranches проверяет признаки, которые ведут компанию в её ветку «Схемы решения».
func TestDecisionBranches(t *testing.T) {
	d := newDataset()
	co := func(inn string) *company { return d.companies[inn] }

	tests := []struct {
		name string
		ok   bool
	}{
		{"gromov: компания ликвидируется", co("7719456720").egrul.State == stateLiquidating},
		{"zaitsev: банкротство", co("7733567830").courts.Bankruptcy},
		{"belov: просрочка больше 90 дней", co("7728334410").bki.MaxOverdueDays > 90},
		{"orlova: регистрация после 2026-04-21", co("772512345601").egrip.RegDate > "2026-04-21"},
		{"sokolova: обороты меньше, чем у smirnov", co("7716678940").turnover.AvgMonthlyIncome < co("7712345678").turnover.AvgMonthlyIncome/3},
		{"morozov: долг больше половины годового оборота", co("7707789050").bki.TotalDebt*2 > co("7707789050").turnover.AvgMonthlyIncome*12},
		{"kozlov: обороты выше выручки ГИР БО больше чем на 30%", float64(co("7724890160").turnover.AvgMonthlyIncome*12) > 1.3*float64(co("7724890160").girbo.Revenue)},
		{"polyakov: стоп-причин нет", co("7731901270").egrul.State == stateActive && !co("7731901270").courts.Bankruptcy && co("7731901270").bki.MaxOverdueDays <= 30},
		{"smirnov: просрочек нет", co("7712345678").bki.MaxOverdueDays == 0 && !co("7712345678").fns.HasDebt},
	}
	for _, tt := range tests {
		if !tt.ok {
			t.Errorf("%s: условие не выполнено", tt.name)
		}
	}
}

func TestMCHD(t *testing.T) {
	r := newRouter()
	tests := []struct {
		number     string
		code       int
		valid      bool
		validUntil string
	}{
		{mchdPetrova, http.StatusOK, true, "2027-03-01"},
		{mchdVolkov, http.StatusOK, false, "2026-09-30"},
		{"unknown", http.StatusNotFound, false, ""},
	}
	for _, tt := range tests {
		var rec mchdRecord
		if code := do(t, r, "GET", "/mock/mchd/"+tt.number, "", &rec); code != tt.code {
			t.Fatalf("%s: %d, want %d", tt.number, code, tt.code)
		}
		if rec.Valid != tt.valid || rec.ValidUntil != tt.validUntil {
			t.Fatalf("%s: %+v", tt.number, rec)
		}
	}
}

func TestErrors(t *testing.T) {
	r := newRouter()
	tests := []struct {
		method, path, body string
		code               int
	}{
		{"GET", "/mock/core-banking/clients/0000000000", "", http.StatusNotFound},
		{"GET", "/mock/courts/0000000000", "", http.StatusNotFound},
		{"POST", "/mock/bki/report", `{"inn":"7712345678"}`, http.StatusBadRequest},
		{"POST", "/mock/bki/report", `{"inn":"0000000000","consentId":"cons-1"}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		if code := do(t, r, tt.method, tt.path, tt.body, nil); code != tt.code {
			t.Errorf("%s %s: %d, want %d", tt.method, tt.path, code, tt.code)
		}
	}
}
