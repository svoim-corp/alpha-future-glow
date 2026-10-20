package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSON пишет успешный ответ.
func JSON(c *gin.Context, status int, v any) { c.JSON(status, v) }

// Fail пишет ошибку в едином формате и прерывает цепочку.
// Неизвестные ошибки логируются и превращаются в 500 INTERNAL.
func Fail(c *gin.Context, err error) {
	trace := TraceIDFrom(c)
	var e *Error
	if errors.As(err, &e) {
		c.AbortWithStatusJSON(e.Status, ErrorBody{Code: e.Code, Message: e.Message, Fields: e.Fields, TraceID: trace})
		return
	}
	slog.Error("internal error", "err", err, "traceId", trace, "path", c.Request.URL.Path)
	c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorBody{
		Code: CodeInternal, Message: "Внутренняя ошибка сервера", TraceID: trace,
	})
}
