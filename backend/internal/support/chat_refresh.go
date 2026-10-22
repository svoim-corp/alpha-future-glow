package support

import (
	"context"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Снимки сохраняют историю, но удалённый черновик нельзя снова открыть из старого ответа.
func (m *Module) refreshApplications(ctx context.Context, companyID string, messages []messageDTO) error {
	ids := []string{}
	seen := map[string]bool{}
	for _, message := range messages {
		for _, b := range message.Blocks {
			if b.Application != nil && !seen[b.Application.ID] {
				ids = append(ids, b.Application.ID)
				seen[b.Application.ID] = true
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if m.ports.Drafts == nil {
		return httpx.Upstream("Не удалось обновить заявки")
	}
	current, err := m.ports.Drafts.Find(ctx, companyID, ids)
	if err != nil {
		return err
	}
	cache := map[string]*shared.ApplicationView{}
	for _, a := range current {
		if a.CompanyID != companyID {
			return httpx.Forbidden("")
		}
		cache[a.ID] = &a
	}
	for i := range messages {
		blocks := make([]block, 0, len(messages[i].Blocks))
		removed := false
		for _, b := range messages[i].Blocks {
			if b.Application == nil {
				blocks = append(blocks, b)
				continue
			}
			id := b.Application.ID
			a := cache[id]
			if a == nil || a.Status == shared.StatusClosed {
				if !removed {
					blocks = append(blocks, block{Kind: "result", Title: "Черновик удалён или заявка закрыта"})
					removed = true
				}
				continue
			}
			b.Application = appCard(*a)
			blocks = append(blocks, b)
		}
		messages[i].Blocks = blocks
	}
	return nil
}
