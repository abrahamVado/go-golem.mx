package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())

				c.AbortWithStatusJSON(500, gin.H{
					"success": false,
					"error":   "internal server error",
				})
			}
		}()

		c.Next()
	}
}