package httpx

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// IfMatch читает версию заявки из заголовка If-Match (допускаются кавычки).
func IfMatch(c *gin.Context) (int, error) {
	raw := strings.Trim(strings.TrimSpace(c.GetHeader("If-Match")), `"`)
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, Validation("Нужен заголовок If-Match с версией заявки")
	}
	return v, nil
}
