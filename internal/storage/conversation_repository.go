package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// ListConversationsForUser returns conversations where the user is a participant, ordered by updated_at desc,
// with per-row unread counts (messages from others after last_read_at).
func (r *ConversationRepository) ListConversationsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]conversations.ConversationListEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	userIDStr := userID.String()
	var convs []conversations.Conversation
	payload, _ := json.Marshal([]string{userIDStr})
	err := r.db.WithContext(ctx).
		Where("participant_ids @> ?::jsonb", string(payload)).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&convs).Error
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	if len(convs) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, len(convs))
	for i := range convs {
		ids[i] = convs[i].ID
	}
	unreadByID, err := r.unreadCountsForConversations(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	out := make([]conversations.ConversationListEntry, len(convs))
	for i := range convs {
		c := convs[i]
		out[i] = conversations.ConversationListEntry{
			Conversation: c,
			UnreadCount:  unreadByID[c.ID],
		}
	}
	return out, nil
}

func (r *ConversationRepository) unreadCountsForConversations(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	type row struct {
		ConversationID uuid.UUID `gorm:"column:conversation_id"`
		Cnt            int       `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT m.conversation_id, COUNT(*)::int AS cnt
		FROM conversation_messages m
		LEFT JOIN conversation_participant_reads r
			ON r.conversation_id = m.conversation_id AND r.user_id = ?
		WHERE m.conversation_id IN ?
			AND m.author_id <> ?
			AND m.deleted_at IS NULL
			AND (
				r.last_read_message_id IS NULL
				OR m.id > r.last_read_message_id
			)
		GROUP BY m.conversation_id
	`, userID, ids, userID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("unread counts: %w", err)
	}
	m := make(map[uuid.UUID]int, len(rows))
	for _, rw := range rows {
		m[rw.ConversationID] = rw.Cnt
	}
	return m, nil
}

// MarkConversationRead sets the user's read cursor to the latest message in the conversation (or now if empty).
func (r *ConversationRepository) MarkConversationRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	var msgID *uuid.UUID
	var ts time.Time
	var latest struct {
		ID        uuid.UUID `gorm:"column:id"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, created_at FROM conversation_messages WHERE conversation_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`,
		conversationID,
	).Scan(&latest).Error; err == nil && latest.ID != uuid.Nil {
		msgID = &latest.ID
		ts = latest.CreatedAt
	}
	if ts.IsZero() {
		ts = time.Now()
	}
	row := conversations.ConversationParticipantRead{
		ConversationID:    conversationID,
		UserID:            userID,
		LastReadAt:        ts,
		LastReadMessageID: msgID,
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "conversation_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_at", "last_read_message_id"}),
	}).Create(&row).Error; err != nil {
		return fmt.Errorf("mark conversation read: %w", err)
	}
	return nil
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
// Soft-deleted messages are excluded.
func (r *ConversationRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*conversations.ConversationMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	var list []*conversations.ConversationMessage
	err := r.db.WithContext(ctx).
		Preload("Attachments").
		Where("conversation_id = ? AND deleted_at IS NULL", conversationID).
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

// EditMessage updates the content of a message and sets edited_at.
func (r *ConversationRepository) EditMessage(ctx context.Context, messageID uuid.UUID, newContent string) (*conversations.ConversationMessage, error) {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&conversations.ConversationMessage{}).
		Where("id = ? AND deleted_at IS NULL", messageID).
		Updates(map[string]interface{}{
			"content":   newContent,
			"edited_at": now,
		}).Error; err != nil {
		return nil, fmt.Errorf("edit message: %w", err)
	}
	return r.GetMessageByID(ctx, messageID)
}

// SoftDeleteMessage marks a message as deleted (soft delete).
func (r *ConversationRepository) SoftDeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&conversations.ConversationMessage{}).
		Where("id = ? AND deleted_at IS NULL", messageID).
		Update("deleted_at", now).Error; err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	return nil
}

// SearchMessages performs full-text search across conversation messages.
// If conversationID is non-nil, restricts to that conversation.
func (r *ConversationRepository) SearchMessages(ctx context.Context, userID uuid.UUID, query string, conversationID *uuid.UUID, limit, offset int) ([]*conversations.ConversationMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	q := r.db.WithContext(ctx).
		Preload("Attachments").
		Where("deleted_at IS NULL").
		Where("content_search @@ websearch_to_tsquery('english', ?)", query)

	if conversationID != nil {
		q = q.Where("conversation_id = ?", *conversationID)
	} else {
		// Only search messages in conversations the user participates in
		payload, _ := json.Marshal([]string{userID.String()})
		q = q.Where("conversation_id IN (SELECT id FROM conversations WHERE participant_ids @> ?::jsonb)", string(payload))
	}

	var list []*conversations.ConversationMessage
	err := q.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	return list, nil
}

// CreateAttachment creates a message attachment record.
func (r *ConversationRepository) CreateAttachment(ctx context.Context, a *conversations.MessageAttachment) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("create attachment: %w", err)
	}
	return nil
}

// ListAttachmentsForMessage returns all attachments for a given message.
func (r *ConversationRepository) ListAttachmentsForMessage(ctx context.Context, messageID uuid.UUID) ([]*conversations.MessageAttachment, error) {
	var list []*conversations.MessageAttachment
	if err := r.db.WithContext(ctx).Where("message_id = ?", messageID).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	return list, nil
}

// DeleteAttachment removes an attachment record.
func (r *ConversationRepository) DeleteAttachment(ctx context.Context, attachmentID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("id = ? AND uploaded_by = ?", attachmentID, userID).
		Delete(&conversations.MessageAttachment{}).Error; err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}
	return nil
}

// GetAttachmentByID returns a single attachment by ID.
func (r *ConversationRepository) GetAttachmentByID(ctx context.Context, id uuid.UUID) (*conversations.MessageAttachment, error) {
	var a conversations.MessageAttachment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get attachment: %w", err)
	}
	return &a, nil
}
