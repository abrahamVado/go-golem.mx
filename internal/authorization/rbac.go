package authorization

import (
	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/tenancy"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Service provides authorization checks for authenticated users.
//
// This service answers questions like:
//
//   - Does this user have this permission?
//   - Does this user have one of these roles?
//
// Important:
//
// This service does NOT authenticate users.
// Authentication belongs to middleware.RequireAuth().
//
// This service only runs AFTER authentication has already injected:
//
//   - user_id
//   - company_id
//   - branch_id, optional
//
// into the request context.
type Service struct {
	DB *gorm.DB
}

// New creates a new authorization service.
func New(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

// UserHasPermission checks whether a user has a specific permission
// inside the current company / tenant.
//
// Tenant safety rule:
//
//   Authorization must always be scoped by company_id.
//
// Without company scoping, a user could accidentally receive permissions
// from another tenant.
//
// Owner bypass:
//
//   Users with the "Owner" role automatically pass permission checks.
//
// Recommended future improvement:
//
//   Replace hardcoded "Owner" with a constant:
//
//     const SystemRoleOwner = "Owner"
//
func (s *Service) UserHasPermission(userID, companyID any, permission string) bool {
	if userID == nil || companyID == nil || permission == "" {
		return false
	}

	var count int64

	err := s.DB.
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("LEFT JOIN permissions p ON p.id = rp.permission_id").
		Where(
			"ur.user_id = ? AND ur.company_id = ? AND ur.deleted_at IS NULL",
			userID,
			companyID,
		).
		Where(
			s.DB.Where("p.name = ?", permission).Or("r.name = ?", "Owner"),
		).
		Count(&count).
		Error

	if err != nil {
		return false
	}

	return count > 0
}

// UserHasRole checks whether a user has at least one of the provided roles
// inside the current company / tenant.
//
// Example:
//
//   UserHasRole(userID, companyID, "Owner", "Admin")
//
// Empty role lists always return false.
func (s *Service) UserHasRole(userID, companyID any, roles ...string) bool {
	if userID == nil || companyID == nil || len(roles) == 0 {
		return false
	}

	var count int64

	err := s.DB.
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where(
			"ur.user_id = ? AND ur.company_id = ? AND ur.deleted_at IS NULL",
			userID,
			companyID,
		).
		Where("r.name IN ?", roles).
		Count(&count).
		Error

	if err != nil {
		return false
	}

	return count > 0
}

// RequirePermission protects a route with one required permission.
//
// Example:
//
//   private.GET(
//       "/users",
//       authorization.RequirePermission(rbac, "users.read"),
//       usersH.List,
//   )
//
// Behavior:
//
//   - If user has permission: continue request
//   - If user does not have permission: return 403
//
// Important:
//
// This middleware assumes authentication already ran.
func RequirePermission(s *Service, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.UserHasPermission(
			tenancy.UserID(c),
			tenancy.CompanyID(c),
			permission,
		) {
			response.Fail(c, 403, "FORBIDDEN", "Missing required permission")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission protects a route with multiple possible permissions.
//
// The user is allowed through if they have AT LEAST ONE permission.
//
// Example:
//
//   RequireAnyPermission(rbac, "users.read", "users.manage")
//
func RequireAnyPermission(s *Service, perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(perms) == 0 {
			response.Fail(c, 403, "FORBIDDEN", "Missing required permission")
			c.Abort()
			return
		}

		for _, permission := range perms {
			if s.UserHasPermission(
				tenancy.UserID(c),
				tenancy.CompanyID(c),
				permission,
			) {
				c.Next()
				return
			}
		}

		response.Fail(c, 403, "FORBIDDEN", "Missing required permission")
		c.Abort()
	}
}

// RequireRole protects a route with one or more allowed roles.
//
// The user is allowed through if they have AT LEAST ONE of the provided roles.
//
// Example:
//
//   RequireRole(rbac, "Owner", "Admin")
//
// Prefer permissions over roles for most API endpoints.
// Roles are better for broad administrative gates.
func RequireRole(s *Service, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.UserHasRole(
			tenancy.UserID(c),
			tenancy.CompanyID(c),
			roles...,
		) {
			response.Fail(c, 403, "FORBIDDEN", "Missing required role")
			c.Abort()
			return
		}

		c.Next()
	}
}