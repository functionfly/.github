package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StartDeployRetryTicker starts a background goroutine that retries failed
// bundle deployments every 60 seconds. Call this once during server startup.
func (h *Handler) StartDeployRetryTicker(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		logrus.Info("Bundle deploy retry ticker started")
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				logrus.Info("Bundle deploy retry ticker stopped")
				return
			case <-ticker.C:
				h.retryFailedDeployments(ctx)
			}
		}
	}()
}

func (h *Handler) retryFailedDeployments(ctx context.Context) {
	subs, err := h.repo.ListPendingDeployments(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to list pending deployments")
		return
	}

	for _, sub := range subs {
		if sub.NextRetryAt != nil && sub.NextRetryAt.After(time.Now()) {
			continue
		}

		log := logrus.WithFields(logrus.Fields{
			"tenant_id":      sub.TenantID,
			"script_name":    sub.ScriptName,
			"deploy_attempts": sub.DeployAttempts,
		})
		log.Info("Retrying bundle deployment")

		// Need a fresh provider ID if it was set
		if sub.ProviderID == nil {
			log.Warn("Skipping retry: no provider ID")
			continue
		}

		go h.DeployBundle(ctx, sub)
	}
}

// TriggerPendingDeployments checks for subscriptions awaiting a provider
// and triggers deployment if a matching provider was just connected.
// Call this after a provider is successfully connected.
func (h *Handler) TriggerPendingDeployments(ctx context.Context, tenantID string, providerSlug string, providerID string) {
	// Parse tenant ID
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return
	}

	awaiting, err := h.repo.ListAwaitingProvider(ctx, tid)
	if err != nil {
		logrus.WithError(err).Error("Failed to list awaiting provider subscriptions")
		return
	}

	for _, sub := range awaiting {
		log := logrus.WithFields(logrus.Fields{
			"tenant_id":    sub.TenantID,
			"bundle_slug":  sub.BundleID,
			"script_name":  sub.ScriptName,
			"provider":     providerSlug,
			"provider_id":  providerID,
		})
		log.Info("Provider connected, triggering bundle deployment")

		pid, _ := uuid.Parse(providerID)
		sub.ProviderID = &pid
		sub.DeployStatus = "deploying"
		if err := h.repo.UpdateBundleSubscription(ctx, sub); err != nil {
			log.WithError(err).Error("Failed to update subscription for deployment")
			continue
		}

		go h.DeployBundle(ctx, sub)
	}
}
