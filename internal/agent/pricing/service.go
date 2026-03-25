package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StrategyType defines the type of pricing strategy
type StrategyType string

const (
	StrategyTypeFixed    StrategyType = "fixed"
	StrategyTypeTiered  StrategyType = "tiered"
	StrategyTypeDynamic StrategyType = "dynamic"
	StrategyTypeAuction StrategyType = "auction"
)

// Service handles pricing calculations
type Service struct {
	db *gorm.DB
}

// NewService creates a new pricing service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CalculatePrice calculates the price for a function call based on its pricing model
func (s *Service) CalculatePrice(ctx context.Context, listing *identity.FunctionListing, callsInPeriod int) (float64, error) {
	switch listing.PricingModel {
	case "free":
		return 0, nil
	case "per_call":
		if listing.PricePerCall == nil {
			return 0, fmt.Errorf("price per call not set")
		}
		return *listing.PricePerCall, nil
	case "subscription":
		// Monthly subscription - divide by estimated calls
		if listing.SubscriptionMonthlyUSD == nil {
			return 0, fmt.Errorf("subscription price not set")
		}
		// Assume 1000 calls per month for subscription calculation
		estimatedCalls := 1000
		if callsInPeriod > 0 {
			estimatedCalls = callsInPeriod
		}
		return *listing.SubscriptionMonthlyUSD / float64(estimatedCalls), nil
	case "tiered":
		return s.calculateTieredPrice(listing, callsInPeriod)
	case "dynamic":
		return s.calculateDynamicPrice(listing, callsInPeriod)
	case "auction":
		return s.calculateAuctionPrice(listing)
	default:
		if listing.PricePerCall != nil {
			return *listing.PricePerCall, nil
		}
		return 0, fmt.Errorf("unsupported pricing model: %s", listing.PricingModel)
	}
}

// calculateTieredPrice calculates price based on volume tiers
func (s *Service) calculateTieredPrice(listing *identity.FunctionListing, callsInPeriod int) (float64, error) {
	if len(listing.Tiers) == 0 {
		// Fall back to per-call price if no tiers
		if listing.PricePerCall == nil {
			return 0, fmt.Errorf("no tiers and no fallback price")
		}
		return *listing.PricePerCall, nil
	}

	// Find the applicable tier based on call volume
	var applicableTier *identity.PricingTier
	for i := range listing.Tiers {
		tier := &listing.Tiers[i]
		if callsInPeriod >= tier.CallsPerMonth {
			applicableTier = tier
		} else {
			break
		}
	}

	if applicableTier == nil {
		// Below first tier, use base price
		if listing.PricePerCall == nil {
			return 0, fmt.Errorf("no applicable tier and no base price")
		}
		return *listing.PricePerCall, nil
	}

	return applicableTier.PricePerCall, nil
}

// calculateDynamicPrice calculates price based on demand factors
func (s *Service) calculateDynamicPrice(listing *identity.FunctionListing, callsInPeriod int) (float64, error) {
	basePrice := 0.001 // Default base price
	if listing.PricePerCall != nil {
		basePrice = *listing.PricePerCall
	}

	minPrice := 0.0001
	maxPrice := 0.01
	demandFactor := 1.0

	if listing.DynamicMinPrice != nil {
		minPrice = *listing.DynamicMinPrice
	}
	if listing.DynamicMaxPrice != nil {
		maxPrice = *listing.DynamicMaxPrice
	}
	if listing.DynamicDemandFactor != nil {
		demandFactor = *listing.DynamicDemandFactor
	}

	// Simple demand-based pricing: increase price when demand (calls) is high
	// Using a sigmoid-like function to smoothly adjust price
	demandRatio := float64(callsInPeriod) / 1000.0 // Normalize to 1000 calls
	priceMultiplier := 1.0 + (demandFactor * math.Tanh(demandRatio-0.5))

	calculatedPrice := basePrice * priceMultiplier

	// Clamp to min/max bounds
	if calculatedPrice < minPrice {
		return minPrice, nil
	}
	if calculatedPrice > maxPrice {
		return maxPrice, nil
	}

	return calculatedPrice, nil
}

// calculateAuctionPrice returns the current auction bid price
func (s *Service) calculateAuctionPrice(listing *identity.FunctionListing) (float64, error) {
	if listing.AuctionCurrentBid != nil {
		return *listing.AuctionCurrentBid, nil
	}
	if listing.AuctionStartPrice != nil {
		return *listing.AuctionStartPrice, nil
	}
	return 0, fmt.Errorf("no auction price set")
}

// PlaceBid places a bid on an auction listing
func (s *Service) PlaceBid(ctx context.Context, listingID uuid.UUID, bidderID string, amount float64) error {
	var listing identity.FunctionListing
	if err := s.db.WithContext(ctx).First(&listing, listingID).Error; err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	if listing.PricingModel != "auction" {
		return fmt.Errorf("listing is not an auction")
	}

	// Check if auction has ended
	if listing.AuctionEndTime != nil && time.Now().After(*listing.AuctionEndTime) {
		return fmt.Errorf("auction has ended")
	}

	// Check if bid is higher than current
	if listing.AuctionCurrentBid != nil && amount <= *listing.AuctionCurrentBid {
		return fmt.Errorf("bid must be higher than current bid: %.6f", *listing.AuctionCurrentBid)
	}

	// Update listing with new bid
	listing.AuctionCurrentBid = &amount
	listing.AuctionBidCount++

	if err := s.db.WithContext(ctx).Save(&listing).Error; err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}

	return nil
}

// StartAuction starts an auction for a listing
func (s *Service) StartAuction(ctx context.Context, listingID uuid.UUID, startPrice float64, reservePrice *float64, duration time.Duration) error {
	var listing identity.FunctionListing
	if err := s.db.WithContext(ctx).First(&listing, listingID).Error; err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	endTime := time.Now().Add(duration)
	listing.PricingModel = "auction"
	listing.AuctionStartPrice = &startPrice
	listing.AuctionReservePrice = reservePrice
	listing.AuctionEndTime = &endTime
	listing.AuctionCurrentBid = nil
	listing.AuctionBidCount = 0

	if err := s.db.WithContext(ctx).Save(&listing).Error; err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}

	return nil
}

// EndAuction ends an auction and determines the winner
func (s *Service) EndAuction(ctx context.Context, listingID uuid.UUID) (winnerID string, finalPrice float64, err error) {
	var listing identity.FunctionListing
	if err := s.db.WithContext(ctx).First(&listing, listingID).Error; err != nil {
		return "", 0, fmt.Errorf("listing not found: %w", err)
	}

	if listing.PricingModel != "auction" {
		return "", 0, fmt.Errorf("listing is not an auction")
	}

	if listing.AuctionEndTime == nil {
		return "", 0, fmt.Errorf("auction has no end time")
	}

	if !time.Now().After(*listing.AuctionEndTime) {
		return "", 0, fmt.Errorf("auction has not ended yet")
	}

	// Check if reserve was met
	if listing.AuctionReservePrice != nil && listing.AuctionCurrentBid != nil {
		if *listing.AuctionCurrentBid < *listing.AuctionReservePrice {
			// Reserve not met - no winner
			return "", 0, nil
		}
	}

	if listing.AuctionCurrentBid != nil {
		return "", *listing.AuctionCurrentBid, nil
	}

	return "", 0, nil
}

// GetTieredPriceSummary returns a summary of tier pricing
func (s *Service) GetTieredPriceSummary(tiers []identity.PricingTier) string {
	if len(tiers) == 0 {
		return "No tiers configured"
	}

	summary := "Volume Tiers:\n"
	for _, tier := range tiers {
		summary += fmt.Sprintf("  %d+ calls: $%.6f/call (%.0f%% discount)\n",
			tier.CallsPerMonth, tier.PricePerCall, tier.DiscountPct)
	}
	return summary
}

// MarshalTiers converts tiers to JSON for storage
func MarshalTiers(tiers []identity.PricingTier) (string, error) {
	if len(tiers) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(tiers)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalTiers converts JSON to tiers
func UnmarshalTiers(data string) ([]identity.PricingTier, error) {
	if data == "" || data == "[]" {
		return []identity.PricingTier{}, nil
	}
	var tiers []identity.PricingTier
	if err := json.Unmarshal([]byte(data), &tiers); err != nil {
		return nil, err
	}
	return tiers, nil
}
