package companies

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }
func (r *Repository) FindByID(companyID uuid.UUID) (Company, error) {
	var c Company
	err := r.DB.First(&c, "id = ?", companyID).Error
	return c, err
}
