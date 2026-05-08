package dashboard

import (
	"github.com/abrahamVado/go-golem.mx/internal/middleware"
	rbacmod "github.com/abrahamVado/go-golem.mx/internal/modules/rbac"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(private *gin.RouterGroup, rbac *rbacmod.Service, handler *Handler) {
	private.GET("/dashboard/summary", middleware.RequirePermission(rbac, "organization:view"), handler.Index)
	private.GET("/dashboard/system-logs", middleware.RequirePermission(rbac, "organization:view"), handler.SystemLogs)
}
