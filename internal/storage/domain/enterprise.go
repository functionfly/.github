package domain

import (
	"context"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// TeamMemoryRepository handles Team Memory Engine operations
type TeamMemoryRepository interface {
	CreateTeamMemory(ctx context.Context, memory *types.TeamMemory) (*types.TeamMemory, error)
	GetTeamMemoryByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*types.TeamMemory, error)
	UpdateTeamMemory(ctx context.Context, memory *types.TeamMemory) (*types.TeamMemory, error)
	DeleteTeamMemory(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error
	ListTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, filter types.TeamMemoryFilter) ([]*types.TeamMemory, int64, error)
	ListTeamMemoriesByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*types.TeamMemory, error)
	SearchTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, query string, limit int) ([]*types.TeamMemorySearchResult, error)
	SearchTeamMemoriesByVector(ctx context.Context, tenantID, teamID uuid.UUID, embedding []float32, limit int) ([]*types.TeamMemorySearchResult, error)
	ValidateTeamMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error
	MarkTeamMemoryAsAccessed(ctx context.Context, memoryID uuid.UUID) error
	CreateEncryptedTeamMemory(ctx context.Context, memory *types.TeamMemory, encryptedContent, iv, tag []byte) (*types.TeamMemory, error)
	GetTeamMemoryDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error)
}

// MemoryExtractionRepository handles memory extraction queue
type MemoryExtractionRepository interface {
	CreateMemoryExtraction(ctx context.Context, extraction *types.MemoryExtraction) (*types.MemoryExtraction, error)
	GetMemoryExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*types.MemoryExtraction, error)
	ApproveMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*types.TeamMemory, error)
	RejectMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error
	ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error)
}

// MemoryShareRepository handles cross-team memory sharing
type MemoryShareRepository interface {
	CreateMemoryShare(ctx context.Context, share *types.MemoryShare) error
	GetMemoryShareByID(ctx context.Context, shareID uuid.UUID) (*types.MemoryShare, error)
	GetMemoryShareBetweenTeams(ctx context.Context, memoryID, sourceTeamID, targetTeamID uuid.UUID) (*types.MemoryShare, error)
	UpdateMemoryShare(ctx context.Context, share *types.MemoryShare) error
	ListMemorySharesByTargetTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*types.MemoryShare, error)
	ListMemorySharesBySourceTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*types.MemoryShare, error)
	ListMemorySharesByMemoryID(ctx context.Context, memoryID uuid.UUID, status string) ([]*types.MemoryShare, error)
}

// BundlePricingRepository handles Backend-in-a-Box pricing bundles
type BundlePricingRepository interface {
	CreatePricingBundle(ctx context.Context, bundle *types.PricingBundle) (*types.PricingBundle, error)
	ListPricingBundles(ctx context.Context, activeOnly bool) ([]*types.PricingBundle, error)
	GetPricingBundleBySlug(ctx context.Context, slug string) (*types.PricingBundle, error)
	GetPricingBundleByID(ctx context.Context, id uuid.UUID) (*types.PricingBundle, error)
	GetPricingBundleByStripePriceID(ctx context.Context, stripePriceID string) (*types.PricingBundle, error)
	UpdatePricingBundleStripePrice(ctx context.Context, slug, stripePriceID string) error
	CountActiveFounderModeRegistrations(ctx context.Context) (int, error)
	CountRecentSuccessfulDeployments(ctx context.Context) (int, error)
}

// FounderModeRepository handles Founder Mode (viral pricing)
type FounderModeRepository interface {
	CreateFounderModeRegistration(ctx context.Context, reg *types.FounderModeRegistration) error
	GetActiveFounderMode(ctx context.Context, tenantID, bundleID uuid.UUID) (*types.FounderModeRegistration, error)
	ListFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.FounderModeRegistration, error)
	ListActiveFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.FounderModeRegistration, error)
	UpdateFounderModeStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateFounderModeProgress(ctx context.Context, id uuid.UUID, users, mrrCents, apiCalls int) error
	ListAllActiveFounderModes(ctx context.Context) ([]*types.FounderModeRegistration, error)
	StartGracePeriod(ctx context.Context, id uuid.UUID, gracePeriodDays int) error
	GetDeferredBillingConfig(ctx context.Context, bundleID uuid.UUID) (*types.DeferredBillingConfig, error)
}

// BundleSubscriptionRepository handles bundle subscriptions
type BundleSubscriptionRepository interface {
	CreateBundleSubscription(ctx context.Context, sub *types.BundleSubscription) error
	UpdateBundleSubscription(ctx context.Context, sub *types.BundleSubscription) error
	GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*types.BundleSubscription, error)
	GetBundleSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*types.BundleSubscription, error)
	ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.BundleSubscription, error)
}

// MaintenanceRepository handles maintenance mode and system settings
type MaintenanceRepository interface {
	IsMaintenanceMode(ctx context.Context) (bool, error)
	GetMaintenanceMessage(ctx context.Context) (string, error)
	SetMaintenanceMode(ctx context.Context, enabled bool, message string) error
}

// FeatureMeasureRepository handles platform feature flags
type FeatureMeasureRepository interface {
	ListFeatureMeasures(ctx context.Context) ([]*types.FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
}
