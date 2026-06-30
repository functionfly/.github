package swarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageService_SendMessage_ValidatesMessageType(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	svc := NewMessageServiceWithoutWorkers(db, nil)

	ctx := context.Background()

	// Create agent identities first to satisfy FK constraints
	fromAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "from-agent-" + uuid.New().String()[:8],
		Name:      "From Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	toAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "to-agent-" + uuid.New().String()[:8],
		Name:      "To Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(fromAgent).Error)
	require.NoError(t, db.Create(toAgent).Error)

	validTypes := []string{
		identity.MessageTypeTaskDelegation,
		identity.MessageTypeTaskResult,
		identity.MessageTypeQuery,
		identity.MessageTypeResponse,
		identity.MessageTypeCapabilityDiscovery,
		identity.MessageTypeHeartbeat,
		identity.MessageTypeEvolutionProposal,
		identity.MessageTypeBudgetRequest,
	}

	for _, msgType := range validTypes {
		msg := &identity.AgentMessage{
			ID:          uuid.New(),
			FromAgentID: fromAgent.AgentID,
			ToAgentID:   toAgent.AgentID,
			MessageType: msgType,
			Payload:     map[string]any{"test": "data"},
		}
		err := svc.SendMessage(ctx, msg, "signing-key")
		assert.NoError(t, err, "should accept valid message type: %s", msgType)
	}

	// Invalid message type
	invalidMsg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgent.AgentID,
		ToAgentID:   toAgent.AgentID,
		MessageType: "invalid_type",
		Payload:     map[string]any{"test": "data"},
	}
	err := svc.SendMessage(ctx, invalidMsg, "signing-key")
	assert.Error(t, err, "should reject invalid message type")
	assert.Contains(t, err.Error(), "invalid message type")
}

func TestMessageService_SendMessage_RequiresSigningKey(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	svc := NewMessageServiceWithoutWorkers(db, nil)

	ctx := context.Background()

	// Create agent identities
	fromAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "from-agent-" + uuid.New().String()[:8],
		Name:      "From Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	toAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "to-agent-" + uuid.New().String()[:8],
		Name:      "To Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(fromAgent).Error)
	require.NoError(t, db.Create(toAgent).Error)

	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgent.AgentID,
		ToAgentID:   toAgent.AgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Payload:     map[string]any{"test": "data"},
	}

	err := svc.SendMessage(ctx, msg, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing key is required")
}

func TestMessageService_SendMessage_SignsMessage(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	svc := NewMessageServiceWithoutWorkers(db, nil)

	ctx := context.Background()

	// Create agent identities
	fromAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "from-agent-" + uuid.New().String()[:8],
		Name:      "From Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	toAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "to-agent-" + uuid.New().String()[:8],
		Name:      "To Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(fromAgent).Error)
	require.NoError(t, db.Create(toAgent).Error)

	signingKey := "test-signing-key-123"
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgent.AgentID,
		ToAgentID:   toAgent.AgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Payload:     map[string]any{"task": "process", "data": "test"},
	}

	err := svc.SendMessage(ctx, msg, signingKey)
	require.NoError(t, err)

	assert.NotEmpty(t, msg.Signature, "message should be signed")
	assert.NotEmpty(t, msg.Nonce, "message should have nonce")

	// Verify signature is valid
	payloadBytes, _ := json.Marshal(msg.Payload)
	valid := svc.signingService.VerifySignature(
		msg.FromAgentID,
		msg.Signature,
		signingKey,
		payloadBytes,
		msg.Nonce,
		msg.SequenceNumber,
	)
	assert.True(t, valid, "signature should verify")
}

func TestMessageService_MarkDelivered(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	svc := NewMessageServiceWithoutWorkers(db, nil)

	ctx := context.Background()

	// Create agent identities
	fromAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "from-agent-" + uuid.New().String()[:8],
		Name:      "From Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	toAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "to-agent-" + uuid.New().String()[:8],
		Name:      "To Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(fromAgent).Error)
	require.NoError(t, db.Create(toAgent).Error)

	msgID := uuid.New()
	msg := identity.AgentMessage{
		ID:          msgID,
		FromAgentID: fromAgent.AgentID,
		ToAgentID:   toAgent.AgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Status:      "pending",
	}
	require.NoError(t, db.Create(&msg).Error)

	err := svc.MarkDelivered(ctx, msgID)
	require.NoError(t, err)

	// Verify status changed
	var updated identity.AgentMessage
	require.NoError(t, db.First(&updated, "id = ?", msgID).Error)
	assert.Equal(t, "delivered", updated.Status)
	assert.NotNil(t, updated.DeliveredAt)
}

func TestMessageService_DeleteMessage(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	svc := NewMessageServiceWithoutWorkers(db, nil)

	ctx := context.Background()

	// Create agent identities
	fromAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "from-agent-" + uuid.New().String()[:8],
		Name:      "From Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	toAgent := &identity.AgentIdentity{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		AgentID:   "to-agent-" + uuid.New().String()[:8],
		Name:      "To Agent",
		Status:    identity.AgentStatusActive,
		PlanTier:  "agent_starter",
		SwarmRole: identity.SwarmRoleWorker,
	}
	require.NoError(t, db.Create(fromAgent).Error)
	require.NoError(t, db.Create(toAgent).Error)

	msgID := uuid.New()
	msg := identity.AgentMessage{
		ID:          msgID,
		FromAgentID: fromAgent.AgentID,
		ToAgentID:   toAgent.AgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Status:      "pending",
	}
	require.NoError(t, db.Create(&msg).Error)

	err := svc.DeleteMessage(ctx, msgID, fromAgent.AgentID)
	require.NoError(t, err)

	// Verify deleted
	var count int64
	require.NoError(t, db.Model(&identity.AgentMessage{}).Where("id = ?", msgID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
