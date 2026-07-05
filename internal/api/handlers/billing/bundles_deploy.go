package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// DeployBundle deploys bundle templates to the user's connected provider.
// It loads templates from the DB, generates a single worker script,
// deploys it, creates a backend, and updates the subscription status.
func (h *Handler) DeployBundle(ctx context.Context, sub *storage.BundleSubscription) {
	log := logrus.WithFields(logrus.Fields{
		"tenant_id":     sub.TenantID,
		"bundle_slug":   sub.BundleID,
		"deploy_status": sub.DeployStatus,
		"script_name":   sub.ScriptName,
	})

	log.Info("Starting bundle deployment")

	// Load provider record
	provider, err := h.repo.GetProviderByID(ctx, *sub.ProviderID)
	if err != nil || provider == nil {
		h.failDeploy(sub, "provider not found: %v", err)
		return
	}

	// Validate provider access
	if err := h.validateProviderAccess(ctx, provider); err != nil {
		h.failDeploy(sub, "provider access denied: %v", err)
		return
	}

	// Load bundle templates
	templates, err := h.repo.ListBundleTemplates(ctx, sub.BundleID.String())
	if err != nil {
		h.failDeploy(sub, "failed to load templates: %v", err)
		return
	}
	if len(templates) == 0 {
		h.failDeploy(sub, "no templates found for bundle")
		return
	}

	// Check idempotency — already deployed?
	app, err := h.repo.GetAppBySlug(ctx, sub.BundleID.String())
	if err == nil && app != nil {
		backends, _ := h.repo.ListBackendsByAppID(ctx, app.ID)
		for _, b := range backends {
			if b.Provider == "workers" && strings.Contains(b.URL, sub.ScriptName) {
				log.Info("Backend already exists, skipping deploy")
				sub.DeployStatus = "deployed"
				now := time.Now()
				sub.DeployedAt = &now
				if err := h.repo.UpdateBundleSubscription(ctx, sub); err != nil {
					log.WithError(err).Error("failed to update bundle subscription")
				}
				return
			}
		}
	}

	// Deploy to provider
	switch provider.Provider {
	case "workers":
		h.deployToCloudflare(ctx, sub, provider, templates)
	default:
		h.failDeploy(sub, "unsupported provider: %s", provider.Provider)
	}
}

func (h *Handler) deployToCloudflare(ctx context.Context, sub *storage.BundleSubscription, provider *storage.Provider, templates []*storage.BundleFunctionTemplate) {
	log := logrus.WithField("script_name", sub.ScriptName)

	// Get account ID from token, then recreate client with it
	tempClient := cloudflare.NewCloudflareDeploymentClient(provider.Token, "")
	accountID, err := tempClient.GetAccountID(ctx)
	if err != nil {
		h.failDeploy(sub, "failed to get CF account ID: %v", err)
		return
	}

	client := cloudflare.NewCloudflareDeploymentClient(provider.Token, accountID)

	// Get workers.dev subdomain
	subdomain, err := client.GetWorkersSubdomain(ctx)
	if err != nil {
		h.failDeploy(sub, "failed to get workers.dev subdomain: %v", err)
		return
	}

	// Create KV namespace for state storage
	kvTitle := sub.ScriptName + "-state"
	kvNamespaceID, err := client.CreateKVNamespace(ctx, kvTitle)
	if err != nil {
		h.failDeploy(sub, "failed to create KV namespace: %v", err)
		return
	}
	log.WithField("kv_namespace_id", kvNamespaceID).Info("Created KV namespace")

	// Generate combined worker script
	script := GenerateWorkerScript(templates)

	// Deploy with KV namespace binding
	result, err := client.DeployWithKVNamespace(ctx, script, sub.ScriptName, kvNamespaceID, "STATE")
	if err != nil {
		// Cleanup KV namespace on failure
		if delErr := client.DeleteKVNamespace(ctx, kvNamespaceID); delErr != nil {
			log.WithError(delErr).Warn("failed to delete KV namespace")
		}
		h.failDeploy(sub, "deploy failed: %v", err)
		return
	}
	if result.Status == "failed" {
		if delErr := client.DeleteKVNamespace(ctx, kvNamespaceID); delErr != nil {
			log.WithError(delErr).Warn("failed to delete KV namespace")
		}
		h.failDeploy(sub, "deploy failed: %s", result.Message)
		return
	}
	log.Info("Worker deployed")

	// Enable workers.dev
	if err := client.EnableWorkersDev(ctx, sub.ScriptName); err != nil {
		// Cleanup on failure
		if delErr := client.DeleteDeployment(ctx, sub.ScriptName); delErr != nil {
			log.WithError(delErr).Warn("failed to delete deployment")
		}
		if delErr := client.DeleteKVNamespace(ctx, kvNamespaceID); delErr != nil {
			log.WithError(delErr).Warn("failed to delete KV namespace")
		}
		h.failDeploy(sub, "enable workers.dev failed: %v", err)
		return
	}
	log.Info("workers.dev enabled")

	// Build worker URL
	workerURL := fmt.Sprintf("https://%s.%s.workers.dev", sub.ScriptName, subdomain)

	// Get or create app
	app, err := h.repo.GetAppBySlug(ctx, sub.BundleID.String())
	if err != nil || app == nil {
		h.failDeploy(sub, "app not found for bundle: %v", err)
		return
	}

	// Create backend
	_, err = h.repo.CreateBackend(ctx, app.ID, "workers", "us-east-1", workerURL, "", nil)
	if err != nil && !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") {
		h.failDeploy(sub, "create backend failed: %v", err)
		return
	}

	// Update subscription status
	sub.DeployStatus = "deployed"
	now := time.Now()
	sub.DeployedAt = &now
	sub.DeployError = ""
	if err := h.repo.UpdateBundleSubscription(ctx, sub); err != nil {
		log.WithError(err).Error("Failed to update subscription after deploy")
	}

	log.WithField("worker_url", workerURL).Info("Bundle deployment complete")
}

// failDeploy records a deployment failure and schedules retry.
func (h *Handler) failDeploy(sub *storage.BundleSubscription, format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	log := logrus.WithFields(logrus.Fields{
		"tenant_id":      sub.TenantID,
		"script_name":    sub.ScriptName,
		"deploy_attempts": sub.DeployAttempts,
	})
	log.WithError(fmt.Errorf("%s", errMsg)).Warn("Bundle deployment failed")

	sub.DeployStatus = "failed"
	sub.DeployAttempts++
	sub.DeployError = errMsg

	// Exponential backoff: 1min, 5min, 15min
	backoff := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	idx := sub.DeployAttempts - 1
	if idx >= len(backoff) {
		idx = len(backoff) - 1
	}
	nextRetry := time.Now().Add(backoff[idx])
	sub.NextRetryAt = &nextRetry

	if err := h.repo.UpdateBundleSubscription(context.Background(), sub); err != nil {
		log.WithError(err).Error("Failed to update subscription with deploy failure")
	}
}

// validateProviderAccess checks if the provider token has required permissions.
func (h *Handler) validateProviderAccess(ctx context.Context, provider *storage.Provider) error {
	switch provider.Provider {
	case "workers":
		client := cloudflare.NewCloudflareDeploymentClient(provider.Token, "")
		if _, err := client.GetAccountID(ctx); err != nil {
			return fmt.Errorf("invalid CF token: %w", err)
		}
		if _, err := client.GetWorkersSubdomain(ctx); err != nil {
			return fmt.Errorf("cannot access workers.dev subdomain: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", provider.Provider)
	}
}
