package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const traceKey = "traceId"

// TraceID берёт X-Request-Id или генерирует новый идентификатор запроса.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			b := make([]byte, 4)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set(traceKey, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

func TraceIDFrom(c *gin.Context) string { return c.GetString(traceKey) }

// Logger пишет одну строку лога на запрос.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		attrs := []any{
			"method", c.Request.Method,
			"route", c.FullPath(),
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"traceId", TraceIDFrom(c),
		}
		if uid := c.GetString("userId"); uid != "" {
			attrs = append(attrs, "userId", uid)
		}
		log.Info("request", attrs...)
	}
}

// Recovery превращает панику в 500 в нашем формате.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic", "panic", recovered, "traceId", TraceIDFrom(c), "path", c.Request.URL.Path)
		Fail(c, &Error{Status: 500, Code: CodeInternal, Message: "Внутренняя ошибка сервера"})
	})
}

func NoRoute(c *gin.Context) {
	Fail(c, NotFound("Ресурс не найден"))
}

func NoMethod(c *gin.Context) {
	Fail(c, &Error{Status: 405, Code: CodeMethodNotAllowed, Message: "Метод не поддерживается"})
}
