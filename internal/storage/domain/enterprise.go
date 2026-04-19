package domain

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// TeamMemoryRepository handles Team Memory Engine operations
type TeamMemoryRepository interface {
	CreateTeamMemory(ctx context.Context, memory *storage.TeamMemory) (*storage.TeamMemory, error)
	GetTeamMemoryByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*storage.TeamMemory, error)
	UpdateTeamMemory(ctx context.Context, memory *storage.TeamMemory) (*storage.TeamMemory, error)
	DeleteTeamMemory(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error
	ListTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, filter storage.TeamMemoryFilter) ([]*storage.TeamMemory, int64, error)
	ListTeamMemoriesByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*storage.TeamMemory, error)
	SearchTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, query string, limit int) ([]*storage.TeamMemorySearchResult, error)
	SearchTeamMemoriesByVector(ctx context.Context, tenantID, teamID uuid.UUID, embedding []float32, limit int) ([]*storage.TeamMemorySearchResult, error)
	ValidateTeamMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error
	MarkTeamMemoryAsAccessed(ctx context.Context, memoryID uuid.UUID) error
	CreateEncryptedTeamMemory(ctx context.Context, memory *storage.TeamMemory, encryptedContent, iv, tag []byte) (*storage.TeamMemory, error)
	GetTeamMemoryDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error)
}

// MemoryExtractionRepository handles memory extraction queue
type MemoryExtractionRepository interface {
	CreateMemoryExtraction(ctx context.Context, extraction *storage.MemoryExtraction) (*storage.MemoryExtraction, error)
	GetMemoryExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*storage.MemoryExtraction, error)
	ApproveMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*storage.TeamMemory, error)
	RejectMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error
	ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error)
}

// MemoryShareRepository handles cross-team memory sharing
type MemoryShareRepository interface {
	CreateMemoryShare(ctx context.Context, share *storage.MemoryShare) error
	GetMemoryShareByID(ctx context.Context, shareID uuid.UUID) (*storage.MemoryShare, error)
	GetMemoryShareBetweenTeams(ctx context.Context, memoryID, sourceTeamID, targetTeamID uuid.UUID) (*storage.MemoryShare, error)
	UpdateMemoryShare(ctx context.Context, share *storage.MemoryShare) error
	ListMemorySharesByTargetTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*storage.MemoryShare, error)
	ListMemorySharesBySourceTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*storage.MemoryShare, error)
	ListMemorySharesByMemoryID(ctx context.Context, memoryID uuid.UUID, status string) ([]*storage.MemoryShare, error)
}

// BundlePricingRepository handles Backend-in-a-Box pricing bundles
type BundlePricingRepository interface {
	CreatePricingBundle(ctx context.Context, bundle *storage.PricingBundle) (*storage.PricingBundle, error)
	ListPricingBundles(ctx context.Context, activeOnly bool) ([]*storage.PricingBundle, error)
	GetPricingBundleBySlug(ctx context.Context, slug string) (*storage.PricingBundle, error)
	GetPricingBundleByID(ctx context.Context, id uuid.UUID) (*storage.PricingBundle, error)
}

// FounderModeRepository handles Founder Mode (viral pricing)
type FounderModeRepository interface {
	CreateFounderModeRegistration(ctx context.Context, reg *storage.FounderModeRegistration) error
	GetActiveFounderMode(ctx context.Context, tenantID, bundleID uuid.UUID) (*storage.FounderModeRegistration, error)
	ListFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.FounderModeRegistration, error)
	ListActiveFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.FounderModeRegistration, error)
	UpdateFounderModeStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateFounderModeProgress(ctx context.Context, id uuid.UUID, users, mrrCents, apiCalls int) error
	ListAllActiveFounderModes(ctx context.Context) ([]*storage.FounderModeRegistration, error)
	StartGracePeriod(ctx context.Context, id uuid.UUID, gracePeriodDays int) error
	GetDeferredBillingConfig(ctx context.Context, bundleID uuid.UUID) (*storage.DeferredBillingConfig, error)
}

// BundleSubscriptionRepository handles bundle subscriptions
type BundleSubscriptionRepository interface {
	CreateBundleSubscription(ctx context.Context, sub *storage.BundleSubscription) error
	UpdateBundleSubscription(ctx context.Context, sub *storage.BundleSubscription) error
	GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*storage.BundleSubscription, error)
	ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.BundleSubscription, error)
}

// MaintenanceRepository handles maintenance mode and system settings
type MaintenanceRepository interface {
	IsMaintenanceMode() (bool, error)
	GetMaintenanceMessage() (string, error)
	SetMaintenanceMode(enabled bool, message string) error
}

// FeatureMeasureRepository handles platform feature flags
type FeatureMeasureRepository interface {
	ListFeatureMeasures(ctx context.Context) ([]*storage.FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
}
