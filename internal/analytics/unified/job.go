package unified

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// SyncJob runs SyncFromSources periodically to fill analytics_rollups from source tables.
type SyncJob struct {
	store    *EventStore
	interval time.Duration
	stopCh   chan struct{}
	log      *logrus.Logger
}

// NewSyncJob creates a sync job that runs every interval (e.g. 24*time.Hour).
func NewSyncJob(store *EventStore, interval time.Duration) *SyncJob {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &SyncJob{
		store:    store,
		interval: interval,
		stopCh:   make(chan struct{}),
		log:      logrus.New(),
	}
}

// Start runs the sync job in the background. Runs once immediately, then every interval.
func (j *SyncJob) Start(ctx context.Context) {
	j.log.WithField("interval", j.interval).Info("Unified analytics sync job starting")
	// Run once soon after start (ensure tables exist, then sync last 2 days)
	go func() {
		if err := j.store.AutoMigrate(ctx); err != nil {
			j.log.WithError(err).Warn("Unified analytics AutoMigrate failed (tables may already exist)")
		}
		start := time.Now().UTC().Add(-48 * time.Hour) // last 2 days
		end := time.Now().UTC()
		if err := j.store.SyncFromSources(ctx, start, end); err != nil {
			j.log.WithError(err).Warn("Unified analytics initial sync failed")
		}
	}()
	go func() {
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				j.log.Info("Unified analytics sync job stopping (context)")
				return
			case <-j.stopCh:
				j.log.Info("Unified analytics sync job stopped")
				return
			case <-ticker.C:
				start := time.Now().UTC().Add(-j.interval - 1*time.Hour)
				end := time.Now().UTC()
				if err := j.store.SyncFromSources(ctx, start, end); err != nil {
					j.log.WithError(err).Warn("Unified analytics periodic sync failed")
				}
			}
		}
	}()
}

// Stop stops the sync job.
func (j *SyncJob) Stop() {
	close(j.stopCh)
}
