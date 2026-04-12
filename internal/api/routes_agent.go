package api

import (
	"net/http"
	"os"

	agentbilling "github.com/functionfly/functionfly/internal/agent/billing"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	conversationshandler "github.com/functionfly/functionfly/internal/api/handlers/conversations"
	flywheelhandler "github.com/functionfly/functionfly/internal/api/handlers/flywheel"
	papercliphandler "github.com/functionfly/functionfly/internal/api/handlers/paperclip"
	flywheelexecution "github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/handlers/webhooks"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/team_memory"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerAgentRoutes wires AEP (Agent Execution Plan), swarm/marketplace/evolution,
// Flywheel network, executable conversations, Paperclip, and Stripe webhook endpoints.
func registerAgentRoutes(
	s *Server,
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	flywheelRateLimiter *middleware.FlywheelRateLimiter,
	aepHandler *agenthandler.Handler,
	swarmHandler *agenthandler.SwarmHandler,
	registryRepo *storageregistry.RegistryRepository,
	cacheService *cache.CacheService,
	platformFeeRepo *storageregistry.PlatformFeeRepository,
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

	// ── Paperclip (public webhook) ────────────────────────────────────────────
	paperclipAdapter := papercliphandler.NewAdapter(logrus.New())
	papercliphandler.RegisterRoutes(api, paperclipAdapter)

	// ── Stripe Webhook (public — no auth) ─────────────────────────────────────
	stripeWebhookHandler := webhooks.NewStripeWebhookHandler(
		storage.NewFinancialTransactionRepository(s.postgresDB.GORM),
		agentbilling.NewController(s.postgresDB.GORM, s.redisClient),
		s.notificationSvc,
		s.repo,
		platformFeeRepo,
		statefabricaddons.NewRepository(s.postgresDB.GORM),
	)
	stripeWebhookHandler.RegisterRoutes(api)

	// ── Flywheel Network (Proof-of-Execution Knowledge Network) ──────────────
	// Disabled via FLYWHEEL_ENABLED=false to focus on core product
	flywheelEnabled := os.Getenv("FLYWHEEL_ENABLED") == "true"
	var flywheelService *flywheel.Service
	if flywheelEnabled {
		flywheelRepo := flywheel.NewRepository(s.postgresDB.GORM)
		flywheelExecSvc := flywheel.NewExecutionAdapter(registryRepo, cacheService, flywheelexecution.NewLocalExecutor(), logrus.New())
		flywheelService = flywheel.NewService(flywheelRepo, flywheelExecSvc, logrus.New())

		flywheelWSHub := flywheelhandler.NewWebSocketHub(logrus.New())
		go flywheelWSHub.Run()

		flywheelHandler := flywheelhandler.NewHandler(flywheelService, flywheelWSHub, logrus.New())

		// Categories (public)
		api.HandleFunc("/flywheel/categories", flywheelHandler.ListCategories).Methods("GET", "OPTIONS")

		// Threads (public read, protected write with body size limits)
		api.HandleFunc("/flywheel/threads", flywheelHandler.ListThreads).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/threads", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxThreadBodySize)(flywheelRateLimiter.LimitCreateThread(authMiddleware.RequireAuth(flywheelHandler.CreateThread)))).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}", flywheelHandler.GetThread).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxThreadBodySize)(authMiddleware.RequireAuth(flywheelHandler.UpdateThread))).Methods("PATCH", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/resolve", authMiddleware.RequireAuth(flywheelHandler.ResolveThread)).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.SubscribeToThread)).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.UnsubscribeFromThread)).Methods("DELETE", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/replies", flywheelHandler.ListReplies).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/replies", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxReplyBodySize)(flywheelRateLimiter.LimitCreateReply(authMiddleware.RequireAuth(flywheelHandler.CreateReply)))).Methods("POST", "OPTIONS")

		// Replies (protected for execution with body size limits)
		api.HandleFunc("/flywheel/replies/{id}/execute", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxReplyBodySize)(flywheelRateLimiter.LimitExecuteReply(authMiddleware.RequireAuth(flywheelHandler.ExecuteReply)))).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/replies/{id}/verify", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxReplyBodySize)(authMiddleware.RequireAuth(flywheelHandler.VerifyReply))).Methods("POST", "OPTIONS")

		// Reputation & Leaderboards (public)
		api.HandleFunc("/flywheel/reputation/me", authMiddleware.RequireAuth(flywheelHandler.GetMyReputation)).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/reputation/{user_id}", flywheelHandler.GetUserReputation).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/leaderboards", flywheelHandler.GetLeaderboardQuery).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/leaderboards/{score_type}", flywheelHandler.GetLeaderboard).Methods("GET", "OPTIONS")

		// Challenges (with body size limits for submissions)
		api.HandleFunc("/flywheel/challenges", flywheelHandler.ListChallenges).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/challenges/{id}", flywheelHandler.GetChallenge).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/challenges/{id}/submit", middleware.BodySizeLimitMiddleware(middleware.FlywheelMaxChallengeBodySize)(flywheelRateLimiter.LimitSubmitChallenge(authMiddleware.RequireAuth(flywheelHandler.SubmitChallenge)))).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/challenges/{id}/leaderboard", flywheelHandler.GetChallengeLeaderboard).Methods("GET", "OPTIONS")

		// Real-time WebSocket
		api.HandleFunc("/flywheel/ws", flywheelHandler.HandleWebSocket)

		// Search & verified solutions
		api.HandleFunc("/flywheel/search", flywheelHandler.Search).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/solutions/verified", flywheelHandler.ListVerifiedSolutions).Methods("GET", "OPTIONS")

		// Thread replay / timeline
		api.HandleFunc("/flywheel/threads/{id}/timeline", flywheelHandler.GetThreadTimeline).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/replay", flywheelHandler.ReplayThread).Methods("POST", "OPTIONS")

		// Agent collaboration (protected with rate limiting)
		api.HandleFunc("/flywheel/threads/{id}/agents", authMiddleware.RequireAuth(flywheelHandler.ListThreadAgents)).Methods("GET", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/invite", flywheelRateLimiter.LimitAgentCollaboration(authMiddleware.RequireAuth(flywheelHandler.InviteAgent))).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}", flywheelRateLimiter.LimitAgentCollaboration(authMiddleware.RequireAuth(flywheelHandler.RemoveAgent))).Methods("DELETE", "OPTIONS")
		api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/respond", flywheelRateLimiter.LimitAgentCollaboration(authMiddleware.RequireAuth(flywheelHandler.AgentRespond))).Methods("POST", "OPTIONS")
		api.HandleFunc("/flywheel/replies/{id}/publish-to-marketplace", flywheelRateLimiter.LimitAgentCollaboration(authMiddleware.RequireAuth(flywheelHandler.PublishToMarketplace))).Methods("POST", "OPTIONS")
	}

	// ── Executable Conversations ──────────────────────────────────────────────
	// Conversations depend on flywheel service
	if flywheelEnabled {
		conversationRepo := storage.NewConversationRepository(s.postgresDB.GORM)
		convHandler := conversationshandler.NewHandler(
			conversationRepo,
			flywheelService,
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
		api.HandleFunc("/conversations/context", authMiddleware.RequireAuth(convHandler.GetConversationContext)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations/from-thread", authMiddleware.RequireAuth(convHandler.CreateFromThread)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/collaboration-profile/{user_id}", authMiddleware.RequireAuth(convHandler.GetCollaborationProfile)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations", authMiddleware.RequireAuth(convHandler.ListConversations)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations", authMiddleware.RequireAuth(convHandler.CreateConversation)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}", authMiddleware.RequireAuth(convHandler.GetConversation)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations/{id}/read", authMiddleware.RequireAuth(convHandler.MarkConversationRead)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}/messages", authMiddleware.RequireAuth(convHandler.ListMessages)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations/{id}/messages/validate", authMiddleware.RequireAuth(convHandler.ValidateMessage)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}/messages", authMiddleware.RequireAuth(convHandler.CreateMessage)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}/resolve", authMiddleware.RequireAuth(convHandler.ResolveConversation)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}/bounties", authMiddleware.RequireAuth(convHandler.ListBounties)).Methods("GET", "OPTIONS")
		api.HandleFunc("/conversations/{id}/bounties", authMiddleware.RequireAuth(convHandler.CreateBounty)).Methods("POST", "OPTIONS")
		api.HandleFunc("/conversations/{id}/bounties/{bounty_id}/claim", authMiddleware.RequireAuth(convHandler.ClaimBounty)).Methods("POST", "OPTIONS")
	}
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
