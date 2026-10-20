package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type fakeResolver struct{}

func (fakeResolver) Resolve(_ context.Context, token string) (*User, error) {
	switch token {
	case "client":
		return &User{ID: "u-1", Role: shared.RoleClient}, nil
	case "officer":
		return &User{ID: "u-9", Role: shared.RoleBankOfficer}, nil
	}
	return nil, nil
}

func router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpx.TraceID(), Middleware(fakeResolver{}))
	r.GET("/client", RequireAuth(), func(c *gin.Context) { c.Status(200) })
	r.GET("/bank", RequireAuth(), RequireRole(shared.RoleBankOfficer), func(c *gin.Context) { c.Status(200) })
	return r
}

func get(r *gin.Engine, path, token string) int {
	req := httptest.NewRequest("GET", path, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuth(t *testing.T) {
	r := router()
	for _, tc := range []struct {
		path, token string
		want        int
	}{
		{"/client", "", 401},
		{"/client", "bogus", 401},
		{"/client", "client", 200},
		{"/bank", "client", 403},
		{"/bank", "officer", 200},
		{"/bank", "", 401},
	} {
		if got := get(r, tc.path, tc.token); got != tc.want {
			t.Errorf("%s token=%q: %d, ожидалось %d", tc.path, tc.token, got, tc.want)
		}
	}
}
