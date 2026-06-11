package dna

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Rollbacker handles rolling back DNA mutations that have been deployed.
type Rollbacker struct {
	repo                     *dna.Repository
	canaryRepo               *registry.CanaryConfigRepository
	functionRepo             *registry.RegistryRepository
	logger                   *logrus.Logger
	platformSettingsProvider PlatformSettingsProvider
	stopCh                   chan struct{}
	autoRollbackCtx          context.Context
	autoRollbackCancel       context.CancelFunc
}

// NewRollbacker creates a new DNA mutation rollbacker.
func NewRollbacker(
	repo *dna.Repository,
	canaryRepo *registry.CanaryConfigRepository,
	functionRepo *registry.RegistryRepository,
	logger *logrus.Logger,
) *Rollbacker {
	return &Rollbacker{
		repo:         repo,
		canaryRepo:   canaryRepo,
		functionRepo: functionRepo,
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// Stop gracefully stops the auto-rollback background monitor.
func (r *Rollbacker) Stop() {
	if r.autoRollbackCancel != nil {
		r.autoRollbackCancel()
	}
	close(r.stopCh)
}

// SetPlatformSettingsProvider sets the platform settings provider for auto-rollback threshold.
func (r *Rollbacker) SetPlatformSettingsProvider(provider PlatformSettingsProvider) {
	r.platformSettingsProvider = provider
}

// getErrorThreshold returns the configured error rate threshold for auto-rollback.
// Falls back to 5% if not configured.
func (r *Rollbacker) getErrorThreshold(ctx context.Context, userID string) float64 {
	if r.platformSettingsProvider != nil {
		settings, err := r.platformSettingsProvider.GetPlatformSettings(ctx, userID)
		if err == nil && settings != nil && settings.AutoRollbackOnError {
			return float64(settings.AutoRollbackErrorThreshold) / 100.0 // convert percentage to rate
		}
	}
	return 0.05 // default: 5%
}

// RollbackMutation rolls back a deployed mutation to the previous version.
// This cancels the active canary, marks the mutation as rolled_back, and
// logs the rollback event for audit purposes.
func (r *Rollbacker) RollbackMutation(ctx context.Context, mutationID, tenantID, reason string) error {
	ctx, span := tracing.StartSpan(ctx, "dna.rollback_mutation")
	defer tracing.Finish(ctx)

	tracing.SetAttribute(ctx, "mutation_id", mutationID)
	tracing.SetAttribute(ctx, "tenant_id", tenantID)
	tracing.SetAttribute(ctx, "reason", reason)

	m, err := r.repo.GetMutation(ctx, mutationID)
	if err != nil {
		tracing.RecordError(ctx, err)
		return fmt.Errorf("get mutation: %w", err)
	}
	if m == nil {
		return ErrMutationNotFound
	}
	if m.TenantID != tenantID {
		return ErrAccessDenied
	}
	if m.Status != "deployed" && m.Status != "accepted" && m.Status != "deploying" {
		return fmt.Errorf("%w: %s", ErrMutationNotRollbackable, m.Status)
	}

	tracing.SetAttribute(ctx, "function_id", m.FunctionID)
	tracing.SetAttribute(ctx, "current_status", m.Status)

	// Cancel active canary if exists
	fnUUID, err := uuid.Parse(m.FunctionID)
	if err == nil {
		existingCanary, _ := r.canaryRepo.GetByFunctionID(fnUUID)
		if existingCanary != nil && existingCanary.Status == "active" {
			if err := r.canaryRepo.UpdateStatus(existingCanary.ID, "rolled_back"); err != nil {
				r.logger.WithError(err).Warn("dna: failed to rollback canary config")
			} else {
				r.logger.WithFields(logrus.Fields{
					"canary_id":   existingCanary.ID,
					"mutation_id": mutationID,
				}).Info("dna: canary config rolled back")
			}
		}
	}

	// Update mutation status to rolled_back
	now := time.Now()
	if err := r.repo.UpdateMutationStatus(ctx, mutationID, "rolled_back", map[string]interface{}{
		"rollback_reason": reason,
		"rolled_back_at":  now,
	}); err != nil {
		tracing.RecordError(ctx, err)
		return fmt.Errorf("update mutation status: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"mutation_id":     mutationID,
		"function_id":      m.FunctionID,
		"rollback_reason":  reason,
	}).Info("dna: mutation rolled back successfully")

	return nil
}

// AutoRollbackOnError monitors canary deployments and automatically rolls back
// if the error rate exceeds the configured threshold.
func (r *Rollbacker) AutoRollbackOnError(ctx context.Context) {
	r.logger.Info("dna: auto-rollback monitor started")

	autoCtx, autoCancel := context.WithCancel(ctx)
	r.autoRollbackCtx = autoCtx
	r.autoRollbackCancel = autoCancel

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("dna: auto-rollback monitor stopped")
			return
		case <-r.stopCh:
			r.logger.Info("dna: auto-rollback monitor stopped via Stop()")
			autoCancel()
			return
		case <-ticker.C:
			r.checkCanariesForRollback(autoCtx)
		}
	}
}

func (r *Rollbacker) checkCanariesForRollback(ctx context.Context) {
	if r.platformSettingsProvider == nil {
		return // auto-rollback disabled without settings provider
	}

	activeCanaries, err := r.canaryRepo.GetAllActive()
	if err != nil {
		r.logger.WithError(err).Error("dna: failed to get active canaries for auto-rollback check")
		return
	}

	for _, canary := range activeCanaries {
		select {
		case <-ctx.Done():
			return
		default:
		}

		r.checkSingleCanary(ctx, canary)
	}
}

func (r *Rollbacker) checkSingleCanary(ctx context.Context, canary *registry.CanaryConfig) {
	ctx, span := tracing.StartSpan(ctx, "dna.check_single_canary")
	defer tracing.Finish(ctx)

	tracing.SetAttribute(ctx, "canary_id", canary.ID)
	tracing.SetAttribute(ctx, "function_id", canary.FunctionID.String())
	tracing.SetAttribute(ctx, "version", canary.Version)

	// Only check canaries that are DNA mutations (version starts with "dna-")
	if len(canary.Version) < 4 || canary.Version[:4] != "dna-" {
		return
	}

	// Calculate current error rate
	errorRate := float64(0)
	if canary.RequestCount > 0 {
		errorRate = float64(canary.RequestCount-canary.SuccessCount) / float64(canary.RequestCount)
	}

	tracing.SetAttribute(ctx, "error_rate", errorRate)
	tracing.SetAttribute(ctx, "request_count", canary.RequestCount)
	tracing.SetAttribute(ctx, "success_count", canary.SuccessCount)

	// Get the threshold
	// Use function ID as user ID for platform settings lookup
	threshold := r.getErrorThreshold(ctx, canary.FunctionID.String())

	tracing.SetAttribute(ctx, "threshold", threshold)

	if errorRate > threshold && errorRate > 0 {
		r.logger.WithFields(logrus.Fields{
			"canary_id":    canary.ID,
			"function_id":  canary.FunctionID,
			"error_rate":   errorRate,
			"threshold":    threshold,
			"version":      canary.Version,
		}).Warn("dna: canary error rate exceeds threshold — triggering auto-rollback")

		// Look up function to get tenant ID for tenant-isolated query
		fn, err := r.functionRepo.GetFunctionByID(canary.FunctionID)
		if err != nil {
			r.logger.WithError(err).WithField("function_id", canary.FunctionID).Error("dna: failed to get function for tenant lookup")
			tracing.RecordError(ctx, err)
			return
		}
		tenantID := ""
		if fn.TenantID != nil {
			tenantID = fn.TenantID.String()
		}
		tracing.SetAttribute(ctx, "tenant_id", tenantID)

		// Find the mutation for this canary (tenant-isolated)
		mutations, _, err := r.repo.ListMutations(ctx, canary.FunctionID.String(), tenantID, "", 10, 0)
		if err != nil {
			r.logger.WithError(err).WithField("canary_id", canary.ID).Error("dna: failed to find mutation for auto-rollback")
			tracing.RecordError(ctx, err)
			return
		}

		var mutationID string
		for _, m := range mutations {
			if m.Status == "accepted" || m.Status == "deploying" || m.Status == "deployed" {
				mutationID = m.ID
				break
			}
		}

		if mutationID == "" {
			r.logger.WithField("canary_id", canary.ID).Warn("dna: no active mutation found for canary")
			return
		}

		// Perform rollback
		if err := r.RollbackMutation(ctx, mutationID, tenantID, fmt.Sprintf("auto-rollback: error rate %.2f%% exceeded threshold %.2f%%", errorRate*100, threshold*100)); err != nil {
			r.logger.WithError(err).WithFields(logrus.Fields{
				"canary_id":    canary.ID,
				"mutation_id":  mutationID,
			}).Error("dna: auto-rollback failed")
			tracing.RecordError(ctx, err)
		} else {
			r.logger.WithFields(logrus.Fields{
				"canary_id":   canary.ID,
				"mutation_id": mutationID,
				"error_rate":  errorRate,
			}).Info("dna: auto-rollback completed")
		}
	}
}
