package authorization

import (
	"github.com/golem-mx/core-api/internal/response"
	"github.com/golem-mx/core-api/internal/tenancy"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct{ DB *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{DB: db} }

func (s *Service) UserHasPermission(userID, companyID any, permission string) bool {
	var count int64
	s.DB.Table("user_roles ur").Joins("JOIN role_permissions rp ON rp.role_id = ur.role_id").Joins("JOIN permissions p ON p.id = rp.permission_id").Joins("JOIN roles r ON r.id = ur.role_id").Where("ur.user_id = ? AND ur.company_id = ? AND p.name = ? AND ur.deleted_at IS NULL", userID, companyID, permission).Or("ur.user_id = ? AND ur.company_id = ? AND r.name = ? AND ur.deleted_at IS NULL", userID, companyID, "Owner").Count(&count)
	return count > 0
}

func (s *Service) UserHasRole(userID, companyID any, roles ...string) bool {
	var count int64
	s.DB.Table("user_roles ur").Joins("JOIN roles r ON r.id = ur.role_id").Where("ur.user_id = ? AND ur.company_id = ? AND r.name IN ? AND ur.deleted_at IS NULL", userID, companyID, roles).Count(&count)
	return count > 0
}

func RequirePermission(s *Service, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.UserHasPermission(tenancy.UserID(c), tenancy.CompanyID(c), permission) {
			response.Fail(c, 403, "FORBIDDEN", "Missing required permission")
			return
		}
		c.Next()
	}
}
func RequireAnyPermission(s *Service, perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, p := range perms {
			if s.UserHasPermission(tenancy.UserID(c), tenancy.CompanyID(c), p) {
				c.Next()
				return
			}
		}
		response.Fail(c, 403, "FORBIDDEN", "Missing required permission")
	}
}
func RequireRole(s *Service, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.UserHasRole(tenancy.UserID(c), tenancy.CompanyID(c), roles...) {
			response.Fail(c, 403, "FORBIDDEN", "Missing required role")
			return
		}
		c.Next()
	}
}
