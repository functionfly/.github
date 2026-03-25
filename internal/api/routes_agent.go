package api

import (
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
	protected.HandleFunc("/agent/register", authMiddleware.RequireAuth(aepHandler.HandleRegisterAgent)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent", authMiddleware.RequireAuth(aepHandler.HandleListAgents)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(aepHandler.HandleGetAgent)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(aepHandler.HandleDeleteAgent)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/quota", authMiddleware.RequireAuth(aepHandler.HandleUpdateQuota)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/usage", authMiddleware.RequireAuth(aepHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(aepHandler.HandleGetPolicy)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(aepHandler.HandleUpdatePolicy)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/executions", authMiddleware.RequireAuth(aepHandler.HandleListExecutions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/executions/{exec_id}", authMiddleware.RequireAuth(aepHandler.HandleGetExecution)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/analytics", authMiddleware.RequireAuth(aepHandler.HandleGetAnalytics)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/start", authMiddleware.RequireAuth(aepHandler.HandleStartSession)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}/end", authMiddleware.RequireAuth(aepHandler.HandleEndSession)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}", authMiddleware.RequireAuth(aepHandler.HandleGetSession)).Methods("GET", "OPTIONS")
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
	flywheelRepo := flywheel.NewRepository(s.postgresDB.GORM)
	flywheelExecSvc := flywheel.NewExecutionAdapter(registryRepo, cacheService, flywheelexecution.NewLocalExecutor(), logrus.New())
	flywheelService := flywheel.NewService(flywheelRepo, flywheelExecSvc, logrus.New())

	flywheelWSHub := flywheelhandler.NewWebSocketHub(logrus.New())
	go flywheelWSHub.Run()

	flywheelHandler := flywheelhandler.NewHandler(flywheelService, flywheelWSHub, logrus.New())

	// Categories (public)
	api.HandleFunc("/flywheel/categories", flywheelHandler.ListCategories).Methods("GET", "OPTIONS")

	// Threads (public read, protected write)
	api.HandleFunc("/flywheel/threads", flywheelHandler.ListThreads).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads", flywheelRateLimiter.LimitCreateThread(authMiddleware.RequireAuth(flywheelHandler.CreateThread))).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}", flywheelHandler.GetThread).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}", authMiddleware.RequireAuth(flywheelHandler.UpdateThread)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/resolve", authMiddleware.RequireAuth(flywheelHandler.ResolveThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.SubscribeToThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.UnsubscribeFromThread)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/replies", flywheelHandler.ListReplies).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/replies", flywheelRateLimiter.LimitCreateReply(authMiddleware.RequireAuth(flywheelHandler.CreateReply))).Methods("POST", "OPTIONS")

	// Replies (protected for execution)
	api.HandleFunc("/flywheel/replies/{id}/execute", flywheelRateLimiter.LimitExecuteReply(authMiddleware.RequireAuth(flywheelHandler.ExecuteReply))).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/replies/{id}/verify", authMiddleware.RequireAuth(flywheelHandler.VerifyReply)).Methods("POST", "OPTIONS")

	// Reputation & Leaderboards (public)
	api.HandleFunc("/flywheel/reputation/me", authMiddleware.RequireAuth(flywheelHandler.GetMyReputation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/reputation/{user_id}", flywheelHandler.GetUserReputation).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/leaderboards", flywheelHandler.GetLeaderboardQuery).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/leaderboards/{score_type}", flywheelHandler.GetLeaderboard).Methods("GET", "OPTIONS")

	// Challenges
	api.HandleFunc("/flywheel/challenges", flywheelHandler.ListChallenges).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}", flywheelHandler.GetChallenge).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}/submit", flywheelRateLimiter.LimitSubmitChallenge(authMiddleware.RequireAuth(flywheelHandler.SubmitChallenge))).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}/leaderboard", flywheelHandler.GetChallengeLeaderboard).Methods("GET", "OPTIONS")

	// Real-time WebSocket
	api.HandleFunc("/flywheel/ws", flywheelHandler.HandleWebSocket)

	// Search & verified solutions
	api.HandleFunc("/flywheel/search", flywheelHandler.Search).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/solutions/verified", flywheelHandler.ListVerifiedSolutions).Methods("GET", "OPTIONS")

	// Thread replay / timeline
	api.HandleFunc("/flywheel/threads/{id}/timeline", flywheelHandler.GetThreadTimeline).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/replay", flywheelHandler.ReplayThread).Methods("POST", "OPTIONS")

	// Agent collaboration (protected)
	api.HandleFunc("/flywheel/threads/{id}/agents", authMiddleware.RequireAuth(flywheelHandler.ListThreadAgents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/invite", authMiddleware.RequireAuth(flywheelHandler.InviteAgent)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}", authMiddleware.RequireAuth(flywheelHandler.RemoveAgent)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/respond", authMiddleware.RequireAuth(flywheelHandler.AgentRespond)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/replies/{id}/publish-to-marketplace", authMiddleware.RequireAuth(flywheelHandler.PublishToMarketplace)).Methods("POST", "OPTIONS")

	// ── Executable Conversations ──────────────────────────────────────────────
	conversationRepo := storage.NewConversationRepository(s.postgresDB.GORM)
	convHandler := conversationshandler.NewHandler(conversationRepo, flywheelService, registryRepo, logrus.New())
	api.HandleFunc("/conversations/context", authMiddleware.RequireAuth(convHandler.GetConversationContext)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations/from-thread", authMiddleware.RequireAuth(convHandler.CreateFromThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/collaboration-profile/{user_id}", authMiddleware.RequireAuth(convHandler.GetCollaborationProfile)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations", authMiddleware.RequireAuth(convHandler.ListConversations)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations", authMiddleware.RequireAuth(convHandler.CreateConversation)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/{id}", authMiddleware.RequireAuth(convHandler.GetConversation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations/{id}/messages", authMiddleware.RequireAuth(convHandler.ListMessages)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations/{id}/messages/validate", authMiddleware.RequireAuth(convHandler.ValidateMessage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/{id}/messages", authMiddleware.RequireAuth(convHandler.CreateMessage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/{id}/resolve", authMiddleware.RequireAuth(convHandler.ResolveConversation)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/{id}/bounties", authMiddleware.RequireAuth(convHandler.ListBounties)).Methods("GET", "OPTIONS")
	api.HandleFunc("/conversations/{id}/bounties", authMiddleware.RequireAuth(convHandler.CreateBounty)).Methods("POST", "OPTIONS")
	api.HandleFunc("/conversations/{id}/bounties/{bounty_id}/claim", authMiddleware.RequireAuth(convHandler.ClaimBounty)).Methods("POST", "OPTIONS")
}
