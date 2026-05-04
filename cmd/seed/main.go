package main

import (
	"log"
	"os"
	"time"

	"github.com/golem-mx/core-api/internal/config"
	"github.com/golem-mx/core-api/internal/database"
	"github.com/golem-mx/core-api/seeders"
)

//
// =====================================================
// GOLEM DATABASE SEED CLI
// =====================================================
//
// This command seeds the database with required baseline data.
//
// Typical seed data:
//
// - default system permissions
// - default RBAC roles
// - initial platform settings
// - optional demo/test data
//
// Usage:
//
//   go run ./cmd/seed
//
// Docker usage:
//
//   docker exec -it golem-api ./seed
//
// Production rule:
//
// Seeders must be IDEMPOTENT.
//
// That means running this command multiple times should NOT create duplicates.
// Use upsert behavior wherever possible.
//
// =====================================================
//

func main() {
	start := time.Now()

	log.Println("starting database seed process")

	// -------------------------------------------------
	// LOAD CONFIGURATION
	// -------------------------------------------------
	// Reads environment variables and application config.
	// DATABASE_URL must point to the internal Docker host:
	//
	//   postgres://user:pass@postgres:5432/db?sslmode=disable
	//
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	// -------------------------------------------------
	// CONNECT TO DATABASE
	// -------------------------------------------------
	// Creates the database connection used by all seeders.
	//
	// Keep this command small:
	// - connect
	// - run seeders
	// - close connection
	//
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		os.Exit(1)
	}

	// -------------------------------------------------
	// CLOSE DATABASE CONNECTION
	// -------------------------------------------------
	// If database.Connect returns a GORM DB, extract the underlying
	// sql.DB so the CLI can close connections cleanly before exiting.
	//
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("failed to get sql DB handle: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	// -------------------------------------------------
	// RUN SEEDERS
	// -------------------------------------------------
	// seeders.Run should orchestrate all required seed modules.
	//
	// Recommended order:
	// 1. Permissions
	// 2. System roles
	// 3. Role permissions
	// 4. Default settings
	// 5. Optional demo data
	//
	if err := seeders.Run(db, cfg); err != nil {
		log.Printf("database seed failed: %v", err)
		os.Exit(1)
	}

	log.Printf("database seed completed successfully in %s", time.Since(start))
}