package permissions

import "gorm.io/gorm"

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) List() ([]Permission, error) {
	var items []Permission
	err := r.DB.Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *Repository) NameExists(name string) (bool, error) {
	var count int64
	err := r.DB.Model(&Permission{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *Repository) Create(name, description string) (Permission, error) {
	item := Permission{Name: name, Description: description}
	return item, r.DB.Create(&item).Error
}
