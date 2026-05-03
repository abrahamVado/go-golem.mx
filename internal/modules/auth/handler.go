package auth

import (
	"github.com/golem-mx/core-api/internal/config"
	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/tenancy"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	svc *Service
	cfg config.Config
}

func NewHandler(s *Service, cfg config.Config) *Handler { return &Handler{svc: s, cfg: cfg} }
func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "refresh_token", Value: token, Path: "/api/v1/auth", Domain: h.cfg.CookieDomain, HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: int(h.cfg.RefreshTokenTTL.Seconds())})
}
func (h *Handler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth", Domain: h.cfg.CookieDomain, HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "VALIDATION_ERROR", "Invalid request")
		return
	}
	response.Created(c, gin.H{"message": "register flow scaffolded; implement transaction in service"})
}
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "VALIDATION_ERROR", "Invalid request")
		return
	}
	out, refresh, err := h.svc.Login(req.Email, req.Password, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.Fail(c, 401, "INVALID_CREDENTIALS", "Invalid credentials")
		return
	}
	h.setRefreshCookie(c, refresh)
	response.OK(c, out)
}
func (h *Handler) Refresh(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Fail(c, 401, "REFRESH_REQUIRED", "Refresh token required")
		return
	}
	out, refresh, err := h.svc.Refresh(cookie, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.Fail(c, 401, "INVALID_REFRESH_TOKEN", "Invalid refresh token")
		return
	}
	h.setRefreshCookie(c, refresh)
	response.OK(c, out)
}
func (h *Handler) Logout(c *gin.Context) {
	cookie, _ := c.Cookie("refresh_token")
	if cookie != "" {
		_ = h.svc.Logout(cookie)
	}
	h.clearRefreshCookie(c)
	response.OK(c, gin.H{"message": "logged out"})
}
func (h *Handler) Recover(c *gin.Context) {
	response.OK(c, gin.H{"message": "If the account exists, reset instructions will be sent"})
}
func (h *Handler) ResetPassword(c *gin.Context) {
	response.OK(c, gin.H{"message": "Password reset scaffolded"})
}
func (h *Handler) Me(c *gin.Context) {
	response.OK(c, gin.H{"user_id": tenancy.UserID(c), "company_id": tenancy.CompanyID(c)})
}
func (h *Handler) UpdateMe(c *gin.Context) {
	response.OK(c, gin.H{"message": "profile updated scaffold"})
}
