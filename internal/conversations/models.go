// Package conversations provides models and types for executable conversations (DMs with function-aware messages).
package conversations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message is a shared interface for any message type across the system
// (conversation messages, support messages, agent messages).
// Implementing this interface enables unified inbox views and cross-system search.
type Message interface {
	GetID() uuid.UUID
	GetConversationID() uuid.UUID
	GetAuthorID() uuid.UUID
	GetContent() string
	GetCreatedAt() time.Time
	IsDeleted() bool
}

// Ensure ConversationMessage implements Message.
var _ Message = (*ConversationMessage)(nil)

// GetID returns the message ID.
func (m *ConversationMessage) GetID() uuid.UUID { return m.ID }

// GetConversationID returns the conversation ID.
func (m *ConversationMessage) GetConversationID() uuid.UUID { return m.ConversationID }

// GetAuthorID returns the author ID.
func (m *ConversationMessage) GetAuthorID() uuid.UUID { return m.AuthorID }

// GetContent returns the message content.
func (m *ConversationMessage) GetContent() string { return m.Content }

// GetCreatedAt returns the creation time.
func (m *ConversationMessage) GetCreatedAt() time.Time { return m.CreatedAt }

// IsDeleted returns whether the message has been soft-deleted.
func (m *ConversationMessage) IsDeleted() bool { return m.DeletedAt != nil }

// ConversationType represents the type of conversation.
type ConversationType string

const (
	TypeDM                 ConversationType = "dm"
	TypeFunctionThread     ConversationType = "function_thread"
	TypeIssueThread        ConversationType = "issue_thread"
	TypeFixMode            ConversationType = "fix_mode"
	TypeBountyThread       ConversationType = "bounty_thread"
	TypeOrgThread          ConversationType = "org_thread"
	TypeSecurityDisclosure ConversationType = "security_disclosure"
)

// ConversationListEntry is one row for GET /conversations (includes unread for the current user).
type ConversationListEntry struct {
	Conversation
	UnreadCount int `json:"unread_count"`
}

// Conversation represents a conversation (DM or typed thread).
type Conversation struct {
	ID                  uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type                ConversationType `json:"type" gorm:"type:conversation_type_enum;not null;default:'dm'"`
	ParticipantIDs      json.RawMessage  `json:"participant_ids" gorm:"type:jsonb;not null;default:'[]'"`
	SourceThreadID      *uuid.UUID       `json:"source_thread_id,omitempty" gorm:"type:uuid"`
	OrganizationID      *uuid.UUID       `json:"organization_id,omitempty" gorm:"type:uuid"`
	Metadata            json.RawMessage  `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	ResolvedAt          *time.Time       `json:"resolved_at,omitempty" gorm:"type:timestamptz"`
	ResolvedByUserID    *uuid.UUID       `json:"resolved_by_user_id,omitempty" gorm:"type:uuid"`
	ResolvedByMessageID *uuid.UUID       `json:"resolved_by_message_id,omitempty" gorm:"type:uuid"`
	CreatedAt           time.Time        `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt           time.Time        `json:"updated_at" gorm:"not null;default:now()"`
}

// ConversationBounty represents a bounty attached to a conversation.
type ConversationBounty struct {
	ID                       uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID           uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null;index"`
	OfferedBy                uuid.UUID  `json:"offered_by" gorm:"type:uuid;not null;index"`
	AmountReputation         int        `json:"amount_reputation" gorm:"not null;default:0"`
	AmountCents              int        `json:"amount_cents" gorm:"default:0"`
	SecurityWeightMultiplier float64    `json:"security_weight_multiplier" gorm:"default:1.0"`
	ClaimedBy                *uuid.UUID `json:"claimed_by,omitempty" gorm:"type:uuid"`
	ClaimedAt                *time.Time `json:"claimed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt                time.Time  `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (ConversationBounty) TableName() string {
	return "conversation_bounties"
}

// TableName returns the table name.
func (Conversation) TableName() string {
	return "conversations"
}

// MessageEmbeddings holds optional function/execution references on a message.
type MessageEmbeddings struct {
	FunctionRef           *FunctionRef   `json:"function_ref,omitempty"`
	ExecutionID           *uuid.UUID     `json:"execution_id,omitempty"`
	ExecutionRootHash     string         `json:"execution_root_hash,omitempty"`
	ReplayLink            string         `json:"replay_link,omitempty"`
	CapabilityDeclaration map[string]any `json:"capability_declaration,omitempty"`
	InputSummary          string         `json:"input_summary,omitempty"`
	OutputSummary         string         `json:"output_summary,omitempty"`
}

// FunctionRef references a function by author/name/version.
type FunctionRef struct {
	Author  string `json:"author"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ConversationMessage represents a single message in a conversation.
type ConversationMessage struct {
	ID             uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID           `json:"conversation_id" gorm:"type:uuid;not null;index"`
	AuthorID       uuid.UUID           `json:"author_id" gorm:"type:uuid;not null;index"`
	Content        string              `json:"content" gorm:"type:text;not null;default:''"`
	Embeddings     json.RawMessage     `json:"embeddings" gorm:"type:jsonb;default:'{}'"`
	EditedAt       *time.Time          `json:"edited_at,omitempty" gorm:"type:timestamptz"`
	DeletedAt      *time.Time          `json:"deleted_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt      time.Time           `json:"created_at" gorm:"not null;default:now()"`
	Attachments    []MessageAttachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
}

// TableName returns the table name.
func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

// ConversationParticipantRead stores when a user last viewed a conversation (for unread counts).
type ConversationParticipantRead struct {
	ConversationID    uuid.UUID  `json:"conversation_id" gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;primaryKey"`
	LastReadAt        time.Time  `json:"last_read_at" gorm:"type:timestamptz;not null"`
	LastReadMessageID *uuid.UUID `json:"last_read_message_id,omitempty" gorm:"type:uuid"`
}

// TableName returns the table name.
func (ConversationParticipantRead) TableName() string {
	return "conversation_participant_reads"
}

// MessageAttachment represents a file/media attachment on a message.
type MessageAttachment struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID      uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;not null;index"`
	UploadedBy     uuid.UUID `json:"uploaded_by" gorm:"type:uuid;not null"`
	Filename       string    `json:"filename" gorm:"type:text;not null"`
	ContentType    string    `json:"content_type" gorm:"type:text;not null;default:'application/octet-stream'"`
	SizeBytes      int64     `json:"size_bytes" gorm:"not null;default:0"`
	StorageURL     string    `json:"storage_url" gorm:"type:text;not null"`
	ThumbnailURL   *string   `json:"thumbnail_url,omitempty" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (MessageAttachment) TableName() string {
	return "message_attachments"
}

// MessageReaction represents an emoji reaction on a message.
type MessageReaction struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Reaction  string   `json:"reaction" gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (MessageReaction) TableName() string {
	return "message_reactions"
}

// ReactionSummary represents aggregated reaction data for a message.
type ReactionSummary struct {
	Reaction string   `json:"reaction"`
	Count    int      `json:"count"`
	UserIDs  []string `json:"user_ids"`
}

// MessageReactions represents all reactions on a message with summary.
type MessageReactions struct {
	MessageID  uuid.UUID         `json:"message_id"`
	Reactions []ReactionSummary `json:"reactions"`
}

// MessageRead represents a read receipt for a message.
type MessageRead struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null;uniqueIndex:idx_message_read_unique"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_message_read_unique"`
	ReadAt    time.Time `json:"read_at" gorm:"type:timestamptz;not null;default:now()"`
}

// TableName returns the table name.
func (MessageRead) TableName() string {
	return "conversation_message_reads"
}
