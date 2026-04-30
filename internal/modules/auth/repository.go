package auth

import (
	"github.com/example/gin-multitenant-backend/internal/modules/users"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type RefreshToken struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID              uuid.UUID
	CompanyID           uuid.UUID
	TokenHash           string
	IPAddress           string
	UserAgent           string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	ReplacedByTokenHash *string
	CreatedAt           time.Time
}
type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }
func (r *Repository) FindUserByEmail(email string) (users.User, error) {
	var u users.User
	err := r.DB.Where("email = ? AND status = ?", email, "active").First(&u).Error
	return u, err
}
func (r *Repository) SaveRefreshToken(t *RefreshToken) error { return r.DB.Create(t).Error }
func (r *Repository) FindRefreshToken(hash string) (RefreshToken, error) {
	var t RefreshToken
	err := r.DB.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&t).Error
	return t, err
}
func (r *Repository) RevokeRefreshToken(hash string, replacedBy *string) error {
	now := time.Now()
	return r.DB.Model(&RefreshToken{}).Where("token_hash = ?", hash).Updates(map[string]any{"revoked_at": &now, "replaced_by_token_hash": replacedBy}).Error
}
