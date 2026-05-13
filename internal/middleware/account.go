package middleware

import (
	"time"

	"github.com/abrahamVado/go-paladin.mx/internal/modules/users"
	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireActiveAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user users.User
		if err := db.
			Where("company_id = ? AND id = ? AND deleted_at IS NULL", tenancy.CompanyID(c), tenancy.UserID(c)).
			First(&user).Error; err != nil {
			response.Unauthorized(c, "Account not found")
			c.Abort()
			return
		}

		snapshot := users.ResolveAccountSnapshot(user, NowUTC())
		if snapshot.Type != user.AccountType || (snapshot.IsBlocked && user.BlockedAt == nil) || (!snapshot.IsBlocked && user.BlockedAt != nil) {
			_ = db.Model(&users.User{}).
				Where("company_id = ? AND id = ? AND deleted_at IS NULL", user.CompanyID, user.ID).
				Updates(map[string]any{
					"account_type": snapshot.Type,
					"blocked_at":   snapshot.BlockedAt,
				}).Error
		}

		if snapshot.IsBlocked {
			response.Fail(c, 402, "PLAN_EXPIRED", "Your premium and free access periods have ended. Renew the account to continue.")
			c.Abort()
			return
		}

		c.Set("account_snapshot", snapshot)
		c.Next()
	}
}

func RequirePremiumAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("account_snapshot")
		if exists {
			if snapshot, ok := value.(users.AccountSnapshot); ok {
				if snapshot.IsPremium {
					c.Next()
					return
				}
				response.Fail(c, 402, "PREMIUM_REQUIRED", "This action requires a premium, founder, or owner account.")
				c.Abort()
				return
			}
		}

		var user users.User
		if err := db.
			Where("company_id = ? AND id = ? AND deleted_at IS NULL", tenancy.CompanyID(c), tenancy.UserID(c)).
			First(&user).Error; err != nil {
			response.Unauthorized(c, "Account not found")
			c.Abort()
			return
		}

		snapshot := users.ResolveAccountSnapshot(user, NowUTC())
		if !snapshot.IsPremium {
			response.Fail(c, 402, "PREMIUM_REQUIRED", "This action requires a premium, founder, or owner account.")
			c.Abort()
			return
		}

		c.Set("account_snapshot", snapshot)
		c.Next()
	}
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
