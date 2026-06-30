package aikeys

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthWorker periodically checks the health of BYOK keys.
type HealthWorker struct {
	repo           *Repository
	interval       time.Duration
	batchSize      int
	maxFailures    int
	stopCh         chan struct{}
}

// NewHealthWorker creates a new BYOK health check worker.
func NewHealthWorker(repo *Repository) *HealthWorker {
	interval := 6 * time.Hour
	if v := os.Getenv("BYOK_HEALTH_CHECK_INTERVAL_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			interval = time.Duration(h) * time.Hour
		}
	}

	batchSize := 100
	if v := os.Getenv("BYOK_HEALTH_CHECK_BATCH_SIZE"); v != "" {
		if b, err := strconv.Atoi(v); err == nil && b > 0 {
			batchSize = b
		}
	}

	maxFailures := 3
	if v := os.Getenv("BYOK_MAX_CONSECUTIVE_FAILURES"); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m > 0 {
			maxFailures = m
		}
	}

	return &HealthWorker{
		repo:        repo,
		interval:    interval,
		batchSize:   batchSize,
		maxFailures: maxFailures,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the health check loop. Run in a goroutine.
func (w *HealthWorker) Start(ctx context.Context) {
	logrus.WithField("interval", w.interval).Info("Starting BYOK health check worker")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("BYOK health check worker stopping (context cancelled)")
			return
		case <-w.stopCh:
			logrus.Info("BYOK health check worker stopped")
			return
		case <-ticker.C:
			w.runCheck(ctx)
		}
	}
}

// Stop signals the worker to stop.
func (w *HealthWorker) Stop() {
	close(w.stopCh)
}

func (w *HealthWorker) runCheck(ctx context.Context) {
	keys, err := w.repo.GetPendingHealthCheck(ctx, w.batchSize)
	if err != nil {
		logrus.WithError(err).Error("BYOK health check: failed to fetch keys")
		return
	}

	if len(keys) == 0 {
		return
	}

	logrus.WithField("count", len(keys)).Info("BYOK health check: checking keys")

	for _, key := range keys {
		result := testProviderAPI(key.Provider, "")

		newStatus := "active"
		message := "health check passed"

		if !result.IsValid {
			message = "health check failed: " + result.Message
			if key.Status == "degraded" {
				newStatus = "expired"
				message = "key failed consecutive health checks"
			} else {
				newStatus = "degraded"
			}
		}

		if newStatus != key.Status {
			logrus.WithFields(logrus.Fields{
				"key_id":     key.ID,
				"provider":   key.Provider,
				"old_status": key.Status,
				"new_status": newStatus,
			}).Info("BYOK key status changed")
		}

		_ = w.repo.UpdateHealthStatus(ctx, key.ID, newStatus, message)
	}
}

// StartHealthWorkerIfEnabled starts the BYOK health worker if BYOK is enabled.
func StartHealthWorkerIfEnabled(ctx context.Context, repo *Repository) *HealthWorker {
	if os.Getenv("BYOK_ENABLED") == "false" {
		return nil
	}

	worker := NewHealthWorker(repo)
	go worker.Start(ctx)
	return worker
}
