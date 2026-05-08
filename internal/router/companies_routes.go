package router

import (
	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	companies "github.com/abrahamVado/go-paladin.mx/internal/modules/companies"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	"github.com/gin-gonic/gin"
)

func registerCompanyRoutes(
	private *gin.RouterGroup,
	rbac *rbacmod.Service,
	companiesH *companies.Handler,
) {
	private.GET(
		"/companies/current",
		middleware.RequirePermission(
			rbac,
			"organization:view",
		),
		companiesH.Current,
	)

	private.PATCH(
		"/companies/current",
		middleware.RequirePermission(
			rbac,
			"organization:update",
		),
		companiesH.UpdateCurrent,
	)
}
