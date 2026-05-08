package router

import (
	"time"

	"github.com/abrahamVado/go-golem.mx/internal/middleware"
	authmod "github.com/abrahamVado/go-golem.mx/internal/modules/auth"
	companiesmod "github.com/abrahamVado/go-golem.mx/internal/modules/companies"
	dashboardmod "github.com/abrahamVado/go-golem.mx/internal/modules/dashboard"
	permissionsmod "github.com/abrahamVado/go-golem.mx/internal/modules/permissions"
	projectsmod "github.com/abrahamVado/go-golem.mx/internal/modules/projects"
	rbacmod "github.com/abrahamVado/go-golem.mx/internal/modules/rbac"
	rolesmod "github.com/abrahamVado/go-golem.mx/internal/modules/roles"
	usersmod "github.com/abrahamVado/go-golem.mx/internal/modules/users"
	whitelistmod "github.com/abrahamVado/go-golem.mx/internal/modules/whitelist"
	"github.com/abrahamVado/go-golem.mx/internal/platform/config"
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	r := gin.New()

	r.Use(
		middleware.Recovery(),
		middleware.Logging(),
		middleware.RequestID(),
		middleware.SecureHeaders(),
		middleware.CORS(cfg.FrontendURL),
		middleware.BodySizeLimit(10<<20),
		middleware.Timeout(15*time.Second),
		middleware.RateLimit(),
	)

	api := r.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	rbacService := rbacmod.New(db)

	authH := authmod.NewHandler(
		authmod.NewService(authmod.NewRepository(db), cfg),
		cfg,
	)

	registerPublicAuthRoutes(api, authH)

	whitelistH := whitelistmod.NewHandler(
		whitelistmod.NewService(whitelistmod.NewRepository(db)),
	)
	whitelistmod.RegisterRoutes(api, whitelistH)

	private := api.Group("")
	private.Use(middleware.RequireAuth(cfg.JWTAccessSecret))

	registerPrivateAuthRoutes(private, authH)

	usersH := usersmod.NewHandler(
		usersmod.NewService(usersmod.NewRepository(db)),
	)
	registerUserRoutes(private, rbacService, usersH)

	rolesH := rolesmod.NewHandler(
		rolesmod.NewService(rolesmod.NewRepository(db)),
	)
	registerRoleRoutes(private, rbacService, rolesH)

	permissionsH := permissionsmod.NewHandler(
		permissionsmod.NewService(permissionsmod.NewRepository(db)),
	)
	registerPermissionRoutes(private, rbacService, permissionsH)

	companiesH := companiesmod.NewHandler(
		companiesmod.NewService(companiesmod.NewRepository(db)),
	)
	registerCompanyRoutes(private, rbacService, companiesH)

	dashboardH := dashboardmod.NewHandler(
		dashboardmod.NewService(db),
	)
	dashboardmod.RegisterRoutes(private, rbacService, dashboardH)

	projectsH := projectsmod.NewHandler(
		projectsmod.NewService(projectsmod.NewRepository(db)),
	)
	projectsmod.RegisterRoutes(private, rbacService, projectsH)

	return r
}
