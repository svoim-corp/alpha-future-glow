package httpx

import "github.com/gin-gonic/gin"

// Routes — группы роутов, в которые зоны регистрируют свои эндпоинты.
// Собирается в cmd/api, чтобы httpx не зависел от auth.
type Routes struct {
	Public *gin.RouterGroup // /api/v1, без сессии
	Client *gin.RouterGroup // /api/v1 + RequireAuth
	Bank   *gin.RouterGroup // /api/v1/bank + RequireAuth + RequireRole(BANK_OFFICER)
}
