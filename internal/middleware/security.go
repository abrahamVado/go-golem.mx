package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	memstore "github.com/ulule/limiter/v3/drivers/store/memory"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}
func CORS(frontend string) gin.HandlerFunc {
	return cors.New(cors.Config{AllowOrigins: []string{frontend}, AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"}, AllowCredentials: true, MaxAge: 12 * time.Hour})
}
func RateLimit() gin.HandlerFunc {
	rate := limiter.Rate{Period: time.Minute, Limit: 120}
	store := memstore.NewStore()
	l := limiter.New(store, rate)
	return func(c *gin.Context) {
		ctx, err := l.Get(c, c.ClientIP())
		if err == nil {
			c.Header("X-RateLimit-Limit", "120")
			if ctx.Reached {
				c.AbortWithStatusJSON(429, gin.H{"success": false, "error": gin.H{"code": "RATE_LIMITED", "message": "Too many requests"}})
				return
			}
		}
		c.Next()
	}
}
