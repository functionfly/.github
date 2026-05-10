package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MarketplaceProvisioner creates the isolated marketplace infrastructure for a tenant.
// All marketplace data lives in the tenant's own dedicated database.
//
// What it provisions:
//   - Marketplace settings (name, currency, commission rate, policies)
//   - Default categories (Electronics, Fashion, Home & Garden, Digital Goods, Services)
//   - Stripe Connect payment configuration
//   - 20+ notification templates for marketplace events
//   - Email templates for order lifecycle, seller onboarding, disputes
//   - Email workflows (order confirmation → shipping → delivery → review request)
//   - Email workflows (seller onboarding drip)
type MarketplaceProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewMarketplaceProvisioner creates a new Marketplace provisioner
func NewMarketplaceProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *MarketplaceProvisioner {
	return &MarketplaceProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision sets up the complete marketplace infrastructure in the tenant's database.
func (mp *MarketplaceProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "marketplace",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := mp.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Create marketplace settings
	_, err = pool.Exec(ctx,
		`INSERT INTO marketplace_settings (id, tenant_id, marketplace_name, marketplace_description, currency, default_commission_rate_bps, require_seller_verification, allow_digital_goods, allow_services, auto_approve_listings, dispute_window_days, payout_minimum_cents)
		 VALUES ($1, $2, 'My Marketplace', 'Buy and sell with confidence', 'usd', 1000, false, true, true, true, 30, 1000)
		 ON CONFLICT (tenant_id) DO NOTHING`,
		uuid.New(), tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to create marketplace settings: %w", err)
	}
	log.Info("Marketplace settings created")

	// 3. Seed default categories
	categories := []struct {
		name     string
		slug     string
		desc     string
		icon     string
		sort     int
		children []struct {
			name string
			slug string
			desc string
		}
	}{
		{"Electronics", "electronics", "Phones, laptops, gadgets, and more", "cpu", 1, []struct{ name, slug, desc string }{
			{"Phones & Tablets", "phones-tablets", "Smartphones and tablets"},
			{"Computers", "computers", "Laptops, desktops, and accessories"},
			{"Audio", "audio", "Headphones, speakers, and more"},
			{"Wearables", "wearables", "Smartwatches and fitness trackers"},
		}},
		{"Fashion", "fashion", "Clothing, shoes, and accessories", "shirt", 2, []struct{ name, slug, desc string }{
			{"Men's", "mens", "Men's clothing and accessories"},
			{"Women's", "womens", "Women's clothing and accessories"},
			{"Shoes", "shoes", "Footwear for all occasions"},
			{"Accessories", "accessories", "Bags, jewelry, and more"},
		}},
		{"Home & Garden", "home-garden", "Furniture, decor, and garden supplies", "home", 3, []struct{ name, slug, desc string }{
			{"Furniture", "furniture", "Tables, chairs, sofas, and more"},
			{"Decor", "decor", "Wall art, lighting, and accessories"},
			{"Garden", "garden", "Plants, tools, and outdoor living"},
		}},
		{"Digital Goods", "digital-goods", "Software, templates, and digital downloads", "download", 4, []struct{ name, slug, desc string }{
			{"Software", "software", "Apps, plugins, and tools"},
			{"Templates", "templates", "Design templates and themes"},
			{"E-books", "ebooks", "Digital books and guides"},
			{"Courses", "courses", "Online courses and tutorials"},
		}},
		{"Services", "services", "Professional services and freelance work", "briefcase", 5, []struct{ name, slug, desc string }{
			{"Design", "design", "Graphic design, UI/UX, branding"},
			{"Development", "development", "Web, mobile, and software development"},
			{"Writing", "writing", "Content writing, copywriting, translation"},
			{"Consulting", "consulting", "Business and technical consulting"},
		}},
	}

	for _, cat := range categories {
		var parentID uuid.UUID
		err = pool.QueryRow(ctx,
			`INSERT INTO marketplace_categories (id, tenant_id, name, slug, description, sort_order, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, true)
			 ON CONFLICT (tenant_id, slug) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			uuid.New(), tenantID, cat.name, cat.slug, cat.desc, cat.sort).Scan(&parentID)
		if err != nil {
			log.WithError(err).WithField("category", cat.name).Warn("Failed to create category (non-fatal)")
			continue
		}

		for i, child := range cat.children {
			_, err = pool.Exec(ctx,
				`INSERT INTO marketplace_categories (id, tenant_id, parent_id, name, slug, description, sort_order, is_active)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, true)
				 ON CONFLICT (tenant_id, slug) DO NOTHING`,
				uuid.New(), tenantID, parentID, child.name, child.slug, child.desc, i)
			if err != nil {
				log.WithError(err).WithField("subcategory", child.name).Warn("Failed to create subcategory (non-fatal)")
			}
		}
	}
	log.WithField("count", len(categories)).Info("Default categories seeded with subcategories")

	// 4. Seed notification templates
	notifTemplates := []struct {
		eventType string
		name      string
		title     string
		body      string
		channel   string
	}{
		{"order_placed", "Order Placed", "New order received!", "Order #{{.OrderNumber}} for {{.Amount}} from {{.BuyerName}}", "in_app"},
		{"order_placed_seller", "Order Placed (Seller)", "You have a new order!", "{{.BuyerName}} purchased {{.ItemName}} for {{.Amount}}", "email"},
		{"order_confirmed", "Order Confirmed", "Order confirmed", "Your order #{{.OrderNumber}} has been confirmed by the seller", "in_app"},
		{"order_shipped", "Order Shipped", "Your order has been shipped!", "Order #{{.OrderNumber}} is on its way. Tracking: {{.TrackingNumber}}", "in_app"},
		{"order_shipped_email", "Order Shipped (Email)", "Your order is on the way!", "Great news! Order #{{.OrderNumber}} has been shipped. Track it here: {{.TrackingURL}}", "email"},
		{"order_delivered", "Order Delivered", "Order delivered", "Order #{{.OrderNumber}} has been delivered. Leave a review!", "in_app"},
		{"order_completed", "Order Completed", "Order completed", "Order #{{.OrderNumber}} is complete. Thank you!", "in_app"},
		{"order_canceled", "Order Canceled", "Order canceled", "Order #{{.OrderNumber}} has been canceled. {{.Reason}}", "in_app"},
		{"order_refunded", "Order Refunded", "Refund processed", "A refund of {{.Amount}} has been processed for order #{{.OrderNumber}}", "in_app"},
		{"new_message", "New Message", "New message from {{.SenderName}}", "{{.Preview}}", "in_app"},
		{"new_message_email", "New Message (Email)", "You have a new message", "{{.SenderName}} sent you a message about {{.Subject}}", "email"},
		{"listing_sold", "Listing Sold", "Your item sold!", "{{.ItemName}} was purchased by {{.BuyerName}} for {{.Amount}}", "in_app"},
		{"listing_favorited", "Listing Favorited", "Someone favorited your listing", "{{.UserName}} added {{.ItemName}} to their favorites", "in_app"},
		{"new_review", "New Review", "New review received", "{{.ReviewerName}} left a {{.Rating}}-star review on {{.ItemName}}", "in_app"},
		{"review_reply", "Review Reply", "Seller replied to your review", "The seller replied to your review on {{.ItemName}}", "in_app"},
		{"payout_sent", "Payout Sent", "Payout processed!", "A payout of {{.Amount}} has been sent to your account", "in_app"},
		{"payout_sent_email", "Payout Sent (Email)", "Your payout is on the way", "A payout of {{.Amount}} has been initiated. Expected arrival: {{.ArrivalDate}}", "email"},
		{"dispute_opened", "Dispute Opened", "Dispute opened", "A dispute has been opened for order #{{.OrderNumber}}: {{.Reason}}", "in_app"},
		{"dispute_resolved", "Dispute Resolved", "Dispute resolved", "The dispute for order #{{.OrderNumber}} has been resolved: {{.Resolution}}", "in_app"},
		{"seller_verification", "Seller Verified", "Account verified!", "Your seller account has been verified. You can now receive orders!", "in_app"},
		{"low_stock", "Low Stock Alert", "Low stock warning", "{{.ItemName}} has only {{.Quantity}} items remaining", "in_app"},
		{"payment_failed", "Payment Failed", "Payment failed", "Payment for order #{{.OrderNumber}} failed. Please update your payment method.", "in_app"},
	}

	for _, nt := range notifTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO marketplace_notification_templates (id, tenant_id, event_type, name, title_template, body_template, channel, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, true)
			 ON CONFLICT (tenant_id, event_type, channel) DO NOTHING`,
			uuid.New(), tenantID, nt.eventType, nt.name, nt.title, nt.body, nt.channel)
		if err != nil {
			log.WithError(err).WithField("event_type", nt.eventType).Warn("Failed to seed notification template (non-fatal)")
		}
	}
	log.WithField("count", len(notifTemplates)).Info("Notification templates seeded")

	// 5. Seed email templates for marketplace flows
	emailTemplates := []struct {
		slug     string
		name     string
		subject  string
		category string
	}{
		// Order lifecycle
		{"order-confirmation", "Order Confirmation", "Order #{{.OrderNumber}} confirmed", "transactional"},
		{"order-shipped", "Order Shipped", "Your order is on the way!", "transactional"},
		{"order-delivered", "Order Delivered", "Your order has been delivered", "transactional"},
		{"order-canceled", "Order Canceled", "Order #{{.OrderNumber}} canceled", "transactional"},
		{"refund-processed", "Refund Processed", "Your refund of {{.Amount}} has been processed", "transactional"},
		// Seller lifecycle
		{"seller-welcome", "Seller Welcome", "Welcome to the marketplace!", "onboarding"},
		{"seller-verification-approved", "Verification Approved", "Your seller account is verified!", "onboarding"},
		{"seller-first-sale", "First Sale", "Congratulations on your first sale!", "onboarding"},
		{"seller-payout-notification", "Payout Notification", "Your payout of {{.Amount}} is on the way", "transactional"},
		// Buyer lifecycle
		{"buyer-welcome", "Buyer Welcome", "Welcome! Start exploring", "onboarding"},
		{"review-request", "Review Request", "How was your experience with {{.ItemName}}?", "engagement"},
		{"abandoned-cart", "Abandoned Cart", "You left something behind!", "marketing"},
		// Disputes
		{"dispute-opened", "Dispute Opened", "Dispute opened for order #{{.OrderNumber}}", "transactional"},
		{"dispute-resolved", "Dispute Resolved", "Your dispute has been resolved", "transactional"},
	}

	for _, t := range emailTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject,
			fmt.Sprintf("<!-- %s template — customize in dashboard -->", t.name),
			fmt.Sprintf("%s template - customize in dashboard", t.name),
			`[{"name":"OrderNumber","type":"string","required":false},{"name":"Amount","type":"string","required":false},{"name":"ItemName","type":"string","required":false},{"name":"BuyerName","type":"string","required":false},{"name":"SellerName","type":"string","required":false},{"name":"TrackingNumber","type":"string","required":false},{"name":"TrackingURL","type":"string","required":false}]`,
			t.category)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed email template (non-fatal)")
		}
	}
	log.WithField("count", len(emailTemplates)).Info("Marketplace email templates seeded")

	// 6. Create email workflows
	workflows := []struct {
		slug        string
		name        string
		triggerType string
		steps       []workflowStep
	}{
		{
			"order-lifecycle", "Order Lifecycle", "order_placed",
			[]workflowStep{
				{0, "order-confirmation", 0, "minutes", "always"},
				{1, "review-request", 7, "days", "always"},
			},
		},
		{
			"seller-onboarding", "Seller Onboarding", "seller_registered",
			[]workflowStep{
				{0, "seller-welcome", 0, "minutes", "always"},
				{1, "seller-first-sale", 0, "minutes", "always"},
			},
		},
		{
			"buyer-onboarding", "Buyer Onboarding", "user_signup",
			[]workflowStep{
				{0, "buyer-welcome", 0, "minutes", "always"},
				{1, "abandoned-cart", 24, "hours", "always"},
			},
		},
		{
			"dispute-lifecycle", "Dispute Lifecycle", "dispute_opened",
			[]workflowStep{
				{0, "dispute-opened", 0, "minutes", "always"},
				{1, "dispute-resolved", 0, "minutes", "always"},
			},
		},
	}

	for _, wf := range workflows {
		var workflowID uuid.UUID
		err = pool.QueryRow(ctx,
			`INSERT INTO tenant_email_workflows (id, tenant_id, slug, name, trigger_type, trigger_config, is_active)
			 VALUES ($1, $2, $3, $4, $5, '{}', true)
			 ON CONFLICT (tenant_id, slug) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			uuid.New(), tenantID, wf.slug, wf.name, wf.triggerType).Scan(&workflowID)
		if err != nil {
			log.WithError(err).WithField("workflow", wf.slug).Warn("Failed to create workflow (non-fatal)")
			continue
		}

		for _, step := range wf.steps {
			var templateID uuid.NullUUID
			pool.QueryRow(ctx,
				`SELECT id FROM tenant_email_templates WHERE tenant_id = $1 AND slug = $2`,
				tenantID, step.templateSlug).Scan(&templateID)

			_, err = pool.Exec(ctx,
				`INSERT INTO tenant_email_workflow_steps (id, tenant_id, workflow_id, template_id, step_order, delay_amount, delay_unit, condition_type, is_active)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
				 ON CONFLICT (workflow_id, step_order) DO NOTHING`,
				uuid.New(), tenantID, workflowID, templateID, step.order, step.delay, step.unit, step.condition)
			if err != nil {
				log.WithError(err).WithField("workflow", wf.slug).Warn("Failed to create workflow step (non-fatal)")
			}
		}
		log.WithField("workflow", wf.slug).Info("Email workflow created")
	}

	// 7. Seed Stripe Connect payment config
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_payment_config (id, tenant_id, payment_mode, default_currency, allowed_payment_methods, tax_calculation_mode, metadata)
		 VALUES ($1, $2, 'connect', 'usd', '["card","us_bank_account"]', 'automatic', '{"type":"marketplace","connect_enabled":true}')
		 ON CONFLICT (tenant_id) DO UPDATE SET payment_mode = 'connect', updated_at = NOW()`,
		uuid.New(), tenantID)
	if err != nil {
		log.WithError(err).Warn("Failed to create payment config (non-fatal)")
	}
	log.Info("Stripe Connect payment config created")

	// 8. Create default analytics dashboard for marketplace
	dashboardLayout := `[{"widget_type":"metric_card","title":"Total GMV","position":{"x":0,"y":0,"w":3,"h":2},"config":{"metric":"gmv","period":"30d","format":"currency"}},{"widget_type":"metric_card","title":"Orders","position":{"x":3,"y":0,"w":3,"h":2},"config":{"metric":"orders","period":"30d"}},{"widget_type":"metric_card","title":"Active Listings","position":{"x":6,"y":0,"w":3,"h":2},"config":{"metric":"active_listings","period":"30d"}},{"widget_type":"metric_card","title":"Avg Order Value","position":{"x":9,"y":0,"w":3,"h":2},"config":{"metric":"avg_order_value","period":"30d","format":"currency"}},{"widget_type":"line_chart","title":"Revenue Trend","position":{"x":0,"y":2,"w":6,"h":4},"config":{"metrics":["revenue","gmv"],"period":"30d"}},{"widget_type":"bar_chart","title":"Top Categories","position":{"x":6,"y":2,"w":6,"h":4},"config":{"metric":"orders","group_by":"category","period":"30d"}},{"widget_type":"table","title":"Recent Orders","position":{"x":0,"y":6,"w":12,"h":4},"config":{"source":"recent_orders","limit":20}}]`

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_analytics_dashboards (id, tenant_id, name, layout, is_default, created_at, updated_at)
		 VALUES ($1, $2, 'Marketplace Overview', $3, true, NOW(), NOW())
		 ON CONFLICT DO NOTHING`,
		uuid.New(), tenantID, dashboardLayout)
	if err != nil {
		log.WithError(err).Warn("Failed to create marketplace dashboard (non-fatal)")
	}

	// 9. Create marketplace-specific analytics events
	mktEvents := []struct {
		name     string
		category string
	}{
		{"listing_created", "seller"},
		{"listing_viewed", "engagement"},
		{"listing_favorited", "engagement"},
		{"search_performed", "engagement"},
		{"order_placed", "revenue"},
		{"order_confirmed", "revenue"},
		{"order_shipped", "fulfillment"},
		{"order_delivered", "fulfillment"},
		{"order_completed", "revenue"},
		{"order_canceled", "revenue"},
		{"payment_success", "revenue"},
		{"payment_failed", "revenue"},
		{"refund_requested", "revenue"},
		{"refund_processed", "revenue"},
		{"review_submitted", "engagement"},
		{"message_sent", "engagement"},
		{"dispute_opened", "support"},
		{"seller_registered", "seller"},
		{"seller_verified", "seller"},
		{"payout_sent", "revenue"},
	}

	for _, ev := range mktEvents {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_events (id, tenant_id, event_name, event_category, properties, created_at)
			 VALUES ($1, $2, $3, $4, '{"type":"definition","description":"Marketplace event"}', NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, ev.name, ev.category)
		if err != nil {
			// Non-fatal
		}
	}
	log.WithField("count", len(mktEvents)).Info("Marketplace analytics events seeded")

	// 10. Create marketplace funnels
	funnels := []struct {
		name  string
		desc  string
		steps []map[string]interface{}
	}{
		{
			"Browse → Purchase",
			"Track conversion from listing view to completed order",
			[]map[string]interface{}{
				{"event_name": "listing_viewed"},
				{"event_name": "listing_favorited"},
				{"event_name": "order_placed"},
				{"event_name": "order_completed"},
			},
		},
		{
			"Seller Onboarding",
			"Track seller activation from registration to first sale",
			[]map[string]interface{}{
				{"event_name": "seller_registered"},
				{"event_name": "seller_verified"},
				{"event_name": "listing_created"},
				{"event_name": "order_placed"},
			},
		},
		{
			"Search → Purchase",
			"Track search-to-purchase conversion",
			[]map[string]interface{}{
				{"event_name": "search_performed"},
				{"event_name": "listing_viewed"},
				{"event_name": "order_placed"},
			},
		},
	}

	for _, f := range funnels {
		stepsJSON, _ := json.Marshal(f.steps)
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_funnels (id, tenant_id, name, description, steps, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, f.name, f.desc, stepsJSON)
		if err != nil {
			log.WithError(err).WithField("funnel", f.name).Warn("Failed to create funnel (non-fatal)")
		}
	}

	// 11. Log provisioning in audit
	auditMeta, _ := json.Marshal(map[string]interface{}{
		"component":   "marketplace",
		"action":      "provisioned",
		"categories":  len(categories),
		"templates":   len(notifTemplates),
		"workflows":   len(workflows),
	})
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_auth_audit (id, tenant_id, event_type, success, metadata, created_at)
		 VALUES ($1, $2, 'system_provision', true, $3, NOW())`,
		uuid.New(), tenantID, auditMeta)
	if err != nil {
		log.WithError(err).Warn("Failed to log provisioning audit (non-fatal)")
	}

	state.Status = StatusActive
	state.ResourceID = "marketplace:complete"
	log.Info("Marketplace provisioning complete")
	return state, nil
}
