// New builds and returns the main Gin HTTP engine.
//
// This function is the HTTP composition root of the application.
// Its responsibility is to:
//
//   - Initialize the web framework
//   - Register global middleware
//   - Define API versioning
//   - Wire authorization and handlers
//   - Register public and private routes
//
// Important design principle:
//
// This function should ONLY wire infrastructure.
// It should NOT contain:
//
//   - business logic
//   - database queries
//   - validation rules
//   - authorization logic
//
// Those belong in services, repositories, and middleware.
func New(db *gorm.DB, cfg config.Config) *gin.Engine {

	// ---------------------------------------------------------------------
	// Create the HTTP engine
	// ---------------------------------------------------------------------
	//
	// gin.New() creates a clean engine without default middleware.
	// This is recommended for production because we explicitly control
	// the middleware stack.
	r := gin.New()

	// ---------------------------------------------------------------------
	// Global middleware stack
	// ---------------------------------------------------------------------
	//
	// Middleware order is intentional and critical.
	//
	// 1) Recovery
	//    Prevents the server from crashing on panic.
	//
	// 2) Logger
	//    Records request metadata.
	//
	// 3) RequestID
	//    Generates a unique request identifier.
	//    Used for tracing, debugging, and audit logs.
	//
	// 4) SecureHeaders
	//    Adds security headers:
	//      - X-Frame-Options
	//      - X-Content-Type-Options
	//      - Content-Security-Policy
	//
	// 5) CORS
	//    Allows frontend communication.
	//
	// 6) RateLimit
	//    Protects API from abuse and brute-force attacks.
	r.Use(
		gin.Recovery(),
		gin.Logger(),
		middleware.RequestID(),
		middleware.SecureHeaders(),
		middleware.CORS(cfg.FrontendURL),
		middleware.RateLimit(),
	)

	// ---------------------------------------------------------------------
	// API versioning
	// ---------------------------------------------------------------------
	//
	// Always version APIs from day one.
	//
	// This prevents breaking clients when new versions are released.
	api := r.Group("/api/v1")

	// ---------------------------------------------------------------------
	// Health check endpoint
	// ---------------------------------------------------------------------
	//
	// Used by:
	//
	//   - Docker health checks
	//   - Kubernetes readiness probes
	//   - Load balancers
	//   - Monitoring systems
	//
	// This endpoint must:
	//
	//   - be fast
	//   - be public
	//   - never access the database unless necessary
	api.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status": "ok",
		})
	})

	// ---------------------------------------------------------------------
	// RBAC authorization service
	// ---------------------------------------------------------------------
	//
	// This service validates permissions such as:
	//
	//   users.read
	//   users.create
	//   roles.update
	//   companies.delete
	//
	// Future optimization:
	//
	//   - Permission caching
	//   - Redis permission store
	//   - Policy engine integration
	rbac := authorization.New(db)

	// ---------------------------------------------------------------------
	// Authentication handler wiring
	// ---------------------------------------------------------------------
	//
	// Repository → database access
	// Service    → business rules
	// Handler    → HTTP interface
	authH := authmod.NewHandler(
		authmod.NewService(
			authmod.NewRepository(db),
			cfg,
		),
		cfg,
	)

	// ---------------------------------------------------------------------
	// Public authentication routes
	// ---------------------------------------------------------------------
	//
	// These routes do NOT require authentication.
	//
	// Security expectations:
	//
	//   - rate limited
	//   - audit logged
	//   - brute-force protected
	registerPublicAuthRoutes(api, authH)

	// ---------------------------------------------------------------------
	// Private route group
	// ---------------------------------------------------------------------
	//
	// All routes inside this group require authentication.
	//
	// RequireAuth middleware should:
	//
	//   - validate JWT signature
	//   - validate expiration
	//   - extract user ID
	//   - extract tenant/company ID
	//   - inject claims into context
	private := api.Group("")

	private.Use(
		middleware.RequireAuth(cfg.JWTAccessSecret),
	)

	// ---------------------------------------------------------------------
	// Private authentication routes
	// ---------------------------------------------------------------------
	//
	// These operate on the current authenticated user.
	//
	// Examples:
	//
	//   logout
	//   get profile
	//   update profile
	registerPrivateAuthRoutes(private, authH)

	// ---------------------------------------------------------------------
	// Users module wiring
	// ---------------------------------------------------------------------
	//
	// This module manages system users.
	//
	// Typical permissions:
	//
	//   users.read
	//   users.create
	//   users.update
	//   users.delete
	usersH := usersmod.NewHandler(
		usersmod.NewService(
			usersmod.NewRepository(db),
		),
	)

	registerUserRoutes(private, rbac, usersH)

	// ---------------------------------------------------------------------
	// Roles module wiring
	// ---------------------------------------------------------------------
	//
	// Roles define permission bundles.
	//
	// Important rule:
	//
	// System roles should be immutable:
	//
	//   owner
	//   admin
	//   support
	rolesH := rolesmod.NewHandler(
		rolesmod.NewService(
			rolesmod.NewRepository(db),
		),
	)

	registerRoleRoutes(private, rbac, rolesH)

	// ---------------------------------------------------------------------
	// Companies / Tenant module wiring
	// ---------------------------------------------------------------------
	//
	// Represents the tenant entity in a multi-tenant SaaS.
	//
	// "current" routes are safer than ID-based routes.
	//
	// This prevents horizontal privilege escalation.
	companiesH := companiesmod.NewHandler(
		companiesmod.NewService(
			companiesmod.NewRepository(db),
		),
	)

	registerCompanyRoutes(private, rbac, companiesH)

	// ---------------------------------------------------------------------
	// Router ready
	// ---------------------------------------------------------------------
	//
	// At this point:
	//
	//   - middleware is configured
	//   - routes are registered
	//   - handlers are wired
	//   - authorization is enabled
	//
	// The server is ready to start.
	return r
}