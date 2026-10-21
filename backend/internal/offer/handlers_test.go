package offer

import (
	"net/http"
	"strings"
	"testing"

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
