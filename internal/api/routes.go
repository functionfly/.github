package api

import (
	"context"
	"net/http"
	"os"
	"strings"

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
	flywheelhandler "github.com/functionfly/functionfly/internal/api/handlers/flywheel"
	followHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/follow"
	"github.com/functionfly/functionfly/internal/api/handlers/functions"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	notificationHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/notifications"
	"github.com/functionfly/functionfly/internal/api/handlers/playground"
	"github.com/functionfly/functionfly/internal/api/handlers/providers"
	"github.com/functionfly/functionfly/internal/api/handlers/recommendations"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	drehandler "github.com/functionfly/functionfly/internal/api/handlers/registry/dre"
	registryexecution "github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/handlers/security"
	"github.com/functionfly/functionfly/internal/api/handlers/state"
	"github.com/functionfly/functionfly/internal/api/handlers/statefabric"
	statushandler "github.com/functionfly/functionfly/internal/api/handlers/status"
	"github.com/functionfly/functionfly/internal/api/handlers/teams"
	usersHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/users"
	"github.com/functionfly/functionfly/internal/api/handlers/vault"
	versionhandler "github.com/functionfly/functionfly/internal/api/handlers/version"
	"github.com/functionfly/functionfly/internal/api/handlers/wellknown"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/captcha"
	"github.com/functionfly/functionfly/internal/flywheel"
	monitoringPkg "github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	statefabricrepo "github.com/functionfly/functionfly/internal/storage/statefabric"
	vaultstorage "github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// setupRoutes configures all API routes
func (s *Server) setupRoutes(realtimeMonitor *monitoringPkg.RealtimeMonitor) {
	// Initialize handlers
	authHandler := authHandlerPkg.NewHandler(s.authSvc)
	usersHandler := usersHandlerPkg.NewHandler(s.repo, s.authSvc)
	billingHandler := billing.NewHandler(s.repo)
	appsHandler := apps.NewHandler(s.repo)
	backendsHandler := backends.NewHandler(s.repo, s.routingSvc)
	deploymentsHandler := deployments.NewHandler(s.repo, s.deploySvc)
	functionsHandler := functions.NewHandler(s.repo, s.deploySvc)
	unifiedAnalyticsSvc := unified.NewService(s.postgresDB.GORM, s.usageMetricsAgg)
	adminHandler := admin.NewHandler(s.repo, s.authSvc, unifiedAnalyticsSvc)
	adminBackendsHandler := admin.NewBackendsHandler(s.repo, s.authSvc)
	adminProvidersHandler := admin.NewProvidersHandler(s.repo, s.authSvc)
	securityHandler := security.NewHandler(s.repo, s.authSvc)

	// Initialize maintenance mode handler
	maintenanceRepo := storage.NewMaintenanceRepository(s.postgresDB.GORM)
	maintenanceHandler := admin.NewMaintenanceHandler(maintenanceRepo, s.authSvc)
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(maintenanceRepo)
	contentHandler := content.NewHandler(s.repo)
	feedbackHandler := feedbackHandlerPkg.NewHandler(s.repo, s.storageService)

	// Initialize follow handler
	followService := services.NewFollowService(s.repo)
	followHandler := followHandlerPkg.NewHandler(followService, s.repo, s.authSvc)

	// Initialize monitoring handler
	monitoringHandler := monitoring.NewHandler(s.repo, s.monitoringSvc, s.realtimeMonitor, s.authSvc)
	mfaHandler := mfaHandlerPkg.NewMFAHandler(s.authSvc)

	// Initialize notification handler
	notificationHandler := notificationHandlerPkg.NewHandler(s.notificationSvc, s.notificationRepo)

	// Initialize notification WebSocket handler
	notificationWSHandler := notificationHandlerPkg.NewWebSocketHandler(
		notificationHandlerPkg.NewWebSocketHub(logrus.New()),
		logrus.New(),
	)

	// Initialize version handler
	versionRepo := versioning.NewRepository(s.postgresDB.DB)
	versionHandler := versionhandler.NewHandler(versionRepo)

	// Initialize app-based playground handler
	appPlaygroundHandler := playground.NewHandler(s.repo)

	// Initialize cache configuration from environment
	cacheConfiguration := cache.LoadCacheConfiguration()
	if err := cacheConfiguration.Validate(); err != nil {
		// Log the validation error and disable all caching features
		// Cache is an optimization, not a hard requirement - server should continue
		logrus.WithError(err).Error("Invalid cache configuration, disabling all caching features")
		// Disable all cache features
		cacheConfiguration.DiskEnabled = false
		cacheConfiguration.RedisEnabled = false
		cacheConfiguration.CDNEnabled = false
		cacheConfiguration.EdgeCacheEnabled = false
		// Set sane fallback defaults for positive-integer fields
		if cacheConfiguration.MemoryMaxMB <= 0 {
			cacheConfiguration.MemoryMaxMB = 100
		}
		if cacheConfiguration.DefaultTTL <= 0 {
			cacheConfiguration.DefaultTTL = 3600 // 1 hour
		}
		if cacheConfiguration.RedisRegistryTTL <= 0 {
			cacheConfiguration.RedisRegistryTTL = 600 // 10 minutes
		}
	}

	// Initialize cache service with comprehensive configuration
	cacheService, err := cache.NewCacheService(s.postgresDB.GORM, s.redisClient, cacheConfiguration.ToCacheConfig())
	if err != nil {
		// Cache service initialization failed - try with minimal fallback config
		logrus.WithError(err).Error("Failed to initialize cache service, attempting fallback configuration")
		// Create fallback config with minimal in-memory-only settings
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
			// Even fallback failed - this should be effectively impossible with safe defaults
			// but we panic to ensure we don't continue with a nil cache service
			logrus.WithError(err).Error("Failed to initialize fallback cache service")
			panic("Failed to initialize cache service even with fallback configuration: " + err.Error())
		}
		logrus.Warn("Cache service initialized with fallback in-memory-only configuration")
	}

	// Initialize CDN service
	cdnService := cache.NewCDNService(cacheConfiguration.ToCDNConfig())

	// Initialize registry repository with Redis cache (must be before edge cache)
	var registryCache *cache.RegistryRedisCache
	if cacheConfiguration.RedisEnabled && s.redisClient != nil {
		registryCache = cacheService.GetRegistryCache()
	}
	registryRepo := registry.NewRegistryRepository(s.postgresDB.GORM, registryCache)

	// Initialize edge cache service (now registryRepo is defined)
	edgeCache := cache.NewEdgeCacheService(
		cacheService.GetRegistryCache(),
		cdnService,
		cacheConfiguration.ToEdgeCacheConfig(),
	)
	edgeCache.SetRepository(registryRepo)

	registryHandler := registryhandler.NewHandler(registryRepo, s.repo, cacheService, cdnService, edgeCache, s.realtimeMonitor)
	adminRegistryHandler := admin.NewRegistryHandler(registryRepo, cacheService)

	// Initialize oversight handler for trust, fraud, execution audit, and economic monitoring
	oversightHandler := admin.NewOversightHandler(registryRepo, s.postgresDB.GORM, nil)

	// Initialize documentation handler
	docsHandler := registryhandler.NewDocumentationHandler(registryRepo)

	// Initialize registry playground handler (for /fx/, /run/{author}/{name}, /replay/)
	registryPlaygroundHandler := registryhandler.NewPlaygroundHandler(registryRepo)

	// Initialize tutorials handler
	tutorialsHandler := registryhandler.NewTutorialsHandler()

	// Initialize team handler
	teamHandler := teams.NewHandler(s.repo, nil)

	// Initialize providers handler
	providersHandler := providers.NewHandler(s.repo)

	// Initialize dashboard handler (tenant-scoped metrics and activity)
	dashboardHandler := dashboard.NewHandler(s.repo)
	enterpriseSLAHandler := enterprise.NewSLAHandler(s.repo)

	// Initialize state handler
	stateRepo := staterepo.NewStateRepository(s.postgresDB.GORM)

	// Initialize trigger engine for state changes
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

	// Initialize agent memory handler
	memoryRepo := state.NewAgentMemoryRepository(s.postgresDB.GORM)
	memoryHandler := state.NewAgentMemoryHandler(memoryRepo)

	// Initialize new agent memory handler (REST API at /agent-memories)
	agentMemoryHandler := agentmemoryhandler.NewHandler(s.postgresDB.GORM)

	// Initialize state fabric handler (dashboard State Fabric feature)
	stateFabricRepo := statefabricrepo.NewRepository(s.postgresDB.GORM)
	stateFabricHandler := statefabric.NewHandler(stateFabricRepo)

	// Initialize vault handler (Secrets Vault feature)
	vaultRepo := vaultstorage.NewRepository(s.postgresDB.GORM)
	vaultHandler := vault.NewHandler(vaultRepo, logrus.New())

	// Initialize AEP (Agent Execution Plan) handler
	aepHandler := agenthandler.NewHandler(s.postgresDB.GORM, s.redisClient, registryRepo)

	// Initialize agent identity repository for swarm services
	agentIdentityRepo := identity.NewRepository(s.postgresDB.GORM)

	// Initialize agent swarm services
	agentEconomyService := economy.NewService(s.postgresDB.GORM)
	agentMarketplaceService := marketplace.NewService(s.postgresDB.GORM)
	agentAutonomyService := autonomy.NewService(s.postgresDB.GORM)
	agentEvolutionService := evolution.NewService(s.postgresDB.GORM)
	agentSwarmMessageService := swarm.NewMessageService(s.postgresDB.GORM)
	agentSwarmService := swarm.NewService(s.postgresDB.GORM, agentIdentityRepo, agentEconomyService)

	// Initialize status page handler
	statusRepo := statushandler.NewRepository(s.postgresDB.DB)
	statusRepo.SetGormDB(s.postgresDB.GORM)
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://127.0.0.1:9091"
	}
	statusHandler := statushandler.NewHandler(statusRepo, prometheusURL, s.authSvc)

	factoryConfig := factorysvc.DefaultConfig("factory-agent")
	factoryDiscovery := discovery.NewService(s.postgresDB.GORM)

	// Initialize generation service with Redis-backed cache if available
	factoryGeneration := initializeGenerationServiceWithCache(s.postgresDB.GORM, s.redisClient)

	factoryTesting := testing.NewService(s.postgresDB.GORM, nil, nil)
	factoryPublisher := agentdeployment.NewPublisher(s.postgresDB.GORM)
	factoryService := factorysvc.NewService(s.postgresDB.GORM, factoryConfig, factoryDiscovery, factoryGeneration, factoryTesting, factoryPublisher)

	// Load factory config from database on startup (creates default if not exists)
	loadedFactoryConfig, err := factoryService.GetConfig(context.Background())
	if err != nil {
		logrus.WithError(err).Warn("failed to load factory config from database, using defaults")
	} else {
		factoryConfig = loadedFactoryConfig
		logrus.Info("loaded factory config from database")
	}

	// Initialize factory pipeline scheduler
	factoryPipelineScheduler := scheduler.NewFactoryPipelineScheduler(factoryService)
	// Start scheduler with config from database (respects enabled flag)
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

	// Initialize experiment service and handler
	experimentService := factorysvc.NewExperimentService(s.postgresDB.GORM)
	experimentAdapter := factorysvc.NewGenerationExperimentAdapter(s.postgresDB.GORM, experimentService)
	experimentHandler := factoryhandler.NewExperimentHandler(s.postgresDB.GORM, experimentService, experimentAdapter)

	// Initialize categorization service and handler
	categorizationSvc := categorization.NewService(s.postgresDB.GORM)
	categorizationHandler := categorizationhandler.NewHandler(s.postgresDB.GORM, categorizationSvc)

	// Initialize analytics service for factory metrics
	analyticsSvc := analytics.NewService(s.postgresDB.GORM, analytics.DefaultServiceConfig(factoryConfig.AgentID))
	analyticsHandler := analyticshandler.NewHandler(analyticsSvc, s.authSvc)

	// Initialize Swarm handler (for swarm/marketplace/evolution features)
	swarmHandler := agenthandler.NewSwarmHandler(
		agentSwarmService,
		agentSwarmMessageService,
		agentEconomyService,
		agentMarketplaceService,
		agentEvolutionService,
		agentAutonomyService,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(s.authSvc)
	advancedSecurityMiddleware := middleware.NewAdvancedSecurityMiddleware(s.repo)
	// loggingMiddleware := middleware.NewLoggingMiddleware() // Temporarily disabled

	// Initialize CAPTCHA service
	captchaService := captcha.NewCaptchaService(nil)
	// Register CAPTCHA providers based on environment variables
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

	// Initialize execution security middleware
	executionSecurityMW := middleware.NewExecutionCoordinatorMiddleware(s.postgresDB.GORM, nil, captchaService)

	// Initialize verification middleware
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
		registryRepo := registry.NewRegistryRepository(s.postgresDB.GORM, nil)
		verificationMiddleware = middleware.NewVerificationMiddleware(registryRepo, clamAVURL, yaraURL, minimumTrustLevel)
	}

	// Apply distributed tracing middleware first (W3C Trace Context)
	s.router.Use(middleware.TracingMiddleware)

	// Apply maintenance mode check middleware
	s.router.Use(maintenanceMiddleware.CheckMaintenanceMode)

	// Apply logging middleware first (temporarily disabled)
	// s.router.Use(func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(loggingMiddleware.StructuredLogger(http.HandlerFunc(next.ServeHTTP)))
	// })

	// Apply CORS middleware to all routes (always enabled for development)
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(advancedSecurityMiddleware.CORSMiddleware(http.HandlerFunc(next.ServeHTTP)))
	})

	// Apply advanced security middleware to all routes (skip in development mode)
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
		// Apply basic security headers in development mode
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(advancedSecurityMiddleware.SecurityHeaders(http.HandlerFunc(next.ServeHTTP)))
		})
	}

	// Apply Prometheus metrics middleware
	s.router.Use(monitoringPkg.HTTPMetricsMiddleware)

	// API versioning
	api := s.router.PathPrefix("/v1").Subrouter()

	// Dedicated rate limiter for auth endpoints: 10 req/min per IP (much stricter than global)
	// Configurable via AUTH_RATE_LIMIT_REQUESTS / AUTH_RATE_LIMIT_WINDOW_SECONDS env vars.
	authRateLimiter := middleware.NewAuthRateLimiter()

	// Auth routes (public) - on main router, not /v1 subrouter
	s.router.HandleFunc("/auth/login", authRateLimiter.Limit(authHandler.HandleLogin)).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/auth/refresh", authHandler.HandleRefreshToken).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/auth/signup", authRateLimiter.Limit(authHandler.HandleSignup)).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/auth/check-username", authHandler.HandleCheckUsernameAvailability).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/auth/verify-email", authHandler.HandleVerifyEmail).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/auth/resend-verification", authRateLimiter.Limit(authHandler.HandleResendVerification)).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/auth/get-session", authHandler.HandleGetSession).Methods("GET", "OPTIONS")

	// OAuth routes (public) - on main router, not /v1 subrouter
	s.router.HandleFunc("/auth/oauth/providers", authHandler.HandleGetOAuthProviders).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/auth/oauth/url", authHandler.HandleGetOAuthURL).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/auth/oauth/{provider}/callback", authHandler.HandleOAuthCallback).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/validate", authMiddleware.RequireAuth(authHandler.HandleValidateToken)).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/logout", authMiddleware.RequireAuth(authHandler.HandleLogout)).Methods("POST", "OPTIONS")

	// Password reset routes (public)
	api.HandleFunc("/auth/password-reset", authRateLimiter.Limit(authHandler.HandlePasswordResetRequest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/password-reset/confirm", authRateLimiter.Limit(authHandler.HandlePasswordResetConfirm)).Methods("POST", "OPTIONS")

	// User profile routes
	// Public: get any user's profile by username
	api.HandleFunc("/users/{username}", usersHandler.HandleGetPublicProfile).Methods("GET", "OPTIONS")
	// Protected: get/update current user's own profile
	api.HandleFunc("/users/me", authMiddleware.RequireAuth(usersHandler.HandleGetMe)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me", authMiddleware.RequireAuth(usersHandler.HandleUpdateMe)).Methods("PATCH", "OPTIONS")

	// User sessions (protected)
	api.HandleFunc("/users/me/sessions", authMiddleware.RequireAuth(usersHandler.HandleListSessions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/sessions/revoke-others", authMiddleware.RequireAuth(usersHandler.HandleRevokeOtherSessions)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/sessions/{id}", authMiddleware.RequireAuth(usersHandler.HandleRevokeSession)).Methods("DELETE", "OPTIONS")
	// User settings (protected; for current authenticated user)
	api.HandleFunc("/users/me/settings", authMiddleware.RequireAuth(usersHandler.HandleGetUserSettingsMe)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/settings/profile", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsProfileMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/notifications", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsNotificationsMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/privacy", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsPrivacyMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/visibility", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsVisibilityMe)).Methods("PATCH", "OPTIONS")
	// User settings by username (protected; requester must match username)
	api.HandleFunc("/users/{username}/settings", authMiddleware.RequireAuth(usersHandler.HandleGetUserSettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/profile", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsProfile)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/notifications", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsNotifications)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/privacy", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsPrivacy)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/visibility", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsVisibility)).Methods("PATCH", "OPTIONS")

	// User profile analytics (public)
	api.HandleFunc("/users/{username}/analytics", usersHandler.HandleGetUserAnalytics).Methods("GET", "OPTIONS")
	// User achievements (public)
	api.HandleFunc("/users/{username}/achievements", usersHandler.HandleGetUserAchievements).Methods("GET", "OPTIONS")
	// User activity feed (public)
	api.HandleFunc("/users/{username}/activity", usersHandler.HandleGetUserActivity).Methods("GET", "OPTIONS")
	// User skills (public)
	api.HandleFunc("/users/{username}/skills", usersHandler.HandleGetUserSkills).Methods("GET", "OPTIONS")
	// User activity management (protected - for creating activity)
	api.HandleFunc("/users/me/activity", authMiddleware.RequireAuth(usersHandler.HandleCreateUserActivity)).Methods("POST", "OPTIONS")
	// User skills management (protected)
	api.HandleFunc("/users/me/skills", authMiddleware.RequireAuth(usersHandler.HandleAddUserSkill)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/skills/{id}", authMiddleware.RequireAuth(usersHandler.HandleRemoveUserSkill)).Methods("DELETE", "OPTIONS")

	// Follow routes (protected for write, public for read where noted)
	// User follows
	api.HandleFunc("/v1/follow/users/{username}/follow", authMiddleware.RequireAuth(followHandler.HandleFollowUser)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/follow/users/{username}/follow", authMiddleware.RequireAuth(followHandler.HandleUnfollowUser)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/v1/follow/users/{username}/followers", followHandler.HandleGetUserFollowers).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/follow/users/{username}/following", followHandler.HandleGetUserFollowing).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/follow/users/{username}/status", authMiddleware.RequireAuth(followHandler.HandleCheckFollowingStatus)).Methods("GET", "OPTIONS")

	// Function follows
	api.HandleFunc("/v1/follow/functions/{functionID}/follow", authMiddleware.RequireAuth(followHandler.HandleFollowFunction)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/follow/functions/{functionID}/follow", authMiddleware.RequireAuth(followHandler.HandleUnfollowFunction)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/v1/follow/functions/{functionID}/followers", followHandler.HandleGetFunctionFollowers).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/follow/functions/{functionID}/status", authMiddleware.RequireAuth(followHandler.HandleCheckFunctionFollowingStatus)).Methods("GET", "OPTIONS")

	// My follows (protected)
	api.HandleFunc("/v1/follow/me/functions", authMiddleware.RequireAuth(followHandler.HandleGetMyFollowedFunctions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/follow/me/stats", authMiddleware.RequireAuth(followHandler.HandleGetMyFollowStats)).Methods("GET", "OPTIONS")

	// Billing portal (Stripe Customer Portal)
	api.HandleFunc("/billing/portal-session", authMiddleware.RequireAuth(billingHandler.HandleCreatePortalSession)).Methods("POST", "OPTIONS")
	// User billing endpoints
	api.HandleFunc("/billing/subscription", authMiddleware.RequireAuth(billingHandler.HandleGetSubscription)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/invoices", authMiddleware.RequireAuth(billingHandler.HandleListInvoices)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/usage", authMiddleware.RequireAuth(billingHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	// @username profile routes (clean URL structure)
	api.HandleFunc("/@/{username}", usersHandler.HandleGetPublicProfileByAt).Methods("GET", "OPTIONS")

	// MFA routes (protected)
	api.HandleFunc("/auth/mfa/setup", authMiddleware.RequireAuth(mfaHandler.SetupMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/verify", authMiddleware.RequireAuth(mfaHandler.VerifyMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/enable", authMiddleware.RequireAuth(mfaHandler.EnableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/disable", authMiddleware.RequireAuth(mfaHandler.DisableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/status", authMiddleware.RequireAuth(mfaHandler.GetMFAStatus)).Methods("GET", "OPTIONS")

	// Notification routes (protected)
	api.HandleFunc("/notifications", authMiddleware.RequireAuth(notificationHandler.HandleListNotifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/unread-count", authMiddleware.RequireAuth(notificationHandler.HandleGetUnreadCount)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/read-all", authMiddleware.RequireAuth(notificationHandler.HandleMarkAllAsRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/notifications/{id}/read", authMiddleware.RequireAuth(notificationHandler.HandleMarkAsRead)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/notifications/{id}", authMiddleware.RequireAuth(notificationHandler.HandleDeleteNotification)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/users/me/notification-preferences", authMiddleware.RequireAuth(notificationHandler.HandleGetPreferences)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/notification-preferences", authMiddleware.RequireAuth(notificationHandler.HandleUpdatePreferences)).Methods("PATCH", "OPTIONS")

	// Notification WebSocket route (protected)
	api.HandleFunc("/notifications/stream", authMiddleware.RequireAuth(notificationWSHandler.HandleWebSocket))

	// Playground routes (public) - App-based functions
	api.HandleFunc("/run/{appSlug}/{functionName}", appPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{appSlug}/{functionName}/info", appPlaygroundHandler.HandleGetFunctionInfo).Methods("GET", "OPTIONS")

	// Perfect Playground routes - Registry functions with new URL structure
	// /fx/{author}/{name} - Public function page (docs + playground)
	// /run/{author}/{name} - Live playground session
	// /run/{author}/{name}/execute - Execute function
	// /replay/{execution_id} - View past execution
	// /fx/{author}/{name}/code - Get code examples
	// /fx/{author}/{name}/ai-schema - Get AI tool schema
	api.HandleFunc("/fx/{author}/{name}", registryPlaygroundHandler.HandleFunctionPage).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}", registryPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/execute", registryPlaygroundHandler.HandlePlaygroundExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/share", registryPlaygroundHandler.HandlePlaygroundShare).Methods("POST", "OPTIONS")
	api.HandleFunc("/replay/{executionId}", registryPlaygroundHandler.HandleReplay).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/code", registryPlaygroundHandler.HandleCodeExamples).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/ai-schema", registryPlaygroundHandler.HandleAIToolSchema).Methods("GET", "OPTIONS")

	// NEW: @username function routes (clean URL structure)
	// /@/{username}/v1/fx/{functionName} - Public function page
	// /@/{username}/v1/fx/{functionName}/execute - Execute function
	// /@/{username}/v1/fx/{functionName}/v/{version} - Versioned function
	// /@/{username}/v1/fx/{functionName}/versions - List versions
	// /@/{username}/v1/fx/{functionName}/stats - Function stats
	api.HandleFunc("/@/{username}/v1/fx/{functionName}", registryPlaygroundHandler.HandleFunctionPageAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/execute", registryPlaygroundHandler.HandleExecuteAt).Methods("POST", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/v/{version}", registryPlaygroundHandler.HandleFunctionPageAtVersion).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/versions", registryHandler.HandleListVersionsAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/stats", registryHandler.HandleGetFunctionStatsAt).Methods("GET", "OPTIONS")

	// Execute playground function with security middleware
	securePlaygroundExecuteHandler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		appSlug := vars["appSlug"]
		functionName := vars["functionName"]

		// Get app and function to extract ID for security middleware
		app, err := s.repo.GetAppBySlug(appSlug)
		if err != nil {
			http.Error(w, "App not found", http.StatusNotFound)
			return
		}

		fn, err := s.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, functionName)
		if err != nil {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}

		// Apply execution security middleware
		securityHandler := executionSecurityMW.SecureExecution(fn.ID, fn.Version)(appPlaygroundHandler.HandleExecute)
		securityHandler.ServeHTTP(w, r)
	}

	// Apply verification middleware on top of security middleware if enabled
	if verificationMiddleware != nil {
		api.Handle("/run/{appSlug}/{functionName}/execute", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(securePlaygroundExecuteHandler))).Methods("POST", "OPTIONS")
	} else {
		api.HandleFunc("/run/{appSlug}/{functionName}/execute", securePlaygroundExecuteHandler).Methods("POST", "OPTIONS")
	}

	// Functions routes (protected)

	// Metrics routes (public - for landing page demo)
	api.HandleFunc("/metrics/global", s.handleGlobalMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/metrics/stream", s.handleMetricsStream).Methods("GET", "OPTIONS")

	// Monitoring routes
	api.HandleFunc("/monitoring/metrics", monitoringHandler.HandleGetMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/alerts", monitoringHandler.HandleGetAlerts).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/health", monitoringHandler.HandleGetSystemHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/events", monitoringHandler.HandleGetMonitoringEvents).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/realtime", monitoringHandler.HandleRealtimeConnection).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/realtime/stats", monitoringHandler.HandleGetRealtimeStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/local-runtime", monitoringHandler.HandleGetLocalRuntimeMetrics).Methods("GET", "OPTIONS")

	// Database monitoring routes
	api.HandleFunc("/monitoring/database/health", monitoringHandler.HandleGetDatabaseHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/metrics", monitoringHandler.HandleGetDatabaseMetrics).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/alerts", monitoringHandler.HandleGetDatabaseAlerts).Methods("GET", "OPTIONS")
	api.HandleFunc("/monitoring/database/check", monitoringHandler.HandleCheckDatabaseHealth).Methods("POST", "OPTIONS")
	api.HandleFunc("/monitoring/database/changes", monitoringHandler.HandleSubscribeToDatabaseChanges).Methods("GET", "OPTIONS")

	// Security metrics (public - for footer and general access)
	api.HandleFunc("/metrics/security", securityHandler.HandleGetSecurityMetrics).Methods("GET")
	api.HandleFunc("/metrics/security/services", securityHandler.HandleGetServiceStatus).Methods("GET")
	api.HandleFunc("/metrics/security/certificates", securityHandler.HandleGetSSLCertificates).Methods("GET")
	api.HandleFunc("/metrics/security/incidents", securityHandler.HandleGetRecentIncidents).Methods("GET")
	api.HandleFunc("/metrics/security/compliance", securityHandler.HandleGetComplianceFrameworks).Methods("GET")
	api.HandleFunc("/metrics/security/measures", securityHandler.HandleGetSecurityMeasures).Methods("GET")
	api.HandleFunc("/metrics/security/incident-response", securityHandler.HandleGetIncidentResponse).Methods("GET")
	api.HandleFunc("/metrics/security/faq", securityHandler.HandleGetSecurityFAQ).Methods("GET")
	api.HandleFunc("/metrics/security/resources", securityHandler.HandleGetSecurityResources).Methods("GET")
	api.HandleFunc("/metrics/security/contacts", securityHandler.HandleGetContactInfo).Methods("GET")

	// Status Page routes (public - for status page)
	// GET /api/v1/status - Current platform status
	api.HandleFunc("/status", statusHandler.HandleGetPlatformStatus).Methods("GET", "OPTIONS")
	// GET /api/v1/status/edge - Edge (edge.functionfly.com) health, uptime, request stats
	api.HandleFunc("/status/edge", s.handleEdgeStatus).Methods("GET", "OPTIONS")
	// GET /api/v1/status/components - Per-component status
	api.HandleFunc("/status/components", statusHandler.HandleGetComponents).Methods("GET", "OPTIONS")
	// GET /api/v1/status/providers - Per-provider status by region
	api.HandleFunc("/status/providers", statusHandler.HandleGetProviders).Methods("GET", "OPTIONS")

	// Incident routes (public read, admin write)
	// GET /api/v1/incidents - List incidents with filtering
	api.HandleFunc("/incidents", statusHandler.HandleListIncidents).Methods("GET", "OPTIONS")
	// GET /api/v1/incidents/:id - Single incident details
	api.HandleFunc("/incidents/{id}", statusHandler.HandleGetIncident).Methods("GET", "OPTIONS")
	// POST /api/v1/incidents - Create incident (admin only)
	api.HandleFunc("/incidents", authMiddleware.RequireAuth(statusHandler.HandleCreateIncident)).Methods("POST", "OPTIONS")
	// PATCH /api/v1/incidents/:id - Update incident (admin only)
	api.HandleFunc("/incidents/{id}", authMiddleware.RequireAuth(statusHandler.HandleUpdateIncident)).Methods("PATCH", "OPTIONS")

	// Metrics routes (public)
	// GET /api/v1/metrics/uptime - Historical uptime data
	api.HandleFunc("/metrics/uptime", statusHandler.HandleGetUptimeMetrics).Methods("GET", "OPTIONS")
	// GET /api/v1/metrics/latency - Latency trends by provider/region
	api.HandleFunc("/metrics/latency", statusHandler.HandleGetLatencyMetrics).Methods("GET", "OPTIONS")

	// Maintenance routes (public read, admin write)
	// GET /api/v1/maintenance - Scheduled maintenance windows
	api.HandleFunc("/maintenance", statusHandler.HandleListMaintenance).Methods("GET", "OPTIONS")
	// POST /api/v1/maintenance - Create maintenance window (admin only)
	api.HandleFunc("/maintenance", authMiddleware.RequireAuth(statusHandler.HandleCreateMaintenance)).Methods("POST", "OPTIONS")

	// API Version management routes (public)
	api.HandleFunc("/api/versions", versionHandler.HandleListVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/api/versions/{version}", versionHandler.HandleGetVersion).Methods("GET", "OPTIONS")
	// Deprecate API version (admin only)
	api.HandleFunc("/api/versions/{version}/deprecate", authMiddleware.RequireAuth(versionHandler.HandleDeprecateVersion)).Methods("POST", "OPTIONS")

	// API Version Lifecycle Management (Phase 2)
	api.HandleFunc("/api/versions", authMiddleware.RequireAuth(versionHandler.HandleCreateAPIVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api/versions/{version}", authMiddleware.RequireAuth(versionHandler.HandleUpdateAPIVersion)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/api/versions/{version}/set-default", authMiddleware.RequireAuth(versionHandler.HandleSetDefaultAPIVersion)).Methods("POST", "OPTIONS")

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

	// Experiment (A/B testing) routes
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

	// Categorization routes
	// GET /api/v1/categorization/taxonomy - Get full category and tag taxonomy
	api.HandleFunc("/categorization/taxonomy", categorizationHandler.HandleGetTaxonomy).Methods("GET", "OPTIONS")
	// GET /api/v1/categorization/categories - List all categories
	api.HandleFunc("/categorization/categories", categorizationHandler.HandleGetCategories).Methods("GET", "OPTIONS")
	// GET /api/v1/categorization/categories/{id} - Get specific category
	api.HandleFunc("/categorization/categories/{id}", categorizationHandler.HandleGetCategory).Methods("GET", "OPTIONS")
	// GET /api/v1/categorization/tags - List all tags
	api.HandleFunc("/categorization/tags", categorizationHandler.HandleGetTags).Methods("GET", "OPTIONS")
	// POST /api/v1/categorization/categorize - Categorize a function spec (without storing)
	api.HandleFunc("/categorization/categorize", categorizationHandler.HandleCategorize).Methods("POST", "OPTIONS")
	// POST /api/v1/categorization/analyze - Analyze code patterns
	api.HandleFunc("/categorization/analyze", categorizationHandler.HandleAnalyzeCode).Methods("POST", "OPTIONS")
	// GET /api/v1/categorization/functions/{id} - Get categorization for a function
	api.HandleFunc("/categorization/functions/{id}", categorizationHandler.HandleGetFunctionCategory).Methods("GET", "OPTIONS")
	// PUT /api/v1/categorization/functions/{id} - Update function categorization (manual override)
	api.HandleFunc("/categorization/functions/{id}", authMiddleware.RequireAuth(categorizationHandler.HandleUpdateFunctionCategory)).Methods("PUT", "OPTIONS")
	// POST /api/v1/categorization/functions/{id}/recategorize - Re-categorize a function
	api.HandleFunc("/categorization/functions/{id}/recategorize", authMiddleware.RequireAuth(categorizationHandler.HandleReCategorize)).Methods("POST", "OPTIONS")
	// GET /api/v1/categorization/category/{category} - Get functions by category
	api.HandleFunc("/categorization/category/{category}", categorizationHandler.HandleGetFunctionsByCategory).Methods("GET", "OPTIONS")
	// GET /api/v1/categorization/tag/{tag} - Get functions by tag
	api.HandleFunc("/categorization/tag/{tag}", categorizationHandler.HandleGetFunctionsByTag).Methods("GET", "OPTIONS")

	// Analytics routes for factory metrics
	analyticsHandler.RegisterRoutes(api, authMiddleware)

	// Unified analytics (tenant summary and time series)
	unifiedAnalyticsHandler := analyticshandler.NewUnifiedHandler(unifiedAnalyticsSvc)
	unifiedAnalyticsHandler.RegisterUnifiedRoutes(api, authMiddleware)

	// Content management (public - for frontend consumption)
	api.HandleFunc("/content/changelog", contentHandler.HandleGetPublishedChangelogEntries).Methods("GET")
	api.HandleFunc("/content/blog", contentHandler.HandleGetPublishedBlogPosts).Methods("GET")
	api.HandleFunc("/content/blog/{slug}", contentHandler.HandleGetPublishedBlogPostBySlug).Methods("GET")
	api.HandleFunc("/content/categories", contentHandler.HandleGetBlogCategories).Methods("GET")
	api.HandleFunc("/content/authors", contentHandler.HandleGetBlogAuthors).Methods("GET")

	// Feedback routes (public)
	api.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	api.HandleFunc("/feedback/history", authMiddleware.RequireAuth(feedbackHandler.GetFeedbackHistory)).Methods("GET")

	// Function Registry routes (public for read, protected for write)
	api.HandleFunc("/registry/functions", registryHandler.HandleListFunctions).Methods("GET")
	api.HandleFunc("/registry/functions", registryHandler.HandleDeleteAllFunctions).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleDeleteFunction).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs/category/{category}", registryHandler.HandleGetChangelogByCategory).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")
	api.HandleFunc("/registry/search", registryHandler.HandleSearchFunctions).Methods("GET")

	// Function version routes (public)
	api.HandleFunc("/functions/{functionId}/versions", versionHandler.HandleListFunctionVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}", versionHandler.HandleGetFunctionVersion).Methods("GET", "OPTIONS")
	// Function version changelog
	api.HandleFunc("/functions/{functionId}/versions/{version}/changelog", authMiddleware.RequireAuth(versionHandler.HandleCreateChangelog)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/changelogs", versionHandler.HandleGetChangelogs).Methods("GET", "OPTIONS")

	// Function Version Lifecycle Management (Phase 2)
	// Publish, Archive, Deprecate
	api.HandleFunc("/functions/{functionId}/versions/{version}/publish", authMiddleware.RequireAuth(versionHandler.HandlePublishVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/archive", authMiddleware.RequireAuth(versionHandler.HandleArchiveVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/deprecate", authMiddleware.RequireAuth(versionHandler.HandleDeprecateFunctionVersion)).Methods("POST", "OPTIONS")

	// Version aliases
	api.HandleFunc("/functions/{functionId}/versions/{version}/alias/{alias}", authMiddleware.RequireAuth(versionHandler.HandleSetAlias)).Methods("POST", "OPTIONS")

	// Rollback endpoints
	api.HandleFunc("/functions/{functionId}/versions/{version}/rollback", authMiddleware.RequireAuth(versionHandler.HandleRollbackVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/rollback", authMiddleware.RequireAuth(versionHandler.HandleRollbackLatest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/rollbacks", versionHandler.HandleGetRollbackHistory).Methods("GET", "OPTIONS")

	// Phase 3: Deployment Version Tracking
	api.HandleFunc("/functions/{functionId}/versions/{version}/deployments", versionHandler.HandleListDeployments).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/deployments/{deploymentId}", versionHandler.HandleGetDeployment).Methods("GET", "OPTIONS")

	// Phase 3: Version Lineage
	api.HandleFunc("/functions/{functionId}/versions/{version}/lineage", versionHandler.HandleGetVersionLineage).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/compare", versionHandler.HandleCompareVersions).Methods("GET", "OPTIONS")

	// Phase 3: Service Contract Versioning (internal API)
	api.HandleFunc("/internal/contracts", versionHandler.HandleListServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/{service}", versionHandler.HandleGetServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/negotiate", versionHandler.HandleNegotiateContractVersion).Methods("POST", "OPTIONS")

	// Recommendations routes (public)
	recommendationHandler := recommendations.NewHandler(s.recommendationSvc)
	api.HandleFunc("/recommendations", recommendationHandler.HandleGetRecommendations).Methods("GET", "OPTIONS")
	api.HandleFunc("/recommendations/interactions", recommendationHandler.HandleRecordInteraction).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/executions", recommendationHandler.HandleRecordExecution).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/feedback", recommendationHandler.HandleRecordFeedback).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/refresh", authMiddleware.RequirePermission(auth.PermSystemWrite)(recommendationHandler.HandleRefreshRecommendations)).Methods("POST", "OPTIONS")

	// Function Registry v2 routes (with camelCase response format)
	apiV2 := s.router.PathPrefix("/v2").Subrouter()
	apiV2.HandleFunc("/registry/functions", registryHandler.HandleListFunctions).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")
	apiV2.HandleFunc("/registry/search", registryHandler.HandleSearchFunctions).Methods("GET")

	// Canary deployment routes
	canaryHandler := registryhandler.NewCanaryHandler(
		registry.NewCanaryConfigRepository(s.postgresDB.GORM),
		registryRepo,
	)
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleCreateCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleGetCanary).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleUpdateCanary).Methods("PATCH")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleCancelCanary).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/promote", canaryHandler.HandlePromoteCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/rollback", canaryHandler.HandleRollbackCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/history", canaryHandler.HandleGetCanaryHistory).Methods("GET")

	// Deprecation and migration guide routes
	versionManager := middleware.NewVersionManager()
	deprecationHandler := registryhandler.NewDeprecationHandler(versionManager)
	migrationHandler := registryhandler.NewMigrationHandler()
	api.HandleFunc("/registry/deprecations", deprecationHandler.HandleGetAllDeprecations).Methods("GET")
	api.HandleFunc("/registry/deprecations/{endpoint}", deprecationHandler.HandleGetEndpointDeprecation).Methods("GET")
	api.HandleFunc("/registry/migration-guide", migrationHandler.HandleGetMigrationGuide).Methods("GET")
	api.HandleFunc("/registry/migration-guide/{endpoint}", migrationHandler.HandleGetEndpointMigration).Methods("GET")
	api.HandleFunc("/registry/versions", migrationHandler.HandleGetVersionInfo).Methods("GET")

	// Documentation routes (public)
	api.HandleFunc("/docs", docsHandler.HandleIndex).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}", docsHandler.HandleFunctionHTMLDocs).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/versions", docsHandler.HandleFunctionVersions).Methods("GET")
	api.HandleFunc("/docs/openapi.json", docsHandler.HandleOpenAPISpec).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/api", docsHandler.HandleFunctionDocs).Methods("GET")

	// Playground routes (public) - Legacy registry routes
	api.HandleFunc("/playground/{author}/{name}", registryPlaygroundHandler.HandlePlaygroundUI).Methods("GET")
	api.HandleFunc("/playground/{author}/{name}/execute", registryPlaygroundHandler.HandlePlaygroundExecute).Methods("POST")
	api.HandleFunc("/playground/{author}/{name}/share", registryPlaygroundHandler.HandlePlaygroundShare).Methods("POST")

	// Tutorials routes (public)
	api.HandleFunc("/tutorials", tutorialsHandler.HandleIndex).Methods("GET")
	api.HandleFunc("/tutorials/getting-started", tutorialsHandler.HandleGettingStarted).Methods("GET")
	api.HandleFunc("/tutorials/api-usage", tutorialsHandler.HandleAPIUsage).Methods("GET")
	api.HandleFunc("/tutorials/function-development", tutorialsHandler.HandleFunctionDevelopment).Methods("GET")
	api.HandleFunc("/tutorials/examples/{example}", tutorialsHandler.HandleInteractiveExample).Methods("GET")

	// CDN Static Asset routes (public)
	api.HandleFunc("/sdk/{sdk}/{version}/{filename}", registryHandler.HandleServeSDK).Methods("GET")
	api.HandleFunc("/docs/{type}/{version}/{path}", registryHandler.HandleServeDocs).Methods("GET")
	api.HandleFunc("/static/{category}/{path}", registryHandler.HandleServeStatic).Methods("GET")

	// Function Embed routes (public)
	// Serves a self-contained JavaScript embed script for any registered function.
	// Supports optional version pinning via the "@version" suffix in the filename.
	//   GET /v1/embed/{author}/{nameVersion}   e.g. /v1/embed/acme/slugify.js
	//                                               /v1/embed/acme/slugify@1.2.0.js
	api.HandleFunc("/embed/{author}/{nameVersion}", registryHandler.HandleServeEmbed).Methods("GET", "OPTIONS")

	// Embed configuration & dashboard routes (public read, protected write)
	api.HandleFunc("/registry/functions/{author}/{name}/embed", registryHandler.HandleGetEmbedConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed/snippet", registryHandler.HandleGetEmbedSnippet).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed/analytics", registryHandler.HandleGetEmbedAnalytics).Methods("GET", "OPTIONS")
	// Update embed config requires authentication (function owner only)
	api.HandleFunc("/registry/functions/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleUpdateEmbedConfig)).Methods("PUT", "OPTIONS")

	// Cache monitoring routes (public for metrics)
	api.HandleFunc("/cache/stats", registryHandler.HandleGetCacheStats).Methods("GET")

	// Execute function (public) with security middleware
	secureExecuteHandler := func(w http.ResponseWriter, r *http.Request) {
		registryRepo := registry.NewRegistryRepository(s.postgresDB.GORM, nil)
		vars := mux.Vars(r)
		author := vars["author"]
		name := vars["name"]
		version := vars["version"]

		// Get function to extract ID for security middleware
		fn, err := registryRepo.GetFunctionByAuthorName(author, name)
		if err != nil {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}

		// Apply execution security middleware
		securityHandler := executionSecurityMW.SecureExecution(fn.ID, version)(registryHandler.HandleExecute)
		securityHandler.ServeHTTP(w, r)
	}

	// Apply verification middleware on top of security middleware if enabled
	if verificationMiddleware != nil {
		api.Handle("/fx/{author}/{name}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
		api.Handle("/fx/{author}/{name}@{version}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
	} else {
		api.HandleFunc("/fx/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/fx/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")
	}

	// Publish function (protected)
	api.HandleFunc("/registry/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST")

	// Function stats (public)
	api.HandleFunc("/registry/functions/{author}/{name}/stats", registryHandler.HandleGetFunctionStats).Methods("GET")

	// Function test (public)
	api.HandleFunc("/registry/functions/{author}/{name}/test", registryHandler.HandleTest).Methods("POST")

	// Submit rating (protected)
	api.HandleFunc("/registry/functions/{author}/{name}/rating", authMiddleware.RequireAuth(registryHandler.HandleSubmitRating)).Methods("POST")

	// Aggregate stats (protected - for admin/owner)
	api.HandleFunc("/registry/functions/{author}/{name}/aggregate", authMiddleware.RequireAuth(registryHandler.HandleAggregateStats)).Methods("POST")

	// Registry replay routes (public)
	api.HandleFunc("/registry/replay/{execId}", registryHandler.HandleGetReplay).Methods("GET")

	// DRE 2.0 routes (public — certificates and passports are public artifacts)
	dreHandler := drehandler.NewHandler(registryRepo)
	api.HandleFunc("/registry/{author}/{name}/cert/{cert_id}", dreHandler.HandleGetCertificate).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/certs", dreHandler.HandleListCertificates).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/replay/{execution_id}", dreHandler.HandleReplay).Methods("POST", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/passport", dreHandler.HandleGetPassport).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/diverge", dreHandler.HandleDivergenceSimulation).Methods("POST", "OPTIONS")

	// Execution Explorer routes (public — execution hashes are public artifacts)
	api.HandleFunc("/registry/{author}/{name}/executions", dreHandler.HandleListExecutions).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/executions/{execution_id}", dreHandler.HandleGetExecution).Methods("GET", "OPTIONS")

	// Execution security routes (public)
	executionSecurityMW.CreateExecutionSecurityRoutes(api)

	// Verification routes (protected)
	api.HandleFunc("/registry/verification/{functionVersionId}/status", authMiddleware.RequireAuth(registryHandler.HandleGetVerificationStatus)).Methods("GET")
	api.HandleFunc("/registry/verification/{functionVersionId}/sign", authMiddleware.RequireAuth(registryHandler.HandleSignFunction)).Methods("POST")
	api.HandleFunc("/registry/verification/signatures/{signatureId}/verify", authMiddleware.RequireAuth(registryHandler.HandleVerifySignature)).Methods("POST")
	api.HandleFunc("/registry/verification/{functionVersionId}/approval", authMiddleware.RequireAuth(registryHandler.HandleRequestApproval)).Methods("POST")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/decide", authMiddleware.RequireAuth(registryHandler.HandleMakeApprovalDecision)).Methods("POST")
	api.HandleFunc("/registry/verification/{functionVersionId}/approvals", authMiddleware.RequireAuth(registryHandler.HandleGetApprovals)).Methods("GET")
	api.HandleFunc("/registry/verification/approvals/pending", authMiddleware.RequireAuth(registryHandler.HandleGetPendingApprovals)).Methods("GET")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleAddApprovalComment)).Methods("POST")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleGetApprovalComments)).Methods("GET")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()

	// Team routes (protected)
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

	// Provider routes (protected)
	protected.HandleFunc("/providers", authMiddleware.RequireAuth(providersHandler.HandleListProviders)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/providers/validate", authMiddleware.RequireAuth(providersHandler.HandleValidateProvider)).Methods("POST")
	protected.HandleFunc("/providers/cost-estimate", authMiddleware.RequireAuth(providersHandler.HandleEstimateCost)).Methods("POST")
	protected.HandleFunc("/providers/{providerId}/share", authMiddleware.RequireAuth(providersHandler.HandleShareProvider)).Methods("POST")

	// Team invites during onboarding (protected)
	protected.HandleFunc("/teams/invites", authMiddleware.RequireAuth(providersHandler.HandleCreateTeamInvite)).Methods("POST")

	// App routes (protected)
	protected.HandleFunc("/apps", authMiddleware.RequireAuth(appsHandler.HandleListApps)).Methods("GET")
	protected.HandleFunc("/apps", authMiddleware.RequireAuth(appsHandler.HandleCreateApp)).Methods("POST")
	protected.HandleFunc("/apps/{appId}", authMiddleware.RequireAuth(appsHandler.HandleGetApp)).Methods("GET")

	// Function routes (protected)
	protected.HandleFunc("/functions", authMiddleware.RequireAuth(functionsHandler.HandleListFunctions)).Methods("GET")
	protected.HandleFunc("/functions", authMiddleware.RequireAuth(functionsHandler.HandleCreateFunction)).Methods("POST")
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleGetFunction)).Methods("GET")
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleUpdateFunction)).Methods("PUT")
	protected.HandleFunc("/functions/{id}", authMiddleware.RequireAuth(functionsHandler.HandleDeleteFunction)).Methods("DELETE")
	protected.HandleFunc("/functions/{id}/logs", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionLogs)).Methods("GET")
	protected.HandleFunc("/functions/{id}/deployments", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionDeployments)).Methods("GET")
	protected.HandleFunc("/functions/deployments/{deploymentId}", authMiddleware.RequireAuth(functionsHandler.HandleGetFunctionDeployment)).Methods("GET")
	protected.HandleFunc("/functions/deploy", authMiddleware.RequireAuth(functionsHandler.HandleDeployFunction)).Methods("POST")
	protected.HandleFunc("/functions/test", authMiddleware.RequireAuth(functionsHandler.HandleTestFunction)).Methods("POST")

	// Dashboard routes (protected, tenant-scoped)
	protected.HandleFunc("/dashboard/usage", authMiddleware.RequireAuth(dashboardHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/execution-rate", authMiddleware.RequireAuth(dashboardHandler.HandleGetExecutionRate)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/activity", authMiddleware.RequireAuth(dashboardHandler.HandleGetActivity)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dashboard/metrics", authMiddleware.RequireAuth(dashboardHandler.HandleGetMetrics)).Methods("GET", "OPTIONS")

	// Enterprise SLA routes (protected, enterprise plan required)
	protected.HandleFunc("/enterprise/sla/overview", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetSLAOverview)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/enterprise/sla/uptime-history", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetUptimeHistory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/enterprise/sla/incidents", authMiddleware.RequireAuth(enterpriseSLAHandler.HandleGetIncidents)).Methods("GET", "OPTIONS")

	// StateFabric routes (protected) - with per-tenant rate limiting
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

	// Agent Memory routes (State Fabric - AI memory with embeddings)
	protected.HandleFunc("/memories", authMiddleware.RequireAuth(memoryHandler.HandleListMemories)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/memories", authMiddleware.RequireAuth(memoryHandler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/memories/search", authMiddleware.RequireAuth(memoryHandler.HandleSearchMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleGetMemory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleUpdateMemory)).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/memories/{id}", authMiddleware.RequireAuth(memoryHandler.HandleDeleteMemory)).Methods("DELETE", "OPTIONS")

	// Agent Memory REST API routes (new handler at /agent-memories)
	protected.HandleFunc("/agent-memories", authMiddleware.RequireAuth(agentMemoryHandler.HandleListMemories)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent-memories", authMiddleware.RequireAuth(agentMemoryHandler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/search", authMiddleware.RequireAuth(agentMemoryHandler.HandleSearchMemories)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/index", authMiddleware.RequireAuth(agentMemoryHandler.HandleRebuildIndex)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}", authMiddleware.RequireAuth(agentMemoryHandler.HandleGetMemory)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent-memories/{id}", authMiddleware.RequireAuth(agentMemoryHandler.HandleDeleteMemory)).Methods("DELETE", "OPTIONS")

	// State Fabric routes (dashboard State Fabric feature) - with per-tenant rate limiting
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

	// ============================================================
	// Secrets Vault routes (protected)
	// ============================================================
	protected.HandleFunc("/vault/secrets", authMiddleware.RequireAuth(vaultHandler.HandleListSecrets)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets", authMiddleware.RequireAuth(vaultHandler.HandleCreateSecret)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleGetSecret)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleUpdateSecret)).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}", authMiddleware.RequireAuth(vaultHandler.HandleDeleteSecret)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/tokens", authMiddleware.RequireAuth(vaultHandler.HandleListTokens)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/vault/secrets/{id}/tokens", authMiddleware.RequireAuth(vaultHandler.HandleGenerateToken)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/vault/tokens/{id}", authMiddleware.RequireAuth(vaultHandler.HandleRevokeToken)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/vault/audit", authMiddleware.RequireAuth(vaultHandler.HandleGetAuditLog)).Methods("GET", "OPTIONS")

	// ============================================================
	// AEP (Agent Execution Plan) routes
	// ============================================================

	// Discovery (public - no auth required for discovery)
	api.HandleFunc("/agent/discover", aepHandler.HandleDiscover).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent/discover/{author}/{name}", aepHandler.HandleDiscoverFunction).Methods("GET", "OPTIONS")

	// Agent execution (supports both X-Agent-API-Key and JWT auth)
	api.HandleFunc("/agent/execute/{author}/{name}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent/execute/{author}/{name}/{version}", aepHandler.HandleExecute).Methods("POST", "OPTIONS")

	// Agent management (protected - JWT required)
	protected.HandleFunc("/agent/register", authMiddleware.RequireAuth(aepHandler.HandleRegisterAgent)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent", authMiddleware.RequireAuth(aepHandler.HandleListAgents)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(aepHandler.HandleGetAgent)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}", authMiddleware.RequireAuth(aepHandler.HandleDeleteAgent)).Methods("DELETE", "OPTIONS")

	// Quota management (protected)
	protected.HandleFunc("/agent/{agent_id}/quota", authMiddleware.RequireAuth(aepHandler.HandleUpdateQuota)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/usage", authMiddleware.RequireAuth(aepHandler.HandleGetUsage)).Methods("GET", "OPTIONS")

	// Policy management (protected)
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(aepHandler.HandleGetPolicy)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/policy", authMiddleware.RequireAuth(aepHandler.HandleUpdatePolicy)).Methods("PUT", "OPTIONS")

	// Attribution & observability (protected)
	protected.HandleFunc("/agent/{agent_id}/executions", authMiddleware.RequireAuth(aepHandler.HandleListExecutions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/executions/{exec_id}", authMiddleware.RequireAuth(aepHandler.HandleGetExecution)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/analytics", authMiddleware.RequireAuth(aepHandler.HandleGetAnalytics)).Methods("GET", "OPTIONS")

	// Session management (protected)
	protected.HandleFunc("/agent/{agent_id}/session/start", authMiddleware.RequireAuth(aepHandler.HandleStartSession)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}/end", authMiddleware.RequireAuth(aepHandler.HandleEndSession)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/session/{session_id}", authMiddleware.RequireAuth(aepHandler.HandleGetSession)).Methods("GET", "OPTIONS")

	// Billing & economic controls (protected)
	protected.HandleFunc("/agent/{agent_id}/billing/summary", authMiddleware.RequireAuth(aepHandler.HandleGetBillingSummary)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/billing/spend-cap", authMiddleware.RequireAuth(aepHandler.HandleUpdateSpendCap)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/cost-breakdown", authMiddleware.RequireAuth(aepHandler.HandleGetCostBreakdown)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/credits/balance", authMiddleware.RequireAuth(aepHandler.HandleGetCreditBalance)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/agent/{agent_id}/credits/purchase", authMiddleware.RequireAuth(aepHandler.HandlePurchaseCredits)).Methods("POST", "OPTIONS")

	// Concurrency stats (protected)
	protected.HandleFunc("/agent/concurrency/stats", authMiddleware.RequireAuth(aepHandler.HandleGetConcurrencyStats)).Methods("GET", "OPTIONS")

	// ============================================================
	// Flywheel Network routes (Proof-of-Execution Knowledge Network)
	// ============================================================
	// ============================================================
	// Flywheel Network components (Proof-of-Execution Knowledge Network)
	// ============================================================
	flywheelRepo := flywheel.NewRepository(s.postgresDB.GORM)
	flywheelExecService := flywheel.NewExecutionAdapter(registryRepo, cacheService, registryexecution.NewLocalExecutor(), logrus.New())
	flywheelService := flywheel.NewService(flywheelRepo, flywheelExecService, logrus.New())

	// Initialize Flywheel WebSocket hub for real-time updates
	flywheelWSHub := flywheelhandler.NewWebSocketHub(logrus.New())
	go flywheelWSHub.Run()

	flywheelHandler := flywheelhandler.NewHandler(flywheelService, flywheelWSHub, logrus.New())

	// Categories (public)
	api.HandleFunc("/flywheel/categories", flywheelHandler.ListCategories).Methods("GET", "OPTIONS")

	// Threads (public read, protected write)
	api.HandleFunc("/flywheel/threads", flywheelHandler.ListThreads).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads", authMiddleware.RequireAuth(flywheelHandler.CreateThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}", flywheelHandler.GetThread).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}", authMiddleware.RequireAuth(flywheelHandler.UpdateThread)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/resolve", authMiddleware.RequireAuth(flywheelHandler.ResolveThread)).Methods("POST", "OPTIONS")

	// Thread subscriptions (protected)
	api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.SubscribeToThread)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/subscribe", authMiddleware.RequireAuth(flywheelHandler.UnsubscribeFromThread)).Methods("DELETE", "OPTIONS")

	// Replies (public read, protected write)
	api.HandleFunc("/flywheel/threads/{id}/replies", flywheelHandler.ListReplies).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/replies", authMiddleware.RequireAuth(flywheelHandler.CreateReply)).Methods("POST", "OPTIONS")

	// Reply execution and verification (protected)
	api.HandleFunc("/flywheel/replies/{id}/execute", authMiddleware.RequireAuth(flywheelHandler.ExecuteReply)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/replies/{id}/verify", authMiddleware.RequireAuth(flywheelHandler.VerifyReply)).Methods("POST", "OPTIONS")

	// Reputation (public read)
	api.HandleFunc("/flywheel/reputation/me", authMiddleware.RequireAuth(flywheelHandler.GetMyReputation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/reputation/{user_id}", flywheelHandler.GetUserReputation).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/leaderboards/{score_type}", flywheelHandler.GetLeaderboard).Methods("GET", "OPTIONS")

	// Challenges (public read, protected write/submit)
	api.HandleFunc("/flywheel/challenges", flywheelHandler.ListChallenges).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}", flywheelHandler.GetChallenge).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}/submit", authMiddleware.RequireAuth(flywheelHandler.SubmitChallenge)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/challenges/{id}/leaderboard", flywheelHandler.GetChallengeLeaderboard).Methods("GET", "OPTIONS")

	// WebSocket endpoint for real-time updates
	api.HandleFunc("/flywheel/ws", flywheelHandler.HandleWebSocket)

	// Search functionality
	api.HandleFunc("/flywheel/search", flywheelHandler.Search).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/solutions/verified", flywheelHandler.ListVerifiedSolutions).Methods("GET", "OPTIONS")

	// Thread replay and timeline
	api.HandleFunc("/flywheel/threads/{id}/timeline", flywheelHandler.GetThreadTimeline).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/replay", flywheelHandler.ReplayThread).Methods("POST", "OPTIONS")

	// Agent collaboration (protected)
	api.HandleFunc("/flywheel/threads/{id}/agents", authMiddleware.RequireAuth(flywheelHandler.ListThreadAgents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/invite", authMiddleware.RequireAuth(flywheelHandler.InviteAgent)).Methods("POST", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}", authMiddleware.RequireAuth(flywheelHandler.RemoveAgent)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/flywheel/threads/{id}/agents/{agent_id}/respond", authMiddleware.RequireAuth(flywheelHandler.AgentRespond)).Methods("POST", "OPTIONS")

	// Marketplace integration (protected)
	api.HandleFunc("/flywheel/replies/{id}/publish-to-marketplace", authMiddleware.RequireAuth(flywheelHandler.PublishToMarketplace)).Methods("POST", "OPTIONS")

	// ============================================================
	// Swarm / Marketplace / Evolution routes (protected)
	// ============================================================
	// Register swarm handler routes (basePath "" because protected is already under api /v1)
	swarmHandler.RegisterRoutes(protected, "")

	// Backend routes (protected)
	protected.HandleFunc("/apps/{appId}/backends", authMiddleware.RequireAuth(backendsHandler.HandleCreateBackend)).Methods("POST")
	protected.HandleFunc("/apps/{appId}/backends", authMiddleware.RequireAuth(backendsHandler.HandleListBackends)).Methods("GET")
	protected.HandleFunc("/apps/{appId}/status", authMiddleware.RequireAuth(appsHandler.HandleGetAppStatus)).Methods("GET")

	// Advanced deployment routes (protected)
	protected.HandleFunc("/apps/{appId}/deploy/blue-green", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(backendsHandler.HandleDeployBlueGreen))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/link", authMiddleware.RequireAuth(backendsHandler.HandleLinkProject)).Methods("POST")
	protected.HandleFunc("/apps/{appId}/secrets", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(backendsHandler.HandleSetSecrets))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/secrets", authMiddleware.RequireAuth(backendsHandler.HandleListSecrets)).Methods("GET")

	// Routing routes (protected)
	protected.HandleFunc("/apps/{appId}/route", authMiddleware.RequireAuth(backendsHandler.HandleGetRoute)).Methods("GET")

	// Monitoring management routes
	protected.HandleFunc("/monitoring/alerts/{alertId}/resolve", authMiddleware.RequireAuth(monitoringHandler.HandleResolveAlert)).Methods("POST")
	protected.HandleFunc("/monitoring/dashboard", authMiddleware.RequireAuth(monitoringHandler.HandleGetDashboardConfig)).Methods("GET")

	// Deployment routes (protected with HMAC for sensitive operations)
	protected.HandleFunc("/apps/{appId}/deploy", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(deploymentsHandler.HandleDeploy))).Methods("POST")
	protected.HandleFunc("/apps/{appId}/deployments", authMiddleware.RequireAuth(deploymentsHandler.HandleListDeployments)).Methods("GET")
	protected.HandleFunc("/deployments/{deploymentId}", authMiddleware.RequireAuth(deploymentsHandler.HandleGetDeployment)).Methods("GET")
	protected.HandleFunc("/deployments/{deploymentId}/rollback", authMiddleware.RequireAuth(advancedSecurityMiddleware.RequireHMACSignature(deploymentsHandler.HandleRollback))).Methods("POST")

	// Health check (public)
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/detailed", s.handleDetailedHealth).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/health/check", s.handleHealthCheck).Methods("GET", "OPTIONS")

	// Maintenance status (public)
	s.router.HandleFunc("/maintenance/status", maintenanceHandler.HandleGetPublicStatus).Methods("GET", "OPTIONS")

	// Status WebSocket endpoint (public - for real-time status updates)
	// WS /ws/v1/status - Real-time status updates
	statusWSHub := statushandler.NewStatusWebSocketHub(statusHandler, logrus.New())
	go statusWSHub.Run()
	s.router.HandleFunc("/ws/v1/status", statusHandler.HandleWebSocket(statusWSHub)).Methods("GET")

	// Prometheus metrics endpoint (public for scraping)
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Well-known AI discovery (public, no auth) — must be before SPA catch-all
	wellknownHandler := wellknown.NewHandler(registryRepo)
	s.router.HandleFunc("/.well-known/functionfly.json", wellknownHandler.HandleWellKnown).Methods("GET", "OPTIONS")

	// Admin routes (protected with RBAC and MFA for admin users)
	adminRoutes := api.PathPrefix("/admin").Subrouter()

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

	// Admin oversight routes (trust dashboard, execution audit, fraud detection, economic leaderboard)
	adminRoutes.HandleFunc("/oversight/trust-dashboard", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetTrustDashboard)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/execution-audit", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetExecutionAudit)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/fraud-detection", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetFraudDetection)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/economic-leaderboard", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleGetEconomicLeaderboard)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/block/{type}/{id}", authMiddleware.RequirePermission(auth.PermSystemWrite)(oversightHandler.HandleBlockEntity)).Methods("POST", "OPTIONS")
	adminRoutes.HandleFunc("/oversight/investigate/{type}/{id}", authMiddleware.RequirePermission(auth.PermSystemRead)(oversightHandler.HandleInvestigateEntity)).Methods("POST", "OPTIONS")

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

	// Playground SPA routes: /fx/*, /run/*, /replay/*
	// Serve the SPA index.html for these reserved paths
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		// Don't match API routes
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 1 && (pathParts[0] == "fx" || pathParts[0] == "run" || pathParts[0] == "replay")
	}).HandlerFunc(s.serveSPAIndex)

	// Public routing endpoint: /{appSlug}/*
	// Use a custom matcher to avoid conflicting with API routes
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		// Don't match API routes or reserved paths
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/v1/") {
			return false
		}
		// Match app slug pattern: /someSlug/somePath
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 2 && pathParts[0] != "" && pathParts[0] != "api" && pathParts[0] != "content" && pathParts[0] != "health" &&
			pathParts[0] != "fx" && pathParts[0] != "run" && pathParts[0] != "replay"
	}).HandlerFunc(s.handlePublicRoute)
}

// initializeGenerationServiceWithCache creates a generation service with Redis-backed cache
// when Redis is available and configured. Falls back to in-memory cache otherwise.
func initializeGenerationServiceWithCache(db *gorm.DB, redisClient *redis.Client) *generation.Service {
	// Check if we should use Redis for generation cache
	useRedisCache := os.Getenv("GENERATION_CACHE_REDIS_ENABLED") == "true"

	// Also check if OPENROUTER_API_KEY is set - if not, there's no point using LLM
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {
		// No API key, use template-based generation
		return generation.NewService(db)
	}

	// Determine if we should use Redis cache
	// Default to true if Redis client is available and not explicitly disabled
	if !useRedisCache && redisClient != nil {
		useRedisCache = os.Getenv("GENERATION_CACHE_REDIS_ENABLED") != "false"
	}

	var codeGen generation.CodeGenerator

	if useRedisCache && redisClient != nil {
		// Use Redis-backed OpenRouter client
		logrus.Info("Initializing generation service with Redis-backed cache")
		codeGen = generation.NewOpenRouterClientWithRedis(apiKey, nil, redisClient, true, nil)
	} else {
		// Use in-memory cache
		logrus.Info("Initializing generation service with in-memory cache")
		codeGen = generation.NewOpenRouterClient(apiKey, nil, nil, nil)
	}

	return generation.NewServiceWithGenerator(db, codeGen)
}
