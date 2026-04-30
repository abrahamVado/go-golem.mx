package settings

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"time"
)

type Setting struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID uuid.UUID      `json:"company_id" gorm:"type:uuid;not null;index"`
	Key       string         `json:"key"`
	Value     datatypes.JSON `json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
