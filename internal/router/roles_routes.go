package router

import (
	"github.com/gin-gonic/gin"
	"github.com/abrahamVado/go-golem.mx/internal/middleware"
	roles "github.com/abrahamVado/go-golem.mx/internal/modules/roles"
	rbacmod "github.com/abrahamVado/go-golem.mx/internal/modules/rbac"
)

func registerRoleRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	rolesH *roles.Handler,
) {
	private.GET(
		"/roles",
		middleware.RequirePermission(rbac, "roles.read"),
		rolesH.List,
	)

	private.POST(
		"/roles",
		middleware.RequirePermission(rbac, "roles.create"),
		rolesH.Create,
	)

	private.GET(
		"/roles/:id",
		middleware.RequirePermission(rbac, "roles.read"),
		rolesH.Get,
	)

	private.PATCH(
		"/roles/:id",
		middleware.RequirePermission(rbac, "roles.update"),
		rolesH.Update,
	)

	private.DELETE(
		"/roles/:id",
		middleware.RequirePermission(rbac, "roles.delete"),
		rolesH.Delete,
	)
}