package rbac

import "gorm.io/gorm"

// Service provides RBAC authorization checks.
type Service struct {
	DB *gorm.DB
}

// New creates a new RBAC service.
func New(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// UserHasPermission checks whether a user has a permission inside a company.
func (s *Service) UserHasPermission(userID, companyID any, permission string) bool {
	if userID == nil || companyID == nil || permission == "" {
		return false
	}

	var count int64

	err := s.DB.
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("LEFT JOIN permissions p ON p.id = rp.permission_id").
		Where(
			"ur.user_id = ? AND ur.company_id = ? AND ur.deleted_at IS NULL",
			userID,
			companyID,
		).
		Where(
			s.DB.Where("p.name = ?", permission).Or("r.name = ?", SystemRoleOwner),
		).
		Count(&count).
		Error

	if err != nil {
		return false
	}

	return count > 0
}

// UserHasRole checks whether a user has at least one role inside a company.
func (s *Service) UserHasRole(userID, companyID any, roles ...string) bool {
	if userID == nil || companyID == nil || len(roles) == 0 {
		return false
	}

	var count int64

	err := s.DB.
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where(
			"ur.user_id = ? AND ur.company_id = ? AND ur.deleted_at IS NULL",
			userID,
			companyID,
		).
		Where("r.name IN ?", roles).
		Count(&count).
		Error

	if err != nil {
		return false
	}

	return count > 0
}
