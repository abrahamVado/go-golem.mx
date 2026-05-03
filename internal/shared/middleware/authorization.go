package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golem-mx/core-api/internal/modules/rbac"
	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/tenancy"
)

// RequirePermission protects a route with one required permission.
func RequirePermission(s *rbac.Service, permission string) gin.HandlerFunc {
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
func RequireAnyPermission(s *rbac.Service, perms ...string) gin.HandlerFunc {
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
func RequireRole(s *rbac.Service, roles ...string) gin.HandlerFunc {
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
