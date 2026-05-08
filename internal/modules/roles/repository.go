package roles

import (
	permissionsmod "github.com/abrahamVado/go-paladin.mx/internal/modules/permissions"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

type RoleSummary struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsSystem        bool      `json:"is_system"`
	PermissionCount int64     `json:"permission_count"`
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }
func (r *Repository) TenantDB(companyID uuid.UUID) *gorm.DB {
	return r.DB.Where("company_id = ?", companyID)
}

func (r *Repository) NameExists(companyID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.DB.Model(&Role{}).
		Where("company_id = ? AND name = ? AND deleted_at IS NULL", companyID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountValidPermissions(permissionIDs []uuid.UUID) (int, error) {
	var count int64
	err := r.DB.Table("permissions").Where("id IN ?", permissionIDs).Count(&count).Error
	return int(count), err
}

func (r *Repository) List(companyID uuid.UUID) ([]RoleSummary, error) {
	var items []RoleSummary
	err := r.DB.Table("roles").
		Select("roles.id, roles.name, roles.description, roles.is_system, COUNT(rp.permission_id) AS permission_count").
		Joins("LEFT JOIN role_permissions rp ON rp.role_id = roles.id").
		Where("(roles.company_id = ? OR roles.company_id IS NULL) AND roles.deleted_at IS NULL", companyID).
		Group("roles.id").
		Order("roles.is_system DESC, roles.created_at ASC").
		Scan(&items).Error
	return items, err
}

func (r *Repository) Create(companyID uuid.UUID, name, description string, permissionIDs []uuid.UUID) (RoleSummary, error) {
	role := Role{
		ID:          uuid.New(),
		CompanyID:   &companyID,
		Name:        name,
		Description: description,
		IsSystem:    false,
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				VALUES (?, ?)
				ON CONFLICT (role_id, permission_id) DO NOTHING
			`, role.ID, permissionID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return RoleSummary{}, err
	}

	items, err := r.List(companyID)
	if err != nil {
		return RoleSummary{}, err
	}
	for _, item := range items {
		if item.ID == role.ID {
			return item, nil
		}
	}

	return RoleSummary{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
	}, nil
}

func (r *Repository) RoleExists(companyID, roleID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&Role{}).
		Where("id = ? AND company_id = ? AND deleted_at IS NULL", roleID, companyID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ReplacePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID).Error; err != nil {
			return err
		}

		for _, permissionID := range permissionIDs {
			if err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				VALUES (?, ?)
				ON CONFLICT (role_id, permission_id) DO NOTHING
			`, roleID, permissionID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) Update(companyID, roleID uuid.UUID, name, description *string, permissionIDs []uuid.UUID, replacePermissions bool) (RoleSummary, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if name != nil {
			updates["name"] = *name
		}
		if description != nil {
			updates["description"] = *description
		}

		if len(updates) > 0 {
			if err := tx.Model(&Role{}).
				Where("id = ? AND company_id = ? AND deleted_at IS NULL", roleID, companyID).
				Updates(updates).Error; err != nil {
				return err
			}
		}

		if replacePermissions {
			if err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID).Error; err != nil {
				return err
			}
			for _, permissionID := range permissionIDs {
				if err := tx.Exec(`
					INSERT INTO role_permissions (role_id, permission_id)
					VALUES (?, ?)
					ON CONFLICT (role_id, permission_id) DO NOTHING
				`, roleID, permissionID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return RoleSummary{}, err
	}

	items, err := r.List(companyID)
	if err != nil {
		return RoleSummary{}, err
	}
	for _, item := range items {
		if item.ID == roleID {
			return item, nil
		}
	}
	return RoleSummary{}, gorm.ErrRecordNotFound
}

func (r *Repository) ListPermissions(companyID, roleID uuid.UUID) ([]permissionsmod.Permission, error) {
	var items []permissionsmod.Permission
	err := r.DB.Table("permissions p").
		Select("p.id, p.name, p.description, p.created_at, p.updated_at").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Where("rp.role_id = ? AND r.company_id = ? AND r.deleted_at IS NULL", roleID, companyID).
		Order("p.name ASC").
		Scan(&items).Error
	return items, err
}

func (r *Repository) Delete(companyID, roleID uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DELETE FROM user_roles WHERE company_id = ? AND role_id = ?`, companyID, roleID).Error; err != nil {
			return err
		}

		return tx.Model(&Role{}).
			Where("id = ? AND company_id = ? AND deleted_at IS NULL", roleID, companyID).
			Update("deleted_at", gorm.Expr("NOW()")).
			Error
	})
}
