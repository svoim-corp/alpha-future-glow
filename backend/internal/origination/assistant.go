package origination

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Assistant отдаёт поддержке чтение и создание черновика через правила Origination.
func (m *Module) Assistant() shared.AssistantApplications { return assistant{m} }

type assistant struct{ m *Module }

func (p assistant) List(ctx context.Context, companyID string) ([]shared.ApplicationView, error) {
	rows, _, err := p.m.store.CompanyApplications(ctx, companyID, 100, 0)
	if err != nil {
		return nil, err
	}
	out := make([]shared.ApplicationView, 0, len(rows))
	for _, a := range rows {
		out = append(out, a.view())
	}
	return out, nil
}

func (p assistant) CreateDraft(ctx context.Context, who shared.DraftApplicant, in shared.AssistantDraft, commandID string) (*shared.ApplicationView, error) {
	if commandID == "" {
		return nil, httpx.Validation("Не указан идентификатор действия")
	}
	// Стабильный id не допускает дубликат при сбое между созданием и сохранением результата в Support.
	id := fmt.Sprintf("a-chat-%x", sha256.Sum256([]byte(who.UserID+":"+commandID)))
	existing, err := p.m.store.Application(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.CompanyID != who.CompanyID || existing.ApplicantID != who.UserID {
			return nil, httpx.Forbidden("")
		}
		v := existing.view()
		return &v, nil
	}
	a, err := p.m.createDraft(ctx, applicant{companyID: who.CompanyID, personINN: who.PersonINN,
		actor: shared.Actor{Type: shared.ActorClient, UserID: who.UserID, Name: who.Name}},
		draftInput{ProductCode: in.ProductCode, Amount: in.Amount, Term: &in.Term, Goal: &in.Goal, CommandID: id})
	if err != nil {
		return nil, err
	}
	v := a.view()
	return &v, nil
}
