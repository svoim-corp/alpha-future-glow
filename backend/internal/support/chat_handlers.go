package support

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
	"net/http"
	"strings"
)

func (m *Module) registerChat(r httpx.Routes) {
	r.Client.GET("/support/conversations", m.listConversations)
	r.Client.POST("/support/conversations", m.createConversation)
	r.Client.GET("/support/conversations/:id/messages", m.listMessages)
	r.Client.POST("/support/conversations/:id/messages", m.sendMessage)
	r.Client.POST("/support/conversations/:id/actions/:actionId/confirm", m.deps.Idem.Middleware(), m.confirmAction)
}

func chatUser(c *gin.Context) (*auth.User, error) {
	u := auth.UserFrom(c)
	if u == nil {
		return nil, httpx.Unauthorized()
	}
	if u.Role != shared.RoleClient || u.CompanyID == nil {
		return nil, httpx.Forbidden("Ассистент доступен клиенту банка")
	}
	return u, nil
}

// listConversations godoc
// @Summary Диалоги пользователя
// @Tags support
// @Produce json
// @Success 200 {array} conversation
// @Failure 403 {object} httpx.ErrorBody
// @Router /support/conversations [get]
func (m *Module) listConversations(c *gin.Context) {
	u, err := chatUser(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out, err := m.store.conversations(c.Request.Context(), u.ID, *u.CompanyID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, out)
}

type conversationRequest struct {
	ApplicationID *string `json:"applicationId" binding:"omitempty,max=100"`
}

// createConversation godoc
// @Summary Новый диалог с ассистентом
// @Tags support
// @Accept json
// @Produce json
// @Param body body conversationRequest true "Контекст"
// @Success 201 {object} conversation
// @Failure 400 {object} httpx.ErrorBody
// @Router /support/conversations [post]
func (m *Module) createConversation(c *gin.Context) {
	u, err := chatUser(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var in conversationRequest
	if err := httpx.Bind(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}
	if in.ApplicationID != nil {
		a, err := m.ports.Applications.Application(c.Request.Context(), *in.ApplicationID)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if a == nil || a.CompanyID != *u.CompanyID {
			httpx.Fail(c, httpx.NotFound("Заявка не найдена"))
			return
		}
	}
	out := conversation{ID: chatID(), UserID: u.ID, CompanyID: *u.CompanyID, ApplicationID: in.ApplicationID}
	err = m.deps.DB.GetContext(c.Request.Context(), &out.CreatedAt, `INSERT INTO support.conversations(id,user_id,company_id,application_id) VALUES($1,$2,$3,$4) RETURNING created_at`, out.ID, out.UserID, out.CompanyID, out.ApplicationID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, out)
}

func (m *Module) conversationFrom(c *gin.Context) (*conversation, *auth.User, error) {
	u, err := chatUser(c)
	if err != nil {
		return nil, nil, err
	}
	v, err := m.store.ownConversation(c.Request.Context(), c.Param("id"), u.ID, *u.CompanyID)
	return v, u, err
}

// listMessages godoc
// @Summary Сообщения диалога
// @Tags support
// @Produce json
// @Param id path string true "Диалог"
// @Success 200 {array} messageDTO
// @Failure 404 {object} httpx.ErrorBody
// @Router /support/conversations/{id}/messages [get]
func (m *Module) listMessages(c *gin.Context) {
	v, _, err := m.conversationFrom(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := m.store.messages(c.Request.Context(), v.ID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]messageDTO, len(rows))
	var completed []struct {
		ID     string         `db:"id"`
		Result db.JSON[block] `db:"result"`
	}
	if err := m.deps.DB.SelectContext(c.Request.Context(), &completed, `SELECT id,result FROM support.actions WHERE conversation_id=$1 AND result IS NOT NULL ORDER BY created_at,id LIMIT 100`, v.ID); err != nil {
		httpx.Fail(c, err)
		return
	}
	results := map[string]block{}
	for _, action := range completed {
		results[action.ID] = action.Result.V
	}
	for i, r := range rows {
		out[i] = r.dto()
		for j, b := range out[i].Blocks {
			if result, ok := results[b.ActionID]; ok {
				out[i].Blocks[j] = result
			}
		}
	}
	httpx.JSON(c, http.StatusOK, out)
}

type messageRequest struct {
	Text string `json:"text" binding:"required,max=4000"`
}

// sendMessage godoc
// @Summary Задать вопрос ассистенту
// @Tags support
// @Accept json
// @Produce json
// @Param id path string true "Диалог"
// @Param Idempotency-Key header string true "Ключ запроса"
// @Param body body messageRequest true "Вопрос"
// @Success 202 {object} messageDTO
// @Failure 409 {object} httpx.ErrorBody
// @Router /support/conversations/{id}/messages [post]
func (m *Module) sendMessage(c *gin.Context) {
	v, _, err := m.conversationFrom(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var in messageRequest
	if err := httpx.Bind(c, &in); err != nil {
		httpx.Fail(c, err)
		return
	}
	key := c.GetHeader("Idempotency-Key")
	if strings.TrimSpace(in.Text) == "" || key == "" || len(key) > 100 {
		httpx.Fail(c, httpx.Validation("Введите вопрос и ключ запроса"))
		return
	}
	if m.ai == nil || m.ai.cfg.URL == "" {
		httpx.Fail(c, httpx.Upstream("Ассистент сейчас недоступен. Частые вопросы и обращения продолжают работать."))
		return
	}
	out, err := m.store.enqueue(c.Request.Context(), v.ID, in.Text, key)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusAccepted, out.dto())
}

// confirmAction godoc
// @Summary Подтвердить предложение ассистента
// @Tags support
// @Produce json
// @Param id path string true "Диалог"
// @Param actionId path string true "Предложение"
// @Param Idempotency-Key header string true "Ключ запроса"
// @Success 200 {object} block
// @Failure 404 {object} httpx.ErrorBody
// @Router /support/conversations/{id}/actions/{actionId}/confirm [post]
func (m *Module) confirmAction(c *gin.Context) {
	conv, u, err := m.conversationFrom(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if key := c.GetHeader("Idempotency-Key"); key == "" || len(key) > 100 {
		httpx.Fail(c, httpx.Validation("Не указан ключ подтверждения"))
		return
	}
	ctx := c.Request.Context()
	var row struct {
		Params db.JSON[actionParams] `db:"params"`
		Result db.JSON[*block]       `db:"result"`
	}
	id := c.Param("actionId")
	if err := m.deps.DB.GetContext(ctx, &row, `SELECT params,result FROM support.actions WHERE id=$1 AND conversation_id=$2`, id, conv.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.Fail(c, httpx.NotFound("Предложение не найдено"))
		} else {
			httpx.Fail(c, err)
		}
		return
	}
	if row.Result.V != nil {
		httpx.JSON(c, http.StatusOK, row.Result.V)
		return
	}
	out := block{Kind: "result", Title: "Действие выполнено"}
	switch p := row.Params.V; p.Kind {
	case "draft":
		if p.Draft == nil {
			httpx.Fail(c, httpx.Validation("Нет параметров черновика"))
			return
		}
		who := shared.DraftApplicant{UserID: u.ID, CompanyID: *u.CompanyID, Name: u.FullName}
		if u.PersonINN != nil {
			who.PersonINN = *u.PersonINN
		}
		a, err := m.ports.Drafts.CreateDraft(ctx, who, *p.Draft, id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		out.Title = "Черновик создан"
		out.Application = appCard(*a)
		out.ResultID = a.ID
	case "ticket":
		t := &ticketRow{ID: "t-chat-" + id, UserID: u.ID, CompanyID: *u.CompanyID, ApplicationID: conv.ApplicationID, Subject: p.Subject, Text: p.Text}
		if err := m.openTicket(ctx, t); err != nil {
			httpx.Fail(c, err)
			return
		}
		out.Title = "Обращение отправлено"
		out.ResultID = t.ID
	default:
		httpx.Fail(c, httpx.Validation("Недоступное действие"))
		return
	}
	if _, err := m.deps.DB.ExecContext(ctx, `UPDATE support.actions SET result=COALESCE(result,$2) WHERE id=$1`, id, db.JSON[block]{V: out}); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, out)
}
