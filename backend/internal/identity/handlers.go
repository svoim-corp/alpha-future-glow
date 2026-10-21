package identity

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const sessionTTL = 7 * 24 * time.Hour

type loginRequest struct {
	Login    string `json:"login" binding:"required" example:"smirnov"`
	Password string `json:"password" binding:"required" example:"demo123"`
}

type userDTO struct {
	ID        string      `json:"id" example:"u-1"`
	FullName  string      `json:"fullName" example:"Смирнов Олег Викторович"`
	Position  string      `json:"position" example:"Генеральный директор"`
	Role      shared.Role `json:"role" example:"CLIENT"`
	CompanyID *string     `json:"companyId" example:"c-1"`
}

type loginResponse struct {
	User     userDTO `json:"user"`
	Redirect string  `json:"redirect" example:"/"`
}

type companyDTO struct {
	ID   string `json:"id" example:"c-1"`
	INN  string `json:"inn" example:"7712345678"`
	Name string `json:"name" example:"ООО «Северная пекарня»"`
}

type meResponse struct {
	ID       string      `json:"id"`
	FullName string      `json:"fullName"`
	Position string      `json:"position"`
	Role     shared.Role `json:"role"`
	Company  *companyDTO `json:"company"`
}

// login godoc
//
//	@Summary	Вход по логину и паролю
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		loginRequest	true	"Учётные данные"
//	@Success	200		{object}	loginResponse
//	@Failure	400		{object}	httpx.ErrorBody
//	@Failure	401		{object}	httpx.ErrorBody
//	@Router		/auth/login [post]
func (m *Module) login(c *gin.Context) {
	var req loginRequest
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	u, err := m.store.FindByLogin(c.Request.Context(), req.Login)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	bad := &httpx.Error{Status: http.StatusUnauthorized, Code: httpx.CodeUnauthorized, Message: "Неверный логин или пароль"}
	if u == nil {
		// Выравниваем время ответа, чтобы не выдавать существование логина.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy"), []byte(req.Password))
		httpx.Fail(c, bad)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		httpx.Fail(c, bad)
		return
	}
	token, err := m.store.CreateSession(c.Request.Context(), u.ID, sessionTTL)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieName, token, int(sessionTTL.Seconds()), "/", "", false, true)

	redirect := "/"
	if shared.Role(u.Role) == shared.RoleBankOfficer {
		redirect = "/bank/queue"
	}
	httpx.JSON(c, http.StatusOK, loginResponse{
		User:     userDTO{ID: u.ID, FullName: u.FullName, Position: u.Position, Role: shared.Role(u.Role), CompanyID: u.CompanyID},
		Redirect: redirect,
	})
}

// logout godoc
//
//	@Summary	Завершить сессию
//	@Tags		auth
//	@Success	204
//	@Failure	401	{object}	httpx.ErrorBody
//	@Router		/auth/logout [post]
func (m *Module) logout(c *gin.Context) {
	if token, err := c.Cookie(auth.CookieName); err == nil {
		if err := m.store.DeleteSession(c.Request.Context(), token); err != nil {
			httpx.Fail(c, err)
			return
		}
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieName, "", -1, "/", "", false, true)
	c.Status(http.StatusNoContent)
}

// me godoc
//
//	@Summary	Текущий пользователь
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	meResponse
//	@Failure	401	{object}	httpx.ErrorBody
//	@Router		/auth/me [get]
func (m *Module) me(c *gin.Context) {
	u := auth.UserFrom(c)
	resp := meResponse{ID: u.ID, FullName: u.FullName, Position: u.Position, Role: u.Role}
	if u.CompanyID != nil {
		cv, err := m.ports.Companies.GetByID(c.Request.Context(), *u.CompanyID)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if cv != nil {
			resp.Company = &companyDTO{ID: cv.ID, INN: cv.INN, Name: cv.Name}
		}
	}
	httpx.JSON(c, http.StatusOK, resp)
}
