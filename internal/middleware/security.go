package middleware

import (
	"strconv"
	"time"
	"context"
	"net/http"	

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	memstore "github.com/ulule/limiter/v3/drivers/store/memory"
)

// RequestID adds a unique request ID to every request.
//
// Purpose:
//
//   - trace logs
//   - debug errors
//   - connect audit events	
//   - expose correlation ID to frontend
//
// Header returned:
//
//   X-Request-ID: <uuid>
//
// The same ID is also stored in gin.Context as:
//
//   request_id
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()

		c.Header("X-Request-ID", id)
		c.Set("request_id", id)

		c.Next()
	}
}

// SecureHeaders adds baseline HTTP security headers.
//
// These headers reduce common browser-based attack surfaces:
//
//   - MIME sniffing
//   - clickjacking
//   - excessive referrer leakage
//   - unauthorized browser feature access
//
// This middleware should run globally.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent browsers from guessing content types.
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent the API from being embedded in iframes.
		c.Header("X-Frame-Options", "DENY")

		// Reduce sensitive URL leakage through referrer headers.
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable browser features that the API does not need.
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Avoid exposing implementation details.
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		c.Next()
	}
}

// CORS configures browser access to the API.
//
// The frontend origin must come from trusted configuration.
//
// Example:
//
//   https://app.golem.mx
//
// Security notes:
//
//   - Do not use "*" with AllowCredentials=true.
//   - Keep allowed methods minimal.
//   - Keep allowed headers explicit.
//   - Do not expose internal/debug headers unnecessarily.
func CORS(frontend string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			frontend,
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},

		ExposeHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	})
}

// RateLimit protects the API from abusive request volume.
//
// Current policy:
//
//   120 requests per minute per client IP
//
// This is useful for:
//
//   - brute-force protection
//   - accidental frontend loops
//   - basic abuse prevention
//
// Current storage:
//
//   in-memory
//
// Important production note:
//
// In-memory rate limiting only works correctly for a single API instance.
// For multiple containers, replicas, or servers, use Redis instead.
func RateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  120,
	}

	store := memstore.NewStore()
	l := limiter.New(store, rate)

	return func(c *gin.Context) {
		limitContext, err := l.Get(c, c.ClientIP())

		// If the limiter fails, we intentionally fail open.
		//
		// Reason:
		//
		//   The API should not become unavailable because the limiter backend
		//   had a transient issue.
		//
		// For high-security systems, you may choose to fail closed instead.
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(limitContext.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limitContext.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(limitContext.Reset, 10))

		if limitContext.Reached {
			c.AbortWithStatusJSON(429, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests",
				},
			})
			return
		}

		c.Next()
	}
}
// BodySizeLimit limits the maximum request body size.
//
// Example:
//
//	BodySizeLimit(10 << 20) // 10 MB
func BodySizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// Timeout adds a request-level context timeout.
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}