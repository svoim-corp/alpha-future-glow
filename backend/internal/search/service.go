package search

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const sourceTimeout = 2 * time.Second

func (m *Module) find(ctx context.Context, scope shared.SearchScope, text string, types []string, limit int, after shared.SearchPosition) ([]groupDTO, error) {
	providers := map[string]shared.SearchProvider{
		"applications": m.ports.Applications, "loans": m.ports.Loans, "plans": m.ports.Plans,
		"products": m.ports.Products, "faq": m.ports.FAQ, "tickets": m.ports.Tickets,
	}
	groups := make([]groupDTO, len(types))
	var work sync.WaitGroup
	for index, kind := range types {
		work.Add(1)
		go func(index int, kind string) {
			defer work.Done()
			group := groupDTO{Type: kind, Items: []itemDTO{}, Unavailable: true}
			defer func() { groups[index] = group }()
			provider := providers[kind]
			if provider == nil {
				return
			}
			sourceCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
			defer cancel()
			hits, err := provider(sourceCtx, scope, shared.SearchQuery{Text: text, Limit: limit + 1, After: after})
			if err != nil {
				if ctx.Err() == nil {
					slog.Warn("search source unavailable", "source", kind, "error", err)
				}
				return
			}
			group.Unavailable = false
			if len(hits) > limit {
				group.HasMore = true
				hits = hits[:limit]
				group.NextCursor, err = nextCursor(scope, text, kind, hits[len(hits)-1])
				if err != nil {
					group.Unavailable = true
					return
				}
			}
			for _, hit := range hits {
				group.Items = append(group.Items, itemDTO{ID: hit.ID, Title: hit.Title, Snippet: hit.Snippet, State: hit.State, Step: hit.Step})
			}
		}(index, kind)
	}
	work.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, group := range groups {
		if !group.Unavailable {
			return groups, nil
		}
	}
	return nil, &httpx.Error{Status: http.StatusServiceUnavailable, Code: httpx.CodeSearchUnavailable, Message: "Поиск временно недоступен. Повторите запрос позже"}
}
