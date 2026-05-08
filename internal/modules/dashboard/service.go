package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

type Summary struct {
	CompanyID          uuid.UUID         `json:"company_id"`
	GeneratedAt        string            `json:"generated_at"`
	ProjectCount       int64             `json:"project_count"`
	TaskCount          int64             `json:"task_count"`
	TeamMemberCount    int64             `json:"team_member_count"`
	CompletedTaskCount int64             `json:"completed_task_count"`
	CompletionRate     float64           `json:"completion_rate"`
	TaskStatus         []StatusMetric    `json:"task_status"`
	ProjectActivity    []ProjectActivity `json:"project_activity"`
	RecentProjects     []RecentProject   `json:"recent_projects"`
}

type SystemLogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	ActorName string `json:"actor_name,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
}

type SystemLogs struct {
	CompanyID   uuid.UUID        `json:"company_id"`
	GeneratedAt string           `json:"generated_at"`
	Items       []SystemLogEntry `json:"items"`
}

type StatusMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type ProjectActivity struct {
	Name  string `json:"name"`
	Tasks int64  `json:"tasks"`
}

type RecentProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	TaskCount   int64  `json:"task_count"`
}

type statusRow struct {
	StatusKey string `gorm:"column:status_key"`
	Count     int64  `gorm:"column:count"`
}

type projectActivityRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Icon        string    `gorm:"column:icon"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	Tasks       int64     `gorm:"column:tasks"`
}

type auditLogRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	Action    string    `gorm:"column:action"`
	Resource  string    `gorm:"column:target_type"`
	ActorName string    `gorm:"column:actor_name"`
	IPAddress string    `gorm:"column:ip_address"`
	UserAgent string    `gorm:"column:user_agent"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (s *Service) Summary(companyID uuid.UUID) (Summary, error) {
	summary := Summary{
		CompanyID:   companyID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.db.Table("projects").
		Where("company_id = ? AND deleted_at IS NULL", companyID).
		Count(&summary.ProjectCount).Error; err != nil {
		return Summary{}, err
	}

	if err := s.db.Table("tasks").
		Where("company_id = ? AND deleted_at IS NULL", companyID).
		Count(&summary.TaskCount).Error; err != nil {
		return Summary{}, err
	}

	if err := s.db.Table("users").
		Where("company_id = ? AND deleted_at IS NULL AND status = ?", companyID, "active").
		Count(&summary.TeamMemberCount).Error; err != nil {
		return Summary{}, err
	}

	if err := s.db.Table("tasks").
		Where("company_id = ? AND deleted_at IS NULL AND status_key = ?", companyID, "done").
		Count(&summary.CompletedTaskCount).Error; err != nil {
		return Summary{}, err
	}

	if summary.TaskCount > 0 {
		summary.CompletionRate = float64(summary.CompletedTaskCount) / float64(summary.TaskCount)
	}

	var statusRows []statusRow
	if err := s.db.Table("tasks").
		Select("status_key, COUNT(*) AS count").
		Where("company_id = ? AND deleted_at IS NULL", companyID).
		Group("status_key").
		Order("count DESC").
		Scan(&statusRows).Error; err != nil {
		return Summary{}, err
	}
	summary.TaskStatus = make([]StatusMetric, 0, len(statusRows))
	for _, row := range statusRows {
		summary.TaskStatus = append(summary.TaskStatus, StatusMetric{
			Key:   row.StatusKey,
			Label: taskStatusLabel(row.StatusKey),
			Count: row.Count,
		})
	}

	var activityRows []projectActivityRow
	if err := s.db.Table("projects p").
		Select("p.id, p.name, p.description, p.icon, p.created_at, COUNT(t.id) AS tasks").
		Joins("LEFT JOIN tasks t ON t.project_id = p.id AND t.deleted_at IS NULL").
		Where("p.company_id = ? AND p.deleted_at IS NULL", companyID).
		Group("p.id").
		Order("tasks DESC, p.created_at DESC").
		Limit(6).
		Scan(&activityRows).Error; err != nil {
		return Summary{}, err
	}

	summary.ProjectActivity = make([]ProjectActivity, 0, len(activityRows))
	summary.RecentProjects = make([]RecentProject, 0, len(activityRows))
	for _, row := range activityRows {
		summary.ProjectActivity = append(summary.ProjectActivity, ProjectActivity{
			Name:  row.Name,
			Tasks: row.Tasks,
		})
		summary.RecentProjects = append(summary.RecentProjects, RecentProject{
			ID:          row.ID.String(),
			Name:        row.Name,
			Description: row.Description,
			Icon:        row.Icon,
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
			TaskCount:   row.Tasks,
		})
	}

	return summary, nil
}

func (s *Service) SystemLogs(companyID uuid.UUID, limit int) (SystemLogs, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 200 {
		limit = 200
	}

	payload := SystemLogs{
		CompanyID:   companyID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Items:       []SystemLogEntry{},
	}

	var rows []auditLogRow
	if err := s.db.Table("audit_logs").
		Select("audit_logs.id, audit_logs.action, audit_logs.target_type, COALESCE(users.name, '') AS actor_name, audit_logs.ip_address, audit_logs.user_agent, audit_logs.created_at").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_user_id").
		Where("audit_logs.company_id = ?", companyID).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return SystemLogs{}, err
	}

	payload.Items = make([]SystemLogEntry, 0, len(rows))
	for _, row := range rows {
		payload.Items = append(payload.Items, SystemLogEntry{
			ID:        row.ID.String(),
			Timestamp: row.CreatedAt.UTC().Format(time.RFC3339),
			Action:    row.Action,
			Resource:  row.Resource,
			ActorName: row.ActorName,
			IPAddress: row.IPAddress,
			UserAgent: row.UserAgent,
			Message:   formatSystemLogMessage(row.Action, row.Resource),
			Severity:  inferSystemLogSeverity(row.Action),
		})
	}

	return payload, nil
}

func taskStatusLabel(key string) string {
	switch key {
	case "backlog":
		return "Backlog"
	case "todo":
		return "To Do"
	case "in_progress":
		return "In Progress"
	case "in_review":
		return "In Review"
	case "done":
		return "Done"
	case "archived":
		return "Archived"
	default:
		if key == "" {
			return "Unassigned"
		}
		return key
	}
}

func formatSystemLogMessage(action string, resource string) string {
	switch {
	case action == "" && resource == "":
		return "Audit event recorded"
	case action == "":
		return resource + " activity recorded"
	case resource == "":
		return action + " event recorded"
	default:
		return action + " " + resource
	}
}

func inferSystemLogSeverity(action string) string {
	switch action {
	case "delete", "revoke", "deny", "lock", "failed_login":
		return "error"
	case "update", "assign", "refresh", "rotate":
		return "warn"
	case "create", "login", "accept", "complete":
		return "success"
	default:
		return "info"
	}
}
