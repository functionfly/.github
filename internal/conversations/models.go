// Package conversations provides models and types for executable conversations (DMs with function-aware messages).
package conversations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ConversationType represents the type of conversation.
type ConversationType string

const (
	TypeDM                  ConversationType = "dm"
	TypeFunctionThread      ConversationType = "function_thread"
	TypeIssueThread         ConversationType = "issue_thread"
	TypeFixMode             ConversationType = "fix_mode"
	TypeBountyThread        ConversationType = "bounty_thread"
	TypeOrgThread           ConversationType = "org_thread"
	TypeSecurityDisclosure  ConversationType = "security_disclosure"
)

// ConversationListEntry is one row for GET /conversations (includes unread for the current user).
type ConversationListEntry struct {
	Conversation
	UnreadCount int `json:"unread_count"`
}

// Conversation represents a conversation (DM or typed thread).
type Conversation struct {
	ID                   uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type                 ConversationType `json:"type" gorm:"type:conversation_type_enum;not null;default:'dm'"`
	ParticipantIDs       json.RawMessage  `json:"participant_ids" gorm:"type:jsonb;not null;default:'[]'"`
	SourceThreadID       *uuid.UUID       `json:"source_thread_id,omitempty" gorm:"type:uuid"`
	OrganizationID       *uuid.UUID       `json:"organization_id,omitempty" gorm:"type:uuid"`
	Metadata             json.RawMessage  `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	ResolvedAt           *time.Time       `json:"resolved_at,omitempty" gorm:"type:timestamptz"`
	ResolvedByUserID     *uuid.UUID       `json:"resolved_by_user_id,omitempty" gorm:"type:uuid"`
	ResolvedByMessageID  *uuid.UUID       `json:"resolved_by_message_id,omitempty" gorm:"type:uuid"`
	CreatedAt            time.Time        `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt            time.Time        `json:"updated_at" gorm:"not null;default:now()"`
}

// ConversationBounty represents a bounty attached to a conversation.
type ConversationBounty struct {
	ID                     uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID        uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null;index"`
	OfferedBy              uuid.UUID  `json:"offered_by" gorm:"type:uuid;not null;index"`
	AmountReputation       int        `json:"amount_reputation" gorm:"not null;default:0"`
	AmountCents            int        `json:"amount_cents" gorm:"default:0"`
	SecurityWeightMultiplier float64  `json:"security_weight_multiplier" gorm:"default:1.0"`
	ClaimedBy              *uuid.UUID `json:"claimed_by,omitempty" gorm:"type:uuid"`
	ClaimedAt              *time.Time `json:"claimed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt              time.Time  `json:"created_at" gorm:"not null;default:now()"`
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
	FunctionRef          *FunctionRef          `json:"function_ref,omitempty"`
	ExecutionID          *uuid.UUID           `json:"execution_id,omitempty"`
	ExecutionRootHash    string               `json:"execution_root_hash,omitempty"`
	ReplayLink           string               `json:"replay_link,omitempty"`
	CapabilityDeclaration map[string]any      `json:"capability_declaration,omitempty"`
	InputSummary         string               `json:"input_summary,omitempty"`
	OutputSummary        string               `json:"output_summary,omitempty"`
}

// FunctionRef references a function by author/name/version.
type FunctionRef struct {
	Author  string `json:"author"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ConversationMessage represents a single message in a conversation.
type ConversationMessage struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID       `json:"conversation_id" gorm:"type:uuid;not null;index"`
	AuthorID       uuid.UUID       `json:"author_id" gorm:"type:uuid;not null;index"`
	Content        string          `json:"content" gorm:"type:text;not null;default:''"`
	Embeddings     json.RawMessage `json:"embeddings" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time       `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

// ConversationParticipantRead stores when a user last viewed a conversation (for unread counts).
type ConversationParticipantRead struct {
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	LastReadAt     time.Time `json:"last_read_at" gorm:"type:timestamptz;not null"`
}

// TableName returns the table name.
func (ConversationParticipantRead) TableName() string {
	return "conversation_participant_reads"
}
