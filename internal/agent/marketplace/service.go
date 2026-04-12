package marketplace

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Service handles marketplace operations for agents and functions
type Service struct {
	db *gorm.DB
}

// NewService creates a new marketplace service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Ranking weights for marketplace
const (
	WeightTrustScore    = 0.30
	WeightEconomicScore = 0.25
	WeightReliability   = 0.20
	WeightROI           = 0.15
	WeightCallVolume    = 0.10
)

// ListingAgent creates a marketplace listing for an agent
func (s *Service) ListingAgent(ctx context.Context, req *CreateAgentListingRequest) (*identity.AgentListing, error) {
	// Validate agent exists
	var agent identity.AgentIdentity
	if err := s.db.WithContext(ctx).Where("agent_id = ?", req.AgentID).First(&agent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Check if listing already exists
	var existing identity.AgentListing
	err := s.db.WithContext(ctx).Where("agent_id = ?", req.AgentID).First(&existing).Error
	if err == nil {
		// Update existing listing
		existing.ListingType = req.ListingType
		existing.PricingModel = req.PricingModel
		existing.PricePerCall = req.PricePerCall
		existing.SubscriptionMonthlyUSD = req.SubscriptionMonthlyUSD
		existing.RevenueSharePercent = req.RevenueSharePercent
		existing.IsActive = req.IsActive
		existing.UpdatedAt = time.Now()
		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new listing
	listing := &identity.AgentListing{
		ID:                     uuid.New(),
		AgentID:                req.AgentID,
		ListingType:            req.ListingType,
		PricingModel:           req.PricingModel,
		PricePerCall:           req.PricePerCall,
		SubscriptionMonthlyUSD: req.SubscriptionMonthlyUSD,
		RevenueSharePercent:    req.RevenueSharePercent,
		RatingScore:            0,
		TotalCalls:             0,
		ROIScore:               0,
		IsActive:               req.IsActive,
		ListedAt:               time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(listing).Error; err != nil {
		return nil, err
	}

	return listing, nil
}

// CreateAgentListingRequest represents a request to create an agent listing
type CreateAgentListingRequest struct {
	AgentID                string   `json:"agent_id"`
	ListingType            string   `json:"listing_type"`  // worker | manager | infrastructure
	PricingModel           string   `json:"pricing_model"` // free | per_call | subscription | revenue_share
	PricePerCall           *float64 `json:"price_per_call,omitempty"`
	SubscriptionMonthlyUSD *float64 `json:"subscription_monthly_usd,omitempty"`
	RevenueSharePercent    *float64 `json:"revenue_share_percent,omitempty"`
	IsActive               bool     `json:"is_active"`
}

// ListingFunction creates a marketplace listing for a function
func (s *Service) ListingFunction(ctx context.Context, req *CreateFunctionListingRequest) (*identity.FunctionListing, error) {
	// Validate function exists
	var function identity.Function // Note: Using identity model or separate function registry model
	if err := s.db.WithContext(ctx).Where("id = ?", req.FunctionID).First(&function).Error; err != nil {
		return nil, fmt.Errorf("function not found: %w", err)
	}

	// Check if listing already exists
	var existing identity.FunctionListing
	err := s.db.WithContext(ctx).Where("function_id = ?", req.FunctionID).First(&existing).Error
	if err == nil {
		// Update existing listing
		existing.PricingModel = req.PricingModel
		existing.PricePerCall = req.PricePerCall
		existing.SubscriptionMonthlyUSD = req.SubscriptionMonthlyUSD
		existing.RevenueSharePercent = req.RevenueSharePercent
		existing.IsActive = req.IsActive
		existing.UpdatedAt = time.Now()
		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new listing
	listing := &identity.FunctionListing{
		ID:                     uuid.New(),
		FunctionID:             req.FunctionID,
		PricingModel:           req.PricingModel,
		PricePerCall:           req.PricePerCall,
		SubscriptionMonthlyUSD: req.SubscriptionMonthlyUSD,
		RevenueSharePercent:    req.RevenueSharePercent,
		IsActive:               req.IsActive,
		RatingScore:            0,
		CallVolume:             0,
		DeterministicVerified:  false,
		ListedAt:               time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(listing).Error; err != nil {
		return nil, err
	}

	return listing, nil
}

// CreateFunctionListingRequest represents a request to create a function listing
type CreateFunctionListingRequest struct {
	FunctionID             uuid.UUID
	PricingModel           string // free | per_call | subscription | revenue_share
	PricePerCall           *float64
	SubscriptionMonthlyUSD *float64
	RevenueSharePercent    *float64
	IsActive               bool
}

// SearchAgents searches for agents in the marketplace
func (s *Service) SearchAgents(ctx context.Context, req *SearchAgentsRequest) ([]AgentSearchResult, int64, error) {
	query := s.db.WithContext(ctx).Model(&identity.AgentListing{}).
		Where("is_active = ?", true).
		Preload("Agent")

	// Apply filters
	if len(req.ListingTypes) > 0 {
		query = query.Where("listing_type IN ?", req.ListingTypes)
	}

	if req.MinRating > 0 {
		query = query.Where("rating_score >= ?", req.MinRating)
	}

	if req.PricingModel != "" {
		query = query.Where("pricing_model = ?", req.PricingModel)
	}

	if req.MaxPricePerCall > 0 {
		query = query.Where("price_per_call <= ?", req.MaxPricePerCall)
	}

	if req.MinROIScore > 0 {
		query = query.Where("roi_score >= ?", req.MinROIScore)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		if isRelationNotFound(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	var listings []identity.AgentListing
	if err := query.Order("rating_score DESC").Limit(req.Limit).Offset(req.Offset).Find(&listings).Error; err != nil {
		if isRelationNotFound(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	// Calculate rankings (reliability from execution data when available)
	results := make([]AgentSearchResult, len(listings))
	for i, listing := range listings {
		reliability := s.agentExecutionSuccessRate(ctx, listing.AgentID)
		results[i] = AgentSearchResult{
			Listing:   listing,
			RankScore: CalculateAgentRankScore(listing, reliability),
		}
	}

	// Sort by rank score
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].RankScore > results[i].RankScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, total, nil
}

// SearchAgentsRequest represents a search request for agents
type SearchAgentsRequest struct {
	ListingTypes    []string // worker | manager | infrastructure
	PricingModel    string
	MinRating       float64
	MaxPricePerCall float64
	MinROIScore     float64
	Limit           int
	Offset          int
}

// AgentSearchResult represents an agent search result with ranking
type AgentSearchResult struct {
	Listing   identity.AgentListing
	RankScore float64
}

// SearchFunctions searches for functions in the marketplace
func (s *Service) SearchFunctions(ctx context.Context, req *SearchFunctionsRequest) ([]FunctionSearchResult, int64, error) {
	query := s.db.WithContext(ctx).Model(&identity.FunctionListing{}).
		Where("is_active = ?", true).
		Preload("Function")

	// Apply filters
	if req.DeterministicVerified {
		query = query.Where("deterministic_verified = ?", true)
	}

	if req.MinRating > 0 {
		query = query.Where("rating_score >= ?", req.MinRating)
	}

	if req.PricingModel != "" {
		query = query.Where("pricing_model = ?", req.PricingModel)
	}

	if req.MaxPricePerCall > 0 {
		query = query.Where("price_per_call <= ?", req.MaxPricePerCall)
	}

	if req.MinCallVolume > 0 {
		query = query.Where("call_volume >= ?", req.MinCallVolume)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var listings []identity.FunctionListing
	if err := query.Order("rating_score DESC").Limit(req.Limit).Offset(req.Offset).Find(&listings).Error; err != nil {
		return nil, 0, err
	}

	// Calculate rankings
	results := make([]FunctionSearchResult, len(listings))
	for i, listing := range listings {
		results[i] = FunctionSearchResult{
			Listing:   listing,
			RankScore: CalculateFunctionRankScore(listing),
		}
	}

	return results, total, nil
}

// SearchFunctionsRequest represents a search request for functions
type SearchFunctionsRequest struct {
	DeterministicVerified bool
	PricingModel          string
	MinRating             float64
	MaxPricePerCall       float64
	MinCallVolume         int
	Limit                 int
	Offset                int
}

// FunctionSearchResult represents a function search result with ranking
type FunctionSearchResult struct {
	Listing   identity.FunctionListing
	RankScore float64
}

// isRelationNotFound returns true if the error is a Postgres "relation does not exist" (42P01).
// Used to return empty results when agent_listings (or related) table has not been migrated yet.
func isRelationNotFound(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "42P01" // undefined_table
	}
	return false
}

// agentExecutionSuccessRate returns the agent's success rate (0-1) from recent execution records.
// Returns -1 if no execution data (caller should fall back to rating).
func (s *Service) agentExecutionSuccessRate(ctx context.Context, agentID string) float64 {
	var total, success int64
	since := time.Now().Add(-30 * 24 * time.Hour)
	s.db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Where("agent_id = ? AND timestamp > ?", agentID, since).
		Count(&total)
	if total == 0 {
		return -1
	}
	s.db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Where("agent_id = ? AND outcome = ? AND timestamp > ?", agentID, "success", since).
		Count(&success)
	return float64(success) / float64(total)
}

// CalculateAgentRankScore calculates the marketplace rank score for an agent.
// executionReliability is success rate 0-1 from execution data, or -1 to use listing rating as proxy.
func CalculateAgentRankScore(listing identity.AgentListing, executionReliability float64) float64 {
	trustScore := listing.RatingScore
	economicScore := listing.ROIScore
	reliability := listing.RatingScore
	if executionReliability >= 0 {
		reliability = executionReliability * 5 // scale 0-1 to 0-5 to match rating
	}
	roi := listing.ROIScore
	callVolumeLog := math.Log(float64(listing.TotalCalls + 1))

	// Normalize call volume (0-1 range)
	callVolumeNormalized := math.Min(callVolumeLog/10, 1.0)

	return (WeightTrustScore * trustScore) +
		(WeightEconomicScore * economicScore) +
		(WeightReliability * reliability) +
		(WeightROI * roi) +
		(WeightCallVolume * callVolumeNormalized)
}

// CalculateFunctionRankScore calculates the marketplace rank score for a function
func CalculateFunctionRankScore(listing identity.FunctionListing) float64 {
	rating := listing.RatingScore
	callVolumeLog := math.Log(float64(listing.CallVolume + 1))
	callVolumeNormalized := math.Min(callVolumeLog/10, 1.0)

	var deterministicBonus float64
	if listing.DeterministicVerified {
		deterministicBonus = 0.1
	}

	return (WeightTrustScore * rating) +
		(WeightReliability * rating) +
		(WeightCallVolume * callVolumeNormalized) +
		deterministicBonus
}

// IncrementCallCount increments the call count for an agent listing
func (s *Service) IncrementCallCount(ctx context.Context, agentID string) error {
	return s.db.WithContext(ctx).Model(&identity.AgentListing{}).
		Where("agent_id = ?", agentID).
		UpdateColumn("total_calls", gorm.Expr("total_calls + 1")).Error
}

// IncrementFunctionCallVolume increments the call volume for a function listing
func (s *Service) IncrementFunctionCallVolume(ctx context.Context, functionID uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&identity.FunctionListing{}).
		Where("function_id = ?", functionID).
		UpdateColumn("call_volume", gorm.Expr("call_volume + 1")).Error
}

// UpdateRating updates the rating score for a listing
func (s *Service) UpdateRating(ctx context.Context, listingType string, listingID uuid.UUID, newRating float64) error {
	if newRating < 0 || newRating > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}

	if listingType == "agent" {
		return s.db.WithContext(ctx).Model(&identity.AgentListing{}).
			Where("id = ?", listingID).
			Update("rating_score", newRating).Error
	}
	return s.db.WithContext(ctx).Model(&identity.FunctionListing{}).
		Where("id = ?", listingID).
		Update("rating_score", newRating).Error
}

// GetListing gets a marketplace listing by type and ID
func (s *Service) GetListing(ctx context.Context, listingType string, listingID uuid.UUID) (any, error) {
	if listingType == "agent" {
		var listing identity.AgentListing
		err := s.db.WithContext(ctx).Where("id = ?", listingID).First(&listing).Error
		return listing, err
	}
	var listing identity.FunctionListing
	err := s.db.WithContext(ctx).Where("id = ?", listingID).First(&listing).Error
	return listing, err
}

// DeactivateListing deactivates a marketplace listing
func (s *Service) DeactivateListing(ctx context.Context, listingType string, ownerAgentID string) error {
	if listingType == "agent" {
		return s.db.WithContext(ctx).Model(&identity.AgentListing{}).
			Where("agent_id = ?", ownerAgentID).
			Update("is_active", false).Error
	}
	return fmt.Errorf("unsupported listing type: %s", listingType)
}

// GetFunctionListingByURI retrieves a function listing by its author/name URI
func (s *Service) GetFunctionListingByURI(ctx context.Context, author, name string) (*identity.FunctionListing, error) {
	// First find the function in registry
	var function identity.Function
	if err := s.db.WithContext(ctx).
		Where("author = ? AND name = ?", author, name).
		First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("function not found: %s/%s", author, name)
		}
		return nil, fmt.Errorf("failed to find function: %w", err)
	}

	// Then find the marketplace listing for this function
	var listing identity.FunctionListing
	err := s.db.WithContext(ctx).
		Where("function_id = ? AND is_active = ?", function.ID, true).
		First(&listing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No active listing found, return a default listing based on function data
			return &identity.FunctionListing{
				FunctionID:   uuid.MustParse(function.ID),
				PricingModel: "per_call",
				PricePerCall: &function.PricePerCall,
				IsActive:     function.PricePerCall > 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to find listing: %w", err)
	}

	return &listing, nil
}
