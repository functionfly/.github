package swarm

import (
	"context"
	"testing"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetDelegationDepth(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ctx := context.Background()

	// Create parent agents first to satisfy FK
	parentAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "parent-" + uuid.New().String()[:8],
		Name:      "Parent Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleManager,
	}
	require.NoError(t, db.Create(parentAgent).Error)

	childAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "child-" + uuid.New().String()[:8],
		Name:      "Child Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(childAgent).Error)

	// Create relationship
	rel := identity.AgentRelationship{
		ID:             uuid.New(),
		ParentAgentID:  parentAgent.AgentID,
		ChildAgentID:   childAgent.AgentID,
		RelationshipType: "parent",
	}
	require.NoError(t, db.Create(&rel).Error)

	identityRepo := identity.NewRepository(db)
	svc := NewService(db, identityRepo, nil)

	depth, err := svc.GetDelegationDepth(ctx, childAgent.AgentID)
	require.NoError(t, err)
	assert.Equal(t, 1, depth) // child -> parent
}

func TestService_GetDelegationDepth_NoChain(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	identityRepo := identity.NewRepository(db)
	svc := NewService(db, identityRepo, nil)
	ctx := context.Background()

	depth, err := svc.GetDelegationDepth(ctx, "orphan-agent")
	require.NoError(t, err)
	assert.Equal(t, 0, depth) // No parent chain
}
