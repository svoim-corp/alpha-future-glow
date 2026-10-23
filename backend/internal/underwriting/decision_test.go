package underwriting

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func fieldsOf(t *testing.T, body []byte) []string {
	t.Helper()
	var b httpx.ErrorBody
	_ = json.Unmarshal(body, &b)
	out := make([]string, len(b.Fields))
	for i, f := range b.Fields {
		out[i] = f.Field
	}
	return out
}

func TestAnswerErrors(t *testing.T) {
	items := []infoItem{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}}
	tests := []struct {
		name    string
		answers []infoAnswer
		fields  []string
	}{
		{"все ответы", []infoAnswer{{"a", "1"}, {"b", "2"}}, nil},
		{"пропущен вопрос", []infoAnswer{{"a", "1"}}, []string{"answers.b"}},
		{"лишний ключ", []infoAnswer{{"a", "1"}, {"b", "2"}, {"c", "3"}}, []string{"answers.c"}},
		{"повтор ключа", []infoAnswer{{"a", "1"}, {"a", "2"}, {"b", "3"}}, []string{"answers.a"}},
	}
	for _, tt := range tests {
		var got []string
		for _, f := range answerErrors(items, tt.answers) {
			got = append(got, f.Field)
		}
		if !slices.Equal(got, tt.fields) {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.fields)
		}
	}
}

// TestDecisionValidation — тело решения проверяется до обращения к заявке и БД.
func TestDecisionValidation(t *testing.T) {
	e := newEnv(t, false)
	tests := []struct {
		body   string
		fields []string
	}{
		{`{}`, []string{"type", "rate"}}, // ставка нужна для всего, кроме отказа
		{`{"type":"MAYBE"}`, []string{"type", "rate"}},
		{`{"type":"APPROVE"}`, []string{"rate"}},
		{`{"type":"COUNTER","rate":21}`, []string{"amount", "term", "productCode"}},
		{`{"type":"REJECT"}`, []string{"reason"}},
	}
	for _, tt := range tests {
		w := e.do("POST", "/bank/applications/a-1/decision", "ivanova", tt.body)
		if got := fieldsOf(t, w.Body.Bytes()); w.Code != http.StatusBadRequest || !slices.Equal(got, tt.fields) {
			t.Errorf("%s: %d %v", tt.body, w.Code, got)
		}
	}
	for _, body := range []string{`{"comment":"x","items":[]}`, `{"comment":"x","items":[{"key":"k","type":"FILE","label":"l"}]}`} {
		if w := e.do("POST", "/bank/applications/a-1/info-requests", "ivanova", body); w.Code != http.StatusBadRequest {
			t.Errorf("запрос информации %s: %d", body, w.Code)
		}
	}
}

func TestInfoRequestsAndDecision(t *testing.T) {
	e := newEnv(t, true)
	a := submittedApp("a-1", "c-1", "WORKING", 1_500_000, today)
	a.Status = shared.StatusOnDecision
	e.apps.add(a)
	b := submittedApp("a-2", "c-1", "WORKING", 1_000_000, today)
	b.Status = shared.StatusSubmitted
	e.apps.add(b)

	items := `{"comment":"Обороты по счетам расходятся с отчётностью","items":[` +
		`{"key":"turnover_note","type":"TEXT","label":"Почему выручка ниже оборотов"},{"key":"purpose","type":"TEXT","label":"У кого закупка"}]}`

	// Запрос: дубли ключей — 400; неподходящий статус — 409 и запроса не остаётся; затем успешный запрос и повтор — 409.
	if w := e.do("POST", "/bank/applications/a-1/info-requests", "ivanova",
		`{"comment":"x","items":[{"key":"k","type":"TEXT","label":"1"},{"key":"k","type":"TEXT","label":"2"}]}`); w.Code != http.StatusBadRequest {
		t.Fatalf("дубли ключей: %d", w.Code)
	}
	if w := e.do("POST", "/bank/applications/a-2/info-requests", "ivanova", items); w.Code != http.StatusConflict {
		t.Fatalf("из SUBMITTED: %d %s", w.Code, w.Body)
	}
	if w := e.do("GET", "/applications/a-2/info-requests", "smirnov", ""); w.Body.String() != "[]" {
		t.Fatalf("запрос после неудачного перехода остался: %s", w.Body)
	}
	var created infoRequestCreated
	w := e.do("POST", "/bank/applications/a-1/info-requests", "ivanova", items)
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if w.Code != http.StatusCreated || created.Status != requestOpen || e.apps.statusOf("a-1") != shared.StatusInfoRequested {
		t.Fatalf("запрос: %d %s", w.Code, w.Body)
	}
	if got := e.notifier.types(); !slices.Equal(got, []shared.NotificationType{shared.NotifyInfoRequested}) || e.notifier.sent[0].userID != "u-1" {
		t.Fatalf("уведомление: %+v", e.notifier.sent)
	}
	e.apps.setStatusForTest("a-1", shared.StatusOnDecision) // второй запрос упрётся в открытый первый
	if w := e.do("POST", "/bank/applications/a-1/info-requests", "ivanova", items); w.Code != http.StatusConflict {
		t.Fatalf("второй открытый запрос: %d", w.Code)
	}
	e.apps.setStatusForTest("a-1", shared.StatusInfoRequested)

	// Клиент: видит только свою заявку.
	if w := e.do("GET", "/applications/a-1/info-requests", "morozov", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужая компания: %d", w.Code)
	}
	var list []infoRequestDTO
	w = e.do("GET", "/applications/a-1/info-requests", "smirnov", "")
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if w.Code != http.StatusOK || len(list) != 1 || list[0].ID != created.ID || len(list[0].Items) != 2 || list[0].Answers != nil {
		t.Fatalf("список: %d %s", w.Code, w.Body)
	}

	answer := "/applications/a-1/info-requests/" + created.ID + "/answer"
	if w := e.do("POST", answer, "smirnov", `{"answers":[{"key":"turnover_note","value":"Возврат займа"}]}`); w.Code != http.StatusBadRequest ||
		!slices.Equal(fieldsOf(t, w.Body.Bytes()), []string{"answers.purpose"}) {
		t.Fatalf("неполный ответ: %d %s", w.Code, w.Body)
	}
	if w := e.do("POST", "/applications/a-1/info-requests/r-404/answer", "smirnov", `{"answers":[{"key":"a","value":"b"}]}`); w.Code != http.StatusNotFound {
		t.Fatalf("нет запроса: %d", w.Code)
	}
	full := `{"answers":[{"key":"turnover_note","value":"Возврат займа"},{"key":"purpose","value":"Мелькомбинат «Заря»"}]}`
	var ans answerResponse
	w = e.do("POST", answer, "smirnov", full)
	_ = json.Unmarshal(w.Body.Bytes(), &ans)
	if w.Code != http.StatusOK || ans.ApplicationStatus != shared.StatusDataCheck || e.apps.statusOf("a-1") != shared.StatusDataCheck {
		t.Fatalf("ответ: %d %s", w.Code, w.Body)
	}
	if w := e.do("POST", answer, "smirnov", full); w.Code != http.StatusConflict {
		t.Fatalf("повторный ответ: %d", w.Code)
	}
	list = nil
	w = e.do("GET", "/applications/a-1/info-requests", "smirnov", "")
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if list[0].Status != requestAnswered || len(list[0].Answers) != 2 || list[0].AnsweredAt == nil {
		t.Fatalf("после ответа: %s", w.Body)
	}

	// Сотрудник видит те же запросы с ответами клиента; клиентской ручкой пользоваться не может.
	var bankList []infoRequestDTO
	w = e.do("GET", "/bank/applications/a-1/info-requests", "ivanova", "")
	_ = json.Unmarshal(w.Body.Bytes(), &bankList)
	if w.Code != http.StatusOK || len(bankList) != 1 || bankList[0].ID != created.ID || len(bankList[0].Answers) != 2 ||
		bankList[0].Answers[0].Value != "Возврат займа" {
		t.Fatalf("запросы у сотрудника: %d %s", w.Code, w.Body)
	}
	if w := e.do("GET", "/bank/applications/a-1/info-requests", "smirnov", ""); w.Code != http.StatusForbidden {
		t.Fatalf("клиент в банковской ручке: %d", w.Code)
	}
	if w := e.do("GET", "/bank/applications/a-404/info-requests", "ivanova", ""); w.Code != http.StatusNotFound {
		t.Fatalf("нет заявки: %d", w.Code)
	}

	// Решение: одобрение без параметров — на условиях заявки; решивший становится исполнителем; клиенту — уведомление.
	e.apps.setStatusForTest("a-1", shared.StatusOnDecision)
	var dr decisionResponse
	w = e.do("POST", "/bank/applications/a-1/decision", "ivanova", `{"type":"APPROVE","rate":20.5}`)
	_ = json.Unmarshal(w.Body.Bytes(), &dr)
	if w.Code != http.StatusOK || dr.Status != shared.StatusApproved || dr.DecisionID != "dec-1" || *dr.DecidedBy != "u-9" {
		t.Fatalf("решение: %d %s", w.Code, w.Body)
	}
	if in := e.decider.last; in.ProductCode != "WORKING" || in.Amount != 1_500_000 || in.Term != 24 || in.Rate != 20.5 {
		t.Fatalf("параметры одобрения: %+v", in)
	}
	if item, _ := e.m.store.QueueItem(t.Context(), "a-1"); item == nil || item.AssigneeID == nil || *item.AssigneeID != "u-9" {
		t.Fatalf("исполнитель после решения: %+v", item)
	}
	if got := e.notifier.types(); !slices.Equal(got, []shared.NotificationType{shared.NotifyInfoRequested, shared.NotifyDecision}) {
		t.Fatalf("уведомления: %v", got)
	}

	// Ошибка Scoring (например, 422 вне рамок продукта) уходит клиенту как есть.
	e.apps.setStatusForTest("a-1", shared.StatusOnDecision)
	e.decider.err = httpx.BusinessRule("Сумма вне рамок продукта")
	if w := e.do("POST", "/bank/applications/a-1/decision", "ivanova",
		`{"type":"COUNTER","amount":20000000,"term":24,"productCode":"WORKING","rate":21}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("422: %d", w.Code)
	}
}
