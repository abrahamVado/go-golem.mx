package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//
// =====================================================
// GOLEM API ENTRYPOINT
// =====================================================
//
// Responsibilities:
//
// - Boot HTTP server
// - Configure routes
// - Provide health endpoint
// - Handle graceful shutdown
// - Expose basic observability hooks
//
// IMPORTANT:
//
// This file should NOT contain business logic.
// It should only wire dependencies and start the server.
//
// =====================================================
//

func main() {
	// -------------------------------------------------
	// CONFIGURATION
	// -------------------------------------------------

	port := getEnv("PORT", "8080")

	// -------------------------------------------------
	// ROUTER SETUP
	// -------------------------------------------------

	mux := http.NewServeMux()

	// Health endpoint (used by Docker / Nginx / monitoring)
	mux.HandleFunc("/health", healthHandler)

	// Root endpoint (basic service info)
	mux.HandleFunc("/", rootHandler)

	// Wrap with middleware (logging, recovery, etc.)
	handler := withMiddleware(mux)

	// -------------------------------------------------
	// HTTP SERVER CONFIG
	// -------------------------------------------------

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// -------------------------------------------------
	// START SERVER (ASYNC)
	// -------------------------------------------------

	go func() {
		log.Printf("🚀 go-golem.mx API running on :%s\n", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// -------------------------------------------------
	// GRACEFUL SHUTDOWN
	// -------------------------------------------------

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("🛑 shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Println("server stopped cleanly")
	}
}

//
// =====================================================
// HANDLERS
// =====================================================
//

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Used by:
	// - Docker healthcheck
	// - Nginx
	// - uptime monitoring
	respondJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "go-golem.mx",
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"service": "go-golem.mx",
		"version": "v1",
	})
}

//
// =====================================================
// MIDDLEWARE
// =====================================================
//

func withMiddleware(next http.Handler) http.Handler {
	return loggingMiddleware(recoveryMiddleware(next))
}

// Logs each request (basic access log)
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// Prevents server crash on panic
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)

				respondJSON(w, http.StatusInternalServerError, map[string]any{
					"success": false,
					"error":   "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

//
// =====================================================
// HELPERS
// =====================================================
//

// JSON response helper (consistent output format)
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Env helper with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}