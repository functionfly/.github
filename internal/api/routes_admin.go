package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	"github.com/functionfly/functionfly/internal/api/handlers/content"
	factoryhandler "github.com/functionfly/functionfly/internal/api/handlers/factory"
	feedbackHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/feedback"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/handlers/security"
	"github.com/functionfly/functionfly/internal/api/handlers/statefabric"
	"github.com/functionfly/functionfly/internal/api/middleware"
	advancedsecurity "github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/gorilla/mux"
)

func registerAdminRoutes(
	s *Server,
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	advancedSecurityMiddleware *advancedsecurity.AdvancedSecurityMiddleware,
	adminHandler *admin.Handler,
	adminBackendsHandler *admin.BackendsHandler,
	adminProvidersHandler *admin.ProvidersHandler,
	maintenanceHandler *admin.MaintenanceHandler,
	feedbackHandler *feedbackHandlerPkg.Handler,
	monitoringHandler *monitoring.Handler,
	securityHandler *security.Handler,
	mfaHandler *mfaHandlerPkg.MFAHandler,
	adminRegistryHandler *admin.RegistryHandler,
	registryHandler *registryhandler.Handler,
	oversightHandler *admin.OversightHandler,
	factoryHandler *factoryhandler.Handler,
	stateFabricHandler *statefabric.Handler,
	contentHandler *content.Handler,
	csrfMiddleware *middleware.CSRFMiddleware,
	rateLimiter *middleware.AdminRateLimiter,
	sessionMiddleware *middleware.AdminSessionMiddleware,
	ipAllowlistMiddleware *middleware.IPAllowlistMiddleware,
	adminIPAllowlistHandler *admin.AdminIPAllowlistHandler,
	adminAuditHandler *admin.AdminAuditHandler,
	securityEventHandler *admin.SecurityEventHandler,
	alertHandler *admin.AlertHandler,
) {
	adminRoutes := api.PathPrefix("/admin").Subrouter()

	// ── Admin middleware wiring ──────────────────────────────────────────────────
	// Order: IP Allowlist → Session Validation → Security Alert (rate limit) → CSRF
	//
	// 1. IP Allowlist check (skips internal IPs automatically)
	adminRoutes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(ipAllowlistMiddleware.RequireIPAllowlist(http.HandlerFunc(next.ServeHTTP)))
	})

	// 2. Session validation middleware
	adminRoutes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(sessionMiddleware.RequireAdminSession(http.HandlerFunc(next.ServeHTTP)))
	})

	// 3. Security Alert middleware (includes AdvancedRateLimit for DDoS protection)
	adminRoutes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(advancedSecurityMiddleware.AdvancedRateLimit(http.HandlerFunc(next.ServeHTTP)))
	})

	// 4. CSRF middleware for mutating requests (POST, PUT, PATCH, DELETE)
	adminRoutes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(csrfMiddleware.RequireCSRF(http.HandlerFunc(next.ServeHTTP)))
	})

	// CSRF endpoint (no auth required - generates token for authenticated sessions)
	adminRoutes.HandleFunc("/csrf", csrfMiddleware.HandleGetCSRFToken).Methods("GET", "OPTIONS")

	// Note: MFA middleware is applied per-route after auth middleware to ensure claims are available
	adminRoutes.HandleFunc("/auth/session", authMiddleware.RequireAuth(adminHandler.HandleGetAdminSession)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/auth/session", authMiddleware.RequireAuth(adminHandler.HandleExtendAdminSession)).Methods("POST", "OPTIONS")

	// Tenant management
	adminRoutes.HandleFunc("/tenants", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenants)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/tenants", authMiddleware.RequirePermission(auth.PermTenantsWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateTenant))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/tenants/{tenantId}", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetTenant)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/tenants/{tenantId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateTenant))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/tenants/{tenantId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleDeleteTenant))).Methods("DELETE", "OPTIONS")

	// User management (read-only user list/stats do not require MFA so admins can always access the UI)
	adminRoutes.HandleFunc("/users", authMiddleware.RequirePermission(auth.PermUsersRead)(adminHandler.HandleListUsers)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/users/stats", authMiddleware.RequirePermission(auth.PermUsersRead)(adminHandler.HandleGetUserStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/users", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateUser))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/users/invite", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleInviteUser))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/users/{userId}", authMiddleware.RequirePermission(auth.PermUsersRead)(adminHandler.HandleGetUser)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/users/{userId}", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateUser))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/users/{userId}", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleDeleteUser))).Methods("DELETE", "OPTIONS")

	// Audit log
	adminRoutes.HandleFunc("/audit-events", authMiddleware.RequirePermission(auth.PermAuditRead)(adminHandler.HandleListAuditEvents)).Methods("GET", "OPTIONS")

	// IP Allowlist management
	adminRoutes.HandleFunc("/ip-allowlist", authMiddleware.RequirePermission(auth.PermSystemRead)(adminIPAllowlistHandler.HandleListIPAllowlist)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminIPAllowlistHandler.HandleCreateIPAllowlist))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(adminIPAllowlistHandler.HandleGetIPAllowlist)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminIPAllowlistHandler.HandleUpdateIPAllowlist))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminIPAllowlistHandler.HandleDeleteIPAllowlist))).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist/{id}/toggle", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminIPAllowlistHandler.HandleToggleIPAllowlist))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/ip-allowlist/check-access", authMiddleware.RequireAuth(adminIPAllowlistHandler.HandleCheckMyIPAccess)).Methods("GET", "OPTIONS")

	// Admin Audit log (new detailed audit system)
	adminRoutes.HandleFunc("/audit", authMiddleware.RequirePermission(auth.PermAuditRead)(adminAuditHandler.HandleListAuditLogs)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/audit/{id}", authMiddleware.RequirePermission(auth.PermAuditRead)(adminAuditHandler.HandleGetAuditLog)).Methods("GET", "OPTIONS")

	// Security Events
	adminRoutes.HandleFunc("/security-events", authMiddleware.RequirePermission(auth.PermSystemRead)(securityEventHandler.HandleListSecurityEvents)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/security-events/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(securityEventHandler.HandleGetSecurityEventStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/security-events/{id}/review", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(securityEventHandler.HandleReviewSecurityEvent))).Methods("POST", "OPTIONS")

	// Security Alert Rules
	adminRoutes.HandleFunc("/security-alerts", authMiddleware.RequirePermission(auth.PermSystemRead)(alertHandler.HandleListSecurityAlerts)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/security-alerts", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(alertHandler.HandleCreateSecurityAlert))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/security-alerts/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(alertHandler.HandleGetSecurityAlert)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/security-alerts/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(alertHandler.HandleUpdateSecurityAlert))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/security-alerts/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(alertHandler.HandleDeleteSecurityAlert))).Methods("DELETE", "OPTIONS")

	// Maintenance mode management
	adminRoutes.HandleFunc("/maintenance", authMiddleware.RequirePermission(auth.PermSystemRead)(maintenanceHandler.HandleGetMaintenanceStatus)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleEnableMaintenance))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleUpdateMaintenance))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleDisableMaintenance))).Methods("DELETE", "OPTIONS")

	// Maintenance templates
	adminRoutes.HandleFunc("/maintenance/templates", authMiddleware.RequirePermission(auth.PermSystemRead)(maintenanceHandler.HandleGetTemplates)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance/templates", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleCreateTemplate))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance/templates/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleUpdateTemplate))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance/templates/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(maintenanceHandler.HandleDeleteTemplate))).Methods("DELETE", "OPTIONS")

	// Maintenance scheduling and audit
	adminRoutes.HandleFunc("/maintenance/schedule", authMiddleware.RequirePermission(auth.PermSystemRead)(maintenanceHandler.HandleGetScheduledMaintenance)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/maintenance/audit", authMiddleware.RequirePermission(auth.PermSystemRead)(maintenanceHandler.HandleGetAuditLog)).Methods("GET", "OPTIONS")

	// Platform backends management
	adminRoutes.HandleFunc("/backends", authMiddleware.RequirePermission(auth.PermSystemRead)(adminBackendsHandler.HandleListPlatformBackends)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/backends/{backendId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminBackendsHandler.HandleUpdateBackendEnabled))).Methods("PATCH", "OPTIONS")

	// Provider management
	adminRoutes.HandleFunc("/providers", authMiddleware.RequirePermission(auth.PermSystemRead)(adminProvidersHandler.HandleListProviders)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/providers/{providerId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminProvidersHandler.HandleUpdateProvider))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/providers/{providerId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminProvidersHandler.HandleDeleteProvider))).Methods("DELETE", "OPTIONS")

	// Incident management
	adminRoutes.HandleFunc("/incidents", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleListIncidents)).Methods("GET")
	adminRoutes.HandleFunc("/incidents", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateIncident))).Methods("POST")
	adminRoutes.HandleFunc("/incidents/{incidentId}", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleGetIncident)).Methods("GET")
	adminRoutes.HandleFunc("/incidents/{incidentId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateIncident))).Methods("PATCH")
	adminRoutes.HandleFunc("/incidents/{incidentId}/resolve", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleResolveIncident))).Methods("POST")

	// System health and metrics
	adminRoutes.HandleFunc("/health", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleSystemHealth)).Methods("GET")
	adminRoutes.HandleFunc("/status/edge", authMiddleware.RequirePermission(auth.PermSystemRead)(s.handleAdminEdgeStatus)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/system/metrics", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleSystemMetrics)).Methods("GET", "OPTIONS")

	// Admin dashboard (activity, revenue, quick stats)
	adminRoutes.HandleFunc("/dashboard/activity", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleDashboardActivity)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/dashboard/revenue", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleDashboardRevenue)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/dashboard/quick-stats", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleDashboardQuickStats)).Methods("GET", "OPTIONS")

	// Analytics management
	adminRoutes.HandleFunc("/analytics", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleGetAnalyticsSettings)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics", authMiddleware.RequirePermission(auth.PermSystemWrite)(adminHandler.HandleUpdateAnalyticsSettings)).Methods("PATCH", "OPTIONS")
	// Unified analytics (platform and tenant summary)
	adminRoutes.HandleFunc("/analytics/platform/summary", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandlePlatformAnalyticsSummary)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics/tenants/{tenantId}/summary", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleTenantAnalyticsSummary)).Methods("GET", "OPTIONS")

	// Billing management
	adminRoutes.HandleFunc("/billing/summary", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleBillingSummary)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tiers", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListPricingTiers)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tiers", authMiddleware.RequirePermission(auth.PermBillingWrite)(adminHandler.HandleCreatePricingTier)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tiers/{tierId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetPricingTier)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tiers/{tierId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(adminHandler.HandleUpdatePricingTier)).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tiers/{tierId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(adminHandler.HandleDeletePricingTier)).Methods("DELETE", "OPTIONS")

	adminRoutes.HandleFunc("/billing/subscriptions", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListSubscriptions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/subscriptions", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateSubscription))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/subscriptions/{subscriptionId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetSubscription)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/subscriptions/{subscriptionId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateSubscription))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/billing/subscriptions/{subscriptionId}/cancel", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCancelSubscription))).Methods("POST", "OPTIONS")

	adminRoutes.HandleFunc("/billing/invoices", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListInvoices)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/invoices", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateInvoice))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/invoices/{invoiceId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetInvoice)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/invoices/{invoiceId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateInvoice))).Methods("PATCH", "OPTIONS")

	adminRoutes.HandleFunc("/billing/usage", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/usage/record", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRecordUsage))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/state-fabric-add-ons/catalog", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListStateFabricAddonCatalog)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/state-fabric-add-ons/entitlements/{tenantId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListStateFabricTenantEntitlements)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/state-fabric-add-ons/entitlements/{tenantId}/{addonId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpsertStateFabricTenantEntitlement))).Methods("PATCH", "OPTIONS")

	adminRoutes.HandleFunc("/billing/coupons", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListCoupons)).Methods("GET", "OPTIONS")

	// Feedback management (admin only)
	adminRoutes.HandleFunc("/feedback", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.ListFeedback)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.GetFeedbackStats)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/analytics", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.GetFeedbackAnalytics)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/export", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.ExportFeedback)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/{id}/status", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(feedbackHandler.UpdateFeedbackStatus))).Methods("PATCH")
	adminRoutes.HandleFunc("/billing/coupons", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateCoupon))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/coupons/{couponId}/redeem", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRedeemCoupon))).Methods("POST", "OPTIONS")

	// Monitoring management
	adminRoutes.HandleFunc("/monitoring/alerts", authMiddleware.RequirePermission(auth.PermSystemWrite)(monitoringHandler.HandleCreateAlert)).Methods("POST")
	adminRoutes.HandleFunc("/monitoring/metrics", authMiddleware.RequirePermission(auth.PermSystemWrite)(monitoringHandler.HandleRecordMetric)).Methods("POST")
	adminRoutes.HandleFunc("/monitoring/health", authMiddleware.RequirePermission(auth.PermSystemWrite)(monitoringHandler.HandleRecordHealthCheck)).Methods("POST")

	// Security routes (admin only)
	adminRoutes.HandleFunc("/security/metrics", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSecurityMetrics)).Methods("GET")
	adminRoutes.HandleFunc("/security/check-ip", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleCheckIPAccess)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/security/services", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetServiceStatus)).Methods("GET")
	adminRoutes.HandleFunc("/security/certificates", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSSLCertificates)).Methods("GET")
	adminRoutes.HandleFunc("/security/incidents", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetRecentIncidents)).Methods("GET")
	adminRoutes.HandleFunc("/security/compliance", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetComplianceFrameworks)).Methods("GET")
	adminRoutes.HandleFunc("/security/measures", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSecurityMeasures)).Methods("GET")
	adminRoutes.HandleFunc("/security/measures/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(securityHandler.HandleUpdateSecurityMeasureEnabled))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/security/incident-response", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetIncidentResponse)).Methods("GET")
	adminRoutes.HandleFunc("/security/faq", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSecurityFAQ)).Methods("GET")
	adminRoutes.HandleFunc("/security/resources", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSecurityResources)).Methods("GET")
	adminRoutes.HandleFunc("/security/contacts", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetContactInfo)).Methods("GET")

	// MFA management (admin only)
	adminRoutes.HandleFunc("/mfa/force-disable", authMiddleware.RequirePermission(auth.PermUsersWrite)(mfaHandler.AdminForceDisableMFA)).Methods("POST")

	// Admin functions (list/CRUD across all tenants)
	adminRoutes.HandleFunc("/functions", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListAdminFunctions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetAdminFunction)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminHandler.HandleUpdateAdminFunction)).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminHandler.HandleDeleteAdminFunction)).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}/toggle", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminHandler.HandleToggleAdminFunction)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}/deployments", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListAdminFunctionDeployments)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}/logs", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListAdminFunctionLogs)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/functions/{functionId}/metrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetAdminFunctionMetrics)).Methods("GET", "OPTIONS")

	// Admin registry (stats, list, get, update, delete, visibility, pricing, flag, versions, metrics)
	adminRoutes.HandleFunc("/registry/stats", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleGetRegistryStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleListRegistryFunctions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleGetRegistryFunction)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleUpdateRegistryFunction)).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleDeleteRegistryFunction)).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/visibility", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleUpdateRegistryVisibility)).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/pricing", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleUpdateRegistryPricing)).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/flag", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleFlagRegistryFunction)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/versions", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleListRegistryFunctionVersions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/versions/{versionId}/deactivate", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleDeactivateRegistryVersion)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/registry/functions/{functionId}/metrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleGetRegistryFunctionMetrics)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/generate-description", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleGenerateRegistryDescription)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/registry/dre/regenerate-bootstrap", authMiddleware.RequirePermission(auth.PermTenantsWrite)(registryHandler.HandleRegenerateBootstrap)).Methods("POST", "OPTIONS")

	// Admin cache management routes
	adminRoutes.HandleFunc("/cache/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(adminRegistryHandler.HandleGetCacheStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/cache", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeAllCache))).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/cache/{functionId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeFunctionCache))).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/cache/{functionId}/{version}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeVersionCache))).Methods("DELETE", "OPTIONS")

	// Cloudflare analytics
	adminRoutes.HandleFunc("/cloudflare/analytics", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleCloudflareAnalytics)).Methods("GET", "OPTIONS")

	// Admin oversight routes (trust dashboard, execution audit, fraud detection, economic leaderboard)
	adminRoutes.HandleFunc("/oversight/trust-dashboard", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetTrustDashboard)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/execution-audit", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetExecutionAudit)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/fraud-detection", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetFraudDetection)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/economic-leaderboard", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetEconomicLeaderboard)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/block/{type}/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(oversightHandler.HandleBlockEntity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/investigate/{type}/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleInvestigateEntity)).Methods("POST", "OPTIONS")

	// Trust Score management (admin only)
	adminRoutes.HandleFunc("/trust/refresh-all", authMiddleware.RequirePermission(auth.PermSystemWrite)(registryHandler.HandleRefreshAllTrustScores)).Methods("POST", "OPTIONS")

	// Admin factory (same handlers as /v1/factory, for admin dashboard calling /v1/admin/factory/*)
	adminRoutes.HandleFunc("/factory/status", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleStatus)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/reviews/pending", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleListPendingReviews)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/config", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleGetConfig)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/config", authMiddleware.RequirePermission(auth.PermSystemWrite)(factoryHandler.HandleUpdateConfig)).Methods("PUT", "PATCH", "OPTIONS")

	// Admin state fabrics (stats and settings before {id} for route precedence)
	adminRoutes.HandleFunc("/state-fabrics/stats", authMiddleware.RequirePermission(auth.PermTenantsRead)(stateFabricHandler.HandleGetStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/settings", authMiddleware.RequirePermission(auth.PermTenantsRead)(stateFabricHandler.HandleGetSettings)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/settings", authMiddleware.RequirePermission(auth.PermTenantsWrite)(advancedSecurityMiddleware.RequireHMACSignature(stateFabricHandler.HandleUpdateSettings))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(stateFabricHandler.HandleListAll)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/{id}/suspend", authMiddleware.RequirePermission(auth.PermTenantsWrite)(stateFabricHandler.HandleSuspendFabric)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/{id}/resume", authMiddleware.RequirePermission(auth.PermTenantsWrite)(stateFabricHandler.HandleResumeFabric)).Methods("POST", "OPTIONS")

	// Content management (admin only)
	contentRoutes := adminRoutes.PathPrefix("/content").Subrouter()

	// Changelog management
	contentRoutes.HandleFunc("/changelog", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleListChangelogEntries)).Methods("GET")
	contentRoutes.HandleFunc("/changelog", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleCreateChangelogEntry)).Methods("POST")
	contentRoutes.HandleFunc("/changelog/{entryId}", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleGetChangelogEntry)).Methods("GET")
	contentRoutes.HandleFunc("/changelog/{entryId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateChangelogEntry)).Methods("PATCH")
	contentRoutes.HandleFunc("/changelog/{entryId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleDeleteChangelogEntry)).Methods("DELETE")

	// Changelog changes management
	contentRoutes.HandleFunc("/changelog/{entryId}/changes", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleCreateChangelogChange)).Methods("POST")
	contentRoutes.HandleFunc("/changes/{changeId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateChangelogChange)).Methods("PATCH")
	contentRoutes.HandleFunc("/changes/{changeId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleDeleteChangelogChange)).Methods("DELETE")

	// Blog management
	contentRoutes.HandleFunc("/blog", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleListBlogPosts)).Methods("GET")
	contentRoutes.HandleFunc("/blog", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleCreateBlogPost)).Methods("POST")
	contentRoutes.HandleFunc("/blog/{postId}", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleGetBlogPost)).Methods("GET")
	contentRoutes.HandleFunc("/blog/{postId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateBlogPost)).Methods("PATCH")
	contentRoutes.HandleFunc("/blog/{postId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleDeleteBlogPost)).Methods("DELETE")

	// Blog categories (admin CRUD)
	contentRoutes.HandleFunc("/categories", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleListAdminCategories)).Methods("GET")
	contentRoutes.HandleFunc("/categories", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleCreateAdminCategory)).Methods("POST")
	contentRoutes.HandleFunc("/categories/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleGetAdminCategory)).Methods("GET")
	contentRoutes.HandleFunc("/categories/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateAdminCategory)).Methods("PATCH")
	contentRoutes.HandleFunc("/categories/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleDeleteAdminCategory)).Methods("DELETE")

	// Blog authors (admin CRUD)
	contentRoutes.HandleFunc("/authors", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleListAdminAuthors)).Methods("GET")
	contentRoutes.HandleFunc("/authors", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleCreateAdminAuthor)).Methods("POST")
	contentRoutes.HandleFunc("/authors/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleGetAdminAuthor)).Methods("GET")
	contentRoutes.HandleFunc("/authors/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateAdminAuthor)).Methods("PATCH")
	contentRoutes.HandleFunc("/authors/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleDeleteAdminAuthor)).Methods("DELETE")

	// Content sync
	contentRoutes.HandleFunc("/sync/github-releases", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleSyncGitHubReleases)).Methods("POST")

	// Content generation (Open Router AI) — also register on adminRoutes so path is unambiguous
	adminRoutes.HandleFunc("/content/generate/changelog", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleGenerateChangelogContent)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/content/generate/blog", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleGenerateBlogContent)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/content/generate/author", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleGenerateAuthorContent)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/content/generate/category", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleGenerateCategoryContent)).Methods("POST", "OPTIONS")

	// Tenant-scoped operations (admin impersonating tenant)
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantApps)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetTenantApp)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/backends", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantBackends)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/deployments", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantDeployments)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/deployments/{deploymentId}/rollback", authMiddleware.RequirePermission(auth.PermDeploymentsWrite)(adminHandler.HandleTenantDeploymentRollback)).Methods("POST")

	// Tenant-scoped observability
	adminRoutes.HandleFunc("/tenants/{tenantId}/metrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleTenantMetrics)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/tenants/{tenantId}/health", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleTenantHealth)).Methods("GET")

	// Seat management (account sharing / user limit enforcement)
	adminRoutes.HandleFunc("/tenants/{tenantId}/seat-usage", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetSeatUsage)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/users/{userId}/deactivate", authMiddleware.RequirePermission(auth.PermUsersWrite)(adminHandler.HandleDeactivateUser)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/users/{userId}/reactivate", authMiddleware.RequirePermission(auth.PermUsersWrite)(adminHandler.HandleReactivateUser)).Methods("POST", "OPTIONS")
}
