package router

import (
	"github.com/golem-mx/core-api/internal/authorization"
	"github.com/golem-mx/core-api/internal/modules/companies"
	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------------
// COMPANY / TENANT ROUTES REGISTRATION
// -----------------------------------------------------------------------------
//
// This file defines routes related to the Company (Tenant) entity.
//
// In a multi-tenant SaaS architecture:
//
//   Company == Tenant
//
// Every authenticated user belongs to exactly one company (tenant),
// and all data access must be scoped to that tenant.
//
// This module is security-critical because:
//
//   - It defines tenant identity
//   - It controls tenant configuration
//   - It affects data isolation boundaries
//
// Responsibilities:
//
//   - Register tenant-scoped routes
//   - Enforce RBAC permissions
//   - Prevent cross-tenant access
//
// This file MUST NOT:
//
//   - Query the database
//   - Contain business logic
//   - Perform authorization logic
//
// Those belong to:
//
//   Repository  -> persistence
//   Service     -> business rules
//   Middleware  -> authentication / authorization
//
// -----------------------------------------------------------------------------

// registerCompanyRoutes registers tenant-scoped company endpoints.
//
// All routes in this function are:
//
//   - Authenticated
//   - Tenant-scoped
//   - Permission-protected
//
// Middleware expected upstream:
//
//   RequireAuth()
//   ResolveTenant()
//
// ResolveTenant() should:
//
//   - Extract company_id from token/session
//   - Validate tenant existence
//   - Inject tenant context into request
//
// IMPORTANT SECURITY RULE:
//
// Never allow arbitrary company IDs in normal user routes.
// Always use the CURRENT tenant derived from authentication.
//
func registerCompanyRoutes(
	private *gin.RouterGroup,
	rbac *authorization.Service,
	companiesH *companies.Handler,
) {

	// -------------------------------------------------------------------------
	// Get current company (tenant)
	// -------------------------------------------------------------------------
	//
	// Returns the company associated with the authenticated user.
	//
	// This endpoint intentionally avoids:
	//
	//   /companies/:id
	//
	// Because exposing arbitrary IDs can allow:
	//
	//   - horizontal privilege escalation
	//   - tenant data leaks
	//
	// Instead, the system resolves the tenant automatically
	// from the authentication context.
	//
	// Typical response data:
	//
	//   - company_id
	//   - company_name
	//   - subscription_plan
	//   - billing_status
	//   - feature_flags
	//
	// Required permission:
	//
	//   companies.read
	//
	private.GET(
		"/companies/current",
		authorization.RequirePermission(
			rbac,
			"companies.read",
		),
		companiesH.Current,
	)

	// -------------------------------------------------------------------------
	// Update current company (tenant)
	// -------------------------------------------------------------------------
	//
	// Updates configuration for the authenticated user's company.
	//
	// Allowed updates typically include:
	//
	//   - company_name
	//   - branding settings
	//   - timezone
	//   - contact information
	//   - feature configuration
	//
	// Forbidden updates should include:
	//
	//   - company_id
	//   - subscription plan
	//   - billing status
	//   - system flags
	//
	// Those fields should only be modified by:
	//
	//   - billing services
	//   - system administrators
	//   - background jobs
	//
	// Required permission:
	//
	//   companies.update
	//
	private.PATCH(
		"/companies/current",
		authorization.RequirePermission(
			rbac,
			"companies.update",
		),
		companiesH.UpdateCurrent,
	)
}