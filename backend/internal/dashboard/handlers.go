package dashboard

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const (
	maxBody    = 64 << 10 // 30 виджетов — несколько килобайт
	maxIdemKey = 128
	// cacheControl — настройка личная: без общего кеша и с перепроверкой при каждом чтении.
	cacheControl = "private, no-cache"
)

var errPreconditionRequired = &httpx.Error{Status: http.StatusPreconditionRequired, Code: httpx.CodePreconditionReq,
	Message: "Нужен заголовок If-Match с версией настройки или If-None-Match: * для первого сохранения"}

type documentDTO struct {
	Revision  string    `json:"revision" example:"8efc6500-8373-4e9d-9f0a-a5d9efae1175"`
	UpdatedAt time.Time `json:"updatedAt" example:"2026-10-22T10:00:00Z"`
	Config    config    `json:"config"`
}

// getDashboard godoc
//
//	@Summary		Настройка главной
//	@Description	Сохранённая раскладка виджетов пользователя в его компании. 204 — настройку ещё не сохраняли, клиент строит исходную.
//	@Tags			dashboard
//	@Produce		json
//	@Success		200	{object}	documentDTO	"ETag — ревизия в кавычках"
//	@Success		204
//	@Failure		401	{object}	httpx.ErrorBody
//	@Failure		403	{object}	httpx.ErrorBody
//	@Router			/me/dashboard [get]
func (m *Module) getDashboard(c *gin.Context) {
	s, err := scopeOf(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	d, err := m.store.Dashboard(c.Request.Context(), s.userID, s.companyID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Header("Cache-Control", cacheControl)
	if d == nil {
		c.Status(http.StatusNoContent)
		return
	}
	writeDocument(c, d)
}

// putDashboard godoc
//
//	@Summary		Сохранить настройку главной
//	@Description	Заменяет документ целиком и атомарно. Нужен If-Match с ревизией или If-None-Match: * для первого сохранения.
//	@Description	Повтор с тем же Idempotency-Key и телом возвращает исходный результат.
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Param			If-Match		header		string	false	"Ревизия сохранённой настройки в кавычках"
//	@Param			If-None-Match	header		string	false	"* — первое сохранение"
//	@Param			Idempotency-Key	header		string	false	"Ключ попытки сохранения"
//	@Param			body			body		config	true	"Настройка формата v1"
//	@Success		200				{object}	documentDTO
//	@Failure		400				{object}	httpx.ErrorBody	"VALIDATION_ERROR или DASHBOARD_SCHEMA_UNSUPPORTED"
//	@Failure		401				{object}	httpx.ErrorBody
//	@Failure		403				{object}	httpx.ErrorBody
//	@Failure		409				{object}	httpx.ErrorBody	"IDEMPOTENCY_KEY_REUSED"
//	@Failure		412				{object}	httpx.ErrorBody	"DASHBOARD_VERSION_CONFLICT"
//	@Failure		428				{object}	httpx.ErrorBody	"PRECONDITION_REQUIRED"
//	@Router			/me/dashboard [put]
func (m *Module) putDashboard(c *gin.Context) {
	s, err := scopeOf(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	pre, err := preconditionOf(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if len(key) > maxIdemKey {
		httpx.Fail(c, httpx.Validation("Ключ идемпотентности длиннее 128 символов"))
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxBody))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpx.Fail(c, httpx.Validation("Настройка главной слишком большая"))
			return
		}
		httpx.Fail(c, err)
		return
	}
	cfg, err := parseConfig(body)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	sum := sha256.Sum256(body)
	d, err := m.save(c.Request.Context(), s, saveInput{cfg: cfg, pre: pre, key: key, bodyHash: hex.EncodeToString(sum[:])})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	writeDocument(c, d)
}

// scopeOf — настройка главной есть только у клиента с компанией.
func scopeOf(c *gin.Context) (scope, error) {
	u := auth.UserFrom(c)
	if u.Role != shared.RoleClient || u.CompanyID == nil {
		return scope{}, httpx.Forbidden("Настройка главной доступна клиенту банка")
	}
	return scope{userID: u.ID, companyID: *u.CompanyID}, nil
}

// preconditionOf — ровно одно из If-Match и If-None-Match: *. Без них старый клиент мог бы молча перезаписать
// настройку, сохранённую на другом устройстве.
func preconditionOf(c *gin.Context) (precondition, error) {
	ifMatch := strings.TrimSpace(c.GetHeader("If-Match"))
	ifNone := strings.TrimSpace(c.GetHeader("If-None-Match"))
	switch {
	case ifMatch != "" && ifNone != "":
		return precondition{}, httpx.Validation("Укажите только один из заголовков If-Match или If-None-Match")
	case ifNone == "*":
		return precondition{create: true}, nil
	case ifNone != "":
		return precondition{}, httpx.Validation("If-None-Match поддерживает только *")
	case ifMatch != "":
		return precondition{revision: strings.Trim(ifMatch, `"`)}, nil
	}
	return precondition{}, errPreconditionRequired
}

func writeDocument(c *gin.Context, d *dashboardRow) {
	c.Header("Cache-Control", cacheControl)
	c.Header("ETag", `"`+d.Revision+`"`)
	httpx.JSON(c, http.StatusOK, documentDTO{Revision: d.Revision, UpdatedAt: d.UpdatedAt.UTC(), Config: d.Config.V})
}
