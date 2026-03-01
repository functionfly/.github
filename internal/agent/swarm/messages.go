package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageService handles agent-to-agent messaging
type MessageService struct {
	db *gorm.DB
}

// NewMessageService creates a new message service
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// SendMessage sends a message from one agent to another
func (s *MessageService) SendMessage(ctx context.Context, msg *identity.AgentMessage) error {
	// Validate message type
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

	valid := false
	for _, t := range validTypes {
		if msg.MessageType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid message type: %s", msg.MessageType)
	}

	// Set defaults
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	if msg.TTLSeconds == 0 {
		msg.TTLSeconds = 3600
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	msg.CreatedAt = time.Now()

	return s.db.WithContext(ctx).Create(msg).Error
}

// GetInbox retrieves pending messages for an agent
func (s *MessageService) GetInbox(ctx context.Context, agentID string, limit int) ([]identity.AgentMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	var messages []identity.AgentMessage
	err := s.db.WithContext(ctx).
		Where("to_agent_id = ? AND status IN ?", agentID, []string{"pending", "delivered"}).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetOutbox retrieves messages sent by an agent
func (s *MessageService) GetOutbox(ctx context.Context, agentID string, limit int) ([]identity.AgentMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	var messages []identity.AgentMessage
	err := s.db.WithContext(ctx).
		Where("from_agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// MarkDelivered marks a message as delivered
func (s *MessageService) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("id = ? AND status = ?", messageID, "pending").
		Updates(map[string]any{
			"status":       "delivered",
			"delivered_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found or already delivered")
	}
	return nil
}

// MarkRead marks a message as read
func (s *MessageService) MarkRead(ctx context.Context, messageID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("id = ?", messageID).
		Update("status", "read")
	return result.Error
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(ctx context.Context, messageID uuid.UUID, agentID string) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND (from_agent_id = ? OR to_agent_id = ?)", messageID, agentID, agentID).
		Delete(&identity.AgentMessage{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found or not authorized")
	}
	return nil
}

// CleanupExpired deletes expired messages
func (s *MessageService) CleanupExpired(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "pending", time.Now().Add(-24*time.Hour)).
		Delete(&identity.AgentMessage{})
	return result.RowsAffected, result.Error
}

// SendTaskDelegation sends a task from one agent to another
func (s *MessageService) SendTaskDelegation(ctx context.Context, fromAgentID, toAgentID string, taskData map[string]any, sessionID string) error {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Payload:     taskData,
		SessionID:   &sessionID,
		TTLSeconds:  3600,
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg)
}

// SendTaskResult sends a task result back to the delegating agent
func (s *MessageService) SendTaskResult(ctx context.Context, fromAgentID, toAgentID string, resultData map[string]any, sessionID, parentExecutionID string) error {
	resultData["parent_execution_id"] = parentExecutionID
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeTaskResult,
		Payload:     resultData,
		SessionID:   &sessionID,
		TTLSeconds:  3600,
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg)
}

// SendHeartbeat sends a heartbeat message between agents
func (s *MessageService) SendHeartbeat(ctx context.Context, fromAgentID, toAgentID string) error {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeHeartbeat,
		Payload:     map[string]any{"timestamp": time.Now().Unix()},
		TTLSeconds:  300, // 5 minutes for heartbeats
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg)
}
