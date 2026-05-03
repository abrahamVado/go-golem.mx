package database

import (
	"database/sql"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// -----------------------------------------------------------------------------
// DATABASE CONNECTION
// -----------------------------------------------------------------------------
//
// Connect establishes a PostgreSQL connection using GORM.
//
// Responsibilities:
//
//   - open database connection
//   - configure connection pool
//   - configure logging level
//   - validate connectivity
//
// Design goals:
//
//   - predictable startup behavior
//   - production-safe defaults
//   - connection stability
//   - controlled logging
//
// This function should be called exactly once during application startup.
//
// -----------------------------------------------------------------------------

func Connect(databaseURL string) (*gorm.DB, error) {

	// -------------------------------------------------------------------------
	// Validate input
	// -------------------------------------------------------------------------
	//
	// Fail fast if the database URL is missing.
	//
	if databaseURL == "" {
		return nil, ErrDatabaseURLMissing
	}

	// -------------------------------------------------------------------------
	// Configure GORM logging
	// -------------------------------------------------------------------------
	//
	// Recommended logging levels:
	//
	//   Silent  -> production high-load systems
	//   Warn    -> default production
	//   Info    -> development / debugging
	//
	gormLogger := logger.Default.LogMode(logger.Warn)

	// -------------------------------------------------------------------------
	// Open database connection
	// -------------------------------------------------------------------------
	//
	// This does not immediately verify connectivity.
	// It only initializes the connection configuration.
	//
	db, err := gorm.Open(
		postgres.Open(databaseURL),
		&gorm.Config{
			Logger: gormLogger,
		},
	)

	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// Extract underlying SQL connection
	// -------------------------------------------------------------------------
	//
	// Required for configuring connection pooling.
	//
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// Configure connection pool
	// -------------------------------------------------------------------------
	//
	// These values are production-safe defaults.
	//
	// Tune based on:
	//
	//   CPU cores
	//   DB capacity
	//   traffic volume
	//
	// -----------------------------------------------------------------------------

	sqlDB.SetMaxOpenConns(25)

	sqlDB.SetMaxIdleConns(25)

	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// -------------------------------------------------------------------------
	// Validate connectivity
	// -------------------------------------------------------------------------
	//
	// Ensures the database is reachable before the server starts.
	//
	// Without this check:
	//
	//   the API could start successfully
	//   but fail on the first request
	//
	if err := ping(sqlDB); err != nil {
		return nil, err
	}

	return db, nil
}

// -----------------------------------------------------------------------------
// CONNECTION HEALTH CHECK
// -----------------------------------------------------------------------------

func ping(db *sql.DB) error {

	// Retry logic could be added here if desired.

	return db.Ping()
}

// -----------------------------------------------------------------------------
// ERRORS
// -----------------------------------------------------------------------------

var ErrDatabaseURLMissing = &DatabaseError{
	Message: "database URL is required",
}

// Simple typed error for infrastructure failures.

type DatabaseError struct {
	Message string
}

func (e *DatabaseError) Error() string {
	return e.Message
}