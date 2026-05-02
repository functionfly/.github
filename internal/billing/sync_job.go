package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BillingSyncJob runs billing integration sync operations asynchronously.
// It processes pending sync records from billing_integration_syncs table
// and executes the actual data synchronization with external billing systems.
type BillingSyncJob struct {
	exportRepo   *storage.ExportRepository
	exporters    map[storage.BillingSystemType]func() ExternalExporter
	stopCh       chan struct{}
	stopOnce     sync.Once
	log          *logrus.Logger
	workerCount  int
	pollInterval time.Duration
}

// BillingSyncJobConfig configures the billing sync job.
type BillingSyncJobConfig struct {
	WorkerCount  int
	PollInterval time.Duration
}

// NewBillingSyncJob creates a new billing sync job.
func NewBillingSyncJob(exportRepo *storage.ExportRepository, config ...BillingSyncJobConfig) *BillingSyncJob {
	cfg := BillingSyncJobConfig{
		WorkerCount:  2,
		PollInterval: 5 * time.Second,
	}
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 2
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 5 * time.Second
	}

	return &BillingSyncJob{
		exportRepo:   exportRepo,
		exporters:    ExporterRegistry,
		stopCh:       make(chan struct{}),
		log:          logrus.New(),
		workerCount:  cfg.WorkerCount,
		pollInterval: cfg.PollInterval,
	}
}

// Start begins the billing sync job workers in the background.
func (j *BillingSyncJob) Start(ctx context.Context) {
	j.log.WithFields(logrus.Fields{
		"workers":       j.workerCount,
		"poll_interval": j.pollInterval,
	}).Info("Billing sync job starting")

	// Start workers
	for i := 0; i < j.workerCount; i++ {
		go j.worker(ctx, i)
	}
}

// Stop stops the billing sync job.
func (j *BillingSyncJob) Stop() {
	j.stopOnce.Do(func() {
		close(j.stopCh)
		j.log.Info("Billing sync job stopped")
	})
}

// worker processes sync jobs from the queue.
func (j *BillingSyncJob) worker(ctx context.Context, id int) {
	j.log.WithField("worker_id", id).Debug("Billing sync worker started")
	defer j.log.WithField("worker_id", id).Debug("Billing sync worker stopped")

	ticker := time.NewTicker(j.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-j.stopCh:
			return
		case <-ticker.C:
			if err := j.processPendingSyncs(ctx); err != nil {
				j.log.WithError(err).WithField("worker_id", id).Warn("Failed to process pending syncs")
			}
		}
	}
}

// processPendingSyncs fetches and processes pending billing syncs.
func (j *BillingSyncJob) processPendingSyncs(ctx context.Context) error {
	// Get pending syncs (status = 'pending' or 'failed' with retry)
	// We need to list across all tenants - get a batch of pending syncs
	// Since there's no direct method, we'll get syncs with status filter
	pendingSyncs, err := j.getPendingSyncs(ctx, 10) // Process up to 10 at a time
	if err != nil {
		return fmt.Errorf("failed to get pending syncs: %w", err)
	}

	if len(pendingSyncs) == 0 {
		return nil // No work to do
	}

	j.log.WithField("count", len(pendingSyncs)).Info("Processing pending billing syncs")

	for _, sync := range pendingSyncs {
		if err := j.processSync(ctx, sync); err != nil {
			j.log.WithError(err).WithField("sync_id", sync.ID).Error("Failed to process sync")
			// Continue with next sync - don't stop processing
		}
	}

	return nil
}

// getPendingSyncs retrieves pending billing syncs that need to be processed
func (j *BillingSyncJob) getPendingSyncs(ctx context.Context, limit int) ([]*storage.BillingIntegrationSync, error) {
	return j.exportRepo.GetPendingSyncs(ctx, limit)
}

// processSync processes a single billing sync
func (j *BillingSyncJob) processSync(ctx context.Context, sync *storage.BillingIntegrationSync) error {
	log := j.log.WithFields(logrus.Fields{
		"sync_id":   sync.ID,
		"tenant_id": sync.TenantID,
		"sync_type": sync.SyncType,
	})

	log.Info("Processing billing sync")

	// Update status to running
	if err := j.exportRepo.UpdateBillingIntegrationSyncStatus(ctx, sync.ID, "running", nil); err != nil {
		return fmt.Errorf("failed to update sync status to running: %w", err)
	}

	// Get the external billing system configuration
	system, err := j.exportRepo.GetExternalBillingSystem(ctx, sync.ExternalSystemID)
	if err != nil {
		j.failSync(ctx, sync.ID, fmt.Sprintf("Failed to get billing system: %v", err))
		return fmt.Errorf("failed to get external billing system: %w", err)
	}

	// Get the appropriate exporter
	systemType := storage.BillingSystemType(system.SystemType)
	exporterFactory, ok := j.exporters[systemType]
	if !ok {
		j.failSync(ctx, sync.ID, fmt.Sprintf("Unsupported billing system type: %s", system.SystemType))
		return fmt.Errorf("unsupported billing system type: %s", system.SystemType)
	}

	exporter := exporterFactory()

	// Test the connection first
	if err := exporter.TestConnection(ctx, system); err != nil {
		j.failSync(ctx, sync.ID, fmt.Sprintf("Connection test failed: %v", err))
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Execute the actual data synchronization based on sync type
	result, err := j.executeDataSync(ctx, exporter, system, sync)
	if err != nil {
		j.failSync(ctx, sync.ID, fmt.Sprintf("Data sync failed: %v", err))
		return err
	}

	// Update sync stats with results
	if err := j.exportRepo.UpdateBillingIntegrationSyncStats(ctx, sync.ID,
		result.RecordsProcessed, result.RecordsCreated, result.RecordsUpdated,
		result.RecordsFailed, 0); err != nil {
		log.WithError(err).Warn("Failed to update sync stats")
	}

	// Mark sync as completed
	completedAt := time.Now()
	if err := j.exportRepo.UpdateBillingIntegrationSyncStatus(ctx, sync.ID, "completed", &completedAt); err != nil {
		return fmt.Errorf("failed to complete sync: %w", err)
	}

	// Update the external system's last sync time
	if err := j.exportRepo.UpdateExternalBillingSystemSyncStatus(ctx, sync.ExternalSystemID, "completed"); err != nil {
		log.WithError(err).Warn("Failed to update external system sync status")
	}

	log.WithFields(logrus.Fields{
		"records_processed": result.RecordsProcessed,
		"records_created":   result.RecordsCreated,
		"records_updated":   result.RecordsUpdated,
	}).Info("Billing sync completed successfully")

	return nil
}

// executeDataSync performs the actual data synchronization based on sync type
func (j *BillingSyncJob) executeDataSync(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {

	switch sync.SyncType {
	case "customers":
		return j.syncCustomers(ctx, exporter, system, sync)
	case "invoices":
		return j.syncInvoices(ctx, exporter, system, sync)
	case "usage":
		return j.syncUsage(ctx, exporter, system, sync)
	case "payments":
		return j.syncPayments(ctx, exporter, system, sync)
	case "all":
		// Sync everything
		return j.syncAll(ctx, exporter, system, sync)
	default:
		return nil, fmt.Errorf("unsupported sync type: %s", sync.SyncType)
	}
}

// syncCustomers syncs tenant customers to the external billing system
func (j *BillingSyncJob) syncCustomers(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {

	// Get tenant users to sync as customers
	// This is a simplified implementation - in production you'd want to batch this
	customers := []CustomerExport{
		{
			Name:     fmt.Sprintf("Tenant %s", sync.TenantID.String()),
			Email:    fmt.Sprintf("billing+%s@tenant.com", sync.TenantID.String()),
			Currency: "USD",
		},
	}

	return exporter.SyncCustomers(ctx, system, customers)
}

// syncInvoices syncs invoices to the external billing system
func (j *BillingSyncJob) syncInvoices(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {
	// Get invoices for the tenant
	// For now, return a placeholder result
	return &ExportResult{
		RecordsProcessed: 0,
		RecordsCreated:   0,
		RecordsUpdated:   0,
		RecordsFailed:    0,
	}, nil
}

// syncUsage syncs usage data to the external billing system
func (j *BillingSyncJob) syncUsage(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {

	// Get cost allocation data for the tenant
	// This would typically fetch usage data and convert to line items
	lineItems := []LineItemExport{
		{
			TenantID:      sync.TenantID,
			TenantName:    sync.TenantID.String(),
			Description:   "Function executions",
			Quantity:      1,
			UnitCostCents: 0,
			TotalCents:    0,
			ServiceDate:   time.Now(),
		},
	}

	return exporter.ExportLineItems(ctx, system, lineItems)
}

// syncPayments syncs payment data to the external billing system
func (j *BillingSyncJob) syncPayments(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {
	// Payment sync would typically involve syncing payment transactions
	// For now, return a placeholder result
	return &ExportResult{
		RecordsProcessed: 0,
		RecordsCreated:   0,
		RecordsUpdated:   0,
		RecordsFailed:    0,
	}, nil
}

// syncAll syncs all billing data types
func (j *BillingSyncJob) syncAll(ctx context.Context, exporter ExternalExporter,
	system *storage.ExternalBillingSystem, sync *storage.BillingIntegrationSync) (*ExportResult, error) {

	totalResult := &ExportResult{}

	// Sync customers
	customerResult, err := j.syncCustomers(ctx, exporter, system, sync)
	if err != nil {
		return nil, fmt.Errorf("customer sync failed: %w", err)
	}
	mergeResults(totalResult, customerResult)

	// Sync usage/line items
	usageResult, err := j.syncUsage(ctx, exporter, system, sync)
	if err != nil {
		return nil, fmt.Errorf("usage sync failed: %w", err)
	}
	mergeResults(totalResult, usageResult)

	// Sync invoices
	invoiceResult, err := j.syncInvoices(ctx, exporter, system, sync)
	if err != nil {
		return nil, fmt.Errorf("invoice sync failed: %w", err)
	}
	mergeResults(totalResult, invoiceResult)

	return totalResult, nil
}

// mergeResults combines two export results
func mergeResults(total, partial *ExportResult) {
	total.RecordsProcessed += partial.RecordsProcessed
	total.RecordsCreated += partial.RecordsCreated
	total.RecordsUpdated += partial.RecordsUpdated
	total.RecordsFailed += partial.RecordsFailed
	if partial.ExternalBatchID != "" {
		total.ExternalBatchID = partial.ExternalBatchID
	}
	if total.ExternalRefs == nil {
		total.ExternalRefs = make(map[string]string)
	}
	for k, v := range partial.ExternalRefs {
		total.ExternalRefs[k] = v
	}
	total.Errors = append(total.Errors, partial.Errors...)
}

// TriggerSync triggers a billing sync asynchronously.
// This is called from the HTTP handler when a user manually triggers a sync.
func (j *BillingSyncJob) TriggerSync(ctx context.Context, syncID uuid.UUID, tenantID uuid.UUID, systemID uuid.UUID) error {
	if j == nil {
		return fmt.Errorf("billing sync job not initialized")
	}

	// Run the sync asynchronously
	go j.executeSync(ctx, syncID, tenantID, systemID)

	return nil
}

// executeSync performs the actual sync operation.
func (j *BillingSyncJob) executeSync(ctx context.Context, syncID uuid.UUID, tenantID uuid.UUID, systemID uuid.UUID) {
	log := j.log.WithFields(logrus.Fields{
		"sync_id":   syncID,
		"tenant_id": tenantID,
		"system_id": systemID,
	})

	log.Info("Starting billing sync execution")

	// Get the pending sync to get the full sync object
	syncs, err := j.exportRepo.GetPendingSyncs(ctx, 100)
	if err != nil {
		log.WithError(err).Error("Failed to get pending syncs")
		j.failSync(ctx, syncID, fmt.Sprintf("Failed to get pending syncs: %v", err))
		return
	}

	var billingSync *storage.BillingIntegrationSync
	for _, s := range syncs {
		if s.ID == syncID {
			billingSync = s
			break
		}
	}

	if billingSync == nil {
		log.Error("Sync not found in pending syncs")
		j.failSync(ctx, syncID, "Sync not found")
		return
	}

	// Use the existing processSync which handles all the logic
	if err := j.processSync(ctx, billingSync); err != nil {
		log.WithError(err).Error("Sync processing failed")
	}
}

// failSync marks a sync as failed with the given error message.
func (j *BillingSyncJob) failSync(ctx context.Context, syncID uuid.UUID, errorMsg string) {
	// Update status to failed - we use the status method with no completed_at
	if err := j.exportRepo.UpdateBillingIntegrationSyncStatus(ctx, syncID, "failed", nil); err != nil {
		j.log.WithError(err).WithField("sync_id", syncID).Error("Failed to mark sync as failed")
	}

	// Note: error message would ideally be stored in error_message column
	// but that's not available in the current repository interface
	j.log.WithField("sync_id", syncID).WithField("error", errorMsg).Error("Billing sync failed")
}
