// Package pricing provides database-driven pricing configuration
// This replaces the hardcoded constants from internal/plans/limits.go
package pricing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
)

// TierCacheTTL is the cache time-to-live for pricing data
const TierCacheTTL = 5 * time.Minute

// Service provides database-driven tier pricing
// This replaces hardcoded constants and allows runtime price changes without deployment
type Service struct {
	repo  storage.Repository
	cache *tierCache
}

// tierCache holds cached tier pricing data
type tierCache struct {
	mu         sync.RWMutex
	tiers      map[string]*storage.AgentTierPricing
	lastUpdate time.Time
	ttl        time.Duration
}

// NewService creates a new pricing service
func NewService(repo storage.Repository) *Service {
	return &Service{
		repo: repo,
		cache: &tierCache{
			tiers: make(map[string]*storage.AgentTierPricing),
			ttl:   TierCacheTTL,
		},
	}
}

// GetTierPricing retrieves tier pricing by slug
// Supported slugs: 'agent-starter', 'agent-scale', 'agent-pro', 'agent-enterprise'
func (s *Service) GetTierPricing(ctx context.Context, slug string) (*storage.AgentTierPricing, error) {
	// Check cache first
	if tier := s.cache.get(slug); tier != nil {
		return tier, nil
	}

	// Load from database
	tier, err := s.repo.GetAgentTierPricingBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing: %w", err)
	}

	if tier == nil {
		// Fall back to hardcoded constants for backward compatibility
		// This ensures the system works even if database is not yet populated
		return s.getFallbackPricing(slug)
	}

	// Update cache
	s.cache.set(slug, tier)
	return tier, nil
}

// GetTierPricingForRegion retrieves tier pricing for a specific currency/region
func (s *Service) GetTierPricingForRegion(ctx context.Context, slug string, currencyCode string) (*storage.AgentTierPricing, error) {
	// Check cache for region-specific key
	cacheKey := slug + ":" + currencyCode
	if tier := s.cache.get(cacheKey); tier != nil {
		return tier, nil
	}

	// Load from database with region pricing
	tier, err := s.repo.GetAgentTierPricingForRegion(ctx, slug, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing for region: %w", err)
	}

	if tier == nil {
		return s.getFallbackPricing(slug)
	}

	s.cache.set(cacheKey, tier)
	return tier, nil
}

// ListAllTiers retrieves all active tier pricing configurations
func (s *Service) ListAllTiers(ctx context.Context) ([]*storage.AgentTierPricing, error) {
	return s.repo.ListAgentTierPricing(ctx, true)
}

// InvalidateCache clears the pricing cache
func (s *Service) InvalidateCache() {
	s.cache.clear()
}

// GetMonthlyPrice returns the monthly price for a tier in the specified currency
// This method provides a direct replacement for the hardcoded AgentStarterPriceCents etc.
func (s *Service) GetMonthlyPrice(ctx context.Context, slug string, currencyCode string) (int, error) {
	tier, err := s.GetTierPricingForRegion(ctx, slug, currencyCode)
	if err != nil {
		// Fall back to hardcoded constants
		return s.getFallbackMonthlyPrice(slug), nil
	}
	return tier.GetMonthlyPrice(currencyCode), nil
}

// GetAnnualPrice returns the annual price for a tier in the specified currency
func (s *Service) GetAnnualPrice(ctx context.Context, slug string, currencyCode string) (int, error) {
	tier, err := s.GetTierPricingForRegion(ctx, slug, currencyCode)
	if err != nil {
		return s.getFallbackAnnualPrice(slug), nil
	}
	annualPrice := tier.GetAnnualPrice(currencyCode)
	if annualPrice == nil {
		return 0, fmt.Errorf("no annual price available for tier %s", slug)
	}
	return *annualPrice, nil
}

// getFallbackPricing returns hardcoded pricing for backward compatibility
// Agent tiers now map to unified plans: agent_starter→starter, agent_scale→professional, agent_pro→enterprise
func (s *Service) getFallbackPricing(slug string) (*storage.AgentTierPricing, error) {
	switch slug {
	case plans.PlanAgentStarter:
		monthly := plans.StarterPriceCents
		annual := plans.StarterAnnualCents
		return &storage.AgentTierPricing{
			TierSlug:                 slug,
			DisplayName:              "Agent Starter",
			MonthlyPriceCents:        monthly,
			AnnualPriceCents:         &annual,
			BaseCurrency:             "USD",
			MaxAgents:                5,
			IncludedAICalls:          10000,
			IncludedExecutions:       100000,
			IncludedStorageGB:        10,
			OveragePricePer1000Cents: 50,
			IsActive:                 true,
		}, nil
	case plans.PlanAgentScale:
		monthly := plans.ProPriceCents
		annual := plans.ProAnnualCents
		return &storage.AgentTierPricing{
			TierSlug:                 slug,
			DisplayName:              "Agent Scale",
			MonthlyPriceCents:        monthly,
			AnnualPriceCents:         &annual,
			BaseCurrency:             "USD",
			MaxAgents:                25,
			IncludedAICalls:          100000,
			IncludedExecutions:       1000000,
			IncludedStorageGB:        100,
			OveragePricePer1000Cents: 40,
			IsActive:                 true,
		}, nil
	case plans.PlanAgentPro:
		monthly := plans.EnterprisePriceCents
		annual := plans.EnterpriseAnnualCents
		return &storage.AgentTierPricing{
			TierSlug:                 slug,
			DisplayName:              "Agent Pro",
			MonthlyPriceCents:        monthly,
			AnnualPriceCents:         &annual,
			BaseCurrency:             "USD",
			MaxAgents:                100,
			IncludedAICalls:          500000,
			IncludedExecutions:       5000000,
			IncludedStorageGB:        500,
			OveragePricePer1000Cents: 30,
			IsActive:                 true,
		}, nil
	case plans.PlanAgentEnterprise:
		monthly := plans.AgentEnterprisePriceCents
		annual := plans.AgentEnterpriseAnnualCents
		return &storage.AgentTierPricing{
			TierSlug:                 slug,
			DisplayName:              "Agent Enterprise",
			MonthlyPriceCents:        monthly,
			AnnualPriceCents:         &annual,
			BaseCurrency:             "USD",
			MaxAgents:                -1, // Unlimited
			IncludedAICalls:          -1,
			IncludedExecutions:       -1,
			IncludedStorageGB:        -1,
			OveragePricePer1000Cents: 20,
			IsActive:                 true,
		}, nil
	default:
		return nil, fmt.Errorf("unknown tier slug: %s", slug)
	}
}

// getFallbackMonthlyPrice returns hardcoded monthly prices for backward compatibility
// Maps agent tier slugs to unified plan prices
func (s *Service) getFallbackMonthlyPrice(slug string) int {
	switch slug {
	case plans.PlanAgentStarter:
		return plans.StarterPriceCents
	case plans.PlanAgentScale:
		return plans.ProPriceCents
	case plans.PlanAgentPro:
		return plans.EnterprisePriceCents
	case plans.PlanAgentEnterprise:
		return plans.AgentEnterprisePriceCents
	default:
		return 0
	}
}

// getFallbackAnnualPrice returns hardcoded annual prices (monthly * 10 for 2 months free)
func (s *Service) getFallbackAnnualPrice(slug string) int {
	monthly := s.getFallbackMonthlyPrice(slug)
	if monthly == 0 {
		return 0
	}
	return monthly * 10
}

// Tier slugs - use these constants when calling the service
const (
	TierAgentStarter    = "agent-starter"
	TierAgentScale      = "agent-scale"
	TierAgentPro        = "agent-pro"
	TierAgentEnterprise = "agent-enterprise"
)

// cache methods

func (c *tierCache) get(key string) *storage.AgentTierPricing {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.lastUpdate) > c.ttl {
		return nil // Cache expired
	}

	return c.tiers[key]
}

func (c *tierCache) set(key string, tier *storage.AgentTierPricing) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tiers[key] = tier
	c.lastUpdate = time.Now()
}

func (c *tierCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tiers = make(map[string]*storage.AgentTierPricing)
	c.lastUpdate = time.Time{}
}
