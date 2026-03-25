// Package support provides models and types for the FunctionFly AI + Human Co-Pilot support system.
package support

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ConversationType represents the type of support conversation.
type ConversationType string

const (
	TypeSupportAI       ConversationType = "support_ai"
	TypeSupportHuman    ConversationType = "support_human"
	TypeSupportEmergency ConversationType = "support_emergency"
)

// MessageType represents the type of support message.
type MessageType string

const (
	TypeMessage    MessageType = "message"
	TypeContext    MessageType = "context"
	TypeEscalation MessageType = "escalation"
	TypeResolution MessageType = "resolution"
	TypeAIResponse MessageType = "ai_response"
	TypeSystem     MessageType = "system"
)

// SupportStatus represents the status of a support conversation.
type SupportStatus string

const (
	StatusActive    SupportStatus = "active"
	StatusPending   SupportStatus = "pending"
	StatusResolved  SupportStatus = "resolved"
	StatusEscalated SupportStatus = "escalated"
)

// Priority represents the priority level of a support conversation.
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// SupportConversation represents a support conversation session.
type SupportConversation struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Type           ConversationType `json:"type" gorm:"type:support_conversation_type_enum;not null"`
	Status         SupportStatus    `json:"status" gorm:"type:support_status_enum;not null;default:'active'"`
	Priority       Priority         `json:"priority" gorm:"type:support_priority_enum;not null;default:'normal'"`
	Title          string           `json:"title" gorm:"type:text"`

	// Context information gathered automatically
	FunctionRef    *FunctionRef     `json:"function_ref,omitempty" gorm:"-"`
	FunctionRefJSON json.RawMessage `json:"-" gorm:"type:jsonb;column:function_ref"`

	// Deployment information
	DeploymentID   *uuid.UUID      `json:"deployment_id,omitempty" gorm:"type:uuid"`
	DeploymentLogs string          `json:"deployment_logs,omitempty" gorm:"type:text"`
	DeploymentError string          `json:"deployment_error,omitempty" gorm:"type:text"`

	// AI handling
	AIHandled      bool            `json:"ai_handled" gorm:"default:false"`
	AIAttempts     int             `json:"ai_attempts" gorm:"default:0"`

	// Human escalation
	StaffID        *uuid.UUID      `json:"staff_id,omitempty" gorm:"type:uuid"`
	StaffJoinedAt  *time.Time      `json:"staff_joined_at,omitempty" gorm:"type:timestamptz"`

	// Resolution
	ResolvedAt     *time.Time      `json:"resolved_at,omitempty" gorm:"type:timestamptz"`
	ResolvedByID   *uuid.UUID      `json:"resolved_by_id,omitempty" gorm:"type:uuid"`
	ResolutionNote string          `json:"resolution_note,omitempty" gorm:"type:text"`

	// Emergency handling
	IsEmergency    bool            `json:"is_emergency" gorm:"default:false"`
	EmergencyCode  string          `json:"emergency_code,omitempty" gorm:"type:text"`

	// Metadata
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	CreatedAt      time.Time       `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (SupportConversation) TableName() string {
	return "support_conversations"
}

// FunctionRef references a function by author/name/version.
type FunctionRef struct {
	Author  string `json:"author"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// SupportMessage represents a single message in a support conversation.
type SupportMessage struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID       `json:"conversation_id" gorm:"type:uuid;not null;index"`
	AuthorID       uuid.UUID       `json:"author_id" gorm:"type:uuid;not null"`
	AuthorType     AuthorType      `json:"author_type" gorm:"type:support_author_type_enum;not null"`
	MessageType    MessageType     `json:"message_type" gorm:"type:support_message_type_enum;not null;default:'message'"`
	Content        string          `json:"content" gorm:"type:text;not null"`

	// AI-specific fields
	AIConfidence   *float64        `json:"ai_confidence,omitempty" gorm:"type:float"`
	AIModel       string          `json:"ai_model,omitempty" gorm:"type:text"`

	// Context embedding
	Embeddings     json.RawMessage `json:"embeddings,omitempty" gorm:"type:jsonb;default:'{}'"`

	// Attachments (logs, screenshots, etc.)
	Attachments    json.RawMessage `json:"attachments,omitempty" gorm:"type:jsonb;default:'[]'"`

	CreatedAt      time.Time       `json:"created_at" gorm:"not null;default:now()"`
}

// AuthorType represents who authored the message.
type AuthorType string

const (
	AuthorUser  AuthorType = "user"
	AuthorAI    AuthorType = "ai"
	AuthorStaff AuthorType = "staff"
	AuthorSystem AuthorType = "system"
)

// TableName returns the table name.
func (SupportMessage) TableName() string {
	return "support_messages"
}

// SupportContext holds contextual information about a support session.
type SupportContext struct {
	FunctionCode    string            `json:"function_code,omitempty"`
	FunctionLogs    []string          `json:"function_logs,omitempty"`
	DeploymentError string            `json:"deployment_error,omitempty"`
	EnvironmentVars map[string]string `json:"environment_vars,omitempty"`
	ExecutionHistory []ExecutionEntry `json:"execution_history,omitempty"`
	UserInfo        *UserContext      `json:"user_info,omitempty"`
}

// ExecutionEntry represents a single execution attempt.
type ExecutionEntry struct {
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Duration  int64     `json:"duration_ms"`
	Error     string    `json:"error,omitempty"`
}

// UserContext holds user information for support context.
type UserContext struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Plan      string    `json:"plan"`
	OrgID     uuid.UUID `json:"org_id"`
	CreatedAt time.Time `json:"created_at"`
}

// StaffAvailability tracks staff online/offline status.
type StaffAvailability struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StaffID   uuid.UUID `json:"staff_id" gorm:"type:uuid;not null;uniqueIndex"`
	IsOnline  bool      `json:"is_online" gorm:"default:false"`
	LastSeen  time.Time `json:"last_seen" gorm:"type:timestamptz"`
	MaxChats  int       `json:"max_chats" gorm:"default:5"`
	ActiveChats int     `json:"active_chats" gorm:"default:0"`
	CanAccept bool      `json:"can_accept" gorm:"default:true"`

	// specialties stored as JSON array
	Specialties  json.RawMessage `json:"specialties" gorm:"type:jsonb;default:'[]'"`

	CreatedAt    time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (StaffAvailability) TableName() string {
	return "staff_availability"
}

// SupportConversationParticipant tracks who is in a conversation.
type SupportConversationParticipant struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Role           string    `json:"role" gorm:"type:text;not null"` // "requester", "helper", "observer"
	JoinedAt       time.Time `json:"joined_at" gorm:"type:timestamptz;not null"`
	LeftAt         *time.Time `json:"left_at,omitempty" gorm:"type:timestamptz"`
}

// TableName returns the table name.
func (SupportConversationParticipant) TableName() string {
	return "support_conversation_participants"
}

// EmergencyFixRequest tracks emergency fix button activations.
type EmergencyFixRequest struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	FunctionID     uuid.UUID `json:"function_id" gorm:"type:uuid;not null"`
	Reason         string    `json:"reason" gorm:"type:text"`

	// Status tracking
	Status        string     `json:"status" gorm:"type:text;not null;default:'pending'"`
	StaffID       *uuid.UUID `json:"staff_id,omitempty" gorm:"type:uuid"`
	StaffAcceptedAt *time.Time `json:"staff_accepted_at,omitempty" gorm:"type:timestamptz"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty" gorm:"type:timestamptz"`

	// What was done
	FixDescription string    `json:"fix_description,omitempty" gorm:"type:text"`

	CreatedAt      time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (EmergencyFixRequest) TableName() string {
	return "emergency_fix_requests"
}

// BeforeCreate marshals the FunctionRef
func (s *SupportConversation) BeforeCreate() error {
	if s.FunctionRef != nil {
		data, err := json.Marshal(s.FunctionRef)
		if err != nil {
			return err
		}
		s.FunctionRefJSON = data
	}
	return nil
}

// AfterFind unmarshals the FunctionRef
func (s *SupportConversation) AfterFind() error {
	if s.FunctionRefJSON != nil && len(s.FunctionRefJSON) > 0 {
		var ref FunctionRef
		if err := json.Unmarshal(s.FunctionRefJSON, &ref); err != nil {
			return err
		}
		s.FunctionRef = &ref
	}
	return nil
}
