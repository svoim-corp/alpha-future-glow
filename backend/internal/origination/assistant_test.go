package origination

import (
	"context"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
	"sync"
	"testing"
)

func TestAssistantDraftReplay(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	who := shared.DraftApplicant{UserID: "u-1", CompanyID: "c-1", PersonINN: smirnovPerson.INN, Name: smirnovPerson.FullName}
	in := shared.AssistantDraft{ProductCode: "WORKING", Amount: 1500000, Term: 24, Goal: "WORKING_CAPITAL"}
	port := e.m.Assistant()
	var wg sync.WaitGroup
	ids := make(chan string, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := port.CreateDraft(ctx, who, in, "one-confirmed-action")
			if err != nil {
				t.Error(err)
				return
			}
			ids <- a.ID
		}()
	}
	wg.Wait()
	close(ids)
	id := ""
	count := 0
	for result := range ids {
		if id != "" && id != result {
			t.Error("повтор создал другой черновик")
		}
		id = result
		count++
	}
	if count != 4 {
		t.Fatalf("результатов: %d", count)
	}
	rows, err := port.List(ctx, "c-1")
	if err != nil || len(rows) != 1 {
		t.Fatalf("дубли: %+v %v", rows, err)
	}
	bad := who
	bad.UserID = "u-3"
	bad.PersonINN = "missing"
	if _, err := port.CreateDraft(ctx, bad, in, "other-action"); err == nil {
		t.Fatal("нет полномочий, но черновик создан")
	}
}
