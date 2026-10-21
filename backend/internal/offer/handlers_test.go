package offer

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestGetOffer(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", shared.DecisionView{Result: shared.StatusCounterOffer, Mode: shared.DecisionManual, ProductCode: "WORKING",
		Amount: 1_200_000, Term: 24, Rate: 21.5, Reason: "Сумма снижена по уровню долговой нагрузки"})
	w := e.do("GET", "/api/v1/applications/a-1/offer", "smirnov", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	got := decode[offerResponse](t, w)
	if got.Decision != shared.StatusCounterOffer || got.Requested.Amount != 1_500_000 || got.Offered.Amount != 1_200_000 ||
		got.Offered.MonthlyPayment != 61_958 || *got.ValidUntil != "2026-10-28" || *got.Explanation != "Сумма снижена по уровню долговой нагрузки" ||
		got.Reason != nil {
		t.Fatalf("иные условия: %+v", got)
	}

	e.decide(t, "a-2", "c-1", shared.DecisionView{Result: shared.StatusRejected, Mode: shared.DecisionAuto, Reason: "Автоматический отказ: балл 220",
		Alternatives: []shared.Alternative{{ProductCode: "EXPRESS", MaxAmount: 600_000}}})
	w = e.do("GET", "/api/v1/applications/a-2/offer", "smirnov", "")
	rej := decode[offerResponse](t, w)
	if rej.Decision != shared.StatusRejected || *rej.Reason != autoRejectText || len(rej.Alternatives) != 1 || rej.Offered != nil ||
		strings.Contains(w.Body.String(), "балл") {
		t.Fatalf("отказ: %s", w.Body.String())
	}

	if w := e.do("GET", "/api/v1/applications/a-1/offer", "orlova", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужая заявка: %d", w.Code)
	}
	e.apps.apps["a-3"] = &shared.ApplicationView{ID: "a-3", CompanyID: "c-1", Status: shared.StatusOnDecision}
	if w := e.do("GET", "/api/v1/applications/a-3/offer", "smirnov", ""); w.Code != http.StatusNotFound {
		t.Fatalf("до решения: %d", w.Code)
	}
}

func TestAcceptOffer(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)

	w := e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", "")
	if got := decode[statusResponse](t, w); w.Code != http.StatusOK || got.Status != shared.StatusSigning || got.Redirect != "/applications/a-1/sign" {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if e.apps.status("a-1") != shared.StatusSigning || len(e.apps.history) != 1 || e.apps.history[0] != "a-1:SIGNING:OFFER_ACCEPT:" {
		t.Fatalf("переход: %v", e.apps.history)
	}
	o, _ := e.m.store.Offer(context.Background(), "a-1")
	if o.AcceptedAt == nil || o.ContractNumber == nil || !strings.HasPrefix(*o.ContractNumber, "КД-2026-") {
		t.Fatalf("договор: %+v", o)
	}
	if w := e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("повторное принятие: %d", w.Code)
	}
}

func TestAcceptOfferErrors(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", shared.DecisionView{Result: shared.StatusRejected, Mode: shared.DecisionAuto})
	if w := e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("отказ: %d", w.Code)
	}

	old := approved
	old.DecidedAt = today.AddDate(0, 0, -8)
	e.decide(t, "a-2", "c-1", old)
	if w := e.do("POST", "/api/v1/applications/a-2/offer/accept", "smirnov", ""); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("истёкшее предложение: %d", w.Code)
	}
	lastDay := approved
	lastDay.DecidedAt = today.AddDate(0, 0, -7).Add(-time.Hour)
	e.decide(t, "a-3", "c-1", lastDay)
	if w := e.do("POST", "/api/v1/applications/a-3/offer/accept", "smirnov", ""); w.Code != http.StatusOK {
		t.Fatalf("последний день действия: %d %s", w.Code, w.Body.String())
	}

	// Переход не удался — принятие откатывается.
	e.decide(t, "a-4", "c-1", approved)
	e.apps.apps["a-4"].Status = shared.StatusClosed
	if w := e.do("POST", "/api/v1/applications/a-4/offer/accept", "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("заявка закрыта: %d", w.Code)
	}
	if o, _ := e.m.store.Offer(context.Background(), "a-4"); o.AcceptedAt != nil || o.ContractNumber != nil {
		t.Fatalf("принятие не откатилось: %+v", o)
	}
}

func TestDeclineOffer(t *testing.T) {
	e := newEnv(t)
	e.decide(t, "a-1", "c-1", approved)
	if w := e.do("POST", "/api/v1/applications/a-1/offer/decline", "smirnov", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("без причины: %d", w.Code)
	}
	w := e.do("POST", "/api/v1/applications/a-1/offer/decline", "smirnov", `{"reason":"Нашли финансирование в другом месте"}`)
	if got := decode[statusResponse](t, w); w.Code != http.StatusOK || got.Status != shared.StatusClosed || strings.Contains(w.Body.String(), "redirect") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if e.apps.history[0] != "a-1:CLOSED:OFFER_DECLINE:Нашли финансирование в другом месте" {
		t.Fatalf("история: %v", e.apps.history)
	}
	if w := e.do("POST", "/api/v1/applications/a-1/offer/accept", "smirnov", ""); w.Code != http.StatusConflict {
		t.Fatalf("принять после отказа: %d", w.Code)
	}

	e.decide(t, "a-2", "c-1", approved)
	e.do("POST", "/api/v1/applications/a-2/offer/accept", "smirnov", "")
	if w := e.do("POST", "/api/v1/applications/a-2/offer/decline", "smirnov", `{"reason":"передумали"}`); w.Code != http.StatusConflict {
		t.Fatalf("отказ после принятия: %d", w.Code)
	}
}
