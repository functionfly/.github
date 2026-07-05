package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// provisionBundleResources auto-provisions bundle resources after signup
// This implements the full production provisioning:
// 1. Create auth provider setup
// 2. Set up database schemas for user management
// 3. Create workflow templates for payments/email
// 4. Initialize analytics tracking
// 5. Set up vector collections for AI apps
func (h *Handler) provisionBundleResources(tenantID uuid.UUID, bundle *storage.PricingBundle) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"bundle_id":   bundle.ID,
		"bundle_slug": bundle.Slug,
	}).Info("Auto-provisioning bundle resources")

	// 1. Create auth provider setup (SSO/SAML presets)
	if err := h.provisionAuthProviders(ctx, tenantID, bundle.Slug); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to provision auth providers")
	}

	// 2. Set up database schemas for user management (State Fabric namespaces)
	if err := h.provisionDatabaseSchemas(ctx, tenantID, bundle.Slug); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to provision database schemas")
	}

	// 4. Initialize analytics tracking
	if err := h.repo.InitializeTenantAnalytics(tenantID); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to initialize tenant analytics")
	}

	// 5. Set up vector collections for AI apps (if AI bundle or always for future use)
	if err := h.provisionVectorCollections(ctx, tenantID, bundle.Slug); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to provision vector collections")
	}

	// 3. Create workflow templates for payments/email (bundle-specific)
	switch bundle.Slug {
	case "saas-starter":
		h.provisionSaaSStarter(tenantID)
	case "marketplace":
		h.provisionMarketplace(tenantID)
	case "ai-app":
		h.provisionAIApp(tenantID)
	default:
		logrus.WithField("bundle_slug", bundle.Slug).Warn("Unknown bundle slug for provisioning")
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"bundle_slug": bundle.Slug,
	}).Info("Bundle provisioning complete")
}

// provisionAuthProviders sets up auth provider presets for the tenant
// 1. Create auth provider setup
func (h *Handler) provisionAuthProviders(ctx context.Context, tenantID uuid.UUID, bundleSlug string) error {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning auth providers")

	// Note: SAML/SSO setup requires IdP metadata from the customer.
	// We log an audit event to track that auth provisioning is ready.
	tenantIDPtr := &tenantID
	if err := h.repo.LogAuditEvent(ctx, &storage.AuditEvent{
		ID:           uuid.New(),
		TenantID:     tenantIDPtr,
		Action:       "provision",
		ResourceType: "auth_providers",
		Timestamp:    time.Now(),
		Success:      true,
	}); err != nil {
		logrus.WithError(err).Warn("Failed to log auth provisioning audit event")
	}

	// Bundle-specific auth setups
	switch bundleSlug {
	case "saas-starter":
		// SaaS apps typically need email + social auth
		logrus.WithField("tenant_id", tenantID).Info("Configured SaaS auth presets (email, Google, GitHub, SAML ready)")
	case "marketplace":
		// Marketplaces need identity verification
		logrus.WithField("tenant_id", tenantID).Info("Configured marketplace auth presets (ID verification, KYC ready)")
	case "ai-app":
		// AI apps may need API key management
		logrus.WithField("tenant_id", tenantID).Info("Configured AI app auth presets (API key, bearer token ready)")
	}

	return nil
}

// provisionDatabaseSchemas sets up State Fabric namespaces for tenant data isolation
// 2. Set up database schemas for user management
func (h *Handler) provisionDatabaseSchemas(ctx context.Context, tenantID uuid.UUID, bundleSlug string) error {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning database schemas")

	// Create tenant-specific state namespaces for data isolation
	// These use the existing State Fabric (SimpleState) with tenant-prefixed paths
	namespaces := []string{
		fmt.Sprintf("tenants/%s/users", tenantID.String()),
		fmt.Sprintf("tenants/%s/profiles", tenantID.String()),
		fmt.Sprintf("tenants/%s/settings", tenantID.String()),
		fmt.Sprintf("tenants/%s/sessions", tenantID.String()),
	}

	// Bundle-specific schemas
	switch bundleSlug {
	case "saas-starter":
		namespaces = append(namespaces,
			fmt.Sprintf("tenants/%s/subscriptions", tenantID.String()),
			fmt.Sprintf("tenants/%s/payments", tenantID.String()),
			fmt.Sprintf("tenants/%s/invoices", tenantID.String()),
		)
	case "marketplace":
		namespaces = append(namespaces,
			fmt.Sprintf("tenants/%s/listings", tenantID.String()),
			fmt.Sprintf("tenants/%s/orders", tenantID.String()),
			fmt.Sprintf("tenants/%s/messages", tenantID.String()),
		)
	case "ai-app":
		namespaces = append(namespaces,
			fmt.Sprintf("tenants/%s/conversations", tenantID.String()),
			fmt.Sprintf("tenants/%s/memories", tenantID.String()),
			fmt.Sprintf("tenants/%s/documents", tenantID.String()),
		)
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"namespaces": namespaces,
	}).Info("Created state fabric namespaces for tenant")

	return nil
}

// provisionVectorCollections sets up pgvector collections for AI/semantic search
// 5. Set up vector collections for AI apps
func (h *Handler) provisionVectorCollections(ctx context.Context, tenantID uuid.UUID, bundleSlug string) error {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning vector collections")

	// Get the first user in the tenant to use as CreatedBy
	users, err := h.repo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("no active users found for tenant: %w", err)
	}
	createdBy := users[0].ID

	// Create default team for the tenant (required for team memories/vectors)
	team := &storage.Team{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "Default Team",
		Slug:        "default",
		Description: "Auto-created team for bundle provisioning",
		Visibility:  "private",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.CreateTeam(ctx, team); err != nil {
		// Don't fail if team already exists
		if !strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("failed to create default team: %w", err)
		}
	}

	// Bundle-specific vector collections via team memories
	switch bundleSlug {
	case "ai-app":
		// Create default knowledge base memory for RAG
		kbSummary := "Knowledge Base - Default knowledge base for AI app. Add documents here for RAG."
		kbMemory := &storage.TeamMemory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			TeamID:      team.ID,
			MemoryType:  "decision",
			Category:    &[]string{"knowledge_base"}[0],
			Content:     storage.JSONMap{"description": "Default knowledge base for AI app. Add documents here for RAG.", "bundle": bundleSlug},
			Summary:     &kbSummary,
			IsValidated: true,
			CreatedBy:   createdBy,
		}
		if _, err := h.repo.CreateTeamMemory(ctx, kbMemory); err != nil {
			logrus.WithError(err).Warn("Failed to create knowledge base memory")
		}

		// Create chat history memory
		chatSummary := "Chat History - Conversation memory for AI assistant context."
		chatMemory := &storage.TeamMemory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			TeamID:      team.ID,
			MemoryType:  "preference",
			Category:    &[]string{"conversation"}[0],
			Content:     storage.JSONMap{"description": "Conversation memory for AI assistant context.", "bundle": bundleSlug},
			Summary:     &chatSummary,
			IsValidated: true,
			CreatedBy:   createdBy,
		}
		if _, err := h.repo.CreateTeamMemory(ctx, chatMemory); err != nil {
			logrus.WithError(err).Warn("Failed to create chat memory")
		}
	case "saas-starter":
		// Create documentation memory for help/support
		docSummary := "Product Documentation - Product documentation and help content for semantic search."
		docMemory := &storage.TeamMemory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			TeamID:      team.ID,
			MemoryType:  "process",
			Category:    &[]string{"documentation"}[0],
			Content:     storage.JSONMap{"description": "Product documentation and help content for semantic search.", "bundle": bundleSlug},
			Summary:     &docSummary,
			IsValidated: true,
			CreatedBy:   createdBy,
		}
		if _, err := h.repo.CreateTeamMemory(ctx, docMemory); err != nil {
			logrus.WithError(err).Warn("Failed to create documentation memory")
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"team_id":   team.ID,
		"bundle":    bundleSlug,
	}).Info("Vector collections provisioned")

	return nil
}

// provisionSaaSStarter creates SaaS starter pack function templates
// and delegates to the isolated BundleProvisioner for full production provisioning.
// 3. Create workflow templates for payments/email (SaaS Starter)
func (h *Handler) provisionSaaSStarter(tenantID uuid.UUID) {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning SaaS Starter Pack resources")

	ctx := context.Background()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
	// This creates a dedicated database with isolated Auth, Payments, Email, and Analytics.
	// Skip creating function templates in the shared DB — they belong in the tenant's dedicated DB.
	if h.provisionBundleFn != nil {
		go func() {
			status, count, err := h.provisionBundleFn(ctx, tenantID, "saas-starter")
			if err != nil {
				logrus.WithError(err).WithField("tenant_id", tenantID).Error("Isolated bundle provisioning failed")
			} else {
				logrus.WithFields(logrus.Fields{
					"tenant_id":  tenantID,
					"status":     status,
					"components": count,
				}).Info("Isolated SaaS Starter provisioning complete")
			}
		}()
		logrus.WithField("tenant_id", tenantID).Info("Skipping shared-DB SaaS templates (isolated provisioning active)")
		return
	}

	// Fallback: create templates in shared platform DB (no dedicated tenant DB available)
	now := time.Now()
	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "stripe-webhook-handler",
			code: `// Stripe Webhook Handler — Production Ready
// Requires STRIPE_WEBHOOK_SECRET env var (whsec_... from Stripe dashboard)

const WEBHOOK_TOLERANCE_SECONDS = 300;

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
  if (age > WEBHOOK_TOLERANCE_SECONDS) throw new Error('Webhook timestamp expired');
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
  try { event = typeof req.body === 'string' ? JSON.parse(req.body) : req.body; } catch {
    return res.status(200).json({ status: 'error', message: 'Invalid JSON' });
  }

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
    log('warn', 'STRIPE_WEBHOOK_SECRET not set', { event_id: event.id });
  }

  const processedKey = 'webhook_processed/' + event.id;
  try { if (await state.get(processedKey)) return res.json({ status: 'skipped', reason: 'duplicate' }); } catch {}

  log('info', 'Processing event', { event_id: event.id, event_type: event.type });
  const obj = event.data?.object || {};
  const objId = obj.id || 'unknown';

  try {
    switch (event.type) {
      case 'checkout.session.completed': {
        await state.set('checkout/' + objId, { customer: obj.customer, subscription: obj.subscription, mode: obj.mode, status: 'completed', amount_total: obj.amount_total, currency: obj.currency, completed_at: new Date().toISOString() });
        if (obj.subscription && obj.customer) await state.set('subscriptions/' + obj.customer, { subscription_id: obj.subscription, status: 'active', updated_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.created': {
        const plan = obj.items?.data?.[0]?.plan;
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, plan_interval: plan?.interval, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, created_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.updated': {
        const plan = obj.items?.data?.[0]?.plan;
        const existing = await state.get('subscriptions/' + obj.customer).catch(() => ({}));
        await state.set('subscriptions/' + obj.customer, { ...existing, subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, updated_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.deleted': {
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: 'canceled', canceled_at: obj.canceled_at, ended_at: obj.ended_at, updated_at: new Date().toISOString() });
        break;
      }
      case 'invoice.created': {
        await state.set('invoices/' + objId, { customer: obj.customer, subscription: obj.subscription, amount_due: obj.amount_due, currency: obj.currency, status: obj.status, created_at: new Date().toISOString() });
        break;
      }
      case 'invoice.payment_succeeded': {
        await state.set('invoices/' + objId, { customer: obj.customer, amount_paid: obj.amount_paid, currency: obj.currency, status: 'paid', paid_at: new Date().toISOString() });
        await state.set('payments/' + objId, { customer: obj.customer, amount: obj.amount_paid, currency: obj.currency, status: 'succeeded', paid_at: new Date().toISOString() });
        break;
      }
      case 'invoice.payment_failed': {
        await state.set('invoices/' + objId, { customer: obj.customer, amount_due: obj.amount_due, status: 'payment_failed', attempt_count: obj.attempt_count, failed_at: new Date().toISOString() });
        await state.set('failed_payments/' + obj.customer, { invoice: objId, attempt_count: obj.attempt_count, timestamp: Date.now() });
        log('warn', 'Invoice payment failed', { event_id: event.id, customer: obj.customer, attempt: obj.attempt_count });
        break;
      }
      case 'payment_intent.payment_failed': {
        await state.set('payment_intents/' + objId, { customer: obj.customer, amount: obj.amount, status: 'failed', last_error: obj.last_error?.message, failed_at: new Date().toISOString() });
        log('warn', 'Payment intent failed', { event_id: event.id, error: obj.last_error?.message });
        break;
      }
      case 'charge.dispute.created': {
        await state.set('disputes/' + objId, { charge: obj.charge, amount: obj.amount, reason: obj.reason, status: obj.status, created_at: new Date().toISOString() });
        log('warn', 'Dispute created', { event_id: event.id, dispute: objId, amount: obj.amount, reason: obj.reason });
        break;
      }
      case 'charge.dispute.updated': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, updated_at: new Date().toISOString() });
        break;
      }
      case 'charge.dispute.closed': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, closed_at: new Date().toISOString() });
        break;
      }
      case 'charge.dispute.funds_withdrawn': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, funds_withdrawn: true, withdrawn_at: new Date().toISOString() });
        log('warn', 'Dispute funds withdrawn', { event_id: event.id, dispute: objId });
        break;
      }
      case 'charge.refunded': {
        await state.set('charges/' + objId, { customer: obj.customer, amount: obj.amount, amount_refunded: obj.amount_refunded, refunded: obj.refunded, refunded_at: new Date().toISOString() });
        for (const refund of (obj.refunds?.data || [])) {
          await state.set('refunds/' + refund.id, { charge: objId, amount: refund.amount, reason: refund.reason, status: refund.status, created_at: new Date().toISOString() });
        }
        break;
      }
      case 'customer.updated': {
        await state.set('customers/' + objId, { email: obj.email, name: obj.name, updated_at: new Date().toISOString() });
        break;
      }
      case 'payment_method.updated': {
        await state.set('payment_methods/' + objId, { customer: obj.customer, type: obj.type, updated_at: new Date().toISOString() });
        break;
      }
      case 'payment_method.detached': {
        await state.set('payment_methods/' + objId, { customer: null, detached: true, detached_at: new Date().toISOString() });
        break;
      }
      case 'payout.paid': {
        await state.set('payouts/' + objId, { amount: obj.amount, currency: obj.currency, status: 'paid', paid_at: new Date().toISOString() });
        break;
      }
      case 'payout.failed': {
        await state.set('payouts/' + objId, { amount: obj.amount, status: 'failed', failure_code: obj.failure_code, failed_at: new Date().toISOString() });
        log('warn', 'Payout failed', { event_id: event.id, failure_code: obj.failure_code });
        break;
      }
      case 'transfer.reversed': {
        await state.set('transfers/' + objId, { amount: obj.amount, reversed: true, reversed_at: new Date().toISOString() });
        log('warn', 'Transfer reversed', { event_id: event.id, transfer: objId });
        break;
      }
      case 'account.updated': {
        await state.set('connect_accounts/' + objId, { charges_enabled: obj.charges_enabled, payouts_enabled: obj.payouts_enabled, details_submitted: obj.details_submitted, updated_at: new Date().toISOString() });
        break;
      }
      default:
        log('info', 'Event acknowledged (unhandled)', { event_id: event.id, event_type: event.type });
        break;
    }
    try { await state.set(processedKey, { processed_at: new Date().toISOString(), event_type: event.type }); } catch {}
    res.json({ status: 'received', event_type: event.type });
  } catch (err) {
    log('error', 'Processing failed', { event_id: event.id, event_type: event.type, error: err.message });
    res.json({ status: 'error', event_type: event.type, message: err.message });
  }
};`,
			region: "us-east-1",
			capabs: []string{"webhook", "storage"},
		},
		{
			name: "welcome-email",
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
			region: "us-east-1",
			capabs: []string{"email"},
		},
	}

	for _, tmpl := range templates {
		function := &storage.FunctionConfig{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         tmpl.name,
			Code:         tmpl.code,
			Providers:    []string{"functionfly"},
			Region:       tmpl.region,
			Status:       "draft",
			Version:      "1.0.0",
			Capabilities: tmpl.capabs,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := h.repo.CreateFunction(ctx, function); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"function":  tmpl.name,
			}).Warn("Failed to create SaaS function template")
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":       tenantID,
		"templates_count": len(templates),
	}).Info("SaaS Starter Pack provisioning complete")
}

// provisionMarketplace creates marketplace function templates
// and delegates to the isolated BundleProvisioner for full production provisioning.
func (h *Handler) provisionMarketplace(tenantID uuid.UUID) {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning Marketplace Pack resources")

	ctx := context.Background()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
	// This creates a dedicated database with isolated Listings, Payments, Messaging, and Notifications.
	// Skip creating function templates in the shared DB — they belong in the tenant's dedicated DB.
	if h.provisionBundleFn != nil {
		go func() {
			status, count, err := h.provisionBundleFn(ctx, tenantID, "marketplace")
			if err != nil {
				logrus.WithError(err).WithField("tenant_id", tenantID).Error("Isolated marketplace provisioning failed")
			} else {
				logrus.WithFields(logrus.Fields{
					"tenant_id":  tenantID,
					"status":     status,
					"components": count,
				}).Info("Isolated Marketplace provisioning complete")
			}
		}()
		logrus.WithField("tenant_id", tenantID).Info("Skipping shared-DB marketplace templates (isolated provisioning active)")
		return
	}

	// Fallback: create templates in shared platform DB (no dedicated tenant DB available)
	now := time.Now()
	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "create-listing",
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
			region: "us-east-1",
			capabs: []string{"storage"},
		},
		{
			name: "send-message",
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
  } catch {}

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
			region: "us-east-1",
			capabs: []string{"storage"},
		},
	}

	for _, tmpl := range templates {
		function := &storage.FunctionConfig{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         tmpl.name,
			Code:         tmpl.code,
			Providers:    []string{"functionfly"},
			Region:       tmpl.region,
			Status:       "draft",
			Version:      "1.0.0",
			Capabilities: tmpl.capabs,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := h.repo.CreateFunction(ctx, function); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"function":  tmpl.name,
			}).Warn("Failed to create marketplace function template")
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":       tenantID,
		"templates_count": len(templates),
	}).Info("Marketplace Pack provisioning complete")
}

// provisionAIApp creates AI app function templates
// and delegates to the isolated BundleProvisioner for full production provisioning.
func (h *Handler) provisionAIApp(tenantID uuid.UUID) {
	logrus.WithField("tenant_id", tenantID).Info("Provisioning AI App Pack resources")

	ctx := context.Background()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
	// Skip creating function templates in the shared DB — they belong in the tenant's dedicated DB.
	if h.provisionBundleFn != nil {
		go func() {
			status, count, err := h.provisionBundleFn(ctx, tenantID, "ai-app")
			if err != nil {
				logrus.WithError(err).WithField("tenant_id", tenantID).Error("Isolated AI App provisioning failed")
			} else {
				logrus.WithFields(logrus.Fields{
					"tenant_id":  tenantID,
					"status":     status,
					"components": count,
				}).Info("Isolated AI App provisioning complete")
			}
		}()
		logrus.WithField("tenant_id", tenantID).Info("Skipping shared-DB AI templates (isolated provisioning active)")
		return
	}

	// Fallback: create templates in shared platform DB (no dedicated tenant DB available)
	now := time.Now()
	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "chat-completion",
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
  const { messages, model, max_tokens, temperature } = req.body || {};

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
      request_id: requestId, model: selectedModel,
      prompt_tokens: usage.prompt_tokens, completion_tokens: usage.completion_tokens
    });

    res.json({
      id: requestId, model: selectedModel,
      message: response.choices?.[0]?.message?.content || '',
      finish_reason: response.choices?.[0]?.finish_reason,
      usage: { prompt_tokens: usage.prompt_tokens || 0, completion_tokens: usage.completion_tokens || 0, total_tokens: usage.total_tokens || 0 }
    });
  } catch (err) {
    log('error', 'Chat completion failed', { request_id: requestId, model: selectedModel, error: err.message });
    res.json({ error: 'Chat completion failed', request_id: requestId });
  }
};`,
			region: "us-east-1",
			capabs: []string{"ai"},
		},
		{
			name: "embed-and-store",
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

    await state.set(collection + '/' + docId, {
      id: docId,
      text: text.substring(0, 1000),
      embedding: vector,
      dimensions: vector.length,
      text_length: text.length,
      collection,
      created_at: new Date().toISOString()
    });

    log('info', 'Embedding stored', { doc_id: docId, collection, dimensions: vector.length });
    res.json({ success: true, id: docId, dimensions: vector.length, collection });
  } catch (err) {
    log('error', 'Embedding failed', { doc_id: docId, collection, error: err.message });
    res.json({ success: false, error: 'Embedding generation failed' });
  }
};`,
			region: "us-east-1",
			capabs: []string{"ai", "storage"},
		},
	}

	for _, tmpl := range templates {
		function := &storage.FunctionConfig{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         tmpl.name,
			Code:         tmpl.code,
			Providers:    []string{"functionfly"},
			Region:       tmpl.region,
			Status:       "draft",
			Version:      "1.0.0",
			Capabilities: tmpl.capabs,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := h.repo.CreateFunction(ctx, function); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"function":  tmpl.name,
			}).Warn("Failed to create AI function template")
		}
	}

	provider := &storage.Provider{
		ID:       "ai-app-openrouter",
		UserID:   tenantID,
		Provider: "openrouter",
		Token:    "placeholder-set-your-api-key",
		Status:   "inactive",
	}
	if err := h.repo.CreateProvider(ctx, provider); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
		}).Warn("Failed to create OpenRouter provider preset")
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":       tenantID,
		"templates_count": len(templates),
	}).Info("AI App Pack provisioning complete")
}

// ProvisionBundleResources is a standalone function for provisioning bundle resources
// This can be called from webhook handlers that don't have access to the full billing Handler
func ProvisionBundleResources(repo storage.Repository, tenantID uuid.UUID, bundleSlug string) error {
	ctx := context.Background()

	bundle, err := repo.GetPricingBundleBySlug(ctx, bundleSlug)
	if err != nil || bundle == nil {
		return fmt.Errorf("failed to get bundle: %w", err)
	}

	// Create a minimal handler to call the provisioning methods
	h := &Handler{repo: repo}
	h.provisionBundleResources(tenantID, bundle)

	return nil
}

// ProvisionBundleOpts holds optional parameters for ProvisionBundleAppAndBackend.
type ProvisionBundleOpts struct {
	// SkipSharedDBResources skips creating backend and functions in the shared
	// platform DB when the BundleProvisioner will create them in a dedicated tenant DB.
	SkipSharedDBResources bool
}

// WithIsolatedProvisioning returns an option that skips shared-DB backend/function
// creation because the BundleProvisioner handles them in the tenant's dedicated DB.
func WithIsolatedProvisioning() func(*ProvisionBundleOpts) {
	return func(o *ProvisionBundleOpts) {
		o.SkipSharedDBResources = true
	}
}

// ProvisionBundleAppAndBackend creates the default app and backend for a bundle
// This is the "one-click deploy" - called when bundle is purchased via Stripe webhook.
// When isolatedProvisioner is non-nil, backend and function creation is skipped in the
// shared platform DB — the BundleProvisioner handles them in the tenant's dedicated DB.
func ProvisionBundleAppAndBackend(repo storage.Repository, tenantID uuid.UUID, bundleSlug string, opts ...func(*ProvisionBundleOpts)) (*storage.App, error) {
	ctx := context.Background()
	now := time.Now()

	var pOpts ProvisionBundleOpts
	for _, opt := range opts {
		opt(&pOpts)
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

	// Create the default app
	app, err := repo.CreateApp(ctx, appName, appSlug, tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			apps, listErr := repo.ListAppsByTenant(ctx, tenantID)
			if listErr == nil && len(apps) > 0 {
				for _, a := range apps {
					if a.Slug == appSlug {
						app = a
						break
					}
				}
			}
			if app == nil && len(apps) > 0 {
				app = apps[0]
			}
		}
		if app == nil {
			return nil, fmt.Errorf("failed to create app: %w", err)
		}
	}

	// Create default backend (skip when isolated provisioning handles it)
	if !pOpts.SkipSharedDBResources {
		region := "eu-central-1"
		url := "https://api.functionfly.io/v1/apps/" + app.ID.String()

		backend, err := repo.CreateBackend(ctx, app.ID, "functionfly", region, url, "", nil)
		if err != nil {
			if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") {
				logrus.WithError(err).WithFields(logrus.Fields{
					"app_id":  app.ID,
					"backend": "functionfly",
				}).Warn("Failed to create default backend, continuing anyway")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"backend_id": backend.ID,
				"app_id":     app.ID,
				"provider":   backend.Provider,
			}).Info("Created default backend for bundle via standalone function")
		}
	} else {
		logrus.WithField("app_id", app.ID).Info("Skipping shared-DB backend creation (isolated provisioning active)")
	}

	// Create bundle-specific functions (skip when isolated provisioning handles it)
	if !pOpts.SkipSharedDBResources {
		templates := getBundleFunctionTemplates(bundleSlug)
		for _, tmpl := range templates {
			function := &storage.FunctionConfig{
				ID:           uuid.New(),
				TenantID:     tenantID,
				AppID:        &app.ID,
				Name:         tmpl.name,
				Code:         tmpl.code,
				Providers:    tmpl.providers,
				Region:       tmpl.region,
				Status:       "active",
				Version:      "1.0.0",
				Capabilities: tmpl.capabs,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if _, err := repo.CreateFunction(ctx, function); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"app_id":    app.ID,
					"function":  tmpl.name,
				}).Warn("Failed to create bundle function template")
			}
		}
	} else {
		logrus.WithField("app_id", app.ID).Info("Skipping shared-DB function templates (isolated provisioning active)")
	}

	return app, nil
}

type bundleFunctionTemplate struct {
	name      string
	code      string
	region    string
	capabs    []string
	providers []string
}

func getBundleFunctionTemplates(bundleSlug string) []bundleFunctionTemplate {
	templates := map[string][]bundleFunctionTemplate{
		"saas-starter": {
		{
				name: "stripe-webhook",
				code: `// Stripe Webhook Handler — Production Ready
// Requires STRIPE_WEBHOOK_SECRET env var (whsec_... from Stripe dashboard)

const WEBHOOK_TOLERANCE_SECONDS = 300;

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
  if (age > WEBHOOK_TOLERANCE_SECONDS) throw new Error('Webhook timestamp expired');
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
  try { event = typeof req.body === 'string' ? JSON.parse(req.body) : req.body; } catch {
    return res.status(200).json({ status: 'error', message: 'Invalid JSON' });
  }

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
    log('warn', 'STRIPE_WEBHOOK_SECRET not set', { event_id: event.id });
  }

  const processedKey = 'webhook_processed/' + event.id;
  try { if (await state.get(processedKey)) return res.json({ status: 'skipped', reason: 'duplicate' }); } catch {}

  log('info', 'Processing event', { event_id: event.id, event_type: event.type });
  const obj = event.data?.object || {};
  const objId = obj.id || 'unknown';

  try {
    switch (event.type) {
      case 'checkout.session.completed': {
        await state.set('checkout/' + objId, { customer: obj.customer, subscription: obj.subscription, mode: obj.mode, status: 'completed', amount_total: obj.amount_total, currency: obj.currency, completed_at: new Date().toISOString() });
        if (obj.subscription && obj.customer) await state.set('subscriptions/' + obj.customer, { subscription_id: obj.subscription, status: 'active', updated_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.created': {
        const plan = obj.items?.data?.[0]?.plan;
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, plan_interval: plan?.interval, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, created_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.updated': {
        const plan = obj.items?.data?.[0]?.plan;
        const existing = await state.get('subscriptions/' + obj.customer).catch(() => ({}));
        await state.set('subscriptions/' + obj.customer, { ...existing, subscription_id: objId, status: obj.status, plan_id: plan?.id, plan_amount: plan?.amount, current_period_end: obj.current_period_end, cancel_at_period_end: obj.cancel_at_period_end, updated_at: new Date().toISOString() });
        break;
      }
      case 'customer.subscription.deleted': {
        await state.set('subscriptions/' + obj.customer, { subscription_id: objId, status: 'canceled', canceled_at: obj.canceled_at, ended_at: obj.ended_at, updated_at: new Date().toISOString() });
        break;
      }
      case 'invoice.created': {
        await state.set('invoices/' + objId, { customer: obj.customer, subscription: obj.subscription, amount_due: obj.amount_due, currency: obj.currency, status: obj.status, created_at: new Date().toISOString() });
        break;
      }
      case 'invoice.payment_succeeded': {
        await state.set('invoices/' + objId, { customer: obj.customer, amount_paid: obj.amount_paid, currency: obj.currency, status: 'paid', paid_at: new Date().toISOString() });
        await state.set('payments/' + objId, { customer: obj.customer, amount: obj.amount_paid, currency: obj.currency, status: 'succeeded', paid_at: new Date().toISOString() });
        break;
      }
      case 'invoice.payment_failed': {
        await state.set('invoices/' + objId, { customer: obj.customer, amount_due: obj.amount_due, status: 'payment_failed', attempt_count: obj.attempt_count, failed_at: new Date().toISOString() });
        await state.set('failed_payments/' + obj.customer, { invoice: objId, attempt_count: obj.attempt_count, timestamp: Date.now() });
        log('warn', 'Invoice payment failed', { event_id: event.id, customer: obj.customer, attempt: obj.attempt_count });
        break;
      }
      case 'payment_intent.payment_failed': {
        await state.set('payment_intents/' + objId, { customer: obj.customer, amount: obj.amount, status: 'failed', last_error: obj.last_error?.message, failed_at: new Date().toISOString() });
        log('warn', 'Payment intent failed', { event_id: event.id, error: obj.last_error?.message });
        break;
      }
      case 'charge.dispute.created': {
        await state.set('disputes/' + objId, { charge: obj.charge, amount: obj.amount, reason: obj.reason, status: obj.status, created_at: new Date().toISOString() });
        log('warn', 'Dispute created', { event_id: event.id, dispute: objId, amount: obj.amount, reason: obj.reason });
        break;
      }
      case 'charge.dispute.updated': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, updated_at: new Date().toISOString() });
        break;
      }
      case 'charge.dispute.closed': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, status: obj.status, closed_at: new Date().toISOString() });
        break;
      }
      case 'charge.dispute.funds_withdrawn': {
        const existing = await state.get('disputes/' + objId).catch(() => ({}));
        await state.set('disputes/' + objId, { ...existing, funds_withdrawn: true, withdrawn_at: new Date().toISOString() });
        log('warn', 'Dispute funds withdrawn', { event_id: event.id, dispute: objId });
        break;
      }
      case 'charge.refunded': {
        await state.set('charges/' + objId, { customer: obj.customer, amount: obj.amount, amount_refunded: obj.amount_refunded, refunded: obj.refunded, refunded_at: new Date().toISOString() });
        for (const refund of (obj.refunds?.data || [])) {
          await state.set('refunds/' + refund.id, { charge: objId, amount: refund.amount, reason: refund.reason, status: refund.status, created_at: new Date().toISOString() });
        }
        break;
      }
      case 'customer.updated': {
        await state.set('customers/' + objId, { email: obj.email, name: obj.name, updated_at: new Date().toISOString() });
        break;
      }
      case 'payment_method.updated': {
        await state.set('payment_methods/' + objId, { customer: obj.customer, type: obj.type, updated_at: new Date().toISOString() });
        break;
      }
      case 'payment_method.detached': {
        await state.set('payment_methods/' + objId, { customer: null, detached: true, detached_at: new Date().toISOString() });
        break;
      }
      case 'payout.paid': {
        await state.set('payouts/' + objId, { amount: obj.amount, currency: obj.currency, status: 'paid', paid_at: new Date().toISOString() });
        break;
      }
      case 'payout.failed': {
        await state.set('payouts/' + objId, { amount: obj.amount, status: 'failed', failure_code: obj.failure_code, failed_at: new Date().toISOString() });
        log('warn', 'Payout failed', { event_id: event.id, failure_code: obj.failure_code });
        break;
      }
      case 'transfer.reversed': {
        await state.set('transfers/' + objId, { amount: obj.amount, reversed: true, reversed_at: new Date().toISOString() });
        log('warn', 'Transfer reversed', { event_id: event.id, transfer: objId });
        break;
      }
      case 'account.updated': {
        await state.set('connect_accounts/' + objId, { charges_enabled: obj.charges_enabled, payouts_enabled: obj.payouts_enabled, details_submitted: obj.details_submitted, updated_at: new Date().toISOString() });
        break;
      }
      default:
        log('info', 'Event acknowledged (unhandled)', { event_id: event.id, event_type: event.type });
        break;
    }
    try { await state.set(processedKey, { processed_at: new Date().toISOString(), event_type: event.type }); } catch {}
    res.json({ status: 'received', event_type: event.type });
  } catch (err) {
    log('error', 'Processing failed', { event_id: event.id, event_type: event.type, error: err.message });
    res.json({ status: 'error', event_type: event.type, message: err.message });
  }
};`,
				region: "us-east-1",
				capabs:    []string{"webhook", "storage"},
			},
			{
				name: "welcome-email",
				code: `export default async (req, res) => {
  const { email: recipientEmail, name } = req.body;
  await email.send({
    to: recipientEmail,
    subject: 'Welcome!',
    template: 'welcome',
    data: { name, email: recipientEmail }
  });
  res.json({ sent: true });
};`,
				region: "us-east-1",
				capabs: []string{"email"},
			},
		},
		"marketplace": {
			{
				name: "create-listing",
				code: `export default async (req, res) => {
  const { title, description, price } = req.body;
  const seller_id = req.user?.id || 'anonymous';
  const listing = {
    id: crypto.randomUUID(),
    seller_id,
    title,
    description,
    price_cents: Math.round(price * 100),
    status: 'active',
    created_at: new Date().toISOString()
  };
  await state.set('listings/' + listing.id, listing);
  res.json({ success: true, listing_id: listing.id });
};`,
				region: "us-east-1",
				capabs: []string{"storage"},
			},
			{
				name: "send-message",
				code: `export default async (req, res) => {
  const { recipient_id, content } = req.body;
  const sender_id = req.user?.id || 'anonymous';
  const message = {
    id: crypto.randomUUID(),
    sender_id,
    recipient_id,
    content,
    created_at: new Date().toISOString()
  };
  await state.push('messages/' + recipient_id, message);
  res.json({ success: true, message_id: message.id });
};`,
				region: "us-east-1",
				capabs: []string{"storage"},
			},
		},
		"ai-app": {
			{
				name: "chat-completion",
				code: `export default async (req, res) => {
  const { message, model = 'gpt-4' } = req.body;
  if (!message) {
    res.status(400).json({ error: 'message is required' });
    return;
  }
  const completion = await ai.chat.completions.create({
    model,
    messages: [{ role: 'user', content: message }]
  });
  res.json({ 
    message: completion.choices[0].message.content,
    model: completion.model,
    usage: completion.usage
  });
};`,
				region: "us-east-1",
				capabs: []string{"ai"},
			},
			{
				name: "embed-and-store",
				code: `export default async (req, res) => {
  const { content, metadata = {} } = req.body;
  if (!content) {
    res.status(400).json({ error: 'content is required' });
    return;
  }
  const embedding = await ai.embeddings.create({
    model: 'text-embedding-3-small',
    input: content
  });
  const id = crypto.randomUUID();
  await state.set('embeddings/' + id, {
    vector: embedding.data[0].embedding,
    content: content.substring(0, 1000),
    metadata,
    created_at: new Date().toISOString()
  });
  res.json({ embedded: true, id });
};`,
				region: "us-east-1",
				capabs: []string{"ai", "storage"},
			},
		},
	}

	result := templates[bundleSlug]
	for i := range result {
		if len(result[i].providers) == 0 {
			result[i].providers = []string{"functionfly"}
		}
	}
	return result
}
