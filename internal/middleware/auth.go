package middleware

import (
	"github.com/example/gin-multitenant-backend/internal/response"
	"github.com/example/gin-multitenant-backend/internal/security"
	"github.com/example/gin-multitenant-backend/internal/tenancy"
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
