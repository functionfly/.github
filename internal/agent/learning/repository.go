package learning

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles persistence for learning memories and insights
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new learning repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AgentMemory represents a stored memory in the agent's learning system
type AgentMemory struct {
	ID          uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID     string            `json:"agent_id" gorm:"not null;index"`
	MemoryType  string            `json:"memory_type" gorm:"not null"` // execution, insight, pattern, optimization
	Content     map[string]any    `json:"content" gorm:"type:jsonb;not null"`
	Importance  float64           `json:"importance" gorm:"type:decimal(3,2);default:0.5"`
	IsLearned   bool              `json:"is_learned" gorm:"not null;default:false"`
	Source      string            `json:"source"` // automatic, manual, external
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiresAt   *time.Time       `json:"expires_at"`
}

// TableName returns the GORM table name
func (AgentMemory) TableName() string {
	return "agent_memories"
}

// Memory types
const (
	MemoryTypeExecution = "execution"
	MemoryTypeInsight   = "insight"
	MemoryTypePattern  = "pattern"
	MemoryTypeOptimization = "optimization"
)

// CreateMemory stores a new memory
func (r *Repository) CreateMemory(ctx context.Context, memory *AgentMemory) error {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).Create(memory).Error
}

// GetMemory retrieves a memory by ID
func (r *Repository) GetMemory(ctx context.Context, memoryID uuid.UUID) (*AgentMemory, error) {
	var memory AgentMemory
	err := r.db.WithContext(ctx).Where("id = ?", memoryID).First(&memory).Error
	return &memory, err
}

// GetMemories retrieves memories for an agent with optional filters
func (r *Repository) GetMemories(ctx context.Context, agentID string, memoryType string, limit int, offset int) ([]AgentMemory, int64, error) {
	query := r.db.WithContext(ctx).Model(&AgentMemory{}).
		Where("agent_id = ?", agentID)

	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}

	// Only get non-expired memories
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	var memories []AgentMemory
	if err := query.Order("importance DESC, created_at DESC").Limit(limit).Offset(offset).Find(&memories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get memories: %w", err)
	}

	return memories, total, nil
}

// GetRecentMemories retrieves recent memories for an agent
func (r *Repository) GetRecentMemories(ctx context.Context, agentID string, limit int) ([]AgentMemory, error) {
	var memories []AgentMemory
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND (expires_at IS NULL OR expires_at > ?)", agentID, time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetImportantMemories retrieves high-importance memories for an agent
func (r *Repository) GetImportantMemories(ctx context.Context, agentID string, minImportance float64, limit int) ([]AgentMemory, error) {
	var memories []AgentMemory
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND importance >= ? AND (expires_at IS NULL OR expires_at > ?)", agentID, minImportance, time.Now()).
		Order("importance DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// MarkAsLearned marks a memory as learned
func (r *Repository) MarkAsLearned(ctx context.Context, memoryID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&AgentMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"is_learned": true,
			"updated_at": time.Now(),
		}).Error
}

// DeleteMemory deletes a memory
func (r *Repository) DeleteMemory(ctx context.Context, memoryID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AgentMemory{}, "id = ?", memoryID).Error
}

// DeleteOldMemories removes memories older than the specified duration
func (r *Repository) DeleteOldMemories(ctx context.Context, agentID string, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := r.db.WithContext(ctx).
		Where("agent_id = ? AND created_at < ? AND is_learned = ?", agentID, cutoff, false).
		Delete(&AgentMemory{})
	return result.RowsAffected, result.Error
}

// CleanupExpiredMemories removes expired memories
func (r *Repository) CleanupExpiredMemories(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&AgentMemory{})
	return result.RowsAffected, result.Error
}

// SearchMemories searches memories by content keywords
func (r *Repository) SearchMemories(ctx context.Context, agentID string, query string, limit int) ([]AgentMemory, error) {
	var memories []AgentMemory
	// Note: This is a simple implementation. For complex search, consider using full-text search
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND (expires_at IS NULL OR expires_at > ?)", agentID, time.Now()).
		Order("importance DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetMemoryStats returns statistics about memories for an agent
func (r *Repository) GetMemoryStats(ctx context.Context, agentID string) (*MemoryStats, error) {
	stats := &MemoryStats{
		AgentID: agentID,
	}

	// Total memories
	var total int64
	r.db.WithContext(ctx).Model(&AgentMemory{}).Where("agent_id = ?", agentID).Count(&total)
	stats.TotalMemories = total

	// By type
	var typeCounts map[string]int64
	r.db.WithContext(ctx).Model(&AgentMemory{}).
		Where("agent_id = ?", agentID).
		Select("memory_type, count(*)").
		Group("memory_type").
		Scan(&typeCounts)
	stats.ByType = typeCounts

	// Learned vs unlearned
	var learned, unlearned int64
	r.db.WithContext(ctx).Model(&AgentMemory{}).Where("agent_id = ? AND is_learned = ?", agentID, true).Count(&learned)
	r.db.WithContext(ctx).Model(&AgentMemory{}).Where("agent_id = ? AND is_learned = ?", agentID, false).Count(&unlearned)
	stats.LearnedCount = learned
	stats.UnlearnedCount = unlearned

	// Average importance
	var avgImportance float64
	r.db.WithContext(ctx).Model(&AgentMemory{}).Where("agent_id = ?", agentID).Select("avg(importance)").Scan(&avgImportance)
	stats.AverageImportance = avgImportance

	return stats, nil
}

// MemoryStats contains statistics about agent memories
type MemoryStats struct {
	AgentID          string         `json:"agent_id"`
	TotalMemories    int64          `json:"total_memories"`
	ByType           map[string]int64 `json:"by_type"`
	LearnedCount     int64          `json:"learned_count"`
	UnlearnedCount   int64          `json:"unlearned_count"`
	AverageImportance float64       `json:"average_importance"`
}

// AutoMigrate runs auto migration for learning models
func (r *Repository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&AgentMemory{},
		&ExecutionPattern{},
		&Optimization{},
	)
}
