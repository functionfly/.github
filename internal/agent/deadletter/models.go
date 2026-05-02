package deadletter

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInspected  Status = "inspected"
	StatusRetried    Status = "retried"
	StatusDiscarded  Status = "discarded"
)

type AgentDeadLetter struct {
	ID                uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID           string                 `json:"agent_id" gorm:"not null;index"`
	TenantID          uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null;index"`
	FunctionID        uuid.UUID              `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionURI       string                 `json:"function_uri" gorm:"size:500"`
	ExecutionID       string                 `json:"execution_id" gorm:"size:100;index"`
	SessionID         string                 `json:"session_id" gorm:"size:100;index"`
	InputPayload      JSONMap                `json:"input_payload" gorm:"type:jsonb"`
	OutputPayload     JSONMap                `json:"output_payload,omitempty" gorm:"type:jsonb"`
	FinalError        string                 `json:"final_error" gorm:"size:2000"`
	ErrorCode         string                 `json:"error_code" gorm:"size:100"`
	Attempts          int                    `json:"attempts" gorm:"not null;default:0"`
	FirstAttemptAt    *time.Time             `json:"first_attempt_at,omitempty"`
	LastAttemptAt     *time.Time             `json:"last_attempt_at,omitempty"`
	FirstAttemptError string                 `json:"first_attempt_error" gorm:"size:2000"`
	Status            Status                 `json:"status" gorm:"size:50;not null;default:'pending';index"`
	CanRetry          bool                   `json:"can_retry" gorm:"not null;default:true"`
	RetryCount        int                    `json:"retry_count" gorm:"not null;default:0"`
	RetriedAt         *time.Time            `json:"retried_at,omitempty"`
	RetrySuccess      bool                   `json:"retry_success" gorm:"default:false"`
	RetryError        string                 `json:"retry_error,omitempty" gorm:"size:2000"`
	AlertSent         bool                   `json:"alert_sent" gorm:"default:false"`
	AlertSentAt       *time.Time            `json:"alert_sent_at,omitempty"`
	AlertThreshold    int                    `json:"alert_threshold" gorm:"default:3"`
	Metadata          JSONMap                `json:"metadata,omitempty" gorm:"type:jsonb"`
	Trace             string                 `json:"trace,omitempty" gorm:"size:5000"`
	CreatedAt         time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt         time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AgentDeadLetter) TableName() string {
	return "agent_dead_letter"
}

type JSONMap map[string]any

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for JSONB")
	}
	if len(b) == 0 {
		*m = make(JSONMap)
		return nil
	}
	return json.Unmarshal(b, m)
}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

type Repository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewRepository(db *gorm.DB, logger *logrus.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) Create(ctx context.Context, dl *AgentDeadLetter) error {
	if dl.ID == uuid.Nil {
		dl.ID = uuid.New()
	}
	if dl.CreatedAt.IsZero() {
		dl.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(dl).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*AgentDeadLetter, error) {
	var dl AgentDeadLetter
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dl).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dead letter entry not found: %s", id)
		}
		return nil, err
	}
	return &dl, nil
}

func (r *Repository) GetByExecutionID(ctx context.Context, executionID string) (*AgentDeadLetter, error) {
	var dl AgentDeadLetter
	err := r.db.WithContext(ctx).Where("execution_id = ?", executionID).First(&dl).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dl, nil
}

func (r *Repository) List(ctx context.Context, agentID string, status Status, limit, offset int) ([]*AgentDeadLetter, int64, error) {
	var entries []*AgentDeadLetter
	var total int64

	query := r.db.WithContext(ctx).Model(&AgentDeadLetter{})
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *Repository) ListPending(ctx context.Context, limit int) ([]*AgentDeadLetter, error) {
	var entries []*AgentDeadLetter
	err := r.db.WithContext(ctx).
		Where("status = ? AND can_retry = ?", StatusPending, true).
		Order("created_at ASC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

func (r *Repository) Update(ctx context.Context, dl *AgentDeadLetter) error {
	dl.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(dl).Error
}

func (r *Repository) MarkRetried(ctx context.Context, id uuid.UUID, success bool, retryError string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        StatusRetried,
		"retried_at":    now,
		"retry_success": success,
	}
	if retryError != "" {
		updates["retry_error"] = retryError
	}
	if success {
		updates["can_retry"] = false
	}
	return r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) MarkInspected(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("id = ?", id).Update("status", StatusInspected).Error
}

func (r *Repository) MarkDiscarded(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   StatusDiscarded,
		"can_retry": false,
	}).Error
}

func (r *Repository) DeleteOlderThan(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := r.db.WithContext(ctx).Where("created_at < ? AND status IN (?, ?)", cutoff, StatusRetried, StatusDiscarded).Delete(&AgentDeadLetter{})
	return result.RowsAffected, result.Error
}

func (r *Repository) CountByStatus(ctx context.Context, status Status) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *Repository) CountPending(ctx context.Context) (int64, error) {
	return r.CountByStatus(ctx, StatusPending)
}

func (r *Repository) GetStats(ctx context.Context, agentID string) (*Stats, error) {
	var stats Stats
	stats.AgentID = agentID

	query := r.db.WithContext(ctx).Model(&AgentDeadLetter{})
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	oneDayAgo := now.Add(-24 * time.Hour)

	err := query.Select(`
		COUNT(*) as total,
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		SUM(CASE WHEN status = 'retried' AND retry_success = true THEN 1 ELSE 0 END) as retry_success,
		SUM(CASE WHEN status = 'retried' AND retry_success = false THEN 1 ELSE 0 END) as retry_failed,
		SUM(CASE WHEN created_at >= ? THEN 1 ELSE 0 END) as last_hour,
		SUM(CASE WHEN created_at >= ? THEN 1 ELSE 0 END) as last_24h
	`, oneHourAgo, oneDayAgo).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

type Stats struct {
	AgentID       string `json:"agent_id"`
	Total         int64  `json:"total"`
	Pending       int64  `json:"pending"`
	RetrySuccess  int64  `json:"retry_success"`
	RetryFailed   int64  `json:"retry_failed"`
	LastHour      int64  `json:"last_hour"`
	Last24h       int64  `json:"last_24h"`
}

func (r *Repository) GetPendingAlerts(ctx context.Context, threshold int) ([]*AgentDeadLetter, error) {
	var entries []*AgentDeadLetter
	err := r.db.WithContext(ctx).
		Where("alert_sent = ? AND attempts >= ? AND status = ?", false, threshold, StatusPending).
		Find(&entries).Error
	return entries, err
}

func (r *Repository) MarkAlertSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("id = ?", id).Updates(map[string]interface{}{
		"alert_sent":    true,
		"alert_sent_at": now,
	}).Error
}