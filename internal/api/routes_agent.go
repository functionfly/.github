package api

import (
	"net/http"

	agentbilling "github.com/functionfly/functionfly/internal/agent/billing"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	conversationshandler "github.com/functionfly/functionfly/internal/api/handlers/conversations"
	papercliphandler "github.com/functionfly/functionfly/internal/api/handlers/paperclip"
	"github.com/functionfly/functionfly/internal/api/handlers/webhooks"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/team_memory"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerAgentRoutes wires AEP (Agent Execution Plan), swarm/marketplace/evolution,
// executable conversations, Paperclip, and Stripe webhook endpoints.
func registerAgentRoutes(
	s *Server,
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	aepHandler *agenthandler.Handler,
	swarmHandler *agenthandler.SwarmHandler,
	sebgHandler *agenthandler.SEBGHandler,
	evolutionHandler *agenthandler.EvolutionHandler,
	daemonHandler *agenthandler.DaemonHandler,
	registryRepo *storageregistry.RegistryRepository,
	cacheService *cache.CacheService,
	platformFeeRepo *storageregistry.PlatformFeeRepository,
	billingOperationalRepo *storage.BillingOperationalRepository,
) {
	// ── AEP Discovery (public) ───────────────────────────────────────────────
	api.HandleFunc("/agent/discover", aepHandler.HandleDiscover).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent/discover/{author}/{name}", aepHandler.HandleDiscoverFunction).Methods("GET", "OPTIONS")

	// Agent execution (supports both X-Agent-API-Key and JWT auth)
	api.HandleFunc("/agent/execute/{author}/{name}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent/execute/{author}/{name}/{version}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")

	// ── AEP Management (protected) ───────────────────────────────────────────

	// Initialize team memory middleware for agent API (auto-injects team context into prompts)
	// This enables automatic team memory context injection for code generation and agent prompts
	var agentTeamMiddleware func(http.Handler) http.Handler
	if s.postgresDB != nil {
		promptInjector := team_memory.NewAgentPromptInjector(s.repo)
		middlewareFactory := team_memory.NewMiddlewareFactory(promptInjector)
		// Enable for generation and execution endpoints
		middlewareFactory.EnablePath("/agent/generate").
			EnablePath("/agent/execute").
			EnablePath("/v1/agent/generate").
			EnablePath("/v1/agent/execute")
		agentTeamMiddleware = middlewareFactory.Create()
	}

	// Wrap agent handlers with team memory middleware
	protected.HandleFunc("/agent/register", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleRegisterAgent, agentTeamMiddleware))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleListAgents, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetAgent, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleDeleteAgent, agentTeamMiddleware))).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/quota", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleUpdateQuota, agentTeamMiddleware))).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/usage", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetUsage, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetPolicy, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleUpdatePolicy, agentTeamMiddleware))).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/executions", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleListExecutions, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/executions/{exec_id}", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetExecution, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/analytics", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetAnalytics, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/start", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleStartSession, agentTeamMiddleware))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}/end", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleEndSession, agentTeamMiddleware))).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleGetSession, agentTeamMiddleware))).Methods("GET", "OPTIONS")
	// NOTE: /wallet is now handled by swarmHandler to avoid route conflicts
	protected.HandleFunc("/agent/{agent_id}/billing/summary", authMiddleware.RequireAuth(aepHandler.HandleGetBillingSummary)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/billing/spend-cap", authMiddleware.RequireAuth(aepHandler.HandleUpdateSpendCap)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/cost-breakdown", authMiddleware.RequireAuth(aepHandler.HandleGetCostBreakdown)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/credits/balance", authMiddleware.RequireAuth(aepHandler.HandleGetCreditBalance)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/credits/purchase", authMiddleware.RequireAuth(aepHandler.HandlePurchaseCredits)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/credits/checkout", authMiddleware.RequireAuth(aepHandler.HandleCreateCreditsCheckout)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/transactions", authMiddleware.RequireAuth(aepHandler.HandleListAgentTransactions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/concurrency/stats", authMiddleware.RequireAuth(aepHandler.HandleGetConcurrencyStats)).Methods("GET", "OPTIONS")

	// ── Swarm / Marketplace / Evolution (protected) ───────────────────────────
	// Swarm routes MUST be registered AFTER AEP routes to take precedence for overlapping paths
	// Swarm handles: /agent/{id}/wallet, /agent/{id}/children, /agent/{id}/parent, /agent/{id}/spawn, etc.
	swarmHandler.RegisterRoutes(protected, "", authMiddleware)

	// ── SEBG (Self-Evolving Backend Graph) ───────────────────────────────────
	// SEBG exposes: /agent/{id}/sebg/proposals, /sebg/decide, /sebg/tier, /sebg/evolve, /sebg/roi
	sebgHandler.RegisterRoutes(protected, "", authMiddleware)

	// ── Agent Evolution API ─────────────────────────────────────────────────
	// Evolution exposes: /agents/{id}/evolution/suggestions, /evolution/auto-enable, /evolution/history
	evolutionHandler.RegisterRoutes(protected, "", authMiddleware)

	// ── Agent Daemon (Always-On) API ─────────────────────────────────────────
	// Daemon exposes: /agents/{id}/daemon/start, /daemon/stop, /daemon/status, /daemon/config
	daemonHandler.RegisterDaemonRoutes(protected, "", authMiddleware)

	// ── Paperclip (public webhook) ────────────────────────────────────────────
	paperclipAdapter := papercliphandler.NewAdapter(logrus.New())
	papercliphandler.RegisterRoutes(api, paperclipAdapter)

	// ── Stripe Webhook (public — no auth) ─────────────────────────────────────
	// Initialize dispute and refund repositories for chargeback/refund handling
	disputeRepo := storage.NewDisputeRepository(s.postgresDB.GORM)
	refundRepo := storage.NewRefundRepository(s.postgresDB.GORM)
	stripeWebhookHandler := webhooks.NewStripeWebhookHandler(
		storage.NewFinancialTransactionRepository(s.postgresDB.GORM),
		agentbilling.NewController(s.postgresDB.GORM, s.redisClient),
		s.notificationSvc,
		s.repo,
		platformFeeRepo,
		statefabricaddons.NewRepository(s.postgresDB.GORM),
		disputeRepo,
		refundRepo,
		registryRepo,
	)
	// Wire up dunning manager for automated payment retry
	stripeWebhookHandler.SetDunningManager(s.dunningManager)
	// Wire up operational repository for webhook payload storage and replay
	stripeWebhookHandler.SetOperationalRepository(billingOperationalRepo)
	stripeWebhookHandler.RegisterRoutes(api)

	// ── Executable Conversations ──────────────────────────────────────────────
	conversationRepo := storage.NewConversationRepository(s.postgresDB.GORM)
	convHandler := conversationshandler.NewHandler(
		conversationRepo,
		registryRepo,
		s.notificationSvc,
		s.repo,
		logrus.New(),
	)

	// Wire up team memory extraction webhook for conversation resolution
	// This enables automatic memory extraction when team conversations are resolved
	if s.postgresDB != nil {
		// Create the auto-updater with AI service integration
		autoUpdater := team_memory.NewAutoUpdater(s.repo, conversationRepo, nil)

		// Register the team memory event handler with the default publisher
		team_memory.RegisterTeamMemoryHandler(autoUpdater)

		// Set the memory event publisher on the conversation handler
		convHandler.SetMemoryEventPublisher(team_memory.DefaultEventPublisher)
	}

	// Initialize conversations WebSocket hub for real-time messaging
	convWSHub := conversationshandler.NewConversationWebSocketHub(logrus.New())
	go convWSHub.Run()
	convHandler.SetWebSocketHub(convWSHub)
	conversationshandler.RegisterConversationWSRoute(api, convWSHub)

	api.HandleFunc("/u/{username}/conversations/context", authMiddleware.RequireAuth(convHandler.GetConversationContext)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/from-thread", authMiddleware.RequireAuth(convHandler.CreateFromThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/collaboration-profile/{user_id}", authMiddleware.RequireAuth(convHandler.GetCollaborationProfile)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/search", authMiddleware.RequireAuth(convHandler.SearchMessages)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations", authMiddleware.RequireAuth(convHandler.ListConversations)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations", authMiddleware.RequireAuth(convHandler.CreateConversation)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}", authMiddleware.RequireAuth(convHandler.GetConversation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/read", authMiddleware.RequireAuth(convHandler.MarkConversationRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages", authMiddleware.RequireAuth(convHandler.ListMessages)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/validate", authMiddleware.RequireAuth(convHandler.ValidateMessage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages", authMiddleware.RequireAuth(convHandler.CreateMessage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", authMiddleware.RequireAuth(convHandler.GetMessage)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", authMiddleware.RequireAuth(convHandler.EditMessage)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", authMiddleware.RequireAuth(convHandler.DeleteMessage)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments", authMiddleware.RequireAuth(convHandler.ListAttachments)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments", authMiddleware.RequireAuth(convHandler.UploadAttachment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments/{attachment_id}", authMiddleware.RequireAuth(convHandler.GetAttachment)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments/{attachment_id}", authMiddleware.RequireAuth(convHandler.DeleteAttachment)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/resolve", authMiddleware.RequireAuth(convHandler.ResolveConversation)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/bounties", authMiddleware.RequireAuth(convHandler.ListBounties)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/bounties", authMiddleware.RequireAuth(convHandler.CreateBounty)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/bounties/{bounty_id}/claim", authMiddleware.RequireAuth(convHandler.ClaimBounty)).Methods("POST", "OPTIONS")
}

// wrapWithTeamMiddleware wraps an HTTP handler with optional team memory middleware
// If middleware is nil, the handler is returned unchanged
func wrapWithTeamMiddleware(handler http.HandlerFunc, middleware func(http.Handler) http.Handler) http.HandlerFunc {
	if middleware == nil {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}
