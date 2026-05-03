package middleware

import (
	"strings"

	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/security"
	"github.com/golem-mx/core-api/internal/tenancy"
	"github.com/gin-gonic/gin"
)

// RequireAuth protects private API routes.
//
// This middleware is responsible for:
//
//   - Reading the Authorization header
//   - Validating the Bearer token format
//   - Parsing and verifying the JWT access token
//   - Rejecting invalid or expired tokens
//   - Injecting authenticated identity context into Gin
//
// Expected header format:
//
//   Authorization: Bearer <access_token>
//
// Security responsibility:
//
// This middleware is the main gatekeeper for all private routes.
// If this middleware fails open, private tenant data can be exposed.
//
// Important:
//
// This middleware should only validate authentication.
// Permission checks belong to the authorization/RBAC layer.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ---------------------------------------------------------------------
		// Read Authorization header
		// ---------------------------------------------------------------------
		//
		// Private routes require an Authorization header.
		//
		// Missing header means the request is unauthenticated.
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

		if authHeader == "" {
			response.Fail(c, 401, "UNAUTHENTICATED", "Authentication required")
			c.Abort()
			return
		}

		// ---------------------------------------------------------------------
		// Validate Bearer token format
		// ---------------------------------------------------------------------
		//
		// Expected format:
		//
		//   Bearer eyJhbGciOi...
		//
		// We split instead of only TrimPrefix to reject malformed values like:
		//
		//   Bearer
		//   Bearer    token
		//   Token abc
		//   Bearer token extra
		parts := strings.Fields(authHeader)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Fail(c, 401, "INVALID_AUTH_HEADER", "Invalid authorization header")
			c.Abort()
			return
		}

		accessToken := strings.TrimSpace(parts[1])

		if accessToken == "" {
			response.Fail(c, 401, "INVALID_AUTH_HEADER", "Access token is required")
			c.Abort()
			return
		}

		// ---------------------------------------------------------------------
		// Parse and validate access token
		// ---------------------------------------------------------------------
		//
		// ParseAccessToken should validate:
		//
		//   - JWT signature
		//   - expiration
		//   - issuer, if configured
		//   - audience, if configured
		//   - token type = access
		//
		// It should return normalized claims used by the rest of the app.
		claims, err := security.ParseAccessToken(secret, accessToken)
		if err != nil {
			response.Fail(c, 401, "INVALID_TOKEN", "Invalid or expired token")
			c.Abort()
			return
		}

		// ---------------------------------------------------------------------
		// Validate required identity claims
		// ---------------------------------------------------------------------
		//
		// A valid authenticated request must include at least:
		//
		//   - user_id
		//   - company_id
		//
		// branch_id may be optional depending on your SaaS model.
		//
		// Rejecting incomplete claims prevents corrupted or downgraded tokens
		// from reaching private handlers.
		if claims.UserID == "" || claims.CompanyID == "" {
			response.Fail(c, 401, "INVALID_TOKEN_CLAIMS", "Invalid token claims")
			c.Abort()
			return
		}

		// ---------------------------------------------------------------------
		// Inject authenticated tenant context
		// ---------------------------------------------------------------------
		//
		// tenancy.Set stores request-scoped identity values in gin.Context.
		//
		// Downstream layers can use this context for:
		//
		//   - tenant-scoped queries
		//   - authorization checks
		//   - audit logging
		//   - ownership validation
		//
		// Every private repository query should eventually be scoped by
		// company_id to prevent cross-tenant data leaks.
		tenancy.Set(c, claims.UserID, claims.CompanyID, claims.BranchID)

		// ---------------------------------------------------------------------
		// Continue request chain
		// ---------------------------------------------------------------------
		//
		// At this point the request is authenticated.
		// RBAC permission checks can run after this middleware.
		c.Next()
	}
}