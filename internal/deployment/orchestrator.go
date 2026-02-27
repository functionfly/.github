package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// Orchestrator coordinates deployment operations across providers
type Orchestrator struct {
	repo            storage.Repository
	adapters        map[string]common.DeploymentAdapter
	store           ArtifactStore
	realtimeMonitor *monitoring.RealtimeMonitor
}

// ArtifactStore defines interface for storing and retrieving deployment artifacts
type ArtifactStore interface {
	Store(ctx context.Context, key string, data []byte) error
	Retrieve(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// NewOrchestrator creates a new deployment orchestrator
func NewOrchestrator(repo storage.Repository, adapters map[string]common.DeploymentAdapter, store ArtifactStore, realtimeMonitor *monitoring.RealtimeMonitor) *Orchestrator {
	return &Orchestrator{
		repo:            repo,
		adapters:        adapters,
		store:           store,
		realtimeMonitor: realtimeMonitor,
	}
}

// DeploySpec represents a deployment specification
type DeploySpec struct {
	AppID         uuid.UUID
	Provider      string
	Region        string
	AppName       string  // Standardized application name
	Environment   string  // Deployment environment (dev, staging, prod)
	Version       string  // Optional version identifier
	Artifact      []byte
	Routes        []string
	EnvVars       map[string]string
	Secrets       map[string]string
	ProviderConfig map[string]interface{}
	Timeout       *time.Duration // Deployment timeout
}

// DeployResult represents the result of a deployment operation
type DeployResult struct {
	DeploymentID uuid.UUID
	Status       string
	Message      string
}

// getProviderConfigFromDeployment extracts provider configuration from deployment metadata
func (o *Orchestrator) getProviderConfigFromDeployment(deploymentID uuid.UUID) (map[string]interface{}, error) {
	deployment, err := o.repo.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	if deployment == nil {
		return nil, fmt.Errorf("deployment not found")
	}

	// Parse the metadata JSON to extract provider config
	var metadata map[string]interface{}
	if deployment.Metadata != "" {
		if err := json.Unmarshal([]byte(deployment.Metadata), &metadata); err != nil {
			// If we can't parse metadata, return basic info as fallback
			return map[string]interface{}{
				"deployment_id": deployment.ID.String(),
				"provider":      deployment.Provider,
				"region":        deployment.Region,
			}, nil
		}
	}

	// Extract provider config from metadata
	if providerConfig, ok := metadata["provider_config"].(map[string]interface{}); ok {
		// Add basic deployment info to provider config
		providerConfig["deployment_id"] = deployment.ID.String()
		providerConfig["provider"] = deployment.Provider
		providerConfig["region"] = deployment.Region
		return providerConfig, nil
	}

	// Fallback: return basic deployment info if no provider config found
	return map[string]interface{}{
		"deployment_id": deployment.ID.String(),
		"provider":      deployment.Provider,
		"region":        deployment.Region,
	}, nil
}

// Deploy initiates a deployment to the specified provider
func (o *Orchestrator) Deploy(ctx context.Context, spec *DeploySpec) (*DeployResult, error) {
	// Get the adapter for this provider
	adapter, ok := o.adapters[spec.Provider]
	if !ok {
		return nil, fmt.Errorf("no adapter available for provider: %s", spec.Provider)
	}

	// Generate artifact key
	artifactKey := fmt.Sprintf("deployments/%s/%s/%d.js", spec.AppID, spec.Provider, time.Now().Unix())

	// Store the artifact
	if err := o.store.Store(ctx, artifactKey, spec.Artifact); err != nil {
		return nil, fmt.Errorf("failed to store artifact: %w", err)
	}

	// Record the artifact in database
	if _, err := o.repo.StoreDeploymentArtifact(spec.AppID, spec.Provider, artifactKey, "application/javascript", "", int64(len(spec.Artifact))); err != nil {
		// Clean up stored artifact on failure
		o.store.Delete(ctx, artifactKey)
		return nil, fmt.Errorf("failed to record artifact: %w", err)
	}

	// Create deployment record with provider config in metadata
	metadata := map[string]interface{}{
		"provider_config": spec.ProviderConfig,
	}
	deployment, err := o.repo.CreateDeployment(spec.AppID, spec.Provider, spec.Region, "", artifactKey, spec.Routes)
	if err != nil {
		// Clean up stored artifact on failure
		o.store.Delete(ctx, artifactKey)
		return nil, fmt.Errorf("failed to create deployment record: %w", err)
	}

	// Update status to deploying with metadata
	if err := o.repo.UpdateDeploymentStatus(deployment.ID, "deploying", "Starting deployment", metadata); err != nil {
		return nil, fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Get tenant ID for real-time broadcast
	app, _ := o.repo.GetAppByID(spec.AppID)
	var tenantID *uuid.UUID
	if app != nil {
		tenantID = &app.TenantID
	}

	// Broadcast deployment started
	o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, deployment.ID, "deploying", map[string]interface{}{
		"message": "Starting deployment",
		"app_id":  spec.AppID,
	})

	// Prepare deployment spec for adapter
	deploymentSpec := &common.DeploymentSpec{
		Artifact:       spec.Artifact,
		ArtifactKey:    artifactKey,
		AppName:        spec.AppName,
		Environment:    spec.Environment,
		Version:        spec.Version,
		Routes:         spec.Routes,
		EnvVars:        spec.EnvVars,
		Secrets:        spec.Secrets,
		ProviderConfig: spec.ProviderConfig,
		Timeout:        spec.Timeout,
	}

	// Execute deployment
	result, err := adapter.Deploy(ctx, deploymentSpec)
	if err != nil {
		// Update status to failed
		o.repo.UpdateDeploymentStatus(deployment.ID, "failed", fmt.Sprintf("Deployment failed: %v", err), nil)

		// Broadcast deployment failed
		o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, deployment.ID, "failed", map[string]interface{}{
			"message": fmt.Sprintf("Deployment failed: %v", err),
			"app_id":  spec.AppID,
			"error":   err.Error(),
		})

		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	// Update deployment record with provider deployment ID and success status
	// Preserve the provider config from initial metadata
	resultMetadata := result.Metadata
	if resultMetadata == nil {
		resultMetadata = make(map[string]interface{})
	}

	// Merge provider config with result metadata
	if providerConfig, ok := metadata["provider_config"]; ok {
		resultMetadata["provider_config"] = providerConfig
	}

	if err := o.repo.UpdateDeploymentStatus(deployment.ID, string(result.Status), result.Message, resultMetadata); err != nil {
		return nil, fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Broadcast deployment completed
	o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, deployment.ID, string(result.Status), map[string]interface{}{
		"message":        result.Message,
		"app_id":         spec.AppID,
		"deployment_url": result.DeploymentURL,
	})

	return &DeployResult{
		DeploymentID: deployment.ID,
		Status:       string(result.Status),
		Message:      result.Message,
	}, nil
}

// GetDeploymentStatus returns the current status of a deployment
func (o *Orchestrator) GetDeploymentStatus(ctx context.Context, deploymentID uuid.UUID) (*storage.Deployment, error) {
	deployment, err := o.repo.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	if deployment == nil {
		return nil, fmt.Errorf("deployment not found")
	}

	// If deployment is still in progress, check with provider
	if deployment.Status == "deploying" || deployment.Status == "pending" {
		adapter, ok := o.adapters[deployment.Provider]
		if ok {
			providerConfig, _ := o.getProviderConfigFromDeployment(deploymentID)
			status, err := adapter.GetDeploymentStatus(ctx, deployment.DeploymentID, providerConfig)
			if err == nil {
				// Update status if it has changed
				if string(status) != deployment.Status {
					o.repo.UpdateDeploymentStatus(deploymentID, string(status), "Status updated from provider", nil)
					deployment.Status = string(status)
				}
			}
		}
	}

	return deployment, nil
}

// ListDeployments returns deployments for an app
func (o *Orchestrator) ListDeployments(ctx context.Context, appID uuid.UUID, limit int) ([]*storage.Deployment, error) {
	if limit <= 0 {
		limit = 10
	}
	return o.repo.ListDeploymentsByAppID(appID, limit)
}

// Rollback rolls back to a previous deployment
func (o *Orchestrator) Rollback(ctx context.Context, appID uuid.UUID, toDeploymentID uuid.UUID) (*DeployResult, error) {
	// Get the target deployment to rollback to
	targetDeployment, err := o.repo.GetDeploymentByID(toDeploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target deployment: %w", err)
	}
	if targetDeployment == nil {
		return nil, fmt.Errorf("target deployment not found")
	}

	// Verify the deployment belongs to the app
	if targetDeployment.AppID != appID {
		return nil, fmt.Errorf("deployment does not belong to app")
	}

	// Get tenant ID for real-time broadcast
	app, _ := o.repo.GetAppByID(appID)
	var tenantID *uuid.UUID
	if app != nil {
		tenantID = &app.TenantID
	}

	// Broadcast rollback started
	o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, toDeploymentID, "rollback_started", map[string]interface{}{
		"message": "Starting rollback to previous deployment",
		"app_id":  appID,
	})

	// Get adapter
	adapter, ok := o.adapters[targetDeployment.Provider]
	if !ok {
		return nil, fmt.Errorf("no adapter available for provider: %s", targetDeployment.Provider)
	}

	// Retrieve the artifact for rollback
	artifact, err := o.store.Retrieve(ctx, targetDeployment.ArtifactKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve artifact for rollback: %w", err)
	}

	// Get provider config
	providerConfig, _ := o.getProviderConfigFromDeployment(toDeploymentID)

	// Create deployment spec for rollback
	// Extract app name from provider config or use a default
	appName := "rollback-app"
	if name, ok := providerConfig["app_name"].(string); ok && name != "" {
		appName = name
	}

	deploymentSpec := &common.DeploymentSpec{
		Artifact:       artifact,
		ArtifactKey:    targetDeployment.ArtifactKey,
		AppName:        appName,
		Environment:    "rollback", // Special environment for rollbacks
		Routes:         targetDeployment.Routes,
		ProviderConfig: providerConfig,
	}

	// Perform rollback deployment
	result, err := adapter.Rollback(ctx, deploymentSpec)
	if err != nil {
		// Create a new deployment record for the rollback attempt
		newDeployment, createErr := o.repo.CreateDeployment(appID, targetDeployment.Provider, targetDeployment.Region, "", targetDeployment.ArtifactKey, targetDeployment.Routes)
		if createErr == nil {
			o.repo.UpdateDeploymentStatus(newDeployment.ID, "failed", fmt.Sprintf("Rollback failed: %v", err), nil)

			// Broadcast rollback failed
			o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, newDeployment.ID, "rollback_failed", map[string]interface{}{
				"message": fmt.Sprintf("Rollback failed: %v", err),
				"app_id":  appID,
				"error":   err.Error(),
			})
		}
		return nil, fmt.Errorf("rollback failed: %w", err)
	}

	// Create deployment record for successful rollback
	newDeployment, err := o.repo.CreateDeployment(appID, targetDeployment.Provider, targetDeployment.Region, result.DeploymentID, targetDeployment.ArtifactKey, targetDeployment.Routes)
	if err != nil {
		return nil, fmt.Errorf("failed to record rollback deployment: %w", err)
	}

	if err := o.repo.UpdateDeploymentStatus(newDeployment.ID, string(result.Status), "Rollback completed: "+result.Message, result.Metadata); err != nil {
		return nil, fmt.Errorf("failed to update rollback status: %w", err)
	}

	// Broadcast rollback completed
	o.realtimeMonitor.BroadcastDeploymentUpdate(tenantID, newDeployment.ID, "rollback_completed", map[string]interface{}{
		"message":        "Rollback completed: " + result.Message,
		"app_id":         appID,
		"deployment_url": result.DeploymentURL,
		"rollback_to":    toDeploymentID,
	})

	return &DeployResult{
		DeploymentID: newDeployment.ID,
		Status:       string(result.Status),
		Message:      "Rollback completed: " + result.Message,
	}, nil
}