package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents valid task statuses
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in-progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusReview     TaskStatus = "review"
)

// TaskPriority represents valid task priorities
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// StudioTask represents a task in the studio task management system
type StudioTask struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      TaskStatus        `json:"status"`
	Priority    TaskPriority      `json:"priority"`
	AssigneeID  *string           `json:"assignee_id,omitempty"`
	CreatedBy   string            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// Validate validates the task data
func (t *StudioTask) Validate() error {
	if t.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(t.Title) > 500 {
		return fmt.Errorf("title must be 500 characters or less")
	}
	if t.Description != "" && len(t.Description) > 10000 {
		return fmt.Errorf("description must be 10000 characters or less")
	}
	if !isValidTaskStatus(t.Status) {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	if !isValidTaskPriority(t.Priority) {
		return fmt.Errorf("invalid priority: %s", t.Priority)
	}
	return nil
}

func isValidTaskStatus(s TaskStatus) bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone, TaskStatusBlocked, TaskStatusReview:
		return true
	default:
		return false
	}
}

func isValidTaskPriority(p TaskPriority) bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh:
		return true
	default:
		return false
	}
}

// TaskRepository handles database operations for studio tasks
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// ListTasksParams contains parameters for listing tasks
type ListTasksParams struct {
	TenantID   string
	Status     *TaskStatus
	AssigneeID *string
	Limit      int
	Offset     int
}

// ListTasks returns tasks filtered by tenant and optional filters
func (r *TaskRepository) ListTasks(ctx context.Context, params ListTasksParams) ([]StudioTask, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, params.TenantID)
	argIdx++

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*params.Status))
		argIdx++
	}

	if params.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *params.AssigneeID)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, title, description, status, priority, assignee_id,
		       created_by, created_at, updated_at, completed_at, metadata
		FROM studio_tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []StudioTask
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

// GetTask returns a single task by ID
func (r *TaskRepository) GetTask(ctx context.Context, tenantID, taskID string) (*StudioTask, error) {
	query := `
		SELECT id, tenant_id, title, description, status, priority, assignee_id,
		       created_by, created_at, updated_at, completed_at, metadata
		FROM studio_tasks
		WHERE tenant_id = $1 AND id = $2
	`
	var task StudioTask
	var desc sql.NullString
	var assigneeID sql.NullString
	var completedAt sql.NullTime
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, taskID).Scan(
		&task.ID, &task.TenantID, &task.Title, &desc,
		&task.Status, &task.Priority, &assigneeID,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt, &completedAt, &metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	if desc.Valid {
		task.Description = desc.String
	}
	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &task.Metadata)
	}

	return &task, nil
}

// CreateTask creates a new task
func (r *TaskRepository) CreateTask(ctx context.Context, task *StudioTask) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	metaRaw, _ := json.Marshal(task.Metadata)
	now := time.Now()

	query := `
		INSERT INTO studio_tasks (id, tenant_id, title, description, status, priority, assignee_id, created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		task.ID, task.TenantID, task.Title, task.Description,
		task.Status, task.Priority, task.AssigneeID,
		task.CreatedBy, now, now, metaRaw,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

// UpdateTask updates an existing task
func (r *TaskRepository) UpdateTask(ctx context.Context, tenantID, taskID string, updates map[string]interface{}) (*StudioTask, error) {
	// Build dynamic update query
	var sets []string
	var args []interface{}
	argIdx := 1

	// Always update updated_at
	sets = append(sets, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	// Handle status update
	if status, ok := updates["status"]; ok {
		s := TaskStatus(status.(string))
		if !isValidTaskStatus(s) {
			return nil, fmt.Errorf("invalid status: %s", s)
		}
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(s))
		argIdx++

		// Set completed_at if status is done
		if s == TaskStatusDone {
			now := time.Now()
			sets = append(sets, fmt.Sprintf("completed_at = $%d", argIdx))
			args = append(args, now)
			argIdx++
		}
	}

	// Handle title update
	if title, ok := updates["title"]; ok {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, title)
		argIdx++
	}

	// Handle description update
	if desc, ok := updates["description"]; ok {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, desc)
		argIdx++
	}

	// Handle priority update
	if priority, ok := updates["priority"]; ok {
		p := TaskPriority(priority.(string))
		if !isValidTaskPriority(p) {
			return nil, fmt.Errorf("invalid priority: %s", p)
		}
		sets = append(sets, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, string(p))
		argIdx++
	}

	// Handle assignee update
	if assigneeID, ok := updates["assignee_id"]; ok {
		sets = append(sets, fmt.Sprintf("assignee_id = $%d", argIdx))
		if assigneeID == nil {
			args = append(args, nil)
		} else {
			args = append(args, assigneeID)
		}
		argIdx++
	}

	// Handle metadata update
	if metadata, ok := updates["metadata"]; ok {
		metaRaw, _ := json.Marshal(metadata)
		sets = append(sets, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, metaRaw)
		argIdx++
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	query := fmt.Sprintf(`
		UPDATE studio_tasks
		SET %s
		WHERE tenant_id = $%d AND id = $%d
		RETURNING id, tenant_id, title, description, status, priority, assignee_id, created_by, created_at, updated_at, completed_at, metadata
	`, strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, tenantID, taskID)

	var task StudioTask
	var desc sql.NullString
	var assigneeID sql.NullString
	var completedAt sql.NullTime
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&task.ID, &task.TenantID, &task.Title, &desc,
		&task.Status, &task.Priority, &assigneeID,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt, &completedAt, &metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	if desc.Valid {
		task.Description = desc.String
	}
	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &task.Metadata)
	}

	return &task, nil
}

// DeleteTask deletes a task
func (r *TaskRepository) DeleteTask(ctx context.Context, tenantID, taskID string) error {
	query := `DELETE FROM studio_tasks WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, taskID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// AssignTask assigns a task to a user
func (r *TaskRepository) AssignTask(ctx context.Context, tenantID, taskID, assigneeID string) error {
	query := `
		UPDATE studio_tasks
		SET assignee_id = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	result, err := r.db.ExecContext(ctx, query, assigneeID, tenantID, taskID)
	if err != nil {
		return fmt.Errorf("assign task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func scanTask(rows interface{ Scan(dst ...interface{}) error }) (*StudioTask, error) {
	var task StudioTask
	var desc sql.NullString
	var assigneeID sql.NullString
	var completedAt sql.NullTime
	var metadata []byte

	err := rows.Scan(
		&task.ID, &task.TenantID, &task.Title, &desc,
		&task.Status, &task.Priority, &assigneeID,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt, &completedAt, &metadata,
	)
	if err != nil {
		return nil, err
	}

	if desc.Valid {
		task.Description = desc.String
	}
	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &task.Metadata)
	}

	return &task, nil
}