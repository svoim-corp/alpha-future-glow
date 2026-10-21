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
	byMCHD := e.draft(t, "petrova").ID // МЧД без полномочия BKI_CONSENT
	own := e.draft(t, "smirnov").ID
	tests := []struct {
		name, id, user, body string
		status               int
	}{
		{"нет полномочия по МЧД", byMCHD, "petrova", consentBody, http.StatusForbidden},
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
