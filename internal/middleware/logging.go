package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging logs every HTTP request with method, path, status, duration, and client IP.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start)

		log.Printf(
			"%s %s %d %s %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
	}
}