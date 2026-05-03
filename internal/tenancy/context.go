package tenancy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// TENANCY CONTEXT MANAGEMENT
// -----------------------------------------------------------------------------
//
// This package manages request-scoped tenant identity.
//
// Every authenticated request must carry:
//
//   - user_id
//   - company_id
//
// Optional:
//
//   - branch_id
//
// These values are injected by authentication middleware and used by:
//
//   - repositories
//   - authorization checks
//   - audit logging
//   - tenant isolation enforcement
//
// Design goals:
//
//   - predictable behavior
//   - safe type handling
//   - no silent cross-tenant access
//
// -----------------------------------------------------------------------------

// Context keys.
//
// Using constants prevents typos and ensures consistency across packages.
const (
	CompanyIDKey = "company_id"
	UserIDKey    = "user_id"
	BranchIDKey  = "branch_id"
)

// Set stores authenticated identity values in the request context.
//
// This function is typically called from:
//
//   middleware.RequireAuth()
//
// After successful token validation.
//
// Important:
//
// These values must be set exactly once per request.
func Set(
	c *gin.Context,
	userID uuid.UUID,
	companyID uuid.UUID,
	branchID *uuid.UUID,
) {
	c.Set(UserIDKey, userID)
	c.Set(CompanyIDKey, companyID)

	if branchID != nil {
		c.Set(BranchIDKey, *branchID)
	}
}

// -----------------------------------------------------------------------------
// SAFE GETTERS
// -----------------------------------------------------------------------------

// CompanyID returns the tenant company ID.
//
// If the value is missing or invalid, uuid.Nil is returned.
//
// Recommended usage:
//
//   companyID := tenancy.CompanyID(c)
//
// Always validate uuid.Nil in critical code paths.
func CompanyID(c *gin.Context) uuid.UUID {
	v, exists := c.Get(CompanyIDKey)
	if !exists {
		return uuid.Nil
	}

	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return id
}

// UserID returns the authenticated user ID.
//
// If missing, uuid.Nil is returned.
func UserID(c *gin.Context) uuid.UUID {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil
	}

	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return id
}

// BranchID returns the optional branch ID.
//
// Behavior:
//
//   - returns nil if not set
//   - returns nil if invalid
//
// This allows branch-level scoping to remain optional.
func BranchID(c *gin.Context) *uuid.UUID {
	v, exists := c.Get(BranchIDKey)
	if !exists {
		return nil
	}

	id, ok := v.(uuid.UUID)
	if !ok {
		return nil
	}

	return &id
}

// -----------------------------------------------------------------------------
// STRICT VARIANTS (OPTIONAL)
// -----------------------------------------------------------------------------
//
// These helpers panic if required context is missing.
// Useful for:
//
//   - repository layer
//   - internal service logic
//   - debugging misconfigured middleware
//
// Not recommended for public handlers.
//

func MustCompanyID(c *gin.Context) uuid.UUID {
	id := CompanyID(c)
	if id == uuid.Nil {
		panic("company_id missing from context")
	}
	return id
}

func MustUserID(c *gin.Context) uuid.UUID {
	id := UserID(c)
	if id == uuid.Nil {
		panic("user_id missing from context")
	}
	return id
}