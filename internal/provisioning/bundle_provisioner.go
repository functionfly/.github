package provisioning

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProvisionerStatus represents the provisioning state of a single component
type ProvisionerStatus string

const (
	StatusPending      ProvisionerStatus = "pending"
	StatusProvisioning ProvisionerStatus = "provisioning"
	StatusActive       ProvisionerStatus = "active"
	StatusFailed       ProvisionerStatus = "failed"
	StatusRolledBack   ProvisionerStatus = "rolled_back"
)

// ComponentState tracks individual component provisioning state
type ComponentState struct {
	Status     ProvisionerStatus `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Error      string            `json:"error,omitempty"`
	ResourceID string            `json:"resource_id,omitempty"` // ID of provisioned resource
}

// ProvisionResult is the final result of a bundle provisioning run
type ProvisionResult struct {
	TenantID   uuid.UUID                  `json:"tenant_id"`
	BundleSlug string                     `json:"bundle_slug"`
	Status     ProvisionerStatus          `json:"status"`
	Components map[string]*ComponentState `json:"components"`
	StartedAt  time.Time                  `json:"started_at"`
	FinishedAt time.Time                  `json:"finished_at"`
	Duration   int64                      `json:"duration_ms"`
	ErrorLog   []string                   `json:"error_log,omitempty"`
}

// BundleProvisioner orchestrates the full one-click provisioning of an isolated
// SaaS Starter bundle. Each component (Auth, Payments, UserDB, Email, Analytics)
// is provisioned in isolation inside the tenant's own dedicated database.
//
// Design principles:
//   - Idempotent: safe to retry any step
//   - Atomic per component: each provisioner either fully succeeds or rolls back
//   - No platform DB leakage: all tenant data goes to the tenant's own DB
//   - Observable: every step is logged with structured fields
type BundleProvisioner struct {
	platformDB     *sql.DB               // Platform DB for tracking state
	platformRepo   storage.Repository    // Repository for user/tenant lookups
	dbProvisioner  *storage.TenantDBProvisioner
	emailService   email.Service

	// Per-component provisioners (shared across all bundles)
	authProvisioner      *AuthProvisioner
	paymentsProvisioner  *PaymentsProvisioner
	emailWfProvisioner   *EmailWorkflowProvisioner
	analyticsProvisioner *AnalyticsProvisioner
	userDBProvisioner    *UserDBProvisioner

	// Bundle-specific provisioners
	marketplaceProvisioner *MarketplaceProvisioner
	aiAppProvisioner       *AIAppProvisioner

	// External infrastructure provisioners
	externalAIProvisioner *ExternalAIProvisioner

	mu sync.Mutex
}

// NewBundleProvisioner creates the orchestrator with all sub-provisioners.
// platformDB is the raw *sql.DB for the platform database (used for tracking state).
// platformRepo is the Repository interface (used for user/tenant lookups).
// dbProvisioner manages per-tenant dedicated databases.
func NewBundleProvisioner(
	platformDB *sql.DB,
	platformRepo storage.Repository,
	dbProvisioner *storage.TenantDBProvisioner,
	emailService email.Service,
) *BundleProvisioner {
	return &BundleProvisioner{
		platformDB:     platformDB,
		platformRepo:   platformRepo,
		dbProvisioner:  dbProvisioner,
		emailService:   emailService,
		authProvisioner:      NewAuthProvisioner(platformRepo, dbProvisioner),
		paymentsProvisioner:  NewPaymentsProvisioner(platformRepo, dbProvisioner),
		emailWfProvisioner:   NewEmailWorkflowProvisioner(platformRepo, dbProvisioner, emailService),
		analyticsProvisioner: NewAnalyticsProvisioner(platformRepo, dbProvisioner),
		userDBProvisioner:    NewUserDBProvisioner(platformRepo, dbProvisioner),
		marketplaceProvisioner: NewMarketplaceProvisioner(platformRepo, dbProvisioner),
		aiAppProvisioner:       NewAIAppProvisioner(platformRepo, dbProvisioner),
		externalAIProvisioner:  NewExternalAIProvisioner(platformDB, platformRepo, dbProvisioner),
	}
}

// ProvisionBundle executes the full one-click provisioning pipeline for a bundle.
// The execution order is:
//
//	1. User DB       — Create dedicated database + apply migrations
//	2. Auth          — Generate JWT keys, seed OAuth configs, create sessions table
//	3. Payments      — Create Stripe customer, webhook endpoint, seed products/prices
//	4. Email         — Seed email templates, create default workflows (welcome, onboarding, dunning)
//	5. Analytics     — Create default dashboards, seed funnel definitions
//
// Each step is idempotent. If a step fails, previously completed steps are NOT rolled back
// (they can be retried independently). The overall status is 'active' only if all steps succeed.
func (bp *BundleProvisioner) ProvisionBundle(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ProvisionResult, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	startTime := time.Now()

	result := &ProvisionResult{
		TenantID:   tenantID,
		BundleSlug: bundleSlug,
		Status:     StatusProvisioning,
		Components: make(map[string]*ComponentState),
		StartedAt:  startTime,
		ErrorLog:   []string{},
	}

	log := logrus.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"bundle_slug": bundleSlug,
		"operation":   "bundle_provision",
	})

	log.Info("Starting one-click bundle provisioning")

	// Track provisioning state in platform DB
	if err := bp.trackProvisioningStart(ctx, tenantID, bundleSlug); err != nil {
		log.WithError(err).Warn("Failed to track provisioning state (non-fatal)")
	}

	// Define the provisioning pipeline (bundle-aware)
	type pipelineStep struct {
		name      string
		provision func(context.Context, uuid.UUID, string) (*ComponentState, error)
	}

	// Check if tenant DBs are disabled (local dev / shared platform mode)
	sharedMode := bp.dbProvisioner == nil || !bp.dbProvisioner.IsEnabled()
	if sharedMode {
		log.Info("Tenant databases disabled — using shared platform mode")
	}

	// Base steps (all bundles get these)
	steps := []pipelineStep{
		{"user_db", bp.userDBProvisioner.Provision},
		{"auth", bp.authProvisioner.Provision},
		{"payments", bp.paymentsProvisioner.Provision},
		{"email_workflows", bp.emailWfProvisioner.Provision},
		{"analytics", bp.analyticsProvisioner.Provision},
	}

	// Bundle-specific steps
	switch bundleSlug {
	case "marketplace":
		steps = append(steps, pipelineStep{"marketplace", bp.marketplaceProvisioner.Provision})
	case "saas-starter":
		// SaaS Starter uses the base steps above (auth, payments, email, analytics are sufficient)
	case "ai-app":
		// External AI DB (Neon serverless) — paid plans only, scales to zero
		if bp.externalAIProvisioner.IsAvailable() {
			steps = append(steps, pipelineStep{"external_ai_db", bp.provisionExternalAIDB})
		}
		steps = append(steps, pipelineStep{"ai_app", bp.aiAppProvisioner.Provision})
	}

	allSucceeded := true

	for i, step := range steps {
		stepLog := log.WithField("component", step.name)
		stepLog.Info("Provisioning component")

		// Mark component as provisioning in result and persist to DB for polling
		result.Components[step.name] = &ComponentState{Status: StatusProvisioning, Timestamp: time.Now()}
		bp.trackProvisioningProgress(ctx, tenantID, result)

		// Record step start in deployment_steps
		stepID := uuid.New()
		bp.recordDeploymentStep(ctx, tenantID, stepID, bundleSlug, step.name, i, "running")

		var state *ComponentState
		var err error

		if sharedMode {
			// Shared platform mode: mark as active using shared infrastructure
			state = &ComponentState{
				Status:     StatusActive,
				Timestamp:  time.Now(),
				ResourceID: "shared-platform",
			}
			stepLog.Info("Component skipped (shared platform mode)")
		} else {
			state, err = step.provision(ctx, tenantID, bundleSlug)
			if state == nil {
				state = &ComponentState{Status: StatusFailed}
			}
		}

		result.Components[step.name] = state

		if err != nil {
			state.Status = StatusFailed
			state.Error = err.Error()
			allSucceeded = false
			result.ErrorLog = append(result.ErrorLog, fmt.Sprintf("%s: %s", step.name, err.Error()))
			stepLog.WithError(err).Error("Component provisioning failed")
			bp.recordDeploymentStep(ctx, tenantID, stepID, bundleSlug, step.name, i, "failed")
			// Continue to next component — each is independent
		} else {
			state.Status = StatusActive
			state.Timestamp = time.Now()
			stepLog.Info("Component provisioned successfully")
			bp.recordDeploymentStep(ctx, tenantID, stepID, bundleSlug, step.name, i, "completed")
		}

		// Persist component result to DB so polling endpoint returns real-time progress
		bp.trackProvisioningProgress(ctx, tenantID, result)
	}

	// Finalize
	endTime := time.Now()
	result.FinishedAt = endTime
	result.Duration = endTime.Sub(startTime).Milliseconds()

	if allSucceeded {
		result.Status = StatusActive
		log.WithField("duration_ms", result.Duration).Info("Bundle provisioning complete — all components active")
	} else {
		result.Status = StatusFailed
		log.WithFields(logrus.Fields{
			"duration_ms": result.Duration,
			"errors":      len(result.ErrorLog),
		}).Warn("Bundle provisioning completed with errors")
	}

	// Update tracking
	if err := bp.trackProvisioningComplete(ctx, tenantID, result); err != nil {
		log.WithError(err).Warn("Failed to update provisioning state (non-fatal)")
	}

	return result, nil
}

// TenantDB returns the tenant database provisioner for direct tenant DB access.
func (bp *BundleProvisioner) TenantDB() *storage.TenantDBProvisioner {
	return bp.dbProvisioner
}

// GetProvisioningStatus returns the current provisioning state for a tenant
func (bp *BundleProvisioner) GetProvisioningStatus(ctx context.Context, tenantID uuid.UUID) (*ProvisionResult, error) {
	if bp.platformDB == nil {
		return nil, fmt.Errorf("platform DB not available")
	}

	row := bp.platformDB.QueryRowContext(ctx,
		`SELECT bundle_slug, provision_status, components, error_log, created_at, updated_at
		 FROM tenant_bundle_state WHERE tenant_id = $1`, tenantID)

	var result ProvisionResult
	var componentsJSON, errorLogJSON []byte
	var createdAt, updatedAt time.Time

	err := row.Scan(&result.BundleSlug, &result.Status, &componentsJSON, &errorLogJSON, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("no provisioning state found for tenant %s: %w", tenantID, err)
	}

	result.TenantID = tenantID
	result.Components = make(map[string]*ComponentState)
	if len(componentsJSON) > 0 {
		json.Unmarshal(componentsJSON, &result.Components)
	}
	if len(errorLogJSON) > 0 {
		json.Unmarshal(errorLogJSON, &result.ErrorLog)
	}
	result.StartedAt = createdAt
	result.FinishedAt = updatedAt
	result.Duration = updatedAt.Sub(createdAt).Milliseconds()

	return &result, nil
}

// trackProvisioningStart records that provisioning has begun
func (bp *BundleProvisioner) trackProvisioningStart(ctx context.Context, tenantID uuid.UUID, bundleSlug string) error {
	if bp.platformDB == nil {
		return fmt.Errorf("platform DB not available")
	}

	_, err := bp.platformDB.ExecContext(ctx,
		`INSERT INTO tenant_bundle_state (id, tenant_id, bundle_slug, provision_status, components, created_at, updated_at)
		 VALUES ($1, $2, $3, 'provisioning', '{}', NOW(), NOW())
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 	provision_status = 'provisioning',
		 	error_log = '[]',
		 	updated_at = NOW()`,
		uuid.New(), tenantID, bundleSlug)
	return err
}

// trackProvisioningComplete records the final provisioning state
func (bp *BundleProvisioner) trackProvisioningComplete(ctx context.Context, tenantID uuid.UUID, result *ProvisionResult) error {
	if bp.platformDB == nil {
		return fmt.Errorf("platform DB not available")
	}

	componentsJSON, _ := json.Marshal(result.Components)
	errorLogJSON, _ := json.Marshal(result.ErrorLog)
	isActive := string(result.Status) == "active"

	_, err := bp.platformDB.ExecContext(ctx,
		`UPDATE tenant_bundle_state SET
		 	provision_status = $1,
		 	components = $2::jsonb,
		 	error_log = $3::jsonb,
		 	provisioned_at = CASE WHEN $5 THEN NOW() ELSE provisioned_at END,
		 	updated_at = NOW()
		 WHERE tenant_id = $4`,
		string(result.Status), string(componentsJSON), string(errorLogJSON), tenantID.String(), isActive)
	return err
}

// trackProvisioningProgress updates the DB with current component states during provisioning.
// This enables real-time polling from the frontend to show loading spinners and checkmarks.
func (bp *BundleProvisioner) trackProvisioningProgress(ctx context.Context, tenantID uuid.UUID, result *ProvisionResult) {
	if bp.platformDB == nil {
		return
	}

	componentsJSON, _ := json.Marshal(result.Components)
	_, err := bp.platformDB.ExecContext(ctx,
		`UPDATE tenant_bundle_state SET
			components = $1::jsonb,
			updated_at = NOW()
		 WHERE tenant_id = $2`,
		string(componentsJSON), tenantID.String())
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to track provisioning progress (non-fatal)")
	}
}

// provisionExternalAIDB is the pipeline step that provisions an external Neon database
// for AI data (vectors, embeddings, documents, memory). Only runs for paid plans.
// Free-tier tenants use the local tenant DB with hard limits instead.
func (bp *BundleProvisioner) provisionExternalAIDB(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "external_ai_db",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: time.Now(),
	}

	// Check tenant plan — free tier does NOT get external DB
	tenant, err := bp.platformRepo.GetTenantByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return state, fmt.Errorf("tenant not found: %w", err)
	}

	// Free/Community plans: skip external DB (use local with limits)
	isPaidPlan := tenant.Plan != "" &&
		tenant.Plan != "free" &&
		tenant.Plan != "community" &&
		tenant.Plan != "founder"

	if !isPaidPlan {
		log.WithField("plan", tenant.Plan).Info("Free plan — skipping external AI DB, using local with limits")
		pool, err := bp.dbProvisioner.GetTenantPool(ctx, tenantID)
		if err == nil {
			pool.Exec(ctx,
				`UPDATE tenant_configs SET settings = settings || '{"ai":{"storage":"local","max_vectors":1000,"max_documents":10}}' WHERE tenant_id = $1`,
				tenantID)
		}
		state.Status = StatusActive
		state.ResourceID = "local:shared"
		return state, nil
	}

	// Paid plan: provision external Neon DB — tenant pays for it
	log.WithField("plan", tenant.Plan).Info("Paid plan — provisioning external AI database on Neon")

	conn, pool, err := bp.externalAIProvisioner.ProvisionExternalAIDB(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to provision external AI DB: %w", err)
	}
	defer pool.Close()

	// Set config flag so AI provisioner knows to use external DB
	localPool, err := bp.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err == nil {
		localPool.Exec(ctx,
			`UPDATE tenant_configs SET settings = settings || '{"ai":{"storage":"external","provider":"neon","max_vectors":1000000,"max_documents":10000}}' WHERE tenant_id = $1`,
			tenantID)
	}

	state.Status = StatusActive
	state.ResourceID = fmt.Sprintf("neon:%s", conn.BranchID)
	log.WithField("branch_id", conn.BranchID).Info("External AI database provisioned")
	return state, nil
}

// recordDeploymentStep writes or updates a deployment step record in the platform DB.
// This enables real-time provisioning status via the deployment status endpoint.
func (bp *BundleProvisioner) recordDeploymentStep(ctx context.Context, tenantID, stepID uuid.UUID, bundleSlug, stepName string, stepOrder int, status string) {
	if bp.platformDB == nil {
		return
	}

	now := time.Now()
	_, err := bp.platformDB.ExecContext(ctx,
		`INSERT INTO deployment_steps (id, tenant_id, deployment_id, bundle_slug, step_name, step_order, status, started_at, completed_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (id) DO UPDATE SET status = $7, completed_at = $9, updated_at = $11`,
		stepID, tenantID, uuid.Nil, bundleSlug, stepName, stepOrder, status,
		now, // started_at
		func() interface{} {
			if status == "completed" || status == "failed" {
				return now
			}
			return nil
		}(),
		now, now)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"step":      stepName,
			"status":    status,
		}).Warn("Failed to record deployment step (non-fatal)")
	}
}
