package discovery

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	result := "{"
	for i, s := range a {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, strings.Replace(s, `"`, `""`, -1))
	}
	result += "}"
	return result, nil
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return a.scanString(string(v))
	case string:
		return a.scanString(v)
	}
	return fmt.Errorf("cannot scan type %T into StringArray", value)
}

func (a *StringArray) scanString(s string) error {
	if s == "" || s == "{}" {
		*a = []string{}
		return nil
	}
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	if s == "" {
		*a = []string{}
		return nil
	}
	parts := strings.Split(s, ",")
	*a = make([]string, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) >= 2 && p[0] == '"' && p[len(p)-1] == '"' {
			p = p[1 : len(p)-1]
		}
		(*a)[i] = strings.Replace(p, `""`, `"`, -1)
	}
	return nil
}

const (
	OpportunityStatusPending     = "pending"
	OpportunityStatusQualified   = "qualified"
	OpportunityStatusRejected    = "rejected"
	OpportunityStatusGenerated   = "generated"
	OpportunityStatusPublished   = "published"
	OpportunityStatusNeedsReview = "needs_review"
	ReviewStatusNotRequired      = "not_required"
	ReviewStatusPending          = "pending"
	ReviewStatusApproved         = "approved"
	ReviewStatusRejected         = "rejected"
	DefaultOpportunityCategory   = "automation"
	DefaultDiscoveryBatchLimit   = 25
)

// Source represents a discovery source implementation.
type Source interface {
	Name() string
	Scan(ctx context.Context) ([]OpportunityCandidate, error)
}

// OpportunityCandidate is a raw opportunity emitted by scanners before persistence.
type OpportunityCandidate struct {
	Source           string         `json:"source"`
	SourceID         string         `json:"source_id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Category         string         `json:"category"`
	Tags             []string       `json:"tags"`
	DemandSignal     float64        `json:"demand_signal"`
	ComplexitySignal int            `json:"complexity_signal"`
	Metadata         map[string]any `json:"metadata"`
	DiscoveredAt     time.Time      `json:"discovered_at"`
}

// Normalize ensures a candidate is usable by the service.
func (c *OpportunityCandidate) Normalize() {
	if c.Category == "" {
		c.Category = DefaultOpportunityCategory
	}
	if c.DiscoveredAt.IsZero() {
		c.DiscoveredAt = time.Now().UTC()
	}
	if c.Metadata == nil {
		c.Metadata = map[string]any{}
	}
	c.Title = strings.TrimSpace(c.Title)
	c.Description = strings.TrimSpace(c.Description)
	if c.DemandSignal < 0 {
		c.DemandSignal = 0
	}
	if c.DemandSignal > 100 {
		c.DemandSignal = 100
	}
	if c.ComplexitySignal < 1 {
		c.ComplexitySignal = 1
	}
	if c.ComplexitySignal > 10 {
		c.ComplexitySignal = 10
	}
}

// Opportunity is the persisted discovery record.
type Opportunity struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Source            string         `json:"source" gorm:"not null;index"`
	SourceID          string         `json:"source_id" gorm:"not null;index"`
	Title             string         `json:"title" gorm:"not null"`
	Description       string         `json:"description" gorm:"type:text"`
	Category          string         `json:"category" gorm:"not null;default:'automation'"`
	Tags              StringArray    `json:"tags" gorm:"type:text[]"`
	DemandScore       float64        `json:"demand_score" gorm:"type:decimal(5,2);default:0"`
	Complexity        int            `json:"complexity" gorm:"not null;default:1"`
	Validated         bool           `json:"validated" gorm:"not null;default:false"`
	Status            string         `json:"status" gorm:"not null;default:'pending';index"`
	QualityScore      float64        `json:"quality_score" gorm:"type:decimal(5,2);default:0"`
	ReviewStatus      string         `json:"review_status" gorm:"not null;default:'not_required'"`
	ReviewReason      *string        `json:"review_reason"`
	ReviewRequestedAt *time.Time     `json:"review_requested_at"`
	Metadata          map[string]any `json:"metadata" gorm:"serializer:json;default:'{}'"`
	GeneratedFuncID   *string        `json:"generated_func_id"`
	GenerationRunID   *string        `json:"generation_run_id"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Opportunity) TableName() string {
	return "factory_opportunities"
}

// DiscoveryBatch summarizes a scan execution.
type DiscoveryBatch struct {
	Source       string        `json:"source"`
	Discovered   int           `json:"discovered"`
	Persisted    int           `json:"persisted"`
	Deduplicated int           `json:"deduplicated"`
	Duration     time.Duration `json:"duration"`
}

// ReviewDecision captures a manual review action.
type ReviewDecision struct {
	Approved bool
	Reason   string
	Actor    string
}
