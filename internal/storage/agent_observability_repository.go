package storage

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

type ObservabilityRun struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	TenantID          uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	AtlasTenantID     string     `json:"atlas_tenant_id" gorm:"not null"`
	AtlasRunID        string     `json:"atlas_run_id" gorm:"not null"`
	AgentID           string     `json:"agent_id" gorm:"not null"`
	AgentType         string     `json:"agent_type" gorm:"not null"`
	SpanID            *string    `json:"span_id,omitempty"`
	ParentSpanID      *string    `json:"parent_span_id,omitempty"`
	Status            string     `json:"status" gorm:"default:'running'"`
	TotalCostUSD      float64    `json:"total_cost_usd" gorm:"default:0"`
	TotalInputTokens  int        `json:"total_input_tokens" gorm:"default:0"`
	TotalOutputTokens int        `json:"total_output_tokens" gorm:"default:0"`
	EventCount        int        `json:"event_count" gorm:"default:0"`
	ErrorCount        int        `json:"error_count" gorm:"default:0"`
	ToolCallCount     int        `json:"tool_call_count" gorm:"default:0"`
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
	LastEventAt       *time.Time `json:"last_event_at,omitempty"`
	Metadata          JSONMap    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (ObservabilityRun) TableName() string {
	return "agent_observability_runs"
}

type ObservabilityConfig struct {
	TenantID          uuid.UUID `json:"tenant_id" gorm:"type:uuid;primaryKey"`
	AtlasTenantID     string    `json:"atlas_tenant_id"`
	SamplingRate      float64   `json:"sampling_rate" gorm:"default:1.0"`
	TraceErrorsOnly   bool      `json:"trace_errors_only" gorm:"default:false"`
	SampleHeadPercent float64   `json:"sample_head_percent" gorm:"default:100"`
	SampleTailCount   int       `json:"sample_tail_count" gorm:"default:10"`
	RetentionDays     int       `json:"retention_days" gorm:"default:90"`
	IsActive          bool      `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ObservabilityConfig) TableName() string {
	return "agent_observability_config"
}

type AgentObservabilityRepository struct {
	db *gorm.DB
}

func NewAgentObservabilityRepository(db *gorm.DB) *AgentObservabilityRepository {
	return &AgentObservabilityRepository{db: db}
}

func (r *AgentObservabilityRepository) CreateRun(ctx context.Context, run *ObservabilityRun) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	run.CreatedAt = time.Now()
	run.StartedAt = time.Now()
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *AgentObservabilityRepository) GetRun(ctx context.Context, tenantID, id uuid.UUID) (*ObservabilityRun, error) {
	var run ObservabilityRun
	if err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *AgentObservabilityRepository) GetRunByAtlasID(ctx context.Context, atlasRunID string) (*ObservabilityRun, error) {
	var run ObservabilityRun
	if err := r.db.WithContext(ctx).
		Where("atlas_run_id = ?", atlasRunID).
		First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *AgentObservabilityRepository) ListRuns(ctx context.Context, tenantID uuid.UUID, agentID, status string, limit, offset int) ([]*ObservabilityRun, int64, error) {
	var runs []*ObservabilityRun
	var total int64

	query := r.db.WithContext(ctx).Model(&ObservabilityRun{}).Where("tenant_id = ?", tenantID)

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("started_at DESC").Limit(limit).Offset(offset).Find(&runs).Error; err != nil {
		return nil, 0, err
	}

	return runs, total, nil
}

func (r *AgentObservabilityRepository) UpdateRunStats(ctx context.Context, id uuid.UUID, stats *RunStats) error {
	updates := map[string]interface{}{
		"total_cost_usd":      stats.TotalCostUSD,
		"total_input_tokens":  stats.InputTokens,
		"total_output_tokens": stats.OutputTokens,
		"event_count":         stats.EventCount,
		"error_count":         stats.ErrorCount,
		"tool_call_count":     stats.ToolCallCount,
	}

	return r.db.WithContext(ctx).Model(&ObservabilityRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AgentObservabilityRepository) UpdateRunEventCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&ObservabilityRun{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"event_count":   gorm.Expr("event_count + 1"),
			"last_event_at": time.Now(),
		}).Error
}

func (r *AgentObservabilityRepository) EndRun(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":   status,
		"ended_at": &now,
	}

	return r.db.WithContext(ctx).Model(&ObservabilityRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AgentObservabilityRepository) GetConfig(ctx context.Context, tenantID uuid.UUID) (*ObservabilityConfig, error) {
	var config ObservabilityConfig
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.createDefaultConfig(ctx, tenantID)
		}
		return nil, err
	}
	return &config, nil
}

func (r *AgentObservabilityRepository) createDefaultConfig(ctx context.Context, tenantID uuid.UUID) (*ObservabilityConfig, error) {
	atlasTenantID := DeriveAtlasTenantID(tenantID)
	config := &ObservabilityConfig{
		TenantID:          tenantID,
		AtlasTenantID:     atlasTenantID,
		SamplingRate:      1.0,
		TraceErrorsOnly:   false,
		SampleHeadPercent: 100,
		SampleTailCount:   10,
		RetentionDays:     90,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return nil, err
	}

	return config, nil
}

func (r *AgentObservabilityRepository) UpsertConfig(ctx context.Context, config *ObservabilityConfig) error {
	config.UpdatedAt = time.Now()

	existing, err := r.GetConfig(ctx, config.TenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		return r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"sampling_rate", "trace_errors_only", "sample_head_percent", "sample_tail_count", "retention_days", "is_active", "updated_at"}),
		}).Create(config).Error
	}

	return r.db.WithContext(ctx).Create(config).Error
}

func (r *AgentObservabilityRepository) ShouldTrace(ctx context.Context, tenantID uuid.UUID, isError bool) (bool, error) {
	config, err := r.GetConfig(ctx, tenantID)
	if err != nil {
		return false, err
	}

	if !config.IsActive {
		return false, nil
	}

	if isError && config.TraceErrorsOnly {
		return true, nil
	}

	if isError && !config.TraceErrorsOnly {
		return true, nil
	}

	if config.SamplingRate >= 1.0 {
		return true, nil
	}

	if config.SamplingRate <= 0.0 {
		return false, nil
	}

	return rand.Float64() < config.SamplingRate, nil
}

func (r *AgentObservabilityRepository) DeleteOldRuns(ctx context.Context, tenantID uuid.UUID, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND ended_at IS NOT NULL AND ended_at < ?", tenantID, cutoff).
		Delete(&ObservabilityRun{})

	return result.RowsAffected, result.Error
}

func (r *AgentObservabilityRepository) ListTenantsForRetention(ctx context.Context) ([]uuid.UUID, error) {
	var tenantIDs []uuid.UUID

	err := r.db.WithContext(ctx).
		Model(&ObservabilityConfig{}).
		Where("is_active = ?", true).
		Pluck("tenant_id", &tenantIDs).Error

	if err != nil {
		return nil, err
	}

	return tenantIDs, nil
}

type RunStats struct {
	AtlasRunID    string  `json:"atlas_run_id"`
	DurationMs    int64   `json:"duration_ms"`
	EventCount    int     `json:"event_count"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	ErrorCount    int     `json:"error_count"`
	ToolCallCount int     `json:"tool_call_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	CostPerToken  float64 `json:"cost_per_token"`
}
