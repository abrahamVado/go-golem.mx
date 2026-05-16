package audit

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }
func (r *Repository) TenantDB(companyID uuid.UUID) *gorm.DB {
	return r.DB.Where("company_id = ?", companyID)
}

func (r *Repository) Create(log AuditLog) error {
	return r.DB.Create(&log).Error
}

func JSONMetadata(value any) []byte {
	if value == nil {
		return []byte(`{}`)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}
