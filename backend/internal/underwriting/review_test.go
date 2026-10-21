package underwriting

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func codes(is []issue) []string {
	out := make([]string, len(is))
	for i, x := range is {
		out[i] = string(x.Severity) + ":" + x.Code
	}
	return out
}

func TestCompleteness(t *testing.T) {
	guarantor := "p-1"
	stranger := "p-404"
	keep := shared.ChoiceKeepAmount
	until := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
	llc := &shared.CompanyProfile{
		CompanyView:         shared.CompanyView{Type: shared.CompanyLLC},
		Turnover:            shared.Turnover{AvgMonthlyIncome: 2_000_000},
		GuarantorCandidates: []shared.Person{{PersonID: "p-1"}},
	}
	ie := &shared.CompanyProfile{CompanyView: shared.CompanyView{Type: shared.CompanyIE}, Turnover: shared.Turnover{AvgMonthlyIncome: 900_000}}
	consent := &shared.ConsentView{ID: "c"}
	girbo := func(revenue int64) *shared.CheckReport {
		return &shared.CheckReport{Facts: shared.CheckFacts{GIRBO: &shared.GIRBOReport{Revenue: revenue}}}
	}

	tests := []struct {
		name     string
		d        dossier
		complete bool
		codes    []string
	}{
		{"всё на месте", dossier{app: &shared.ApplicationView{GuarantorPersonID: &guarantor}, company: llc, product: working,
			consent: consent, guarantor: &llc.GuarantorCandidates[0], guarantorBKI: consent, checks: girbo(23_000_000)}, true, []string{}},
		{"ИП без поручителя", dossier{app: &shared.ApplicationView{}, company: ie, product: working, consent: consent}, true, []string{}},
		{"нет согласия заявителя", dossier{app: &shared.ApplicationView{}, company: ie, product: working}, false, []string{"BLOCKER:BKI_CONSENT_MISSING"}},
		{"ООО без поручителя", dossier{app: &shared.ApplicationView{}, company: llc, product: working, consent: consent}, false, []string{"BLOCKER:GUARANTOR_MISSING"}},
		{"продукт без поручителя", dossier{app: &shared.ApplicationView{}, company: llc, product: express, consent: consent}, true, []string{}},
		{"поручитель не из кандидатов", dossier{app: &shared.ApplicationView{GuarantorPersonID: &stranger}, company: llc, product: working, consent: consent}, false, []string{"BLOCKER:GUARANTOR_UNKNOWN"}},
		{"нет согласия поручителя", dossier{app: &shared.ApplicationView{GuarantorPersonID: &guarantor}, company: llc, product: working,
			consent: consent, guarantor: &llc.GuarantorCandidates[0]}, false, []string{"BLOCKER:GUARANTOR_CONSENT_MISSING"}},
		{"kozlov: обороты выше выручки", dossier{app: &shared.ApplicationView{}, company: ie, product: working, consent: consent, checks: girbo(6_000_000)},
			true, []string{"WARN:TURNOVER_MISMATCH"}},
		{"источник не ответил и проверка провалена", dossier{app: &shared.ApplicationView{}, company: ie, product: working, consent: consent,
			checks: &shared.CheckReport{Results: []shared.CheckResult{{Source: shared.SourceBKI, Status: shared.CheckUnavailable}, {Source: shared.SourceCourts, Status: shared.CheckFail}}}},
			true, []string{"WARN:SOURCE_UNAVAILABLE", "WARN:CHECK_FAILED"}},
		{"сумма выше лимита и МЧД", dossier{app: &shared.ApplicationView{Authority: shared.AuthorityView{Type: shared.AuthorityMCHD, ValidUntil: &until}},
			company: ie, product: working, consent: consent, prelim: &shared.PrelimView{Limit: 900_000, Choice: &keep}},
			true, []string{"INFO:AMOUNT_ABOVE_LIMIT", "INFO:MCHD_APPLICANT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete, issues := completeness(&tt.d)
			if got := codes(issues); complete != tt.complete || !slices.Equal(got, tt.codes) {
				t.Fatalf("complete %v %v", complete, got)
			}
		})
	}
	_, issues := completeness(&dossier{app: &shared.ApplicationView{}, company: ie, product: working, consent: consent, checks: girbo(6_000_000)})
	if issues[0].Text != "Обороты по счетам в банке выше выручки в ГИР БО на 80%" {
		t.Fatalf("%q", issues[0].Text)
	}
}

func TestReview(t *testing.T) {
	e := newEnv(t, false)
	guarantor := "p-1"
	head := shared.Person{PersonID: "p-1", FullName: "Смирнов Олег Викторович", INN: "771234567890"}
	bakery.Head, bakery.GuarantorCandidates = head, []shared.Person{head}
	bakery.Turnover = shared.Turnover{AvgMonthlyIncome: 1_840_000, PeriodMonths: 12}
	bakery.RegDate = time.Date(2015, 3, 12, 0, 0, 0, 0, time.UTC)

	a := submittedApp("a-1", "c-1", "WORKING", 1_500_000, today)
	a.Status, a.GuarantorPersonID = shared.StatusOnDecision, &guarantor
	a.Authority = shared.AuthorityView{Type: shared.AuthorityHead}
	e.apps.add(a)
	e.apps.add(submittedApp("a-draft", "c-1", "WORKING", 1_500_000, today)) // статус DRAFT
	e.apps.consents["a-1/APPLICANT"] = &shared.ConsentView{ID: "cons-1"}
	e.apps.consents["a-1/GUARANTOR"] = &shared.ConsentView{ID: "cons-2"}
	e.prelims["a-1"] = &shared.PrelimView{Outcome: shared.PrelimLimit, Limit: 3_300_000, PrelimRate: 20.5, AmountFits: true}
	e.checks.reports["a-1"] = &shared.CheckReport{
		Results: []shared.CheckResult{{Source: shared.SourceGIRBO, Status: shared.CheckOK, Summary: "Отчётность за 2025 год учтена"}},
		Facts:   shared.CheckFacts{GIRBO: &shared.GIRBOReport{Year: 2025, Revenue: 21_400_000, NetProfit: 2_100_000}},
	}

	for _, path := range []string{"/bank/applications/a-draft", "/bank/applications/a-404", "/bank/applications/a-draft/checks"} {
		if w := e.do("GET", path, "ivanova", ""); w.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	if w := e.do("GET", "/bank/applications/a-1", "smirnov", ""); w.Code != http.StatusForbidden {
		t.Errorf("клиент в разборе: %d", w.Code)
	}

	var r reviewResponse
	w := e.do("GET", "/bank/applications/a-1", "ivanova", "")
	_ = json.Unmarshal(w.Body.Bytes(), &r)
	if w.Code != http.StatusOK || r.Applicant.FullName != head.FullName || r.Company.RegDate != "2015-03-12" || r.GIRBO == nil ||
		r.GIRBO.Revenue != 21_400_000 || r.Guarantor == nil || !r.Guarantor.BKIConsent || r.Prelim == nil || r.Prelim.Limit != 3_300_000 ||
		r.Turnover.AvgMonthlyIncome != 1_840_000 {
		t.Fatalf("анкета: %d %s", w.Code, w.Body)
	}

	var c completenessResponse
	w = e.do("GET", "/bank/applications/a-1/completeness", "ivanova", "")
	_ = json.Unmarshal(w.Body.Bytes(), &c)
	if w.Code != http.StatusOK || !c.Complete || c.Issues == nil {
		t.Fatalf("полнота: %d %s", w.Code, w.Body)
	}

	if w := e.do("POST", "/bank/applications/a-1/checks/refresh", "ivanova", `{"sources":["BKI","SPACE"]}`); w.Code != http.StatusBadRequest {
		t.Fatalf("неизвестный источник: %d %s", w.Code, w.Body)
	}
	var checks []checkDTO
	w = e.do("POST", "/bank/applications/a-1/checks/refresh", "ivanova", `{"sources":["BKI","COURTS"]}`)
	_ = json.Unmarshal(w.Body.Bytes(), &checks)
	if w.Code != http.StatusOK || len(checks) != 3 || e.checks.last.ConsentID != "cons-1" || e.checks.last.INN != "7712345678" {
		t.Fatalf("перезапрос: %d %s %+v", w.Code, w.Body, e.checks.last)
	}
	if w := e.do("POST", "/bank/applications/a-1/checks/refresh", "ivanova", ""); w.Code != http.StatusOK || len(e.checks.last.Sources) != 0 {
		t.Fatalf("перезапрос всех: %d %+v", w.Code, e.checks.last)
	}
	w = e.do("GET", "/bank/applications/a-1/checks", "ivanova", "")
	checks = nil
	_ = json.Unmarshal(w.Body.Bytes(), &checks)
	if w.Code != http.StatusOK || len(checks) != 3 || checks[0].Source != shared.SourceGIRBO {
		t.Fatalf("проверки: %d %s", w.Code, w.Body)
	}
}
