package offer

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
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

func TestGetDocument(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)
	if w := e.do("GET", "/api/v1/applications/a-1/contract/documents/d-1", "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("до принятия: %d", w.Code)
	}
	e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", "")

	w := e.do("GET", "/api/v1/applications/a-1/contract/documents/d-1", "smirnov", "")
	body := w.Body.String()
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "text/html; charset=utf-8" || w.Header().Get("Content-Disposition") != "inline" {
		t.Fatalf("%d %v", w.Code, w.Header())
	}
	for _, want := range []string{"Кредитный договор", "ООО «Северная пекарня»", "7712345678", "1 500 000 ₽", "21.5% годовых", "21.11.2026"} {
		if !strings.Contains(body, want) {
			t.Errorf("в договоре нет %q", want)
		}
	}
	surety := e.do("GET", "/api/v1/applications/a-1/contract/documents/d-2", "smirnov", "").Body.String()
	if !strings.Contains(surety, "Договор поручительства") || !strings.Contains(surety, "Смирнов Олег Викторович") {
		t.Fatalf("договор поручительства: %s", surety)
	}
	if w := e.do("GET", "/api/v1/applications/a-1/contract/documents/d-9", "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("нет документа: %d", w.Code)
	}

	e.decide(t, "a-2", "c-3", approved)
	e.do("POST", "/api/v1/applications/a-2/offer/accept", "orlova", "")
	if w := e.do("GET", "/api/v1/applications/a-2/contract/documents/d-2", "orlova", ""); w.Code != http.StatusNotFound {
		t.Fatalf("поручительство у ИП: %d", w.Code)
	}
}

func TestSignContract(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)
	path := "/api/v1/applications/a-1/contract/sign"
	if w := e.do("POST", path, "smirnov", `{"confirm":true}`); w.Code != http.StatusNotFound {
		t.Fatalf("до принятия: %d", w.Code)
	}
	e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", "")
	if w := e.do("POST", path, "smirnov", `{"confirm":false}`); w.Code != http.StatusBadRequest {
		t.Fatalf("без подтверждения: %d", w.Code)
	}

	key := fmt.Sprintf("sign-%d", time.Now().UnixNano()) // ключи хранятся между прогонами
	w := e.do("POST", path, "smirnov", `{"confirm":true}`, "Idempotency-Key", key)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[signResponse](t, w)
	if !got.Signed || !got.SignedAt.Equal(today) || !regexp.MustCompile(`^SIM-[0-9A-F]{4}$`).MatchString(got.SignatureID) {
		t.Fatalf("подпись: %+v", got)
	}
	if e.apps.status("a-1") != shared.StatusSigning {
		t.Fatalf("подпись не меняет статус: %s", e.apps.status("a-1"))
	}
	if again := e.do("POST", path, "smirnov", `{"confirm":true}`, "Idempotency-Key", key); again.Body.String() != w.Body.String() {
		t.Fatalf("повтор по ключу: %s", again.Body.String())
	}
	if w := e.do("POST", path, "smirnov", `{"confirm":true}`); w.Code != http.StatusConflict {
		t.Fatalf("повторная подпись: %d", w.Code)
	}
	if k := decode[contractResponse](t, e.do("GET", "/api/v1/applications/a-1/contract", "smirnov", "")); !k.Signed {
		t.Fatalf("договор подписан: %+v", k)
	}
	if doc := e.do("GET", "/api/v1/applications/a-1/contract/documents/d-1", "smirnov", "").Body.String(); !strings.Contains(doc, got.SignatureID) {
		t.Fatal("в печатной форме нет отметки о подписи")
	}
}

func TestDisburse(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)
	path := "/api/v1/applications/a-1/contract/disburse"
	e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", "")
	if w := e.do("POST", path, "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("без подписи: %d", w.Code)
	}
	e.do("POST", "/api/v1/applications/a-1/contract/sign", "smirnov", `{"confirm":true}`)

	key := fmt.Sprintf("disburse-%d", time.Now().UnixNano())
	w := e.do("POST", path, "smirnov", "", "Idempotency-Key", key)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[disburseResponse](t, w)
	if got.Status != shared.StatusIssued || got.LoanID != "l-a-1" || !got.DisbursedAt.Equal(today) || got.AccountMasked != "•••• 5678" {
		t.Fatalf("выдача: %+v", got)
	}
	in := e.loans.opened["a-1"]
	if in.Amount != 1_500_000 || in.Rate != 21.5 || in.Term != 24 || in.ApplicantID != "u-1" || !strings.HasPrefix(in.ContractNumber, "КД-") {
		t.Fatalf("условия кредита: %+v", in)
	}
	if e.apps.status("a-1") != shared.StatusIssued {
		t.Fatalf("статус: %s", e.apps.status("a-1"))
	}
	if again := e.do("POST", path, "smirnov", "", "Idempotency-Key", key); again.Body.String() != w.Body.String() {
		t.Fatalf("повтор по ключу: %s", again.Body.String())
	}
	if w := e.do("POST", path, "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("повторная выдача: %d", w.Code)
	}
	if k := decode[contractResponse](t, e.do("GET", "/api/v1/applications/a-1/contract", "smirnov", "")); !k.Disbursed || k.Terms.FirstPaymentDate != "2026-11-21" {
		t.Fatalf("договор после выдачи: %+v", k)
	}
}
