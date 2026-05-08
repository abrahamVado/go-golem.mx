package router

import (
	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	users "github.com/abrahamVado/go-paladin.mx/internal/modules/users"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	usersH *users.Handler,
) {
	private.GET(
		"/users",
		middleware.RequirePermission(rbac, "member:update"),
		usersH.List,
	)

	private.POST(
		"/users",
		middleware.RequirePermission(rbac, "member:invite"),
		usersH.Create,
	)

	private.GET(
		"/users/:id",
		middleware.RequirePermission(rbac, "member:update"),
		usersH.Get,
	)

	private.PATCH(
		"/users/:id",
		middleware.RequirePermission(rbac, "member:update"),
		usersH.Update,
	)

	private.DELETE(
		"/users/:id",
		middleware.RequirePermission(rbac, "member:remove"),
		usersH.Delete,
	)
}
