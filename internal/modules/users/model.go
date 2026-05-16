package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID        uuid.UUID      `json:"company_id" gorm:"type:uuid;not null;index"`
	BranchID         *uuid.UUID     `json:"branch_id,omitempty" gorm:"type:uuid;index"`
	Email            string         `json:"email" gorm:"not null"`
	EmailVerifiedAt  *time.Time     `json:"email_verified_at,omitempty"`
	Name             string         `json:"name" gorm:"not null"`
	AvatarURL        string         `json:"avatar_url,omitempty"`
	PasswordHash     string         `json:"-" gorm:"not null"`
	Status           string         `json:"status" gorm:"not null;default:active"`
	AccountType      string         `json:"account_type" gorm:"not null;default:free_client"`
	PremiumExpiresAt *time.Time     `json:"premium_expires_at,omitempty"`
	FreeExpiresAt    *time.Time     `json:"free_expires_at,omitempty"`
	BlockedAt        *time.Time     `json:"blocked_at,omitempty"`
	FailedLoginCount int            `json:"-"`
	LockedUntil      *time.Time     `json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}
