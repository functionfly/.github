package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ReconciliationSettings struct {
	ID                        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantID                  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	AutoReconcileEnabled      bool      `json:"auto_reconcile_enabled" gorm:"default:false"`
	ScheduledReconcileEnabled bool      `json:"scheduled_reconcile_enabled" gorm:"default:false"`
	ScheduleCron              string    `json:"schedule_cron" gorm:"type:varchar(100);default:'0 2 * * *'"`
	AuditExportEnabled        bool      `json:"audit_export_enabled" gorm:"default:false"`
	NotifyOnCompletion        bool      `json:"notify_on_completion" gorm:"default:true"`
	NotifyOnFailure           bool      `json:"notify_on_failure" gorm:"default:true"`
	LastReconciliationAt      time.Time `json:"last_reconciliation_at" gorm:"default:null"`
	CreatedAt                 time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ReconciliationSettings) TableName() string {
	return "reconciliation_settings"
}

type ReconciliationUsage struct {
	TotalReconciliations        int64 `json:"total_reconciliations"`
	TotalExecutionsReconciled   int64 `json:"total_executions_reconciled"`
	AvgDurationMs                int64 `json:"avg_duration_ms"`
	SuccessfulReconciliations    int64 `json:"successful_reconciliations"`
	FailedReconciliations        int64 `json:"failed_reconciliations"`
}

type ReconciliationStats struct {
	TotalReconciliations     int64 `json:"total_reconciliations"`
	SuccessfulReconciliations int64 `json:"successful_reconciliations"`
	FailedReconciliations    int64 `json:"failed_reconciliations"`
}

func (r *BillingRepository) GetReconciliationSettings(ctx context.Context, tenantID uuid.UUID) (*ReconciliationSettings, error) {
	settings := &ReconciliationSettings{}
	err := r.db.GORM.WithContext(ctx).Where("tenant_id = ?", tenantID).First(settings).Error
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return settings, nil
}

func (r *BillingRepository) UpsertReconciliationSettings(ctx context.Context, settings *ReconciliationSettings) error {
	existing, err := r.GetReconciliationSettings(ctx, settings.TenantID)
	if err != nil {
		return err
	}

	if existing == nil {
		settings.ID = uuid.New()
		settings.CreatedAt = time.Now()
		settings.UpdatedAt = time.Now()
		return r.db.GORM.WithContext(ctx).Create(settings).Error
	}

	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	settings.UpdatedAt = time.Now()
	return r.db.GORM.WithContext(ctx).Save(settings).Error
}

func (r *BillingRepository) GetReconciliationStats(ctx context.Context, tenantID uuid.UUID) (*ReconciliationStats, error) {
	var stats ReconciliationStats

	err := r.db.GORM.WithContext(ctx).
		Model(&ReconciliationSettings{}).
		Where("tenant_id = ?", tenantID).
		Select(`
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND status = 'completed'), 0) as total_reconciliations,
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND status = 'completed' AND error_message IS NULL), 0) as successful_reconciliations,
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND status = 'failed'), 0) as failed_reconciliations
		`, tenantID, tenantID, tenantID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *BillingRepository) GetLiveReconciliationUsage(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*ReconciliationUsage, error) {
	var usage ReconciliationUsage

	err := r.db.GORM.WithContext(ctx).
		Model(&ReconciliationSettings{}).
		Where("tenant_id = ?", tenantID).
		Select(`
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND started_at BETWEEN ? AND ?), 0) as total_reconciliations,
			COALESCE((SELECT SUM(executions_reconciled) FROM reconciliation_runs WHERE tenant_id = ? AND started_at BETWEEN ? AND ?), 0) as total_executions_reconciled,
			COALESCE((SELECT AVG(duration_ms) FROM reconciliation_runs WHERE tenant_id = ? AND started_at BETWEEN ? AND ? AND duration_ms > 0), 0) as avg_duration_ms,
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND started_at BETWEEN ? AND ? AND status = 'completed' AND error_message IS NULL), 0) as successful_reconciliations,
			COALESCE((SELECT COUNT(*) FROM reconciliation_runs WHERE tenant_id = ? AND started_at BETWEEN ? AND ? AND status = 'failed'), 0) as failed_reconciliations
		`, tenantID, start, end, tenantID, start, end, tenantID, start, end, tenantID, start, end, tenantID, start, end).
		Scan(&usage).Error

	if err != nil {
		return nil, err
	}

	return &usage, nil
}

func (r *BillingRepository) UpdateLastReconciliationAt(ctx context.Context, tenantID uuid.UUID, at time.Time) error {
	return r.db.GORM.WithContext(ctx).
		Model(&ReconciliationSettings{}).
		Where("tenant_id = ?", tenantID).
		Update("last_reconciliation_at", at).Error
}