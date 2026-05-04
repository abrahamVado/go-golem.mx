package router

import (
	"github.com/abrahamVado/go-golem.mx/internal/modules/auth"
	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------------
// AUTHENTICATION ROUTES REGISTRATION
// -----------------------------------------------------------------------------
//
// This file defines the HTTP routes for authentication and user session
// management.
//
// Design philosophy:
//
//   - This file ONLY registers routes.
//   - It must NOT contain business logic.
//   - It must NOT access the database.
//   - It must NOT implement authorization rules.
//
// Responsibilities:
//
//   - Define endpoint paths
//   - Attach handlers
//   - Define public vs private boundaries
//
// The actual logic lives in:
//
//   Handler  -> HTTP translation
//   Service  -> business rules
//   Repository -> database access
//
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// registerPublicAuthRoutes
// -----------------------------------------------------------------------------
//
// Registers PUBLIC authentication endpoints.
//
// These routes DO NOT require authentication.
//
// Security expectations:
//
//   - Rate-limited
//   - Logged
//   - Protected against brute-force attacks
//   - Protected against enumeration attacks
//
// These endpoints are typically exposed to:
//
//   - Web frontend
//   - Mobile applications
//   - External integrations
//
// Important:
//
// These routes must NEVER leak sensitive information such as:
//
//   - Whether a user exists
//   - Whether an email is registered
//   - Whether a password reset token is valid
//
// Future improvements:
//
//   - CAPTCHA integration
//   - Login attempt throttling
//   - IP reputation checks
//   - Device fingerprinting
//
func registerPublicAuthRoutes(api *gin.RouterGroup, authH *auth.Handler) {

	// -------------------------------------------------------------------------
	// User registration
	// -------------------------------------------------------------------------
	//
	// Creates a new user account.
	//
	// Expected behavior:
	//
	//   - Validate email format
	//   - Validate password strength
	//   - Hash password
	//   - Create user record
	//   - Assign default role
	//   - Emit audit log
	//
	api.POST("/auth/register", authH.Register)

	// -------------------------------------------------------------------------
	// User login
	// -------------------------------------------------------------------------
	//
	// Authenticates a user and returns access and refresh tokens.
	//
	// Security requirements:
	//
	//   - Constant-time password comparison
	//   - Account lock after repeated failures
	//   - Audit logging
	//
	api.POST("/auth/login", authH.Login)

	// -------------------------------------------------------------------------
	// Token refresh
	// -------------------------------------------------------------------------
	//
	// Issues a new access token using a valid refresh token.
	//
	// Recommended behavior:
	//
	//   - Rotate refresh tokens
	//   - Invalidate previous refresh tokens
	//   - Track token device/session
	//
	api.POST("/auth/refresh", authH.Refresh)

	// -------------------------------------------------------------------------
	// Password recovery request
	// -------------------------------------------------------------------------
	//
	// Initiates password reset flow.
	//
	// Must NOT reveal:
	//
	//   - Whether email exists
	//
	// Always return success response.
	//
	api.POST("/auth/recover", authH.Recover)

	// -------------------------------------------------------------------------
	// Password reset
	// -------------------------------------------------------------------------
	//
	// Completes password reset using reset token.
	//
	// Required validations:
	//
	//   - Token validity
	//   - Token expiration
	//   - Password strength
	//
	api.POST("/auth/reset-password", authH.ResetPassword)
}

// -----------------------------------------------------------------------------
// registerPrivateAuthRoutes
// -----------------------------------------------------------------------------
//
// Registers PRIVATE authentication endpoints.
//
// These routes REQUIRE authentication.
//
// Middleware expected:
//
//   RequireAuth()
//
// This middleware must:
//
//   - Validate JWT signature
//   - Validate expiration
//   - Extract user identity
//   - Inject claims into context
//
// These endpoints operate on the CURRENT authenticated user.
// No user ID should be required in the request.
//
// Future extensions:
//
//   - Session management
//   - Device management
//   - MFA / 2FA
//   - Security events
//
func registerPrivateAuthRoutes(private *gin.RouterGroup, authH *auth.Handler) {

	// -------------------------------------------------------------------------
	// Logout
	// -------------------------------------------------------------------------
	//
	// Terminates the current session.
	//
	// Recommended behavior:
	//
	//   - Invalidate refresh token
	//   - Record logout event
	//   - Clear session metadata
	//
	private.POST("/auth/logout", authH.Logout)

	// -------------------------------------------------------------------------
	// Get current user profile
	// -------------------------------------------------------------------------
	//
	// Returns information about the authenticated user.
	//
	// Data typically returned:
	//
	//   - User ID
	//   - Email
	//   - Roles
	//   - Permissions
	//   - Tenant / Company
	//
	private.GET("/me", authH.Me)

	// -------------------------------------------------------------------------
	// Update current user profile
	// -------------------------------------------------------------------------
	//
	// Allows the authenticated user to update their own profile.
	//
	// Allowed updates typically include:
	//
	//   - Name
	//   - Avatar
	//   - Phone
	//   - Preferences
	//
	// Forbidden updates:
	//
	//   - Roles
	//   - Permissions
	//   - Tenant ownership
	//
	private.PATCH("/me", authH.UpdateMe)
}