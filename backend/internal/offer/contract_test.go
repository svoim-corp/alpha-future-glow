package offer

import (
	"net/http"
	"strings"
	"testing"
)

func TestGetContract(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved) // ООО с поручителем
	if w := e.do("GET", "/api/v1/applications/a-1/contract", "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("до принятия: %d", w.Code)
	}
	e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", "")

	w := e.do("GET", "/api/v1/applications/a-1/contract", "smirnov", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[contractResponse](t, w)
	if !strings.HasPrefix(got.Number, "КД-2026-") || got.Terms.Amount != 1_500_000 || got.Terms.Rate != 21.5 || got.Terms.MonthlyPayment == 0 ||
		got.Terms.FirstPaymentDate != "2026-11-21" || got.Signed || got.Disbursed {
		t.Fatalf("договор: %+v", got)
	}
	if len(got.Documents) != 2 || got.Documents[1].ID != docSurety || got.Documents[0].URL != "/api/v1/applications/a-1/contract/documents/d-1" {
		t.Fatalf("документы ООО: %+v", got.Documents)
	}

	e.decide(t, "a-2", "c-3", approved) // ИП: поручителя нет
	e.do("POST", "/api/v1/applications/a-2/offer/accept", "orlova", "")
	if got := decode[contractResponse](t, e.do("GET", "/api/v1/applications/a-2/contract", "orlova", "")); len(got.Documents) != 1 {
		t.Fatalf("документы ИП: %+v", got.Documents)
	}
	if w := e.do("GET", "/api/v1/applications/a-2/contract", "smirnov", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужой договор: %d", w.Code)
	}
}
