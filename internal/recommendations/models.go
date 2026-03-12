package recommendations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeSimilar            RecommendationType = "similar"
	RecommendationTypeFrequentlyUsedTogether RecommendationType = "frequently_used_together"
	RecommendationTypeSameCategory        RecommendationType = "same_category"
	RecommendationTypeTrending            RecommendationType = "trending"
	RecommendationTypePersonalized        RecommendationType = "personalized"
)

// InteractionType represents how a user interacts with a function
type InteractionType string

const (
	InteractionTypeView       InteractionType = "view"
	InteractionTypeExecute   InteractionType = "execute"
	InteractionTypeSave      InteractionType = "save"
	InteractionTypeFollow    InteractionType = "follow"
	InteractionTypeRate      InteractionType = "rate"
	InteractionTypeCopyCode  InteractionType = "copy_code"
	InteractionTypeShare     InteractionType = "share"
)

// FunctionCooccurrence tracks functions used together
type FunctionCooccurrence struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	FunctionIDA       uuid.UUID `json:"function_id_a" gorm:"type:uuid;index"`
	FunctionIDB       uuid.UUID `json:"function_id_b" gorm:"type:uuid;index"`
	CooccurrenceCount int       `json:"co_occurrence_count" gorm:"default:1"`
	LastCooccurredAt  time.Time `json:"last_co_occurred_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// FunctionSimilarity stores pre-computed similarity scores
type FunctionSimilarity struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	FunctionIDA           uuid.UUID `json:"function_id_a" gorm:"type:uuid;index"`
	FunctionIDB           uuid.UUID `json:"function_id_b" gorm:"type:uuid;index"`
	ContentSimilarity     float64   `json:"content_similarity" gorm:"default:0"`
	CollaborativeSimilarity float64 `json:"collaborative_similarity" gorm:"default:0"`
	CategorySimilarity   float64   `json:"category_similarity" gorm:"default:0"`
	CombinedSimilarity   float64   `json:"combined_similarity" gorm:"default:0"`
	ComputedAt           time.Time `json:"computed_at"`
	ComputationVersion   int       `json:"computation_version" gorm:"default:1"`
}

// UserFunctionInteraction tracks user interactions with functions
type UserFunctionInteraction struct {
	ID                  uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	UserID              *uuid.UUID     `json:"user_id" gorm:"type:uuid;index"`
	FunctionID          uuid.UUID       `json:"function_id" gorm:"type:uuid;index"`
	InteractionType    InteractionType `json:"interaction_type" gorm:"type:varchar(20)"`
	SessionID           *string        `json:"session_id" gorm:"type:varchar(100);index"`
	ReferrerFunctionID *uuid.UUID     `json:"referrer_function_id" gorm:"type:uuid"`
	Timestamp          time.Time       `json:"timestamp"`
}

// FunctionRecommendation represents a pre-computed recommendation
type FunctionRecommendation struct {
	ID                     uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey"`
	FunctionID             uuid.UUID          `json:"function_id" gorm:"type:uuid;index"`
	RecommendedFunctionID  uuid.UUID          `json:"recommended_function_id" gorm:"type:uuid;index"`
	RecommendationScore   float64            `json:"recommendation_score" gorm:"type:numeric(5,4)"`
	RecommendationType    RecommendationType `json:"recommendation_type" gorm:"type:varchar(20)"`
	RankPosition           int                `json:"rank_position" gorm:"default:0"`
	ComputedAt             time.Time          `json:"computed_at"`
	ExpiresAt             *time.Time        `json:"expires_at" gorm:"type:timestamp"`
}

// SessionFunctionUsage tracks functions used in a session
type SessionFunctionUsage struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	SessionID     string    `json:"session_id" gorm:"type:varchar(100);index"`
	FunctionID    uuid.UUID `json:"function_id" gorm:"type:uuid;index"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	ExecutionCount int      `json:"execution_count" gorm:"default:1"`
	FirstUsedAt   time.Time `json:"first_used_at"`
	LastUsedAt    time.Time `json:"last_used_at"`
}

// FunctionEmbedding stores a vector embedding for a function (semantic similarity via pgvector).
type FunctionEmbedding struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID     uuid.UUID `json:"function_id" gorm:"type:uuid;not null;uniqueIndex"`
	Embedding      []float32 `json:"embedding" gorm:"type:vector(1536)"`
	EmbeddedText   *string   `json:"embedded_text,omitempty" gorm:"type:text"`
	EmbeddingModel string    `json:"embedding_model" gorm:"type:varchar(100);default:text-embedding-ada-002"`
	ComputedAt     time.Time `json:"computed_at"`
}

// TableName overrides the table name for FunctionEmbedding.
func (FunctionEmbedding) TableName() string {
	return "function_embeddings"
}

// CategorySimilarity stores category relationship scores
type CategorySimilarity struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	CategoryA        string    `json:"category_a" gorm:"type:varchar(50)"`
	CategoryB        string    `json:"category_b" gorm:"type:varchar(50)"`
	SimilarityScore  float64   `json:"similarity_score" gorm:"type:numeric(5,4)"`
	SharedFunctions  int       `json:"shared_functions" gorm:"default:0"`
	CooccurrenceCount int      `json:"co_occurrence_count" gorm:"default:0"`
	ComputedAt        time.Time `json:"computed_at"`
}

// RecommendationFeedback tracks user feedback on recommendations
type RecommendationFeedback struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID                *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	FunctionID            uuid.UUID `json:"function_id" gorm:"type:uuid;index"`
	RecommendedFunctionID uuid.UUID `json:"recommended_function_id" gorm:"type:uuid;index"`
	FeedbackType         string    `json:"feedback_type" gorm:"type:varchar(20)"`
	RecommendationType   *string   `json:"recommendation_type" gorm:"type:varchar(20)"`
	Timestamp            time.Time `json:"timestamp"`
}

// RecommendationRequest represents a request for recommendations
type RecommendationRequest struct {
	FunctionID       *uuid.UUID `json:"function_id,omitempty"`       // Get related to this function
	UserID          *uuid.UUID `json:"user_id,omitempty"`          // Personalized for this user
	Category        *string    `json:"category,omitempty"`         // Get functions in this category
	Query           *string    `json:"query,omitempty"`            // Search query for use case
	Limit           int        `json:"limit,omitempty"`            // Number of recommendations (default 10)
	Offset          int        `json:"offset,omitempty"`           // Pagination offset
	Types           []RecommendationType `json:"types,omitempty"`  // Types of recommendations to include
	IncludePersonalized bool       `json:"include_personalized"`   // Include personalized recommendations
}

// RecommendationResult represents a single recommendation
type RecommendationResult struct {
	FunctionID       uuid.UUID `json:"function_id"`
	Author           string    `json:"author"`
	Name             string    `json:"name"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	Tags             []string  `json:"tags"`
	PopularityScore  int       `json:"popularity_score"`
	ReliabilityScore float64   `json:"reliability_score"`
	Score            float64   `json:"score"`
	RecommendationType RecommendationType `json:"recommendation_type"`
}

// RecommendationResponse represents the API response for recommendations
type RecommendationResponse struct {
	Recommendations []RecommendationResult `json:"recommendations"`
	Total           int                   `json:"total"`
	Limit           int                   `json:"limit"`
	Offset          int                   `json:"offset"`
	TypesIncluded   []RecommendationType   `json:"types_included"`
}

// RecommendationConfig holds configuration for the recommendation service
type RecommendationConfig struct {
	// Content-based similarity weights
	ContentWeight float64 `json:"content_weight"`
	// Collaborative filtering weight
	CollaborativeWeight float64 `json:"collaborative_weight"`
	// Category similarity weight
	CategoryWeight float64 `json:"category_weight"`
	// Minimum similarity score to be considered
	MinSimilarityScore float64 `json:"min_similarity_score"`
	// Maximum recommendations to compute per function
	MaxRecommendationsPerFunction int `json:"max_recommendations_per_function"`
	// Cache TTL for recommendations (minutes)
	CacheTTLMinutes int `json:"cache_ttl_minutes"`
	// Enable personalized recommendations
	EnablePersonalized bool `json:"enable_personalized"`
	// Enable trending recommendations
	EnableTrending bool `json:"enable_trending"`
}

// DefaultRecommendationConfig returns sensible defaults
func DefaultRecommendationConfig() *RecommendationConfig {
	return &RecommendationConfig{
		ContentWeight:               0.4,
		CollaborativeWeight:         0.35,
		CategoryWeight:              0.25,
		MinSimilarityScore:          0.1,
		MaxRecommendationsPerFunction: 20,
		CacheTTLMinutes:            60,
		EnablePersonalized:         true,
		EnableTrending:             true,
	}
}

// MarshalJSON implements custom JSON marshaling for RecommendationConfig
func (r *RecommendationConfig) MarshalJSON() ([]byte, error) {
	type Alias RecommendationConfig
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}
