package support

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStructuredPlan(t *testing.T) {
	for _, tt := range []struct {
		name, content string
		fail          bool
	}{
		{"справка", `{"tools":[]}`, false},
		{"актуальные данные", `{"tools":["applications","payments"]}`, false},
		{"неизвестная функция", `{"tools":["execute_sql"]}`, true},
		{"дублирование", `{"tools":["applications","applications"]}`, true},
		{"отсутствует поле", `{}`, true},
		{"посторонние поля", `{"tools":[],"companyId":"c-3"}`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Format json.RawMessage `json:"format"`
					Think  bool            `json:"think"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if len(body.Format) == 0 || body.Think {
					t.Error("классификация должна использовать JSON-схему без режима размышления")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"message": llmMessage{Role: "assistant", Content: tt.content}})
			}))
			defer server.Close()
			m := &Module{}
			m.ConfigureAI(AIConfig{URL: server.URL, Model: "test"})
			_, err := m.ai.plan(context.Background(), []llmMessage{{Role: "user", Content: "Покажи мои заявки"}})
			if (err != nil) != tt.fail {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestProposalTextIsNotExecution(t *testing.T) {
	for _, name := range []string{"propose_draft", "propose_ticket"} {
		text := cardAnswer(chatPlan{Tools: []string{name}}, nil)
		if !strings.Contains(text, "подтвержд") || strings.Contains(text, "уже создан") {
			t.Fatalf("неверный статус: %s", text)
		}
	}
	plan := chatPlan{Tools: []string{"calculate"}}
	available := plannedTools(plan, map[string]bool{})
	for _, tool := range available {
		if tool.Function.Name != "calculate" && tool.Function.Name != "products" {
			t.Fatalf("лишняя функция %s", tool.Function.Name)
		}
	}
	if planComplete(plan, map[string]bool{}) || !planComplete(plan, map[string]bool{"calculate": true}) {
		t.Fatal("неверное завершение плана")
	}
}

func TestFreshApplicationsIgnoreOldAnswer(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	e.m.ports.Drafts = &chatDrafts{company: "c-1"}
	inference := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Format json.RawMessage `json:"format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if len(body.Format) == 0 {
			t.Error("для списка заявок генерация фактов моделью не нужна")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": llmMessage{Role: "assistant", Content: `{"tools":["applications"]}`}})
	}))
	defer inference.Close()
	e.m.ConfigureAI(AIConfig{URL: inference.URL, Model: "test"})
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	if _, err := e.m.store.enqueue(ctx, c.ID, "Покажи мои заявки", "first"); err != nil {
		t.Fatal(err)
	}
	rows, err := e.m.store.messages(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	old := messagePayload{Blocks: []block{{Kind: "application", Application: &applicationCard{Amount: 100}}}}
	if err := e.m.store.complete(ctx, rows[len(rows)-1], "Старая сумма: 100 рублей", "COMPLETED", old); err != nil {
		t.Fatal(err)
	}
	if _, err := e.m.store.enqueue(ctx, c.ID, "Покажи мои заявки ещё раз", "second"); err != nil {
		t.Fatal(err)
	}
	rows, err = e.m.store.messages(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Для чтения личных данных embedding не нужен: это проверяется тестовым сервером,
	// поддерживающим только классификацию, и отсутствием вызова /api/embed.
	text, payload, err := e.m.answer(ctx, rows[len(rows)-1])
	if err != nil {
		t.Fatal(err)
	}
	if len(payload.Blocks) != 1 || payload.Blocks[0].Application.Amount != 2000000 {
		t.Fatalf("использован старый ответ: %+v", payload)
	}
	if !strings.Contains(text, "актуальные") {
		t.Fatalf("не показаны актуальные данные: %s", text)
	}
}

func TestGreetingAfterApplications(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	inference := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("приветствие не должно продолжать старую задачу или обращаться к модели")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer inference.Close()
	e.m.ConfigureAI(AIConfig{URL: inference.URL, Model: "test"})
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	if _, err := e.m.store.enqueue(ctx, c.ID, "Покажи мои заявки", "old"); err != nil {
		t.Fatal(err)
	}
	rows, _ := e.m.store.messages(ctx, c.ID)
	if err := e.m.store.complete(ctx, rows[len(rows)-1], "Показываю заявки", "COMPLETED", messagePayload{}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.m.store.enqueue(ctx, c.ID, "Привет!", "greeting"); err != nil {
		t.Fatal(err)
	}
	rows, _ = e.m.store.messages(ctx, c.ID)
	text, payload, err := e.m.answer(ctx, rows[len(rows)-1])
	if err != nil || !strings.Contains(text, "Здравствуйте") || len(payload.Blocks) != 0 || len(payload.Sources) != 0 {
		t.Fatalf("приветствие продолжило старую задачу: text=%s payload=%+v err=%v", text, payload, err)
	}
}

func TestPlanHasOneCurrentRequest(t *testing.T) {
	inference := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []llmMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		users := []string{}
		for _, msg := range body.Messages {
			if msg.Role == "user" {
				users = append(users, msg.Content)
			}
		}
		if len(users) != 1 || users[0] != "Чем отличается овердрафт?" {
			t.Errorf("история стала активными запросами: %v", users)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": llmMessage{Content: `{"tools":[]}`}})
	}))
	defer inference.Close()
	m := &Module{}
	m.ConfigureAI(AIConfig{URL: inference.URL, Model: "test"})
	_, err := m.ai.plan(context.Background(), []llmMessage{{Role: "user", Content: "Покажи мои заявки"}, {Role: "assistant", Content: "Показываю заявки"}, {Role: "user", Content: "Чем отличается овердрафт?"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGreetingDoesNotSwallowRequest(t *testing.T) {
	for _, question := range []string{"Привет, покажи мои заявки", "Добрый день! Посчитай платёж", "Как создать заявку?"} {
		if isGreeting(question) {
			t.Errorf("запрос принят за отдельное приветствие: %s", question)
		}
	}
	for _, question := range []string{"  Привет!  ", "ЗДРАВСТВУЙТЕ", "Добрый день."} {
		if !isGreeting(question) {
			t.Errorf("не распознано приветствие: %s", question)
		}
	}
}
