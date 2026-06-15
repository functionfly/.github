package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// UsageForecastRepository handles usage forecasting and alerting
type UsageForecastRepository interface {
	// Usage alerts
	CreateUsageAlert(ctx context.Context, alert *types.UsageAlert) error
	GetUsageAlertByID(ctx context.Context, id uuid.UUID) (*types.UsageAlert, error)
	ListUsageAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.UsageAlert, error)
	UpdateUsageAlert(ctx context.Context, alert *types.UsageAlert) error
	DeleteUsageAlert(ctx context.Context, id uuid.UUID) error
	RecordAlertTrigger(ctx context.Context, history *types.UsageAlertHistory) error
	GetAlertHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*types.UsageAlertHistory, error)

	// Spend caps
	CreateOrUpdateSpendCap(ctx context.Context, cap *types.SpendCap) error
	GetSpendCapByTenant(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*types.SpendCap, error)
	UpdateCurrentSpend(ctx context.Context, capID uuid.UUID, spendCents int) error

	// Forecasts
	SaveUsageForecast(ctx context.Context, forecast *types.UsageForecast) error
	GetLatestForecast(ctx context.Context, tenantID uuid.UUID, forecastType string) (*types.UsageForecast, error)
	GetDailyUsageHistory(ctx context.Context, tenantID uuid.UUID, eventType string, days int) ([]*types.DailyUsagePoint, error)
	GetDailySpendHistory(ctx context.Context, tenantID uuid.UUID, days int) ([]*types.DailyUsagePoint, error)
	GetCurrentPeriodUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*types.UsageSummary, error)
}

// CostAllocationRepository handles detailed cost tracking
type CostAllocationRepository interface {
	RecordCostAllocationEntry(ctx context.Context, entry *types.CostAllocationEntry) error
	GetCostAllocationByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*types.CostAllocationSummary, error)
	GetCostAllocationDailyBreakdown(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*types.DailyCostBreakdown, error)
	GetTenantCostSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*types.TenantCostSummary, error)
	GetAllTenantsCostSummary(ctx context.Context, start, end time.Time) ([]*types.TenantCostSummary, error)
	GetCostAllocationEntries(ctx context.Context, filter *types.CostAllocationFilter, limit, offset int) ([]*types.CostAllocationEntry, int, error)
	GetCostAllocationReport(ctx context.Context, start, end time.Time) (*types.CostAllocationReport, error)
	GetCostAllocationByRegion(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (map[string]*types.CostAllocationSummary, error)
	DeleteOldCostAllocationEntries(ctx context.Context, before time.Time) (int64, error)
}

// ExecutionRetentionRepository handles execution log retention
type ExecutionRetentionRepository interface {
	DeleteOldExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldPublicExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldResourceUsage(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldMEGRecords(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldDriftReports(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldExecutionCertificates(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	GetExecutionRetentionStats(ctx context.Context) (map[string]interface{}, error)

	// Settings
	GetExecutionRetentionSettings(ctx context.Context) (*types.ExecutionRetentionSettings, error)
	UpdateExecutionRetentionSettings(ctx context.Context, updates *types.ExecutionRetentionSettingsUpdate) (*types.ExecutionRetentionSettings, error)
	GetOrCreateExecutionRetentionSettings(ctx context.Context) (*types.ExecutionRetentionSettings, error)
	ResetExecutionRetentionSettingsToDefaults(ctx context.Context, updatedBy *uuid.UUID) (*types.ExecutionRetentionSettings, error)
}

// UsageExportRepository handles usage data export operations
type UsageExportRepository interface {
	// Configurations
	CreateUsageExportConfiguration(ctx context.Context, config *types.UsageExportConfiguration) error
	GetUsageExportConfiguration(ctx context.Context, id uuid.UUID) (*types.UsageExportConfiguration, error)
	UpdateUsageExportConfiguration(ctx context.Context, config *types.UsageExportConfiguration) error
	DeleteUsageExportConfiguration(ctx context.Context, id uuid.UUID) error
	ListUsageExportConfigurations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.UsageExportConfiguration, error)

	// Jobs
	CreateUsageExportJob(ctx context.Context, job *types.UsageExportJob) error
	GetUsageExportJob(ctx context.Context, id uuid.UUID) (*types.UsageExportJob, error)
	UpdateUsageExportJobStatus(ctx context.Context, id uuid.UUID, status types.UsageExportStatus, errorMessage string) error
	CompleteUsageExportJob(ctx context.Context, id uuid.UUID, storagePath, storageURL, checksum string, recordCount, fileSize int64) error
	UpdateDeliveryStatus(ctx context.Context, jobID uuid.UUID, status, errorMessage string) error
	UpdateLastExecution(ctx context.Context, configID, jobID uuid.UUID, executedAt time.Time) error
	ListUsageExportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.UsageExportJob, error)
	GetPendingScheduledConfigs(ctx context.Context, now time.Time) ([]*types.UsageExportConfiguration, error)
}

// ExternalBillingRepository handles external billing system integrations
type ExternalBillingRepository interface {
	CreateExternalBillingSystem(ctx context.Context, system *types.ExternalBillingSystem) error
	GetExternalBillingSystem(ctx context.Context, id uuid.UUID) (*types.ExternalBillingSystem, error)
	UpdateExternalBillingSystem(ctx context.Context, system *types.ExternalBillingSystem) error
	DeleteExternalBillingSystem(ctx context.Context, id uuid.UUID) error
	ListExternalBillingSystems(ctx context.Context, tenantID uuid.UUID, limit, offset int, activeOnly bool) ([]*types.ExternalBillingSystem, error)

	CreateBillingIntegrationSync(ctx context.Context, sync *types.BillingIntegrationSync) error
	GetBillingIntegrationSync(ctx context.Context, id uuid.UUID) (*types.BillingIntegrationSync, error)
	ListBillingIntegrationSyncs(ctx context.Context, tenantID uuid.UUID, systemID *uuid.UUID, status string, limit, offset int) ([]*types.BillingIntegrationSync, error)
}

// UsageTemplateRepository handles export templates
type UsageTemplateRepository interface {
	CreateUsageExportTemplate(ctx context.Context, template *types.UsageExportTemplate) error
	GetUsageExportTemplate(ctx context.Context, id uuid.UUID) (*types.UsageExportTemplate, error)
	ListUsageExportTemplates(ctx context.Context, category string) ([]*types.UsageExportTemplate, error)
}
