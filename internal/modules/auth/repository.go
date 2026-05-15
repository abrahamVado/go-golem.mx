package auth

import (
	"strings"
	"time"

	companiesmod "github.com/abrahamVado/go-paladin.mx/internal/modules/companies"
	rolesmod "github.com/abrahamVado/go-paladin.mx/internal/modules/roles"
	"github.com/abrahamVado/go-paladin.mx/internal/modules/users"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// -----------------------------------------------------------------------------
// AUTH REPOSITORY
// -----------------------------------------------------------------------------
//
// This file owns persistence for authentication-related data.
//
// Responsibilities:
//
//   - lookup active users
//   - store refresh token hashes
//   - find valid refresh tokens
//   - revoke refresh tokens
//
// Security rule:
//
// Raw refresh tokens must NEVER be stored.
// Only token hashes should be persisted.
//
// -----------------------------------------------------------------------------

// RefreshToken stores a hashed opaque refresh token.
//
// TokenHash:
//
//	SHA-256 hash of the raw refresh token.
//
// RevokedAt:
//
//	nil means token is still active.
//	non-nil means token was revoked.
//
// ReplacedByTokenHash:
//
//	points to the next token hash during refresh-token rotation.
type RefreshToken struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID    uuid.UUID
	CompanyID uuid.UUID

	TokenHash string `gorm:"not null;uniqueIndex"`

	IPAddress string
	UserAgent string

	ExpiresAt time.Time
	RevokedAt *time.Time

	ReplacedByTokenHash *string

	CreatedAt time.Time
}

type PasswordResetToken struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID    uuid.UUID `gorm:"not null;index"`
	CompanyID uuid.UUID `gorm:"not null;index"`

	TokenHash string `gorm:"not null;uniqueIndex"`

	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"index"`

	CreatedAt time.Time
}

// Repository provides database access for the auth module.
type Repository struct {
	DB *gorm.DB
}

// NewRepository creates an auth repository instance.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

// FindUserByEmail returns an active user by normalized email.
//
// Only active users are allowed to authenticate.
//
// Recommended DB index:
//
//	CREATE INDEX idx_users_email_status ON users(email, status);
func (r *Repository) FindUserByEmail(email string) (users.User, error) {
	var user users.User

	err := r.DB.
		Where(
			"email = ? AND status = ? AND deleted_at IS NULL",
			email,
			"active",
		).
		First(&user).
		Error

	return user, err
}

func (r *Repository) FindUserByEmailAndCompanySlug(email string, companySlug string) (users.User, error) {
	var user users.User

	query := r.DB.Model(&users.User{}).
		Joins("JOIN companies c ON c.id = users.company_id").
		Where("users.email = ? AND users.status = ? AND users.deleted_at IS NULL AND c.deleted_at IS NULL", email, "active")

	if companySlug != "" {
		query = query.Where("c.slug = ?", companySlug)
	}

	err := query.First(&user).Error
	return user, err
}

func (r *Repository) SavePasswordResetToken(token *PasswordResetToken) error {
	return r.DB.Create(token).Error
}

func (r *Repository) FindActivePasswordResetToken(hash string) (PasswordResetToken, error) {
	var token PasswordResetToken
	err := r.DB.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now().UTC()).
		First(&token).
		Error
	return token, err
}

func (r *Repository) MarkPasswordResetTokenUsed(id uuid.UUID) error {
	now := time.Now().UTC()
	return r.DB.Model(&PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", &now).
		Error
}

func (r *Repository) UpdatePasswordConsumeResetTokenAndRevokeSessions(
	resetTokenID uuid.UUID,
	companyID uuid.UUID,
	userID uuid.UUID,
	passwordHash string,
) error {
	now := time.Now().UTC()

	return r.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&PasswordResetToken{}).
			Where("id = ? AND used_at IS NULL", resetTokenID).
			Update("used_at", &now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(&users.User{}).
			Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
			Update("password_hash", passwordHash).
			Error; err != nil {
			return err
		}

		return tx.Model(&RefreshToken{}).
			Where("company_id = ? AND user_id = ? AND revoked_at IS NULL", companyID, userID).
			Updates(map[string]any{
				"revoked_at": &now,
			}).
			Error
	})
}

func (r *Repository) UpdatePasswordAndRevokeSessions(companyID, userID uuid.UUID, passwordHash string) error {
	now := time.Now().UTC()
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&users.User{}).
			Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
			Update("password_hash", passwordHash).
			Error; err != nil {
			return err
		}

		return tx.Model(&RefreshToken{}).
			Where("company_id = ? AND user_id = ? AND revoked_at IS NULL", companyID, userID).
			Updates(map[string]any{
				"revoked_at": &now,
			}).
			Error
	})
}

// SaveRefreshToken stores a hashed refresh token.
//
// Important:
//
//   - TokenHash must be a hash, not the raw token.
//   - ExpiresAt must be set by the service.
//   - IP/UserAgent help with audit and suspicious-session detection.
func (r *Repository) SaveRefreshToken(token *RefreshToken) error {
	return r.DB.Create(token).Error
}

// FindRefreshToken returns a valid active refresh token.
//
// A token is valid only when:
//
//   - token_hash matches
//   - revoked_at IS NULL
//   - expires_at is in the future
//
// This protects against:
//
//   - expired token reuse
//   - revoked token reuse
//   - replay of rotated tokens
func (r *Repository) FindRefreshToken(hash string) (RefreshToken, error) {
	var token RefreshToken

	err := r.DB.
		Where(
			"token_hash = ? AND revoked_at IS NULL AND expires_at > ?",
			hash,
			time.Now().UTC(),
		).
		First(&token).
		Error

	return token, err
}

// RevokeRefreshToken revokes a refresh token.
//
// Used by:
//
//   - logout
//   - refresh-token rotation
//   - suspicious activity response
//
// replacedBy:
//
//	nil during logout
//	new token hash during rotation
func (r *Repository) RevokeRefreshToken(
	hash string,
	replacedBy *string,
) error {
	now := time.Now().UTC()

	return r.DB.
		Model(&RefreshToken{}).
		Where(
			"token_hash = ? AND revoked_at IS NULL",
			hash,
		).
		Updates(map[string]any{
			"revoked_at":             &now,
			"replaced_by_token_hash": replacedBy,
		}).
		Error
}

func (r *Repository) RevokeAllRefreshTokensForUser(companyID, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.DB.Model(&RefreshToken{}).
		Where("company_id = ? AND user_id = ? AND revoked_at IS NULL", companyID, userID).
		Updates(map[string]any{
			"revoked_at": &now,
		}).
		Error
}

func (r *Repository) FindUserByID(companyID, userID uuid.UUID) (users.User, error) {
	var user users.User

	err := r.DB.
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		First(&user).
		Error

	return user, err
}

func (r *Repository) FindPrimaryRoleName(companyID, userID uuid.UUID) (string, error) {
	type roleRow struct {
		Name string `gorm:"column:name"`
	}

	var row roleRow
	err := r.DB.
		Table("user_roles ur").
		Select("r.name").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.company_id = ? AND ur.user_id = ? AND ur.deleted_at IS NULL", companyID, userID).
		Order("CASE WHEN LOWER(r.name) = 'owner' THEN 0 WHEN LOWER(r.name) = 'admin' THEN 1 WHEN LOWER(r.name) = 'client' THEN 2 ELSE 3 END, r.name ASC").
		Limit(1).
		Scan(&row).
		Error
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.Name), nil
}

func (r *Repository) FindPermissionNames(companyID, userID uuid.UUID) ([]string, error) {
	type permissionRow struct {
		Name string `gorm:"column:name"`
	}

	var rows []permissionRow
	err := r.DB.
		Table("user_roles ur").
		Select("DISTINCT p.name").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("ur.company_id = ? AND ur.user_id = ? AND ur.deleted_at IS NULL", companyID, userID).
		Order("p.name ASC").
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	return out, nil
}

func (r *Repository) UpdateAccountState(companyID, userID uuid.UUID, accountType string, blockedAt *time.Time) error {
	updates := map[string]any{
		"account_type": accountType,
		"blocked_at":   blockedAt,
	}
	if accountType != users.AccountTypeInvalidClient {
		updates["blocked_at"] = nil
	}

	return r.DB.Model(&users.User{}).
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		Updates(updates).
		Error
}

func (r *Repository) UpdateMyProfile(companyID, userID uuid.UUID, name string) (users.User, error) {
	if err := r.DB.Model(&users.User{}).
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		Update("name", name).
		Error; err != nil {
		return users.User{}, err
	}

	return r.FindUserByID(companyID, userID)
}

func (r *Repository) UpdateMyAvatar(companyID, userID uuid.UUID, avatarURL string) (users.User, error) {
	if err := r.DB.Model(&users.User{}).
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		Update("avatar_url", avatarURL).
		Error; err != nil {
		return users.User{}, err
	}

	return r.FindUserByID(companyID, userID)
}

func (r *Repository) UpdateMyPassword(companyID, userID uuid.UUID, passwordHash string) error {
	return r.DB.Model(&users.User{}).
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, userID).
		Update("password_hash", passwordHash).
		Error
}

func (r *Repository) FindCompanySlugByID(companyID uuid.UUID) (string, error) {
	var company companiesmod.Company
	err := r.DB.Where("id = ? AND deleted_at IS NULL", companyID).First(&company).Error
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(company.Slug), nil
}

func (r *Repository) CompanySlugExists(slug string) (bool, error) {
	var count int64
	err := r.DB.Model(&companiesmod.Company{}).
		Where("slug = ? AND deleted_at IS NULL", slug).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.DB.Model(&users.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateCompanyWithClientUser(company companiesmod.Company, user users.User, role rolesmod.Role, permissionNames []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&company).Error; err != nil {
			return err
		}

		role.CompanyID = &company.ID
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

		if len(permissionNames) > 0 {
			if err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT ?, p.id
				FROM permissions p
				WHERE p.name IN ?
				ON CONFLICT (role_id, permission_id) DO NOTHING
			`, role.ID, permissionNames).Error; err != nil {
				return err
			}
		}

		user.CompanyID = company.ID
		now := time.Now().UTC()
		premiumUntil := now.AddDate(0, 0, 30)
		freeUntil := premiumUntil.AddDate(0, 0, 30)
		user.AccountType = users.AccountTypePremiumClient
		user.PremiumExpiresAt = &premiumUntil
		user.FreeExpiresAt = &freeUntil
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO user_roles (user_id, company_id, role_id, deleted_at)
			VALUES (?, ?, ?, NULL)
			ON CONFLICT (user_id, company_id, role_id)
			DO UPDATE SET deleted_at = NULL
		`, user.ID, company.ID, role.ID).Error; err != nil {
			return err
		}

		return tx.Exec(`
			WITH created_team AS (
				INSERT INTO teams (id, company_id, name, slug, created_by_user_id, created_at, updated_at, deleted_at)
				VALUES (gen_random_uuid(), ?, ?, ?, ?, NOW(), NOW(), NULL)
				RETURNING id
			)
			INSERT INTO team_members (team_id, user_id, added_by_user_id, created_at, updated_at, deleted_at)
			SELECT created_team.id, ?, ?, NOW(), NOW(), NULL
			FROM created_team
			ON CONFLICT (team_id, user_id)
			DO UPDATE SET deleted_at = NULL, updated_at = NOW()
		`, company.ID, company.Name, company.Slug, user.ID, user.ID, user.ID).Error
	})
}
