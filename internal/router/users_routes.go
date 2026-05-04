package router

import (
	"github.com/gin-gonic/gin"
	"github.com/abrahamVado/go-golem.mx/internal/middleware"
	users "github.com/abrahamVado/go-golem.mx/internal/modules/users"
	rbacmod "github.com/abrahamVado/go-golem.mx/internal/modules/rbac"
)

func registerUserRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	usersH *users.Handler,
) {
	private.GET(
		"/users",
		middleware.RequirePermission(rbac, "users.read"),
		usersH.List,
	)

	private.POST(
		"/users",
		middleware.RequirePermission(rbac, "users.create"),
		usersH.Create,
	)

	private.GET(
		"/users/:id",
		middleware.RequirePermission(rbac, "users.read"),
		usersH.Get,
	)

	private.PATCH(
		"/users/:id",
		middleware.RequirePermission(rbac, "users.update"),
		usersH.Update,
	)

	private.DELETE(
		"/users/:id",
		middleware.RequirePermission(rbac, "users.delete"),
		usersH.Delete,
	)
}