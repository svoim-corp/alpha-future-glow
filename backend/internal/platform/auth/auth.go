// Package auth — middleware сессии и проверки ролей. Сами сессии хранит зона Identity.
package auth

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const CookieName = "session"

const (
	userKey   = "user"
	userIDKey = "userId"
)

type User struct {
	ID        string
	Role      shared.Role
	CompanyID *string
	PersonINN *string // ИНН человека в данных банка; по нему определяются полномочия. У сотрудника банка — nil.
	FullName  string
	Position  string
}

// SessionResolver находит пользователя по токену сессии (nil, nil — сессии нет).
type SessionResolver interface {
	Resolve(ctx context.Context, token string) (*User, error)
}

// Middleware кладёт пользователя в контекст, если есть валидная сессия.
// Без сессии запрос идёт дальше: доступ ограничивают RequireAuth/RequireRole.
func Middleware(res SessionResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie(CookieName); err == nil && token != "" {
			u, err := res.Resolve(c.Request.Context(), token)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			if u != nil {
				c.Set(userKey, u)
				c.Set(userIDKey, u.ID)
			}
		}
		c.Next()
	}
}

// RequireAuth отвечает 401, если пользователя нет.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if UserFrom(c) == nil {
			httpx.Fail(c, httpx.Unauthorized())
			return
		}
		c.Next()
	}
}

// RequireRole отвечает 403, если у пользователя другая роль.
func RequireRole(role shared.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := UserFrom(c)
		if u == nil {
			httpx.Fail(c, httpx.Unauthorized())
			return
		}
		if u.Role != role {
			httpx.Fail(c, httpx.Forbidden("Недостаточно прав для этого раздела"))
			return
		}
		c.Next()
	}
}

// UserFrom возвращает текущего пользователя или nil.
func UserFrom(c *gin.Context) *User {
	if v, ok := c.Get(userKey); ok {
		u, _ := v.(*User)
		return u
	}
	return nil
}
