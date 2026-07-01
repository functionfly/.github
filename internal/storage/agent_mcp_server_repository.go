package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentMCPServerRepository handles database operations for agent MCP servers
type AgentMCPServerRepository struct {
	db *gorm.DB
}

// NewAgentMCPServerRepository creates a new agent MCP server repository
func NewAgentMCPServerRepository(db *gorm.DB) *AgentMCPServerRepository {
	return &AgentMCPServerRepository{db: db}
}

// Create creates a new MCP server for an agent
func (r *AgentMCPServerRepository) Create(ctx context.Context, s *AgentMCPServer) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetByID retrieves an MCP server by ID, scoped to tenant
func (r *AgentMCPServerRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*AgentMCPServer, error) {
	var s AgentMCPServer
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListByAgent lists all MCP servers for an agent, scoped to tenant
func (r *AgentMCPServerRepository) ListByAgent(ctx context.Context, agentID string, tenantID uuid.UUID) ([]*AgentMCPServer, error) {
	var servers []*AgentMCPServer
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND tenant_id = ?", agentID, tenantID).
		Order("created_at DESC").
		Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// Update updates specific fields of an MCP server
func (r *AgentMCPServerRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&AgentMCPServer{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete deletes an MCP server by ID
func (r *AgentMCPServerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&AgentMCPServer{}, "id = ?", id).Error
}

// UpdateConnectionStatus updates last_connected_at and clears last_error on successful connection
func (r *AgentMCPServerRepository) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, connected bool, errMsg string) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if connected {
		now := time.Now()
		updates["last_connected_at"] = now
		updates["last_error"] = ""
	} else {
		updates["last_error"] = errMsg
	}
	return r.db.WithContext(ctx).
		Model(&AgentMCPServer{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateToolCount updates the tool count for an MCP server
func (r *AgentMCPServerRepository) UpdateToolCount(ctx context.Context, id uuid.UUID, count int) error {
	return r.db.WithContext(ctx).
		Model(&AgentMCPServer{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tool_count": count,
			"updated_at": time.Now(),
		}).Error
}

// GetByAgentAndURL finds a server by agent+URL (for uniqueness checks)
func (r *AgentMCPServerRepository) GetByAgentAndURL(ctx context.Context, agentID, url string, tenantID uuid.UUID) (*AgentMCPServer, error) {
	var s AgentMCPServer
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND url = ? AND tenant_id = ?", agentID, url, tenantID).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
