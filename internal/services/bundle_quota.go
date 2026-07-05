package services

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BundleQuotaLimitType string

const (
	BundleLimitFunctions BundleQuotaLimitType = "functions"
	BundleLimitProviders BundleQuotaLimitType = "providers"
	BundleLimitAICalls   BundleQuotaLimitType = "ai_calls"
	BundleLimitRequests  BundleQuotaLimitType = "requests"
	BundleLimitWorkflows BundleQuotaLimitType = "workflows"
	BundleLimitStorageGB BundleQuotaLimitType = "storage_gb"
)

type BundleQuotaResult struct {
	Allowed   bool
	LimitType BundleQuotaLimitType
	Current   int
	Limit     int
	Message   string
}

type BundleQuotaStatus struct {
	TenantID uuid.UUID

	FunctionsUsed  int
	FunctionsLimit int
	FunctionsOK    bool

	ProvidersUsed  int
	ProvidersLimit int
	ProvidersOK    bool

	AICallsUsed  int
	AICallsLimit int
	AICallsOK    bool

	RequestsUsed  int
	RequestsLimit int
	RequestsOK    bool

	WorkflowsUsed  int
	WorkflowsLimit int
	WorkflowsOK    bool

	StorageGBUsed  int
	StorageGBLimit int
	StorageGBOK    bool
}

type BundleQuotaService struct {
	repo   storage.Repository
	logger *logrus.Logger
}

func NewBundleQuotaService(repo storage.Repository) *BundleQuotaService {
	return &BundleQuotaService{
		repo:   repo,
		logger: logrus.StandardLogger(),
	}
}

func (s *BundleQuotaService) getBundleSubscriptionAndLimits(ctx context.Context, tenantID uuid.UUID) (*storage.BundleSubscription, *storage.PricingBundle, error) {
	sub, err := s.repo.GetBundleSubscriptionByTenant(ctx, tenantID)
	if err != nil || sub == nil {
		return nil, nil, err
	}

	bundle, err := s.repo.GetPricingBundleByID(ctx, sub.BundleID)
	if err != nil || bundle == nil {
		return sub, nil, err
	}

	return sub, bundle, nil
}

func (s *BundleQuotaService) CheckFunctionLimit(ctx context.Context, tenantID uuid.UUID) (*BundleQuotaResult, error) {
	_, bundle, err := s.getBundleSubscriptionAndLimits(ctx, tenantID)
	if err != nil || bundle == nil {
		return &BundleQuotaResult{Allowed: true}, err
	}

	limit, exists := bundle.FeatureLimits["functions"]
	if !exists || limit <= 0 {
		return &BundleQuotaResult{Allowed: true}, nil
	}

	functions, err := s.repo.ListFunctionsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	current := len(functions)
	allowed := current < limit

	result := &BundleQuotaResult{
		Allowed:   allowed,
		LimitType: BundleLimitFunctions,
		Current:   current,
		Limit:     limit,
	}

	if !allowed {
		result.Message = "Function limit reached. Upgrade your bundle to create more functions."
	}

	return result, nil
}

func (s *BundleQuotaService) CheckProviderLimit(ctx context.Context, tenantID uuid.UUID) (*BundleQuotaResult, error) {
	_, bundle, err := s.getBundleSubscriptionAndLimits(ctx, tenantID)
	if err != nil || bundle == nil {
		return &BundleQuotaResult{Allowed: true}, err
	}

	limit, exists := bundle.FeatureLimits["providers"]
	if !exists || limit <= 0 {
		return &BundleQuotaResult{Allowed: true}, nil
	}

	current, err := s.repo.CountBackendsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	allowed := current < limit

	result := &BundleQuotaResult{
		Allowed:   allowed,
		LimitType: BundleLimitProviders,
		Current:   current,
		Limit:     limit,
	}

	if !allowed {
		result.Message = "Provider limit reached. Upgrade your bundle to add more providers."
	}

	return result, nil
}

func (s *BundleQuotaService) CheckAICallsLimit(ctx context.Context, tenantID uuid.UUID, callsToAdd int) (*BundleQuotaResult, error) {
	_, bundle, err := s.getBundleSubscriptionAndLimits(ctx, tenantID)
	if err != nil || bundle == nil {
		return &BundleQuotaResult{Allowed: true}, err
	}

	limit, exists := bundle.FeatureLimits["ai_calls"]
	if !exists || limit <= 0 {
		return &BundleQuotaResult{Allowed: true}, nil
	}

	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	rollups, err := s.repo.GetUsageByTenant(ctx, tenantID, "ai_call", periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	current := 0
	for _, rollup := range rollups {
		current += rollup.TotalQuantity
	}

	allowed := (current + callsToAdd) <= limit

	result := &BundleQuotaResult{
		Allowed:   allowed,
		LimitType: BundleLimitAICalls,
		Current:   current,
		Limit:     limit,
	}

	if !allowed {
		result.Message = "AI calls limit reached. Upgrade your bundle for more AI calls."
	}

	return result, nil
}

func (s *BundleQuotaService) GetQuotaStatus(ctx context.Context, tenantID uuid.UUID) (*BundleQuotaStatus, error) {
	status := &BundleQuotaStatus{
		TenantID: tenantID,
	}

	sub, bundle, err := s.getBundleSubscriptionAndLimits(ctx, tenantID)
	if err != nil || sub == nil || bundle == nil {
		status.FunctionsOK = true
		status.ProvidersOK = true
		status.AICallsOK = true
		status.RequestsOK = true
		status.WorkflowsOK = true
		status.StorageGBOK = true
		return status, nil
	}

	if limit, exists := bundle.FeatureLimits["functions"]; exists && limit > 0 {
		status.FunctionsLimit = limit
		functions, err := s.repo.ListFunctionsByTenant(ctx, tenantID)
		if err == nil {
			status.FunctionsUsed = len(functions)
		}
		status.FunctionsOK = status.FunctionsUsed < limit
	}

	if limit, exists := bundle.FeatureLimits["providers"]; exists && limit > 0 {
		status.ProvidersLimit = limit
		count, err := s.repo.CountBackendsByTenant(ctx, tenantID)
		if err == nil {
			status.ProvidersUsed = count
		}
		status.ProvidersOK = status.ProvidersUsed < limit
	}

	if limit, exists := bundle.FeatureLimits["ai_calls"]; exists && limit > 0 {
		status.AICallsLimit = limit
		now := time.Now().UTC()
		periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)
		rollups, err := s.repo.GetUsageByTenant(ctx, tenantID, "ai_call", periodStart, periodEnd)
		if err == nil {
			for _, rollup := range rollups {
				status.AICallsUsed += rollup.TotalQuantity
			}
		}
		status.AICallsOK = status.AICallsUsed < limit
	}

	if limit, exists := bundle.FeatureLimits["requests"]; exists && limit > 0 {
		status.RequestsLimit = limit
		now := time.Now().UTC()
		periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)
		rollups, err := s.repo.GetUsageByTenant(ctx, tenantID, "function_execution", periodStart, periodEnd)
		if err == nil {
			for _, rollup := range rollups {
				status.RequestsUsed += rollup.TotalQuantity
			}
		}
		status.RequestsOK = status.RequestsUsed < limit
	}

	if limit, exists := bundle.FeatureLimits["workflows"]; exists && limit > 0 {
		status.WorkflowsLimit = limit
	}

	if limit, exists := bundle.FeatureLimits["storage_gb"]; exists && limit > 0 {
		status.StorageGBLimit = limit
	}

	return status, nil
}

func (s *BundleQuotaService) HasBundleQuota(ctx context.Context, tenantID uuid.UUID) bool {
	sub, _, err := s.getBundleSubscriptionAndLimits(ctx, tenantID)
	if err != nil || sub == nil {
		return false
	}
	return sub.Status == "active" || sub.Status == "deferred"
}
