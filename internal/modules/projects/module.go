package projects

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abrahamVado/go-golem.mx/internal/modules/users"
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/abrahamVado/go-golem.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid;not null"`
	ProjectKey      string    `gorm:"column:project_key"`
	Name            string
	Description     string
	Icon            string
	Status          string
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

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type projectCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
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

type projectSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type memberSummary struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
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

type taskHydrated struct {
	Task
	ColumnKey   string
	Assignees   []string
	Watchers    []string
	Checklist   []taskChecklistEntry
	TimeEntries []taskTimeEntryDTO
}

func RegisterRoutes(private *gin.RouterGroup, h *Handler) {
	private.GET("/teams", h.ListTeams)
	private.GET("/teams/:id/projects", h.ListProjectsForTeam)
	private.GET("/teams/:id/members", h.ListTeamMembers)
	private.GET("/projects", h.ListProjects)
	private.POST("/projects", h.CreateProject)
	private.GET("/projects/:id/board", h.GetBoard)
	private.POST("/projects/:id/tasks", h.CreateTask)
	private.POST("/projects/:id/board/columns", h.CreateColumn)
	private.DELETE("/projects/:id/board/columns/:columnId", h.DeleteColumn)
	private.GET("/tasks/:id", h.GetTask)
	private.PUT("/tasks/:id", h.UpdateTask)
	private.DELETE("/tasks/:id", h.DeleteTask)
	private.PATCH("/tasks/:id/move", h.MoveTask)
}

func (h *Handler) ListTeams(c *gin.Context) {
	companyID := tenancy.CompanyID(c)
	company, err := h.svc.repo.getCompany(companyID)
	if err != nil {
		response.Internal(c, "Failed to load company")
		return
	}
	response.OK(c, []gin.H{{
		"id":   company.ID.String(),
		"name": company.Name,
		"slug": company.Slug,
	}})
}

func (h *Handler) ListTeamMembers(c *gin.Context) {
	companyID := tenancy.CompanyID(c)
	if c.Param("id") != companyID.String() {
		response.Forbidden(c, "Team not available")
		return
	}
	members, err := h.svc.repo.listMembers(companyID)
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
	if c.Param("id") != companyID.String() {
		response.Forbidden(c, "Team not available")
		return
	}
	h.respondProjects(c, companyID)
}

func (h *Handler) ListProjects(c *gin.Context) {
	h.respondProjects(c, tenancy.CompanyID(c))
}

func (h *Handler) respondProjects(c *gin.Context, companyID uuid.UUID) {
	projects, err := h.svc.repo.listProjects(companyID)
	if err != nil {
		response.Internal(c, "Failed to load projects")
		return
	}
	out := make([]projectSummary, 0, len(projects))
	for _, project := range projects {
		out = append(out, projectSummary{
			ID:          project.ID.String(),
			Name:        project.Name,
			Description: project.Description,
			Icon:        project.Icon,
			CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		})
	}
	response.OK(c, out)
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req projectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid project payload")
		return
	}
	project, err := h.svc.createProject(tenancy.CompanyID(c), tenancy.UserID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, projectSummary{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		Icon:        project.Icon,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) GetBoard(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project id")
		return
	}
	board, err := h.svc.getBoard(tenancy.CompanyID(c), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.Internal(c, "Failed to load board")
		return
	}
	response.OK(c, board)
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
	task, err := h.svc.getTask(tenancy.CompanyID(c), taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Task not found")
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
	if err := h.svc.deleteTask(tenancy.CompanyID(c), taskID); err != nil {
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
	column, err := h.svc.createColumn(tenancy.CompanyID(c), projectID, req)
	if err != nil {
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
	if err := h.svc.deleteColumn(tenancy.CompanyID(c), projectID, columnID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

type Company struct {
	ID   uuid.UUID
	Name string
	Slug string
}

func (r *Repository) getCompany(companyID uuid.UUID) (Company, error) {
	var company Company
	err := r.db.Table("companies").Where("id = ? AND deleted_at IS NULL", companyID).First(&company).Error
	return company, err
}

func (r *Repository) listMembers(companyID uuid.UUID) ([]users.User, error) {
	var members []users.User
	err := r.db.Where("company_id = ? AND deleted_at IS NULL AND status = ?", companyID, "active").Order("name asc").Find(&members).Error
	return members, err
}

func (r *Repository) listProjects(companyID uuid.UUID) ([]Project, error) {
	var projects []Project
	err := r.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Order("created_at asc").Find(&projects).Error
	return projects, err
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	normalized = nonAlnum.ReplaceAllString(normalized, "-")
	return strings.Trim(normalized, "-")
}

func (s *Service) createProject(companyID, userID uuid.UUID, req projectCreateRequest) (Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Project{}, errors.New("project name is required")
	}
	keyBase := strings.ToUpper(strings.ReplaceAll(slugify(name), "-", ""))
	if keyBase == "" {
		keyBase = "PROJECT"
	}
	project := Project{
		ID:              uuid.New(),
		CompanyID:       companyID,
		ProjectKey:      keyBase,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		Icon:            strings.TrimSpace(req.Icon),
		Status:          "active",
		CreatedByUserID: &userID,
	}
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		project.ProjectKey = s.nextProjectKey(tx, companyID, project.ProjectKey)
		if err := tx.Create(&project).Error; err != nil {
			return err
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
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "backlog", Title: "Backlog", Color: "gray", Position: 1},
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "todo", Title: "To Do", Color: "blue", Position: 2},
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "in_progress", Title: "In Progress", Color: "amber", Position: 3},
		{ID: uuid.New(), CompanyID: companyID, BoardID: boardID, ColumnKey: "done", Title: "Done", Color: "green", Position: 4},
	}
	return tx.Create(&columns).Error
}

func (s *Service) getBoard(companyID, projectID uuid.UUID) (boardResponse, error) {
	var out boardResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var project Project
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", projectID, companyID).First(&project).Error; err != nil {
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

func (s *Service) getTask(companyID, taskID uuid.UUID) (taskResponse, error) {
	var out taskResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("id = ? AND company_id = ? AND deleted_at IS NULL", taskID, companyID).First(&task).Error; err != nil {
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

func (s *Service) deleteTask(companyID, taskID uuid.UUID) error {
	return s.repo.db.Where("id = ? AND company_id = ?", taskID, companyID).Delete(&Task{}).Error
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
		lookup = "backlog"
	}
	var column BoardColumn
	err := tx.Where("board_id = ? AND column_key = ? AND deleted_at IS NULL", boardID, lookup).First(&column).Error
	return column, err
}

func (s *Service) createColumn(companyID, projectID uuid.UUID, req columnCreateRequest) (columnResponse, error) {
	var out columnResponse
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
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

func (s *Service) deleteColumn(companyID, projectID, columnID uuid.UUID) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		board, err := ensureDefaultBoard(tx, companyID, projectID)
		if err != nil {
			return err
		}
		var column BoardColumn
		if err := tx.Where("id = ? AND board_id = ? AND deleted_at IS NULL", columnID, board.ID).First(&column).Error; err != nil {
			return err
		}
		if column.ColumnKey == "backlog" {
			return errors.New("backlog column cannot be deleted")
		}
		backlog, err := findColumnByKey(tx, board.ID, "backlog")
		if err != nil {
			return err
		}
		if err := tx.Model(&Task{}).Where("column_id = ? AND company_id = ? AND deleted_at IS NULL", columnID, companyID).Updates(map[string]any{
			"column_id":  backlog.ID,
			"status_key": backlog.ColumnKey,
		}).Error; err != nil {
			return err
		}
		return tx.Delete(&column).Error
	})
}
