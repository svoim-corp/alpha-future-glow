package support

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStrictToolArguments(t *testing.T) {
	for _, raw := range []string{`{}`, `{"productCode":"INVEST","amount":-1,"term":24,"goal":"EQUIPMENT"}`, `{"productCode":"INVEST","amount":2000000,"term":24,"goal":"EQUIPMENT","companyId":"c-3"}`, `{} {}`, `null`} {
		t.Run(raw, func(t *testing.T) {
			var in shared.AssistantDraft
			if err := strictArgs([]byte(raw), &in); err == nil {
				t.Fatal("ожидалась ошибка аргументов")
			}
		})
	}
	var valid shared.AssistantDraft
	if err := strictArgs([]byte(`{"productCode":"INVEST","amount":2000000,"term":24,"goal":"EQUIPMENT"}`), &valid); err != nil {
		t.Fatal(err)
	}
}

type chatCatalog struct{}

func (chatCatalog) Search(context.Context, string, int64, int) ([]shared.ProductView, error) {
	return []shared.ProductView{{Code: "INVEST", Name: "Инвестиционный", Goals: []string{"EQUIPMENT"}, AmountMin: 1000000, AmountMax: 30000000, TermMin: 12, TermMax: 84}}, nil
}
func (chatCatalog) Calculate(context.Context, string, int64, int) (*shared.PaymentEstimate, error) {
	return &shared.PaymentEstimate{Amount: 2000000, Term: 24, Min: 100000, Max: 110000}, nil
}

type chatDrafts struct {
	company string
	creates int
}

func (f *chatDrafts) List(_ context.Context, company string) ([]shared.ApplicationView, error) {
	return []shared.ApplicationView{{ID: "a-1", CompanyID: f.company, Status: shared.StatusDraft, Amount: 2000000, Term: 24}}, nil
}
func (f *chatDrafts) CreateDraft(context.Context, shared.DraftApplicant, shared.AssistantDraft, string) (*shared.ApplicationView, error) {
	f.creates++
	return nil, errors.New("не должно вызываться из LLM")
}

type chatPayments struct{}

func (chatPayments) Upcoming(context.Context, string) ([]shared.UpcomingPayment, error) {
	return []shared.UpcomingPayment{{LoanID: "l-1", Date: "2026-11-01", Amount: 100000}}, nil
}

func TestToolAllowlistAndProposals(t *testing.T) {
	f := &chatDrafts{company: "c-1"}
	m := &Module{ports: Ports{Catalog: chatCatalog{}, Drafts: f, Payments: chatPayments{}}}
	conv := conversation{CompanyID: "c-1"}
	for _, tt := range []struct {
		name, args, kind string
		fail             bool
	}{
		{"products", `{}`, "product", false},
		{"products", `{"amount":1000000}`, "", true},
		{"calculate", `{"productCode":"INVEST","amount":2000000,"term":24}`, "calculation", false},
		{"applications", `{}`, "application", false},
		{"applications", `{"companyId":"c-3"}`, "", true},
		{"payments", `{}`, "payment", false},
		{"propose_draft", `{"productCode":"INVEST","amount":2000000,"term":24,"goal":"EQUIPMENT"}`, "action", false},
		{"propose_ticket", `{"subject":"Ошибка","text":"Не открывается заявка"}`, "action", false},
		{"propose_ticket", `{"subject":" ","text":" "}`, "", true},
		{"execute_sql", `{"sql":"select * from identity.users"}`, "", true},
		{"submit_application", `{}`, "", true},
	} {
		t.Run(tt.name+tt.args, func(t *testing.T) {
			blocks, err := m.executeTool(context.Background(), conv, toolCall{Function: toolFunction{Name: tt.name, Arguments: json.RawMessage(tt.args)}})
			if (err != nil) != tt.fail {
				t.Fatalf("err=%v", err)
			}
			if !tt.fail && (len(blocks) != 1 || blocks[0].Kind != tt.kind) {
				t.Fatalf("карточки: %+v", blocks)
			}
		})
	}
	if f.creates != 0 {
		t.Fatal("модель выполнила команду без подтверждения")
	}
	f.company = "c-3"
	if _, err := m.executeTool(context.Background(), conv, toolCall{Function: toolFunction{Name: "applications", Arguments: json.RawMessage(`{}`)}}); err == nil {
		t.Fatal("чужая компания должна быть запрещена")
	}
}

func TestOllamaLocalBoundary(t *testing.T) {
	for _, tt := range []struct {
		name       string
		status     int
		dimensions int
		fail       bool
	}{
		{"успех", 200, 1024, false}, {"неверная размерность", 200, 12, true}, {"модель недоступна", 503, 1024, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/embed" {
					t.Errorf("путь %s", r.URL.Path)
				}
				w.WriteHeader(tt.status)
				_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": [][]float32{make([]float32, tt.dimensions)}})
			}))
			defer s.Close()
			m := &Module{}
			m.ConfigureAI(AIConfig{URL: s.URL, EmbeddingModel: "bge-m3:567m"})
			_, err := m.ai.embed(context.Background(), "Как оформить заявку?")
			if (err != nil) != tt.fail {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestKnowledgeChunks(t *testing.T) {
	text := "# Кабинет\nКраткая справка\n## Кредиты\n" + strings.Repeat("финансирование ", 650)
	chunks := splitKnowledge("test", text)
	if len(chunks) < 3 {
		t.Fatal("длинный текст не разбит")
	}
	for _, c := range chunks {
		if c.Hash == "" || c.Title == "" || len(strings.Fields(c.Content)) > 302 {
			t.Fatalf("некорректный фрагмент %+v", c)
		}
	}
	if _, err := vectorLiteral(make([]float32, 12)); err == nil {
		t.Fatal("размерность должна проверяться")
	}
}

func TestConversationPrivacyAndQueue(t *testing.T) {
	e := newEnv(t)
	foreign := e.as("smirnov", "POST", "/api/v1/support/conversations", `{"applicationId":"a-2"}`)
	if foreign.Code != 404 {
		t.Fatalf("чужая заявка: %d %s", foreign.Code, foreign.Body)
	}
	created := e.as("smirnov", "POST", "/api/v1/support/conversations", `{"applicationId":"a-1"}`)
	if created.Code != 201 {
		t.Fatalf("создание: %s", created.Body)
	}
	c := decode[conversation](t, created)
	for _, u := range []string{"petrova", "orlova", "ivanova"} {
		if w := e.as(u, "GET", "/api/v1/support/conversations/"+c.ID+"/messages", ""); w.Code != 404 && w.Code != 403 {
			t.Fatalf("доступ %s: %d", u, w.Code)
		}
	}
	ctx := context.Background()
	first, err := e.m.store.enqueue(ctx, c.ID, "Покажи заявки", "key-1")
	if err != nil {
		t.Fatal(err)
	}
	replay, err := e.m.store.enqueue(ctx, c.ID, "Покажи заявки", "key-1")
	if err != nil || first.ID != replay.ID {
		t.Fatal("повтор создал сообщение", err)
	}
	if _, err := e.m.store.enqueue(ctx, c.ID, "Другой вопрос", "key-1"); err == nil {
		t.Fatal("ключ с другим текстом")
	}
	if _, err := e.m.store.enqueue(ctx, c.ID, "Новый вопрос", "key-2"); err == nil {
		t.Fatal("второй незавершённый вопрос")
	}
	if err := e.m.RecoverChat(ctx); err != nil {
		t.Fatal(err)
	}
	rows, err := e.m.store.messages(ctx, c.ID)
	if err != nil || len(rows) != 2 || rows[1].Status != "FAILED" {
		t.Fatalf("восстановление: %+v %v", rows, err)
	}
	if w := e.as("smirnov", "POST", "/api/v1/support/conversations/"+c.ID+"/messages", `{"text":"Привет"}`); w.Code != 400 {
		t.Fatalf("нет ключа: %d", w.Code)
	}
}

func TestConfirmedTicketIdempotency(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	params := db.JSON[actionParams]{V: actionParams{Kind: "ticket", Subject: "Ошибка", Text: "Не открывается заявка"}}
	if _, err := e.m.deps.DB.ExecContext(ctx, `INSERT INTO support.actions(id,conversation_id,params) VALUES('test-action',$1,$2)`, c.ID, params); err != nil {
		t.Fatal(err)
	}
	// Стабильный id защищает доменную запись даже до сохранения результата предложения.
	for i := 0; i < 2; i++ {
		row := &ticketRow{ID: "t-chat-test-action", UserID: "u-1", CompanyID: "c-1", Subject: "Ошибка", Text: "Не открывается заявка"}
		if err := e.m.openTicket(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := e.m.store.Tickets(ctx, "u-1", 100)
	if err != nil || len(rows) != 1 {
		t.Fatalf("дубли: %+v %v", rows, err)
	}
}

func TestRAGAndConfirmedAction(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	chatCalls := 0
	inference := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/embed" {
			v := make([]float32, 1024)
			v[0] = 1
			_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": [][]float32{v}})
			return
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("неожиданный путь %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		chatCalls++
		reply := llmMessage{Role: "assistant", Content: "Проверьте текст обращения и подтвердите отправку на карточке."}
		if chatCalls == 1 {
			reply.Calls = []toolCall{{Function: toolFunction{Name: "propose_ticket", Arguments: json.RawMessage(`{"subject":"Неверный адрес","text":"В карточке компании указан старый адрес"}`)}}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": reply})
	}))
	defer inference.Close()
	e.m.ConfigureAI(AIConfig{URL: inference.URL, Model: "qwen-test", EmbeddingModel: "bge-test"})
	if err := e.m.IndexKnowledge(ctx); err != nil {
		t.Fatal(err)
	}
	if err := e.m.IndexKnowledge(ctx); err != nil {
		t.Fatal("повтор индексации", err)
	}
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	if _, err := e.m.store.enqueue(ctx, c.ID, "В карточке компании неверный адрес. Подготовь обращение.", "rag-test"); err != nil {
		t.Fatal(err)
	}
	rows, err := e.m.store.messages(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	pending := rows[len(rows)-1]
	text, payload, err := e.m.answer(ctx, pending)
	if err != nil {
		t.Fatal(err)
	}
	if len(payload.Sources) == 0 || len(payload.Blocks) != 1 || payload.Blocks[0].Kind != "action" {
		t.Fatalf("ответ: %+v", payload)
	}
	if err := e.m.store.complete(ctx, pending, text, "COMPLETED", payload); err != nil {
		t.Fatal(err)
	}
	tickets, err := e.m.store.Tickets(ctx, "u-1", 100)
	if err != nil || len(tickets) != 0 {
		t.Fatal("до подтверждения создано обращение", err)
	}
	path := "/api/v1/support/conversations/" + c.ID + "/actions/" + payload.Blocks[0].ActionID + "/confirm"
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", path, nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "smirnov"})
		req.Header.Set("Idempotency-Key", "confirm-"+payload.Blocks[0].ActionID)
		w := httptest.NewRecorder()
		e.router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("подтверждение %d: %d %s", i, w.Code, w.Body)
		}
	}
	tickets, err = e.m.store.Tickets(ctx, "u-1", 100)
	if err != nil || len(tickets) != 1 {
		t.Fatalf("повтор создал дубли: %+v %v", tickets, err)
	}
	visible := decode[[]messageDTO](t, e.as("smirnov", "GET", "/api/v1/support/conversations/"+c.ID+"/messages", ""))
	if visible[len(visible)-1].Blocks[0].Kind != "result" {
		t.Fatal("после перезагрузки не виден результат")
	}
	if w := e.as("petrova", "POST", path, ""); w.Code != 404 {
		t.Fatalf("коллега получил доступ к подтверждению: %d", w.Code)
	}
}
