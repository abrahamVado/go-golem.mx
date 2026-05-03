package auth

import (
	"time"

	"github.com/golem-mx/core-api/internal/modules/users"
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
//   SHA-256 hash of the raw refresh token.
//
// RevokedAt:
//
//   nil means token is still active.
//   non-nil means token was revoked.
//
// ReplacedByTokenHash:
//
//   points to the next token hash during refresh-token rotation.
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
//   CREATE INDEX idx_users_email_status ON users(email, status);
//
func (r *Repository) FindUserByEmail(email string) (users.User, error) {
	var user users.User

	err := r.DB.
		Where(
			"email = ? AND status = ?",
			email,
			"active",
		).
		First(&user).
		Error

	return user, err
}

// SaveRefreshToken stores a hashed refresh token.
//
// Important:
//
//   - TokenHash must be a hash, not the raw token.
//   - ExpiresAt must be set by the service.
//   - IP/UserAgent help with audit and suspicious-session detection.
//
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
//
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
//   nil during logout
//   new token hash during rotation
//
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