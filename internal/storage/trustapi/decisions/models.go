package decisions

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DecisionStatus represents the status of a team decision
type DecisionStatus string

const (
	DecisionStatusPending    DecisionStatus = "pending"
	DecisionStatusApproved   DecisionStatus = "approved"
	DecisionStatusSuperseded DecisionStatus = "superseded"
	DecisionStatusDeprecated DecisionStatus = "deprecated"
)

// DecisionSourceType represents how the decision was created
type DecisionSourceType string

const (
	DecisionSourceManual     DecisionSourceType = "manual"
	DecisionSourceAIExtracted DecisionSourceType = "ai_extracted"
	DecisionSourceImported   DecisionSourceType = "imported"
)

// TeamDecision represents a recorded team decision
type TeamDecision struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID      uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"size:500;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Rationale   string    `json:"rationale" gorm:"type:text"`          // Why this decision was made
	Outcome     string    `json:"outcome" gorm:"type:text"`             // What was decided
	Alternatives string   `json:"alternatives" gorm:"type:text"`        // JSON array of alternatives considered

	// Audit
	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Approval workflow
	Status     string `json:"status" gorm:"size:50;not null;default:'pending'"`
	ApprovedBy *uuid.UUID `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt *time.Time `json:"approved_at"`

	// Source tracking
	SourceType string  `json:"source_type" gorm:"size:50;default:'manual'"`
	SourceID   *string `json:"source_id" gorm:"size:255"` // Original message/task ID if AI extracted

	// Link to TeamMemory for AI-extracted content
	TeamMemoryID *uuid.UUID `json:"team_memory_id" gorm:"type:uuid"`

	// Tags for categorization
	Tags []string `json:"tags" gorm:"type:text"`

	// Importance for sorting/filtering
	ImportanceScore float64 `json:"importance_score" gorm:"default:0.5"`
}

// TableName returns the table name for TeamDecision
func (TeamDecision) TableName() string {
	return "team_decisions"
}

// GetAlternativesSlice returns alternatives as a slice of strings
func (d *TeamDecision) GetAlternativesSlice() []string {
	if d.Alternatives == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(d.Alternatives), &result); err != nil {
		return []string{}
	}
	return result
}

// SetAlternativesSlice stores alternatives as a JSON string
func (d *TeamDecision) SetAlternativesSlice(alternatives []string) error {
	if alternatives == nil {
		d.Alternatives = "[]"
		return nil
	}
	data, err := json.Marshal(alternatives)
	if err != nil {
		return err
	}
	d.Alternatives = string(data)
	return nil
}

// GetTagsSlice returns tags as a slice of strings
func (d *TeamDecision) GetTagsSlice() []string {
	if len(d.Tags) == 0 {
		return []string{}
	}
	return d.Tags
}

// ============================================
// Request/Response DTOs
// ============================================

// CreateDecisionRequest represents a request to create a new decision
type CreateDecisionRequest struct {
	Title        string   `json:"title" binding:"required,min=3,max=500"`
	Description  string   `json:"description"`
	Rationale    string   `json:"rationale"`
	Outcome      string   `json:"outcome"`
	Alternatives []string `json:"alternatives"`
	Tags         []string `json:"tags"`
	ImportanceScore float64 `json:"importance_score"`
}

// UpdateDecisionRequest represents a request to update a decision
type UpdateDecisionRequest struct {
	Title        *string  `json:"title" binding:"omitempty,min=3,max=500"`
	Description  *string  `json:"description"`
	Rationale    *string  `json:"rationale"`
	Outcome      *string  `json:"outcome"`
	Alternatives *[]string `json:"alternatives"`
	Tags         *[]string `json:"tags"`
	Status       *string  `json:"status" binding:"omitempty,oneof=pending approved superseded deprecated"`
	ImportanceScore *float64 `json:"importance_score"`
}

// ApproveDecisionRequest represents a request to approve/supersede/deprecate a decision
type ApproveDecisionRequest struct {
	Status string `json:"status" binding:"required,oneof=approved superseded deprecated"`
}

// ListDecisionsRequest represents query parameters for listing decisions
type ListDecisionsRequest struct {
	Status string `form:"status" binding:"omitempty,oneof=pending approved superseded deprecated"`
	Tag    string `form:"tag"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
}

// DecisionResponse represents a decision in API responses
type DecisionResponse struct {
	ID              uuid.UUID  `json:"id"`
	TeamID          uuid.UUID  `json:"team_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Rationale       string     `json:"rationale,omitempty"`
	Outcome         string     `json:"outcome,omitempty"`
	Alternatives    []string   `json:"alternatives,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Status          string     `json:"status"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	SourceType      string     `json:"source_type"`
	SourceID        *string    `json:"source_id,omitempty"`
	ImportanceScore float64    `json:"importance_score"`
}

// ToResponse converts a TeamDecision to a DecisionResponse
func (d *TeamDecision) ToResponse() *DecisionResponse {
	return &DecisionResponse{
		ID:              d.ID,
		TeamID:          d.TeamID,
		Title:           d.Title,
		Description:     d.Description,
		Rationale:       d.Rationale,
		Outcome:         d.Outcome,
		Alternatives:    d.GetAlternativesSlice(),
		Tags:            d.GetTagsSlice(),
		CreatedBy:       d.CreatedBy,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
		Status:          d.Status,
		ApprovedBy:      d.ApprovedBy,
		ApprovedAt:      d.ApprovedAt,
		SourceType:      d.SourceType,
		SourceID:        d.SourceID,
		ImportanceScore: d.ImportanceScore,
	}
}

// ListDecisionsResponse represents a paginated list of decisions
type ListDecisionsResponse struct {
	Decisions []DecisionResponse `json:"decisions"`
	TotalCount int64             `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// SearchDecisionsRequest represents query parameters for searching decisions
type SearchDecisionsRequest struct {
	Q     string `form:"q" binding:"required,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=50"`
}
