package whitelist

import "gorm.io/gorm"

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) Create(item *Request) error {
	return r.DB.Create(item).Error
}
