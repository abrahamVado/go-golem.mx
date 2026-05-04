package router

import (
	"github.com/gin-gonic/gin"
	"github.com/abrahamVado/go-golem.mx/internal/middleware"
	companies "github.com/abrahamVado/go-golem.mx/internal/modules/companies"
	rbacmod "github.com/abrahamVado/go-golem.mx/internal/modules/rbac"
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
			"companies.read",
		),
		companiesH.Current,
	)

	private.PATCH(
		"/companies/current",
		middleware.RequirePermission(
			rbac,
			"companies.update",
		),
		companiesH.UpdateCurrent,
	)
}