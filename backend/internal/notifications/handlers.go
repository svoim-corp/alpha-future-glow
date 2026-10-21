package notifications

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// listLimit — в шапке показываются последние уведомления, а не вся история.
const listLimit = 50

type notificationDTO struct {
	ID        string                  `json:"id" example:"n-15"`
	Type      shared.NotificationType `json:"type" example:"INFO_REQUESTED"`
	Title     string                  `json:"title" example:"Банку нужны пояснения по заявке"`
	Text      string                  `json:"text" example:"Уточните фактический адрес склада"`
	Link      string                  `json:"link" example:"/applications/a-1024"`
	Read      bool                    `json:"read" example:"false"`
	CreatedAt time.Time               `json:"createdAt" example:"2026-10-20T12:40:00Z"`
}

type listResponse struct {
	UnreadCount int               `json:"unreadCount" example:"2"`
	Items       []notificationDTO `json:"items"`
}

// list godoc
//
//	@Summary		Уведомления в шапке
//	@Description	Последние 50 уведомлений пользователя, новые первыми. unread=true — только непрочитанные.
//	@Tags			notifications
//	@Produce		json
//	@Param			unread	query		bool	false	"Только непрочитанные"
//	@Success		200		{object}	listResponse
//	@Failure		401		{object}	httpx.ErrorBody
//	@Router			/notifications [get]
func (m *Module) list(c *gin.Context) {
	ctx, u := c.Request.Context(), auth.UserFrom(c)
	rows, err := m.store.List(ctx, u.ID, c.Query("unread") == "true", listLimit)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	unread, err := m.store.UnreadCount(ctx, u.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	resp := listResponse{UnreadCount: unread, Items: make([]notificationDTO, len(rows))}
	for i, r := range rows {
		resp.Items[i] = notificationDTO{
			ID: r.ID, Type: shared.NotificationType(r.Type), Title: r.Title, Text: r.Text, Link: r.Link, Read: r.Read, CreatedAt: r.CreatedAt,
		}
	}
	httpx.JSON(c, http.StatusOK, resp)
}
