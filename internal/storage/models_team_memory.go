package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// Team Memory Engine - Shared Brain Models
// ============================================

// TeamMemory represents a shared team memory (decision, preference, process, client context)
type TeamMemory struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	TeamID   uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	Team     *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`

	// Memory categorization
	MemoryType string  `json:"memory_type" gorm:"not null;size:50;index"` // 'decision', 'preference', 'process', 'client_context'
	Category   *string `json:"category,omitempty" gorm:"size:100;index"`

	// Content (structured JSON) - NULL when encrypted
	Content   JSONMap   `json:"content,omitempty" gorm:"type:jsonb"`
	Summary   *string   `json:"summary,omitempty"` // Human-readable summary (always plaintext for search)
	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	Creator   *User     `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`

	// Source tracking
	SourceConversationID *uuid.UUID `json:"source_conversation_id,omitempty" gorm:"type:uuid"`
	SourceEventID        *uuid.UUID `json:"source_event_id,omitempty" gorm:"type:uuid"`

	// Vector embedding for semantic search (generated from plaintext before encryption)
	Embedding []float32 `json:"-" gorm:"type:vector(1536)"` // Exclude from JSON

	// Confidence & validation
	ConfidenceScore float64    `json:"confidence_score" gorm:"default:0.9"`
	IsValidated     bool       `json:"is_validated" gorm:"default:false;index"`
	ValidatedBy     *uuid.UUID `json:"validated_by,omitempty" gorm:"type:uuid"`
	Validator       *User      `json:"validator,omitempty" gorm:"foreignKey:ValidatedBy"`
	ValidatedAt     *time.Time `json:"validated_at,omitempty"`

	// Retention & priority (auto-decays with disuse)
	ImportanceScore float64    `json:"importance_score" gorm:"default:0.5"`
	TTLDays         int        `json:"ttl_days" gorm:"default:0"` // 0 = never expire
	ExpiresAt       *time.Time `json:"expires_at,omitempty" gorm:"index"`

	// Access tracking
	AccessCount    int        `json:"access_count" gorm:"default:0"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`

	// Auto-update tracking
	AutoUpdateEnabled    bool       `json:"auto_update_enabled" gorm:"default:true;index"`
	LastAutoUpdatedAt    *time.Time `json:"last_auto_updated_at,omitempty"`
	ExtractionConfidence *float64   `json:"extraction_confidence,omitempty"` // AI extraction confidence

	// Client-side encryption toggle (user-controlled zero-knowledge option)
	IsEncrypted      bool   `json:"is_encrypted" gorm:"default:false;index"`
	EncryptedContent []byte `json:"-" gorm:"type:bytea"` // Never serialize to JSON
	EncryptionIV     []byte `json:"-" gorm:"type:bytea"` // Never serialize to JSON
	EncryptionTag    []byte `json:"-" gorm:"type:bytea"` // Never serialize to JSON

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (TeamMemory) TableName() string {
	return "team_memories"
}

// BeforeCreate hook for TeamMemory
func (tm *TeamMemory) BeforeCreate(tx *gorm.DB) error {
	if tm.ID == uuid.Nil {
		tm.ID = uuid.New()
	}
	if tm.CreatedAt.IsZero() {
		tm.CreatedAt = time.Now()
	}
	if tm.UpdatedAt.IsZero() {
		tm.UpdatedAt = time.Now()
	}

	// Set default TTL based on memory type if not specified
	if tm.TTLDays == 0 {
		switch tm.MemoryType {
		case "decision":
			tm.TTLDays = 730 // 2 years
		case "preference":
			tm.TTLDays = 365 // 1 year
		case "process":
			tm.TTLDays = 0 // Never expire
		case "client_context":
			tm.TTLDays = 0 // Until marked inactive
		}
	}

	// Calculate expires_at from TTL
	if tm.TTLDays > 0 && tm.ExpiresAt == nil {
		expiresAt := tm.CreatedAt.AddDate(0, 0, tm.TTLDays)
		tm.ExpiresAt = &expiresAt
	}

	return nil
}

// BeforeUpdate hook for TeamMemory
func (tm *TeamMemory) BeforeUpdate(tx *gorm.DB) error {
	tm.UpdatedAt = time.Now()

	// Recalculate expires_at if TTL changed
	if tm.TTLDays > 0 {
		expiresAt := tm.UpdatedAt.AddDate(0, 0, tm.TTLDays)
		tm.ExpiresAt = &expiresAt
	} else if tm.TTLDays == 0 {
		tm.ExpiresAt = nil
	}

	return nil
}

// IsActive checks if memory is not expired
func (tm *TeamMemory) IsActive() bool {
	if tm.ExpiresAt == nil {
		return true
	}
	return tm.ExpiresAt.After(time.Now())
}

// GetContent returns content based on encryption state
// For encrypted memories, returns nil (client must decrypt)
func (tm *TeamMemory) GetContent() JSONMap {
	if tm.IsEncrypted {
		return nil // Client-side decryption required
	}
	return tm.Content
}

// GetSearchableText returns text for embedding generation
func (tm *TeamMemory) GetSearchableText() string {
	text := ""
	if tm.Summary != nil {
		text += *tm.Summary + " "
	}
	if !tm.IsEncrypted && tm.Content != nil {
		// Extract text from structured content
		text += extractTextFromContent(tm.Content, tm.MemoryType)
	}
	return text
}

// extractTextFromContent extracts searchable text from structured content
func extractTextFromContent(content JSONMap, memoryType string) string {
	switch memoryType {
	case "decision":
		if title, ok := content["title"].(string); ok {
			return title
		}
		if rationale, ok := content["rationale"].(string); ok {
			return rationale
		}
	case "preference":
		if subject, ok := content["subject"].(string); ok {
			if value, ok := content["value"].(string); ok {
				return subject + " " + value
			}
		}
	case "process":
		if name, ok := content["name"].(string); ok {
			return name
		}
	case "client_context":
		if clientName, ok := content["client_name"].(string); ok {
			return clientName
		}
	}

	// Fallback: convert entire content to string representation
	if content != nil {
		return fmt.Sprintf("%v", content)
	}
	return ""
}

// MemoryContent represents the structured content for different memory types
// This is used for type-safe content creation/validation
type MemoryContent struct {
	// For 'decision' type
	Decision *DecisionContent `json:"decision,omitempty"`

	// For 'preference' type
	Preference *PreferenceContent `json:"preference,omitempty"`

	// For 'process' type
	Process *ProcessContent `json:"process,omitempty"`

	// For 'client_context' type
	Client *ClientContextContent `json:"client,omitempty"`
}

// DecisionContent represents a team decision
type DecisionContent struct {
	Title         string   `json:"title"`
	Rationale     string   `json:"rationale"`
	Alternatives  []string `json:"alternatives,omitempty"`
	DecisionMaker string   `json:"decision_maker"`
	Date          string   `json:"date"`
	Status        string   `json:"status"` // 'active', 'superseded', 'deprecated'
}

// PreferenceContent represents a team/client preference
type PreferenceContent struct {
	Subject     string `json:"subject"`               // e.g., "email_style", "meeting_times"
	Value       string `json:"value"`                 // e.g., "short", "mornings"
	Context     string `json:"context,omitempty"`     // e.g., "for_client_emails"
	Stakeholder string `json:"stakeholder,omitempty"` // who prefers this
	Priority    int    `json:"priority"`              // 1-10
}

// ProcessContent represents a team process
type ProcessContent struct {
	Name      string   `json:"name"`
	Steps     []string `json:"steps"`
	Owner     string   `json:"owner"`
	Frequency string   `json:"frequency,omitempty"` // 'daily', 'weekly', 'on_demand'
	Tools     []string `json:"tools,omitempty"`
}

// ClientContextContent represents client-specific context
type ClientContextContent struct {
	ClientID       string            `json:"client_id"`
	ClientName     string            `json:"client_name"`
	Industry       string            `json:"industry,omitempty"`
	KeyContacts    []ContactInfo     `json:"key_contacts,omitempty"`
	Preferences    map[string]string `json:"preferences,omitempty"`
	ImportantDates []ImportantDate   `json:"important_dates,omitempty"`
	Notes          string            `json:"notes,omitempty"`
}

// ContactInfo represents a contact within client context
type ContactInfo struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email,omitempty"`
}

// ImportantDate represents a date relevant to the client
type ImportantDate struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

// TeamMemoryFilter provides filtering options for memory queries
type TeamMemoryFilter struct {
	MemoryType    *string
	Category      *string
	IsValidated   *bool
	MinConfidence *float64
	IsEncrypted   *bool
	CreatedAfter  *time.Time
	Limit         int
	Offset        int
}

// TeamMemorySearchResult represents a search result with relevance score
type TeamMemorySearchResult struct {
	TeamMemory
	RelevanceScore float64 `json:"relevance_score"` // Cosine similarity score
}

// MemoryExtraction represents an AI-extracted memory pending validation
type MemoryExtraction struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID         uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;not null;index"`

	// Extracted data
	MemoryType string  `json:"memory_type" gorm:"not null;size:50"`
	Category   *string `json:"category,omitempty" gorm:"size:100"`
	Content    JSONMap `json:"content" gorm:"type:jsonb;not null"`
	Summary    string  `json:"summary" gorm:"not null"`
	Confidence float64 `json:"confidence" gorm:"not null"`
	Rationale  string  `json:"rationale"` // Why AI thinks this is relevant

	// Review status
	Status          string     `json:"status" gorm:"default:'pending';size:20"` // 'pending', 'approved', 'rejected'
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`

	// Auto-apply if confidence >= 0.9 (MVP default)
	AutoApplyThreshold float64 `json:"auto_apply_threshold" gorm:"default:0.9"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (MemoryExtraction) TableName() string {
	return "memory_extractions"
}

// ShouldAutoApply determines if this extraction should be auto-applied based on confidence
func (me *MemoryExtraction) ShouldAutoApply() bool {
	return me.Confidence >= me.AutoApplyThreshold && me.Status == "pending"
}

// ToTeamMemory converts an approved extraction to a TeamMemory
func (me *MemoryExtraction) ToTeamMemory(createdBy uuid.UUID) *TeamMemory {
	now := time.Now()
	return &TeamMemory{
		TeamID:               me.TeamID,
		MemoryType:           me.MemoryType,
		Category:             me.Category,
		Content:              me.Content,
		Summary:              &me.Summary,
		CreatedBy:            createdBy,
		SourceConversationID: &me.ConversationID,
		ConfidenceScore:      me.Confidence,
		IsValidated:          true, // Pre-validated since human reviewed
		ValidatedAt:          &now,
		AutoUpdateEnabled:    true,
		ExtractionConfidence: &me.Confidence,
	}
}

// ============================================
// Memory Share Model (Cross-team collaboration)
// ============================================

// MemoryShare represents a shared memory between teams
type MemoryShare struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MemoryID       uuid.UUID  `json:"memory_id" gorm:"type:uuid;not null;index"`
	SourceTeamID   uuid.UUID  `json:"source_team_id" gorm:"type:uuid;not null;index"`
	TargetTeamID   uuid.UUID  `json:"target_team_id" gorm:"type:uuid;not null;index"`
	SharedBy       uuid.UUID  `json:"shared_by" gorm:"type:uuid;not null"`
	TargetMemoryID *uuid.UUID `json:"target_memory_id,omitempty" gorm:"type:uuid;index"`      // Memory created in target team (for revocation)
	ShareType      string     `json:"share_type" gorm:"size:20;not null;default:'reference'"` // 'reference', 'copy', 'fork'
	Permission     string     `json:"permission" gorm:"size:20;not null;default:'read'"`      // 'read', 'write', 'admin'
	Status         string     `json:"status" gorm:"size:20;not null;default:'pending'"`       // 'pending', 'accepted', 'rejected', 'revoked'
	Message        *string    `json:"message,omitempty" gorm:"type:text"`
	AcceptedBy     *uuid.UUID `json:"accepted_by,omitempty" gorm:"type:uuid"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty" gorm:"type:timestamptz"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MemoryShare) TableName() string {
	return "memory_shares"
}
