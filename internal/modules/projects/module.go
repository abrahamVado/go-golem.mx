package projects

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	rolesmod "github.com/abrahamVado/go-paladin.mx/internal/modules/roles"
	"github.com/abrahamVado/go-paladin.mx/internal/modules/users"
	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/security"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null"`
	TeamID          uuid.UUID `gorm:"type:uuid;not null"`
	ProjectKey      string    `gorm:"column:project_key"`
	Name            string
	Description     string
	Icon            string
	Status          string
	SprintSize      *int       `gorm:"column:sprint_size"`
	SprintStartDate *time.Time `gorm:"column:sprint_start_date"`
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Project) TableName() string { return "projects" }

type Board struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null"`
	Name      string
	IsDefault bool `gorm:"column:is_default"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Board) TableName() string { return "boards" }

type BoardColumn struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null"`
	BoardID   uuid.UUID `gorm:"type:uuid;not null"`
	ColumnKey string    `gorm:"column:column_key"`
	Title     string
	Color     string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (BoardColumn) TableName() string { return "board_columns" }

type Task struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID  `gorm:"type:uuid;not null"`
	ProjectID       uuid.UUID  `gorm:"type:uuid;not null"`
	BoardID         *uuid.UUID `gorm:"column:board_id"`
	ColumnID        *uuid.UUID `gorm:"column:column_id"`
	ParentTaskID    *uuid.UUID `gorm:"column:parent_task_id"`
	TaskNumber      int64      `gorm:"column:task_number"`
	Title           string
	Description     string
	StatusKey       string `gorm:"column:status_key"`
	Priority        string
	DueDate         *time.Time `gorm:"column:due_date"`
	StoryPoints     *float64   `gorm:"column:story_points"`
	Position        float64
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"`
	UpdatedByUserID *uuid.UUID `gorm:"column:updated_by_user_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Task) TableName() string { return "tasks" }

type TaskAssignee struct {
	TaskID uuid.UUID `gorm:"column:task_id"`
	UserID uuid.UUID `gorm:"column:user_id"`
}

func (TaskAssignee) TableName() string { return "task_assignees" }

type TaskWatcher struct {
	TaskID uuid.UUID `gorm:"column:task_id"`
	UserID uuid.UUID `gorm:"column:user_id"`
}

func (TaskWatcher) TableName() string { return "task_watchers" }

type TaskChecklistItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null"`
	TaskID      uuid.UUID `gorm:"column:task_id"`
	Text        string
	Position    int
	IsCompleted bool `gorm:"column:is_completed"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (TaskChecklistItem) TableName() string { return "task_checklist_items" }

type TaskTimeEntry struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompanyID    uuid.UUID  `gorm:"type:uuid;not null"`
	TaskID       uuid.UUID  `gorm:"column:task_id"`
	UserID       *uuid.UUID `gorm:"column:user_id"`
	EntryDate    time.Time  `gorm:"column:entry_date"`
	MinutesSpent int        `gorm:"column:minutes_spent"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (TaskTimeEntry) TableName() string { return "task_time_entries" }

type ProjectMember struct {
	ProjectID       uuid.UUID  `gorm:"column:project_id;primaryKey"`
	UserID          uuid.UUID  `gorm:"column:user_id;primaryKey"`
	Role            string     `gorm:"column:role"`
	InvitedByUserID *uuid.UUID `gorm:"column:invited_by_user_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (ProjectMember) TableName() string { return "project_members" }

type Team struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null"`
	Name            string
	Slug            string
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Team) TableName() string { return "teams" }

type TeamMember struct {
	TeamID        uuid.UUID  `gorm:"column:team_id;primaryKey"`
	UserID        uuid.UUID  `gorm:"column:user_id;primaryKey"`
	AddedByUserID *uuid.UUID `gorm:"column:added_by_user_id"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (TeamMember) TableName() string { return "team_members" }

type ProjectPublicPage struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null"`
	ProjectID       uuid.UUID `gorm:"type:uuid;not null"`
	Slug            string
	Title           string
	HTMLTemplate    string     `gorm:"column:html_template"`
	AccessMode      string     `gorm:"column:access_mode"`
	PasswordHash    *string    `gorm:"column:password_hash"`
	IsEnabled       bool       `gorm:"column:is_enabled"`
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"`
	UpdatedByUserID *uuid.UUID `gorm:"column:updated_by_user_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (ProjectPublicPage) TableName() string { return "project_public_pages" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type projectCreateRequest struct {
	Name            string                     `json:"name" binding:"required"`
	Description     string                     `json:"description"`
	Icon            string                     `json:"icon"`
	SprintSize      *int                       `json:"sprint_size"`
	SprintStartDate string                     `json:"sprint_start_date"`
	Members         []projectMembershipPayload `json:"members"`
}

type teamCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug"`
}

type teamUpdateRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug"`
}

type addTeamMemberRequest struct {
	Email string `json:"email" binding:"required"`
	Name  string `json:"name"`
}

type projectUpdateRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	Icon            string `json:"icon"`
	SprintSize      *int   `json:"sprint_size"`
	SprintStartDate string `json:"sprint_start_date"`
}

type projectMembershipPayload struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

type updateProjectMembersRequest struct {
	Members []projectMembershipPayload `json:"members"`
}

type taskPayload struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	ColumnKey   string             `json:"column_key"`
	Priority    string             `json:"priority"`
	DueDate     string             `json:"due_date"`
	Assignees   []string           `json:"assignees"`
	Watchers    []string           `json:"watchers"`
	OwnerID     *string            `json:"owner_id"`
	StoryPoints *float64           `json:"story_points"`
	Checklist   []checklistPayload `json:"checklist"`
	Tags        []string           `json:"tags"`
	TimeEntries []timeEntryPayload `json:"time_entries"`
}

type checklistPayload struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type timeEntryPayload struct {
	EntryDate    string `json:"entry_date"`
	MinutesSpent int    `json:"minutes_spent"`
}

type columnCreateRequest struct {
	Title string `json:"title" binding:"required"`
	Color string `json:"color"`
}

type moveTaskRequest struct {
	ColumnKey string `json:"column_key" binding:"required"`
}

type publicPageUpdateRequest struct {
	Enabled      bool   `json:"enabled"`
	AccessMode   string `json:"access_mode"`
	Password     string `json:"password"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	HTMLTemplate string `json:"html_template"`
}

type portalAccessRequest struct {
	Password string `json:"password"`
}

type publicProjectRegistrationRequest struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=10"`
	PortalPassword string `json:"portal_password"`
}

type projectSummary struct {
	ID              string `json:"id"`
	TeamID          string `json:"team_id,omitempty"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Icon            string `json:"icon,omitempty"`
	SprintSize      *int   `json:"sprint_size,omitempty"`
	SprintStartDate string `json:"sprint_start_date,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type memberSummary struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type teamSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	MemberCount  int64  `json:"member_count,omitempty"`
	ProjectCount int64  `json:"project_count,omitempty"`
}

type projectMemberSummary struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type boardResponse struct {
	Columns []columnResponse `json:"columns"`
}

type columnResponse struct {
	ID    string         `json:"id"`
	Key   string         `json:"key"`
	Title string         `json:"title"`
	Color string         `json:"color"`
	Tasks []taskResponse `json:"tasks"`
}

type taskResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	Priority    string               `json:"priority"`
	ColumnKey   string               `json:"column_key,omitempty"`
	Assignees   []string             `json:"assignees,omitempty"`
	Watchers    []string             `json:"watchers,omitempty"`
	OwnerID     *string              `json:"owner_id,omitempty"`
	StoryPoints *float64             `json:"story_points,omitempty"`
	DueDate     string               `json:"due_date,omitempty"`
	Checklist   []taskChecklistEntry `json:"checklist,omitempty"`
	Subtasks    []taskChecklistEntry `json:"subtasks,omitempty"`
	TimeEntries []taskTimeEntryDTO   `json:"time_entries,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
}

type taskChecklistEntry struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Position  int    `json:"position"`
	Completed bool   `json:"completed"`
}

type taskTimeEntryDTO struct {
	ID           string `json:"id"`
	EntryDate    string `json:"entry_date"`
	MinutesSpent int    `json:"minutes_spent"`
}

type publicPageResponse struct {
	ProjectID            string `json:"project_id"`
	ProjectName          string `json:"project_name"`
	CompanySlug          string `json:"company_slug"`
	Slug                 string `json:"slug"`
	Title                string `json:"title"`
	HTMLTemplate         string `json:"html_template,omitempty"`
	AccessMode           string `json:"access_mode"`
	IsEnabled            bool   `json:"is_enabled"`
	RequiresPassword     bool   `json:"requires_password"`
	LoginEndpoint        string `json:"login_endpoint"`
	RegisterEndpoint     string `json:"register_endpoint"`
	TicketCreateEndpoint string `json:"ticket_create_endpoint"`
}

type taskHydrated struct {
	Task
	ColumnKey   string
	Assignees   []string
	Watchers    []string
	Checklist   []taskChecklistEntry
	TimeEntries []taskTimeEntryDTO
}

func RegisterRoutes(private *gin.RouterGroup, rbac *rbacmod.Service, h *Handler) {
	private.GET("/teams", middleware.RequirePermission(rbac, "organization:view"), h.ListTeams)
	private.POST("/teams", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "organization:update"), h.CreateTeam)
	private.PUT("/teams/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "organization:update"), h.UpdateTeam)
	private.DELETE("/teams/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "organization:update"), h.DeleteTeam)
	private.GET("/teams/:id/projects", middleware.RequirePermission(rbac, "project:view"), h.ListProjectsForTeam)
	private.GET("/teams/:id/members", middleware.RequirePermission(rbac, "organization:view"), h.ListTeamMembers)
	private.POST("/teams/:id/members", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "member:invite"), h.AddTeamMember)
	private.GET("/projects", middleware.RequirePermission(rbac, "project:view"), h.ListProjects)
	private.POST("/projects", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:create"), h.CreateProject)
	private.GET("/projects/:id", middleware.RequirePermission(rbac, "project:view"), h.GetProject)
	private.PUT("/projects/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:update"), h.UpdateProject)
	private.DELETE("/projects/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:delete"), h.DeleteProject)
	private.GET("/projects/:id/members", middleware.RequirePermission(rbac, "project:view"), h.ListProjectMembers)
	private.PUT("/projects/:id/members", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "member:invite"), h.UpdateProjectMembers)
	private.GET("/projects/:id/board", middleware.RequirePermission(rbac, "project:view"), h.GetBoard)
	private.GET("/projects/:id/public-page", middleware.RequirePermission(rbac, "project:view"), h.GetPublicPage)
	private.PUT("/projects/:id/public-page", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:update"), h.UpdatePublicPage)
	private.POST("/projects/:id/tasks", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "task:create"), h.CreateTask)
	private.POST("/projects/:id/board/columns", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:update"), h.CreateColumn)
	private.DELETE("/projects/:id/board/columns/:columnId", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "project:update"), h.DeleteColumn)
	private.GET("/tasks/:id", middleware.RequirePermission(rbac, "task:view"), h.GetTask)
	private.PUT("/tasks/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "task:update"), h.UpdateTask)
	private.DELETE("/tasks/:id", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "task:delete"), h.DeleteTask)
	private.PATCH("/tasks/:id/move", middleware.RequirePremiumAccount(rbac.DB), middleware.RequirePermission(rbac, "task:update"), h.MoveTask)
}

func RegisterPublicRoutes(api *gin.RouterGroup, h *Handler) {
	public := api.Group("/public/projects")
	public.GET("/:slug", h.GetPublicProjectPage)
	public.POST("/:slug/access", h.AccessPublicProjectPage)
	public.POST("/:slug/register", h.RegisterPublicProjectUser)
}

func (h *Handler) ListTeams(c *gin.Context) {
	companyID := tenancy.CompanyID(c)
	teams, err := h.svc.repo.listTeams(companyID)
	if err != nil {
		response.Internal(c, "Failed to load teams")
		return
	}
	out := make([]teamSummary, 0, len(teams))
	for _, team := range teams {
		out = append(out, teamSummary{
			ID:           team.ID.String(),
			Name:         team.Name,
			Slug:         team.Slug,
			MemberCount:  team.MemberCount,
			ProjectCount: team.ProjectCount,
		})
	}
	response.OK(c, out)
}

func (h *Handler) CreateTeam(c *gin.Context) {
	var req teamCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid team payload")
		return
	}
	team, err := h.svc.createTeam(tenancy.CompanyID(c), tenancy.UserID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, toTeamSummary(team))
}

func (h *Handler) UpdateTeam(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid team id")
		return
	}
	var req teamUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid team payload")
		return
	}
	team, err := h.svc.updateTeam(tenancy.CompanyID(c), teamID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Team not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, toTeamSummary(team))
}

func (h *Handler) DeleteTeam(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid team id")
		return
	}
	if err := h.svc.deleteTeam(tenancy.CompanyID(c), teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Team not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

func (h *Handler) ListTeamMembers(c *gin.Context) {
	companyID := tenancy.CompanyID(c)
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid team id")
		return
	}
	members, err := h.svc.repo.listMembersForTeam(companyID, teamID)
	if err != nil {
		response.Internal(c, "Failed to load members")
		return
	}
	out := make([]memberSummary, 0, len(members))
	for _, member := range members {
		out = append(out, memberSummary{
			UserID: member.ID.String(),
			Name:   member.Name,
			Email:  member.Email,
		})
	}
	response.OK(c, out)
}

func (h *Handler) ListProjectsForTeam(c *gin.Context) {
	companyID := tenancy.CompanyID(c)
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid team id")
		return
	}
	h.respondProjects(c, companyID, &teamID)
}

func (h *Handler) ListProjects(c *gin.Context) {
	teamID, err := h.svc.resolveRequestedTeamID(tenancy.CompanyID(c), c.GetHeader("X-Team-ID"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	h.respondProjects(c, tenancy.CompanyID(c), teamID)
}

func (h *Handler) respondProjects(c *gin.Context, companyID uuid.UUID, teamID *uuid.UUID) {
	projects, err := h.svc.repo.listProjects(companyID, tenancy.UserID(c), teamID)
	if err != nil {
		response.Internal(c, "Failed to load projects")
		return
	}
	out := make([]projectSummary, 0, len(projects))
	for _, project := range projects {
		out = append(out, projectSummary{
			ID:          project.ID.String(),
			TeamID:      project.TeamID.String(),
			Name:        project.Name,
			Description: project.Description,
			Icon:        project.Icon,
			SprintSize:  project.SprintSize,
			CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		})
		if project.SprintStartDate != nil {
			out[len(out)-1].SprintStartDate = project.SprintStartDate.Format("2006-01-02")
		}
	}
	response.OK(c, out)
}

func (h *Handler) AddTeamMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid team id")
		return
	}
	var req addTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid team member payload")
		return
	}
	member, err := h.svc.addTeamMember(tenancy.CompanyID(c), tenancy.UserID(c), teamID, req)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Team not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, member)
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req projectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid project payload")
		return
	}
	project, err := h.svc.createProject(tenancy.CompanyID(c), tenancy.UserID(c), c.GetHeader("X-Team-ID"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, toProjectSummary(project))
}

func (h *Handler) GetProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	project, err := h.svc.getProject(tenancy.CompanyID(c), tenancy.UserID(c), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.Internal(c, "Failed to load project")
		return
	}
	response.OK(c, toProjectSummary(project))
}

func (h *Handler) UpdateProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	var req projectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid project payload")
		return
	}
	project, err := h.svc.updateProject(tenancy.CompanyID(c), tenancy.UserID(c), projectID, req)
	if err != nil {
		switch err.Error() {
		case "project not available", "only project admins can update this project":
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, toProjectSummary(project))
}

func (h *Handler) DeleteProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	if err := h.svc.deleteProject(tenancy.CompanyID(c), tenancy.UserID(c), projectID); err != nil {
		switch err.Error() {
		case "project not available", "only project admins can delete this project":
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

func (h *Handler) ListProjectMembers(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	members, err := h.svc.listProjectMembers(tenancy.CompanyID(c), tenancy.UserID(c), projectID)
	if err != nil {
		switch err.Error() {
		case "project not available":
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, members)
}

func (h *Handler) UpdateProjectMembers(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	var req updateProjectMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid project members payload")
		return
	}
	members, err := h.svc.updateProjectMembers(tenancy.CompanyID(c), tenancy.UserID(c), projectID, req.Members)
	if err != nil {
		switch err.Error() {
		case "project not available", "only project admins can manage project members":
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, members)
}

func (h *Handler) GetBoard(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	board, err := h.svc.getBoard(tenancy.CompanyID(c), tenancy.UserID(c), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.Internal(c, "Failed to load board")
		return
	}
	response.OK(c, board)
}

func (h *Handler) GetPublicPage(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	page, err := h.svc.getOrCreatePublicPage(tenancy.CompanyID(c), tenancy.UserID(c), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.Internal(c, "Failed to load public page")
		return
	}
	resp, err := h.svc.buildPublicPageResponse(page, true)
	if err != nil {
		response.Internal(c, "Failed to load public page")
		return
	}
	response.OK(c, resp)
}

func (h *Handler) UpdatePublicPage(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	var req publicPageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid public page payload")
		return
	}
	page, err := h.svc.updatePublicPage(tenancy.CompanyID(c), tenancy.UserID(c), projectID, req)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Project not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	resp, err := h.svc.buildPublicPageResponse(page, true)
	if err != nil {
		response.Internal(c, "Failed to load public page")
		return
	}
	response.OK(c, resp)
}

func (h *Handler) GetPublicProjectPage(c *gin.Context) {
	page, err := h.svc.getPublicPageBySlug(strings.TrimSpace(c.Param("slug")), "")
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Public project page not found")
		case err.Error() == "public page password required":
			response.OK(c, gin.H{
				"slug":              strings.TrimSpace(c.Param("slug")),
				"requires_password": true,
				"access_mode":       "password_protected",
			})
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	resp, err := h.svc.buildPublicPageResponse(page, page.AccessMode == "public")
	if err != nil {
		response.Internal(c, "Failed to load public page")
		return
	}
	response.OK(c, resp)
}

func (h *Handler) AccessPublicProjectPage(c *gin.Context) {
	var req portalAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid portal access payload")
		return
	}
	page, err := h.svc.getPublicPageBySlug(strings.TrimSpace(c.Param("slug")), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Public project page not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	resp, err := h.svc.buildPublicPageResponse(page, true)
	if err != nil {
		response.Internal(c, "Failed to load public page")
		return
	}
	response.OK(c, resp)
}

func (h *Handler) RegisterPublicProjectUser(c *gin.Context) {
	var req publicProjectRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid project registration payload")
		return
	}
	member, err := h.svc.registerPublicProjectUser(strings.TrimSpace(c.Param("slug")), req)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Public project page not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, member)
}

func (h *Handler) CreateTask(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	var req taskPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid task payload")
		return
	}
	task, err := h.svc.createTask(tenancy.CompanyID(c), tenancy.UserID(c), projectID, req)
	if err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, task)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	task, err := h.svc.getTask(tenancy.CompanyID(c), tenancy.UserID(c), taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Task not found")
			return
		}
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.Internal(c, "Failed to load task")
		return
	}
	response.OK(c, task)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	var req taskPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid task payload")
		return
	}
	task, err := h.svc.updateTask(tenancy.CompanyID(c), tenancy.UserID(c), taskID, req)
	if err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	if err := h.svc.deleteTask(tenancy.CompanyID(c), tenancy.UserID(c), taskID); err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

func (h *Handler) MoveTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	var req moveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid move payload")
		return
	}
	task, err := h.svc.moveTask(tenancy.CompanyID(c), tenancy.UserID(c), taskID, req.ColumnKey)
	if err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, task)
}

func (h *Handler) CreateColumn(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	var req columnCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid column payload")
		return
	}
	column, err := h.svc.createColumn(tenancy.CompanyID(c), tenancy.UserID(c), projectID, req)
	if err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, column)
}

func (h *Handler) DeleteColumn(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	columnID, err := uuid.Parse(c.Param("columnId"))
	if err != nil {
		response.BadRequest(c, "Invalid column id")
		return
	}
	if err := h.svc.deleteColumn(tenancy.CompanyID(c), tenancy.UserID(c), projectID, columnID); err != nil {
		if err.Error() == "project not available" {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

type teamListRow struct {
	Team
	MemberCount  int64 `gorm:"column:member_count"`
	ProjectCount int64 `gorm:"column:project_count"`
}

func (r *Repository) listAllCompanyMembers(companyID uuid.UUID) ([]users.User, error) {
	var members []users.User
	err := r.db.Where("company_id = ? AND deleted_at IS NULL AND status = ?", companyID, "active").Order("name asc").Find(&members).Error
	return members, err
}

func (r *Repository) listTeams(companyID uuid.UUID) ([]teamListRow, error) {
	var teams []teamListRow
	err := r.db.Table("teams t").
		Select(`
			t.*,
			COUNT(DISTINCT tm.user_id) AS member_count,
			COUNT(DISTINCT p.id) AS project_count
		`).
		Joins("LEFT JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL").
		Joins("LEFT JOIN projects p ON p.team_id = t.id AND p.deleted_at IS NULL").
		Where("t.company_id = ? AND t.deleted_at IS NULL", companyID).
		Group("t.id").
		Order("t.created_at asc, t.name asc").
		Scan(&teams).Error
	return teams, err
}

func (r *Repository) getTeam(companyID, teamID uuid.UUID) (Team, error) {
	var team Team
	err := r.db.Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, teamID).First(&team).Error
	return team, err
}

func (r *Repository) firstTeam(companyID uuid.UUID) (Team, error) {
	var team Team
	err := r.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Order("created_at asc, name asc").First(&team).Error
	return team, err
}

func (r *Repository) listMembersForTeam(companyID, teamID uuid.UUID) ([]users.User, error) {
	if _, err := r.getTeam(companyID, teamID); err != nil {
		return nil, err
	}
	var members []users.User
	err := r.db.Table("team_members tm").
		Select("u.*").
		Joins("JOIN users u ON u.id = tm.user_id").
		Where("tm.team_id = ? AND tm.deleted_at IS NULL", teamID).
		Where("u.company_id = ? AND u.deleted_at IS NULL AND u.status = ?", companyID, "active").
		Order("u.name asc, u.email asc").
		Scan(&members).Error
	return members, err
}

func (r *Repository) findUserByEmail(companyID uuid.UUID, email string) (users.User, error) {
	var user users.User
	err := r.db.Where("company_id = ? AND LOWER(email) = ? AND deleted_at IS NULL", companyID, strings.ToLower(strings.TrimSpace(email))).First(&user).Error
	return user, err
}

func (r *Repository) listProjects(companyID, userID uuid.UUID, teamID *uuid.UUID) ([]Project, error) {
	var projects []Project
	query := r.db.Where("projects.company_id = ? AND projects.deleted_at IS NULL", companyID)
	if teamID != nil {
		query = query.Where("projects.team_id = ?", *teamID)
	}
	if !r.userCanManageAllProjects(companyID, userID) {
		query = query.Joins(
			"JOIN project_members pm ON pm.project_id = projects.id AND pm.user_id = ? AND pm.deleted_at IS NULL",
			userID,
		)
	}
	err := query.Order("projects.created_at desc").Find(&projects).Error
	return projects, err
}

func (r *Repository) userCanManageAllProjects(companyID, userID uuid.UUID) bool {
	var count int64
	err := r.db.
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("LEFT JOIN permissions p ON p.id = rp.permission_id").
		Where("ur.user_id = ? AND ur.company_id = ? AND ur.deleted_at IS NULL", userID, companyID).
		Where(
			r.db.
				Where("r.name = ?", rbacmod.SystemRoleOwner).
				Or("p.name IN ?", []string{"project:create", "project:update", "member:invite"}),
		).
		Count(&count).
		Error
	if err != nil {
		return false
	}
	return count > 0
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	normalized = nonAlnum.ReplaceAllString(normalized, "-")
	return strings.Trim(normalized, "-")
}

func toProjectSummary(project Project) projectSummary {
	summary := projectSummary{
		ID:          project.ID.String(),
		TeamID:      project.TeamID.String(),
		Name:        project.Name,
		Description: project.Description,
		Icon:        project.Icon,
		SprintSize:  project.SprintSize,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
	}
	if project.SprintStartDate != nil {
		summary.SprintStartDate = project.SprintStartDate.Format("2006-01-02")
	}
	return summary
}

func toTeamSummary(team Team) teamSummary {
	return teamSummary{
		ID:   team.ID.String(),
		Name: team.Name,
		Slug: team.Slug,
	}
}

func parseProjectDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, errors.New("invalid sprint start date")
	}
	return &parsed, nil
}

func normalizeProjectRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "admin", "admins":
		return "admin"
	default:
		return "member"
	}
}

func normalizeTeamSlug(name, provided string) string {
	if slug := slugify(provided); slug != "" {
		return slug
	}
	return slugify(name)
}

func (s *Service) resolveRequestedTeamID(companyID uuid.UUID, rawTeamID string) (*uuid.UUID, error) {
	rawTeamID = strings.TrimSpace(rawTeamID)
	if rawTeamID == "" {
		team, err := s.repo.firstTeam(companyID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("no teams are available")
			}
			return nil, err
		}
		return &team.ID, nil
	}
	teamID, err := uuid.Parse(rawTeamID)
	if err != nil {
		return nil, errors.New("invalid team id")
	}
	if _, err := s.repo.getTeam(companyID, teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}
	return &teamID, nil
}

func (s *Service) ensureProjectAccess(tx *gorm.DB, companyID, userID, projectID uuid.UUID) (Project, error) {
	var project Project
	if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", projectID, companyID).First(&project).Error; err != nil {
		return Project{}, err
	}
	if s.repo.userCanManageAllProjects(companyID, userID) {
		return project, nil
	}
	var count int64
	if err := tx.Model(&ProjectMember{}).Where("project_id = ? AND user_id = ? AND deleted_at IS NULL", projectID, userID).Count(&count).Error; err != nil {
		return Project{}, err
	}
	if count == 0 {
		return Project{}, errors.New("project not available")
	}
	return project, nil
}

func (s *Service) createTeam(companyID, userID uuid.UUID, req teamCreateRequest) (Team, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Team{}, errors.New("team name is required")
	}
	slug := normalizeTeamSlug(name, req.Slug)
	if slug == "" {
		return Team{}, errors.New("team slug is required")
	}
	team := Team{
		ID:              uuid.New(),
		CompanyID:       companyID,
		Name:            name,
		Slug:            slug,
		CreatedByUserID: &userID,
	}
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Team{}).
			Where("company_id = ? AND slug = ? AND deleted_at IS NULL", companyID, slug).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("team slug already exists")
		}
		if err := tx.Create(&team).Error; err != nil {
			return err
		}
		return tx.Exec(`
			INSERT INTO team_members (team_id, user_id, added_by_user_id, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, NOW(), NOW(), NULL)
			ON CONFLICT (team_id, user_id)
			DO UPDATE SET deleted_at = NULL, updated_at = NOW()
		`, team.ID, userID, userID).Error
	})
	return team, err
}

func (s *Service) updateTeam(companyID, teamID uuid.UUID, req teamUpdateRequest) (Team, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Team{}, errors.New("team name is required")
	}
	slug := normalizeTeamSlug(name, req.Slug)
	if slug == "" {
		return Team{}, errors.New("team slug is required")
	}
	var team Team
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, teamID).First(&team).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&Team{}).
			Where("company_id = ? AND slug = ? AND id <> ? AND deleted_at IS NULL", companyID, slug, teamID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("team slug already exists")
		}
		team.Name = name
		team.Slug = slug
		return tx.Save(&team).Error
	})
	return team, err
}

func (s *Service) deleteTeam(companyID, teamID uuid.UUID) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		var team Team
		if err := tx.Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, teamID).First(&team).Error; err != nil {
			return err
		}
		var teamCount int64
		if err := tx.Model(&Team{}).Where("company_id = ? AND deleted_at IS NULL", companyID).Count(&teamCount).Error; err != nil {
			return err
		}
		if teamCount <= 1 {
			return errors.New("at least one team must remain")
		}
		var projectCount int64
		if err := tx.Model(&Project{}).Where("company_id = ? AND team_id = ? AND deleted_at IS NULL", companyID, teamID).Count(&projectCount).Error; err != nil {
			return err
		}
		if projectCount > 0 {
			return errors.New("move or delete team projects before deleting this team")
		}
		if err := tx.Where("team_id = ? AND deleted_at IS NULL", teamID).Delete(&TeamMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&team).Error
	})
}

func (s *Service) addTeamMember(companyID, actorUserID, teamID uuid.UUID, req addTeamMemberRequest) (memberSummary, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return memberSummary{}, errors.New("member email is required")
	}
	var created memberSummary
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.repo.getTeam(companyID, teamID); err != nil {
			return err
		}
		user, err := s.repo.findUserByEmail(companyID, email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user must already exist in this tenant before being added to a team")
			}
			return err
		}
		if err := tx.Exec(`
			INSERT INTO team_members (team_id, user_id, added_by_user_id, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, NOW(), NOW(), NULL)
			ON CONFLICT (team_id, user_id)
			DO UPDATE SET deleted_at = NULL, updated_at = NOW()
		`, teamID, user.ID, actorUserID).Error; err != nil {
			return err
		}
		created = memberSummary{
			UserID: user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
		}
		return nil
	})
	return created, err
}

func (s *Service) ensureProjectAdmin(tx *gorm.DB, companyID, userID, projectID uuid.UUID) (Project, error) {
	project, err := s.ensureProjectAccess(tx, companyID, userID, projectID)
	if err != nil {
		return Project{}, err
	}
	if s.repo.userCanManageAllProjects(companyID, userID) {
		return project, nil
	}
	var member ProjectMember
	if err := tx.Where("project_id = ? AND user_id = ? AND deleted_at IS NULL", projectID, userID).First(&member).Error; err != nil {
		return Project{}, errors.New("only project admins can manage project members")
	}
	if normalizeProjectRole(member.Role) != "admin" {
		return Project{}, errors.New("only project admins can manage project members")
	}
	return project, nil
}

func (s *Service) createProject(companyID, userID uuid.UUID, rawTeamID string, req projectCreateRequest) (Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Project{}, errors.New("project name is required")
	}
	if req.SprintSize != nil && *req.SprintSize <= 0 {
		return Project{}, errors.New("sprint size must be greater than zero")
	}
	sprintStartDate, err := parseProjectDate(req.SprintStartDate)
	if err != nil {
		return Project{}, err
	}
	keyBase := strings.ToUpper(strings.ReplaceAll(slugify(name), "-", ""))
	if keyBase == "" {
		keyBase = "PROJECT"
	}
	project := Project{
		ID:              uuid.New(),
		CompanyID:       companyID,
		TeamID:          uuid.Nil,
		ProjectKey:      keyBase,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		Icon:            strings.TrimSpace(req.Icon),
		Status:          "active",
		SprintSize:      req.SprintSize,
		SprintStartDate: sprintStartDate,
		CreatedByUserID: &userID,
	}
	teamID, err := s.resolveRequestedTeamID(companyID, rawTeamID)
	if err != nil {
		return Project{}, err
	}
	project.TeamID = *teamID
	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		project.ProjectKey = s.nextProjectKey(tx, companyID, project.ProjectKey)
		if err := tx.Create(&project).Error; err != nil {
			return err
		}
		if err := tx.Create(&ProjectMember{
			ProjectID:       project.ID,
			UserID:          userID,
			Role:            "admin",
			InvitedByUserID: &userID,
		}).Error; err != nil {
			return err
		}
		if len(req.Members) > 0 {
			if _, err := s.replaceProjectMembers(tx, companyID, userID, project, req.Members); err != nil {
				return err
			}
		}
		_, err := ensureDefaultBoard(tx, companyID, project.ID)
		return err
	})
	return project, err
}

func (s *Service) nextProjectKey(tx *gorm.DB, companyID uuid.UUID, base string) string {
	key := base
	for i := 1; i < 1000; i++ {
		var count int64
		tx.Model(&Project{}).Where("company_id = ? AND project_key = ?", companyID, key).Count(&count)
		if count == 0 {
			return key
		}
		key = fmt.Sprintf("%s%d", base, i+1)
	}
	return fmt.Sprintf("%s%s", base, time.Now().Format("150405"))
}

func ensureDefaultBoard(tx *gorm.DB, companyID, projectID uuid.UUID) (Board, error) {
	var board Board
	err := tx.Where("company_id = ? AND project_id = ? AND is_default = ? AND deleted_at IS NULL", companyID, projectID, true).First(&board).Error
	if err == nil {
		if err := ensureDefaultColumns(tx, companyID, board.ID); err != nil {
			return Board{}, err
		}
		return board, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Board{}, err
	}
	board = Board{
		ID:        uuid.New(),
		CompanyID: companyID,
		ProjectID: projectID,
		Name:      "Main Board",
		IsDefault: true,
	}
	if err := tx.Create(&board).Error; err != nil {
		return Board{}, err
	}
	if err := ensureDefaultColumns(tx, companyID, board.ID); err != nil {
		return Board{}, err
	}
	return board, nil
}

func ensureDefaultColumns(tx *gorm.DB, companyID, boardID uuid.UUID) error {
	var count int64
	if err := tx.Model(&BoardColumn{}).Where("board_id = ? AND deleted_at IS NULL", boardID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	columns := []BoardColumn{
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "todo", Title: "To Do", Color: "blue", Position: 1},
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "in_progress", Title: "In Progress", Color: "amber", Position: 2},
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "done", Title: "Done", Color: "green", Position: 3},
	}
	return tx.Create(&columns).Error
}

func (s *Service) getProject(companyID, userID, projectID uuid.UUID) (Project, error) {
	var project Project
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		found, err := s.ensureProjectAccess(tx, companyID, userID, projectID)
		if err != nil {
			return err
		}
		project = found
		return nil
	})
	return project, err
}

func (s *Service) updateProject(companyID, userID, projectID uuid.UUID, req projectUpdateRequest) (Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Project{}, errors.New("project name is required")
	}
	if req.SprintSize != nil && *req.SprintSize <= 0 {
		return Project{}, errors.New("sprint size must be greater than zero")
	}
	sprintStartDate, err := parseProjectDate(req.SprintStartDate)
	if err != nil {
		return Project{}, err
	}
	var updated Project
	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		project, err := s.ensureProjectAdmin(tx, companyID, userID, projectID)
		if err != nil {
			if err.Error() == "only project admins can manage project members" {
				return errors.New("only project admins can update this project")
			}
			return err
		}
		project.Name = name
		project.Description = strings.TrimSpace(req.Description)
		project.Icon = strings.TrimSpace(req.Icon)
		project.SprintSize = req.SprintSize
		project.SprintStartDate = sprintStartDate
		if err := tx.Save(&project).Error; err != nil {
			return err
		}
		updated = project
		return nil
	})
	return updated, err
}

func (s *Service) deleteProject(companyID, userID, projectID uuid.UUID) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		project, err := s.ensureProjectAdmin(tx, companyID, userID, projectID)
		if err != nil {
			if err.Error() == "only project admins can manage project members" {
				return errors.New("only project admins can delete this project")
			}
			return err
		}
		return tx.Delete(&project).Error
	})
}

func (s *Service) listProjectMembers(companyID, userID, projectID uuid.UUID) ([]projectMemberSummary, error) {
	members := []projectMemberSummary{}
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAccess(tx, companyID, userID, projectID); err != nil {
			return err
		}
		type row struct {
			UserID uuid.UUID
			Name   string
			Email  string
			Role   string
		}
		var rows []row
		if err := tx.Table("project_members pm").
			Select("u.id as user_id, u.name, u.email, pm.role").
			Joins("JOIN users u ON u.id = pm.user_id").
			Where("pm.project_id = ? AND pm.deleted_at IS NULL AND u.deleted_at IS NULL", projectID).
			Order("CASE WHEN pm.role = 'admin' THEN 0 ELSE 1 END, u.name asc, u.email asc").
			Scan(&rows).Error; err != nil {
			return err
		}
		members = make([]projectMemberSummary, 0, len(rows))
		for _, row := range rows {
			members = append(members, projectMemberSummary{
				UserID: row.UserID.String(),
				Name:   row.Name,
				Email:  row.Email,
				Role:   normalizeProjectRole(row.Role),
			})
		}
		return nil
	})
	return members, err
}

func (s *Service) updateProjectMembers(companyID, userID, projectID uuid.UUID, members []projectMembershipPayload) ([]projectMemberSummary, error) {
	var updated []projectMemberSummary
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		project, err := s.ensureProjectAdmin(tx, companyID, userID, projectID)
		if err != nil {
			return err
		}
		updated, err = s.replaceProjectMembers(tx, companyID, userID, project, members)
		return err
	})
	return updated, err
}

func (s *Service) replaceProjectMembers(tx *gorm.DB, companyID, userID uuid.UUID, project Project, members []projectMembershipPayload) ([]projectMemberSummary, error) {
	allowedMembers, err := s.repo.listAllCompanyMembers(companyID)
	if err != nil {
		return nil, err
	}
	allowedByID := make(map[uuid.UUID]users.User, len(allowedMembers))
	for _, member := range allowedMembers {
		allowedByID[member.ID] = member
	}
	targets := make(map[uuid.UUID]string)
	targets[userID] = "admin"
	if project.CreatedByUserID != nil {
		targets[*project.CreatedByUserID] = "admin"
	}
	for _, member := range members {
		parsed, err := uuid.Parse(strings.TrimSpace(member.UserID))
		if err != nil {
			return nil, errors.New("invalid project member")
		}
		if _, ok := allowedByID[parsed]; !ok {
			return nil, errors.New("project members must belong to the company")
		}
		role := normalizeProjectRole(member.Role)
		if existing, ok := targets[parsed]; !ok || existing != "admin" {
			targets[parsed] = role
		}
	}
	if err := tx.Unscoped().Where("project_id = ?", project.ID).Delete(&ProjectMember{}).Error; err != nil {
		return nil, err
	}
	rows := make([]ProjectMember, 0, len(targets))
	summaries := make([]projectMemberSummary, 0, len(targets))
	for memberID, role := range targets {
		invitedBy := userID
		rows = append(rows, ProjectMember{
			ProjectID:       project.ID,
			UserID:          memberID,
			Role:            role,
			InvitedByUserID: &invitedBy,
		})
		user := allowedByID[memberID]
		summaries = append(summaries, projectMemberSummary{
			UserID: memberID.String(),
			Name:   user.Name,
			Email:  user.Email,
			Role:   role,
		})
	}
	if len(rows) > 0 {
		if err := tx.Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	return summaries, nil
}

func (s *Service) getBoard(companyID, userID, projectID uuid.UUID) (boardResponse, error) {
	var out boardResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAccess(tx, companyID, userID, projectID); err != nil {
			return err
		}
		board, err := ensureDefaultBoard(tx, companyID, projectID)
		if err != nil {
			return err
		}
		var columns []BoardColumn
		if err := tx.Where("board_id = ? AND deleted_at IS NULL", board.ID).Order("position asc").Find(&columns).Error; err != nil {
			return err
		}
		var tasks []Task
		if err := tx.Where("project_id = ? AND company_id = ? AND deleted_at IS NULL", projectID, companyID).Order("position asc, created_at asc").Find(&tasks).Error; err != nil {
			return err
		}
		hydrated, err := hydrateTasks(tx, companyID, tasks, columns)
		if err != nil {
			return err
		}
		grouped := map[string][]taskResponse{}
		for _, task := range hydrated {
			grouped[task.ColumnKey] = append(grouped[task.ColumnKey], toTaskResponse(task))
		}
		out.Columns = make([]columnResponse, 0, len(columns))
		for _, column := range columns {
			tasksForColumn := grouped[column.ColumnKey]
			if tasksForColumn == nil {
				tasksForColumn = []taskResponse{}
			}
			out.Columns = append(out.Columns, columnResponse{
				ID:    column.ID.String(),
				Key:   column.ColumnKey,
				Title: column.Title,
				Color: column.Color,
				Tasks: tasksForColumn,
			})
		}
		return nil
	})
	return out, err
}

func hydrateTasks(tx *gorm.DB, companyID uuid.UUID, tasks []Task, columns []BoardColumn) ([]taskHydrated, error) {
	columnByID := map[uuid.UUID]string{}
	for _, col := range columns {
		columnByID[col.ID] = col.ColumnKey
	}
	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	assigneesByTask := map[uuid.UUID][]string{}
	watchersByTask := map[uuid.UUID][]string{}
	checklistByTask := map[uuid.UUID][]taskChecklistEntry{}
	timeByTask := map[uuid.UUID][]taskTimeEntryDTO{}
	if len(taskIDs) > 0 {
		var assignees []TaskAssignee
		if err := tx.Where("task_id IN ?", taskIDs).Find(&assignees).Error; err != nil {
			return nil, err
		}
		for _, row := range assignees {
			assigneesByTask[row.TaskID] = append(assigneesByTask[row.TaskID], row.UserID.String())
		}
		var watchers []TaskWatcher
		if err := tx.Where("task_id IN ?", taskIDs).Find(&watchers).Error; err != nil {
			return nil, err
		}
		for _, row := range watchers {
			watchersByTask[row.TaskID] = append(watchersByTask[row.TaskID], row.UserID.String())
		}
		var checklist []TaskChecklistItem
		if err := tx.Where("company_id = ? AND task_id IN ? AND deleted_at IS NULL", companyID, taskIDs).Order("position asc").Find(&checklist).Error; err != nil {
			return nil, err
		}
		for _, item := range checklist {
			checklistByTask[item.TaskID] = append(checklistByTask[item.TaskID], taskChecklistEntry{
				ID: item.ID.String(), Text: item.Text, Position: item.Position, Completed: item.IsCompleted,
			})
		}
		var timeEntries []TaskTimeEntry
		if err := tx.Where("company_id = ? AND task_id IN ? AND deleted_at IS NULL", companyID, taskIDs).Order("entry_date asc").Find(&timeEntries).Error; err != nil {
			return nil, err
		}
		for _, entry := range timeEntries {
			timeByTask[entry.TaskID] = append(timeByTask[entry.TaskID], taskTimeEntryDTO{
				ID: entry.ID.String(), EntryDate: entry.EntryDate.Format("2006-01-02"), MinutesSpent: entry.MinutesSpent,
			})
		}
	}
	out := make([]taskHydrated, 0, len(tasks))
	for _, task := range tasks {
		columnKey := task.StatusKey
		if task.ColumnID != nil {
			if key, ok := columnByID[*task.ColumnID]; ok {
				columnKey = key
			}
		}
		out = append(out, taskHydrated{
			Task:        task,
			ColumnKey:   columnKey,
			Assignees:   assigneesByTask[task.ID],
			Watchers:    watchersByTask[task.ID],
			Checklist:   checklistByTask[task.ID],
			TimeEntries: timeByTask[task.ID],
		})
	}
	return out, nil
}

func toTaskResponse(task taskHydrated) taskResponse {
	var ownerID *string
	if task.CreatedByUserID != nil {
		value := task.CreatedByUserID.String()
		ownerID = &value
	}
	resp := taskResponse{
		ID:          task.ID.String(),
		Title:       task.Title,
		Description: task.Description,
		Priority:    strings.ToUpper(task.Priority),
		ColumnKey:   task.ColumnKey,
		Assignees:   task.Assignees,
		Watchers:    task.Watchers,
		OwnerID:     ownerID,
		StoryPoints: task.StoryPoints,
		Checklist:   task.Checklist,
		Subtasks:    task.Checklist,
		TimeEntries: task.TimeEntries,
	}
	if task.DueDate != nil {
		resp.DueDate = task.DueDate.Format(time.RFC3339)
	}
	return resp
}

func parseDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	layouts := []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05Z07:00"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("invalid due date")
}

func priorityOrDefault(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "LOW", "MEDIUM", "HIGH", "URGENT":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "medium"
	}
}

func nextTaskPosition(tx *gorm.DB, projectID uuid.UUID, columnID *uuid.UUID) float64 {
	query := tx.Model(&Task{}).Where("project_id = ? AND deleted_at IS NULL", projectID)
	if columnID != nil {
		query = query.Where("column_id = ?", *columnID)
	}
	var maxPosition float64
	query.Select("COALESCE(MAX(position), 0)").Scan(&maxPosition)
	return maxPosition + 1
}

func (s *Service) createTask(companyID, userID, projectID uuid.UUID, req taskPayload) (taskResponse, error) {
	var result taskResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAccess(tx, companyID, userID, projectID); err != nil {
			return err
		}
		board, err := ensureDefaultBoard(tx, companyID, projectID)
		if err != nil {
			return err
		}
		column, err := findColumnByKey(tx, board.ID, req.ColumnKey)
		if err != nil {
			return err
		}
		var maxNumber int64
		tx.Model(&Task{}).Where("company_id = ?", companyID).Select("COALESCE(MAX(task_number), 0)").Scan(&maxNumber)
		dueDate, err := parseDate(req.DueDate)
		if err != nil {
			return err
		}
		task := Task{
			ID:              uuid.New(),
			CompanyID:       companyID,
			ProjectID:       projectID,
			BoardID:         &board.ID,
			ColumnID:        &column.ID,
			TaskNumber:      maxNumber + 1,
			Title:           strings.TrimSpace(req.Title),
			Description:     strings.TrimSpace(req.Description),
			StatusKey:       column.ColumnKey,
			Priority:        priorityOrDefault(req.Priority),
			DueDate:         dueDate,
			StoryPoints:     req.StoryPoints,
			Position:        nextTaskPosition(tx, projectID, &column.ID),
			CreatedByUserID: &userID,
			UpdatedByUserID: &userID,
		}
		if task.Title == "" {
			return errors.New("task title is required")
		}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		if err := replaceTaskRelations(tx, companyID, task.ID, req); err != nil {
			return err
		}
		hydrated, err := hydrateTasks(tx, companyID, []Task{task}, []BoardColumn{column})
		if err != nil {
			return err
		}
		result = toTaskResponse(hydrated[0])
		return nil
	})
	return result, err
}

func (s *Service) getTask(companyID, userID, taskID uuid.UUID) (taskResponse, error) {
	var out taskResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", taskID, companyID).First(&task).Error; err != nil {
			return err
		}
		if _, err := s.ensureProjectAccess(tx, companyID, userID, task.ProjectID); err != nil {
			return err
		}
		var columns []BoardColumn
		if task.BoardID != nil {
			if err := tx.Where("board_id = ? AND deleted_at IS NULL", *task.BoardID).Find(&columns).Error; err != nil {
				return err
			}
		}
		hydrated, err := hydrateTasks(tx, companyID, []Task{task}, columns)
		if err != nil {
			return err
		}
		out = toTaskResponse(hydrated[0])
		return nil
	})
	return out, err
}

func (s *Service) updateTask(companyID, userID, taskID uuid.UUID, req taskPayload) (taskResponse, error) {
	var out taskResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", taskID, companyID).First(&task).Error; err != nil {
			return err
		}
		if _, err := s.ensureProjectAccess(tx, companyID, userID, task.ProjectID); err != nil {
			return err
		}
		var column BoardColumn
		if task.BoardID != nil {
			found, err := findColumnByKey(tx, *task.BoardID, req.ColumnKey)
			if err != nil {
				return err
			}
			column = found
			task.ColumnID = &column.ID
			task.StatusKey = column.ColumnKey
		}
		dueDate, err := parseDate(req.DueDate)
		if err != nil {
			return err
		}
		task.Title = strings.TrimSpace(req.Title)
		task.Description = strings.TrimSpace(req.Description)
		task.Priority = priorityOrDefault(req.Priority)
		task.DueDate = dueDate
		task.StoryPoints = req.StoryPoints
		task.UpdatedByUserID = &userID
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if err := replaceTaskRelations(tx, companyID, task.ID, req); err != nil {
			return err
		}
		columns := []BoardColumn{}
		if column.ID != uuid.Nil {
			columns = append(columns, column)
		}
		hydrated, err := hydrateTasks(tx, companyID, []Task{task}, columns)
		if err != nil {
			return err
		}
		out = toTaskResponse(hydrated[0])
		return nil
	})
	return out, err
}

func (s *Service) moveTask(companyID, userID, taskID uuid.UUID, columnKey string) (taskResponse, error) {
	var out taskResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", taskID, companyID).First(&task).Error; err != nil {
			return err
		}
		if _, err := s.ensureProjectAccess(tx, companyID, userID, task.ProjectID); err != nil {
			return err
		}
		if task.BoardID == nil {
			return errors.New("task has no board")
		}
		column, err := findColumnByKey(tx, *task.BoardID, columnKey)
		if err != nil {
			return err
		}
		task.ColumnID = &column.ID
		task.StatusKey = column.ColumnKey
		task.UpdatedByUserID = &userID
		task.Position = nextTaskPosition(tx, task.ProjectID, &column.ID)
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		hydrated, err := hydrateTasks(tx, companyID, []Task{task}, []BoardColumn{column})
		if err != nil {
			return err
		}
		out = toTaskResponse(hydrated[0])
		return nil
	})
	return out, err
}

func (s *Service) deleteTask(companyID, userID, taskID uuid.UUID) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", taskID, companyID).First(&task).Error; err != nil {
			return err
		}
		if _, err := s.ensureProjectAccess(tx, companyID, userID, task.ProjectID); err != nil {
			return err
		}
		return tx.Delete(&task).Error
	})
}

func replaceTaskRelations(tx *gorm.DB, companyID, taskID uuid.UUID, req taskPayload) error {
	if err := tx.Where("task_id = ?", taskID).Delete(&TaskAssignee{}).Error; err != nil {
		return err
	}
	if len(req.Assignees) > 0 {
		rows := make([]TaskAssignee, 0, len(req.Assignees))
		for _, raw := range req.Assignees {
			id, err := uuid.Parse(raw)
			if err != nil {
				continue
			}
			rows = append(rows, TaskAssignee{TaskID: taskID, UserID: id})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Where("task_id = ?", taskID).Delete(&TaskWatcher{}).Error; err != nil {
		return err
	}
	if len(req.Watchers) > 0 {
		rows := make([]TaskWatcher, 0, len(req.Watchers))
		for _, raw := range req.Watchers {
			id, err := uuid.Parse(raw)
			if err != nil {
				continue
			}
			rows = append(rows, TaskWatcher{TaskID: taskID, UserID: id})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Where("task_id = ?", taskID).Delete(&TaskChecklistItem{}).Error; err != nil {
		return err
	}
	if len(req.Checklist) > 0 {
		items := make([]TaskChecklistItem, 0, len(req.Checklist))
		for index, item := range req.Checklist {
			if strings.TrimSpace(item.Text) == "" {
				continue
			}
			items = append(items, TaskChecklistItem{
				ID: uuid.New(), CompanyID: companyID, TaskID: taskID, Text: strings.TrimSpace(item.Text), Position: index + 1, IsCompleted: item.Completed,
			})
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Where("task_id = ?", taskID).Delete(&TaskTimeEntry{}).Error; err != nil {
		return err
	}
	if len(req.TimeEntries) > 0 {
		rows := make([]TaskTimeEntry, 0, len(req.TimeEntries))
		for _, entry := range req.TimeEntries {
			date, err := time.Parse("2006-01-02", entry.EntryDate)
			if err != nil || entry.MinutesSpent <= 0 {
				continue
			}
			rows = append(rows, TaskTimeEntry{ID: uuid.New(), CompanyID: companyID, TaskID: taskID, EntryDate: date, MinutesSpent: entry.MinutesSpent})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func findColumnByKey(tx *gorm.DB, boardID uuid.UUID, key string) (BoardColumn, error) {
	lookup := strings.TrimSpace(key)
	if lookup == "" {
		lookup = "todo"
	}
	var column BoardColumn
	err := tx.Where("board_id = ? AND column_key = ? AND deleted_at IS NULL", boardID, lookup).First(&column).Error
	return column, err
}

func (s *Service) createColumn(companyID, userID, projectID uuid.UUID, req columnCreateRequest) (columnResponse, error) {
	var out columnResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAccess(tx, companyID, userID, projectID); err != nil {
			return err
		}
		board, err := ensureDefaultBoard(tx, companyID, projectID)
		if err != nil {
			return err
		}
		var maxPosition int
		tx.Model(&BoardColumn{}).Where("board_id = ? AND deleted_at IS NULL", board.ID).Select("COALESCE(MAX(position), 0)").Scan(&maxPosition)
		key := slugify(req.Title)
		if key == "" {
			key = fmt.Sprintf("column-%d", maxPosition+1)
		}
		key = strings.ReplaceAll(key, "-", "_")
		for i := 0; i < 100; i++ {
			var count int64
			tx.Model(&BoardColumn{}).Where("board_id = ? AND column_key = ?", board.ID, key).Count(&count)
			if count == 0 {
				break
			}
			key = fmt.Sprintf("%s_%d", key, i+2)
		}
		column := BoardColumn{
			ID: uuid.New(), CompanyID: companyID, BoardID: board.ID, ColumnKey: key,
			Title: strings.TrimSpace(req.Title), Color: strings.TrimSpace(req.Color), Position: maxPosition + 1,
		}
		if column.Title == "" {
			return errors.New("column title is required")
		}
		if column.Color == "" {
			column.Color = "gray"
		}
		if err := tx.Create(&column).Error; err != nil {
			return err
		}
		out = columnResponse{ID: column.ID.String(), Key: column.ColumnKey, Title: column.Title, Color: column.Color, Tasks: []taskResponse{}}
		return nil
	})
	return out, err
}

func (s *Service) deleteColumn(companyID, userID, projectID, columnID uuid.UUID) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAccess(tx, companyID, userID, projectID); err != nil {
			return err
		}
		board, err := ensureDefaultBoard(tx, companyID, projectID)
		if err != nil {
			return err
		}
		var column BoardColumn
		if err := tx.Where("id = ? AND board_id = ? AND deleted_at IS NULL", columnID, board.ID).First(&column).Error; err != nil {
			return err
		}
		if column.ColumnKey == "todo" {
			return errors.New("to do column cannot be deleted")
		}
		todo, err := findColumnByKey(tx, board.ID, "todo")
		if err != nil {
			return err
		}
		if err := tx.Model(&Task{}).Where("column_id = ? AND company_id = ? AND deleted_at IS NULL", columnID, companyID).Updates(map[string]any{
			"column_id":  todo.ID,
			"status_key": todo.ColumnKey,
		}).Error; err != nil {
			return err
		}
		return tx.Delete(&column).Error
	})
}

func (s *Service) getOrCreatePublicPage(companyID, userID, projectID uuid.UUID) (ProjectPublicPage, error) {
	var page ProjectPublicPage
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		project, err := s.ensureProjectAccess(tx, companyID, userID, projectID)
		if err != nil {
			return err
		}
		if err := tx.Where("project_id = ? AND company_id = ? AND deleted_at IS NULL", projectID, companyID).First(&page).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		slug := s.nextPublicPageSlug(tx, project.Name, project.ProjectKey)
		page = ProjectPublicPage{
			ID:              uuid.New(),
			CompanyID:       companyID,
			ProjectID:       projectID,
			Slug:            slug,
			Title:           project.Name,
			HTMLTemplate:    "",
			AccessMode:      "public",
			IsEnabled:       false,
			CreatedByUserID: &userID,
			UpdatedByUserID: &userID,
		}
		return tx.Create(&page).Error
	})
	return page, err
}

func (s *Service) updatePublicPage(companyID, userID, projectID uuid.UUID, req publicPageUpdateRequest) (ProjectPublicPage, error) {
	page, err := s.getOrCreatePublicPage(companyID, userID, projectID)
	if err != nil {
		return ProjectPublicPage{}, err
	}
	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.ensureProjectAdmin(tx, companyID, userID, projectID); err != nil {
			return err
		}
		updates := map[string]any{
			"is_enabled":         req.Enabled,
			"title":              strings.TrimSpace(req.Title),
			"html_template":      req.HTMLTemplate,
			"updated_by_user_id": userID,
		}
		if updates["title"] == "" {
			updates["title"] = page.Title
		}
		accessMode := normalizeAccessMode(req.AccessMode)
		if accessMode == "" {
			accessMode = page.AccessMode
			if accessMode == "" {
				accessMode = "public"
			}
		}
		updates["access_mode"] = accessMode
		if slug := normalizePublicSlug(req.Slug); slug != "" && slug != page.Slug {
			var count int64
			if err := tx.Model(&ProjectPublicPage{}).
				Where("slug = ? AND project_id <> ? AND deleted_at IS NULL", slug, projectID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("public slug already exists")
			}
			updates["slug"] = slug
		}
		if accessMode == "password_protected" {
			password := strings.TrimSpace(req.Password)
			if password == "" && page.PasswordHash == nil {
				return errors.New("password is required for password protected pages")
			}
			if password != "" {
				hash, err := security.HashPassword(password, 12)
				if err != nil {
					return err
				}
				updates["password_hash"] = hash
			}
		} else {
			updates["password_hash"] = nil
		}
		if err := tx.Model(&ProjectPublicPage{}).
			Where("id = ? AND deleted_at IS NULL", page.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND deleted_at IS NULL", page.ID).First(&page).Error
	})
	return page, err
}

func (s *Service) getPublicPageBySlug(slug, password string) (ProjectPublicPage, error) {
	slug = normalizePublicSlug(slug)
	if slug == "" {
		return ProjectPublicPage{}, errors.New("public project slug is required")
	}
	var page ProjectPublicPage
	if err := s.repo.db.Where("slug = ? AND is_enabled = ? AND deleted_at IS NULL", slug, true).First(&page).Error; err != nil {
		return ProjectPublicPage{}, err
	}
	if page.AccessMode == "password_protected" {
		if page.PasswordHash == nil || !security.CheckPassword(*page.PasswordHash, strings.TrimSpace(password)) {
			return ProjectPublicPage{}, errors.New("public page password required")
		}
	}
	return page, nil
}

func (s *Service) registerPublicProjectUser(slug string, req publicProjectRegistrationRequest) (memberSummary, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if name == "" {
		return memberSummary{}, errors.New("name is required")
	}
	if err := security.ValidatePassword(password); err != nil {
		return memberSummary{}, err
	}

	var created memberSummary
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		page, err := s.getPublicPageBySlug(slug, req.PortalPassword)
		if err != nil {
			return err
		}
		var project Project
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", page.ProjectID, page.CompanyID).First(&project).Error; err != nil {
			return err
		}
		if _, err := s.repo.findUserByEmail(page.CompanyID, email); err == nil {
			return errors.New("a user with this email already exists in this company; please log in instead")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		hash, err := security.HashPassword(password, 12)
		if err != nil {
			return err
		}
		clientRoleID, err := ensureClientRole(tx, page.CompanyID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		freeUntil := now.AddDate(0, 0, 30)
		user := users.User{
			ID:              uuid.New(),
			CompanyID:       page.CompanyID,
			Email:           email,
			Name:            name,
			PasswordHash:    hash,
			Status:          "active",
			AccountType:     users.AccountTypeFreeClient,
			FreeExpiresAt:   &freeUntil,
			EmailVerifiedAt: &now,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO user_roles (user_id, company_id, role_id, deleted_at)
			VALUES (?, ?, ?, NULL)
			ON CONFLICT (user_id, company_id, role_id)
			DO UPDATE SET deleted_at = NULL
		`, user.ID, page.CompanyID, clientRoleID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO team_members (team_id, user_id, added_by_user_id, created_at, updated_at, deleted_at)
			VALUES (?, ?, NULL, NOW(), NOW(), NULL)
			ON CONFLICT (team_id, user_id)
			DO UPDATE SET deleted_at = NULL, updated_at = NOW()
		`, project.TeamID, user.ID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO project_members (project_id, user_id, role, invited_by_user_id, created_at, updated_at, deleted_at)
			VALUES (?, ?, 'member', NULL, NOW(), NOW(), NULL)
			ON CONFLICT (project_id, user_id)
			DO UPDATE SET deleted_at = NULL, role = EXCLUDED.role, updated_at = NOW()
		`, project.ID, user.ID).Error; err != nil {
			return err
		}
		created = memberSummary{
			UserID: user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
		}
		return nil
	})
	return created, err
}

func ensureClientRole(tx *gorm.DB, companyID uuid.UUID) (uuid.UUID, error) {
	var role rolesmod.Role
	err := tx.Where("company_id = ? AND LOWER(name) = ? AND deleted_at IS NULL", companyID, "client").First(&role).Error
	if err == nil {
		return role.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}
	role = rolesmod.Role{
		ID:          uuid.New(),
		CompanyID:   &companyID,
		Name:        "Client",
		Description: "Client-facing access limited to project ticket intake",
		IsSystem:    true,
	}
	if err := tx.Create(&role).Error; err != nil {
		return uuid.Nil, err
	}
	if err := tx.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT ?, p.id
		FROM permissions p
		WHERE p.name IN ?
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, role.ID, []string{"project:view", "task:create", "task:view"}).Error; err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}

func (s *Service) nextPublicPageSlug(tx *gorm.DB, projectName, projectKey string) string {
	base := normalizePublicSlug(projectName)
	if base == "" {
		base = normalizePublicSlug(projectKey)
	}
	if base == "" {
		base = "project"
	}
	candidate := base
	for i := 0; i < 100; i++ {
		var count int64
		tx.Model(&ProjectPublicPage{}).Where("slug = ? AND deleted_at IS NULL", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix())
}

func normalizeAccessMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "public":
		return "public"
	case "password_protected", "password-protected", "password":
		return "password_protected"
	default:
		return ""
	}
}

func normalizePublicSlug(value string) string {
	value = slugify(strings.TrimSpace(value))
	return strings.Trim(value, "-")
}

func (s *Service) buildPublicPageResponse(page ProjectPublicPage, includeHTML bool) (publicPageResponse, error) {
	type pageMeta struct {
		ProjectName string `gorm:"column:project_name"`
		CompanySlug string `gorm:"column:company_slug"`
	}
	var meta pageMeta
	if err := s.repo.db.Table("project_public_pages ppp").
		Select("p.name AS project_name, c.slug AS company_slug").
		Joins("JOIN projects p ON p.id = ppp.project_id").
		Joins("JOIN companies c ON c.id = ppp.company_id").
		Where("ppp.id = ?", page.ID).
		Scan(&meta).Error; err != nil {
		return publicPageResponse{}, err
	}
	resp := publicPageResponse{
		ProjectID:            page.ProjectID.String(),
		ProjectName:          meta.ProjectName,
		CompanySlug:          meta.CompanySlug,
		Slug:                 page.Slug,
		Title:                page.Title,
		AccessMode:           page.AccessMode,
		IsEnabled:            page.IsEnabled,
		RequiresPassword:     page.AccessMode == "password_protected",
		LoginEndpoint:        "/api/v1/auth/login",
		RegisterEndpoint:     fmt.Sprintf("/api/v1/public/projects/%s/register", page.Slug),
		TicketCreateEndpoint: fmt.Sprintf("/api/v1/projects/%s/tasks", page.ProjectID.String()),
	}
	if includeHTML {
		resp.HTMLTemplate = page.HTMLTemplate
	}
	return resp, nil
}
