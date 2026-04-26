package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	"github.com/functionfly/functionfly/internal/api/handlers/billing"
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
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
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
	newsletterHandler *admin.NewsletterHandler,
	usageHandler *billing.UsageHandler,
	costAllocationHandler *billing.CostAllocationHandler,
	retentionHandler *admin.RetentionHandler,
	disputesHandler *admin.DisputesHandler,
	stateUsageHandler *billing.StateUsageHandler,
	unfairAdvantageHandler *agenthandler.UnfairAdvantageHandler,
) {
	adminRoutes := api.PathPrefix("/admin").Subrouter()

	// ── Admin middleware wiring ──────────────────────────────────────────────────
	// Order: CORS Preflight → IP Allowlist → Session Validation → Security Alert (rate limit) → CSRF
	//
	// 0. CORS preflight: must be first so OPTIONS requests get proper headers without auth
	adminRoutes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				origin := r.Header.Get("Origin")
				if origin != "" && middleware.IsOriginAllowedForRequest(r) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint, x-device-fingerprint, X-Environment")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

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
	adminRoutes.HandleFunc("/auth/last-login", authMiddleware.RequireAuth(adminHandler.HandleGetAdminLastLogin)).Methods("GET", "OPTIONS")

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

	// Signup invite codes (invite-only launch)
	adminRoutes.HandleFunc("/signup-invites", authMiddleware.RequirePermission(auth.PermUsersRead)(adminHandler.HandleListSignupInvites)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/signup-invites", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateSignupInvite))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/signup-invites/{id}/revoke", authMiddleware.RequirePermission(auth.PermUsersWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRevokeSignupInvite))).Methods("POST", "OPTIONS")

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
	// Client-side tracked security events (login attempts, suspicious activity, etc.)
	adminRoutes.HandleFunc("/security/events", authMiddleware.RequirePermission(auth.PermSystemWrite)(securityEventHandler.HandleCreateSecurityEvents)).Methods("POST", "OPTIONS")

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

	// Revenue analytics - MRR/ARR metrics
	adminRoutes.HandleFunc("/analytics/mrr", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsMRR)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics/mrr-series", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsMRRSeries)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics/arr", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsARR)).Methods("GET", "OPTIONS")

	// Churn metrics
	adminRoutes.HandleFunc("/analytics/churn", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsChurn)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics/churn-series", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsChurnSeries)).Methods("GET", "OPTIONS")

	// LTV metrics
	adminRoutes.HandleFunc("/analytics/ltv", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleAnalyticsLTV)).Methods("GET", "OPTIONS")

	// Financial reporting
	adminRoutes.HandleFunc("/analytics/financial-report", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleFinancialReport)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics/tax-jurisdiction", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleTaxJurisdictionReport)).Methods("GET", "OPTIONS")

	// Billing management
	adminRoutes.HandleFunc("/billing/summary", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleBillingSummary)).Methods("GET", "OPTIONS")

	// Wallet admin management (freeze, suspend, adjustments, reconciliation)
	adminRoutes.HandleFunc("/wallets", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListWallets)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/{walletId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetWallet)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/{walletId}/freeze", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleFreezeWallet))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/{walletId}/unfreeze", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUnfreezeWallet))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/{walletId}/close", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCloseWallet))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/{walletId}/adjust", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleAdjustWalletBalance))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/reconciliation", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleTriggerReconciliation))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/wallets/reconciliation/runs", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetReconciliationRuns)).Methods("GET", "OPTIONS")

	// Payout approval management
	adminRoutes.HandleFunc("/payouts/pending", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListPendingPayouts)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/payouts/{payoutId}/approve", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleApprovePayout))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/payouts/{payoutId}/reject", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRejectPayout))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/payouts/approval-rules", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListPayoutApprovalRules)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/payouts/approval-rules", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreatePayoutApprovalRule))).Methods("POST", "OPTIONS")

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

	// Dispute and refund management
	adminRoutes.HandleFunc("/billing/disputes", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleListDisputes)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/disputes/open", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetOpenDisputes)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/disputes/stats", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetDisputeStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/disputes/{disputeId}", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetDispute)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/disputes/{disputeId}/status", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(disputesHandler.HandleUpdateDisputeStatus))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/billing/disputes/{disputeId}/evidence", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(disputesHandler.HandleUpdateDisputeEvidence))).Methods("POST", "OPTIONS")

	// Refund management
	adminRoutes.HandleFunc("/billing/refunds", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleListRefunds)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/refunds/stats", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetRefundStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/refunds/{refundId}", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetRefund)).Methods("GET", "OPTIONS")

	// Credit Notes (for refund accounting / SOX compliance)
	adminRoutes.HandleFunc("/billing/credit-notes", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListCreditNotes)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateCreditNote))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes/stats", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetCreditNoteStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes/{creditNoteId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetCreditNote)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes/{creditNoteId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateCreditNote))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes/{creditNoteId}/void", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleVoidCreditNote))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/credit-notes/{creditNoteId}/apply", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleApplyCreditNote))).Methods("POST", "OPTIONS")

	// Chargeback reconciliation
	adminRoutes.HandleFunc("/billing/chargebacks/reconciliation", authMiddleware.RequirePermission(auth.PermBillingRead)(disputesHandler.HandleGetChargebackReconciliation)).Methods("GET", "OPTIONS")

	// Billing operational readiness - webhook replay and monitoring
	adminRoutes.HandleFunc("/billing/webhooks", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListStoredWebhooks)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/webhooks/{webhookId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetStoredWebhook)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/webhooks/{webhookId}/replay", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleReplayWebhook))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/webhooks/replay-requests", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListWebhookReplayRequests)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/webhooks/cleanup", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCleanupExpiredWebhooks))).Methods("POST", "OPTIONS")

	// Tax exemption certificate management
	adminRoutes.HandleFunc("/billing/tax-exemptions/pending", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListPendingTaxCertificates)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/tax-exemptions/{certificateId}/review", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleReviewTaxCertificate))).Methods("POST", "OPTIONS")

	// Feedback management (admin only)
	adminRoutes.HandleFunc("/feedback", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.ListFeedback)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.GetFeedbackStats)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/analytics", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.GetFeedbackAnalytics)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/export", authMiddleware.RequirePermission(auth.PermSystemRead)(feedbackHandler.ExportFeedback)).Methods("GET")
	adminRoutes.HandleFunc("/feedback/{id}/status", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(feedbackHandler.UpdateFeedbackStatus))).Methods("PATCH")
	adminRoutes.HandleFunc("/billing/coupons", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateCoupon))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/coupons/{couponId}/redeem", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRedeemCoupon))).Methods("POST", "OPTIONS")

	// Affiliate / Referral Commission System
	adminRoutes.HandleFunc("/billing/affiliate-codes", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListAffiliateCodes)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-codes", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateAffiliateCode))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-codes/{codeId}", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleGetAffiliateCode)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-codes/{codeId}", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateAffiliateCode))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-codes/{codeId}/referrals", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListAffiliateReferrals)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-codes/{codeId}/commissions", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleListAffiliateCommissions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-referrals", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleRecordAffiliateReferral))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-referrals/{referralId}/status", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateAffiliateReferralStatus))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-commissions/{commissionId}/approve", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleApproveAffiliateCommission))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-commissions/{commissionId}/paid", authMiddleware.RequirePermission(auth.PermBillingWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleMarkAffiliateCommissionPaid))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/billing/affiliate-commissions/calculate", authMiddleware.RequirePermission(auth.PermBillingRead)(adminHandler.HandleCalculateAffiliateCommission)).Methods("POST", "OPTIONS")

	// Admin usage management (real-time usage tracking)
	adminRoutes.HandleFunc("/usage/tenants/{tenantId}", authMiddleware.RequirePermission(auth.PermBillingRead)(usageHandler.AdminGetTenantUsage)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/usage/metrics", authMiddleware.RequirePermission(auth.PermSystemRead)(usageHandler.GetUsageMetrics)).Methods("GET", "OPTIONS")

	// Admin state usage management (state fabric billing/quota)
	adminRoutes.HandleFunc("/usage/state/{tenant_id}", authMiddleware.RequirePermission(auth.PermBillingRead)(stateUsageHandler.AdminGetTenantStateUsage)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/usage/state", authMiddleware.RequirePermission(auth.PermSystemRead)(stateUsageHandler.AdminListAllStateUsage)).Methods("GET", "OPTIONS")

	// Admin cost allocation (detailed cost tracking and chargebacks)
	adminRoutes.HandleFunc("/costs/tenants/{tenant_id}/summary", authMiddleware.RequirePermission(auth.PermBillingRead)(costAllocationHandler.AdminGetTenantCostSummary)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/costs/chargeback", authMiddleware.RequirePermission(auth.PermBillingRead)(costAllocationHandler.GetChargebackReport)).Methods("GET", "OPTIONS")

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
	adminRoutes.HandleFunc("/registry/functions/{functionId}/versions", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminRegistryHandler.HandleListRegistryFunctionVersions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/registry/generate-description", authMiddleware.RequirePermission(auth.PermTenantsWrite)(adminRegistryHandler.HandleGenerateRegistryDescription)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/registry/dre/regenerate-bootstrap", authMiddleware.RequirePermission(auth.PermTenantsWrite)(registryHandler.HandleRegenerateBootstrap)).Methods("POST", "OPTIONS")

	// Admin cache management routes
	adminRoutes.HandleFunc("/cache/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(adminRegistryHandler.HandleGetCacheStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/cache", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeAllCache))).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/cache/{functionId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeFunctionCache))).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/cache/{functionId}/{version}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminRegistryHandler.HandlePurgeVersionCache))).Methods("DELETE", "OPTIONS")

	// Admin retention management routes (execution log data retention policies)
	adminRoutes.HandleFunc("/retention/settings", authMiddleware.RequirePermission(auth.PermSystemRead)(retentionHandler.HandleGetRetentionSettings)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/retention/settings", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(retentionHandler.HandleUpdateRetentionSettings))).Methods("PUT", "OPTIONS")
	adminRoutes.HandleFunc("/retention/settings/reset", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(retentionHandler.HandleResetRetentionDefaults))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/retention/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(retentionHandler.HandleGetRetentionStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/retention/metrics", authMiddleware.RequirePermission(auth.PermSystemRead)(retentionHandler.HandleGetCleanupMetrics)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/retention/cleanup", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(retentionHandler.HandleRunManualCleanup))).Methods("POST", "OPTIONS")

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

	// Sliding window trust score management
	adminRoutes.HandleFunc("/trust/calculate-sliding", authMiddleware.RequirePermission(auth.PermSystemWrite)(registryHandler.HandleCalculateSlidingWindow)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/trust/sliding-window/{functionId}", authMiddleware.RequirePermission(auth.PermSystemRead)(registryHandler.HandleGetSlidingWindowState)).Methods("GET", "OPTIONS")

	// Admin factory (same handlers as /v1/factory, for admin dashboard calling /v1/admin/factory/*)
	adminRoutes.HandleFunc("/factory/status", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleStatus)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/reviews/pending", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleListPendingReviews)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/pipeline/run", authMiddleware.RequirePermission(auth.PermSystemWrite)(factoryHandler.HandleRunPipeline)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/factory/config", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleGetConfig)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/config", authMiddleware.RequirePermission(auth.PermSystemWrite)(factoryHandler.HandleUpdateConfig)).Methods("PUT", "PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/factory/opportunities", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleListOpportunities)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/opportunities/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleGetOpportunity)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/factory/opportunities/{id}/approve", authMiddleware.RequirePermission(auth.PermSystemWrite)(factoryHandler.HandleApproveOpportunity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/factory/opportunities/{id}/reject", authMiddleware.RequirePermission(auth.PermSystemWrite)(factoryHandler.HandleRejectOpportunity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/factory/functions", authMiddleware.RequirePermission(auth.PermSystemRead)(factoryHandler.HandleListFunctions)).Methods("GET", "OPTIONS")

	// ── Unfair Advantage Engine (ADMIN ONLY - FunctionFly Internal) ─────────────────────────────
	adminRoutes.HandleFunc("/unfair-advantage/dashboard", authMiddleware.RequirePermission(auth.PermSystemRead)(unfairAdvantageHandler.HandleGetDashboard)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/value-report", authMiddleware.RequirePermission(auth.PermBillingRead)(unfairAdvantageHandler.HandleGetValueReport)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/opportunities", authMiddleware.RequirePermission(auth.PermSystemRead)(unfairAdvantageHandler.HandleListInternalOpportunities)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/opportunities/seed", authMiddleware.RequirePermission(auth.PermSystemWrite)(unfairAdvantageHandler.HandleSeedOpportunity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/opportunities/custom", authMiddleware.RequirePermission(auth.PermSystemWrite)(unfairAdvantageHandler.HandleSeedCustomOpportunity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/rdlab/run", authMiddleware.RequirePermission(auth.PermSystemWrite)(unfairAdvantageHandler.HandleRunRDLab)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/functions/generate", authMiddleware.RequirePermission(auth.PermSystemWrite)(unfairAdvantageHandler.HandleGenerateInternalFunction)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/functions", authMiddleware.RequirePermission(auth.PermSystemRead)(unfairAdvantageHandler.HandleListInternalFunctions)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/unfair-advantage/stealth/run", authMiddleware.RequirePermission(auth.PermSystemWrite)(unfairAdvantageHandler.HandleRunStealthPipeline)).Methods("POST", "OPTIONS")

	// Admin state fabrics (stats and settings before {id} for route precedence)
	adminRoutes.HandleFunc("/state-fabrics/stats", authMiddleware.RequirePermission(auth.PermTenantsRead)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleGetStats))).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/settings", authMiddleware.RequirePermission(auth.PermTenantsRead)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleGetSettings))).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/settings", authMiddleware.RequirePermission(auth.PermTenantsWrite)(advancedSecurityMiddleware.RequireHMACSignature(stateFabricHandler.HandleUpdateSettings))).Methods("PATCH", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleListAll))).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/{id}/suspend", authMiddleware.RequirePermission(auth.PermTenantsWrite)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleSuspendFabric))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/{id}/resume", authMiddleware.RequirePermission(auth.PermTenantsWrite)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleResumeFabric))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/cleanup", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(stateFabricHandler.HandleRunTTLCleanup))).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/state-fabrics/cleanup/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(rateLimiter.RequireRateLimit(stateFabricHandler.HandleGetTTLCleanupStats))).Methods("GET", "OPTIONS")

	// Trigger Engine Admin Endpoints
	// GET /admin/triggers/stats - Get trigger engine statistics
	adminRoutes.HandleFunc("/triggers/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.triggerEngine == nil {
			http.Error(w, `{"error":"trigger engine not initialized"}`, http.StatusServiceUnavailable)
			return
		}
		stats := s.triggerEngine.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))).Methods("GET", "OPTIONS")

	// GET /admin/triggers/queue-stats - Get detailed queue statistics
	adminRoutes.HandleFunc("/triggers/queue-stats", authMiddleware.RequirePermission(auth.PermSystemRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.triggerEngine == nil {
			http.Error(w, `{"error":"trigger engine not initialized"}`, http.StatusServiceUnavailable)
			return
		}
		stats, err := s.triggerEngine.GetQueueStats(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))).Methods("GET", "OPTIONS")

	// GET /admin/triggers/dead-letter - List dead letter queue entries
	adminRoutes.HandleFunc("/triggers/dead-letter", authMiddleware.RequirePermission(auth.PermSystemRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.triggerEngine == nil {
			http.Error(w, `{"error":"trigger engine not initialized"}`, http.StatusServiceUnavailable)
			return
		}
		// Parse pagination params
		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 1000 {
				limit = n
			}
		}
		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if n, err := strconv.Atoi(o); err == nil && n >= 0 {
				offset = n
			}
		}

		var entries []staterepo.TriggerDeadLetter
		var total int64
		db := s.postgresDB.GORM.WithContext(r.Context()).Model(&staterepo.TriggerDeadLetter{})
		if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
			db = db.Where("tenant_id = ?", tenantID)
		}
		if canRetry := r.URL.Query().Get("can_retry"); canRetry != "" {
			db = db.Where("can_retry = ?", canRetry == "true")
		}
		db.Count(&total)
		db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&entries)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"entries": entries,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	}))).Methods("GET", "OPTIONS")

	// POST /admin/triggers/dead-letter/{id}/retry - Retry a dead letter entry
	adminRoutes.HandleFunc("/triggers/dead-letter/{id}/retry", authMiddleware.RequirePermission(auth.PermSystemWrite)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.triggerEngine == nil {
			http.Error(w, `{"error":"trigger engine not initialized"}`, http.StatusServiceUnavailable)
			return
		}
		vars := mux.Vars(r)
		dlqID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, `{"error":"invalid dead letter id"}`, http.StatusBadRequest)
			return
		}
		if err := s.triggerEngine.RetryDeadLetterEvent(r.Context(), dlqID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "retry queued"})
	}))).Methods("POST", "OPTIONS")

	// POST /admin/triggers/purge-completed - Purge old completed events
	adminRoutes.HandleFunc("/triggers/purge-completed", authMiddleware.RequirePermission(auth.PermSystemWrite)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.triggerEngine == nil {
			http.Error(w, `{"error":"trigger engine not initialized"}`, http.StatusServiceUnavailable)
			return
		}
		// Default to 30 days retention
		retentionDays := 30
		if d := r.URL.Query().Get("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				retentionDays = n
			}
		}
		deleted, err := s.triggerEngine.PurgeCompletedEvents(r.Context(), time.Duration(retentionDays)*24*time.Hour)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted":        deleted,
			"retention_days": retentionDays,
		})
	}))).Methods("POST", "OPTIONS")

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

	// Blog settings
	contentRoutes.HandleFunc("/blog/settings", authMiddleware.RequirePermission(auth.PermSystemRead)(contentHandler.HandleGetBlogSettings)).Methods("GET")
	contentRoutes.HandleFunc("/blog/settings", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleUpdateBlogSettings)).Methods("PATCH")

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

	// Newsletter management
	adminRoutes.HandleFunc("/newsletter/subscribers", authMiddleware.RequirePermission(auth.PermSystemRead)(newsletterHandler.ListSubscribers)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/subscribers", authMiddleware.RequirePermission(auth.PermSystemWrite)(newsletterHandler.CreateSubscriber)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/subscribers/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(newsletterHandler.DeleteSubscriber)).Methods("DELETE", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/stats", authMiddleware.RequirePermission(auth.PermSystemRead)(newsletterHandler.GetStats)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/campaigns", authMiddleware.RequirePermission(auth.PermSystemRead)(newsletterHandler.ListCampaigns)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/campaigns", authMiddleware.RequirePermission(auth.PermSystemWrite)(newsletterHandler.CreateCampaign)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/campaigns/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(newsletterHandler.GetCampaign)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/newsletter/campaigns/{id}/send", authMiddleware.RequirePermission(auth.PermSystemWrite)(newsletterHandler.SendCampaign)).Methods("POST", "OPTIONS")
}
