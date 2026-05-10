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

// AnalyticsProvisioner creates the isolated analytics infrastructure for a tenant.
// All analytics data lives in the tenant's own dedicated database.
//
// What it provisions:
//   - Default analytics dashboard with standard widgets
//   - Pre-built funnel definitions (signup, activation, retention, revenue)
//   - Daily metric rollup configuration
//   - Cohort analysis setup
//   - Event taxonomy (standard SaaS events pre-registered)
type AnalyticsProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewAnalyticsProvisioner creates a new Analytics provisioner
func NewAnalyticsProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *AnalyticsProvisioner {
	return &AnalyticsProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision sets up the complete analytics infrastructure in the tenant's database.
func (ap *AnalyticsProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "analytics",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := ap.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Create default dashboard with standard SaaS widgets
	defaultDashboardLayout := []map[string]interface{}{
		{
			"widget_type": "metric_card",
			"title":       "Active Users",
			"position":    map[string]int{"x": 0, "y": 0, "w": 3, "h": 2},
			"config":      map[string]interface{}{"metric": "active_users", "period": "7d", "comparison": "previous_period"},
		},
		{
			"widget_type": "metric_card",
			"title":       "New Signups",
			"position":    map[string]int{"x": 3, "y": 0, "w": 3, "h": 2},
			"config":      map[string]interface{}{"metric": "signups", "period": "7d", "comparison": "previous_period"},
		},
		{
			"widget_type": "metric_card",
			"title":       "MRR",
			"position":    map[string]int{"x": 6, "y": 0, "w": 3, "h": 2},
			"config":      map[string]interface{}{"metric": "mrr", "period": "30d", "format": "currency"},
		},
		{
			"widget_type": "metric_card",
			"title":       "Conversion Rate",
			"position":    map[string]int{"x": 9, "y": 0, "w": 3, "h": 2},
			"config":      map[string]interface{}{"metric": "conversion_rate", "period": "30d", "format": "percent"},
		},
		{
			"widget_type": "line_chart",
			"title":       "User Growth",
			"position":    map[string]int{"x": 0, "y": 2, "w": 6, "h": 4},
			"config":      map[string]interface{}{"metrics": []string{"signups", "active_users"}, "period": "30d"},
		},
		{
			"widget_type": "line_chart",
			"title":       "Revenue Trend",
			"position":    map[string]int{"x": 6, "y": 2, "w": 6, "h": 4},
			"config":      map[string]interface{}{"metrics": []string{"revenue", "mrr"}, "period": "30d", "format": "currency"},
		},
		{
			"widget_type": "funnel_chart",
			"title":       "Signup → Activation Funnel",
			"position":    map[string]int{"x": 0, "y": 6, "w": 6, "h": 4},
			"config":      map[string]interface{}{"funnel": "signup-activation"},
		},
		{
			"widget_type": "bar_chart",
			"title":       "Top Events",
			"position":    map[string]int{"x": 6, "y": 6, "w": 6, "h": 4},
			"config":      map[string]interface{}{"metric": "event_count", "group_by": "event_name", "period": "7d"},
		},
		{
			"widget_type": "table",
			"title":       "Recent Activity",
			"position":    map[string]int{"x": 0, "y": 10, "w": 12, "h": 4},
			"config":      map[string]interface{}{"source": "recent_events", "limit": 20},
		},
	}

	layoutJSON, _ := json.Marshal(defaultDashboardLayout)

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_analytics_dashboards (id, tenant_id, name, layout, is_default, created_at, updated_at)
		 VALUES ($1, $2, 'Overview', $3, true, NOW(), NOW())
		 ON CONFLICT DO NOTHING`,
		uuid.New(), tenantID, layoutJSON)
	if err != nil {
		log.WithError(err).Warn("Failed to create default dashboard (non-fatal)")
	}
	log.Info("Default analytics dashboard created")

	// 3. Create default funnels
	funnels := []struct {
		name        string
		description string
		steps       []map[string]interface{}
	}{
		{
			"Signup → Activation",
			"Track how many signups become activated users",
			[]map[string]interface{}{
				{"event_name": "user_signup"},
				{"event_name": "profile_completed"},
				{"event_name": "first_action"},
				{"event_name": "activated"},
			},
		},
		{
			"Free → Paid Conversion",
			"Track conversion from free trial to paid subscription",
			[]map[string]interface{}{
				{"event_name": "trial_started"},
				{"event_name": "feature_used"},
				{"event_name": "subscription_created"},
			},
		},
		{
			"Onboarding Completion",
			"Track onboarding step completion rates",
			[]map[string]interface{}{
				{"event_name": "onboarding_started"},
				{"event_name": "onboarding_step_1"},
				{"event_name": "onboarding_step_2"},
				{"event_name": "onboarding_step_3"},
				{"event_name": "onboarding_completed"},
			},
		},
		{
			"Payment Funnel",
			"Track payment flow from checkout to success",
			[]map[string]interface{}{
				{"event_name": "checkout_started"},
				{"event_name": "payment_submitted"},
				{"event_name": "payment_success"},
			},
		},
	}

	for _, f := range funnels {
		stepsJSON, _ := json.Marshal(f.steps)
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_funnels (id, tenant_id, name, description, steps, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, f.name, f.description, stepsJSON)
		if err != nil {
			log.WithError(err).WithField("funnel", f.name).Warn("Failed to create funnel (non-fatal)")
		}
	}
	log.WithField("count", len(funnels)).Info("Default funnels created")

	// 4. Create default cohort analysis
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_analytics_cohorts (id, tenant_id, name, cohort_type, cohort_config, is_active, created_at)
		 VALUES ($1, $2, 'Weekly Signup Retention', 'signup_date', '{"period":"week","retention_days":[1,7,14,30,60,90]}', true, NOW())
		 ON CONFLICT DO NOTHING`,
		uuid.New(), tenantID)
	if err != nil {
		log.WithError(err).Warn("Failed to create cohort (non-fatal)")
	}

	// 5. Seed some example analytics events for immediate dashboard value
	exampleEvents := []struct {
		name     string
		category string
	}{
		{"page_view", "engagement"},
		{"user_signup", "conversion"},
		{"profile_completed", "engagement"},
		{"first_action", "activation"},
		{"feature_used", "engagement"},
		{"checkout_started", "revenue"},
		{"payment_success", "revenue"},
		{"subscription_created", "revenue"},
		{"subscription_canceled", "retention"},
		{"trial_started", "conversion"},
		{"trial_ended", "retention"},
		{"invite_sent", "viral"},
		{"export_downloaded", "engagement"},
		{"settings_updated", "engagement"},
		{"api_key_created", "activation"},
	}

	for _, ev := range exampleEvents {
		// Insert as a "definition" event (created_at in the past, used for event taxonomy)
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_events (id, tenant_id, event_name, event_category, properties, created_at)
			 VALUES ($1, $2, $3, $4, '{"type":"definition","description":"Standard SaaS event"}', NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, ev.name, ev.category)
		if err != nil {
			// Non-fatal: table might not have unique constraint on event_name
		}
	}
	log.WithField("count", len(exampleEvents)).Info("Standard event taxonomy seeded")

	// 6. Initialize daily rollup config in tenant configs
	_, err = pool.Exec(ctx,
		`UPDATE tenant_configs SET
		 	settings = settings || '{"analytics":{"daily_rollup_enabled":true,"retention_days":365,"real_time_window_minutes":5}}',
		 	updated_at = NOW()
		 WHERE tenant_id = $1`,
		tenantID)
	if err != nil {
		log.WithError(err).Warn("Failed to update analytics config (non-fatal)")
	}

	state.Status = StatusActive
	state.ResourceID = "analytics:default"
	log.Info("Analytics provisioning complete")
	return state, nil
}
