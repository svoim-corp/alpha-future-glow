package origination

import (
	"slices"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestCompleteness(t *testing.T) {
	llc := newFakeCompanies().profiles["c-1"]
	ie := newFakeCompanies().profiles["c-3"]
	limit := &shared.PrelimView{Outcome: shared.PrelimLimit, AmountFits: true}
	consent := &consentRow{ID: "cs-1"}
	agreed := &consentRow{ID: "cs-2", PersonID: ptr("p-1")}

	// complete — заявка ООО, у которой есть всё.
	complete := func() *dossier {
		return &dossier{
			app:     &appRow{Goal: ptr("WORKING_CAPITAL"), GuarantorPersonID: ptr("p-1"), Status: string(shared.StatusDraft)},
			product: &working, profile: llc, stage: stageState{stage: stageFull, prelim: limit},
			consent: consent, guarantorConsent: agreed,
		}
	}
	tests := []struct {
		name  string
		edit  func(*dossier)
		truth bool
		want  []string
	}{
		{"всё есть", func(*dossier) {}, true, nil},
		{"не подтверждена достоверность", func(*dossier) {}, false, []string{"stage2.truthConfirmed"}},
		{"нет цели", func(d *dossier) { d.app.Goal = nil }, true, []string{"params.goal"}},
		{"этап 1 не пройден", func(d *dossier) { d.stage = stageState{stage: stagePrelim} }, true, []string{"prelim"}},
		{"этап 1 устарел", func(d *dossier) { d.stage = stageState{stage: stagePrelim, invalidated: true} }, true, []string{"prelim"}},
		{"стоп-отказ", func(d *dossier) {
			d.stage = stageState{stage: stagePrelim, prelim: &shared.PrelimView{Outcome: shared.PrelimStop}}
		}, true, []string{"prelim"}},
		{"сумма выше лимита без выбора", func(d *dossier) {
			d.stage = stageState{stage: stagePrelim, prelim: &shared.PrelimView{Outcome: shared.PrelimLimit}}
		}, true, []string{"prelim.choice"}},
		{"нет согласия на БКИ", func(d *dossier) { d.consent = nil }, true, []string{"consents.bki"}},
		{"нет поручителя", func(d *dossier) { d.app.GuarantorPersonID = nil }, true, []string{"stage2.guarantor.personId"}},
		{"поручитель больше не кандидат", func(d *dossier) { d.app.GuarantorPersonID = ptr("p-9") }, true, []string{"stage2.guarantor.personId"}},
		{"нет согласия поручителя", func(d *dossier) { d.guarantorConsent = nil }, true, []string{"stage2.guarantor.bkiConsent"}},
		{"согласие дал прежний поручитель", func(d *dossier) { d.app.GuarantorPersonID = ptr("p-3") }, true, []string{"stage2.guarantor.bkiConsent"}},
		{"ИП — поручитель не нужен", func(d *dossier) { d.profile, d.app.GuarantorPersonID, d.guarantorConsent = ie, nil, nil }, true, nil},
		{"продукт без поручителя", func(d *dossier) { d.product, d.app.GuarantorPersonID = &express, nil }, true, nil},
		{"пустая заявка", func(d *dossier) {
			d.app.Goal, d.app.GuarantorPersonID, d.stage, d.consent, d.guarantorConsent = nil, nil, stageState{stage: stagePrelim}, nil, nil
		}, false, []string{"params.goal", "prelim", "consents.bki", "stage2.guarantor.personId", "stage2.truthConfirmed"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := complete()
			tt.edit(d)
			var got []string
			for _, f := range completeness(d, tt.truth) {
				got = append(got, f.Field)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
