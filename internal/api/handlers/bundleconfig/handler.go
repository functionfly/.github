package bundleconfig

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/apputil"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/provisioning"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Handler provides endpoints for reading/updating bundle component configs
// stored in the tenant's isolated database.
type Handler struct {
	bundleProv *provisioning.BundleProvisioner
	repo       storage.Repository
}

// NewHandler creates a new bundle config handler.
func NewHandler(bp *provisioning.BundleProvisioner, repo storage.Repository) *Handler {
	return &Handler{bundleProv: bp, repo: repo}
}

// bundleConfigResponse is the top-level response for GET /apps/{appId}/bundle/config.
type bundleConfigResponse struct {
	BundleSlug string           `json:"bundle_slug"`
	TenantID   string           `json:"tenant_id"`
	Auth       *authConfig      `json:"auth"`
	Payments   *paymentsConfig  `json:"payments"`
	Email      *emailConfig     `json:"email"`
	Analytics  *analyticsConfig `json:"analytics"`
}

type authConfig struct {
	OAuthProviders []oauthProvider `json:"oauth_providers"`
	HasJWTKey      bool            `json:"has_jwt_key"`
	JWTKeyID       string          `json:"jwt_key_id"`
	EmailTemplates int             `json:"email_templates"`
}

type oauthProvider struct {
	Provider  string `json:"provider"`
	Enabled   bool   `json:"enabled"`
	ClientID  string `json:"client_id,omitempty"`
	HasSecret bool   `json:"has_secret"`
}

type paymentsConfig struct {
	Mode             string         `json:"mode"`
	DefaultCurrency  string         `json:"default_currency"`
	TaxMode          string         `json:"tax_calculation_mode"`
	WebhookURL       string         `json:"webhook_url"`
	Products         []productInfo  `json:"products"`
	EmailTemplates   int            `json:"email_templates"`
}

type productInfo struct {
	Name   string       `json:"name"`
	Active bool         `json:"active"`
	Prices []priceInfo  `json:"prices"`
}

type priceInfo struct {
	AmountCents int    `json:"amount_cents"`
	Currency    string `json:"currency"`
	Interval    string `json:"interval"`
	TrialDays   int    `json:"trial_days"`
}

type emailConfig struct {
	TransactionTemplates int              `json:"transaction_templates"`
	WorkflowTemplates    int              `json:"workflow_templates"`
	Workflows            []workflowInfo   `json:"workflows"`
}

type workflowInfo struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Trigger  string `json:"trigger_type"`
	Active   bool   `json:"is_active"`
	Steps    int    `json:"steps"`
}

type analyticsConfig struct {
	Dashboards int    `json:"dashboards"`
	Funnels    int    `json:"funnels"`
	Events     int    `json:"events"`
	RetentionDays int `json:"retention_days"`
}

// HandleGetBundleConfig returns all bundle component configs for an app's tenant.
// GET /v1/apps/{appId}/bundle/config
func (h *Handler) HandleGetBundleConfig(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	app, resolveErr := apputil.ResolveAppForRequest(ctx, h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	if h.bundleProv.TenantDB() == nil {
		apierror.WriteError(w, apierror.NewInternal("Tenant database not configured"))
		return
	}

	pool, err := h.bundleProv.TenantDB().GetTenantPool(ctx, app.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", app.TenantID).Error("Failed to get tenant pool")
		apierror.WriteError(w, apierror.NewInternal("Failed to connect to tenant database"))
		return
	}

	resp := bundleConfigResponse{
		TenantID: app.TenantID.String(),
	}

	resp.Auth = h.readAuthConfig(pool)
	resp.Payments = h.readPaymentsConfig(pool)
	resp.Email = h.readEmailConfig(pool)
	resp.Analytics = h.readAnalyticsConfig(pool)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleUpdateOAuthProvider enables/disables an OAuth provider or updates its credentials.
// PUT /v1/apps/{appId}/bundle/config/auth/oauth/{provider}
func (h *Handler) HandleUpdateOAuthProvider(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	app, resolveErr := apputil.ResolveAppForRequest(ctx, h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	vars := mux.Vars(r)
	provider := vars["provider"]

	var req struct {
		Enabled  bool   `json:"enabled"`
		ClientID string `json:"client_id,omitempty"`
		Secret   string `json:"secret,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	pool, err := h.bundleProv.TenantDB().GetTenantPool(ctx, app.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to connect to tenant database"))
		return
	}

	_, err = pool.Exec(ctx,
		`UPDATE tenant_oauth_configs SET enabled = $1, client_id = $2, updated_at = NOW()
		 WHERE tenant_id = $3 AND provider = $4`,
		req.Enabled, req.ClientID, app.TenantID.String(), provider)
	if err != nil {
		logrus.WithError(err).Error("Failed to update OAuth provider")
		apierror.WriteError(w, apierror.NewInternal("Failed to update OAuth provider"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleToggleWorkflow enables/disables an email workflow.
// PUT /v1/apps/{appId}/bundle/config/email/workflows/{slug}
func (h *Handler) HandleToggleWorkflow(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	app, resolveErr := apputil.ResolveAppForRequest(ctx, h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	var req struct {
		Active bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	pool, err := h.bundleProv.TenantDB().GetTenantPool(ctx, app.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to connect to tenant database"))
		return
	}

	tag, err := pool.Exec(ctx,
		`UPDATE tenant_email_workflows SET is_active = $1, updated_at = NOW()
		 WHERE tenant_id = $2 AND slug = $3`,
		req.Active, app.TenantID.String(), slug)
	if err != nil {
		logrus.WithError(err).Error("Failed to toggle workflow")
		apierror.WriteError(w, apierror.NewInternal("Failed to toggle workflow"))
		return
	}
	if tag.RowsAffected() == 0 {
		apierror.WriteError(w, apierror.NewNotFound("Workflow not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ─── Internal readers ────────────────────────────────────────────────────────

func (h *Handler) readAuthConfig(pool *pgxpool.Pool) *authConfig {
	ctx := sqlQueryCtx()
	cfg := &authConfig{}

	// JWT key
	var keyID, algorithm string
	err := pool.QueryRow(ctx,
		`SELECT key_id, algorithm FROM tenant_auth_keys WHERE is_active = true LIMIT 1`).
		Scan(&keyID, &algorithm)
	if err == nil {
		cfg.HasJWTKey = true
		cfg.JWTKeyID = keyID
	}

	// OAuth providers
	rows, err := pool.Query(ctx,
		`SELECT provider, enabled, client_id, encrypted_client_secret FROM tenant_oauth_configs ORDER BY provider`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p oauthProvider
			var clientID, encSecret sql.NullString
			rows.Scan(&p.Provider, &p.Enabled, &clientID, &encSecret)
			if clientID.Valid {
				p.ClientID = clientID.String
			}
			p.HasSecret = encSecret.Valid && encSecret.String != ""
			cfg.OAuthProviders = append(cfg.OAuthProviders, p)
		}
	}

	// Email template count
	var count int
	pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_email_templates WHERE category = 'auth'`).
		Scan(&count)
	cfg.EmailTemplates = count

	return cfg
}

func (h *Handler) readPaymentsConfig(pool *pgxpool.Pool) *paymentsConfig {
	ctx := sqlQueryCtx()
	cfg := &paymentsConfig{}

	err := pool.QueryRow(ctx,
		`SELECT payment_mode, default_currency, tax_calculation_mode, webhook_endpoint_url
		 FROM tenant_payment_config LIMIT 1`).
		Scan(&cfg.Mode, &cfg.DefaultCurrency, &cfg.TaxMode, &cfg.WebhookURL)
	if err != nil {
		return cfg
	}

	// Products with prices
	rows, err := pool.Query(ctx,
		`SELECT p.name, p.active, pr.amount_cents, pr.currency, pr.interval, pr.trial_days
		 FROM tenant_products p
		 LEFT JOIN tenant_prices pr ON pr.product_id = p.id
		 ORDER BY p.name, pr.amount_cents`)
	if err == nil {
		defer rows.Close()
		productMap := map[string]*productInfo{}
		var order []string
		for rows.Next() {
			var name, currency, interval sql.NullString
			var active sql.NullBool
			var amount, trial sql.NullInt32
			rows.Scan(&name, &active, &amount, &currency, &interval, &trial)
			n := name.String
			if _, ok := productMap[n]; !ok {
				productMap[n] = &productInfo{Name: n, Active: active.Bool}
				order = append(order, n)
			}
			if amount.Valid {
				productMap[n].Prices = append(productMap[n].Prices, priceInfo{
					AmountCents: int(amount.Int32),
					Currency:    currency.String,
					Interval:    interval.String,
					TrialDays:   int(trial.Int32),
				})
			}
		}
		for _, n := range order {
			cfg.Products = append(cfg.Products, *productMap[n])
		}
	}

	var count int
	pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_email_templates WHERE category = 'billing'`).
		Scan(&count)
	cfg.EmailTemplates = count

	return cfg
}

func (h *Handler) readEmailConfig(pool *pgxpool.Pool) *emailConfig {
	ctx := sqlQueryCtx()
	cfg := &emailConfig{}

	pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_email_templates WHERE category = 'transactional'`).
		Scan(&cfg.TransactionTemplates)
	pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_email_templates WHERE category = 'workflow'`).
		Scan(&cfg.WorkflowTemplates)

	rows, err := pool.Query(ctx,
		`SELECT w.slug, w.name, w.trigger_type, w.is_active,
		        (SELECT COUNT(*) FROM tenant_email_workflow_steps s WHERE s.workflow_id = w.id) as steps
		 FROM tenant_email_workflows w ORDER BY w.name`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var wf workflowInfo
			rows.Scan(&wf.Slug, &wf.Name, &wf.Trigger, &wf.Active, &wf.Steps)
			cfg.Workflows = append(cfg.Workflows, wf)
		}
	}

	return cfg
}

func (h *Handler) readAnalyticsConfig(pool *pgxpool.Pool) *analyticsConfig {
	ctx := sqlQueryCtx()
	cfg := &analyticsConfig{}

	pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenant_analytics_dashboards`).Scan(&cfg.Dashboards)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenant_analytics_funnels`).Scan(&cfg.Funnels)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenant_analytics_events`).Scan(&cfg.Events)

	// Read retention from tenant_configs
	var settingsJSON []byte
	err := pool.QueryRow(ctx,
		`SELECT settings FROM tenant_configs LIMIT 1`).Scan(&settingsJSON)
	if err == nil && len(settingsJSON) > 0 {
		var settings map[string]interface{}
		json.Unmarshal(settingsJSON, &settings)
		if analytics, ok := settings["analytics"].(map[string]interface{}); ok {
			if rd, ok := analytics["retention_days"].(float64); ok {
				cfg.RetentionDays = int(rd)
			}
		}
	}
	if cfg.RetentionDays == 0 {
		cfg.RetentionDays = 365
	}

	return cfg
}

func sqlQueryCtx() context.Context {
	return context.Background()
}
