package router

import (
	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	roles "github.com/abrahamVado/go-paladin.mx/internal/modules/roles"
	"github.com/gin-gonic/gin"
)

func registerRoleRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	rolesH *roles.Handler,
) {
	private.GET(
		"/roles",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.List,
	)

	private.POST(
		"/roles",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.Create,
	)

	private.GET(
		"/roles/:id",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.Get,
	)

	private.GET(
		"/roles/:id/permissions",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.GetPermissions,
	)

	private.PATCH(
		"/roles/:id",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.Update,
	)

	private.DELETE(
		"/roles/:id",
		middleware.RequirePermission(rbac, "role:manage"),
		rolesH.Delete,
	)
}
