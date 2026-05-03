package middleware

import (
	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/security"
	"github.com/golem-mx/core-api/internal/tenancy"
	"github.com/gin-gonic/gin"
	"strings"
)

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			response.Fail(c, 401, "UNAUTHENTICATED", "Authentication required")
			return
		}
		claims, err := security.ParseAccessToken(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			response.Fail(c, 401, "INVALID_TOKEN", "Invalid or expired token")
			return
		}
		tenancy.Set(c, claims.UserID, claims.CompanyID, claims.BranchID)
		c.Next()
	}
}
