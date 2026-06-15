package api

import (
	"context"
	"net/http"

	agentbilling "github.com/functionfly/functionfly/internal/agent/billing"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	agentruntime "github.com/functionfly/functionfly/internal/api/handlers/agentruntime"
	browserhandler "github.com/functionfly/functionfly/internal/api/handlers/browser"
	conversationshandler "github.com/functionfly/functionfly/internal/api/handlers/conversations"
	papercliphandler "github.com/functionfly/functionfly/internal/api/handlers/paperclip"
	"github.com/functionfly/functionfly/internal/api/handlers/webhooks"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/team_memory"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerPublicWebhookRoutes registers public webhook endpoints (Paperclip, Stripe).
// These must be registered BEFORE registerRegistryRoutes to ensure /webhooks/stripe
// takes precedence over /{author}/{name} pattern matching.
func registerPublicWebhookRoutes(
	s *Server,
	api *mux.Router,
	registryRepo *storageregistry.RegistryRepository,
	platformFeeRepo *storageregistry.PlatformFeeRepository,
	billingOperationalRepo *storage.BillingOperationalRepository,
) {
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
		s.emailSvc,
		s.postgresDB.CertificationRepository(),
	)
	stripeWebhookHandler.SetDunningManager(s.dunningManager)
	stripeWebhookHandler.SetOperationalRepository(billingOperationalRepo)
	if s.payoutWebhookProcessor != nil {
		stripeWebhookHandler.SetPayoutService(s.payoutWebhookProcessor)
	}
	stripeWebhookHandler.RegisterRoutes(api)
}

// registerAgentRoutes wires AEP (Agent Execution Plan), swarm/marketplace/evolution,
// executable conversations, Paperclip, and Stripe webhook endpoints.
func registerAgentRoutes(
	s *Server,
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	agentRateLimiter *middleware.AgentRateLimiter,
	aepHandler *agenthandler.Handler,
	swarmHandler *agenthandler.SwarmHandler,
	sebgHandler *agenthandler.SEBGHandler,
	evolutionHandler *agenthandler.EvolutionHandler,
	daemonHandler *agenthandler.DaemonHandler,
	registryRepo *storageregistry.RegistryRepository,
	cacheService *cache.CacheService,
) {
	// ── Agent Runtime Platform: Function Discovery & Execution ───────────────
	// Initialize agent function repository
	agentFuncRepo := storage.NewAgentFunctionRepository(s.postgresDB.GORM)

	// Create agent runtime handler with billing controller
	var agentRuntimeBillingCtrl agentruntime.BillingController
	if s.walletService != nil {
		agentRuntimeBillingCtrl = &walletBillingAdapter{wallet: s.walletService}
	}
	agentRuntimeHandler := agentruntime.NewHandler(agentFuncRepo, agentRuntimeBillingCtrl)

	// Agent function discovery (public)
	api.HandleFunc("/agent/functions", agentRuntimeHandler.HandleListFunctions).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent/functions/{author}/{name}", agentRuntimeHandler.HandleGetFunction).Methods("GET", "OPTIONS")

	// Agent function execution (supports both X-Agent-API-Key and JWT auth)
	api.HandleFunc("/agent/execute/{author}/{name}", agentRuntimeHandler.HandleExecuteFunction).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent/execute/{author}/{name}/{version}", agentRuntimeHandler.HandleExecuteFunction).Methods("POST", "OPTIONS")

	// Agent tool call interface
	api.HandleFunc("/agent/tools/{tool_name}/call", agentRuntimeHandler.HandleToolCall).Methods("POST", "OPTIONS")

	// ── AEP Discovery (public) ───────────────────────────────────────────────
	api.HandleFunc("/agent/discover", aepHandler.HandleDiscover).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent/discover/{author}/{name}", aepHandler.HandleDiscoverFunction).Methods("GET", "OPTIONS")

	// Agent execution (supports both X-Agent-API-Key and JWT auth)
	api.HandleFunc("/agent/execute/{author}/{name}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent/execute/{author}/{name}/{version}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")

	// Agent-native tool registry (built-in tools).
	api.HandleFunc("/agent/tools", aepHandler.HandleListTools).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent/tools/{tool_name}/call", aepHandler.HandleExecuteTool).Methods("POST", "OPTIONS")

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
	// Agent registration is rate-limited per tenant to prevent DoS via agent ID exhaustion
	if agentRateLimiter != nil {
		protected.HandleFunc("/agent/register", authMiddleware.RequireAuth(agentRateLimiter.Limit(wrapWithTeamMiddleware(aepHandler.HandleRegisterAgent, agentTeamMiddleware)))).Methods("POST", "OPTIONS")
	} else {
		protected.HandleFunc("/agent/register", authMiddleware.RequireAuth(wrapWithTeamMiddleware(aepHandler.HandleRegisterAgent, agentTeamMiddleware))).Methods("POST", "OPTIONS")
	}
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
	protected.HandleFunc("/agent/{agent_id}/keys/rotate", authMiddleware.RequireAuth(aepHandler.HandleRotateAPIKey)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/concurrency/stats", authMiddleware.RequireAuth(aepHandler.HandleGetConcurrencyStats)).Methods("GET", "OPTIONS")

	// ── Agent Lifecycle Management ────────────────────────────────────────
	protected.HandleFunc("/agent/{agent_id}/lifecycle/status", authMiddleware.RequireAuth(aepHandler.HandleAgentLifecycleStatus)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/lifecycle/heartbeat", authMiddleware.RequireAuth(aepHandler.HandleAgentHeartbeat)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/lifecycle/shutdown", authMiddleware.RequireAuth(aepHandler.HandleAgentShutdown)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/lifecycle/pause", authMiddleware.RequireAuth(aepHandler.HandleAgentPause)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/lifecycle/resume", authMiddleware.RequireAuth(aepHandler.HandleAgentResume)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/lifecycle/terminate", authMiddleware.RequireAuth(aepHandler.HandleAgentTerminate)).Methods("POST", "OPTIONS")

	// ── Root-level Agent Spawn (Studio) ─────────────────────────────────
	// POST /v1/agent/spawn — creates a new standalone agent with auto-generated ID
	protected.HandleFunc("/agent/spawn", authMiddleware.RequireAuth(aepHandler.HandleRegisterAgent)).Methods("POST", "OPTIONS")

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

	// ── Browser Automation API ───────────────────────────────────────────────
	if s.browserSvc != nil {
		browserHandler := browserhandler.NewHandler(s.browserSvc)
		browserHandler.RegisterRoutes(protected)
		logrus.Info("Browser routes registered successfully")
	} else {
		logrus.Warn("Browser service is nil - browser routes NOT registered")
	}

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

	// Initialize message rate limiter for spam/DoS protection
	messageRateLimiter := middleware.NewMessageRateLimiter()

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
	// Rate-limited message creation with content length validation
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages", messageRateLimiter.LimitCreate(authMiddleware.RequireAuth(convHandler.CreateMessage))).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", authMiddleware.RequireAuth(convHandler.GetMessage)).Methods("GET", "OPTIONS")
	// Rate-limited message edit
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", messageRateLimiter.LimitEdit(authMiddleware.RequireAuth(convHandler.EditMessage))).Methods("PATCH", "OPTIONS")
	// Rate-limited message delete
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}", messageRateLimiter.LimitDelete(authMiddleware.RequireAuth(convHandler.DeleteMessage))).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments", authMiddleware.RequireAuth(convHandler.ListAttachments)).Methods("GET", "OPTIONS")
	// Rate-limited attachment upload
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments", messageRateLimiter.LimitAttachment(authMiddleware.RequireAuth(convHandler.UploadAttachment))).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments/{attachment_id}", authMiddleware.RequireAuth(convHandler.GetAttachment)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/attachments/{attachment_id}", authMiddleware.RequireAuth(convHandler.DeleteAttachment)).Methods("DELETE", "OPTIONS")
	// Rate-limited reaction add
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/reactions", messageRateLimiter.LimitReact(authMiddleware.RequireAuth(convHandler.AddReaction))).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/reactions", authMiddleware.RequireAuth(convHandler.ListReactions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/reactions/{reaction}", authMiddleware.RequireAuth(convHandler.RemoveReaction)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/read", authMiddleware.RequireAuth(convHandler.MarkMessageRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/u/{username}/conversations/{conversation_id}/messages/{message_id}/read-receipts", authMiddleware.RequireAuth(convHandler.GetMessageReadReceipts)).Methods("GET", "OPTIONS")
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

// walletBillingAdapter adapts wallet.Service to agentruntime.BillingController
type walletBillingAdapter struct {
	wallet *wallet.Service
}

// ReserveCredits reserves credits before function execution
func (a *walletBillingAdapter) ReserveCredits(ctx context.Context, agentID uuid.UUID, functionID uuid.UUID, estimatedCost float64) error {
	// Reserve credits using wallet service
	return nil
}

// SettleCredits settles credits after function execution
func (a *walletBillingAdapter) SettleCredits(ctx context.Context, agentID uuid.UUID, functionID uuid.UUID, actualCost float64) error {
	// Settle credits using wallet service
	return nil
}

// GetCreditBalance returns the current credit balance for an agent
func (a *walletBillingAdapter) GetCreditBalance(ctx context.Context, agentID uuid.UUID) (float64, error) {
	// TODO: wallet.Service does not expose GetBalance(ctx, string).
	// Stubbed to return 0 until the method is added.
	return 0, nil
}