package users

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ DB *gorm.DB }

type UserSummary struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Status    string    `json:"status"`
	RoleNames []string  `json:"role_names"`
}

type userRoleRow struct {
	UserID   uuid.UUID `gorm:"column:user_id"`
	RoleName string    `gorm:"column:role_name"`
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) TenantDB(companyID uuid.UUID) *gorm.DB {
	return r.DB.Where("company_id = ?", companyID)
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.DB.Model(&User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) EmailExistsForOtherUser(companyID, userID uuid.UUID, email string) (bool, error) {
	var count int64
	err := r.DB.Model(&User{}).
		Where("company_id = ? AND id <> ? AND email = ? AND deleted_at IS NULL", companyID, userID, email).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) UserExists(companyID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&User{}).
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountValidRoles(companyID uuid.UUID, roleIDs []uuid.UUID) (int, error) {
	var count int64
	err := r.DB.Table("roles").
		Where("id IN ? AND deleted_at IS NULL AND (company_id = ? OR company_id IS NULL)", roleIDs, companyID).
		Count(&count).Error
	return int(count), err
}

func (r *Repository) List(companyID uuid.UUID) ([]UserSummary, error) {
	var users []User
	if err := r.DB.
		Where("company_id = ? AND deleted_at IS NULL", companyID).
		Order("created_at ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}

	return r.enrichRoleNames(users)
}

func (r *Repository) Create(companyID uuid.UUID, email, name, passwordHash, status string, roleIDs []uuid.UUID) (UserSummary, error) {
	user := User{
		ID:           uuid.New(),
		CompanyID:    companyID,
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Status:       status,
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if len(roleIDs) == 0 {
			return nil
		}

		for _, roleID := range roleIDs {
			if err := tx.Exec(`
				INSERT INTO user_roles (user_id, company_id, role_id, deleted_at)
				VALUES (?, ?, ?, NULL)
				ON CONFLICT (user_id, company_id, role_id)
				DO UPDATE SET deleted_at = NULL
			`, user.ID, companyID, roleID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return UserSummary{}, err
	}

	summaries, err := r.enrichRoleNames([]User{user})
	if err != nil {
		return UserSummary{}, err
	}
	return summaries[0], nil
}

func (r *Repository) Update(companyID, userID uuid.UUID, email, name, status string, roleIDs []uuid.UUID) (UserSummary, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).
			Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
			Updates(map[string]any{
				"email":  email,
				"name":   name,
				"status": status,
			}).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DELETE FROM user_roles WHERE company_id = ? AND user_id = ?`, companyID, userID).Error; err != nil {
			return err
		}

		for _, roleID := range roleIDs {
			if err := tx.Exec(`
				INSERT INTO user_roles (user_id, company_id, role_id, deleted_at)
				VALUES (?, ?, ?, NULL)
				ON CONFLICT (user_id, company_id, role_id)
				DO UPDATE SET deleted_at = NULL
			`, userID, companyID, roleID).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return UserSummary{}, err
	}

	var user User
	if err := r.DB.Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).First(&user).Error; err != nil {
		return UserSummary{}, err
	}

	summaries, err := r.enrichRoleNames([]User{user})
	if err != nil {
		return UserSummary{}, err
	}
	return summaries[0], nil
}

func (r *Repository) Delete(companyID, userID uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE user_roles SET deleted_at = NOW() WHERE company_id = ? AND user_id = ? AND deleted_at IS NULL`, companyID, userID).Error; err != nil {
			return err
		}
		return tx.Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).Delete(&User{}).Error
	})
}

func (r *Repository) enrichRoleNames(users []User) ([]UserSummary, error) {
	out := make([]UserSummary, 0, len(users))
	if len(users) == 0 {
		return out, nil
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var roleRows []userRoleRow
	if err := r.DB.Table("user_roles ur").
		Select("ur.user_id, r.name AS role_name").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id IN ? AND ur.deleted_at IS NULL", userIDs).
		Order("r.name ASC").
		Scan(&roleRows).Error; err != nil {
		return nil, err
	}

	roleNamesByUser := map[uuid.UUID][]string{}
	for _, row := range roleRows {
		roleNamesByUser[row.UserID] = append(roleNamesByUser[row.UserID], row.RoleName)
	}

	for _, user := range users {
		out = append(out, UserSummary{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
			Status:    user.Status,
			RoleNames: uniqueStrings(roleNamesByUser[user.ID]),
		})
	}
	return out, nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
