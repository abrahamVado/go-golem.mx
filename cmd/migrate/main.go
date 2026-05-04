package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//
// =====================================================
// GOLEM DATABASE MIGRATION CLI
// =====================================================
//
// This CLI runs database migrations inside the container.
//
// Why this exists:
//
// - Avoid manual SQL execution
// - Keep schema versioned and reproducible
// - Allow safe deploys (up/down migrations)
//
// Usage (inside container):
//
//   ./migrate up
//   ./migrate down
//   ./migrate force 1
//   ./migrate version
//
// Docker usage:
//
//   docker exec -it golem-api ./migrate up
//
// IMPORTANT:
//
// - DATABASE_URL must be set
// - migrations folder must exist inside container
//
// Example:
//
//   DATABASE_URL=postgres://user:pass@postgres:5432/db?sslmode=disable
//
// =====================================================
//

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Path inside container (ensure Dockerfile copies migrations)
	migrationsPath := "file://migrations"

	m, err := migrate.New(
		migrationsPath,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}

	switch command {

	// -------------------------------------------------
	// APPLY ALL PENDING MIGRATIONS
	// -------------------------------------------------
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration up failed: %v", err)
		}
		fmt.Println("migrations applied")

	// -------------------------------------------------
	// ROLLBACK ONE STEP
	// -------------------------------------------------
	case "down":
		if err := m.Steps(-1); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}
		fmt.Println("rolled back 1 migration")

	// -------------------------------------------------
	// FORCE VERSION (DANGEROUS)
	// -------------------------------------------------
	// Use only if migration state is broken
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("force requires version number")
		}

		version := os.Args[2]

		var v int
		_, err := fmt.Sscanf(version, "%d", &v)
		if err != nil {
			log.Fatalf("invalid version: %s", version)
		}

		if err := m.Force(v); err != nil {
			log.Fatalf("force failed: %v", err)
		}
		fmt.Printf("forced version to %d\n", v)

	// -------------------------------------------------
	// SHOW CURRENT VERSION
	// -------------------------------------------------
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("no migrations applied")
				return
			}
			log.Fatalf("failed to get version: %v", err)
		}

		fmt.Printf("version: %d, dirty: %v\n", v, dirty)

	// -------------------------------------------------
	// UNKNOWN COMMAND
	// -------------------------------------------------
	default:
		fmt.Printf("unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

//
// =====================================================
// HELP
// =====================================================
//

func printHelp() {
	fmt.Println(`
Golem Migration CLI

Usage:
  migrate <command>

Commands:
  up           Apply all pending migrations
  down         Rollback last migration
  force <ver>  Force migration version (dangerous)
  version      Show current migration version

Examples:
  migrate up
  migrate down
  migrate version
  migrate force 1
`)
}