package brain

import (
	"context"

	"github.com/google/uuid"
)

type BrainScope string

const (
	ScopeTenant BrainScope = "tenant"
	ScopeAgent  BrainScope = "agent"
)

type AgentIdentityExt struct {
	BrainEnabled   bool      `json:"brain_enabled"`
	BrainOwnerType string    `json:"brain_owner_type"`
	BrainOwnerID   uuid.UUID `json:"brain_owner_id"`
}

type ScopeResolver struct {
	agentBrains map[uuid.UUID]*AgentIdentityExt
}

func NewScopeResolver() *ScopeResolver {
	return &ScopeResolver{
		agentBrains: make(map[uuid.UUID]*AgentIdentityExt),
	}
}

// GetBrainScope determines whether an agent uses its own brain or the tenant brain
func (sr *ScopeResolver) GetBrainScope(ctx context.Context, agentID, tenantID uuid.UUID) (BrainScope, uuid.UUID) {
	ext, ok := sr.agentBrains[agentID]
	if ok && ext.BrainEnabled && ext.BrainOwnerType == "agent" {
		return ScopeAgent, agentID
	}
	return ScopeTenant, tenantID
}

// RegisterAgentBrain registers an agent as having its own isolated brain
func (sr *ScopeResolver) RegisterAgentBrain(agentID, tenantID uuid.UUID) {
	sr.agentBrains[agentID] = &AgentIdentityExt{
		BrainEnabled:   true,
		BrainOwnerType: "agent",
		BrainOwnerID:   agentID,
	}
}

// GetRedisKey returns the correct Redis key for the given brain scope
func GetRedisKey(scope BrainScope, id uuid.UUID) string {
	if scope == ScopeAgent {
		return "brain:agent:" + id.String() + ":signals"
	}
	return "brain:tenant:" + id.String() + ":signals"
}
