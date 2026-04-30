package audit

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"time"
)

type AuditLog struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID      `json:"company_id" gorm:"type:uuid;not null;index"`
	UserID     *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID *uuid.UUID     `json:"resource_id,omitempty"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Metadata   datatypes.JSON `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}
