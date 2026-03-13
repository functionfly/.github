package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationRepository handles persistence for conversations and messages.
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new ConversationRepository.
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// CreateConversation creates a new conversation.
func (r *ConversationRepository) CreateConversation(ctx context.Context, c *conversations.Conversation) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}
	return nil
}

// GetConversationByID returns a conversation by ID.
func (r *ConversationRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*conversations.Conversation, error) {
	var c conversations.Conversation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return &c, nil
}

// ListConversationsForUser returns conversations where the user is a participant, ordered by updated_at desc.
func (r *ConversationRepository) ListConversationsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*conversations.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	userIDStr := userID.String()
	var list []*conversations.Conversation
	// participant_ids is JSONB array of UUID strings; check containment
	payload, _ := json.Marshal([]string{userIDStr})
	err := r.db.WithContext(ctx).
		Where("participant_ids @> ?::jsonb", string(payload)).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	return list, nil
}

// IsParticipant returns true if userID is in the conversation's participant_ids.
func (r *ConversationRepository) IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var c conversations.Conversation
	if err := r.db.WithContext(ctx).Select("participant_ids").Where("id = ?", conversationID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("get conversation: %w", err)
	}
	var ids []string
	if err := json.Unmarshal(c.ParticipantIDs, &ids); err != nil {
		return false, fmt.Errorf("parse participant_ids: %w", err)
	}
	s := userID.String()
	for _, id := range ids {
		if id == s {
			return true, nil
		}
	}
	return false, nil
}

// CreateMessage creates a new message in a conversation.
func (r *ConversationRepository) CreateMessage(ctx context.Context, m *conversations.ConversationMessage) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	// Update conversation updated_at
	if err := r.db.WithContext(ctx).Model(&conversations.Conversation{}).Where("id = ?", m.ConversationID).Update("updated_at", m.CreatedAt).Error; err != nil {
		return fmt.Errorf("update conversation updated_at: %w", err)
	}
	return nil
}

// ListMessages returns messages for a conversation, paginated by created_at desc (newest first).
func (r *ConversationRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*conversations.ConversationMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	var list []*conversations.ConversationMessage
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return list, nil
}

// GetMessageByID returns a single message by ID.
func (r *ConversationRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*conversations.ConversationMessage, error) {
	var m conversations.ConversationMessage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get message: %w", err)
	}
	return &m, nil
}

// ResolveConversation marks a conversation as resolved. messageID is optional.
func (r *ConversationRepository) ResolveConversation(ctx context.Context, conversationID, resolvedByUserID uuid.UUID, messageID *uuid.UUID) error {
	updates := map[string]interface{}{
		"resolved_at":         time.Now(),
		"resolved_by_user_id": resolvedByUserID,
		"updated_at":          time.Now(),
	}
	if messageID != nil && *messageID != uuid.Nil {
		updates["resolved_by_message_id"] = *messageID
	}
	if err := r.db.WithContext(ctx).Model(&conversations.Conversation{}).Where("id = ?", conversationID).Updates(updates).Error; err != nil {
		return fmt.Errorf("resolve conversation: %w", err)
	}
	return nil
}

// CreateBounty attaches a bounty to a conversation.
func (r *ConversationRepository) CreateBounty(ctx context.Context, b *conversations.ConversationBounty) error {
	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		return fmt.Errorf("create bounty: %w", err)
	}
	return nil
}

// ListBountiesForConversation returns bounties for a conversation.
func (r *ConversationRepository) ListBountiesForConversation(ctx context.Context, conversationID uuid.UUID) ([]*conversations.ConversationBounty, error) {
	var list []*conversations.ConversationBounty
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list bounties: %w", err)
	}
	return list, nil
}

// GetBountyByID returns a bounty by ID.
func (r *ConversationRepository) GetBountyByID(ctx context.Context, id uuid.UUID) (*conversations.ConversationBounty, error) {
	var b conversations.ConversationBounty
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get bounty: %w", err)
	}
	return &b, nil
}

// ClaimBounty marks a bounty as claimed by the given user.
func (r *ConversationRepository) ClaimBounty(ctx context.Context, bountyID, claimedBy uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&conversations.ConversationBounty{}).Where("id = ? AND claimed_by IS NULL", bountyID).Updates(map[string]interface{}{
		"claimed_by": claimedBy,
		"claimed_at": now,
	}).Error; err != nil {
		return fmt.Errorf("claim bounty: %w", err)
	}
	return nil
}
