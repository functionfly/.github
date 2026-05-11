package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/actuator"
	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/categorization"
	agentdeployment "github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/discovery"
	agenttesting "github.com/functionfly/functionfly/internal/agent/testing"
	secureSandbox "github.com/functionfly/functionfly/internal/agent/testing/sandbox"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/evolution"
	factorysvc "github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/learning"
	"github.com/functionfly/functionfly/internal/agent/marketplace"
	agentsecurity "github.com/functionfly/functionfly/internal/agent/security"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/functionfly/functionfly/internal/analytics"
	"github.com/functionfly/functionfly/internal/analytics/unified"
	"github.com/functionfly/functionfly/internal/api/docs"
	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	agentmemoryhandler "github.com/functionfly/functionfly/internal/api/handlers/agent_memory"
	analyticshandler "github.com/functionfly/functionfly/internal/api/handlers/analytics"
	"github.com/functionfly/functionfly/internal/api/handlers/apikeys"
	"github.com/functionfly/functionfly/internal/api/handlers/apps"
	"github.com/functionfly/functionfly/internal/api/handlers/blog"
	authHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/auth"
	"github.com/functionfly/functionfly/internal/api/handlers/backends"
	billinghandler "github.com/functionfly/functionfly/internal/api/handlers/billing"
	categorizationhandler "github.com/functionfly/functionfly/internal/api/handlers/categorization"
	"github.com/functionfly/functionfly/internal/api/handlers/chat"
	"github.com/functionfly/functionfly/internal/api/handlers/content"
	"github.com/functionfly/functionfly/internal/api/handlers/dashboard"
	"github.com/functionfly/functionfly/internal/api/handlers/decisions"
	"github.com/functionfly/functionfly/internal/api/handlers/deploykeys"
	"github.com/functionfly/functionfly/internal/api/handlers/deployments"
	enterprisePkg "github.com/functionfly/functionfly/internal/api/handlers/enterprise"
	factoryhandler "github.com/functionfly/functionfly/internal/api/handlers/factory"
	feedbackHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/feedback"
	followHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/follow"
	"github.com/functionfly/functionfly/internal/api/handlers/function_webhooks"
	"github.com/functionfly/functionfly/internal/api/handlers/functions"
	githubhandler "github.com/functionfly/functionfly/internal/api/handlers/github"
	"github.com/functionfly/functionfly/internal/manifest"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	"github.com/functionfly/functionfly/internal/api/handlers/newsletter"
	notificationHandlerPkg 	"github.com/functionfly/functionfly/internal/api/handlers/notifications"
	payoutsHandler "github.com/functionfly/functionfly/internal/api/handlers/payouts"
	"github.com/functionfly/functionfly/internal/api/handlers/playground"
	privacyhandler "github.com/functionfly/functionfly/internal/api/handlers/privacy"
	"github.com/functionfly/functionfly/internal/api/handlers/providers"
	"github.com/functionfly/functionfly/internal/api/handlers/recommendations"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	registryexecution "github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/handlers/security"
	"github.com/functionfly/functionfly/internal/api/handlers/state"
	"github.com/functionfly/functionfly/internal/api/handlers/statefabric"
	statushandler "github.com/functionfly/functionfly/internal/api/handlers/status"
	supportHandler "github.com/functionfly/functionfly/internal/api/handlers/support"
	"github.com/functionfly/functionfly/internal/api/handlers/teams"
	usersHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/users"
	"github.com/functionfly/functionfly/internal/api/handlers/vault"
	versionhandler "github.com/functionfly/functionfly/internal/api/handlers/version"
	"github.com/functionfly/functionfly/internal/api/handlers/wellknown"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/billing"
	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/captcha"
	"github.com/functionfly/functionfly/internal/currency"
	monitoringPkg "github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/privacy"
	paymentPkg "github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	statefabricrepo "github.com/functionfly/functionfly/internal/storage/statefabric"
	trustapirepo "github.com/functionfly/functionfly/internal/storage/trustapi"
	decisionsrepo "github.com/functionfly/functionfly/internal/storage/trustapi/decisions"
	vaultstorage "github.com/functionfly/functionfly/internal/storage/vault"
	timemachine "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/functionfly/functionfly/internal/support"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/functionfly/functionfly/internal/wasm"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// setupRoutes configures all API routes.
//
// The function is organised into three sections:
//  1. Handler / service initialization
//  2. Global middleware wiring
//  3. Calls to the focused register* helpers (one file per domain)
func (s *Server) setupRoutes(realtimeMonitor *monitoringPkg.RealtimeMonitor) {
	// ── Handler initialization ────────────────────────────────────────────────
	authHandler := authHandlerPkg.NewHandler(s.authSvc)
	tenantAuthHandler := authHandlerPkg.NewTenantAuthHandler(s.repo)
	usersHandler := usersHandlerPkg.NewHandler(s.repo, s.authSvc)
	favoritesHandler := usersHandlerPkg.NewFavoritesHandler(s.repo)
	presenceHandler := usersHandlerPkg.NewPresenceHandler(s.repo, s.authSvc, s.redisClient, logrus.New())
	// ── Wallet Service Initialization ────────────────────────────────────────────
	// Initialize the unified wallet system (replaces user_wallets and agent_billing_controls)
	walletRepo := wallet.NewRepository(s.postgresDB.GORM)
	walletService := wallet.NewService(walletRepo, s.redisClient)
	_ = walletService // Used by admin handlers and webhooks

	// Legacy platform fee repository (still used by registry handlers during migration)
	platformFeeRepo := registry.NewPlatformFeeRepository(s.postgresDB.GORM)
	sfAddonRepo := statefabricaddons.NewRepository(s.postgresDB.GORM)
	billingHandler := billinghandler.NewHandler(s.repo, platformFeeRepo, sfAddonRepo, s.redisClient)
	billingHandler.SetWalletService(walletService)
	tenantWebhookHandler := billinghandler.NewTenantWebhookHandler(s.repo)
	appsHandler := apps.NewHandler(s.repo)
	backendsHandler := backends.NewHandler(s.repo, s.routingSvc)
	deploymentsHandler := deployments.NewHandler(s.repo, s.deploySvc)
	pasteHandler := functions.NewPasteHandler(s.repo)
	functionsHandler := functions.NewHandler(s.repo, s.deploySvc, pasteHandler)
	unifiedAnalyticsSvc := unified.NewService(s.postgresDB.GORM, s.usageMetricsAgg)
	adminHandler := admin.NewHandler(s.repo, s.postgresDB.LoginAttemptRepository(), s.postgresDB.AnalyticsRepository(), s.authSvc, unifiedAnalyticsSvc, sfAddonRepo)
	adminHandler.SetWalletService(walletService)

	// Initialize billing operational repository for webhook replay and tax exemption management
	billingOperationalRepo := storage.NewBillingOperationalRepository(s.postgresDB.GORM)
	adminHandler.SetBillingOperationalRepository(billingOperationalRepo)
	adminBackendsHandler := admin.NewBackendsHandler(s.repo, s.authSvc)

	// Initialize dispute and refund repositories for chargeback/refund handling
	disputeRepo := storage.NewDisputeRepository(s.postgresDB.GORM)
	refundRepo := storage.NewRefundRepository(s.postgresDB.GORM)
	disputesHandler := admin.NewDisputesHandler(disputeRepo, refundRepo, s.repo)
	adminProvidersHandler := admin.NewProvidersHandler(s.repo, s.authSvc)
	securityHandler := security.NewHandler(s.repo, s.authSvc)

	maintenanceRepo := storage.NewMaintenanceRepository(s.postgresDB.GORM)
	maintenanceHandler := admin.NewMaintenanceHandler(maintenanceRepo, s.authSvc)
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(maintenanceRepo)
	contentHandler := content.NewHandler(s.repo)
	blogRepo := blog.NewBlogRepository(s.postgresDB.DB)
	blogHandler := blog.NewHandler(blogRepo)
	feedbackHandler := feedbackHandlerPkg.NewHandler(s.repo, s.storageService)

	followService := services.NewFollowService(s.repo)
	followHandler := followHandlerPkg.NewHandler(followService, s.repo, s.authSvc)

	apikeyRepo := apikey.NewRepository(s.postgresDB.GORM)
	apiKeysHandler := apikeys.NewHandler(apikeyRepo)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logrus.Fatal("FATAL: JWT_SECRET environment variable is required. Refusing to start with empty secret.")
	}
	apiKeyAuthHandler := apikeys.NewAPIKeyAuthHandler(apikeyRepo, jwtSecret)

	monitoringHandler := monitoring.NewHandler(s.repo, s.monitoringSvc, s.realtimeMonitor, s.authSvc)
	mfaHandler := mfaHandlerPkg.NewMFAHandler(s.authSvc)

	notificationHandler := notificationHandlerPkg.NewHandler(s.notificationSvc, s.notificationRepo)
	notificationWSHandler := notificationHandlerPkg.NewWebSocketHandler(
		notificationHandlerPkg.NewWebSocketHub(logrus.New()),
		logrus.New(),
	)
	s.notificationWSHandler = notificationWSHandler

	// Create a dedicated pgxpool for LISTEN connections (PostgreSQL notification subscription)
	// Reuse the same connection string as the main DB; pool is tiny so it doesn't compete with app queries.
	connStr := storage.GetConnectionString()
	if pool, err := pgxpool.New(context.Background(), connStr); err == nil {
		s.notificationPool = pool
		logrus.Info("Notification pgxpool created for LISTEN subscriptions")
	} else {
		logrus.WithError(err).Warn("Failed to create notification pgxpool – WebSocket push will rely on polling")
	}

	newsletterHandler := newsletter.NewHandler(s.repo, s.emailSvc)

	deployKeysRepo := storage.NewDeployKeysRepository(s.postgresDB.GORM)
	deployKeysHandler := deploykeys.NewHandler(deployKeysRepo)
	functionWebhookRepo := storage.NewFunctionWebhookRepository(s.postgresDB.GORM)
	functionWebhookService := storage.NewFunctionWebhookService(functionWebhookRepo)
	functionWebhooksHandler := function_webhooks.NewHandler(functionWebhookRepo, functionWebhookService)

	// ── GitHub Integration ────────────────────────────────────────────────────
	githubRepo := storage.NewGitHubRepository(s.postgresDB.DB)
	githubVaultKey := os.Getenv("GITHUB_VAULT_KEY")
	if githubVaultKey == "" {
		githubVaultKey = "default-dev-key-must-be-32-bytes!" // 32 bytes for dev
	}
	githubBaseURL := os.Getenv("FRONTEND_URL")
	if githubBaseURL == "" {
		githubBaseURL = "http://localhost:3000"
	}
	githubHandler := githubhandler.NewHandler(s.repo, githubRepo, nil, logrus.New(), githubVaultKey, githubBaseURL)
	githubHandler.SetAuthService(s.authSvc)

	// ── Real-time Usage Tracking ─────────────────────────────────────────────
	// Initialize the real-time usage tracker with Redis counters for quota enforcement
	realtimeUsageTracker := services.NewRealtimeUsageTracker(
		s.redisClient,
		s.repo,
		s.notificationSvc,
		services.DefaultRealtimeUsageConfig(),
	)
	// Start the background sync job to periodically sync counters to the database
	realtimeUsageTracker.Start(context.Background())
	logrus.Info("Real-time usage tracker initialized with background sync")

	// Initialize quota middleware for synchronous quota enforcement
	quotaMiddleware := middleware.NewQuotaMiddleware(realtimeUsageTracker, s.repo, &middleware.QuotaMiddlewareConfig{
		Enabled:              true,
		Mode:                 middleware.QuotaModeEnforce,
		AllowPublicFunctions: true,
	})
	_ = quotaMiddleware // Used for API routes that need quota enforcement

	// Initialize usage handler for real-time usage API endpoints
	usageHandler := billinghandler.NewUsageHandler(realtimeUsageTracker, s.repo)

	// Initialize state usage handler for state fabric billing/quota integration
	stateUsageHandler := billinghandler.NewStateUsageHandler(s.stateUsageAggregator, s.repo)

	// Initialize cost allocation handler for detailed cost tracking
	costAllocationHandler := billinghandler.NewCostAllocationHandler(s.repo)

	// Initialize export repository, service, and handlers for usage data export
	exportRepo := storage.NewExportRepository(s.postgresDB.DB)
	exportService := services.NewExportService(exportRepo, s.repo, s.emailSvc, os.Getenv("API_BASE_URL"))
	exportHandler := billinghandler.NewExportHandler(exportRepo, exportService, s.repo)

	// Initialize billing sync job for external billing integrations
	s.billingSyncJob = billing.NewBillingSyncJob(exportRepo)
	s.billingSyncJob.Start(context.Background())
	logrus.Info("Billing sync job initialized")

	externalBillingHandler := billinghandler.NewExternalBillingHandler(exportRepo, s.repo, s.billingSyncJob)

	// Initialize export scheduler for automated exports
	s.exportScheduler = services.NewExportScheduler(exportRepo, exportService)
	s.exportScheduler.Start()
	logrus.Info("Export scheduler initialized")

	// Initialize usage alert repository and forecast/alerter services
	alertRepo := storage.NewUsageAlertRepository(s.postgresDB.DB)
	forecastConfig := services.DefaultUsageForecasterConfig()
	usageForecaster := services.NewUsageForecaster(alertRepo, s.repo, forecastConfig)
	usageAlerter := services.NewUsageAlerter(alertRepo, s.repo, s.notificationSvc, usageForecaster, services.DefaultUsageAlerterConfig())

	// Initialize forecast handler for usage forecasting and alerts
	forecastHandler := billinghandler.NewUsageForecastHandler(alertRepo, s.repo, usageForecaster, usageAlerter)

	versionRepo := versioning.NewRepository(s.postgresDB.DB)
	versionHandler := versionhandler.NewHandler(versionRepo)

	appPlaygroundHandler := playground.NewHandler(s.repo)

	cacheConfiguration := cache.LoadCacheConfiguration()
	if err := cacheConfiguration.Validate(); err != nil {
		logrus.WithError(err).Error("Invalid cache configuration, disabling all caching features")
		cacheConfiguration.DiskEnabled = false
		cacheConfiguration.RedisEnabled = false
		cacheConfiguration.CDNEnabled = false
		cacheConfiguration.EdgeCacheEnabled = false
		if cacheConfiguration.MemoryMaxMB <= 0 {
			cacheConfiguration.MemoryMaxMB = 100
		}
		if cacheConfiguration.DefaultTTL <= 0 {
			cacheConfiguration.DefaultTTL = 3600
		}
		if cacheConfiguration.RedisRegistryTTL <= 0 {
			cacheConfiguration.RedisRegistryTTL = 600
		}
	}

	cacheService, err := cache.NewCacheService(s.postgresDB.GORM, s.redisClient, cacheConfiguration.ToCacheConfig())
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize cache service, attempting fallback configuration")
		fallbackConfig := &cache.CacheConfig{
			MaxMemoryMB:      100,
			EnableDiskCache:  false,
			EnableRedisCache: false,
			EnableCDNCaching: false,
			DefaultTTL:       3600,
			RedisRegistryTTL: 600,
		}
		cacheService, err = cache.NewCacheService(s.postgresDB.GORM, s.redisClient, fallbackConfig)
		if err != nil {
			logrus.WithError(err).Error("Failed to initialize fallback cache service")
			panic("Failed to initialize cache service even with fallback configuration: " + err.Error())
		}
		logrus.Warn("Cache service initialized with fallback in-memory-only configuration")
	}

	cdnService := cache.NewCDNService(cacheConfiguration.ToCDNConfig())

	var registryCache *cache.RegistryRedisCache
	if cacheConfiguration.RedisEnabled && s.redisClient != nil {
		registryCache = cacheService.GetRegistryCache()
	}
	registryRepo := registry.NewRegistryRepository(s.postgresDB.GORM, registryCache)

	// Initialize the persistent SandboxClient daemon for function execution.
	// This replaces per-request process spawning with a single long-lived runtime.
	if err := registryexecution.InitSandboxClient(); err != nil {
		logrus.WithError(err).Warn("Failed to init SandboxClient daemon, falling back to per-request executor")
	}

	edgeCache := cache.NewEdgeCacheService(
		cacheService.GetRegistryCache(),
		cdnService,
		cacheConfiguration.ToEdgeCacheConfig(),
	)
	edgeCache.SetRepository(registryRepo)

	// Create function repository for remix/fork operations
	functionRepo := storage.NewFunctionRepository(s.postgresDB.DB)

	registryHandler := registryhandler.NewHandler(registryRepo, s.repo, functionRepo, cacheService, cdnService, edgeCache, s.realtimeMonitor, platformFeeRepo, s.recommendationSvc, realtimeUsageTracker)
	registryHandler.SetWalletService(walletService)

	// Initialize privacy service and wire it into the registry handler
	privacyRepo := privacy.NewRepository(s.postgresDB)
	privacySalt := os.Getenv("PRIVACY_SALT")
	if privacySalt == "" {
		logrus.Fatal("FATAL: PRIVACY_SALT environment variable is required. Refusing to start with predictable salt.")
	}
	privacyService := privacy.NewService(privacyRepo, privacySalt)
	registryHandler.SetPrivacyService(privacyService)

	// Wire up billing controller for paid function execution using the unified wallet system
	registryHandler.SetBillingController(wallet.NewBillingControllerWrapper(walletService))

	// ---------------------------------------------------------------------------
	// RuntimeRouter + eager bundling wiring
	// ---------------------------------------------------------------------------
	// BundleService: compile source → WASM at publish time (eliminates cold-start).
	bundleSvc, _ := bundler.NewBundleService(os.Getenv("REDIS_ADDR"))
	if bundleSvc != nil {
		// Register Python WASM compiler for eager bundling.
		bundleSvc.RegisterCompiler("python3.11", &pythonWasmCompiler{})
		bundleSvc.RegisterCompiler("python3.12", &pythonWasmCompiler{})
		bundleSvc.RegisterCompiler("python3.13", &pythonWasmCompiler{})
		registryHandler.SetBundleService(bundleSvc)
	}

	// ---------------------------------------------------------------------------
	// WASM instance pool: pre-warms MicroPython runtime cells for fast Python exec.
	// ---------------------------------------------------------------------------
	micropythonPath := "internal/bundler/python/micropython.wasm"
	if _, err := os.Stat(micropythonPath); os.IsNotExist(err) {
		micropythonPath = "bundler/python/micropython.wasm"
	}
	var wasmPool *wasm.InstancePool
	if _, err := os.Stat(micropythonPath); err == nil {
		factory := func() (*wasm.PythonRuntime, error) {
			return wasm.NewPythonRuntime(micropythonPath, nil, nil, nil)
		}
		poolSize := 4
		if ps := os.Getenv("WASM_POOL_SIZE"); ps != "" {
			if n, _ := strconv.Atoi(ps); n > 0 {
				poolSize = n
			}
		}
		wasm.InitPoolsWithConfig(factory, poolSize, 30*time.Minute)
		wasmPool = wasm.PerTenantPools
		logrus.WithField("pool_size", poolSize).Info("WASM instance pool initialized")
	} else {
		logrus.Warn("MicroPython WASM not found, skipping instance pool initialization")
	}

	// RuntimeRouter: selects engine by runtime + tier, with fallback to legacy sandbox.
	runtimeRouter := registryexecution.BuildRuntimeRouter(wasmPool, cacheService, bundleSvc, micropythonPath)
	registryHandler.SetRuntimeRouter(runtimeRouter)
	// ---------------------------------------------------------------------------
	// ---------------------------------------------------------------------------

	// Initialize privacy handler for API routes
	privacyHandler := privacyhandler.NewHandler(privacyService, s.authSvc)

	adminRegistryHandler := admin.NewRegistryHandler(registryRepo, cacheService)
	oversightHandler := admin.NewOversightHandler(registryRepo, s.postgresDB.GORM, nil)
	docsHandler := registryhandler.NewDocumentationHandler(registryRepo)
	registryPlaygroundHandler := registryhandler.NewPlaygroundHandler(registryRepo)
	tutorialsHandler := registryhandler.NewTutorialsHandler()

	teamHandler := teams.NewHandler(s.repo, s.notificationSvc, nil)
	providersHandler := providers.NewHandler(s.repo, s.notificationSvc)
	dashboardHandler := dashboard.NewHandler(s.repo)
	enterpriseSLAHandler := enterprisePkg.NewSLAHandler(s.repo)
	decisionsRepo := decisionsrepo.NewRepository(s.postgresDB.GORM)
	decisionsHandler := decisions.NewHandler(decisionsRepo)

	stateRepo := staterepo.NewStateRepository(s.postgresDB.GORM)

	// Create function executor for internal function triggers
	functionExecutor := staterepo.NewFunctionTriggerExecutor(
		os.Getenv("FUNCTION_EXECUTION_URL"),
		os.Getenv("FUNCTION_API_KEY"),
		logrus.New(),
	)

	// Create webhook executor for external HTTP triggers
	webhookExecutor := staterepo.NewWebhookTriggerExecutor(logrus.New())

	// Combine into multi-executor that handles both types
	triggerExecutor := staterepo.NewMultiExecutor(functionExecutor, webhookExecutor, logrus.New())

	triggerEngineConfig := staterepo.DefaultTriggerEngineConfig()
	if envEnabled := os.Getenv("TRIGGER_ENGINE_ENABLED"); envEnabled != "" {
		triggerEngineConfig.Enabled = envEnabled == "true"
	}
	if envMaxRetries := os.Getenv("TRIGGER_ENGINE_MAX_RETRIES"); envMaxRetries != "" {
		if n, err := strconv.Atoi(envMaxRetries); err == nil && n > 0 {
			triggerEngineConfig.MaxRetries = n
		}
	}
	if envRateLimit := os.Getenv("TRIGGER_ENGINE_TENANT_RATE_LIMIT"); envRateLimit != "" {
		if n, err := strconv.Atoi(envRateLimit); err == nil && n > 0 {
			triggerEngineConfig.TenantRateLimit = n
		}
	}
	triggerEngine := staterepo.NewTriggerEngine(
		s.postgresDB.GORM,
		triggerEngineConfig,
		triggerExecutor,
		logrus.New(),
	)
	s.triggerEngine = triggerEngine

	stateHandler := state.NewHandlerWithTriggerEngine(stateRepo, triggerEngine).
		WithUserTenantResolver(state.RepoUserTenantResolver(s.repo))

	memoryRepo := state.NewAgentMemoryRepository(s.postgresDB.GORM)
	memoryHandler := state.NewAgentMemoryHandler(memoryRepo)
	agentMemoryHandler := agentmemoryhandler.NewHandler(s.postgresDB.GORM)

	stateFabricRepo := statefabricrepo.NewRepository(s.postgresDB.GORM)
	stateFabricRepo.ConfigureExecution(
		os.Getenv("FUNCTION_EXECUTION_URL"),
		os.Getenv("FUNCTION_API_KEY"),
	)
	stateFabricHandler := statefabric.NewHandlerWithCleanup(stateFabricRepo, sfAddonRepo, s.stateFabricCleanup)

	vaultRepo := vaultstorage.NewRepository(s.postgresDB.GORM)
	s.vaultRepo = vaultRepo
	vaultHandler := vault.NewHandler(vaultRepo, logrus.New())

	// Support handler initialization
	supportRepo := support.NewPostgresRepository(s.postgresDB.DB)
	supportLogger := logrus.New()

	// Initialize AI client for support
	aiSupportConfig := &support.AIChatClientConfig{
		BaseURL: os.Getenv("AI_SERVICE_URL"),
		APIKey:  os.Getenv("AI_SERVICE_API_KEY"),
		Timeout: 30 * time.Second,
		Model:   os.Getenv("AI_SUPPORT_MODEL"),
		Enabled: os.Getenv("AI_SUPPORT_ENABLED") != "false",
	}
	if aiSupportConfig.BaseURL == "" {
		aiSupportConfig.BaseURL = "http://localhost:18081"
	}
	if aiSupportConfig.Model == "" {
		aiSupportConfig.Model = "gpt-4o-mini"
	}
	aiSupportClient := support.NewAIServiceClient(aiSupportConfig, supportLogger)

	supportService := support.NewService(supportRepo, aiSupportClient, nil, supportLogger)
	supportHdlr := supportHandler.NewHandler(supportService, supportLogger)
	supportAdminHdlr := supportHandler.NewAdminHandler(supportRepo, supportLogger)

	// Initialize support WebSocket hub for real-time chat
	supportWSHub := supportHandler.NewWebSocketHub(supportService, nil, s.authSvc, supportLogger)
	go supportWSHub.Run()

	// Chat handler initialization
	chatRepo := chat.NewRepository(s.postgresDB.GORM)
	aiChatBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiChatBaseURL == "" {
		aiChatBaseURL = "http://localhost:18081"
	}
	aiChatAPIKey := os.Getenv("AI_SERVICE_API_KEY")
	chatAIClient := chat.NewAIServiceClient(aiChatBaseURL, aiChatAPIKey, logrus.New())
	chatConnectorRegistry := chat.NewConnectorRegistry(logrus.New())
	chatService := chat.NewService(chatRepo, chatAIClient, chatConnectorRegistry, logrus.New())
	chatWSHub := chat.NewWebSocketHub(chatService, chatRepo, chatAIClient, logrus.New())
	go chatWSHub.Run()
	chatHandler := chat.NewHandler(chatRepo, chatService, chatWSHub, logrus.New())
	chatConnectorHandler := chat.NewConnectorHandler(chatRepo, chatConnectorRegistry, logrus.New())

	aepHandler := agenthandler.NewHandler(s.postgresDB.GORM, s.redisClient, registryRepo, s.repo, s.notificationSvc)

	agentIdentityRepo := identity.NewRepository(s.postgresDB.GORM)
	agentGraphSvc := graph.NewService(s.postgresDB.GORM)
	agentActuatorSvc := actuator.NewService(s.postgresDB.GORM, agentGraphSvc)
	agentEconomyService := economy.NewService(s.postgresDB.GORM)
	agentMarketplaceService := marketplace.NewService(s.postgresDB.GORM)
	agentAutonomyService := autonomy.NewService(s.postgresDB.GORM)
	agentEvolutionService := evolution.NewService(s.postgresDB.GORM, agentGraphSvc, agentActuatorSvc)
	agentSwarmMessageService := swarm.NewMessageService(s.postgresDB.GORM, s.redisClient)
	agentSwarmService := swarm.NewService(s.postgresDB.GORM, agentIdentityRepo, agentEconomyService)

	prometheusURL := os.Getenv("PROMETHEUS_URL")
	// When PROMETHEUS_URL is not set, status queries will return ErrPrometheusNotConfigured
	// instead of attempting to connect to localhost (which causes connection errors in production)
	statusRepo := statushandler.NewRepository(s.postgresDB.DB)
	statusRepo.SetGormDB(s.postgresDB.GORM)
	statusHandlerInst := statushandler.NewHandler(statusRepo, prometheusURL, s.authSvc)

	factoryConfig := factorysvc.DefaultConfig("factory-agent")

	// Determine GitHub scanner type based on configuration
	var githubScanner discovery.Source
	if os.Getenv("GITHUB_TOKEN") != "" && (os.Getenv("GITHUB_OWNER") == "" || os.Getenv("GITHUB_REPO") == "") {
		// Global search mode: token provided but no specific repo
		logrus.Info("GitHub scanner: using global search mode (no specific repo configured)")
		githubScanner = discovery.NewGitHubGlobalScanner(os.Getenv("GITHUB_TOKEN"), nil, 100)
	} else {
		// Specific repo mode (or no token - will return empty)
		githubScanner = discovery.NewGitHubScanner()
	}

	factorySources := []discovery.Source{
		discovery.NewRedditScanner(),
		githubScanner,
		discovery.NewStackOverflowScanner(),
		discovery.NewGoogleScanner(),
	}
	factoryDiscovery := discovery.NewService(s.postgresDB.GORM, factorySources...)
	factoryGeneration := initializeGenerationServiceWithCache(s.postgresDB.GORM, s.redisClient)
	if detExec := registryexecution.NewRegistryDeterminismExecutorWithSandbox(registryRepo); detExec != nil {
		factoryGeneration.SetDeterminismExecutor(detExec)
	}
	// factory_test_results schema is managed by migration 20260428182910_create_factory_test_results.up.sql

	// Initialize secure sandbox executor for factory testing
	var factorySandbox agenttesting.SandboxExecutor
	secureSandbox, err := secureSandbox.NewSecureSandboxExecutor()
	if err != nil {
		logrus.WithError(err).Warn("failed to create secure sandbox, falling back to heuristic")
		factorySandbox = nil
	} else {
		factorySandbox = secureSandbox
		logrus.Infof("secure sandbox executor initialized (gVisor available: %v)", secureSandbox.IsGvisorAvailable())
	}

	factoryTesting := agenttesting.NewService(s.postgresDB.GORM, factorySandbox, nil)
	factoryPublisher := agentdeployment.NewPublisher(s.postgresDB.GORM)
	factoryService := factorysvc.NewService(s.postgresDB.GORM, factoryConfig, factoryDiscovery, factoryGeneration, factoryTesting, factoryPublisher)

	loadedFactoryConfig, err := factoryService.GetConfig(context.Background())
	if err != nil {
		logrus.WithError(err).Warn("failed to load factory config from database, using defaults")
	} else {
		factoryConfig = loadedFactoryConfig
		logrus.Info("loaded factory config from database")
	}

	factoryDiscoveryWithThreshold := discovery.NewServiceWithThreshold(s.postgresDB.GORM, factoryConfig.MinimumQualityScore, factorySources...)

	factoryPipelineScheduler := scheduler.NewFactoryPipelineScheduler(factoryService)
	scheduleConfig := scheduler.FactoryScheduleConfig{
		Enabled:  factoryConfig.ScheduleEnabled,
		Cron:     factoryConfig.ScheduleCron,
		Timezone: factoryConfig.ScheduleTimezone,
	}
	if err := factoryPipelineScheduler.Start(context.Background(), scheduleConfig); err != nil {
		logrus.WithError(err).Error("failed to start factory pipeline scheduler")
	} else if factoryConfig.ScheduleEnabled {
		logrus.Infof("factory pipeline scheduler started with cron: %s", factoryConfig.ScheduleCron)
	}

	factoryService.UpdateDiscoveryService(factoryDiscoveryWithThreshold)

	factoryHandler := factoryhandler.NewHandler(s.postgresDB.GORM, factoryService, factoryDiscovery, factoryPublisher, &factoryConfig, factoryPipelineScheduler)

	platformController := swarm.NewPlatformController(
		s.postgresDB.GORM,
		agentSwarmService,
		agentIdentityRepo,
		agentSwarmMessageService,
		factoryDiscovery,
	)
	metricsCollector := swarm.NewMetricsCollector(s.postgresDB.GORM, factoryService, factoryDiscovery)
	workerService := swarm.NewWorkerService(
		s.postgresDB.GORM,
		agentSwarmMessageService,
		factoryDiscovery,
		factoryService,
		agentIdentityRepo,
	)
	swarmControllerHandler := agenthandler.NewSwarmControllerHandler(platformController, metricsCollector, workerService)
	if err := platformController.Initialize(context.Background()); err != nil {
		logrus.WithError(err).Warn("failed to initialize platform controller")
	}

	unfairAdvantageEngine := swarm.NewUnfairAdvantageEngine(s.postgresDB.GORM, platformController, metricsCollector)
	if err := unfairAdvantageEngine.Initialize(context.Background()); err != nil {
		logrus.WithError(err).Warn("failed to initialize unfair advantage engine")
	}
	unfairAdvantageHandler := agenthandler.NewUnfairAdvantageHandler(unfairAdvantageEngine)

	unfairAdvantageScheduler := scheduler.NewUnfairAdvantageScheduler(unfairAdvantageEngine)
	unfairAdvantageSchedulerEnabled := os.Getenv("UNFAIR_ADVANTAGE_SCHEDULER_ENABLED") == "true"
	unfairAdvantageSchedulerCron := os.Getenv("UNFAIR_ADVANTAGE_SCHEDULER_CRON")
	if unfairAdvantageSchedulerCron == "" {
		unfairAdvantageSchedulerCron = "0 2 * * *"
	}
	unfairAdvantageSchedulerConfig := scheduler.UnfairAdvantageScheduleConfig{
		Enabled:  unfairAdvantageSchedulerEnabled,
		Cron:     unfairAdvantageSchedulerCron,
		Timezone: "UTC",
	}
	if err := unfairAdvantageScheduler.Start(context.Background(), unfairAdvantageSchedulerConfig); err != nil {
		logrus.WithError(err).Error("failed to start unfair advantage scheduler")
	} else if unfairAdvantageSchedulerEnabled {
		logrus.Infof("unfair advantage scheduler started with cron: %s", unfairAdvantageSchedulerCron)
	}

	if err := workerService.Start(context.Background()); err != nil {
		logrus.WithError(err).Warn("failed to start swarm worker service")
	}

	// Trust Score Scheduler - hourly trust score recalculation
	trustScoreScheduler := scheduler.NewTrustScoreScheduler(registryRepo)
	trustScoreEnabled := os.Getenv("TRUST_SCORE_SCHEDULER_ENABLED") == "true"
	trustScoreCron := os.Getenv("TRUST_SCORE_SCHEDULER_CRON")
	if trustScoreCron == "" {
		trustScoreCron = "0 * * * *" // Default: every hour at minute 0
	}
	trustScoreConfig := scheduler.TrustScoreScheduleConfig{
		Enabled: trustScoreEnabled,
		Cron:    trustScoreCron,
	}
	if err := trustScoreScheduler.Start(context.Background(), trustScoreConfig); err != nil {
		logrus.WithError(err).Error("failed to start trust score scheduler")
	} else if trustScoreEnabled {
		logrus.Infof("trust score scheduler started with cron: %s", trustScoreCron)
	}

	// Expired Evaluation Scheduler - cleans up old cached trust policy evaluations
	trustapiRevocationRepo := trustapirepo.NewRevocationRepository(s.postgresDB.GORM)
	expiredEvalScheduler := scheduler.NewExpiredEvaluationScheduler(trustapiRevocationRepo)
	expiredEvalEnabled := os.Getenv("EXPIRED_EVAL_SCHEDULER_ENABLED") != "false" // Default: enabled
	expiredEvalCron := os.Getenv("EXPIRED_EVAL_SCHEDULER_CRON")
	if expiredEvalCron == "" {
		expiredEvalCron = "0 */6 * * *" // Default: every 6 hours
	}
	expiredEvalMaxAgeHours := 24
	if maxAgeHours := os.Getenv("EXPIRED_EVAL_MAX_AGE_HOURS"); maxAgeHours != "" {
		if h, err := strconv.Atoi(maxAgeHours); err == nil {
			expiredEvalMaxAgeHours = h
		}
	}
	expiredEvalScheduler.CronExpression = expiredEvalCron
	expiredEvalScheduler.Enabled = expiredEvalEnabled
	expiredEvalScheduler.MaxAge = time.Duration(expiredEvalMaxAgeHours) * time.Hour
	if err := expiredEvalScheduler.Start(context.Background()); err != nil {
		logrus.WithError(err).Error("failed to start expired evaluation scheduler")
	} else if expiredEvalEnabled {
		logrus.Infof("expired evaluation scheduler started with cron: %s, max_age: %dh", expiredEvalCron, expiredEvalMaxAgeHours)
	}

	// Subscription Sync Scheduler - syncs Stripe subscription status periodically
	subscriptionSyncScheduler := scheduler.NewSubscriptionSyncScheduler(s.repo, nil)       // notification service will be injected later
	subscriptionSyncEnabled := os.Getenv("SUBSCRIPTION_SYNC_SCHEDULER_ENABLED") != "false" // Default: enabled
	subscriptionSyncCron := os.Getenv("SUBSCRIPTION_SYNC_SCHEDULER_CRON")
	if subscriptionSyncCron == "" {
		subscriptionSyncCron = "0 */6 * * *" // Default: every 6 hours
	}
	subscriptionSyncConfig := scheduler.SubscriptionSyncScheduleConfig{
		Cron: subscriptionSyncCron,
	}
	if subscriptionSyncEnabled {
		if err := subscriptionSyncScheduler.Start(context.Background(), subscriptionSyncConfig); err != nil {
			logrus.WithError(err).Error("failed to start subscription sync scheduler")
		} else {
			logrus.Infof("subscription sync scheduler started with cron: %s", subscriptionSyncCron)
		}
	}

	// Upcoming Renewal Scheduler - sends notifications about upcoming subscription renewals
	upcomingRenewalScheduler := scheduler.NewUpcomingRenewalScheduler(s.repo, s.notificationSvc)
	upcomingRenewalEnabled := os.Getenv("UPCOMING_RENEWAL_SCHEDULER_ENABLED") != "false" // Default: enabled
	upcomingRenewalCron := os.Getenv("UPCOMING_RENEWAL_SCHEDULER_CRON")
	if upcomingRenewalCron == "" {
		upcomingRenewalCron = "0 9 * * *" // Default: daily at 9 AM
	}
	upcomingRenewalConfig := scheduler.UpcomingRenewalConfig{
		Cron: upcomingRenewalCron,
	}
	if upcomingRenewalEnabled {
		if err := upcomingRenewalScheduler.Start(context.Background(), upcomingRenewalConfig); err != nil {
			logrus.WithError(err).Error("failed to start upcoming renewal scheduler")
		} else {
			logrus.Infof("upcoming renewal scheduler started with cron: %s", upcomingRenewalCron)
		}
	}

	// Exchange Rate Scheduler - syncs exchange rates from external providers
	exchangeRateScheduler := currency.NewExchangeRateScheduler(s.repo, s.redisClient)
	if err := exchangeRateScheduler.Start(context.Background()); err != nil {
		logrus.WithError(err).Error("failed to start exchange rate scheduler")
	}

	experimentService := factorysvc.NewExperimentService(s.postgresDB.GORM)
	experimentAdapter := factorysvc.NewGenerationExperimentAdapter(s.postgresDB.GORM, experimentService)
	experimentHandler := factoryhandler.NewExperimentHandler(s.postgresDB.GORM, experimentService, experimentAdapter)

	categorizationSvc := categorization.NewService(s.postgresDB.GORM)
	categorizationHandler := categorizationhandler.NewHandler(s.postgresDB.GORM, categorizationSvc)

	analyticsSvc := analytics.NewService(s.postgresDB.GORM, analytics.DefaultServiceConfig(factoryConfig.AgentID))
	analyticsHandler := analyticshandler.NewHandler(analyticsSvc, s.authSvc)

	// Initialize learning and deployment services for swarm
	agentLearningRepo := learning.NewRepository(s.postgresDB.GORM)
	agentAnalyzer := learning.NewAnalyzer(s.postgresDB.GORM)
	agentOptimizer := learning.NewOptimizer(learning.OptimizerDeps{
		DB: s.postgresDB.GORM,
	})
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	agentGenerator := agentdeployment.NewGenerator(s.postgresDB.GORM, openRouterAPIKey)
	agentPublisher := agentdeployment.NewPublisher(s.postgresDB.GORM)
	agentSecurityService := agentsecurity.NewSwarmSecurityService(s.postgresDB.GORM)

	swarmHandler := agenthandler.NewSwarmHandler(
		agentSwarmService,
		agentSwarmMessageService,
		agentEconomyService,
		storage.NewFinancialTransactionRepository(s.postgresDB.GORM),
		agentMarketplaceService,
		agentEvolutionService,
		agentAutonomyService,
		agentIdentityRepo,
		agentLearningRepo,
		agentAnalyzer,
		agentOptimizer,
		agentGenerator,
		agentPublisher,
		agentSecurityService,
	)

	sebgHandler := agenthandler.NewSEBGHandler(s.postgresDB.GORM, agentIdentityRepo)
	evolutionHandler := agenthandler.NewEvolutionHandler(s.postgresDB.GORM, agentEvolutionService, agentIdentityRepo)
	optimizationHandler := agenthandler.NewOptimizationHandler(s.postgresDB.GORM)
	daemonHandler := agenthandler.NewDaemonHandler(s.postgresDB.GORM)

	recommendationHandler := recommendations.NewHandler(s.recommendationSvc)

	// ── Middleware initialization ─────────────────────────────────────────────
	authMiddleware := middleware.NewAuthMiddleware(s.authSvc)
	advancedSecurityMiddleware := middleware.NewAdvancedSecurityMiddleware(s.repo)

	captchaService := captcha.NewCaptchaService(nil)
	if recaptchaV2SiteKey := os.Getenv("RECAPTCHA_V2_SITE_KEY"); recaptchaV2SiteKey != "" {
		recaptchaV2Secret := os.Getenv("RECAPTCHA_V2_SECRET_KEY")
		captchaService.RegisterProvider(captcha.NewRecaptchaV2Provider(recaptchaV2SiteKey, recaptchaV2Secret, nil))
	}
	if recaptchaV3SiteKey := os.Getenv("RECAPTCHA_V3_SITE_KEY"); recaptchaV3SiteKey != "" {
		recaptchaV3Secret := os.Getenv("RECAPTCHA_V3_SECRET_KEY")
		captchaService.RegisterProvider(captcha.NewRecaptchaV3Provider(recaptchaV3SiteKey, recaptchaV3Secret, nil))
	}
	if hcaptchaSiteKey := os.Getenv("HCAPTCHA_SITE_KEY"); hcaptchaSiteKey != "" {
		hcaptchaSecret := os.Getenv("HCAPTCHA_SECRET_KEY")
		captchaService.RegisterProvider(captcha.NewHCaptchaProvider(hcaptchaSiteKey, hcaptchaSecret, nil))
	}
	if turnstileSiteKey := os.Getenv("TURNSTILE_SITE_KEY"); turnstileSiteKey != "" {
		turnstileSecret := os.Getenv("TURNSTILE_SECRET_KEY")
		captchaService.RegisterProvider(captcha.NewTurnstileProvider(turnstileSiteKey, turnstileSecret, nil))
	}

	executionSecurityMW := middleware.NewExecutionCoordinatorMiddleware(s.postgresDB.GORM, nil, captchaService)

	clamAVURL := os.Getenv("CLAMAV_URL")
	if clamAVURL == "" {
		clamAVURL = "http://clamav:3310"
	}
	yaraURL := os.Getenv("YARA_URL")
	if yaraURL == "" {
		yaraURL = "http://yara:8080"
	}
	minimumTrustLevel := os.Getenv("MINIMUM_TRUST_LEVEL")
	if minimumTrustLevel == "" {
		minimumTrustLevel = "standard"
	}
	verificationEnabled := os.Getenv("VERIFICATION_ENABLED") == "true"

	var verificationMiddleware *middleware.VerificationMiddleware
	if verificationEnabled {
		verificationRegistryRepo := registry.NewRegistryRepository(s.postgresDB.GORM, nil)
		verificationMiddleware = middleware.NewVerificationMiddleware(verificationRegistryRepo, clamAVURL, yaraURL, minimumTrustLevel)
	}

	// ── Global middleware wiring ──────────────────────────────────────────────
	s.router.Use(middleware.RecoveryMiddleware)
	s.router.Use(middleware.TracingMiddleware)
	s.router.Use(maintenanceMiddleware.CheckMaintenanceMode)
	s.router.Use(middleware.EnvironmentMiddleware)

	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(advancedSecurityMiddleware.CORSMiddleware(http.HandlerFunc(next.ServeHTTP)))
	})

	// SECURITY FIX: Require explicit PRODUCTION_ENV=true for security middleware
	// Previously, security was disabled just by setting DEVELOPMENT=true
	// Now we require PRODUCTION_ENV=true to enable security features
	isDev := os.Getenv("DEVELOPMENT") == "true"
	productionEnv := os.Getenv("PRODUCTION_ENV") == "true"

	if !isDev || productionEnv {
		// Apply advanced security middleware when not in dev mode OR when explicitly in production mode
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.SecurityHeaders(http.HandlerFunc(next.ServeHTTP)))
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.GeoBlocking(http.HandlerFunc(next.ServeHTTP)))
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.DDoSProtection(http.HandlerFunc(next.ServeHTTP)))
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.AdvancedInputValidation(http.HandlerFunc(next.ServeHTTP)))
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.AdvancedRateLimit(http.HandlerFunc(next.ServeHTTP)))
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.TrafficManagement(http.HandlerFunc(next.ServeHTTP)))
		})
	} else {
		// In development mode without PRODUCTION_ENV, only apply basic security headers
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.SecurityHeaders(http.HandlerFunc(next.ServeHTTP)))
		})
		// Log warning about missing security middleware in development
		logrus.Warn("SECURITY WARNING: Running in DEVELOPMENT mode without PRODUCTION_ENV=true. Advanced security middleware (DDoS protection, geo-blocking, rate limiting, input validation) is DISABLED. This is NOT safe for production!")
	}

	s.router.Use(monitoringPkg.HTTPMetricsMiddleware)

	// ── Subrouters & rate limiters ────────────────────────────────────────────
	api := s.router.PathPrefix("/v1").Subrouter()
	apiV2 := s.router.PathPrefix("/v2").Subrouter()
	protected := api.PathPrefix("").Subrouter()

	// Online presence: optional JWT on any /v1 and /v2 call sets user context; then we bump last_active_at.
	userRepoActivity := storage.NewUserRepository(s.postgresDB)
	api.Use(authMiddleware.OptionalAuth)
	api.Use(middleware.SimpleActivityTrackerWithRedis(userRepoActivity, s.redisClient))
	apiV2.Use(authMiddleware.OptionalAuth)
	apiV2.Use(middleware.SimpleActivityTrackerWithRedis(userRepoActivity, s.redisClient))

	authRateLimiter := middleware.NewAuthRateLimiter()
	vaultRateLimiter := middleware.NewVaultRateLimiter()
	providerRateLimiter := middleware.NewProviderRateLimiter()
	walletRateLimiter := middleware.NewWalletRateLimiter()
	mfaRateLimiter := middleware.NewMFARateLimiter()

	// Initialize CSRF middleware early for billing route protection
	csrfMiddleware := middleware.NewCSRFMiddleware(s.upstashRedis, s.authSvc)

	// Status WebSocket hub — must be wired before route registration
	statusWSHub := statushandler.NewStatusWebSocketHub(statusHandlerInst, logrus.New())
	go statusWSHub.Run()
	statusHandlerInst.SetStatusHub(statusWSHub)

	// Public newsletter routes
	newsletterHandler.RegisterRoutes(api)

	// ── GitHub Integration ─────────────────────────────────────────────────────
	registerGitHubRoutes(api, authMiddleware, githubHandler)

	// ── Domain-scoped route registration ─────────────────────────────────────
	registerAuthRoutes(
		s.router, api,
		authRateLimiter, walletRateLimiter, mfaRateLimiter, authMiddleware, csrfMiddleware,
		authHandler, tenantAuthHandler, apiKeyAuthHandler, usersHandler, favoritesHandler,
		followHandler, apiKeysHandler, billingHandler, tenantWebhookHandler,
		usageHandler, forecastHandler, costAllocationHandler,
		exportHandler, externalBillingHandler,
		mfaHandler, notificationHandler, notificationWSHandler,
		presenceHandler,
	)

	registerRegistryRoutes(
		s, api, apiV2,
		authMiddleware, executionSecurityMW, verificationMiddleware,
		registryRepo, registryHandler, registryPlaygroundHandler,
		appPlaygroundHandler, docsHandler, tutorialsHandler,
		versionHandler, blogHandler, contentHandler, feedbackHandler, recommendationHandler,
	)

	// ── FRG (Function Registry + Live Runtime Graph) ─────────────────────────
	// Register on root router for /frg/* paths (not /v1/frg/*)
	registerFRGRoutesOnRoot(s, s.router, authMiddleware, executionSecurityMW, registryRepo, advancedSecurityMiddleware, realtimeUsageTracker)

	// Also register on /v1 for compatibility
	registerFRGRoutes(
		s, api,
		authMiddleware, executionSecurityMW,
		registryRepo, advancedSecurityMiddleware,
		realtimeUsageTracker,
	)

	if err := platformController.Initialize(context.Background()); err != nil {
		logrus.WithError(err).Warn("failed to initialize platform controller")
	}

	registerPlatformRoutes(
		s, api, protected,
		authMiddleware, advancedSecurityMiddleware, vaultRateLimiter, providerRateLimiter,
		monitoringHandler, securityHandler, statusHandlerInst,
		factoryHandler, experimentHandler, categorizationHandler,
		analyticsHandler, unifiedAnalyticsSvc,
		stateHandler, stateFabricHandler, vaultHandler,
		memoryHandler, agentMemoryHandler,
		dashboardHandler, enterpriseSLAHandler,
		teamHandler, providersHandler,
		appsHandler, functionsHandler,
		backendsHandler, deploymentsHandler,
		versionHandler, maintenanceHandler,
		supportHdlr, supportAdminHdlr, supportWSHub,
		decisionsHandler,
		stateUsageHandler,
		deployKeysHandler,
		functionWebhooksHandler,
		swarmControllerHandler,
		unfairAdvantageHandler,
		chatHandler, chatConnectorHandler, chatWSHub,
	)

	agentRateLimiter := middleware.NewAgentRateLimiter(s.redisClient)

	registerAgentRoutes(
		s, api, protected,
		authMiddleware,
		agentRateLimiter,
		aepHandler, swarmHandler, sebgHandler, evolutionHandler, daemonHandler,
		registryRepo, cacheService,
		platformFeeRepo,
		billingOperationalRepo,
	)

	// Initialize admin security middleware (csrfMiddleware already initialized above for billing)
	adminRateLimiter := middleware.NewAdminRateLimiter(s.redisClient)
	adminSessionMiddleware := middleware.NewAdminSessionMiddleware(s.postgresDB.DB, s.authSvc)
	ipAllowlistMiddleware := middleware.NewIPAllowlistMiddleware(s.postgresDB.DB, s.redisClient)
	adminIPAllowlistHandler := admin.NewAdminIPAllowlistHandler(s.postgresDB.DB, ipAllowlistMiddleware)
	adminAuditHandler := admin.NewAdminAuditHandler(s.postgresDB.DB)
	securityEventHandler := admin.NewSecurityEventHandler(s.postgresDB.DB)
	alertHandler := admin.NewAlertHandler(s.postgresDB.DB)
	adminNewsletterHandler := admin.NewNewsletterHandler(s.repo, s.emailSvc)

	// Initialize retention handler with cleanup service for admin management
	retentionHandler := admin.NewRetentionHandler(s.postgresDB, s.executionLogCleanup)

	registerAdminRoutes(
		s, api, authMiddleware, advancedSecurityMiddleware,
		adminHandler, adminBackendsHandler, adminProvidersHandler,
		maintenanceHandler, feedbackHandler, monitoringHandler,
		securityHandler, mfaHandler, adminRegistryHandler,
		registryHandler, oversightHandler, factoryHandler,
		stateFabricHandler, blogHandler, contentHandler,
		csrfMiddleware, adminRateLimiter, adminSessionMiddleware,
		ipAllowlistMiddleware, adminIPAllowlistHandler, adminAuditHandler, securityEventHandler, alertHandler,
		adminNewsletterHandler, usageHandler, costAllocationHandler,
		retentionHandler, disputesHandler,
		stateUsageHandler,
		unfairAdvantageHandler,
	)

	// Trust API for external platform partners
	registerTrustAPIRoutes(s, api, registryRepo)

	// Privacy API routes (GDPR compliance, data export/deletion, consent management)
	registerPrivacyRoutes(api, authMiddleware, privacyHandler)

	// ── Time Machine ──────────────────────────────────────────────────────
	tmRepo := timemachine.NewRepository(s.postgresDB.GORM)
	tmHandler := newTimeMachineHandler(tmRepo, s.repo, s.redisClient, realtimeUsageTracker, s.notificationSvc, s.authSvc)
	registerTimeMachineRoutes(api, tmHandler, authMiddleware)

	// ── Payout System (Stripe Connect) ───────────────────────────────────────
	payoutRepo := storage.NewPayoutRepository(s.postgresDB.DB)
	payoutServiceExtended := paymentPkg.NewPayoutServiceExtended(payoutRepo, s.notificationSvc)
	payoutExtendedHandler := payoutsHandler.NewExtendedHandler(payoutServiceExtended, payoutRepo, s.repo)
	payoutLegacyHandler := payoutsHandler.NewHandler(payoutServiceExtended.PayoutService, payoutRepo, s.repo)

	// Payout scheduler for auto-payouts
	payoutSchedulerConfig := scheduler.EnvPayoutScheduleConfig()
	payoutScheduler := scheduler.NewPayoutScheduler(payoutServiceExtended)
	if err := payoutScheduler.Start(context.Background(), payoutSchedulerConfig); err != nil {
		logrus.WithError(err).Error("failed to start payout scheduler")
	}

	// Wire payout service to Stripe webhook handler (registered later in registerAgentRoutes)
	// Store on server for access by agent routes
	s.payoutWebhookProcessor = payoutServiceExtended

	// ── Payout Routes (user-facing, protected) ──────────────────────────────
	// Account management
	api.HandleFunc("/payouts/connect-account", authMiddleware.RequireAuth(payoutLegacyHandler.HandleGetConnectAccount)).Methods("GET", "OPTIONS")
	api.HandleFunc("/payouts/connect-account", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(payoutLegacyHandler.HandleStartOnboarding))).Methods("POST", "OPTIONS")
	api.HandleFunc("/payouts/connect-account/refresh", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(payoutLegacyHandler.HandleRefreshConnectAccount))).Methods("POST", "OPTIONS")

	// Balance and ledger
	api.HandleFunc("/payouts/balance", authMiddleware.RequireAuth(payoutLegacyHandler.HandleGetBalance)).Methods("GET", "OPTIONS")
	api.HandleFunc("/payouts/ledger", authMiddleware.RequireAuth(payoutLegacyHandler.HandleListLedgerEntries)).Methods("GET", "OPTIONS")

	// Payout requests with fee calculation
	api.HandleFunc("/payouts/request", authMiddleware.RequireAuth(walletRateLimiter.LimitTopUp(csrfMiddleware.RequireCSRF(payoutExtendedHandler.HandleRequestPayoutWithFees)))).Methods("POST", "OPTIONS")
	api.HandleFunc("/payouts/{id}/cancel", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(payoutExtendedHandler.HandleCancelPayout))).Methods("POST", "OPTIONS")

	// Payout history (enhanced)
	api.HandleFunc("/payouts/requests", authMiddleware.RequireAuth(payoutLegacyHandler.HandleListPayoutRequests)).Methods("GET", "OPTIONS")
	api.HandleFunc("/payouts/history", authMiddleware.RequireAuth(payoutExtendedHandler.HandleGetPayoutHistory)).Methods("GET", "OPTIONS")

	// Schedule preferences
	api.HandleFunc("/payouts/schedule", authMiddleware.RequireAuth(payoutExtendedHandler.HandleGetSchedulePreference)).Methods("GET", "OPTIONS")
	api.HandleFunc("/payouts/schedule", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(payoutExtendedHandler.HandleUpdateSchedulePreference))).Methods("PUT", "OPTIONS")

	// Suppress unused variable warnings (admin routes registered in routes_admin.go)
	_ = payoutExtendedHandler
	_ = payoutLegacyHandler

	// ── Runtime Optimization Endpoint (receives suggestions from Rust GraphOptimizer) ──
	api.HandleFunc("/optimizations", optimizationHandler.ReceiveOptimizationSuggestion).Methods("POST", "OPTIONS")

	// ── AI Service Proxy (for AI Composer + Gallery features) ───────────────
	aiProxyHandler := NewAIProxyHandler()

	// AI Composer routes - paths are relative to /v1 subrouter
	protected.HandleFunc("/ai/composer/generate", authMiddleware.RequireAuth(aiProxyHandler.HandleGenerateFunction)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/ai/composer/generate/stream", authMiddleware.RequireAuth(aiProxyHandler.HandleGenerateFunctionStream)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/ai/composer/refine", authMiddleware.RequireAuth(aiProxyHandler.HandleRefineFunction)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/ai/composer/refine/stream", authMiddleware.RequireAuth(aiProxyHandler.HandleRefineFunctionStream)).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/health", aiProxyHandler.HandleHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/status", aiProxyHandler.HandleAIStatus).Methods("GET", "OPTIONS")

	// ── Infrastructure endpoints ──────────────────────────────────────────────
	wellknownHandler := wellknown.NewHandler(registryRepo)
	s.router.HandleFunc("/.well-known/functionfly.json", wellknownHandler.HandleWellKnown).Methods("GET", "OPTIONS")

	s.router.HandleFunc("/health", s.handleHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/healthz", s.handleHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/detailed", s.handleDetailedHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/check", s.handleHealthCheck).Methods("GET", "OPTIONS")
	s.router.Handle("/metrics", middleware.RequireAuthInProduction(authMiddleware)(promhttp.Handler())).Methods("GET")
	s.router.HandleFunc("/ws/v1/status", statusHandlerInst.HandleWebSocketStatus).Methods("GET")

	// ── API Documentation (Swagger/OpenAPI) ────────────────────────────────────
	// OpenAPI spec - JSON endpoint (for OpenAI compatibility and Swagger UI)
	s.router.Handle("/swagger/doc.json", middleware.RequireAuthInProduction(authMiddleware)(http.HandlerFunc(docs.ServeJSONSpec))).Methods("GET", "OPTIONS")
	s.router.Handle("/swagger/doc.yaml", middleware.RequireAuthInProduction(authMiddleware)(http.HandlerFunc(docs.ServeYAMLSpec))).Methods("GET", "OPTIONS")

	// Swagger UI - interactive API documentation
	s.router.Handle("/swagger", middleware.RequireAuthInProduction(authMiddleware)(http.HandlerFunc(docs.ServeSwaggerUI))).Methods("GET", "OPTIONS")
	s.router.Handle("/swagger/", middleware.RequireAuthInProduction(authMiddleware)(http.HandlerFunc(docs.ServeSwaggerUI))).Methods("GET", "OPTIONS")
	s.router.Handle("/swagger/index.html", middleware.RequireAuthInProduction(authMiddleware)(http.HandlerFunc(docs.ServeSwaggerUI))).Methods("GET", "OPTIONS")

	// ── SPA catch-all routes ──────────────────────────────────────────────────
	// Serve index.html for /fx/*, /run/*, /replay/* (playground SPA paths).
	// Only match GET/HEAD/OPTIONS so POST/PUT/PATCH/DELETE API calls are not
	// shadowed by the SPA catch-all.
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			return false
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") ||
			strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") ||
			r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 1 && (pathParts[0] == "fx" || pathParts[0] == "run" || pathParts[0] == "replay" || pathParts[0] == "functions" || pathParts[0] == "v1" || pathParts[0] == "v2")
	}).HandlerFunc(s.serveSPAIndex)

	// Public routing endpoint: /{appSlug}/*
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/content/") ||
			strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" || r.URL.Path == "/healthz" ||
			strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/auth/") ||
			r.URL.Path == "/waitlist" ||
			strings.HasPrefix(r.URL.Path, "/frg/") || r.URL.Path == "/frg" ||
			strings.HasPrefix(r.URL.Path, "/gx/") || r.URL.Path == "/gx" ||
			strings.HasPrefix(r.URL.Path, "/webhook/") {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 2 && pathParts[0] != "" && pathParts[0] != "api" &&
			pathParts[0] != "content" && pathParts[0] != "health" && pathParts[0] != "auth" &&
			pathParts[0] != "fx" && pathParts[0] != "run" && pathParts[0] != "replay" &&
			pathParts[0] != "frg" && pathParts[0] != "gx" && pathParts[0] != "functions" &&
			pathParts[0] != "v1" && pathParts[0] != "v2"
	}).HandlerFunc(s.handlePublicRoute)
}

// initializeGenerationServiceWithCache creates a generation service with Redis-backed
// cache when Redis is available. Falls back to in-memory cache otherwise.
func initializeGenerationServiceWithCache(db *gorm.DB, redisClient *redis.Client) *generation.Service {
	useRedisCache := os.Getenv("GENERATION_CACHE_REDIS_ENABLED") == "true"
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {
		return generation.NewService(db)
	}

	if !useRedisCache && redisClient != nil {
		useRedisCache = os.Getenv("GENERATION_CACHE_REDIS_ENABLED") != "false"
	}

	var codeGen generation.CodeGenerator

	if useRedisCache && redisClient != nil {
		logrus.Info("Initializing generation service with Redis-backed cache")
		codeGen = generation.NewOpenRouterClientWithRedis(apiKey, nil, redisClient, true, nil)
	} else {
		logrus.Info("Initializing generation service with in-memory cache")
		codeGen = generation.NewOpenRouterClient(apiKey, nil, nil, nil)
	}

	return generation.NewServiceWithGenerator(db, codeGen)
}

// pythonWasmCompiler implements bundler.RuntimeCompiler for Python → WASM.
// It writes source to a temp file and delegates to the existing bundler.
type pythonWasmCompiler struct{}

func (c *pythonWasmCompiler) Compile(ctx context.Context, sourceCode string, m *manifest.Manifest) (*bundler.CompiledBundle, error) {
	// Write source to a temp entry file so the bundler can read it.
	tmpDir, err := os.MkdirTemp("", "py-wasm-compile-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	entry := filepath.Join(tmpDir, "main.py")
	if err := os.WriteFile(entry, []byte(sourceCode), 0600); err != nil {
		return nil, fmt.Errorf("write source: %w", err)
	}

	bm := &manifest.Manifest{
		Name:      m.Name,
		Version:   m.Version,
		Runtime:   m.Runtime,
		TimeoutMS: m.TimeoutMS,
		MemoryMB:  m.MemoryMB,
		Entry:     entry,
	}

	wasmBytes, err := bundler.Bundle(bm)
	if err != nil {
		return nil, err
	}
	return &bundler.CompiledBundle{
		Bytes:   wasmBytes,
		Runtime: m.Runtime,
		IsValid: true,
	}, nil
}
