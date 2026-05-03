// -----------------------------------------------------------------------------
// CORE API MODULE DEFINITION
// -----------------------------------------------------------------------------
//
// Module identity for the main backend service.
//
// This service is responsible for:
//
//   - Authentication
//   - Authorization (RBAC)
//   - Multi-tenant company management
//   - User lifecycle management
//   - Audit logging
//   - Platform configuration
//
// IMPORTANT:
//
// The module path is part of the public identity of the service.
// Changing it later can break imports, CI/CD pipelines, and deployments.
//
// Naming convention:
//
//   github.com/<organization>/<service>
//
// -----------------------------------------------------------------------------

module github.com/golem-mx/core-api

// -----------------------------------------------------------------------------
// GO VERSION
// -----------------------------------------------------------------------------
//
// Defines the minimum Go version required to build the project.
//
// All developers, CI pipelines, and Docker images should use this version
// or newer to ensure consistent behavior.
//
go 1.22

// -----------------------------------------------------------------------------
// TOOLCHAIN PINNING
// -----------------------------------------------------------------------------
//
// Locks the Go compiler version used to build the project.
//
// Benefits:
//
//   - reproducible builds
//   - consistent CI behavior
//   - predictable dependency resolution
//
// Recommended practice for production systems.
//
toolchain go1.22.5

// -----------------------------------------------------------------------------
// DEPENDENCIES
// -----------------------------------------------------------------------------
//
// These libraries provide the runtime infrastructure for the API.
//
// Dependency selection rules:
//
//   - stable releases only
//   - actively maintained
//   - security-reviewed
//   - widely adopted
//
// Avoid:
//
//   - experimental libraries
//   - abandoned projects
//   - unnecessary dependencies
//
// -----------------------------------------------------------------------------

require (

	// -------------------------------------------------------------------------
	// HTTP FRAMEWORK
	// -------------------------------------------------------------------------
	//
	// Core web framework responsible for:
	//
	//   - routing
	//   - middleware
	//   - request handling
	//   - response serialization
	//
	// This is the primary entry point for all HTTP traffic.
	//
	github.com/gin-gonic/gin v1.10.1


	// -------------------------------------------------------------------------
	// CORS (Cross-Origin Resource Sharing)
	// -------------------------------------------------------------------------
	//
	// Controls which frontend domains are allowed to call the API.
	//
	// Critical for:
	//
	//   - browser security
	//   - frontend integration
	//   - preventing unauthorized cross-origin requests
	//
	github.com/gin-contrib/cors v1.7.2


	// -------------------------------------------------------------------------
	// JWT AUTHENTICATION
	// -------------------------------------------------------------------------
	//
	// Provides JSON Web Token support.
	//
	// Used for:
	//
	//   - access tokens
	//   - refresh tokens
	//   - session validation
	//
	// Security-critical dependency.
	//
	github.com/golang-jwt/jwt/v5 v5.2.1


	// -------------------------------------------------------------------------
	// UUID GENERATION
	// -------------------------------------------------------------------------
	//
	// Generates globally unique identifiers.
	//
	// Used for:
	//
	//   - users
	//   - companies
	//   - sessions
	//   - audit events
	//   - API resources
	//
	github.com/google/uuid v1.6.0


	// -------------------------------------------------------------------------
	// ENVIRONMENT CONFIGURATION
	// -------------------------------------------------------------------------
	//
	// Loads environment variables from .env files.
	//
	// Intended for:
	//
	//   - local development
	//   - testing
	//
	// In production:
	//
	// configuration should come from:
	//
	//   - environment variables
	//   - container runtime
	//   - secret manager
	//
	github.com/joho/godotenv v1.5.1


	// -------------------------------------------------------------------------
	// RATE LIMITING
	// -------------------------------------------------------------------------
	//
	// Protects the API against:
	//
	//   - brute-force attacks
	//   - denial-of-service attempts
	//   - abusive clients
	//
	github.com/ulule/limiter/v3 v3.11.2


	// -------------------------------------------------------------------------
	// RATE LIMIT STORAGE DRIVER
	// -------------------------------------------------------------------------
	//
	// In-memory store used for rate limiting.
	//
	// Suitable for:
	//
	//   - development
	//   - single-node deployments
	//
	// Recommended replacement for production clusters:
	//
	//   Redis
	//   Distributed cache
	//
	github.com/ulule/limiter/v3/drivers/store/memory v3.11.2


	// -------------------------------------------------------------------------
	// CRYPTOGRAPHY
	// -------------------------------------------------------------------------
	//
	// Provides secure cryptographic primitives.
	//
	// Used for:
	//
	//   - password hashing
	//   - secure token generation
	//   - encryption utilities
	//
	golang.org/x/crypto v0.26.0


	// -------------------------------------------------------------------------
	// POSTGRESQL DATABASE DRIVER
	// -------------------------------------------------------------------------
	//
	// Enables database connectivity for PostgreSQL.
	//
	// Responsibilities:
	//
	//   - SQL execution
	//   - connection pooling
	//   - transaction handling
	//
	gorm.io/driver/postgres v1.5.9


	// -------------------------------------------------------------------------
	// ORM (Object Relational Mapper)
	// -------------------------------------------------------------------------
	//
	// Provides database abstraction and query building.
	//
	// Used for:
	//
	//   - migrations
	//   - models
	//   - transactions
	//   - query execution
	//
	gorm.io/gorm v1.25.12


	// -------------------------------------------------------------------------
	// ADVANCED DATABASE TYPES
	// -------------------------------------------------------------------------
	//
	// Adds support for advanced PostgreSQL data types.
	//
	// Examples:
	//
	//   JSONB
	//   arrays
	//   structured metadata
	//
	gorm.io/datatypes v1.2.7

)


// -----------------------------------------------------------------------------
// MAINTENANCE COMMANDS
// -----------------------------------------------------------------------------
//
// Keep dependencies healthy by running periodically:
//
//   go mod tidy
//   go mod verify
//   go list -m -u all
//
// These commands:
//
//   - remove unused dependencies
//   - verify module integrity
//   - check for updates
//
// -----------------------------------------------------------------------------