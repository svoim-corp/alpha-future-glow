package origination

import (
	"net/http"
	"testing"
)

const consentBody = `{"given":true,"text":"Согласен на запрос моей кредитной истории в бюро кредитных историй"}`

func TestConsentBKI(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID

	w := e.do("POST", "/api/v1/applications/"+id+"/consents/bki", "smirnov", consentBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if got := decode[consentResponse](t, w); !got.Given || got.By != "u-1" || !got.At.Equal(today) {
		t.Fatalf("согласие: %+v", got)
	}
	// Повтор от другого пользователя той же компании — действует первое согласие.
	if got := decode[consentResponse](t, e.do("POST", "/api/v1/applications/"+id+"/consents/bki", "smirnov", consentBody)); got.By != "u-1" {
		t.Fatalf("повтор: %+v", got)
	}
	a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", ""))
	if a.Consents.BKI == nil || !a.Consents.BKI.Given {
		t.Fatalf("согласие в заявке: %+v", a.Consents)
	}
}

func TestConsentBKIErrors(t *testing.T) {
	e := newEnv(t)
	byMCHD := e.draft(t, "lebedev").ID // МЧД без полномочия BKI_CONSENT
	own := e.draft(t, "smirnov").ID
	tests := []struct {
		name, id, user, body string
		status               int
	}{
		{"нет полномочия по МЧД", byMCHD, "lebedev", consentBody, http.StatusForbidden},
		{"отказ от согласия", own, "smirnov", `{"given":false,"text":"нет"}`, http.StatusBadRequest},
		{"без текста", own, "smirnov", `{"given":true}`, http.StatusBadRequest},
		{"чужая заявка", own, "orlova", consentBody, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := e.do("POST", "/api/v1/applications/"+tt.id+"/consents/bki", tt.user, tt.body); w.Code != tt.status {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestGuarantorConsent(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	path := "/api/v1/applications/" + id + "/guarantor/consent-bki"

	if w := e.do("POST", path, "smirnov", `{"given":true}`); w.Code != http.StatusConflict {
		t.Fatalf("до этапа 2: %d", w.Code)
	}
	e.passPrelim(t, id)
	if w := e.do("POST", path, "smirnov", `{"given":true}`); w.Code != http.StatusBadRequest {
		t.Fatalf("поручитель не выбран: %d", w.Code)
	}
	if st, _, _ := e.patch(id, "smirnov", "1", `{"stage2":{"guarantor":{"personId":"p-1"}}}`); st != http.StatusOK {
		t.Fatalf("выбор поручителя: %d", st)
	}
	w := e.do("POST", path, "smirnov", `{"given":true}`)
	if got := decode[guarantorConsentResponse](t, w); w.Code != http.StatusCreated || !got.Given || !got.At.Equal(today) {
		t.Fatalf("согласие: %d %s", w.Code, w.Body.String())
	}
	if a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", "")); !a.Stage2.Guarantor.BKIConsent {
		t.Fatalf("согласие поручителя в заявке: %+v", a.Stage2.Guarantor)
	}

	// Другой поручитель — согласие прежнего не действует.
	if st, _, _ := e.patch(id, "smirnov", "2", `{"stage2":{"guarantor":{"personId":"p-3"}}}`); st != http.StatusOK {
		t.Fatalf("смена поручителя: %d", st)
	}
	if a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", "")); a.Stage2.Guarantor.PersonID != "p-3" || a.Stage2.Guarantor.BKIConsent {
		t.Fatalf("после смены поручителя: %+v", a.Stage2.Guarantor)
	}
	// Тот же поручитель повторно — согласие сохраняется.
	if w := e.do("POST", path, "smirnov", `{"given":true}`); w.Code != http.StatusCreated {
		t.Fatal(w.Body.String())
	}
	if st, _, _ := e.patch(id, "smirnov", "3", `{"stage2":{"guarantor":{"personId":"p-3"}}}`); st != http.StatusOK {
		t.Fatalf("тот же поручитель: %d", st)
	}
	if a := decode[applicationDTO](t, e.do("GET", "/api/v1/applications/"+id, "smirnov", "")); !a.Stage2.Guarantor.BKIConsent {
		t.Fatalf("согласие того же поручителя не сбрасывается: %+v", a.Stage2.Guarantor)
	}
}
