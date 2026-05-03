package router

import (
	"github.com/golem-mx/core-api/internal/authorization"
	"github.com/golem-mx/core-api/internal/modules/roles"
	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------------
// ROLE MANAGEMENT ROUTES REGISTRATION
// -----------------------------------------------------------------------------
//
// This file defines routes for managing Roles in the RBAC system.
//
// Roles are collections of permissions that define what actions a user
// is allowed to perform in the system.
//
// Example:
//
//   Role: "Admin"
//   Permissions:
//     - users.read
//     - users.create
//     - roles.update
//
// Roles are one of the most security-sensitive resources in the platform.
// Misconfiguration can lead to privilege escalation or data exposure.
//
// Responsibilities:
//
//   - Register role management endpoints
//   - Enforce permission checks via RBAC middleware
//   - Maintain tenant isolation
//
// This file MUST NOT:
//
//   - Execute business logic
//   - Query the database
//   - Perform authorization decisions
//
// Those belong to:
//
//   Handler     -> HTTP interface
//   Service     -> business logic
//   Repository  -> persistence
//   Authorization -> permission validation
//
// -----------------------------------------------------------------------------

// registerRoleRoutes registers all role-related endpoints.
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
//   - Roles must belong to the current tenant
//   - Cross-tenant role access must never be allowed
//   - System roles should be protected from modification
//
func registerRoleRoutes(
	private *gin.RouterGroup,
	rbac *authorization.Service,
	rolesH *roles.Handler,
) {

	// -------------------------------------------------------------------------
	// List roles
	// -------------------------------------------------------------------------
	//
	// Returns all roles available within the current tenant.
	//
	// Typical use cases:
	//
	//   - Admin panel role management
	//   - Permission assignment UI
	//   - User onboarding workflows
	//
	// Expected behavior:
	//
	//   - Filter roles by tenant
	//   - Exclude deleted roles
	//   - Support pagination
	//   - Support search/filtering
	//
	// Required permission:
	//
	//   roles.read
	//
	private.GET(
		"/roles",
		authorization.RequirePermission(rbac, "roles.read"),
		rolesH.List,
	)

	// -------------------------------------------------------------------------
	// Create role
	// -------------------------------------------------------------------------
	//
	// Creates a new role within the current tenant.
	//
	// Expected validations:
	//
	//   - Role name uniqueness within tenant
	//   - Permission list validation
	//   - Role scope enforcement
	//
	// Security considerations:
	//
	//   - Prevent creation of system-reserved role names
	//   - Validate permissions against allowed set
	//
	// Required permission:
	//
	//   roles.create
	//
	private.POST(
		"/roles",
		authorization.RequirePermission(rbac, "roles.create"),
		rolesH.Create,
	)

	// -------------------------------------------------------------------------
	// Get role by ID
	// -------------------------------------------------------------------------
	//
	// Returns details for a specific role.
	//
	// Expected behavior:
	//
	//   - Validate role belongs to current tenant
	//   - Return associated permissions
	//   - Return metadata (created_at, updated_at)
	//
	// Security rule:
	//
	//   Never expose roles from other tenants.
	//
	// Required permission:
	//
	//   roles.read
	//
	private.GET(
		"/roles/:id",
		authorization.RequirePermission(rbac, "roles.read"),
		rolesH.Get,
	)

	// -------------------------------------------------------------------------
	// Update role
	// -------------------------------------------------------------------------
	//
	// Modifies an existing role.
	//
	// Allowed updates typically include:
	//
	//   - Role name
	//   - Permission assignments
	//
	// Forbidden updates should include:
	//
	//   - Tenant ownership
	//   - System role flags
	//
	// Security considerations:
	//
	//   - Prevent modification of protected system roles
	//   - Validate permission changes
	//   - Log permission changes in audit log
	//
	// Required permission:
	//
	//   roles.update
	//
	private.PATCH(
		"/roles/:id",
		authorization.RequirePermission(rbac, "roles.update"),
		rolesH.Update,
	)

	// -------------------------------------------------------------------------
	// Delete role
	// -------------------------------------------------------------------------
	//
	// Removes a role from the system.
	//
	// Recommended implementation behavior:
	//
	//   - Use soft delete
	//   - Prevent deletion of system roles
	//   - Prevent deletion of roles currently assigned to users
	//
	// Security considerations:
	//
	//   - Log deletion events
	//   - Maintain referential integrity
	//
	// Required permission:
	//
	//   roles.delete
	//
	private.DELETE(
		"/roles/:id",
		authorization.RequirePermission(rbac, "roles.delete"),
		rolesH.Delete,
	)
}