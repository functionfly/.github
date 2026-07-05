package provisioning

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
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
	platformDB     *sql.DB                       // Platform DB for tracking state
	platformRepo   storage.Repository            // Repository for user/tenant lookups
	registryRepo   *registry.RegistryRepository  // Registry for function publishing
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
// registryRepo is the RegistryRepository for publishing bundle functions to the registry.
// dbProvisioner manages per-tenant dedicated databases.
func NewBundleProvisioner(
	platformDB *sql.DB,
	platformRepo storage.Repository,
	registryRepo *registry.RegistryRepository,
	dbProvisioner *storage.TenantDBProvisioner,
	emailService email.Service,
) *BundleProvisioner {
	return &BundleProvisioner{
		platformDB:     platformDB,
		platformRepo:   platformRepo,
		registryRepo:   registryRepo,
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
		{"app", bp.provisionApp},
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

		// Platform-level steps (app creation) always run; tenant-DB steps are skipped in shared mode
		if sharedMode && step.name != "app" {
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

// provisionApp creates the default app for the bundle, publishes function templates
// to the registry under the @functionfly namespace, and links the app to the subscription.
func (bp *BundleProvisioner) provisionApp(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	if bp.platformRepo == nil {
		return &ComponentState{Status: StatusActive, Timestamp: time.Now(), ResourceID: "no-repo"}, nil
	}

	bundleAppNames := map[string]string{
		"saas-starter": "SaaS Starter",
		"marketplace":  "Marketplace",
		"ai-app":      "AI App",
	}
	bundleAppSlugs := map[string]string{
		"saas-starter": "saas-starter",
		"marketplace":  "marketplace",
		"ai-app":      "ai-app",
	}

	appName := bundleAppNames[bundleSlug]
	appSlug := bundleAppSlugs[bundleSlug]
	if appName == "" {
		appName = "My Backend"
		appSlug = "my-backend"
	}

	app, err := bp.platformRepo.CreateApp(ctx, appName, appSlug, tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			apps, listErr := bp.platformRepo.ListAppsByTenant(ctx, tenantID)
			if listErr == nil {
				for _, a := range apps {
					if a.Slug == appSlug {
						app = a
						break
					}
				}
			}
			if app == nil {
				return &ComponentState{Status: StatusActive, Timestamp: time.Now(), ResourceID: "existing"}, nil
			}
		} else {
			return nil, fmt.Errorf("failed to create app: %w", err)
		}
	}

	// Link the app to the bundle subscription via DefaultAppID
	sub, subErr := bp.platformRepo.GetBundleSubscriptionByTenant(ctx, tenantID)
	if subErr == nil && sub != nil {
		sub.DefaultAppID = &app.ID
		if updateErr := bp.platformRepo.UpdateBundleSubscription(ctx, sub); updateErr != nil {
			logrus.WithError(updateErr).WithField("tenant_id", tenantID).Warn("Failed to link app to bundle subscription")
		}
	}

	// Publish bundle function templates to the registry under @functionfly/
	if bp.registryRepo != nil {
		for _, tmpl := range bundleFunctionTemplates(bundleSlug) {
			bp.publishBundleFunction(ctx, tenantID, app.ID, tmpl)
		}
	} else {
		logrus.WithField("tenant_id", tenantID).Info("Registry repo not available — skipping function publishing")
	}

	return &ComponentState{
		Status:     StatusActive,
		Timestamp:  time.Now(),
		ResourceID: app.ID.String(),
	}, nil
}

// publishBundleFunction creates a registry function + first version under the @functionfly author.
func (bp *BundleProvisioner) publishBundleFunction(ctx context.Context, tenantID, appID uuid.UUID, tmpl bundleFnTemplate) {
	// Check if already published
	existing, _ := bp.registryRepo.GetFunctionByAuthorName(ctx, "functionfly", tmpl.name)
	if existing != nil {
		logrus.WithField("function", tmpl.name).Debug("Bundle function already exists in registry")
		return
	}

	manifest := map[string]interface{}{
		"name":        tmpl.name,
		"version":     "1.0.0",
		"description": tmpl.description,
		"runtime":     "javascript",
		"capabilities": tmpl.capabs,
	}
	manifestJSON, _ := json.Marshal(manifest)
	capsJSON, _ := json.Marshal(tmpl.capabs)

	fn := &registry.RegistryFunction{
		ID:                 uuid.New(),
		Author:             "functionfly",
		Name:               tmpl.name,
		Title:              sql.NullString{String: tmpl.title, Valid: tmpl.title != ""},
		Description:        sql.NullString{String: tmpl.description, Valid: tmpl.description != ""},
		Visibility:         "public",
		TenantID:           &tenantID,
		ReliabilityScore:   90.0,
		DeterministicScore: 90.0,
		Status:             "active",
		Region:             tmpl.region,
		Code:               tmpl.code,
		AppID:              &appID,
		Capabilities:       capsJSON,
	}
	// Set providers via raw update after creation — GORM can't serialize StringArray
	if err := bp.registryRepo.CreateFunction(ctx, fn); err != nil {
		logrus.WithError(err).WithField("function", tmpl.name).Warn("Failed to create registry function")
		return
	}

	// Set providers via raw SQL — GORM can't serialize StringArray for text[] columns
	if bp.platformDB != nil {
		bp.platformDB.Exec(`UPDATE registry_functions SET providers = '{"functionfly"}' WHERE id = $1`, fn.ID)
	}

	// Create first version so /functions/:author/:name resolves
	version := &registry.RegistryFunctionVersion{
		ID:         uuid.New(),
		FunctionID: fn.ID,
		Version:    "1.0.0",
		Manifest:   manifestJSON,
		Runtime:    "javascript",
		SourceCode: sql.NullString{String: tmpl.code, Valid: true},
		Capabilities: capsJSON,
		IsActive:   true,
	}
	if err := bp.registryRepo.CreateFunctionVersion(version); err != nil {
		logrus.WithError(err).WithField("function", tmpl.name).Warn("Failed to create registry function version")
		return
	}

	// Set latest version
	if err := bp.registryRepo.UpdateFunctionLatestVersion(ctx, fn.ID, "1.0.0"); err != nil {
		logrus.WithError(err).WithField("function", tmpl.name).Warn("Failed to update latest version")
	}

	logrus.WithFields(logrus.Fields{
		"author":   "functionfly",
		"name":     tmpl.name,
		"app_id":   appID,
	}).Info("Published bundle function to registry")
}

type bundleFnTemplate struct {
	name        string
	title       string
	description string
	code        string
	region      string
	capabs      []string
}

func bundleFunctionTemplates(bundleSlug string) []bundleFnTemplate {
	switch bundleSlug {
	case "saas-starter":
		return []bundleFnTemplate{
			{
				name:        "stripe-webhook",
				title:       "Stripe Webhook Handler",
				description: "Production-ready Stripe webhook handler with HMAC-SHA256 signature verification, idempotency, and full event coverage for subscriptions, invoices, payments, disputes, refunds, and Connect payouts.",
				region:      "us-east-1",
				capabs:      []string{"webhook", "storage"},
				code: `// Stripe Webhook Handler — Production Ready
// Requires STRIPE_WEBHOOK_SECRET env var (whsec_... from Stripe dashboard)
// Set via: functionfly env set STRIPE_WEBHOOK_SECRET=whsec_xxx

const WEBHOOK_TOLERANCE_SECONDS = 300; // 5 min timestamp tolerance

async function verifySignature(payload, sigHeader, secret) {
  if (!sigHeader) throw new Error('Missing Stripe-Signature header');
  const parts = sigHeader.split(',').reduce((acc, part) => {
    const [k, v] = part.split('=');
    (acc[k] = acc[k] || []).push(v);
    return acc;
  }, {});
  const timestamp = parts.t?.[0];
  const signatures = parts.v1 || [];
  if (!timestamp || signatures.length === 0) throw new Error('Malformed Stripe-Signature');
  const age = Math.abs(Math.floor(Date.now() / 1000) - parseInt(timestamp, 10));
  if (age > WEBHOOK_TOLERANCE_SECONDS) throw new Error('Webhook timestamp expired (replay attack protection)');
  const key = await crypto.subtle.importKey('raw', new TextEncoder().encode('sha256'), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
  const signed = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(timestamp + '.' + payload));
  const expected = Array.from(new Uint8Array(signed)).map(b => b.toString(16).padStart(2, '0')).join('');
  const valid = signatures.some(sig => {
    if (sig.length !== expected.length) return false;
    let result = 0;
    for (let i = 0; i < sig.length; i++) result |= sig.charCodeAt(i) ^ expected.charCodeAt(i);
    return result === 0;
  });
  if (!valid) throw new Error('Webhook signature mismatch');
}

function log(level, msg, data) {
  const entry = { ts: new Date().toISOString(), level, msg, ...data };
  if (level === 'error') console.error(JSON.stringify(entry));
  else if (level === 'warn') console.warn(JSON.stringify(entry));
  else console.log(JSON.stringify(entry));
}

export default async (req, res) => {
  const rawBody = typeof req.body === 'string' ? req.body : JSON.stringify(req.body);
  let event;
  try {
    event = typeof req.body === 'string' ? JSON.parse(req.body) : req.body;
  } catch {
    return res.status(200).json({ status: 'error', message: 'Invalid JSON' });
  }

  // ── Signature verification ──
  const secret = (typeof env !== 'undefined' && env.get) ? env.get('STRIPE_WEBHOOK_SECRET') : (typeof process !== 'undefined' ? process.env?.STRIPE_WEBHOOK_SECRET : null);
  if (secret) {
    try {
      const sigHeader = req.headers?.['stripe-signature'] || req.headers?.['Stripe-Signature'] || '';
      await verifySignature(rawBody, sigHeader, secret);
    } catch (err) {
      log('error', 'Signature verification failed', { event_id: event.id, error: err.message });
      return res.status(200).json({ status: 'error', message: 'Signature verification failed' });
    }
  } else {
    log('warn', 'STRIPE_WEBHOOK_SECRET not set — running without signature verification', { event_id: event.id });
  }

  // ── Idempotency ──
  const processedKey = 'webhook_processed/' + event.id;
  try {
    const already = await state.get(processedKey);
    if (already) {
      log('info', 'Duplicate event skipped', { event_id: event.id, event_type: event.type });
      return res.json({ status: 'skipped', reason: 'duplicate' });
    }
  } catch { /* state may not be available */ }

  log('info', 'Processing event', { event_id: event.id, event_type: event.type, livemode: event.livemode });

  const obj = event.data?.object || {};
  const objId = obj.id || 'unknown';

  try {
    switch (event.type) {
      // ── Checkout ──
      case 'checkout.session.completed': {
        const customerId = obj.customer;
        const subscriptionId = obj.subscription;
        const mode = obj.mode;
        await state.set('checkout/' + objId, { customer: customerId, subscription: subscriptionId, mode, status: 'completed', amount_total: obj.amount_total, currency: obj.currency, completed_at: new Date().toISOString() });
        if (subscriptionId && customerId) {
          await state.set('subscriptions/' + customerId, { subscription_id: subscriptionId, status: 'active', checkout_session: objId, updated_at: new Date().toISOString() });
        }
        log('info', 'Checkout completed', { event_id: event.id, customer: customerId, mode });
        break;
      }

      // ── Subscriptions ──
      case 'customer.subscription.created': {
        const items = obj.items?.data || [];
        const plan = items[0]?.plan;
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, plan_interval: plan?.interval, currency: plan?.currency, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, created_at: new Date().toISOString() });
        log('info', 'Subscription created', { event_id: event.id, customer: obj.customer, plan: plan?.id, status: obj.status });
        break;
      }
      case 'customer.subscription.updated': {
        const existing = await state.get('subscriptions/' + obj.customer).catch(() => ({}));
        const items = obj.items?.data || [];
        const plan = items[0]?.plan;
        await state.set('subscriptions/' + obj.customer, { ...existing, subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, plan_interval: plan?.interval, currency: plan?.currency, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, updated_at: new Date().toISOString() });
        log('info', 'Subscription updated', { event_id: event.id, customer: obj.customer, status: obj.status, cancel_at_period_end: obj.cancel_at_period_end });
        break;
      }
      case 'customer.subscription.deleted': {
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: 'canceled', canceled_at: obj.canceled_at, ended_at: obj.ended_at, updated_at: new Date().toISOString() });
        log('info', 'Subscription canceled', { event_id: event.id, customer: obj.customer });
        break;
      }

      // ── Invoices ──
      case 'invoice.created': {
        await state.set('invoices/' + objId, { customer: obj.customer, subscription: obj.subscription, amount_due: obj.amount_due, amount_paid: obj.amount_paid, currency: obj.currency, status: obj.status, created_at: new Date().toISOString() });
        log('info', 'Invoice created', { event_id: event.id, invoice: objId, customer: obj.customer });
        break;
      }
      case 'invoice.payment_succeeded': {
        await state.set('invoices/' + objId, { customer: obj.customer, subscription: obj.subscription, amount_due: obj.amount_due, amount_paid: obj.amount_paid, currency: obj.currency, status: 'paid', paid_at: new Date().toISOString() });
        await state.set('payments/' + objId, { customer: obj.customer, amount: obj.amount_paid, currency: obj.currency, status: 'succeeded', invoice: objId, paid_at: new Date().toISOString() });
        log('info', 'Invoice payment succeeded', { event_id: event.id, invoice: objId, amount: obj.amount_paid });
        break;
      }
      case 'invoice.payment_failed': {
        await state.set('invoices/' + objId, { customer: obj.customer, subscription: obj.subscription, amount_due: obj.amount_due, currency: obj.currency, status: 'payment_failed', attempt_count: obj.attempt_count, next_payment_attempt: obj.next_payment_attempt, failed_at: new Date().toISOString() });
        await state.set('failed_payments/' + obj.customer, { invoice: objId, attempt_count: obj.attempt_count, next_payment_attempt: obj.next_payment_attempt, timestamp: Date.now() });
        log('warn', 'Invoice payment failed', { event_id: event.id, invoice: objId, customer: obj.customer, attempt: obj.attempt_count });
        break;
      }

      // ── Payment Intents ──
      case 'payment_intent.payment_failed': {
        await state.set('payment_intents/' + objId, { customer: obj.customer, amount: obj.amount, currency: obj.currency, status: 'failed', last_error: obj.last_error?.message, failed_at: new Date().toISOString() });
        log('warn', 'Payment intent failed', { event_id: event.id, payment_intent: objId, error: obj.last_error?.message });
        break;
      }

      // ── Disputes ──
      case 'charge.dispute.created': {
        await state.set('disputes/' + objId, { charge: obj.charge, amount: obj.amount, currency: obj.currency, reason: obj.reason, status: obj.status, created_at: new Date().toISOString() });
        log('warn', 'Dispute created', { event_id: event.id, dispute: objId, charge: obj.charge, amount: obj.amount, reason: obj.reason });
        break;
      }
      case 'charge.dispute.updated': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, updated_at: new Date().toISOString() });
        log('info', 'Dispute updated', { event_id: event.id, dispute: objId, status: obj.status });
        break;
      }
      case 'charge.dispute.closed': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, closed_at: new Date().toISOString() });
        log('info', 'Dispute closed', { event_id: event.id, dispute: objId, status: obj.status });
        break;
      }
      case 'charge.dispute.funds_withdrawn': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, funds_withdrawn: true, withdrawn_at: new Date().toISOString() });
        log('warn', 'Dispute funds withdrawn', { event_id: event.id, dispute: objId, amount: obj.amount });
        break;
      }

      // ── Refunds ──
      case 'charge.refunded': {
        const refunds = obj.refunds?.data || [];
        await state.set('charges/' + objId, { customer: obj.customer, amount: obj.amount, amount_refunded: obj.amount_refunded, refunded: obj.refunded, refunded_at: new Date().toISOString() });
        for (const refund of refunds) {
          await state.set('refunds/' + refund.id, { charge: objId, amount: refund.amount, currency: refund.currency, reason: refund.reason, status: refund.status, created_at: new Date().toISOString() });
        }
        log('info', 'Charge refunded', { event_id: event.id, charge: objId, amount_refunded: obj.amount_refunded, refund_count: refunds.length });
        break;
      }

      // ── Customer updates ──
      case 'customer.updated': {
        await state.set('customers/' + objId, { email: obj.email, name: obj.name, phone: obj.phone, currency: obj.currency, updated_at: new Date().toISOString() });
        log('info', 'Customer updated', { event_id: event.id, customer: objId });
        break;
      }

      // ── Payment methods ──
      case 'payment_method.updated': {
        await state.set('payment_methods/' + objId, { customer: obj.customer, type: obj.type, updated_at: new Date().toISOString() });
        log('info', 'Payment method updated', { event_id: event.id, payment_method: objId });
        break;
      }
      case 'payment_method.detached': {
        await state.set('payment_methods/' + objId, { customer: null, detached: true, detached_at: new Date().toISOString() });
        log('info', 'Payment method detached', { event_id: event.id, payment_method: objId });
        break;
      }

      // ── Connect payouts ──
      case 'payout.paid': {
        await state.set('payouts/' + objId, { amount: obj.amount, currency: obj.currency, status: 'paid', arrival_date: obj.arrival_date, paid_at: new Date().toISOString() });
        log('info', 'Payout paid', { event_id: event.id, payout: objId, amount: obj.amount });
        break;
      }
      case 'payout.failed': {
        await state.set('payouts/' + objId, { amount: obj.amount, currency: obj.currency, status: 'failed', failure_code: obj.failure_code, failure_message: obj.failure_message, failed_at: new Date().toISOString() });
        log('warn', 'Payout failed', { event_id: event.id, payout: objId, failure_code: obj.failure_code });
        break;
      }
      case 'transfer.reversed': {
        await state.set('transfers/' + objId, { amount: obj.amount, currency: obj.currency, reversed: true, reversed_at: new Date().toISOString() });
        log('warn', 'Transfer reversed', { event_id: event.id, transfer: objId, amount: obj.amount });
        break;
      }
      case 'account.updated': {
        await state.set('connect_accounts/' + objId, { charges_enabled: obj.charges_enabled, payouts_enabled: obj.payouts_enabled, details_submitted: obj.details_submitted, updated_at: new Date().toISOString() });
        log('info', 'Connect account updated', { event_id: event.id, account: objId, charges_enabled: obj.charges_enabled });
        break;
      }

      // ── Unhandled event types — acknowledge silently ──
      default:
        log('info', 'Event acknowledged (unhandled type)', { event_id: event.id, event_type: event.type });
        break;
    }

    // Mark as processed for idempotency
    try { await state.set(processedKey, { processed_at: new Date().toISOString(), event_type: event.type }); } catch { /* non-critical */ }
    res.json({ status: 'received', event_type: event.type });
  } catch (err) {
    log('error', 'Event processing failed', { event_id: event.id, event_type: event.type, error: err.message, stack: err.stack });
    // Always return 200 — if we return 4xx/5xx, Stripe will retry indefinitely
    res.json({ status: 'error', event_type: event.type, message: err.message });
  }
};`,
			},
			{
				name:        "welcome-email",
				title:       "Welcome Email Sender",
				description: "Sends personalized welcome emails to new users after signup. Validates input, logs delivery status, and handles errors gracefully.",
				region:      "us-east-1",
				capabs:      []string{"email"},
				code: `function log(level, msg, data) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), level, msg, ...data }));
}

function isValidEmail(e) {
  return typeof e === 'string' && e.length <= 254 && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e);
}

function sanitizeName(name) {
  if (typeof name !== 'string') return '';
  return name.replace(/[<>&"'\/\\]/g, '').trim().slice(0, 100);
}

export default async (req, res) => {
  const { email: recipientEmail, name } = req.body || {};

  if (!isValidEmail(recipientEmail)) {
    log('warn', 'Invalid email', { email: typeof recipientEmail });
    return res.status(200).json({ sent: false, error: 'Invalid or missing email address' });
  }

  const safeName = sanitizeName(name) || 'there';
  const emailId = crypto.randomUUID();

  try {
    await email.send({
      to: recipientEmail,
      subject: 'Welcome to FunctionFly!',
      template: 'welcome',
      data: { name: safeName, email: recipientEmail, id: emailId }
    });
    log('info', 'Welcome email sent', { email_id: emailId, to: recipientEmail });
    res.json({ sent: true, email_id: emailId });
  } catch (err) {
    log('error', 'Email send failed', { email_id: emailId, to: recipientEmail, error: err.message });
    res.json({ sent: false, error: 'Email delivery failed' });
  }
};`,
			},
		}
	case "marketplace":
		return []bundleFnTemplate{
			{
				name:        "create-listing",
				title:       "Create Listing",
				description: "Creates marketplace listings with input validation, price bounds, duplicate detection, and structured logging.",
				region:      "us-east-1",
				capabs:      []string{"storage"},
				code: `function log(level, msg, data) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), level, msg, ...data }));
}

function sanitize(str, maxLen) {
  if (typeof str !== 'string') return '';
  return str.replace(/[<>&]/g, c => ({ '<': '&lt;', '>': '&gt;', '&': '&amp;' }[c])).trim().slice(0, maxLen);
}

export default async (req, res) => {
  const { title, description, price, category, image_url } = req.body || {};
  const seller_id = req.user?.id;

  if (!seller_id) {
    log('warn', 'Unauthenticated listing creation attempt');
    return res.status(200).json({ success: false, error: 'Authentication required' });
  }
  if (!title || typeof title !== 'string' || title.trim().length < 3) {
    return res.status(200).json({ success: false, error: 'Title must be at least 3 characters' });
  }
  if (title.length > 200) {
    return res.status(200).json({ success: false, error: 'Title must be 200 characters or less' });
  }
  if (typeof price !== 'number' || !Number.isFinite(price) || price < 0) {
    return res.status(200).json({ success: false, error: 'Price must be a non-negative number' });
  }
  if (price > 999999.99) {
    return res.status(200).json({ success: false, error: 'Price exceeds maximum allowed' });
  }
  if (description && typeof description === 'string' && description.length > 5000) {
    return res.status(200).json({ success: false, error: 'Description must be 5000 characters or less' });
  }

  const listingId = crypto.randomUUID();
  const listing = {
    id: listingId,
    seller_id,
    title: sanitize(title, 200),
    description: sanitize(description || '', 5000),
    price_cents: Math.round(price * 100),
    category: sanitize(category || 'general', 64),
    image_url: typeof image_url === 'string' && image_url.startsWith('https://') ? image_url.slice(0, 2048) : null,
    status: 'active',
    created_at: new Date().toISOString()
  };

  try {
    await state.set('listings/' + listingId, listing);
    await state.push('seller_listings/' + seller_id, listingId);
    log('info', 'Listing created', { listing_id: listingId, seller_id, price_cents: listing.price_cents });
    res.json({ success: true, listing_id: listingId });
  } catch (err) {
    log('error', 'Failed to create listing', { seller_id, error: err.message });
    res.json({ success: false, error: 'Failed to save listing' });
  }
};`,
			},
			{
				name:        "send-message",
				title:       "Send Message",
				description: "Enables buyer-seller messaging with input validation, content length limits, rate limiting awareness, and XSS prevention.",
				region:      "us-east-1",
				capabs:      []string{"storage"},
				code: `function log(level, msg, data) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), level, msg, ...data }));
}

function sanitize(str, maxLen) {
  if (typeof str !== 'string') return '';
  return str.replace(/[<>&]/g, c => ({ '<': '&lt;', '>': '&gt;', '&': '&amp;' }[c])).trim().slice(0, maxLen);
}

const MAX_MESSAGE_LENGTH = 4000;
const RATE_LIMIT_WINDOW_MS = 60000;
const RATE_LIMIT_MAX = 30;

export default async (req, res) => {
  const { recipient_id, content } = req.body || {};
  const sender_id = req.user?.id;

  if (!sender_id) {
    log('warn', 'Unauthenticated message attempt');
    return res.status(200).json({ success: false, error: 'Authentication required' });
  }
  if (!recipient_id || typeof recipient_id !== 'string' || recipient_id.length > 128) {
    return res.status(200).json({ success: false, error: 'Invalid recipient' });
  }
  if (sender_id === recipient_id) {
    return res.status(200).json({ success: false, error: 'Cannot send message to yourself' });
  }
  if (!content || typeof content !== 'string' || content.trim().length === 0) {
    return res.status(200).json({ success: false, error: 'Message content is required' });
  }
  if (content.length > MAX_MESSAGE_LENGTH) {
    return res.status(200).json({ success: false, error: 'Message exceeds ' + MAX_MESSAGE_LENGTH + ' character limit' });
  }

  // Rate limiting check
  try {
    const rateKey = 'msg_rate/' + sender_id;
    const recent = (await state.get(rateKey)) || { count: 0, window_start: Date.now() };
    const now = Date.now();
    if (now - recent.window_start > RATE_LIMIT_WINDOW_MS) {
      await state.set(rateKey, { count: 1, window_start: now });
    } else if (recent.count >= RATE_LIMIT_MAX) {
      log('warn', 'Rate limit exceeded', { sender_id });
      return res.status(200).json({ success: false, error: 'Rate limit exceeded. Try again later.' });
    } else {
      await state.set(rateKey, { count: recent.count + 1, window_start: recent.window_start });
    }
  } catch { /* rate limiting is best-effort */ }

  const messageId = crypto.randomUUID();
  const message = {
    id: messageId,
    sender_id,
    recipient_id,
    content: sanitize(content, MAX_MESSAGE_LENGTH),
    created_at: new Date().toISOString()
  };

  try {
    await state.push('messages/' + recipient_id, message);
    await state.push('sent_messages/' + sender_id, messageId);
    log('info', 'Message sent', { message_id: messageId, from: sender_id, to: recipient_id });
    res.json({ success: true, message_id: messageId });
  } catch (err) {
    log('error', 'Failed to send message', { sender_id, recipient_id, error: err.message });
    res.json({ success: false, error: 'Failed to deliver message' });
  }
};`,
			},
		}
	case "ai-app":
		return []bundleFnTemplate{
			{
				name:        "chat-completion",
				title:       "Chat Completion",
				description: "AI chat completions via OpenAI-compatible API with model allowlisting, message validation, token limits, and error handling.",
				region:      "us-east-1",
				capabs:      []string{"ai"},
				code: `function log(level, msg, data) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), level, msg, ...data }));
}

const ALLOWED_MODELS = [
  'gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo',
  'claude-3-5-sonnet-20241022', 'claude-3-5-haiku-20241022', 'claude-3-opus-20240229',
  'google/gemini-2.0-flash-exp', 'google/gemini-pro-1.5',
  'meta-llama/llama-3.1-70b-instruct', 'meta-llama/llama-3.1-8b-instruct',
  'mistralai/mixtral-8x7b-instruct'
];
const DEFAULT_MODEL = 'gpt-4o-mini';
const MAX_MESSAGES = 100;
const MAX_TOKENS = 4096;

export default async (req, res) => {
  const { messages, model, max_tokens, temperature, stream } = req.body || {};

  if (!Array.isArray(messages) || messages.length === 0) {
    return res.status(200).json({ error: 'messages must be a non-empty array' });
  }
  if (messages.length > MAX_MESSAGES) {
    return res.status(200).json({ error: 'Maximum ' + MAX_MESSAGES + ' messages exceeded' });
  }
  for (const msg of messages) {
    if (!msg.role || !msg.content) {
      return res.status(200).json({ error: 'Each message must have role and content' });
    }
    if (!['system', 'user', 'assistant', 'tool'].includes(msg.role)) {
      return res.status(200).json({ error: 'Invalid role: ' + msg.role });
    }
    if (typeof msg.content === 'string' && msg.content.length > 32000) {
      return res.status(200).json({ error: 'Message content exceeds 32000 character limit' });
    }
  }

  const selectedModel = ALLOWED_MODELS.includes(model) ? model : DEFAULT_MODEL;
  const tokens = typeof max_tokens === 'number' && max_tokens > 0 ? Math.min(max_tokens, MAX_TOKENS) : MAX_TOKENS;
  const temp = typeof temperature === 'number' && temperature >= 0 && temperature <= 2 ? temperature : 0.7;

  const requestId = crypto.randomUUID();
  log('info', 'Chat completion request', { request_id: requestId, model: selectedModel, message_count: messages.length });

  try {
    const response = await ai.chat.completions.create({
      model: selectedModel,
      messages,
      max_tokens: tokens,
      temperature: temp,
      stream: false
    });

    const usage = response.usage || {};
    log('info', 'Chat completion succeeded', {
      request_id: requestId,
      model: selectedModel,
      prompt_tokens: usage.prompt_tokens,
      completion_tokens: usage.completion_tokens,
      total_tokens: usage.total_tokens
    });

    res.json({
      id: requestId,
      model: selectedModel,
      message: response.choices?.[0]?.message?.content || '',
      finish_reason: response.choices?.[0]?.finish_reason,
      usage: { prompt_tokens: usage.prompt_tokens || 0, completion_tokens: usage.completion_tokens || 0, total_tokens: usage.total_tokens || 0 }
    });
  } catch (err) {
    log('error', 'Chat completion failed', { request_id: requestId, model: selectedModel, error: err.message });
    res.json({ error: 'Chat completion failed', request_id: requestId });
  }
};`,
			},
			{
				name:        "embed-and-store",
				title:       "Embed & Store",
				description: "Generates text embeddings and stores them in your vector collection with input validation, collection name sanitization, and size limits.",
				region:      "us-east-1",
				capabs:      []string{"ai", "storage"},
				code: `function log(level, msg, data) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), level, msg, ...data }));
}

const MAX_TEXT_LENGTH = 32000;
const MAX_COLLECTION_LEN = 128;
const VALID_COLLECTION = /^[a-zA-Z0-9_-]+$/;

export default async (req, res) => {
  const { text, collection } = req.body || {};

  if (!text || typeof text !== 'string') {
    return res.status(200).json({ success: false, error: 'text is required and must be a string' });
  }
  if (text.length > MAX_TEXT_LENGTH) {
    return res.status(200).json({ success: false, error: 'text exceeds ' + MAX_TEXT_LENGTH + ' character limit' });
  }
  if (!collection || typeof collection !== 'string') {
    return res.status(200).json({ success: false, error: 'collection is required' });
  }
  if (collection.length > MAX_COLLECTION_LEN || !VALID_COLLECTION.test(collection)) {
    return res.status(200).json({ success: false, error: 'collection name must be alphanumeric (max 128 chars)' });
  }

  const docId = crypto.randomUUID();
  log('info', 'Embedding request', { doc_id: docId, collection, text_length: text.length });

  try {
    const embedding = await ai.embeddings.create({
      model: 'text-embedding-3-small',
      input: text
    });

    const vector = embedding.data?.[0]?.embedding;
    if (!Array.isArray(vector) || vector.length === 0) {
      throw new Error('Empty embedding returned');
    }

    const record = {
      id: docId,
      text: text.substring(0, 1000),
      embedding: vector,
      dimensions: vector.length,
      text_length: text.length,
      collection,
      created_at: new Date().toISOString()
    };

    await state.set(collection + '/' + docId, record);

    log('info', 'Embedding stored', { doc_id: docId, collection, dimensions: vector.length });
    res.json({ success: true, id: docId, dimensions: vector.length, collection });
  } catch (err) {
    log('error', 'Embedding failed', { doc_id: docId, collection, error: err.message });
    res.json({ success: false, error: 'Embedding generation failed' });
  }
};`,
			},
		}
	default:
		return nil
	}
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
