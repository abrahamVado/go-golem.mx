package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("migration source close error: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("migration database close error: %v", dbErr)
		}
	}()

	switch command {
	case "up":
		err := m.Up()

		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		if err != nil {
			log.Fatalf("migration up failed: %v", err)
		}

		fmt.Println("migrations applied")

	case "down":
		err := m.Steps(-1)

		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migration to rollback")
			return
		}

		if err != nil {
			log.Fatalf("migration down failed: %v", err)
		}

		fmt.Println("rolled back 1 migration")

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("force requires version number")
		}

		var version int
		if _, err := fmt.Sscanf(os.Args[2], "%d", &version); err != nil {
			log.Fatalf("invalid version: %s", os.Args[2])
		}

		if err := m.Force(version); err != nil {
			log.Fatalf("force failed: %v", err)
		}

		fmt.Printf("forced version to %d\n", version)

	case "version":
		version, dirty, err := m.Version()

		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("no migrations applied")
			return
		}

		if err != nil {
			log.Fatalf("failed to get version: %v", err)
		}

		fmt.Printf("version: %d, dirty: %v\n", version, dirty)

	default:
		fmt.Printf("unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`
Golem Migration CLI

Usage:
  migrate <command>

Commands:
  up           Apply all pending migrations
  down         Rollback last migration
  force <ver>  Force migration version
  version      Show current migration version

Examples:
  migrate up
  migrate down
  migrate version
  migrate force 1
`)
}
