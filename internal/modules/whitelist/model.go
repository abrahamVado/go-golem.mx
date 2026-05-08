package whitelist

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Request struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	Email     string         `json:"email" gorm:"not null;index"`
	Company   string         `json:"company"`
	Subject   string         `json:"subject"`
	Message   string         `json:"message" gorm:"not null"`
	Source    string         `json:"source" gorm:"not null;default:landing_page"`
	Status    string         `json:"status" gorm:"not null;default:pending"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
