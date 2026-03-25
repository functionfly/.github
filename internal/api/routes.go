package api

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/categorization"
	agentdeployment "github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/evolution"
	factorysvc "github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/marketplace"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/functionfly/functionfly/internal/agent/testing"
	"github.com/functionfly/functionfly/internal/analytics"
	"github.com/functionfly/functionfly/internal/analytics/unified"
	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	agentmemoryhandler "github.com/functionfly/functionfly/internal/api/handlers/agent_memory"
	analyticshandler "github.com/functionfly/functionfly/internal/api/handlers/analytics"
	"github.com/functionfly/functionfly/internal/api/handlers/apikeys"
	"github.com/functionfly/functionfly/internal/api/handlers/apps"
	authHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/auth"
	"github.com/functionfly/functionfly/internal/api/handlers/backends"
	"github.com/functionfly/functionfly/internal/api/handlers/billing"
	categorizationhandler "github.com/functionfly/functionfly/internal/api/handlers/categorization"
	"github.com/functionfly/functionfly/internal/api/handlers/content"
	"github.com/functionfly/functionfly/internal/api/handlers/dashboard"
	"github.com/functionfly/functionfly/internal/api/handlers/deployments"
	"github.com/functionfly/functionfly/internal/api/handlers/enterprise"
	factoryhandler "github.com/functionfly/functionfly/internal/api/handlers/factory"
	feedbackHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/feedback"
	followHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/follow"
	"github.com/functionfly/functionfly/internal/api/handlers/functions"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	notificationHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/notifications"
	"github.com/functionfly/functionfly/internal/api/handlers/playground"
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
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/captcha"
	monitoringPkg "github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	statefabricrepo "github.com/functionfly/functionfly/internal/storage/statefabric"
	vaultstorage "github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/functionfly/functionfly/internal/support"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/gorilla/mux"
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
	usersHandler := usersHandlerPkg.NewHandler(s.repo, s.authSvc)
	platformFeeRepo := registry.NewPlatformFeeRepository(s.postgresDB.GORM)
	sfAddonRepo := statefabricaddons.NewRepository(s.postgresDB.GORM)
	billingHandler := billing.NewHandler(s.repo, platformFeeRepo, sfAddonRepo)
	appsHandler := apps.NewHandler(s.repo)
	backendsHandler := backends.NewHandler(s.repo, s.routingSvc)
	deploymentsHandler := deployments.NewHandler(s.repo, s.deploySvc)
	functionsHandler := functions.NewHandler(s.repo, s.deploySvc)
	unifiedAnalyticsSvc := unified.NewService(s.postgresDB.GORM, s.usageMetricsAgg)
	adminHandler := admin.NewHandler(s.repo, s.authSvc, unifiedAnalyticsSvc, sfAddonRepo)
	adminBackendsHandler := admin.NewBackendsHandler(s.repo, s.authSvc)
	adminProvidersHandler := admin.NewProvidersHandler(s.repo, s.authSvc)
	securityHandler := security.NewHandler(s.repo, s.authSvc)

	maintenanceRepo := storage.NewMaintenanceRepository(s.postgresDB.GORM)
	maintenanceHandler := admin.NewMaintenanceHandler(maintenanceRepo, s.authSvc)
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(maintenanceRepo)
	contentHandler := content.NewHandler(s.repo)
	feedbackHandler := feedbackHandlerPkg.NewHandler(s.repo, s.storageService)

	followService := services.NewFollowService(s.repo)
	followHandler := followHandlerPkg.NewHandler(followService, s.repo, s.authSvc)

	apikeyRepo := apikey.NewRepository(s.postgresDB.GORM)
	apiKeysHandler := apikeys.NewHandler(apikeyRepo)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "functionfly-jwt-secret-key-2026"
	}
	apiKeyAuthHandler := apikeys.NewAPIKeyAuthHandler(apikeyRepo, jwtSecret)

	monitoringHandler := monitoring.NewHandler(s.repo, s.monitoringSvc, s.realtimeMonitor, s.authSvc)
	mfaHandler := mfaHandlerPkg.NewMFAHandler(s.authSvc)

	notificationHandler := notificationHandlerPkg.NewHandler(s.notificationSvc, s.notificationRepo)
	notificationWSHandler := notificationHandlerPkg.NewWebSocketHandler(
		notificationHandlerPkg.NewWebSocketHub(logrus.New()),
		logrus.New(),
	)

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

	edgeCache := cache.NewEdgeCacheService(
		cacheService.GetRegistryCache(),
		cdnService,
		cacheConfiguration.ToEdgeCacheConfig(),
	)
	edgeCache.SetRepository(registryRepo)

	registryHandler := registryhandler.NewHandler(registryRepo, s.repo, cacheService, cdnService, edgeCache, s.realtimeMonitor, platformFeeRepo)
	adminRegistryHandler := admin.NewRegistryHandler(registryRepo, cacheService)
	oversightHandler := admin.NewOversightHandler(registryRepo, s.postgresDB.GORM, nil)
	docsHandler := registryhandler.NewDocumentationHandler(registryRepo)
	registryPlaygroundHandler := registryhandler.NewPlaygroundHandler(registryRepo)
	tutorialsHandler := registryhandler.NewTutorialsHandler()

	teamHandler := teams.NewHandler(s.repo, nil)
	providersHandler := providers.NewHandler(s.repo)
	dashboardHandler := dashboard.NewHandler(s.repo)
	enterpriseSLAHandler := enterprise.NewSLAHandler(s.repo)

	stateRepo := staterepo.NewStateRepository(s.postgresDB.GORM)
	triggerExecutor := staterepo.NewHTTPTriggerExecutor(
		os.Getenv("FUNCTION_EXECUTION_URL"),
		os.Getenv("FUNCTION_API_KEY"),
		logrus.New(),
	)
	triggerEngineConfig := staterepo.DefaultTriggerEngineConfig()
	if envEnabled := os.Getenv("TRIGGER_ENGINE_ENABLED"); envEnabled != "" {
		triggerEngineConfig.Enabled = envEnabled == "true"
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
	stateFabricHandler := statefabric.NewHandler(stateFabricRepo, sfAddonRepo)

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
		aiSupportConfig.BaseURL = "http://localhost:8081"
	}
	if aiSupportConfig.Model == "" {
		aiSupportConfig.Model = "gpt-4o-mini"
	}
	aiSupportClient := support.NewAIServiceClient(aiSupportConfig, supportLogger)

	supportService := support.NewService(supportRepo, aiSupportClient, nil, supportLogger)
	supportHdlr := supportHandler.NewHandler(supportService, supportLogger)
	supportAdminHdlr := supportHandler.NewAdminHandler(supportRepo, supportLogger)

	aepHandler := agenthandler.NewHandler(s.postgresDB.GORM, s.redisClient, registryRepo, s.repo, s.notificationSvc)

	agentIdentityRepo := identity.NewRepository(s.postgresDB.GORM)
	agentEconomyService := economy.NewService(s.postgresDB.GORM)
	agentMarketplaceService := marketplace.NewService(s.postgresDB.GORM)
	agentAutonomyService := autonomy.NewService(s.postgresDB.GORM)
	agentEvolutionService := evolution.NewService(s.postgresDB.GORM)
	agentSwarmMessageService := swarm.NewMessageService(s.postgresDB.GORM)
	agentSwarmService := swarm.NewService(s.postgresDB.GORM, agentIdentityRepo, agentEconomyService)

	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://127.0.0.1:9091"
	}
	statusRepo := statushandler.NewRepository(s.postgresDB.DB)
	statusRepo.SetGormDB(s.postgresDB.GORM)
	statusHandlerInst := statushandler.NewHandler(statusRepo, prometheusURL, s.authSvc)

	factoryConfig := factorysvc.DefaultConfig("factory-agent")
	factoryDiscovery := discovery.NewService(s.postgresDB.GORM)
	factoryGeneration := initializeGenerationServiceWithCache(s.postgresDB.GORM, s.redisClient)
	if detExec := registryexecution.NewRegistryDeterminismExecutorWithSandbox(registryRepo); detExec != nil {
		factoryGeneration.SetDeterminismExecutor(detExec)
	}
	factoryTesting := testing.NewService(s.postgresDB.GORM, nil, nil)
	factoryPublisher := agentdeployment.NewPublisher(s.postgresDB.GORM)
	factoryService := factorysvc.NewService(s.postgresDB.GORM, factoryConfig, factoryDiscovery, factoryGeneration, factoryTesting, factoryPublisher)

	loadedFactoryConfig, err := factoryService.GetConfig(context.Background())
	if err != nil {
		logrus.WithError(err).Warn("failed to load factory config from database, using defaults")
	} else {
		factoryConfig = loadedFactoryConfig
		logrus.Info("loaded factory config from database")
	}

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

	factoryHandler := factoryhandler.NewHandler(s.postgresDB.GORM, factoryService, factoryDiscovery, factoryPublisher, &factoryConfig, factoryPipelineScheduler)

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

	experimentService := factorysvc.NewExperimentService(s.postgresDB.GORM)
	experimentAdapter := factorysvc.NewGenerationExperimentAdapter(s.postgresDB.GORM, experimentService)
	experimentHandler := factoryhandler.NewExperimentHandler(s.postgresDB.GORM, experimentService, experimentAdapter)

	categorizationSvc := categorization.NewService(s.postgresDB.GORM)
	categorizationHandler := categorizationhandler.NewHandler(s.postgresDB.GORM, categorizationSvc)

	analyticsSvc := analytics.NewService(s.postgresDB.GORM, analytics.DefaultServiceConfig(factoryConfig.AgentID))
	analyticsHandler := analyticshandler.NewHandler(analyticsSvc, s.authSvc)

	swarmHandler := agenthandler.NewSwarmHandler(
		agentSwarmService,
		agentSwarmMessageService,
		agentEconomyService,
		agentMarketplaceService,
		agentEvolutionService,
		agentAutonomyService,
		agentIdentityRepo,
	)

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
	s.router.Use(middleware.TracingMiddleware)
	s.router.Use(maintenanceMiddleware.CheckMaintenanceMode)

	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(advancedSecurityMiddleware.CORSMiddleware(http.HandlerFunc(next.ServeHTTP)))
	})

	if os.Getenv("DEVELOPMENT") != "true" {
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
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.SecurityHeaders(http.HandlerFunc(next.ServeHTTP)))
		})
	}

	s.router.Use(monitoringPkg.HTTPMetricsMiddleware)

	// ── Subrouters & rate limiters ────────────────────────────────────────────
	api := s.router.PathPrefix("/v1").Subrouter()
	apiV2 := s.router.PathPrefix("/v2").Subrouter()
	protected := api.PathPrefix("").Subrouter()

	authRateLimiter := middleware.NewAuthRateLimiter()
	vaultRateLimiter := middleware.NewVaultRateLimiter()
	flywheelRateLimiter := middleware.NewFlywheelRateLimiter()

	// Status WebSocket hub — must be wired before route registration
	statusWSHub := statushandler.NewStatusWebSocketHub(statusHandlerInst, logrus.New())
	go statusWSHub.Run()
	statusHandlerInst.SetStatusHub(statusWSHub)

	// ── Domain-scoped route registration ─────────────────────────────────────
	registerAuthRoutes(
		s.router, api,
		authRateLimiter, authMiddleware,
		authHandler, apiKeyAuthHandler, usersHandler,
		followHandler, apiKeysHandler, billingHandler,
		mfaHandler, notificationHandler, notificationWSHandler,
	)

	registerRegistryRoutes(
		s, api, apiV2,
		authMiddleware, executionSecurityMW, verificationMiddleware,
		registryRepo, registryHandler, registryPlaygroundHandler,
		appPlaygroundHandler, docsHandler, tutorialsHandler,
		versionHandler, contentHandler, feedbackHandler, recommendationHandler,
	)

	registerPlatformRoutes(
		s, api, protected,
		authMiddleware, advancedSecurityMiddleware, vaultRateLimiter,
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
		supportHdlr, supportAdminHdlr,
	)

	registerAgentRoutes(
		s, api, protected,
		authMiddleware, flywheelRateLimiter,
		aepHandler, swarmHandler,
		registryRepo, cacheService,
		platformFeeRepo,
	)

	// Initialize admin security middleware
	csrfMiddleware := middleware.NewCSRFMiddleware(s.redisClient, s.authSvc)
	adminRateLimiter := middleware.NewAdminRateLimiter(s.redisClient)
	adminSessionMiddleware := middleware.NewAdminSessionMiddleware(s.postgresDB.DB, s.authSvc)
	ipAllowlistMiddleware := middleware.NewIPAllowlistMiddleware(s.postgresDB.DB, s.redisClient)
	adminIPAllowlistHandler := admin.NewAdminIPAllowlistHandler(s.postgresDB.DB, ipAllowlistMiddleware)
	adminAuditHandler := admin.NewAdminAuditHandler(s.postgresDB.DB)
	securityEventHandler := admin.NewSecurityEventHandler(s.postgresDB.DB)
	alertHandler := admin.NewAlertHandler(s.postgresDB.DB)

	registerAdminRoutes(
		s, api, authMiddleware, advancedSecurityMiddleware,
		adminHandler, adminBackendsHandler, adminProvidersHandler,
		maintenanceHandler, feedbackHandler, monitoringHandler,
		securityHandler, mfaHandler, adminRegistryHandler,
		registryHandler, oversightHandler, factoryHandler,
		stateFabricHandler, contentHandler,
		csrfMiddleware, adminRateLimiter, adminSessionMiddleware,
		ipAllowlistMiddleware, adminIPAllowlistHandler, adminAuditHandler, securityEventHandler, alertHandler,
	)

	// Trust API for external platform partners
	registerTrustAPIRoutes(s, api, registryRepo)

	// ── Infrastructure endpoints ──────────────────────────────────────────────
	wellknownHandler := wellknown.NewHandler(registryRepo)
	s.router.HandleFunc("/.well-known/functionfly.json", wellknownHandler.HandleWellKnown).Methods("GET", "OPTIONS")

	s.router.HandleFunc("/health", s.handleHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/healthz", s.handleHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/detailed", s.handleDetailedHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/check", s.handleHealthCheck).Methods("GET", "OPTIONS")
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.router.HandleFunc("/ws/v1/status", statusHandlerInst.HandleWebSocketStatus).Methods("GET")

	// ── SPA catch-all routes ──────────────────────────────────────────────────
	// Serve index.html for /fx/*, /run/*, /replay/* (playground SPA paths)
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") ||
			strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") ||
			r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 1 && (pathParts[0] == "fx" || pathParts[0] == "run" || pathParts[0] == "replay")
	}).HandlerFunc(s.serveSPAIndex)

	// Public routing endpoint: /{appSlug}/*
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/content/") ||
			strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" || r.URL.Path == "/healthz" ||
			strings.HasPrefix(r.URL.Path, "/v1/") {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 2 && pathParts[0] != "" && pathParts[0] != "api" &&
			pathParts[0] != "content" && pathParts[0] != "health" &&
			pathParts[0] != "fx" && pathParts[0] != "run" && pathParts[0] != "replay"
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
