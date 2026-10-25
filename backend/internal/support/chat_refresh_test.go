package support

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type currentApps struct {
	chatDrafts
	row   *shared.ApplicationView
	calls int
}

func (a *currentApps) Find(context.Context, string, []string) ([]shared.ApplicationView, error) {
	a.calls++
	return []shared.ApplicationView{*a.row}, nil
}

func TestHistoryRefreshesDeletedDrafts(t *testing.T) {
	apps := &currentApps{row: &shared.ApplicationView{ID: "a-1", CompanyID: "c-1", Status: shared.StatusDraft, Amount: 150000, Term: 12}}
	m := &Module{ports: Ports{Drafts: apps}}
	messages := []messageDTO{{Blocks: []block{{Kind: "application", Application: &applicationCard{ID: "a-1"}}, {Kind: "application", Application: &applicationCard{ID: "a-1"}}}}}
	if err := m.refreshApplications(context.Background(), "c-1", messages); err != nil {
		t.Fatal(err)
	}
	if apps.calls != 1 || !messages[0].Blocks[0].Application.CanDelete {
		t.Fatal("черновик не актуализирован", messages, apps.calls)
	}
	apps.row.Status = shared.StatusClosed
	if err := m.refreshApplications(context.Background(), "c-1", messages); err != nil {
		t.Fatal(err)
	}
	if len(messages[0].Blocks) != 1 || messages[0].Blocks[0].Application != nil {
		t.Fatal("удалённый черновик остался доступен", messages)
	}
	apps.row.CompanyID = "c-3"
	messages[0].Blocks = []block{{Application: &applicationCard{ID: "a-1"}}}
	if err := m.refreshApplications(context.Background(), "c-1", messages); err == nil {
		t.Fatal("раскрыта чужая заявка")
	}
}

type viewerCompanies struct{}

func (viewerCompanies) GetByID(context.Context, string) (*shared.CompanyView, error) {
	return &shared.CompanyView{ID: "c-1", Name: "Северная пекарня"}, nil
}
func (viewerCompanies) Profile(context.Context, string) (*shared.CompanyProfile, error) {
	return &shared.CompanyProfile{CompanyView: shared.CompanyView{ID: "c-1", Name: "Северная пекарня"}}, nil
}

func TestAccountQuestionUsesSessionContext(t *testing.T) {
	e := newEnv(t)
	e.m.ports.Companies = viewerCompanies{}
	c := decode[conversation](t, e.as("smirnov", "POST", "/api/v1/support/conversations", `{}`))
	ctx := context.Background()
	if _, err := e.m.store.enqueue(ctx, c.ID, "Вот у меня как компания называется и кто я?", "identity", viewerContext{Name: "Олег Смирнов", Position: "Генеральный директор"}); err != nil {
		t.Fatal(err)
	}
	rows, _ := e.m.store.messages(ctx, c.ID)
	text, payload, err := e.m.answer(ctx, rows[len(rows)-1])
	if err != nil || !strings.Contains(text, "Олег Смирнов") || !strings.Contains(text, "Северная пекарня") || payload.Blocks[0].Company.ViewerPosition != "Генеральный директор" {
		t.Fatalf("не использованы данные входа: %s %+v %v", text, payload, err)
	}
	raw, _ := json.Marshal(rows[0].dto())
	if strings.Contains(string(raw), "Олег Смирнов") {
		t.Fatal("внутренний контекст попал в API сообщения пользователя")
	}
	conversations, err := e.m.store.conversations(ctx, "u-1", "c-1")
	if err != nil || conversations[0].Title != "Вот у меня как компания называется и кто я?" {
		t.Fatal("нет названия диалога", conversations, err)
	}
}
