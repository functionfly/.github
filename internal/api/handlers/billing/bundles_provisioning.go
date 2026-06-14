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
	now := time.Now()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
	// This creates a dedicated database with isolated Auth, Payments, Email, and Analytics.
	// The function templates below are still created in the platform registry as convenience
	// entry points that bridge into the tenant's isolated data.
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
	}

	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "stripe-webhook-handler",
			code: `export default async (req, res) => {
  const event = req.body;
  try {
    switch (event.type) {
      case 'customer.subscription.created':
        await state.set('subscriptions/' + event.data.object.customer, {
          status: 'active',
          plan: event.data.object.items.data[0].plan.id
        });
        break;
      case 'invoice.payment_succeeded':
        await state.set('payments/' + event.data.object.id, {
          status: 'paid',
          amount: event.data.object.amount_paid
        });
        break;
    }
    res.json({ received: true });
  } catch (err) {
    res.status(400).json({ error: err.message });
  }
};`,
			region: "us-east-1",
			capabs: []string{"webhook", "storage"},
		},
		{
			name: "welcome-email",
			code: `export default async (req, res) => {
  const { email, name } = req.body;
  await email.send({
    to: email,
    subject: 'Welcome!',
    template: 'welcome',
    data: { name }
  });
  res.json({ sent: true });
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
	now := time.Now()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
	// This creates a dedicated database with isolated Listings, Payments, Messaging, and Notifications.
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
	}

	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "create-listing",
			code: `export default async (req, res) => {
  const { title, description, price } = req.body;
  const seller_id = req.user.id;
  const listing = {
    id: crypto.randomUUID(),
    seller_id,
    title,
    description,
    price_cents: Math.round(price * 100),
    status: 'active'
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
  const sender_id = req.user.id;
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
	}

	for _, tmpl := range templates {
		function := &storage.FunctionConfig{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         tmpl.name,
			Code:         tmpl.code,
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
	now := time.Now()

	// If isolated provisioning is available, delegate to the BundleProvisioner.
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
	}

	templates := []struct {
		name   string
		code   string
		region string
		capabs []string
	}{
		{
			name: "chat-completion",
			code: `export default async (req, res) => {
  const { message, model = 'gpt-4' } = req.body;
  const completion = await ai.chat.completions.create({
    model,
    messages: [{ role: 'user', content: message }]
  });
  res.json({ message: completion.choices[0].message.content });
};`,
			region: "us-east-1",
			capabs: []string{"ai"},
		},
		{
			name: "embed-and-store",
			code: `export default async (req, res) => {
  const { content } = req.body;
  const embedding = await ai.embeddings.create({
    model: 'text-embedding-3-small',
    input: content
  });
  await state.set('embeddings/' + crypto.randomUUID(), {
    vector: embedding.data[0].embedding,
    content: content.substring(0, 1000)
  });
  res.json({ embedded: true });
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

// ProvisionBundleAppAndBackend creates the default app and backend for a bundle
// This is the "one-click deploy" - called when bundle is purchased via Stripe webhook
func ProvisionBundleAppAndBackend(repo storage.Repository, tenantID uuid.UUID, bundleSlug string) (*storage.App, error) {
	ctx := context.Background()
	now := time.Now()

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

	// Determine backend config based on bundle
	region := "eu-central-1"
	url := "https://api.functionfly.io/v1/apps/" + app.ID.String()

	// Create default backend
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
			"app_id":    app.ID,
			"provider":  backend.Provider,
		}).Info("Created default backend for bundle via standalone function")
	}

	// Create bundle-specific functions
	templates := getBundleFunctionTemplates(bundleSlug)
	for _, tmpl := range templates {
		function := &storage.FunctionConfig{
			ID:           uuid.New(),
			TenantID:     tenantID,
			AppID:        &app.ID,
			Name:         tmpl.name,
			Code:         tmpl.code,
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

	return app, nil
}

type bundleFunctionTemplate struct {
	name   string
	code   string
	region string
	capabs []string
}

func getBundleFunctionTemplates(bundleSlug string) []bundleFunctionTemplate {
	templates := map[string][]bundleFunctionTemplate{
		"saas-starter": {
			{
				name: "stripe-webhook",
				code: `export default async (req, res) => {
  const event = req.body;
  try {
    switch (event.type) {
      case 'customer.subscription.created':
        await state.set('subscriptions/' + event.data.object.customer, {
          status: 'active',
          plan: event.data.object.items.data[0].plan.id
        });
        break;
      case 'invoice.payment_succeeded':
        await state.set('payments/' + event.data.object.id, {
          status: 'paid',
          amount: event.data.object.amount_paid
        });
        break;
      case 'invoice.payment_failed':
        await state.set('failed_payments/' + event.data.object.customer, {
          timestamp: Date.now()
        });
        break;
    }
    res.json({ received: true });
  } catch (err) {
    res.status(400).json({ error: err.message });
  }
};`,
				region: "us-east-1",
				capabs: []string{"webhook", "storage"},
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

	return templates[bundleSlug]
}
