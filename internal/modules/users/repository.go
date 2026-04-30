package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }
func (r *Repository) TenantDB(companyID uuid.UUID) *gorm.DB {
	return r.DB.Where("company_id = ?", companyID)
}
