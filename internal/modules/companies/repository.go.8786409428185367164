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

func (r *Repository) UpdateCurrent(companyID uuid.UUID, name string) (Company, error) {
	if err := r.DB.Model(&Company{}).
		Where("id = ? AND deleted_at IS NULL", companyID).
		Updates(map[string]any{
			"name": name,
		}).Error; err != nil {
		return Company{}, err
	}

	return r.FindByID(companyID)
}
