package billing

import (
	"github.com/functionfly/functionfly/internal/api/routes"
	"github.com/gorilla/mux"
)

// RouteRegistrar implements the routes.RouteRegistrar interface for billing routes
type RouteRegistrar struct {
	handler         *Handler
	usageHandler    *UsageHandler
	forecastHandler *UsageForecastHandler
	costHandler     *CostAllocationHandler
	exportHandler   *ExportHandler
	externalHandler *ExternalBillingHandler
}

// NewRouteRegistrar creates a new billing route registrar
// Note: Authentication and rate limiting are applied via middleware chain
func NewRouteRegistrar(
	handler *Handler,
	usageHandler *UsageHandler,
	forecastHandler *UsageForecastHandler,
	costHandler *CostAllocationHandler,
	exportHandler *ExportHandler,
	externalHandler *ExternalBillingHandler,
) *RouteRegistrar {
	return &RouteRegistrar{
		handler:         handler,
		usageHandler:    usageHandler,
		forecastHandler: forecastHandler,
		costHandler:     costHandler,
		exportHandler:   exportHandler,
		externalHandler: externalHandler,
	}
}

// Name returns the registrar name
func (r *RouteRegistrar) Name() string {
	return "billing"
}

// Priority returns the registration priority
func (r *RouteRegistrar) Priority() int {
	return routes.PriorityCore
}

// Register implements the RouteRegistrar interface
// The auth and rate limiting middleware should be applied via the MiddlewareChain
func (r *RouteRegistrar) Register(router *mux.Router, api *mux.Router, protected *mux.Router, mw *routes.MiddlewareChain) {
	// Billing routes
	billingRouter := api.PathPrefix("/billing").Subrouter()

	// Portal and checkout
	billingRouter.HandleFunc("/portal-session", r.handler.HandleCreatePortalSession).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/checkout", r.handler.HandleCreateCheckoutSession).Methods("POST", "OPTIONS")

	// Subscription management
	billingRouter.HandleFunc("/subscription", r.handler.HandleGetSubscription).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/subscription/cancel", r.handler.HandleCancelSubscription).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/subscription/webhook", r.handler.HandleSubscriptionWebhook).Methods("POST", "OPTIONS")

	// Invoices and usage
	billingRouter.HandleFunc("/invoices", r.handler.HandleListInvoices).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/usage", r.handler.HandleGetUsage).Methods("GET", "OPTIONS")

	// Bundles and founder mode
	billingRouter.HandleFunc("/bundles", r.handler.HandleGetBundles).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/bundles/{slug}", r.handler.HandleGetBundle).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/bundles/founder", r.handler.HandleRegisterFounderMode).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/bundles/founder-status", r.handler.HandleGetFounderModeStatus).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/bundles/deferred-status", r.handler.HandleGetDeferredBillingStatus).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/bundles/checkout", r.handler.HandleCreateBundleCheckout).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/bundles/subscription", r.handler.HandleGetBundleSubscription).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/bundles/convert", r.handler.HandleConvertToPaid).Methods("POST", "OPTIONS")

	// Revenue/plans
	billingRouter.HandleFunc("/plans", r.handler.HandleGetPlans).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/plans/verification-cost", r.handler.HandleGetVerificationCost).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/plans/verify-function", r.handler.HandleVerifyFunction).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/earnings", r.handler.HandleGetEarnings).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/agent-usage", r.handler.HandleGetAgentUsage).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/subscribe", r.handler.HandleSubscribe).Methods("POST", "OPTIONS")

	// Stripe Meter Events integration for metered billing
	billingRouter.HandleFunc("/meter/verify", r.handler.HandleVerifyStripeMeterIntegration).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/meter/report", r.handler.HandleReportUsage).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/meter/status", r.handler.HandleGetMeteredBillingStatus).Methods("GET", "OPTIONS")

	// Tax/VAT compliance endpoints
	billingRouter.HandleFunc("/tax/settings", r.handler.HandleGetTaxSettings).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/tax/settings", r.handler.HandleUpdateTaxSettings).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/tax/types", r.handler.HandleGetTaxTypes).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/tax/calculate", r.handler.HandleCalculateTax).Methods("POST", "OPTIONS")
	billingRouter.HandleFunc("/tax/validate", r.handler.HandleValidateTaxID).Methods("POST", "OPTIONS")

	// Wallet
	billingRouter.HandleFunc("/wallet", r.handler.HandleGetWallet).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/wallet/transactions", r.handler.HandleListWalletTransactions).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/wallet/fees", r.handler.HandleListPlatformFees).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/wallet/top-up", r.handler.HandleWalletTopUp).Methods("POST", "OPTIONS")

	// State Fabric addons
	billingRouter.HandleFunc("/state-fabric-addons", r.handler.HandleListStateFabricAddOnCatalog).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/state-fabric-addons/entitlements", r.handler.HandleGetStateFabricAddOnEntitlements).Methods("GET", "OPTIONS")
	billingRouter.HandleFunc("/state-fabric-addons/checkout", r.handler.HandleCreateStateFabricAddOnCheckout).Methods("POST", "OPTIONS")

	// Usage and cost allocation (if handlers are provided)
	if r.usageHandler != nil {
		billingRouter.HandleFunc("/usage/realtime", r.usageHandler.GetRealtimeUsage).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/usage/quota", r.usageHandler.CheckQuota).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/usage/history", r.usageHandler.GetUsageHistory).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/usage/by-function", r.usageHandler.GetUsageByFunction).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/usage/period", r.usageHandler.GetCurrentPeriodUsage).Methods("GET", "OPTIONS")
	}

	if r.costHandler != nil {
		billingRouter.HandleFunc("/costs", r.costHandler.GetCostSummary).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/costs/by-function", r.costHandler.GetCostByFunction).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/costs/by-period", r.costHandler.GetCostByPeriod).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/costs/by-region", r.costHandler.GetCostByRegion).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/costs/entries", r.costHandler.GetCostEntries).Methods("GET", "OPTIONS")
	}

	// Forecasting and alerts
	if r.forecastHandler != nil {
		billingRouter.HandleFunc("/forecast", r.forecastHandler.GetCurrentForecast).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/forecast/{type}", r.forecastHandler.GetForecastByType).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/forecast/refresh", r.forecastHandler.RefreshForecast).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/alerts", r.forecastHandler.ListAlerts).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/alerts", r.forecastHandler.CreateAlert).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/alerts/{id}", r.forecastHandler.GetAlert).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/alerts/{id}", r.forecastHandler.UpdateAlert).Methods("PUT", "OPTIONS")
		billingRouter.HandleFunc("/alerts/{id}", r.forecastHandler.DeleteAlert).Methods("DELETE", "OPTIONS")
		billingRouter.HandleFunc("/alerts/history", r.forecastHandler.GetAlertHistory).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/spend-cap", r.forecastHandler.GetSpendCap).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/spend-cap", r.forecastHandler.UpdateSpendCap).Methods("PUT", "OPTIONS")
		billingRouter.HandleFunc("/trends", r.forecastHandler.GetUsageTrends).Methods("GET", "OPTIONS")
	}

	// Export endpoints
	if r.exportHandler != nil {
		billingRouter.HandleFunc("/exports/configs", r.exportHandler.ListExportConfigurations).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/configs", r.exportHandler.CreateExportConfiguration).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/exports/configs/{id}", r.exportHandler.GetExportConfiguration).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/configs/{id}", r.exportHandler.UpdateExportConfiguration).Methods("PUT", "OPTIONS")
		billingRouter.HandleFunc("/exports/configs/{id}", r.exportHandler.DeleteExportConfiguration).Methods("DELETE", "OPTIONS")
		billingRouter.HandleFunc("/exports/jobs", r.exportHandler.ListExportJobs).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/jobs/{id}", r.exportHandler.GetExportJob).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/jobs/{id}/download", r.exportHandler.DownloadExport).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/execute", r.exportHandler.ExecuteExport).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/exports/templates", r.exportHandler.ListExportTemplates).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/exports/templates/{id}", r.exportHandler.GetExportTemplate).Methods("GET", "OPTIONS")
	}

	// External billing integrations
	if r.externalHandler != nil {
		billingRouter.HandleFunc("/external-systems", r.externalHandler.ListExternalBillingSystems).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/external-systems", r.externalHandler.CreateExternalBillingSystem).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/{id}", r.externalHandler.GetExternalBillingSystem).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/{id}", r.externalHandler.UpdateExternalBillingSystem).Methods("PUT", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/{id}", r.externalHandler.DeleteExternalBillingSystem).Methods("DELETE", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/{id}/test", r.externalHandler.TestExternalBillingSystem).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/syncs", r.externalHandler.ListBillingSyncs).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/syncs/{id}", r.externalHandler.GetBillingSync).Methods("GET", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/syncs/{id}/trigger", r.externalHandler.TriggerBillingSync).Methods("POST", "OPTIONS")
		billingRouter.HandleFunc("/external-systems/types", r.externalHandler.GetBillingSystemTypes).Methods("GET", "OPTIONS")
	}
}
