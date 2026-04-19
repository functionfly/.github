package graph

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionTrace represents a single node execution within a graph run,
// tagged with conversion/revenue events for the SEBG analyzer.
type ExecutionTrace struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID  uuid.UUID `json:"execution_id" gorm:"type:uuid;not null;index"`
	GraphID      uuid.UUID `json:"graph_id" gorm:"type:uuid;not null;index"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	NodeID       uuid.UUID `json:"node_id" gorm:"type:uuid;not null"`
	NodeName     string    `json:"node_name" gorm:"not null"`
	VerticalTag  string    `json:"vertical_tag" gorm:"type:text;index"`
	Input        string    `json:"input" gorm:"type:jsonb"`
	Output       string    `json:"output" gorm:"type:jsonb"`
	LatencyMs    int64     `json:"latency_ms"`
	Status       string    `json:"status" gorm:"not null"`
	EventType    string    `json:"event_type" gorm:"type:text;index"`
	RevenueCents int64     `json:"revenue_cents" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName returns the GORM table name.
func (ExecutionTrace) TableName() string { return "agent_execution_traces" }

// TraceRepository handles persistence of execution traces.
type TraceRepository struct {
	db *gorm.DB
}

// NewTraceRepository creates a new trace repository.
func NewTraceRepository(db *gorm.DB) *TraceRepository {
	return &TraceRepository{db: db}
}

// Create inserts a new execution trace.
func (r *TraceRepository) Create(ctx context.Context, trace *ExecutionTrace) error {
	if trace.ID == uuid.Nil {
		trace.ID = uuid.New()
	}
	if trace.CreatedAt.IsZero() {
		trace.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(trace).Error
}

// BatchCreate inserts multiple traces in a single transaction.
func (r *TraceRepository) BatchCreate(ctx context.Context, traces []ExecutionTrace) error {
	if len(traces) == 0 {
		return nil
	}
	for i := range traces {
		if traces[i].ID == uuid.Nil {
			traces[i].ID = uuid.New()
		}
		if traces[i].CreatedAt.IsZero() {
			traces[i].CreatedAt = time.Now().UTC()
		}
	}
	return r.db.WithContext(ctx).CreateInBatches(traces, 100).Error
}

// ListByTenant returns all traces for a tenant within a time window.
func (r *TraceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]ExecutionTrace, error) {
	var traces []ExecutionTrace
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Order("created_at DESC").
		Limit(limit).
		Find(&traces).Error
	return traces, err
}

// ListByGraph returns all traces for a specific graph.
func (r *TraceRepository) ListByGraph(ctx context.Context, graphID uuid.UUID, since time.Time, limit int) ([]ExecutionTrace, error) {
	var traces []ExecutionTrace
	err := r.db.WithContext(ctx).
		Where("graph_id = ? AND created_at >= ?", graphID, since).
		Order("created_at DESC").
		Limit(limit).
		Find(&traces).Error
	return traces, err
}

// ListByEventType returns traces filtered by event type for a tenant.
func (r *TraceRepository) ListByEventType(ctx context.Context, tenantID uuid.UUID, eventType string, since time.Time, limit int) ([]ExecutionTrace, error) {
	var traces []ExecutionTrace
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND event_type = ? AND created_at >= ?", tenantID, eventType, since).
		Order("created_at DESC").
		Limit(limit).
		Find(&traces).Error
	return traces, err
}

// GetRevenueSum returns total revenue for a tenant within a time window.
func (r *TraceRepository) GetRevenueSum(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.WithContext(ctx).
		Model(&ExecutionTrace{}).
		Select("COALESCE(SUM(revenue_cents), 0) as total").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Scan(&result).Error
	return result.Total, err
}

// AutoMigrate runs database migrations for trace components.
func (r *TraceRepository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&ExecutionTrace{})
}
