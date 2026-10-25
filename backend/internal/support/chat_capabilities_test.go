package support

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type recordingCatalog struct {
	chatCatalog
	amount int64
	term   int
	goal   string
}

func (c *recordingCatalog) Search(_ context.Context, goal string, amount int64, term int) ([]shared.ProductView, error) {
	c.amount, c.term, c.goal = amount, term, goal
	return []shared.ProductView{{Code: "EXPRESS", Name: "Экспресс-кредит", Goals: []string{"EQUIPMENT"}}}, nil
}

func TestColloquialCreditKeepsKnownParameters(t *testing.T) {
	e := newEnv(t)
	catalog := &recordingCatalog{}
	e.m.ports.Catalog = catalog
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"message": llmMessage{Content: `{"tools":["products"],"params":{"amount":150000,"term":12,"goal":"EQUIPMENT","businessUse":"business"}}`}})
	}))
	defer s.Close()
	e.m.ConfigureAI(AIConfig{URL: s.URL, Model: "test"})
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	_, err := e.m.store.enqueue(context.Background(), c.ID, "Хочу 150к на год, пару телефонов для сотрудников", "colloquial")
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := e.m.store.messages(context.Background(), c.ID)
	_, payload, err := e.m.answer(context.Background(), rows[len(rows)-1])
	if err != nil || len(payload.Blocks) != 1 || catalog.amount != 150000 || catalog.term != 12 || catalog.goal != "EQUIPMENT" {
		t.Fatalf("потеряны параметры: %+v catalog=%+v err=%v", payload, catalog, err)
	}
}

func TestNavigationUsesRealRoutes(t *testing.T) {
	m := &Module{}
	blocks, err := m.executeTool(context.Background(), conversation{}, toolCall{Function: toolFunction{Name: "navigate", Arguments: json.RawMessage(`{"section":"planner"}`)}})
	if err != nil || len(blocks) != 1 || blocks[0].Navigation == nil || blocks[0].Navigation.Href != "/planner" {
		t.Fatalf("нет ссылки на планер: %+v %v", blocks, err)
	}
	for _, raw := range []string{`{"section":"https://example.com"}`, `{"section":"planner","href":"https://example.com"}`} {
		if _, err := m.executeTool(context.Background(), conversation{}, toolCall{Function: toolFunction{Name: "navigate", Arguments: json.RawMessage(raw)}}); err == nil {
			t.Fatal("принят произвольный адрес")
		}
	}
}
