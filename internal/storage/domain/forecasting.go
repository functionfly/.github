package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// UsageForecastRepository handles usage forecasting and alerting
type UsageForecastRepository interface {
	// Usage alerts
	CreateUsageAlert(ctx context.Context, alert *storage.UsageAlert) error
	GetUsageAlertByID(ctx context.Context, id uuid.UUID) (*storage.UsageAlert, error)
	ListUsageAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.UsageAlert, error)
	UpdateUsageAlert(ctx context.Context, alert *storage.UsageAlert) error
	DeleteUsageAlert(ctx context.Context, id uuid.UUID) error
	RecordAlertTrigger(ctx context.Context, history *storage.UsageAlertHistory) error
	GetAlertHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*storage.UsageAlertHistory, error)

	// Spend caps
	CreateOrUpdateSpendCap(ctx context.Context, cap *storage.SpendCap) error
	GetSpendCapByTenant(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*storage.SpendCap, error)
	UpdateCurrentSpend(ctx context.Context, capID uuid.UUID, spendCents int) error

	// Forecasts
	SaveUsageForecast(ctx context.Context, forecast *storage.UsageForecast) error
	GetLatestForecast(ctx context.Context, tenantID uuid.UUID, forecastType string) (*storage.UsageForecast, error)
	GetDailyUsageHistory(ctx context.Context, tenantID uuid.UUID, eventType string, days int) ([]*storage.DailyUsagePoint, error)
	GetDailySpendHistory(ctx context.Context, tenantID uuid.UUID, days int) ([]*storage.DailyUsagePoint, error)
	GetCurrentPeriodUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*storage.UsageSummary, error)
}

// CostAllocationRepository handles detailed cost tracking
type CostAllocationRepository interface {
	RecordCostAllocationEntry(ctx context.Context, entry *storage.CostAllocationEntry) error
	GetCostAllocationByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*storage.CostAllocationSummary, error)
	GetCostAllocationDailyBreakdown(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*storage.DailyCostBreakdown, error)
	GetTenantCostSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*storage.TenantCostSummary, error)
	GetAllTenantsCostSummary(ctx context.Context, start, end time.Time) ([]*storage.TenantCostSummary, error)
	GetCostAllocationEntries(ctx context.Context, filter *storage.CostAllocationFilter, limit, offset int) ([]*storage.CostAllocationEntry, int, error)
	GetCostAllocationReport(ctx context.Context, start, end time.Time) (*storage.CostAllocationReport, error)
	GetCostAllocationByRegion(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (map[string]*storage.CostAllocationSummary, error)
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
	GetExecutionRetentionSettings(ctx context.Context) (*storage.ExecutionRetentionSettings, error)
	UpdateExecutionRetentionSettings(ctx context.Context, updates *storage.ExecutionRetentionSettingsUpdate) (*storage.ExecutionRetentionSettings, error)
	GetOrCreateExecutionRetentionSettings(ctx context.Context) (*storage.ExecutionRetentionSettings, error)
	ResetExecutionRetentionSettingsToDefaults(ctx context.Context, updatedBy *uuid.UUID) (*storage.ExecutionRetentionSettings, error)
}

// UsageExportRepository handles usage data export operations
type UsageExportRepository interface {
	// Configurations
	CreateUsageExportConfiguration(ctx context.Context, config *storage.UsageExportConfiguration) error
	GetUsageExportConfiguration(ctx context.Context, id uuid.UUID) (*storage.UsageExportConfiguration, error)
	UpdateUsageExportConfiguration(ctx context.Context, config *storage.UsageExportConfiguration) error
	DeleteUsageExportConfiguration(ctx context.Context, id uuid.UUID) error
	ListUsageExportConfigurations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storage.UsageExportConfiguration, error)

	// Jobs
	CreateUsageExportJob(ctx context.Context, job *storage.UsageExportJob) error
	GetUsageExportJob(ctx context.Context, id uuid.UUID) (*storage.UsageExportJob, error)
	UpdateUsageExportJobStatus(ctx context.Context, id uuid.UUID, status storage.UsageExportStatus, errorMessage string) error
	CompleteUsageExportJob(ctx context.Context, id uuid.UUID, storagePath, storageURL, checksum string, recordCount, fileSize int64) error
	UpdateDeliveryStatus(ctx context.Context, jobID uuid.UUID, status, errorMessage string) error
	UpdateLastExecution(ctx context.Context, configID, jobID uuid.UUID, executedAt time.Time) error
	ListUsageExportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storage.UsageExportJob, error)
	GetPendingScheduledConfigs(ctx context.Context, now time.Time) ([]*storage.UsageExportConfiguration, error)
}

// ExternalBillingRepository handles external billing system integrations
type ExternalBillingRepository interface {
	CreateExternalBillingSystem(ctx context.Context, system *storage.ExternalBillingSystem) error
	GetExternalBillingSystem(ctx context.Context, id uuid.UUID) (*storage.ExternalBillingSystem, error)
	UpdateExternalBillingSystem(ctx context.Context, system *storage.ExternalBillingSystem) error
	DeleteExternalBillingSystem(ctx context.Context, id uuid.UUID) error
	ListExternalBillingSystems(ctx context.Context, tenantID uuid.UUID, limit, offset int, activeOnly bool) ([]*storage.ExternalBillingSystem, error)

	CreateBillingIntegrationSync(ctx context.Context, sync *storage.BillingIntegrationSync) error
	GetBillingIntegrationSync(ctx context.Context, id uuid.UUID) (*storage.BillingIntegrationSync, error)
	ListBillingIntegrationSyncs(ctx context.Context, tenantID uuid.UUID, systemID *uuid.UUID, status string, limit, offset int) ([]*storage.BillingIntegrationSync, error)
}

// UsageTemplateRepository handles export templates
type UsageTemplateRepository interface {
	CreateUsageExportTemplate(ctx context.Context, template *storage.UsageExportTemplate) error
	GetUsageExportTemplate(ctx context.Context, id uuid.UUID) (*storage.UsageExportTemplate, error)
	ListUsageExportTemplates(ctx context.Context, category string) ([]*storage.UsageExportTemplate, error)
}
