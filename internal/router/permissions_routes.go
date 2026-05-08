package router

import (
	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	permissions "github.com/abrahamVado/go-paladin.mx/internal/modules/permissions"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	"github.com/gin-gonic/gin"
)

func registerPermissionRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	permissionsH *permissions.Handler,
) {
	private.GET(
		"/permissions",
		middleware.RequirePermission(rbac, "role:manage"),
		permissionsH.List,
	)

	private.POST(
		"/permissions",
		middleware.RequirePermission(rbac, "role:manage"),
		permissionsH.Create,
	)
}
