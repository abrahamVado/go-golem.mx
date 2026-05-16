package apikeys

import (
	"time"

	"gorm.io/datatypes"

	"github.com/google/uuid"
)

type APIClient struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID          uuid.UUID      `json:"company_id" gorm:"type:uuid;not null"`
	Name               string         `json:"name" gorm:"not null"`
	Description        string         `json:"description" gorm:"not null;default:''"`
	Status             string         `json:"status" gorm:"not null;default:active"`
	AllowedIPs         datatypes.JSON `json:"allowed_ips,omitempty"`
	RateLimitPerMinute *int           `json:"rate_limit_per_minute,omitempty"`
	CreatedByUserID    *uuid.UUID     `json:"created_by_user_id,omitempty" gorm:"type:uuid"`
	RevokedAt          *time.Time     `json:"revoked_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"-" gorm:"index"`
}

type APIKey struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID         uuid.UUID      `gorm:"type:uuid;not null"`
	ClientID          uuid.UUID      `gorm:"type:uuid;not null"`
	KeyID             string         `gorm:"not null;uniqueIndex"`
	SecretHash        string         `gorm:"not null"`
	Scopes            datatypes.JSON `gorm:"not null"`
	LastUsedAt        *time.Time
	LastUsedIP        *string
	LastUsedUserAgent *string
	ExpiresAt         *time.Time
	Status            string `gorm:"not null;default:active"`
	RevokedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type APIClientPublicKey struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID         uuid.UUID  `json:"company_id" gorm:"type:uuid;not null"`
	ClientID          uuid.UUID  `json:"client_id" gorm:"type:uuid;not null"`
	Algorithm         string     `json:"algorithm" gorm:"not null"`
	PublicKeyRaw      []byte     `json:"-" gorm:"not null"`
	FingerprintSHA256 string     `json:"fingerprint_sha256" gorm:"not null;uniqueIndex"`
	SourceFormat      string     `json:"source_format" gorm:"not null;default:openssh"`
	Status            string     `json:"status" gorm:"not null;default:pending"`
	ActivatedAt       *time.Time `json:"activated_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
