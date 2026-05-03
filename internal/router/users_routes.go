package router

import (
	"github.com/golem-mx/core-api/internal/authorization"
	"github.com/golem-mx/core-api/internal/modules/users"
	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------------
// USER MANAGEMENT ROUTES REGISTRATION
// -----------------------------------------------------------------------------
//
// This file defines routes responsible for managing User entities.
//
// In a multi-tenant RBAC system:
//
//   User is the principal identity interacting with the system.
//
// Users are security-sensitive resources because they represent:
//
//   - Authentication identities
//   - Access permissions
//   - Audit accountability
//
// Responsibilities:
//
//   - Register user management endpoints
//   - Enforce RBAC permission checks
//   - Maintain tenant isolation boundaries
//
// This file MUST NOT:
//
//   - Contain business logic
//   - Query the database
//   - Implement authorization decisions
//
// Those responsibilities belong to:
//
//   Handler        -> HTTP interface
//   Service        -> business logic
//   Repository     -> persistence
//   Authorization  -> permission validation
//
// -----------------------------------------------------------------------------

// registerUserRoutes registers all user-related endpoints.
//
// All routes in this function are:
//
//   - Authenticated
//   - Tenant-scoped
//   - Permission-protected
//
// Required upstream middleware:
//
//   RequireAuth()
//   ResolveTenant()
//
// Security rules:
//
//   - Users must belong to the current tenant
//   - Cross-tenant access must never be allowed
//   - Sensitive changes must be audit logged
//
func registerUserRoutes(
	private *gin.RouterGroup,
	rbac *authorization.Service,
	usersH *users.Handler,
) {

	// -------------------------------------------------------------------------
	// List users
	// -------------------------------------------------------------------------
	//
	// Returns users belonging to the current tenant.
	//
	// Expected behavior:
	//
	//   - Filter by tenant_id
	//   - Support pagination
	//   - Support search/filtering
	//   - Exclude soft-deleted users
	//
	// Typical consumers:
	//
	//   - Admin user management UI
	//   - Role assignment workflows
	//
	// Required permission:
	//
	//   users.read
	//
	private.GET(
		"/users",
		authorization.RequirePermission(rbac, "users.read"),
		usersH.List,
	)

	// -------------------------------------------------------------------------
	// Create user
	// -------------------------------------------------------------------------
	//
	// Creates a new user within the current tenant.
	//
	// Expected validations:
	//
	//   - Email uniqueness within tenant
	//   - Password strength validation
	//   - Default role assignment
	//   - Tenant ownership enforcement
	//
	// Security considerations:
	//
	//   - Hash password before storage
	//   - Prevent privilege escalation
	//   - Validate assigned roles
	//   - Log user creation event
	//
	// Required permission:
	//
	//   users.create
	//
	private.POST(
		"/users",
		authorization.RequirePermission(rbac, "users.create"),
		usersH.Create,
	)

	// -------------------------------------------------------------------------
	// Get user by ID
	// -------------------------------------------------------------------------
	//
	// Returns details for a specific user.
	//
	// Expected behavior:
	//
	//   - Validate user belongs to current tenant
	//   - Return profile and role information
	//   - Exclude sensitive fields
	//
	// Sensitive fields that must never be returned:
	//
	//   - password_hash
	//   - password_reset_token
	//   - internal_security_flags
	//
	// Required permission:
	//
	//   users.read
	//
	private.GET(
		"/users/:id",
		authorization.RequirePermission(rbac, "users.read"),
		usersH.Get,
	)

	// -------------------------------------------------------------------------
	// Update user
	// -------------------------------------------------------------------------
	//
	// Modifies an existing user's profile or role assignment.
	//
	// Allowed updates typically include:
	//
	//   - name
	//   - email
	//   - role assignments
	//   - status (active/inactive)
	//
	// Forbidden updates should include:
	//
	//   - tenant ownership
	//   - system identity fields
	//
	// Security considerations:
	//
	//   - Validate role changes
	//   - Prevent self-demotion from required roles
	//   - Record audit trail
	//
	// Required permission:
	//
	//   users.update
	//
	private.PATCH(
		"/users/:id",
		authorization.RequirePermission(rbac, "users.update"),
		usersH.Update,
	)

	// -------------------------------------------------------------------------
	// Delete user
	// -------------------------------------------------------------------------
	//
	// Removes a user from the system.
	//
	// Recommended implementation behavior:
	//
	//   - Use soft delete instead of hard delete
	//   - Prevent deletion of system users
	//   - Prevent self-deletion if disallowed
	//   - Preserve audit history
	//
	// Security considerations:
	//
	//   - Validate user ownership
	//   - Log deletion event
	//
	// Required permission:
	//
	//   users.delete
	//
	private.DELETE(
		"/users/:id",
		authorization.RequirePermission(rbac, "users.delete"),
		usersH.Delete,
	)
}