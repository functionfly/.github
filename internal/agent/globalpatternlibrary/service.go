package globalpatternlibrary

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SharingTier constants define who can access a pattern.
const (
	SharingTierUniversal   = "universal"   // any tenant
	SharingTierVertical   = "vertical"     // same industry vertical only
	SharingTierCrossVertical = "cross_vertical" // enterprise add-on only
)

// PatternType constants for classifying optimization patterns.
const (
	PatternTypeCheckoutRetry     = "checkout_retry"
	PatternTypeFraudRouting      = "fraud_routing"
	PatternTypeFrictionReduction = "friction_reduction"
	PatternTypeLatencyOpt       = "latency_opt"
	PatternTypeConversionBoost  = "conversion_boost"
	PatternTypeRetryLogic       = "retry_logic"
	PatternTypeCircuitBreaker   = "circuit_breaker"
)

// VerticalTag constants for industry classification.
const (
	VerticalECommerce  = "e-commerce"
	VerticalSaaS       = "saas"
	VerticalMarketplace = "marketplace"
	VerticalUniversal  = "universal"
)

// GlobalOptimizationPattern represents an anonymized, aggregated optimization pattern.
type GlobalOptimizationPattern struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	VerticalTag           string    `json:"vertical_tag" gorm:"type:text;not null;index"`
	PatternType           string    `json:"pattern_type" gorm:"type:text;not null;index"`
	Description          string    `json:"description" gorm:"type:text"`
	AnonymizedSignature   string    `json:"anonymized_signature" gorm:"type:jsonb"`
	ObservedImprovementPct float64  `json:"observed_improvement_pct"`
	SampleSize            int       `json:"sample_size"`
	ConfidenceScore       float64   `json:"confidence_score"`
	SharingTier           string    `json:"sharing_tier" gorm:"type:text;not null;default:'vertical'"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name.
func (GlobalOptimizationPattern) TableName() string { return "global_optimization_patterns" }

// Service manages the global pattern library.
type Service struct {
	db *gorm.DB
}

// NewService creates a new global pattern library service.
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Query returns patterns matching the given criteria.
// It filters by vertical and sharing tier to enforce isolation rules.
func (s *Service) Query(ctx context.Context, params QueryParams) ([]GlobalOptimizationPattern, error) {
	query := s.db.WithContext(ctx).
		Where("vertical_tag IN ?", params.VerticalTags).
		Where("sharing_tier IN ?", params.SharingTiers)

	if params.PatternType != "" {
		query = query.Where("pattern_type = ?", params.PatternType)
	}

	if params.MinConfidence > 0 {
		query = query.Where("confidence_score >= ?", params.MinConfidence)
	}

	var patterns []GlobalOptimizationPattern
	err := query.
		Order("confidence_score DESC").
		Limit(params.Limit).
		Find(&patterns).Error

	return patterns, err
}

// QueryParams defines filtering parameters for pattern queries.
type QueryParams struct {
	VerticalTags   []string  // e.g., ["e-commerce", "universal"]
	SharingTiers   []string  // e.g., ["universal", "vertical"]
	PatternType    string
	MinConfidence  float64
	Limit          int
}

// RecordPattern records a new optimization pattern.
// Patterns are anonymized before storage — no tenant-specific data is stored.
func (s *Service) RecordPattern(ctx context.Context, pattern *GlobalOptimizationPattern) error {
	if pattern.ID == uuid.Nil {
		pattern.ID = uuid.New()
	}
	if pattern.CreatedAt.IsZero() {
		pattern.CreatedAt = time.Now().UTC()
	}
	// Ensure anonymized_signature is valid JSON
	if pattern.AnonymizedSignature == "" {
		pattern.AnonymizedSignature = "{}"
	}
	return s.db.WithContext(ctx).Create(pattern).Error
}

// GetPatternsForVertical returns all patterns accessible to a given vertical.
func (s *Service) GetPatternsForVertical(ctx context.Context, verticalTag string, limit int) ([]GlobalOptimizationPattern, error) {
	var patterns []GlobalOptimizationPattern
	err := s.db.WithContext(ctx).
		Where("vertical_tag IN ?", []string{verticalTag, VerticalUniversal}).
		Where("sharing_tier IN ?", []string{SharingTierUniversal, SharingTierVertical}).
		Order("confidence_score DESC").
		Limit(limit).
		Find(&patterns).Error
	return patterns, err
}

// AutoMigrate runs database migrations for the pattern library.
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&GlobalOptimizationPattern{})
}