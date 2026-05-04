package auth

import (
	"net/http"

	"github.com/abrahamVado/go-golem.mx/internal/platform/config"
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/abrahamVado/go-golem.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

// Handler owns the HTTP layer for authentication.
//
// Responsibilities:
//
//   - bind request JSON
//   - call auth service methods
//   - set/clear refresh cookies
//   - return standardized API responses
//
// This layer should NOT:
//
//   - hash passwords
//   - query the database
//   - create JWTs directly
//   - implement business rules
type Handler struct {
	svc *Service
	cfg config.Config
}

// NewHandler creates the auth HTTP handler.
func NewHandler(s *Service, cfg config.Config) *Handler {
	return &Handler{
		svc: s,
		cfg: cfg,
	}
}

// setRefreshCookie stores the refresh token in a secure HTTP-only cookie.
//
// Security notes:
//
//   - HttpOnly prevents JavaScript access.
//   - Secure should be true in production.
//   - SameSite=Lax reduces CSRF risk while keeping normal navigation usable.
//   - Path limits where the browser sends the cookie.
//
// Important:
//
// The raw refresh token is sent to the browser only as a cookie.
// The database should store only its hash.
func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Domain:   h.cfg.CookieDomain,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.RefreshTokenTTL.Seconds()),
	})
}

// clearRefreshCookie removes the refresh cookie from the browser.
func (h *Handler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   h.cfg.CookieDomain,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// Register handles tenant/user registration.
//
// Future behavior:
//
//   - create company
//   - create owner user
//   - create/attach Owner role
//   - assign owner permissions
//   - run everything inside a DB transaction
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Currently scaffolded in the service.
	_, err := h.svc.Register(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"message": "registered",
	})
}

// Login authenticates a user and returns an access token.
//
// Refresh token behavior:
//
//   - returned as secure HTTP-only cookie
//   - not returned in JSON body
//
// JSON response contains only the access token payload.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	out, refresh, err := h.svc.Login(
		req.Email,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	h.setRefreshCookie(c, refresh)

	response.OK(c, out)
}

// Refresh rotates the refresh token and returns a new access token.
//
// Token rotation:
//
//   - old refresh token is revoked
//   - new refresh token is stored as hash
//   - browser receives new refresh cookie
func (h *Handler) Refresh(c *gin.Context) {
	refreshCookie, err := c.Cookie("refresh_token")
	if err != nil || refreshCookie == "" {
		response.Unauthorized(c, "Refresh token required")
		return
	}

	out, refresh, err := h.svc.Refresh(
		refreshCookie,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		h.clearRefreshCookie(c)
		response.Unauthorized(c, "Invalid refresh token")
		return
	}

	h.setRefreshCookie(c, refresh)

	response.OK(c, out)
}

// Logout revokes the current refresh token and clears the cookie.
//
// Logout should be idempotent:
//
//   - missing cookie still returns success
//   - already-revoked token still clears cookie
func (h *Handler) Logout(c *gin.Context) {
	refreshCookie, _ := c.Cookie("refresh_token")

	if refreshCookie != "" {
		_ = h.svc.Logout(refreshCookie)
	}

	h.clearRefreshCookie(c)

	response.OK(c, gin.H{
		"message": "logged out",
	})
}

// Recover starts the password recovery flow.
//
// Security rule:
//
// Always return the same response whether the email exists or not.
// This prevents account enumeration.
func (h *Handler) Recover(c *gin.Context) {
	response.OK(c, gin.H{
		"message": "If the account exists, reset instructions will be sent",
	})
}

// ResetPassword completes password reset.
//
// Future behavior:
//
//   - validate reset token
//   - validate new password
//   - update password hash
//   - revoke active sessions
func (h *Handler) ResetPassword(c *gin.Context) {
	response.OK(c, gin.H{
		"message": "Password reset scaffolded",
	})
}

// Me returns the current authenticated identity.
//
// Source:
//
//   tenancy context injected by middleware.RequireAuth().
func (h *Handler) Me(c *gin.Context) {
	response.OK(c, gin.H{
		"user_id":    tenancy.UserID(c),
		"company_id": tenancy.CompanyID(c),
		"branch_id":  tenancy.BranchID(c),
	})
}

// UpdateMe updates the current authenticated user's profile.
//
// Future allowed fields:
//
//   - name
//   - avatar
//   - phone
//   - preferences
//
// Forbidden fields:
//
//   - roles
//   - permissions
//   - company_id
func (h *Handler) UpdateMe(c *gin.Context) {
	response.OK(c, gin.H{
		"message": "profile updated scaffold",
	})
}