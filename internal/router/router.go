package router

import (
	"github.com/example/gin-multitenant-backend/internal/authorization"
	"github.com/example/gin-multitenant-backend/internal/config"
	"github.com/example/gin-multitenant-backend/internal/middleware"
	auditmod "github.com/example/gin-multitenant-backend/internal/modules/audit"
	authmod "github.com/example/gin-multitenant-backend/internal/modules/auth"
	branchesmod "github.com/example/gin-multitenant-backend/internal/modules/branches"
	companiesmod "github.com/example/gin-multitenant-backend/internal/modules/companies"
	dashboardmod "github.com/example/gin-multitenant-backend/internal/modules/dashboard"
	permissionsmod "github.com/example/gin-multitenant-backend/internal/modules/permissions"
	rolesmod "github.com/example/gin-multitenant-backend/internal/modules/roles"
	settingsmod "github.com/example/gin-multitenant-backend/internal/modules/settings"
	usersmod "github.com/example/gin-multitenant-backend/internal/modules/users"
	"github.com/example/gin-multitenant-backend/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.RequestID(), middleware.SecureHeaders(), middleware.CORS(cfg.FrontendURL), middleware.RateLimit())
	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) { response.OK(c, gin.H{"status": "ok"}) })

	rbac := authorization.New(db)
	authH := authmod.NewHandler(authmod.NewService(authmod.NewRepository(db), cfg), cfg)
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/refresh", authH.Refresh)
	api.POST("/auth/recover", authH.Recover)
	api.POST("/auth/reset-password", authH.ResetPassword)

	private := api.Group("")
	private.Use(middleware.RequireAuth(cfg.JWTAccessSecret))
	private.POST("/auth/logout", authH.Logout)
	private.GET("/me", authH.Me)
	private.PATCH("/me", authH.UpdateMe)

	private.GET("/dashboard", authorization.RequirePermission(rbac, "dashboard.read"), dashboardmod.NewHandler(dashboardmod.NewService()).Index)

	usersH := usersmod.NewHandler(usersmod.NewService(usersmod.NewRepository(db)))
	private.GET("/users", authorization.RequirePermission(rbac, "users.read"), usersH.List)
	private.POST("/users", authorization.RequirePermission(rbac, "users.create"), usersH.Create)
	private.GET("/users/:id", authorization.RequirePermission(rbac, "users.read"), usersH.Get)
	private.PATCH("/users/:id", authorization.RequirePermission(rbac, "users.update"), usersH.Update)
	private.DELETE("/users/:id", authorization.RequirePermission(rbac, "users.delete"), usersH.Delete)

	rolesH := rolesmod.NewHandler(rolesmod.NewService(rolesmod.NewRepository(db)))
	private.GET("/roles", authorization.RequirePermission(rbac, "roles.read"), rolesH.List)
	private.POST("/roles", authorization.RequirePermission(rbac, "roles.create"), rolesH.Create)
	private.GET("/roles/:id", authorization.RequirePermission(rbac, "roles.read"), rolesH.Get)
	private.PATCH("/roles/:id", authorization.RequirePermission(rbac, "roles.update"), rolesH.Update)
	private.DELETE("/roles/:id", authorization.RequirePermission(rbac, "roles.delete"), rolesH.Delete)

	permsH := permissionsmod.NewHandler(permissionsmod.NewService(permissionsmod.NewRepository(db)))
	private.GET("/permissions", authorization.RequirePermission(rbac, "permissions.read"), permsH.List)

	settingsH := settingsmod.NewHandler(settingsmod.NewService(settingsmod.NewRepository(db)))
	private.GET("/settings", authorization.RequirePermission(rbac, "settings.read"), settingsH.List)
	private.PATCH("/settings", authorization.RequirePermission(rbac, "settings.update"), settingsH.Update)

	companiesH := companiesmod.NewHandler(companiesmod.NewService(companiesmod.NewRepository(db)))
	private.GET("/companies/current", authorization.RequirePermission(rbac, "companies.read"), companiesH.Current)
	private.PATCH("/companies/current", authorization.RequirePermission(rbac, "companies.update"), companiesH.UpdateCurrent)

	branchesH := branchesmod.NewHandler(branchesmod.NewService(branchesmod.NewRepository(db)))
	private.GET("/branches", authorization.RequirePermission(rbac, "branches.read"), branchesH.List)
	private.POST("/branches", authorization.RequirePermission(rbac, "branches.create"), branchesH.Create)
	private.GET("/branches/:id", authorization.RequirePermission(rbac, "branches.read"), branchesH.Get)
	private.PATCH("/branches/:id", authorization.RequirePermission(rbac, "branches.update"), branchesH.Update)
	private.DELETE("/branches/:id", authorization.RequirePermission(rbac, "branches.delete"), branchesH.Delete)

	auditH := auditmod.NewHandler(auditmod.NewService(auditmod.NewRepository(db)))
	private.GET("/audit-logs", authorization.RequirePermission(rbac, "audit.read"), auditH.List)
	return r
}
