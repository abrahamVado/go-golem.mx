package roles

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Role struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   *uuid.UUID     `json:"company_id,omitempty" gorm:"type:uuid;index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsSystem    bool           `json:"is_system"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
