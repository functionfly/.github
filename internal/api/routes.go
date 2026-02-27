package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/api/handlers/admin"
	agenthandler "github.com/functionfly/functionfly/internal/api/handlers/agent"
	"github.com/functionfly/functionfly/internal/api/handlers/apps"
	authHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/auth"
	"github.com/functionfly/functionfly/internal/api/handlers/backends"
	"github.com/functionfly/functionfly/internal/api/handlers/content"
	"github.com/functionfly/functionfly/internal/api/handlers/deployments"
	feedbackHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/feedback"
	"github.com/functionfly/functionfly/internal/api/handlers/functions"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	"github.com/functionfly/functionfly/internal/api/handlers/monitoring"
	"github.com/functionfly/functionfly/internal/api/handlers/playground"
	"github.com/functionfly/functionfly/internal/api/handlers/providers"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	drehandler "github.com/functionfly/functionfly/internal/api/handlers/registry/dre"
	"github.com/functionfly/functionfly/internal/api/handlers/security"
	"github.com/functionfly/functionfly/internal/api/handlers/state"
	"github.com/functionfly/functionfly/internal/api/handlers/teams"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/captcha"
	monitoringPkg "github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage/registry"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// setupRoutes configures all API routes
func (s *Server) setupRoutes(realtimeMonitor *monitoringPkg.RealtimeMonitor) {
	// Initialize handlers
	authHandler := authHandlerPkg.NewHandler(s.authSvc)
	appsHandler := apps.NewHandler(s.repo)
	backendsHandler := backends.NewHandler(s.repo, s.routingSvc)
	deploymentsHandler := deployments.NewHandler(s.repo, s.deploySvc)
	functionsHandler := functions.NewHandler(s.repo, s.deploySvc)
	adminHandler := admin.NewHandler(s.repo, s.authSvc)
	securityHandler := security.NewHandler(s.repo, s.authSvc)
	contentHandler := content.NewHandler(s.repo)
	feedbackHandler := feedbackHandlerPkg.NewHandler(s.repo, s.storageService)
	// Initialize monitoring handler
	monitoringHandler := monitoring.NewHandler(s.repo, s.monitoringSvc, s.realtimeMonitor, s.authSvc)
	mfaHandler := mfaHandlerPkg.NewMFAHandler(s.authSvc)

	// Initialize app-based playground handler
	appPlaygroundHandler := playground.NewHandler(s.repo)

	// Initialize cache configuration from environment
	cacheConfiguration := cache.LoadCacheConfiguration()
	if err := cacheConfiguration.Validate(); err != nil {
		logrus.WithError(err).Error("Invalid cache configuration")
		panic("Invalid cache configuration: " + err.Error())
	}

	// Initialize cache service with comprehensive configuration
	cacheService, err := cache.NewCacheService(s.postgresDB.GORM, s.redisClient, cacheConfiguration.ToCacheConfig())
	if err != nil {
		// Handle cache service initialization error
		panic("Failed to initialize cache service: " + err.Error())
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
	adminRegistryHandler := admin.NewRegistryHandler(registryRepo)

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

	// Initialize state handler
	stateRepo := staterepo.NewStateRepository(s.postgresDB.GORM)
	stateHandler := state.NewHandler(stateRepo)

	// Initialize AEP (Agent Execution Plan) handler
	aepHandler := agenthandler.NewHandler(s.postgresDB.GORM, s.redisClient, registryRepo)

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

	// Auth routes (public)
	api.HandleFunc("/auth/login", authRateLimiter.Limit(authHandler.HandleLogin)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/signup", authRateLimiter.Limit(authHandler.HandleSignup)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/verify-email", authHandler.HandleVerifyEmail).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/resend-verification", authRateLimiter.Limit(authHandler.HandleResendVerification)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/get-session", authHandler.HandleGetSession).Methods("GET", "OPTIONS")

	// OAuth routes (public)
	api.HandleFunc("/auth/oauth/providers", authHandler.HandleGetOAuthProviders).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/oauth/url", authHandler.HandleGetOAuthURL).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/oauth/{provider}/callback", authHandler.HandleOAuthCallback).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/validate", authMiddleware.RequireAuth(authHandler.HandleValidateToken)).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/logout", authMiddleware.RequireAuth(authHandler.HandleLogout)).Methods("POST", "OPTIONS")

	// MFA routes (protected)
	api.HandleFunc("/auth/mfa/setup", authMiddleware.RequireAuth(mfaHandler.SetupMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/verify", authMiddleware.RequireAuth(mfaHandler.VerifyMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/enable", authMiddleware.RequireAuth(mfaHandler.EnableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/disable", authMiddleware.RequireAuth(mfaHandler.DisableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/status", authMiddleware.RequireAuth(mfaHandler.GetMFAStatus)).Methods("GET", "OPTIONS")

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
	api.HandleFunc("/registry/search", registryHandler.HandleSearchFunctions).Methods("GET")

	// Function Registry v2 routes (with camelCase response format)
	apiV2 := s.router.PathPrefix("/v2").Subrouter()
	apiV2.HandleFunc("/registry/functions", registryHandler.HandleListFunctions).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
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
	protected.HandleFunc("/providers/validate", authMiddleware.RequireAuth(providersHandler.HandleValidateProvider)).Methods("POST")
	protected.HandleFunc("/providers/cost-estimate", authMiddleware.RequireAuth(providersHandler.HandleEstimateCost)).Methods("POST")
	protected.HandleFunc("/providers/{providerId}/share", authMiddleware.RequireAuth(providersHandler.HandleShareProvider)).Methods("POST")

	// Team invites during onboarding (protected)
	protected.HandleFunc("/teams/invites", authMiddleware.RequireAuth(providersHandler.HandleCreateTeamInvite)).Methods("POST")

	// App routes (protected)
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

	// StateFabric routes (protected)
	protected.HandleFunc("/state", authMiddleware.RequireAuth(stateHandler.HandleListStates)).Methods("GET")
	protected.HandleFunc("/state", authMiddleware.RequireAuth(stateHandler.HandleCreateState)).Methods("POST")
	protected.HandleFunc("/state/{path}", authMiddleware.RequireAuth(stateHandler.HandleGetState)).Methods("GET")
	protected.HandleFunc("/state/{path}", authMiddleware.RequireAuth(stateHandler.HandleUpdateState)).Methods("PUT")
	protected.HandleFunc("/state/{path}", authMiddleware.RequireAuth(stateHandler.HandleDeleteState)).Methods("DELETE")
	protected.HandleFunc("/state/{path}/value", authMiddleware.RequireAuth(stateHandler.HandleSetValue)).Methods("PUT")
	protected.HandleFunc("/state/{path}/value", authMiddleware.RequireAuth(stateHandler.HandlePatchValue)).Methods("PATCH")
	protected.HandleFunc("/state/{path}/value", authMiddleware.RequireAuth(stateHandler.HandleGetValue)).Methods("GET")
	protected.HandleFunc("/state/{path}/value", authMiddleware.RequireAuth(stateHandler.HandleDeleteValue)).Methods("DELETE")
	protected.HandleFunc("/state/{path}/history", authMiddleware.RequireAuth(stateHandler.HandleGetHistory)).Methods("GET")
	protected.HandleFunc("/state/{path}/snapshot", authMiddleware.RequireAuth(stateHandler.HandleCreateSnapshot)).Methods("POST")
	protected.HandleFunc("/state/{path}/snapshots", authMiddleware.RequireAuth(stateHandler.HandleListSnapshots)).Methods("GET")
	protected.HandleFunc("/state/{path}/restore", authMiddleware.RequireAuth(stateHandler.HandleRestoreSnapshot)).Methods("POST")
	protected.HandleFunc("/state/{path}/time-travel", authMiddleware.RequireAuth(stateHandler.HandleTimeTravel)).Methods("GET")
	protected.HandleFunc("/state/{path}/permissions", authMiddleware.RequireAuth(stateHandler.HandleGrantPermission)).Methods("POST")
	protected.HandleFunc("/state/{path}/permissions", authMiddleware.RequireAuth(stateHandler.HandleGetPermissions)).Methods("GET")
	protected.HandleFunc("/triggers", authMiddleware.RequireAuth(stateHandler.HandleGetTriggers)).Methods("GET")
	protected.HandleFunc("/triggers", authMiddleware.RequireAuth(stateHandler.HandleCreateTrigger)).Methods("POST")
	protected.HandleFunc("/triggers/{id}", authMiddleware.RequireAuth(stateHandler.HandleDeleteTrigger)).Methods("DELETE")

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
	protected.HandleFunc("/agent/{agent_id}/credits/balance", authMiddleware.RequireAuth(aepHandler.HandleGetCreditBalance)).Methods("GET", "OPTIONS")

	// Concurrency stats (protected)
	protected.HandleFunc("/agent/concurrency/stats", authMiddleware.RequireAuth(aepHandler.HandleGetConcurrencyStats)).Methods("GET", "OPTIONS")

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

	// Prometheus metrics endpoint (public for scraping)
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Admin routes (protected with RBAC and MFA for admin users)
	adminRoutes := api.PathPrefix("/admin").Subrouter()

	// Note: MFA middleware is applied per-route after auth middleware to ensure claims are available

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

	// Incident management
	adminRoutes.HandleFunc("/incidents", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleListIncidents)).Methods("GET")
	adminRoutes.HandleFunc("/incidents", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleCreateIncident))).Methods("POST")
	adminRoutes.HandleFunc("/incidents/{incidentId}", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleGetIncident)).Methods("GET")
	adminRoutes.HandleFunc("/incidents/{incidentId}", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleUpdateIncident))).Methods("PATCH")
	adminRoutes.HandleFunc("/incidents/{incidentId}/resolve", authMiddleware.RequirePermission(auth.PermSystemWrite)(advancedSecurityMiddleware.RequireHMACSignature(adminHandler.HandleResolveIncident))).Methods("POST")

	// System health
	adminRoutes.HandleFunc("/health", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleSystemHealth)).Methods("GET")

	// Analytics management
	adminRoutes.HandleFunc("/analytics", authMiddleware.RequirePermission(auth.PermSystemRead)(adminHandler.HandleGetAnalyticsSettings)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/analytics", authMiddleware.RequirePermission(auth.PermSystemWrite)(adminHandler.HandleUpdateAnalyticsSettings)).Methods("PATCH", "OPTIONS")

	// Billing management
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
	adminRoutes.HandleFunc("/security/services", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetServiceStatus)).Methods("GET")
	adminRoutes.HandleFunc("/security/certificates", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSSLCertificates)).Methods("GET")
	adminRoutes.HandleFunc("/security/incidents", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetRecentIncidents)).Methods("GET")
	adminRoutes.HandleFunc("/security/compliance", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetComplianceFrameworks)).Methods("GET")
	adminRoutes.HandleFunc("/security/measures", authMiddleware.RequirePermission(auth.PermSystemRead)(securityHandler.HandleGetSecurityMeasures)).Methods("GET")
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

	// Content sync
	contentRoutes.HandleFunc("/sync/github-releases", authMiddleware.RequirePermission(auth.PermSystemWrite)(contentHandler.HandleSyncGitHubReleases)).Methods("POST")

	// Tenant-scoped operations (admin impersonating tenant)
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantApps)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleGetTenantApp)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/backends", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantBackends)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/deployments", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleListTenantDeployments)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/apps/{appId}/deployments/{deploymentId}/rollback", authMiddleware.RequirePermission(auth.PermDeploymentsWrite)(adminHandler.HandleTenantDeploymentRollback)).Methods("POST")

	// Tenant-scoped observability
	adminRoutes.HandleFunc("/tenants/{tenantId}/metrics", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleTenantMetrics)).Methods("GET")
	adminRoutes.HandleFunc("/tenants/{tenantId}/health", authMiddleware.RequirePermission(auth.PermTenantsRead)(adminHandler.HandleTenantHealth)).Methods("GET")

	// Playground SPA routes: /fx/*, /run/*, /replay/*
	// Serve the SPA index.html for these reserved paths
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		// Don't match API routes
		if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" {
			return false
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 1 && (pathParts[0] == "fx" || pathParts[0] == "run" || pathParts[0] == "replay")
	}).HandlerFunc(s.serveSPAIndex)

	// Public routing endpoint: /{appSlug}/*
	// Use a custom matcher to avoid conflicting with API routes
	s.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		// Don't match API routes
		if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/content/") || strings.HasPrefix(r.URL.Path, "/admin/") || r.URL.Path == "/health" {
			return false
		}
		// Match app slug pattern: /someSlug/somePath
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		return len(pathParts) >= 2 && pathParts[0] != "" && pathParts[0] != "v1" && pathParts[0] != "content" && pathParts[0] != "health" &&
			pathParts[0] != "fx" && pathParts[0] != "run" && pathParts[0] != "replay"
	}).HandlerFunc(s.handlePublicRoute)
}
