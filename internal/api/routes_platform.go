package api

import (
	"github.com/functionfly/functionfly/internal/analytics/unified"
	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	agentmemoryhandler "github.com/functionfly/functionfly/internal/api/handlers/agent_memory"
	analyticshandler "github.com/functionfly/functionfly/internal/api/handlers/analytics"
	"github.com/functionfly/functionfly/internal/api/handlers/apps"
	"github.com/functionfly/functionfly/internal/api/handlers/backends"
	billinghandler "github.com/functionfly/functionfly/internal/api/handlers/billing"
	categorizationhandler "github.com/functionfly/functionfly/internal/api/handlers/categorization"
	"github.com/functionfly/functionfly/internal/api/handlers/dashboard"
	"github.com/functionfly/functionfly/internal/api/handlers/decisions"
	"github.com/functionfly/functionfly/internal/api/handlers/deploykeys"
	"github.com/functionfly/functionfly/internal/api/handlers/deployments"
	"github.com/functionfly/functionfly/internal/api/handlers/enterprise"
	factoryhandler "github.com/functionfly/functionfly/internal/api/handlers/factory"
	"github.com/functionfly/functionfly/internal/api/handlers/function_webhooks"
	"github.com/functionfly/functionfly/internal/api/handlers/functions"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	"github.com/functionfly/functionfly/internal/api/handlers/providers"
	"github.com/functionfly/functionfly/internal/api/handlers/security"
	"github.com/functionfly/functionfly/internal/api/handlers/state"
	"github.com/functionfly/functionfly/internal/api/handlers/statefabric"
	statushandler "github.com/functionfly/functionfly/internal/api/handlers/status"
	"github.com/functionfly/functionfly/internal/api/handlers/support"
	teammemoryhandler "github.com/functionfly/functionfly/internal/api/handlers/team_memory"
	"github.com/functionfly/functionfly/internal/api/handlers/teams"
	"github.com/functionfly/functionfly/internal/api/handlers/vault"
	versionhandler "github.com/functionfly/functionfly/internal/api/handlers/version"
	"github.com/functionfly/functionfly/internal/api/middleware"
	advancedsecurity "github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	"github.com/gorilla/mux"
)

// registerPlatformRoutes wires monitoring, status, factory, analytics, state,
// vault, dashboard, teams, apps, functions, deployment, and support endpoints.
func registerPlatformRoutes(
	s *Server,
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	advancedSecurityMiddleware *advancedsecurity.AdvancedSecurityMiddleware,
	vaultRateLimiter *middleware.VaultRateLimiter,
	providerRateLimiter *middleware.ProviderRateLimiter,
	monitoringHandler *monitoring.Handler,
	securityHandler *security.Handler,
	statusHandler *statushandler.Handler,
	factoryHandler *factoryhandler.Handler,
	experimentHandler *factoryhandler.ExperimentHandler,
	categorizationHandler *categorizationhandler.Handler,
	analyticsHandler *analyticshandler.Handler,
	unifiedAnalyticsSvc *unified.Service,
	stateHandler *state.Handler,
	stateFabricHandler *statefabric.Handler,
	vaultHandler *vault.Handler,
	memoryHandler *state.AgentMemoryHandler,
	agentMemoryHandler *agentmemoryhandler.AgentMemoryHandler,
	dashboardHandler *dashboard.Handler,
	enterpriseSLAHandler *enterprise.SLAHandler,
	teamHandler *teams.Handler,
	providersHandler *providers.Handler,
	appsHandler *apps.Handler,
	functionsHandler *functions.Handler,
	backendsHandler *backends.Handler,
	deploymentsHandler *deployments.Handler,
	versionHandler *versionhandler.Handler,
	maintenanceHandler *admin.MaintenanceHandler,
	supportHdlr *support.Handler,
	supportAdminHdlr *support.AdminHandler,
	supportWSHub *support.WebSocketHub,
	decisionsHandler *decisions.Handler,
	stateUsageHandler *billinghandler.StateUsageHandler,
	deployKeysHandler *deploykeys.Handler,
	functionWebhooksHandler *function_webhooks.Handler,
	swarmControllerHandler agenthandler.SwarmControllerHandlerInterface,
	unfairAdvantageHandler *agenthandler.UnfairAdvantageHandler,
) {
	// ── Metrics (public) ─────────────────────────────────────────────────────
	api.HandleFunc("/metrics/global", s.handleGlobalMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/stream", s.handleMetricsStream).Methods("GET", "OPTIONS")

	// ── Monitoring (public read) ──────────────────────────────────────────────
	api.HandleFunc("/monitoring/metrics", monitoringHandler.HandleGetMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/alerts", monitoringHandler.HandleGetAlerts).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/health", monitoringHandler.HandleGetSystemHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/events", monitoringHandler.HandleGetMonitoringEvents).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/realtime", monitoringHandler.HandleRealtimeConnection).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/realtime/stats", monitoringHandler.HandleGetRealtimeStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/local-runtime", monitoringHandler.HandleGetLocalRuntimeMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/health", monitoringHandler.HandleGetDatabaseHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/metrics", monitoringHandler.HandleGetDatabaseMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/alerts", monitoringHandler.HandleGetDatabaseAlerts).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/check", monitoringHandler.HandleCheckDatabaseHealth).Methods("POST", "OPTIONS")
	api.HandleFunc("/monitoring/database/changes", monitoringHandler.HandleSubscribeToDatabaseChanges).Methods("GET", "OPTIONS")

	// Monitoring management (protected)
	protected.HandleFunc("/monitoring/alerts/{alertId}/resolve", authMiddleware.RequireAuth(monitoringHandler.HandleResolveAlert)).Methods("POST")
	protected.HandleFunc("/monitoring/dashboard", authMiddleware.RequireAuth(monitoringHandler.HandleGetDashboardConfig)).Methods("GET")

	// ── Security Metrics (public) ─────────────────────────────────────────────
	api.HandleFunc("/metrics/security", securityHandler.HandleGetSecurityMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/services", securityHandler.HandleGetServiceStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/certificates", securityHandler.HandleGetSSLCertificates).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/incidents", securityHandler.HandleGetRecentIncidents).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/compliance", securityHandler.HandleGetComplianceFrameworks).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/measures", securityHandler.HandleGetSecurityMeasures).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/incident-response", securityHandler.HandleGetIncidentResponse).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/faq", securityHandler.HandleGetSecurityFAQ).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/resources", securityHandler.HandleGetSecurityResources).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/security/contacts", securityHandler.HandleGetContactInfo).Methods("GET", "OPTIONS")

	// ── Status Page (public read, admin write) ────────────────────────────────
	api.HandleFunc("/status", statusHandler.HandleGetPlatformStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/status/edge", s.handleEdgeStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/status/components", statusHandler.HandleGetComponents).Methods("GET", "OPTIONS")
	api.HandleFunc("/status/providers", statusHandler.HandleGetProviders).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents", statusHandler.HandleListIncidents).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents/{id}", statusHandler.HandleGetIncident).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents", authMiddleware.RequireAuth(statusHandler.HandleCreateIncident)).Methods("POST", "OPTIONS")
	api.HandleFunc("/incidents/{id}", authMiddleware.RequireAuth(statusHandler.HandleUpdateIncident)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/metrics/uptime", statusHandler.HandleGetUptimeMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/latency", statusHandler.HandleGetLatencyMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/maintenance", statusHandler.HandleListMaintenance).Methods("GET", "OPTIONS")
	api.HandleFunc("/maintenance", authMiddleware.RequireAuth(statusHandler.HandleCreateMaintenance)).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/maintenance/status", maintenanceHandler.HandleGetPublicStatus).Methods("GET", "OPTIONS")

	// ── Status RSS Feed (public) ──────────────────────────────────────────────
	api.HandleFunc("/status/rss", statusHandler.HandleGetRSSFeed).Methods("GET", "OPTIONS")

	// ── API Version Management (public read, admin write) ─────────────────────
	api.HandleFunc("/api/versions", versionHandler.HandleListVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/api/versions/{version}", versionHandler.HandleGetVersion).Methods("GET", "OPTIONS")
	api.HandleFunc("/api/versions/{version}/deprecate", authMiddleware.RequireAuth(versionHandler.HandleDeprecateVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api/versions", authMiddleware.RequireAuth(versionHandler.HandleCreateAPIVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api/versions/{version}", authMiddleware.RequireAuth(versionHandler.HandleUpdateAPIVersion)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/api/versions/{version}/set-default", authMiddleware.RequireAuth(versionHandler.HandleSetDefaultAPIVersion)).Methods("POST", "OPTIONS")

	// ── Factory (public read, protected write) ────────────────────────────────
	api.HandleFunc("/factory/status", factoryHandler.HandleStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/opportunities", factoryHandler.HandleListOpportunities).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/opportunities/{id}", factoryHandler.HandleGetOpportunity).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/opportunities/{id}/approve", authMiddleware.RequireAuth(factoryHandler.HandleApproveOpportunity)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/opportunities/{id}/reject", authMiddleware.RequireAuth(factoryHandler.HandleRejectOpportunity)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/reviews/pending", factoryHandler.HandleListPendingReviews).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/pipeline/run", authMiddleware.RequireAuth(factoryHandler.HandleRunPipeline)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/functions", factoryHandler.HandleListFunctions).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/config", authMiddleware.RequireAuth(factoryHandler.HandleGetConfig)).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/config", authMiddleware.RequireAuth(factoryHandler.HandleUpdateConfig)).Methods("PUT", "PATCH", "OPTIONS")
	api.HandleFunc("/factory/schedule/status", factoryHandler.HandleGetScheduleStatus).Methods("GET", "OPTIONS")

	// Factory Experiments (A/B testing)
	api.HandleFunc("/factory/experiments", experimentHandler.HandleListExperiments).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/experiments", authMiddleware.RequireAuth(experimentHandler.HandleCreateExperiment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/experiments/{id}", experimentHandler.HandleGetExperiment).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/experiments/{id}/status", authMiddleware.RequireAuth(experimentHandler.HandleUpdateExperimentStatus)).Methods("PUT", "PATCH", "OPTIONS")
	api.HandleFunc("/factory/experiments/{id}/stats", experimentHandler.HandleGetExperimentStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/experiments/{id}/winner", authMiddleware.RequireAuth(experimentHandler.HandleDetermineWinner)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/experiments/{id}/variants", authMiddleware.RequireAuth(experimentHandler.HandleAddVariant)).Methods("POST", "OPTIONS")
	api.HandleFunc("/factory/experiments/{expID}/variants/{variantID}", authMiddleware.RequireAuth(experimentHandler.HandleUpdateVariant)).Methods("PUT", "PATCH", "OPTIONS")
	api.HandleFunc("/factory/experiments/{expID}/variants/{variantID}", authMiddleware.RequireAuth(experimentHandler.HandleDeleteVariant)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/factory/experiments/{expID}/variants/{variantID}/metrics", experimentHandler.HandleGetVariantMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/factory/experiments/metrics", experimentHandler.HandleRecordMetric).Methods("POST", "OPTIONS")

	// ── Categorization (public read, protected write) ─────────────────────────
	api.HandleFunc("/categorization/taxonomy", categorizationHandler.HandleGetTaxonomy).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/categories", categorizationHandler.HandleGetCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/categories/{id}", categorizationHandler.HandleGetCategory).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/tags", categorizationHandler.HandleGetTags).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/categorize", categorizationHandler.HandleCategorize).Methods("POST", "OPTIONS")
	api.HandleFunc("/categorization/analyze", categorizationHandler.HandleAnalyzeCode).Methods("POST", "OPTIONS")
	api.HandleFunc("/categorization/functions/{id}", categorizationHandler.HandleGetFunctionCategory).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/functions/{id}", authMiddleware.RequireAuth(categorizationHandler.HandleUpdateFunctionCategory)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/categorization/functions/{id}/recategorize", authMiddleware.RequireAuth(categorizationHandler.HandleReCategorize)).Methods("POST", "OPTIONS")
	api.HandleFunc("/categorization/category/{category}", categorizationHandler.HandleGetFunctionsByCategory).Methods("GET", "OPTIONS")
	api.HandleFunc("/categorization/tag/{tag}", categorizationHandler.HandleGetFunctionsByTag).Methods("GET", "OPTIONS")

	// ── Analytics ────────────────────────────────────────────────────────────
	analyticsHandler.RegisterRoutes(api, authMiddleware)
	unifiedAnalyticsHandler := analyticshandler.NewUnifiedHandler(unifiedAnalyticsSvc)
	unifiedAnalyticsHandler.RegisterUnifiedRoutes(api, authMiddleware)

	// ── Teams (protected) ─────────────────────────────────────────────────────
	protected.HandleFunc("/teams", authMiddleware.RequireAuth(teamHandler.HandleCreateTeam)).Methods("POST")
	protected.HandleFunc("/teams", authMiddleware.RequireAuth(teamHandler.HandleListTeams)).Methods("GET")
	protected.HandleFunc("/teams/{teamId}", authMiddleware.RequireAuth(teamHandler.HandleGetTeam)).Methods("GET")
	protected.HandleFunc("/teams/{teamId}", authMiddleware.RequireAuth(teamHandler.HandleUpdateTeam)).Methods("PATCH")
	protected.HandleFunc("/teams/{teamId}", authMiddleware.RequireAuth(teamHandler.HandleDeleteTeam)).Methods("DELETE")
	protected.HandleFunc("/teams/{teamId}/members", authMiddleware.RequireAuth(teamHandler.HandleAddTeamMember)).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/members/{userId}", authMiddleware.RequireAuth(teamHandler.HandleUpdateTeamMember)).Methods("PATCH")
	protected.HandleFunc("/teams/{teamId}/members/{userId}", authMiddleware.RequireAuth(teamHandler.HandleRemoveTeamMember)).Methods("DELETE")
	protected.HandleFunc("/teams/{teamId}/permissions", authMiddleware.RequireAuth(teamHandler.HandleGrantTeamPermission)).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/permissions/{resourceType}/{resourceId}", authMiddleware.RequireAuth(teamHandler.HandleRevokeTeamPermission)).Methods("DELETE")
	protected.HandleFunc("/teams/{teamId}/permissions", authMiddleware.RequireAuth(teamHandler.HandleGetTeamPermissions)).Methods("GET")
	protected.HandleFunc("/user/teams", authMiddleware.RequireAuth(teamHandler.HandleGetUserTeams)).Methods("GET")
	protected.HandleFunc("/permissions/{resourceType}/{resourceId}", authMiddleware.RequireAuth(teamHandler.HandleCheckUserResourcePermission)).Methods("GET")

	// ── Team Memory Engine (Shared Brain) ────────────────────────────────────
	teamMemoryHandler := teammemoryhandler.NewHandler(s.repo)
	protected.HandleFunc("/teams/{teamId}/memories", authMiddleware.RequireAuth(teamMemoryHandler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories", authMiddleware.RequireAuth(teamMemoryHandler.HandleListMemories)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/search", authMiddleware.RequireAuth(teamMemoryHandler.HandleSearchMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/query", authMiddleware.RequireAuth(teamMemoryHandler.HandleQueryMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/{memoryId}", authMiddleware.RequireAuth(teamMemoryHandler.HandleGetMemory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/{memoryId}", authMiddleware.RequireAuth(teamMemoryHandler.HandleUpdateMemory)).Methods("PUT", "PATCH", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/{memoryId}", authMiddleware.RequireAuth(teamMemoryHandler.HandleDeleteMemory)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/{memoryId}/validate", authMiddleware.RequireAuth(teamMemoryHandler.HandleValidateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/extractions", authMiddleware.RequireAuth(teamMemoryHandler.HandleListExtractions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/extractions/{extractionId}/approve", authMiddleware.RequireAuth(teamMemoryHandler.HandleApproveExtraction)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/teams/{teamId}/memories/extractions/{extractionId}/reject", authMiddleware.RequireAuth(teamMemoryHandler.HandleRejectExtraction)).Methods("POST", "OPTIONS")

	// ── Team Decisions (protected) ───────────────────────────────────────────
	decisionsHandler.RegisterRoutes(protected)

	// ── Providers (protected) ─────────────────────────────────────────────────
	// Provider operations are rate-limited per tenant to prevent abuse
	protected.HandleFunc("/providers", authMiddleware.RequireAuth(providersHandler.HandleListProviders)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/providers/credentials", authMiddleware.RequireAuth(providersHandler.HandleGetProviderCredentials)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/providers/connect", authMiddleware.RequireAuth(providerRateLimiter.LimitConnect(providersHandler.HandleConnectProvider))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/providers/validate", authMiddleware.RequireAuth(providersHandler.HandleValidateProvider)).Methods("POST")
	protected.HandleFunc("/providers/cost-estimate", authMiddleware.RequireAuth(providersHandler.HandleEstimateCost)).Methods("POST")
	protected.HandleFunc("/providers/{providerId}", authMiddleware.RequireAuth(providerRateLimiter.LimitDisconnect(providersHandler.HandleDisconnectProvider))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/providers/{providerId}/test", authMiddleware.RequireAuth(providerRateLimiter.LimitTest(providersHandler.HandleTestConnection))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/providers/{providerId}/rotate", authMiddleware.RequireAuth(providerRateLimiter.LimitConnect(providersHandler.HandleRotateProvider))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/providers/failover-test", authMiddleware.RequireAuth(providersHandler.HandleRunFailoverTest)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/providers/{providerId}/share", authMiddleware.RequireAuth(providersHandler.HandleShareProvider)).Methods("POST")
	protected.HandleFunc("/teams/invites", authMiddleware.RequireAuth(providersHandler.HandleCreateTeamInvite)).Methods("POST")

	// ── Apps (protected) ──────────────────────────────────────────────────────
	protected.HandleFunc("/apps", authMiddleware.RequireAuth(appsHandler.HandleListApps)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/apps", authMiddleware.RequireAuth(appsHandler.HandleCreateApp)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/apps/{appId}", authMiddleware.RequireAuth(appsHandler.HandleGetApp)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/apps/{appId}/status", authMiddleware.RequireAuth(appsHandler.HandleGetAppStatus)).Methods("GET", "OPTIONS")

	// ── Functions (protected) ─────────────────────────────────────────────────
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleGetFunction)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleUpdateFunction)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleDeleteFunction)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/functions/{id}/logs", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionLogs)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/functions/{id}/deployments", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionDeployments)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/functions/{id}/metrics", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionMetrics)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/functions/deployments/{deploymentId}", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionDeployment)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/functions/deploy", authMiddleware.RequireAuth(functionsHandler.HandleDeployFunction)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/functions/test", authMiddleware.RequireAuth(functionsHandler.HandleTestFunction)).Methods("POST", "OPTIONS")

	// ── Dashboard (protected, tenant-scoped) ──────────────────────────────────
	protected.HandleFunc("/dashboard/usage", authMiddleware.RequireAuth(dashboardHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/execution-rate", authMiddleware.RequireAuth(dashboardHandler.HandleGetExecutionRate)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/activity", authMiddleware.RequireAuth(dashboardHandler.HandleGetActivity)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/metrics", authMiddleware.RequireAuth(dashboardHandler.HandleGetMetrics)).Methods("GET", "OPTIONS")

	// ── State Usage (billing/quota integration) ─────────────────────────────
	protected.HandleFunc("/usage/state", authMiddleware.RequireAuth(stateUsageHandler.GetCurrentStateUsage)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/usage/state/history", authMiddleware.RequireAuth(stateUsageHandler.GetStateUsageHistory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/usage/state/quota", authMiddleware.RequireAuth(stateUsageHandler.GetStateQuotaStatus)).Methods("GET", "OPTIONS")

	// ── Enterprise SLA (protected) ────────────────────────────────────────────
	protected.HandleFunc("/enterprise/sla/overview", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetSLAOverview)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/enterprise/sla/uptime-history", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetUptimeHistory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/enterprise/sla/incidents", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetIncidents)).Methods("GET", "OPTIONS")

	// ── State (protected, rate-limited per-tenant) ────────────────────────────
	protected.HandleFunc("/state", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleListStates))).Methods("GET")
	protected.HandleFunc("/state", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleCreateState))).Methods("POST")
	protected.HandleFunc("/state/{path}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetState))).Methods("GET")
	protected.HandleFunc("/state/{path}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleUpdateState))).Methods("PUT")
	protected.HandleFunc("/state/{path}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleDeleteState))).Methods("DELETE")
	protected.HandleFunc("/state/{path}/value", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleSetValue))).Methods("PUT")
	protected.HandleFunc("/state/{path}/value", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandlePatchValue))).Methods("PATCH")
	protected.HandleFunc("/state/{path}/value", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetValue))).Methods("GET")
	protected.HandleFunc("/state/{path}/value", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleDeleteValue))).Methods("DELETE")
	protected.HandleFunc("/state/{path}/history", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetHistory))).Methods("GET")
	protected.HandleFunc("/state/{path}/snapshot", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleCreateSnapshot))).Methods("POST")
	protected.HandleFunc("/state/{path}/snapshots", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleListSnapshots))).Methods("GET")
	protected.HandleFunc("/state/{path}/restore", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleRestoreSnapshot))).Methods("POST")
	protected.HandleFunc("/state/{path}/time-travel", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleTimeTravel))).Methods("GET")
	protected.HandleFunc("/state/{path}/permissions", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGrantPermission))).Methods("POST")
	protected.HandleFunc("/state/{path}/permissions", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetPermissions))).Methods("GET")
	protected.HandleFunc("/triggers", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetTriggers))).Methods("GET")
	protected.HandleFunc("/triggers", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleCreateTrigger))).Methods("POST")
	protected.HandleFunc("/triggers/{id}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleDeleteTrigger))).Methods("DELETE")

	// ── State Encryption (protected, admin only) ─────────────────────────────
	protected.HandleFunc("/state/encrypt", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleMigrateEncryption))).Methods("POST")
	protected.HandleFunc("/state/encryption-stats", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleGetEncryptionStats))).Methods("GET")
	protected.HandleFunc("/state/{path}/enable-encryption", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleEnableEncryption))).Methods("POST")
	protected.HandleFunc("/state/rotate-key", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateHandler.HandleRotateEncryptionKey))).Methods("POST")

	// ── Agent Memory (protected) ──────────────────────────────────────────────
	protected.HandleFunc("/memories", authMiddleware.RequireAuth(memoryHandler.HandleListMemories)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/memories", authMiddleware.RequireAuth(memoryHandler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/memories/search", authMiddleware.RequireAuth(memoryHandler.HandleSearchMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleGetMemory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleUpdateMemory)).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleDeleteMemory)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/agent-memories", authMiddleware.RequireAuth(agentMemoryHandler.HandleListMemories)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent-memories", authMiddleware.RequireAuth(agentMemoryHandler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/search", authMiddleware.RequireAuth(agentMemoryHandler.HandleSearchMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/index", authMiddleware.RequireAuth(agentMemoryHandler.HandleRebuildIndex)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}", authMiddleware.RequireAuth(agentMemoryHandler.HandleGetMemory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}", authMiddleware.RequireAuth(agentMemoryHandler.HandleUpdateMemory)).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}", authMiddleware.RequireAuth(agentMemoryHandler.HandleDeleteMemory)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}/accessed", authMiddleware.RequireAuth(agentMemoryHandler.HandleMarkAccessed)).Methods("POST", "OPTIONS")

	// ── State Fabric (protected, rate-limited per-tenant) ─────────────────────
	protected.HandleFunc("/state-fabrics", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleList))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleCreate))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleGet))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleUpdate))).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleDelete))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/metrics", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleGetMetrics))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/stores", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleListStores))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/stores", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleCreateStore))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/stores/{storeId}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleDeleteStore))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/pipelines", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleListPipelines))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/pipelines", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleCreatePipeline))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/pipelines/{pipelineId}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleUpdatePipeline))).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/pipelines/{pipelineId}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleDeletePipeline))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/pipelines/{pipelineId}/execute", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleExecutePipeline))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/events", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleListEvents))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/snapshots", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleListSnapshots))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/snapshots", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleCreateSnapshot))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/snapshots/{snapshotId}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleDeleteSnapshot))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/replays", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleListReplays))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/replays", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleCreateReplay))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/state-fabrics/{id}/replays/{replayId}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(stateFabricHandler.HandleGetReplay))).Methods("GET", "OPTIONS")

	// ── Secrets Vault (protected) ─────────────────────────────────────────────
	protected.HandleFunc("/vault/secrets", authMiddleware.RequireAuth(vaultHandler.HandleListSecrets)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets", authMiddleware.RequireAuth(vaultRateLimiter.LimitCreate(vaultHandler.HandleCreateSecret))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleGetSecret)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleUpdateSecret)).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleDeleteSecret)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/tokens", authMiddleware.RequireAuth(vaultHandler.HandleListTokens)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/tokens", authMiddleware.RequireAuth(vaultRateLimiter.LimitGenerateToken(vaultHandler.HandleGenerateToken))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/vault/tokens/{id}", authMiddleware.RequireAuth(vaultHandler.HandleRevokeToken)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/vault/audit", authMiddleware.RequireAuth(vaultHandler.HandleGetAuditLog)).Methods("GET", "OPTIONS")

	// ── Secret Versions (protected) ─────────────────────────────────────────
	protected.HandleFunc("/vault/secrets/{id}/versions", authMiddleware.RequireAuth(vaultHandler.HandleListSecretVersions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/versions/{version}", authMiddleware.RequireAuth(vaultHandler.HandleGetSecretVersion)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/versions/diff", authMiddleware.RequireAuth(vaultHandler.HandleDiffSecretVersions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/rollback", authMiddleware.RequireAuth(vaultHandler.HandleRollbackSecret)).Methods("POST", "OPTIONS")

	// ── Backends (protected) ──────────────────────────────────────────────────
	protected.HandleFunc("/apps/{appId}/backends", authMiddleware.RequireAuth(backendsHandler.HandleCreateBackend)).Methods("POST")
	protected.HandleFunc("/apps/{appId}/backends", authMiddleware.RequireAuth(backendsHandler.HandleListBackends)).Methods("GET")
	protected.HandleFunc("/apps/{appId}/deploy/blue-green", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(backendsHandler.HandleDeployBlueGreen))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/link", authMiddleware.RequireAuth(backendsHandler.HandleLinkProject)).Methods("POST")
	protected.HandleFunc("/apps/{appId}/secrets", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(backendsHandler.HandleSetSecrets))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/secrets", authMiddleware.RequireAuth(backendsHandler.HandleListSecrets)).Methods("GET")
	protected.HandleFunc("/apps/{appId}/route", authMiddleware.RequireAuth(backendsHandler.HandleGetRoute)).Methods("GET")

	// ── Deployments (protected) ───────────────────────────────────────────────
	protected.HandleFunc("/apps/{appId}/deploy", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(deploymentsHandler.HandleDeploy))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/deployments", authMiddleware.RequireAuth(deploymentsHandler.HandleListDeployments)).Methods("GET")
	protected.HandleFunc("/deployments/{deploymentId}", authMiddleware.RequireAuth(deploymentsHandler.HandleGetDeployment)).Methods("GET")
	protected.HandleFunc("/deployments/{deploymentId}/rollback", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(deploymentsHandler.HandleRollback))).Methods("POST")

	// ── Deploy Keys (protected) ─────────────────────────────────────────────
	protected.HandleFunc("/deploy-keys", authMiddleware.RequireAuth(deployKeysHandler.HandleCreate)).Methods("POST")
	protected.HandleFunc("/deploy-keys", authMiddleware.RequireAuth(deployKeysHandler.HandleList)).Methods("GET")
	protected.HandleFunc("/deploy-keys/{id}", authMiddleware.RequireAuth(deployKeysHandler.HandleGet)).Methods("GET")
	protected.HandleFunc("/deploy-keys/{id}", authMiddleware.RequireAuth(deployKeysHandler.HandleDelete)).Methods("DELETE")
	protected.HandleFunc("/deploy-keys/{id}/verify", authMiddleware.RequireAuth(deployKeysHandler.HandleVerify)).Methods("POST")

	// ── Function Webhooks (protected) ───────────────────────────────────────
	protected.HandleFunc("/function-webhooks", authMiddleware.RequireAuth(functionWebhooksHandler.HandleCreate)).Methods("POST")
	protected.HandleFunc("/function-webhooks", authMiddleware.RequireAuth(functionWebhooksHandler.HandleList)).Methods("GET")
	protected.HandleFunc("/function-webhooks/{id}", authMiddleware.RequireAuth(functionWebhooksHandler.HandleGet)).Methods("GET")
	protected.HandleFunc("/function-webhooks/{id}", authMiddleware.RequireAuth(functionWebhooksHandler.HandleUpdate)).Methods("PUT")
	protected.HandleFunc("/function-webhooks/{id}", authMiddleware.RequireAuth(functionWebhooksHandler.HandleDelete)).Methods("DELETE")
	protected.HandleFunc("/function-webhooks/{id}/deliveries", authMiddleware.RequireAuth(functionWebhooksHandler.HandleListDeliveries)).Methods("GET")
	protected.HandleFunc("/function-webhooks/{id}/test", authMiddleware.RequireAuth(functionWebhooksHandler.HandleTest)).Methods("POST")

	// ── Platform Swarm Controller (protected) ────────────────────────────────
	swarmControllerHandler.RegisterRoutes(protected, "/platform")

	// ── Support (protected; register on api so /v1/support/... is matched) ─────────
	supportHdlr.RegisterRoutes(api, authMiddleware)
	supportAdminHdlr.RegisterRoutes(protected)

	// ── Support WebSocket (real-time chat) ──────────────────────────────────────
	// Register on api subrouter so /v1/support/ws matches (auth middleware runs first)
	support.RegisterWebSocketRoutes(api, supportWSHub)
}
