package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SyncScheduler struct {
	connectorRepo *storage.ConnectorRepository
	brainRepo     *storage.BrainRepository
	registry      *Registry
	logger        *logrus.Logger
	stopCh        chan struct{}
}

func NewSyncScheduler(
	connectorRepo *storage.ConnectorRepository,
	brainRepo *storage.BrainRepository,
	registry *Registry,
	logger *logrus.Logger,
) *SyncScheduler {
	return &SyncScheduler{
		connectorRepo: connectorRepo,
		brainRepo:     brainRepo,
		registry:      registry,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the background sync scheduler
func (s *SyncScheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.runSyncCycle(ctx)
			}
		}
	}()
	s.logger.Info("Brain sync scheduler started")
}

// Stop gracefully stops the scheduler
func (s *SyncScheduler) Stop() {
	close(s.stopCh)
	s.logger.Info("Brain sync scheduler stopped")
}

// runSyncCycle fetches connectors due for sync and processes them
func (s *SyncScheduler) runSyncCycle(ctx context.Context) {
	// Get all active connectors that need syncing
	connectors, err := s.connectorRepo.GetActiveConnectorsForSync(ctx, "5 minutes")
	if err != nil {
		s.logger.WithError(err).Error("Failed to get connectors for sync")
		return
	}

	for _, uc := range connectors {
		syncFreq := plans.StarterBrainSyncFrequency // Default
		s.logger.WithFields(logrus.Fields{
			"connector": uc.ConnectorSlug,
			"tenant":    uc.TenantID,
			"frequency": syncFreq,
		}).Debug("Syncing connector")

		if err := s.syncConnector(ctx, uc); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"connector": uc.ConnectorSlug,
				"tenant":    uc.TenantID,
			}).Error("Connector sync failed")
			s.connectorRepo.UpdateUserConnectorStatus(ctx, uc.ID, "sync_error", err.Error())
			continue
		}

		if err := s.connectorRepo.UpdateLastSyncAt(ctx, uc.ID); err != nil {
			s.logger.WithError(err).Error("Failed to update last sync time")
		}
	}
}

// syncConnector handles the sync for a single user connector
func (s *SyncScheduler) syncConnector(ctx context.Context, uc *storage.UserConnector) error {
	ext, ok := s.registry.Get(uc.ConnectorSlug)
	if !ok {
		return fmt.Errorf("no extractor for connector: %s", uc.ConnectorSlug)
	}

	// In a real implementation, this would:
	// 1. Decrypt credentials (zero-knowledge, client-derived key)
	// 2. Fetch data from the external API using the OAuth token
	// 3. Pass raw data to the extractor
	// 4. Save resulting signals

	// For now, the extractor is ready to receive webhook data
	_ = ext
	return nil
}

// ProcessWebhookData handles real-time webhook data for a connector
func (s *SyncScheduler) ProcessWebhookData(ctx context.Context, tenantID uuid.UUID, connectorSlug string, data []byte) (int, error) {
	signals, err := s.registry.ExtractSignals(ctx, connectorSlug, tenantID, data)
	if err != nil {
		return 0, fmt.Errorf("extract signals: %w", err)
	}

	if len(signals) == 0 {
		return 0, nil
	}

	count, err := s.brainRepo.SaveSignalsBatch(ctx, signals)
	if err != nil {
		return 0, fmt.Errorf("save signals: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"connector": connectorSlug,
		"tenant":    tenantID,
		"signals":   count,
	}).Info("Processed webhook signals")

	return count, nil
}
