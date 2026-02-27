package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// Agent Memory Operations
// ============================================

// StoreAgentMemory stores an agent memory with embedding
func (r *StateRepository) StoreAgentMemory(ctx context.Context, memory *AgentMemory) (*AgentMemory, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Create(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}

	return memory, nil
}

// GetAgentMemory retrieves agent memory by ID
func (r *StateRepository) GetAgentMemory(ctx context.Context, memoryID uuid.UUID) (*AgentMemory, error) {
	var memory AgentMemory
	err := r.db.WithContext(ctx).First(&memory, "id = ?", memoryID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("memory not found")
		}
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	return &memory, nil
}

// GetAgentMemories retrieves all memories for an agent
func (r *StateRepository) GetAgentMemories(ctx context.Context, tenantID uuid.UUID, agentID string, memoryType string, limit, offset int) ([]*AgentMemory, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&AgentMemory{}).Where("tenant_id = ? AND agent_id = ?", tenantID, agentID)

	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	var memories []*AgentMemory
	err = query.
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get memories: %w", err)
	}

	return memories, total, nil
}

// SearchAgentMemory searches agent memories by embedding similarity
func (r *StateRepository) SearchAgentMemory(ctx context.Context, tenantID uuid.UUID, agentID string, queryVector []float32, memoryType string, limit int) ([]*AgentMemory, error) {
	// Use GORM's raw SQL with proper vector handling for pgvector
	query := r.db.WithContext(ctx).Model(&AgentMemory{}).
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Where("expires_at IS NULL OR expires_at > NOW()")

	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}

	// For pgvector, we need raw SQL for the <=> operator
	// Convert float32 slice to PostgreSQL vector format
	vectorStr := "["
	for i, v := range queryVector {
		if i > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%.6f", v)
	}
	vectorStr += "]"

	var memories []*AgentMemory
	err := query.Raw(`
		SELECT id, tenant_id, agent_id, memory_type, content, structured_data,
			   embedding, importance_score, access_count, last_accessed_at,
			   ttl_days, expires_at, source_event_id, created_at, updated_at
		FROM agent_memories
		WHERE tenant_id = ? AND agent_id = ?
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND memory_type = COALESCE(?, memory_type)
		ORDER BY embedding <=> ?
		LIMIT ?
	`, tenantID, agentID, memoryType, vectorStr, limit).
		Scan(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	return memories, nil
}

// DeleteAgentMemory deletes an agent memory
func (r *StateRepository) DeleteAgentMemory(ctx context.Context, memoryID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&AgentMemory{}, "id = ?", memoryID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete memory: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

// UpdateAgentMemoryAccessCount increments the access count for a memory
func (r *StateRepository) UpdateAgentMemoryAccessCount(ctx context.Context, memoryID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&AgentMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"access_count":     gorm.Expr("access_count + 1"),
			"last_accessed_at": time.Now(),
			"updated_at":       time.Now(),
		}).Error
}