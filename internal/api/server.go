package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/agent/browser"
	"github.com/functionfly/functionfly/internal/adapters/aws"
	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/deno"
	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/adapters/functionfly"
	"github.com/functionfly/functionfly/internal/adapters/vercel"
	"github.com/functionfly/functionfly/internal/analytics/unified"
	dnahandler "github.com/functionfly/functionfly/internal/api/handlers/dna"
	"github.com/functionfly/functionfly/internal/api/handlers/billing"
	"github.com/functionfly/functionfly/internal/api/handlers/notifications"
	regexec "github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/handlers/trustapi"
	"github.com/functionfly/functionfly/internal/auth"
	billingpkg "github.com/functionfly/functionfly/internal/billing"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/deployment"
	"github.com/functionfly/functionfly/internal/dna"
	dnaStorage "github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/health"
	"github.com/functionfly/functionfly/internal/logging"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/provisioning"
	"github.com/functionfly/functionfly/internal/recommendations"
	"github.com/functionfly/functionfly/internal/routing"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	statefabricrepo "github.com/functionfly/functionfly/internal/storage/statefabric"
	trustapirepo "github.com/functionfly/functionfly/internal/storage/trustapi"
	vaultstorage "github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Server struct {
	// Database
	postgresDB *storage.PostgresDB

	// Logger is the server's structured logger (initialized in NewServer).
	// Routes, handlers, and background workers should use this rather than the
	// package-level logrus.* helpers, so log output flows through the configured
	// format/level and is shared across the process.
	logger *logrus.Logger

	// Repository (unified interface)
	repo storage.Repository

	router              *mux.Router
	authSvc             *auth.AuthService
	routingSvc          *routing.Router
	deploySvc           *deployment.Orchestrator
	monitoringSvc       *monitoring.Service
	realtimeMonitor     *monitoring.RealtimeMonitor
	storageService      *services.StorageService
	sessionCleanup      *storage.SessionCleanupService
	oauthStateCleanup   *storage.OAuthStateCleanupService
	loginAttemptCleanup *storage.LoginAttemptCleanupService
	authEventCleanup    *storage.AuthEventCleanupService
	healthMonitor       *health.Monitor
	redisClient         *redis.Client
	upstashRedis        *cache.UpstashRedisClient
	httpServer          *http.Server
	shutdownTimeout     time.Duration

	// Server lifecycle context - cancelled on shutdown to signal background workers
	serverCtx    context.Context
	serverCancel context.CancelFunc

	// WaitGroup for tracking background worker goroutines (pointer to avoid copy)
	wg *sync.WaitGroup

	// Notification service
	notificationSvc       *notification.Service
	notificationRepo      notification.Repository
	notificationPool      *pgxpool.Pool // dedicated pool for LISTEN connections
	notificationWSHandler *notifications.WebSocketHandler

	// Trigger engine for state changes
	triggerEngine *staterepo.TriggerEngine

	// Execution log cleanup service
	executionLogCleanup *storage.ExecutionLogCleanupService

	// State cleanup service for TTL-based cleanup
	stateCleanup *staterepo.CleanupService

	// State fabric cleanup service for TTL-based cleanup
	stateFabricCleanup *statefabricrepo.CleanupService

	// Recommendations service
	recommendationSvc *recommendations.Service

	// Usage metrics aggregation service
	usageMetricsAgg *services.AggregationService

	// State usage aggregator for billing/quota integration
	stateUsageAggregator *services.StateUsageAggregator

	// Email service for sending emails
	emailSvc email.Service

	// Unified analytics sync job (Phase 3: fills analytics_rollups from source tables)
	unifiedSyncJob *unified.SyncJob

	// Vault repository for token cleanup job (set in setupRoutes)
	stateFabricRepo *statefabricrepo.Repository
	vaultRepo *vaultstorage.Repository

	// Deferred billing checker for Backend-in-a-Box founder mode
	deferredBillingChecker *billing.DeferredBillingChecker

	// Dunning management repository and service for automated payment retry
	dunningRepo        *storage.DunningRepository
	usageReportingRepo *storage.UsageReportingRepository
	dunningManager     *billingpkg.DunningManager

	// Trust API billing service (with usage reporting)
	trustBillingService *trustapi.BillingService

	// Billing sync job for external billing integrations
	billingSyncJob *billingpkg.BillingSyncJob

	// Export scheduler for automated exports
	exportScheduler *services.ExportScheduler

	// Payout webhook processor for Stripe Connect payout events
	payoutWebhookProcessor interface {
		ProcessTransferReversed(ctx context.Context, stripeTransferID string) error
		ProcessPayoutPaid(ctx context.Context, stripePayoutID, stripeAccountID string) error
		RefreshAccountStatus(ctx context.Context, stripeAccountID string) error
	}

	// Bundle provisioning
	bundleProvisioner   *provisioning.BundleProvisioner
	provisioningHandler *provisioning.Handler

	// Certification schedulers (use repo directly, not handler)
	certExamExpiryScheduler *scheduler.CertExamExpiryScheduler
	certCredExpiryScheduler *scheduler.CertCredentialExpiryScheduler

	// Consciousness scheduler for periodic awareness analysis
	consciousnessScheduler        *scheduler.ConsciousnessScheduler
	consciousnessCleanupScheduler *scheduler.CleanupScheduler
	consciousnessRetryScheduler   *scheduler.RetryScheduler

	// DNA service for mutation analysis
	dnaRepo                 *dnaStorage.Repository
	dnaService              *dna.Service
	dnaHandler              *dnahandler.Handler
	dnaPartitionScheduler   *scheduler.DNAPartitionScheduler
	dnaInsightsScheduler    *scheduler.DNAInsightsScheduler
	browserSvc           browser.Browser

	walletService *wallet.Service
}

func NewServer(db *storage.PostgresDB) *Server {
	// Centralized environment validation — captures missing required vars early.
	// We use the singleton logger so this Fatal output is formatted like the rest
	// of the process, and so test code that imports this package can stub it.
	if err := config.ValidateEnv(); err != nil {
		logging.Logger().Fatal(err)
	}

	// SECURITY: Block DEVELOPMENT=true on machines whose hostname looks like a deployed server.
	// Many dev laptops/WSL hosts use a real hostname (not "localhost"), so allow an explicit opt-in.
	if os.Getenv("DEVELOPMENT") == "true" {
		hostname, _ := os.Hostname()
		isLocalhost := strings.HasPrefix(hostname, "localhost") || strings.HasPrefix(hostname, "127.0.0.1") || strings.Contains(hostname, ".local")
		allowNonlocal := os.Getenv("DEVELOPMENT_ALLOW_NONLOCAL_HOST") == "true"
		if !isLocalhost && !allowNonlocal {
			logging.Logger().Fatal("FATAL: DEVELOPMENT=true is set but this machine's hostname is not localhost-like (" + hostname + "). " +
				"For a normal dev workstation (e.g. WSL), set DEVELOPMENT_ALLOW_NONLOCAL_HOST=true or unset DEVELOPMENT. " +
				"Never set DEVELOPMENT=true in real production.")
		}
		logging.Logger().Warn("WARNING: DEVELOPMENT mode is enabled. Do not use in production.")
	}

	// Use PostgreSQL as the database backend
	repo := db.Repository()
	logging.Logger().Info("Using PostgreSQL as database backend")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logging.Logger().Fatal("FATAL: JWT_SECRET environment variable is required. Refusing to start with empty secret.")
	}

	// Initialize adapters
	ffEdge := functionfly.NewFunctionFlyAdapter()
	adapters := map[string]common.DeploymentAdapter{
		"workers":          cloudflare.NewCloudflareAdapter(),
		"vercel":           vercel.NewVercelAdapter(),
		"fly":              fly.NewFlyDeploymentAdapter(),
		"deno-deploy":      deno.NewDenoAdapter(),
		"functionfly-edge": ffEdge,
		"functionfly":      ffEdge, // alias for CLI/engine compatibility
		"aws-lambda":       aws.NewAWSAdapter(),
	}

	// Initialize Redis client for caching and artifact store
	redisClient, err := initializeRedisClient()
	if err != nil {
		logging.Logger().WithError(err).Warn("Failed to initialize Redis client, some features may not work")
		redisClient = nil
	}

	// Initialize Upstash Redis client (used by CSRF and other middleware when Upstash is configured)
	upstashRedis, err := initializeUpstashRedis()
	if err != nil {
		logging.Logger().WithError(err).Warn("Failed to initialize Upstash Redis client, CSRF and other features may not work")
		upstashRedis = nil
	}

	// Fallback: If Upstash is not configured but standard Redis is available,
	// wrap the standard Redis client for CSRF and other middleware that requires Upstash interface
	if upstashRedis == nil && redisClient != nil {
		upstashRedis = cache.NewUpstashRedisClientFromStandardRedis(redisClient)
		logging.Logger().Info("Using standard Redis client for CSRF and middleware (Upstash not configured)")
	}

	// Initialize artifact store (Redis for production, fallback to memory for development)
	artifactStore, err := initializeArtifactStore()
	if err != nil {
		logging.Logger().WithError(err).Warn("Failed to initialize Redis artifact store, falling back to memory store")
		artifactStore = deployment.NewMemoryArtifactStore()
	}

	router := mux.NewRouter()
	router.StrictSlash(true) // /v1/admin/functions/ redirects to /v1/admin/functions so proxy/client don't get 404

	// Default shutdown timeout (configurable via environment)
	shutdownTimeout := 30 * time.Second
	if timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT"); timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil {
			shutdownTimeout = parsed
		}
	}

	// Initialize email service
	// Default email config for testing/development
	emailConfig := email.Config{
		SMTPHost:     "localhost",
		SMTPPort:     1025, // Mailpit default port
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "noreply@functionfly.dev",
		FromName:     "FunctionFly",
		BaseURL:      "http://localhost:8080",
	}
	var emailSvc email.Service = email.NewMockService(emailConfig) // Default to mock service with config
	if svc, ok := email.NewServiceFromEnv(); ok {
		emailSvc = svc
	}

	// Validate email service configuration at startup
	if err := emailSvc.ValidateConfiguration(); err != nil {
		logging.Logger().WithError(err).Warn("Email service validation failed - magic links and other emails may not send")
	} else {
		logging.Logger().Info("Email service configuration validated successfully")
	}

	authSvc, err := auth.NewAuthService(repo, jwtSecret)
	if err != nil {
		logging.Logger().Fatal("Failed to initialize auth service: ", err)
	}
	authSvc.SetEmailService(emailSvc)
	// Notification service is set below after it is created

	// Set base URL for OAuth redirects
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		authSvc.SetBaseURL(baseURL)
	}

	// Initialize bundle provisioning infrastructure
	tenantDBConfig := storage.LoadTenantDatabaseConfig()
	dbProvisioner, err := storage.NewTenantDBProvisioner(tenantDBConfig, db)
	if err != nil {
		logging.Logger().WithError(err).Warn("TenantDBProvisioner: failed to initialize, bundle provisioning may not work")
		dbProvisioner = nil
	}
	bundleProvisioner := provisioning.NewBundleProvisioner(db.DB, repo, dbProvisioner, emailSvc)
	provisioningHandler := provisioning.NewHandler(bundleProvisioner, repo)

	// Initialize monitoring services
	monitoringSvc := monitoring.NewService(repo)
	monitoringSvc.SetDatabase(db.DB) // Provide database access for uptime calculations
	realtimeMonitor := monitoring.NewRealtimeMonitor(monitoringSvc)

	// Initialize storage service (local filesystem)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	storageService := services.NewStorageService(baseURL)
	logging.Logger().Info("Storage service initialized (backend from env: local, s3, or r2)")

	// Initialize session cleanup service
	sessionCleanup := storage.NewSessionCleanupService(repo)

	// Initialize OAuth state cleanup service
	oauthStateCleanup := storage.NewOAuthStateCleanupService(repo)

	// Initialize login attempt cleanup service
	loginAttemptCleanup := storage.NewLoginAttemptCleanupService(repo)

	// Initialize auth event cleanup service
	authEventCleanup := storage.NewAuthEventCleanupService(repo)

	// Initialize execution log cleanup service with monitoring callback
	executionRetentionConfig := storage.ExecutionRetentionConfigFromEnv()
	cleanupCallback := func(tableName string, deleted int64, err error) {
		if err != nil {
			monitoring.RecordExecutionLogCleanupError(tableName, "cleanup_failed")
		} else {
			monitoring.RecordExecutionLogCleanupDeleted(tableName, deleted)
		}
	}
	executionLogCleanup := storage.NewExecutionLogCleanupServiceWithCallback(repo, executionRetentionConfig, cleanupCallback)

	// Update retention age metrics
	monitoring.UpdateExecutionLogRetentionAge("registry_function_executions", executionRetentionConfig.ExecutionRetentionDays)
	monitoring.UpdateExecutionLogRetentionAge("registry_executions_public", executionRetentionConfig.PublicExecutionRetentionDays)
	monitoring.UpdateExecutionLogRetentionAge("execution_resource_usage", executionRetentionConfig.ResourceUsageRetentionDays)
	monitoring.UpdateExecutionLogRetentionAge("execution_meg_records", executionRetentionConfig.MEGRecordRetentionDays)
	monitoring.UpdateExecutionLogRetentionAge("drift_reports", executionRetentionConfig.DriftReportRetentionDays)
	monitoring.UpdateExecutionLogRetentionAge("execution_certificates", executionRetentionConfig.ExecutionCertRetentionDays)

	// Initialize state cleanup service for TTL-based cleanup
	stateCleanupConfig := staterepo.DefaultCleanupConfig()
	stateCleanupConfig.Interval = 1 * time.Hour
	stateCleanup := staterepo.NewCleanupService(db.GORM, stateCleanupConfig)

	// Initialize state fabric cleanup service for TTL-based cleanup
	stateFabricCleanupConfig := statefabricrepo.DefaultCleanupConfig()
	stateFabricCleanupConfig.Interval = 1 * time.Hour
	var stateFabricCleanup *statefabricrepo.CleanupService
	if statefabricrepo.IsR2StorageConfigured() {
		if r2Backend, err := statefabricrepo.NewR2StorageBackend(); err == nil {
			stateFabricCleanup = statefabricrepo.NewCleanupServiceWithR2(db.GORM, r2Backend, stateFabricCleanupConfig)
		} else {
			logging.Logger().WithError(err).Warn("Failed to initialize R2 backend for state fabric cleanup")
			stateFabricCleanup = statefabricrepo.NewCleanupService(db.GORM, stateFabricCleanupConfig)
		}
	} else {
		stateFabricCleanup = statefabricrepo.NewCleanupService(db.GORM, stateFabricCleanupConfig)
	}

	// Initialize verification service
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

	// Initialize deployment orchestrator with realtime monitor
	deploySvc := deployment.NewOrchestrator(repo, adapters, artifactStore, realtimeMonitor)

	// Initialize health monitor for backend health checks and circuit breaker
	healthMonitor := health.NewMonitor(repo)

	// Initialize notification repository and service
	notificationRepo := notification.NewPostgresRepository(db.DB)
	notificationSvc := notification.NewService(notificationRepo, db, emailSvc, logging.Logger())
	authSvc.SetNotificationService(notificationSvc)

	// Initialize recommendations service
	recommendationConfig := recommendations.DefaultRecommendationConfig()
	recommendationSvc := recommendations.NewService(db.GORM, nil, recommendationConfig)

	// Initialize usage metrics aggregation service
	usageMetricsConfig := services.LoadAggregationConfig()
	usageMetricsAgg := services.NewAggregationService(db.GORM, usageMetricsConfig)

	// Initialize state usage aggregator for billing/quota integration
	stateUsageAggregatorConfig := services.LoadStateUsageAggregatorConfig()
	stateUsageAggregator := services.NewStateUsageAggregator(db.GORM, stateUsageAggregatorConfig)

	// Initialize deferred billing checker for Backend-in-a-Box founder mode
	deferredBillingCheckInterval := 24 * time.Hour
	if v := os.Getenv("DEFERRED_BILLING_CHECK_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= time.Hour {
			deferredBillingCheckInterval = d
		}
	}
	deferredBillingChecker := billing.NewDeferredBillingChecker(repo, notificationSvc, deferredBillingCheckInterval)

	// Initialize dunning management and usage reporting repositories
	dunningRepo := storage.NewDunningRepository(db)
	usageReportingRepo := storage.NewUsageReportingRepository(db)

	// Initialize dunning manager for automated payment retry
	dunningManager := billingpkg.NewDunningManager(dunningRepo, repo, notificationSvc)

	// Initialize Trust API billing service with usage reporting
	trustBillingRepo := trustapirepo.NewBillingRepository(db.GORM)
	trustBillingService := trustapi.NewBillingService(trustBillingRepo, usageReportingRepo)

	// Create server context and WaitGroup for background workers
	serverCtx, serverCancel := context.WithCancel(context.Background())
	var serverWg sync.WaitGroup

	// Start dunning background processors
	startDunningProcessors(serverCtx, &serverWg, dunningManager)

	s := &Server{
		postgresDB:             db,
		logger:                 logging.Logger(),
		repo:                   repo,
		router:                 router,
		authSvc:                authSvc,
		routingSvc:             routing.NewRouter(repo),
		deploySvc:              deploySvc,
		monitoringSvc:          monitoringSvc,
		realtimeMonitor:        realtimeMonitor,
		storageService:         storageService,
		sessionCleanup:         sessionCleanup,
		oauthStateCleanup:      oauthStateCleanup,
		loginAttemptCleanup:    loginAttemptCleanup,
		authEventCleanup:       authEventCleanup,
		executionLogCleanup:    executionLogCleanup,
		stateCleanup:           stateCleanup,
		stateFabricCleanup:     stateFabricCleanup,
		healthMonitor:          healthMonitor,
		redisClient:            redisClient,
		upstashRedis:           upstashRedis,
		shutdownTimeout:        shutdownTimeout,
		serverCtx:              serverCtx,
		serverCancel:           serverCancel,
		wg:                     &serverWg,
		notificationSvc:        notificationSvc,
		notificationRepo:       notificationRepo,
		notificationPool:       nil,
		notificationWSHandler:  nil,
		recommendationSvc:      recommendationSvc,
		usageMetricsAgg:        usageMetricsAgg,
		stateUsageAggregator:   stateUsageAggregator,
		emailSvc:               emailSvc,
		deferredBillingChecker: deferredBillingChecker,
		dunningRepo:            dunningRepo,
		usageReportingRepo:     usageReportingRepo,
		dunningManager:         dunningManager,
		trustBillingService:    trustBillingService,
		bundleProvisioner:      bundleProvisioner,
		provisioningHandler:    provisioningHandler,
		httpServer: &http.Server{
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	s.setupRoutes(s.realtimeMonitor)

	// Serve /metrics before any wrapper so Prometheus (no Origin, from Docker) can scrape without 403
	metricsHandler := monitoring.Handler()
	// Gate localhost CORS behind development mode: in production, only explicitly configured CORS_ALLOWED_ORIGINS apply.
	var mainHandler http.Handler
	if os.Getenv("DEVELOPMENT") == "true" {
		mainHandler = localhostCORSWrapper(s.router)
	} else {
		mainHandler = s.router
	}
	handlerWithMetrics := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" && r.Method == http.MethodGet {
			metricsHandler.ServeHTTP(w, r)
			return
		}
		mainHandler.ServeHTTP(w, r)
	})
	s.httpServer.Handler = handlerWithMetrics

	return s
}

// corsResponseWriter wraps http.ResponseWriter to add CORS headers.
// WriteHeader must be idempotent: outer middleware (e.g. response tracking) or net/http may
// trigger multiple WriteHeader paths; only the first may call the underlying writer.
type corsResponseWriter struct {
	http.ResponseWriter
	origin      string
	wroteHeader bool
}

func (w *corsResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if w.origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", w.origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint, x-device-fingerprint, X-Environment")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write ensures CORS headers are applied when handlers write a body without calling WriteHeader.
func (w *corsResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// localhostCORSWrapper ensures CORS headers are set for localhost origins on every response.
// This guarantees the dashboard at :3000 can call the API at :8080 even when the route returns 404.
// SECURITY: Only active in DEVELOPMENT mode - production must use explicit CORS_ALLOWED_ORIGINS.
func localhostCORSWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		isDev := os.Getenv("DEVELOPMENT") == "true"
		isLocalhost := origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:"))

		// SECURITY: Reject localhost origins in production even via this wrapper
		if !isDev && origin != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Handle preflight first so the browser always gets CORS headers on OPTIONS.
		if r.Method == "OPTIONS" {
			if origin != "" && isLocalhost {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint, x-device-fingerprint, X-Environment")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		// Note: X-Environment header added to allow environment selection in dashboard requests

		// For WebSocket upgrades, use the original response writer to preserve Hijacker interface
		if r.Header.Get("Upgrade") == "websocket" {
			if isLocalhost && isDev {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint, x-device-fingerprint, X-Environment")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			next.ServeHTTP(w, r)
			return
		}

		// Wrap the response writer for regular HTTP requests only in dev mode with localhost
		if isLocalhost && isDev {
			w = &corsResponseWriter{ResponseWriter: w, origin: origin}
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// runFunctionLogRetention deletes function_logs older than retention (e.g. daily job, 90-day retention).
// Set retentionDays to 0 to disable (goroutine not started by caller).
func runFunctionLogRetention(ctx context.Context, db *storage.PostgresDB, interval time.Duration, retentionDays int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	retention := time.Duration(retentionDays) * 24 * time.Hour
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().UTC().Add(-retention)
			n, err := db.DeleteFunctionLogsOlderThan(ctx, cutoff)
			if err != nil {
				logging.Logger().WithError(err).Warn("Function log retention cleanup failed")
				continue
			}
			if n > 0 {
				logging.Logger().WithFields(logrus.Fields{
					"deleted_rows":   n,
					"cutoff_utc":     cutoff.Format(time.RFC3339),
					"retention_days": retentionDays,
				}).Info("Function log retention: deleted old rows")
			}
		}
	}
}

// runVaultTokenCleanup runs CleanupExpiredTokens periodically (e.g. daily) with the given olderThan (e.g. 30 days).
func runVaultTokenCleanup(ctx context.Context, repo *vaultstorage.Repository, interval, olderThan time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.CleanupExpiredTokens(ctx, olderThan); err != nil {
				logging.Logger().WithError(err).Warn("Vault token cleanup failed")
			}
		}
	}
}

func (s *Server) ListenAndServe(addr string) error {
	// Set the server address
	s.httpServer.Addr = addr

	// Channel to listen for interrupt signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Ensure server context is cancelled on function exit (shutdown)
	defer s.serverCancel()

	// Start HTTP listener first so Fly.io (and health checks) see the app listening immediately.
	// Background services are started after so a slow init does not delay the port binding.
	go func() {
		logging.Logger().WithField("addr", addr).Info("API server listening")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger().WithError(err).Fatal("Server failed to start")
		}
	}()

	// Start session cleanup routine (runs every hour)
	s.sessionCleanup.StartCleanupRoutine(s.serverCtx, time.Hour)

	// Start OAuth state cleanup routine (runs every 6 hours)
	s.oauthStateCleanup.StartCleanupRoutine(s.serverCtx, 6*time.Hour)

	// Start login attempt cleanup routine (runs daily, keeps 30 days of history)
	s.loginAttemptCleanup.StartCleanupRoutine(s.serverCtx, 24*time.Hour, 30*24*time.Hour)

	// Start auth event cleanup routine (runs daily, keeps 90 days of history for security/compliance)
	s.authEventCleanup.StartCleanupRoutine(s.serverCtx, 24*time.Hour, 90*24*time.Hour)

	// Start execution log cleanup routine (configurable retention, default daily)
	if s.executionLogCleanup != nil {
		go s.executionLogCleanup.StartCleanupRoutine(s.serverCtx)
		logging.Logger().Info("Execution log cleanup routine started")
	}

	// Start local runtime cleanup routine (runs every 5 minutes)
	go s.monitoringSvc.StartLocalRuntimeCleanup(s.serverCtx, 5*time.Minute, 10*time.Minute)

	// Start health monitor for backend health checks and circuit breaker (MVP gap fix)
	s.healthMonitor.Start()

	// Start uptime metrics updater
	s.monitoringSvc.StartUptimeUpdater()

	// Start notification service
	s.notificationSvc.Start(s.serverCtx)
	logging.Logger().Info("Notification service started")

	// Start notification WebSocket hub and PostgreSQL LISTEN subscription
	if s.notificationWSHandler != nil {
		s.notificationWSHandler.RunHub()
		if s.notificationPool != nil {
			s.notificationWSHandler.RunNotificationSubscription(s.serverCtx, s.notificationPool)
		}
		logging.Logger().Info("Notification WebSocket hub and LISTEN subscription started")
	} else {
		logging.Logger().Warn("Notification WebSocket handler not initialized – skipping hub start")
	}

	// Start deferred billing checker for Backend-in-a-Box founder mode
	if s.deferredBillingChecker != nil {
		s.deferredBillingChecker.Start()
		logging.Logger().Info("Deferred billing checker started")
	}

	// Start trigger engine for state changes
	if s.triggerEngine != nil {
		s.triggerEngine.Start(s.serverCtx)
		logging.Logger().Info("Trigger engine started")
	}

	// Start state cleanup routine for TTL-based cleanup (runs every hour)
	if s.stateCleanup != nil {
		go s.stateCleanup.StartCleanupRoutine(s.serverCtx)
		logging.Logger().Info("State cleanup routine started")
	}

	// Start state fabric TTL cleanup routine for expired snapshots (runs every hour)
	if s.stateFabricCleanup != nil {
		go s.stateFabricCleanup.StartCleanupRoutine(s.serverCtx)
		logging.Logger().Info("State fabric TTL cleanup routine started")
	}

	// Start vault expired-token cleanup (runs daily; prunes tokens expired/revoked > 30 days ago)
	if s.vaultRepo != nil {
		go runVaultTokenCleanup(s.serverCtx, s.vaultRepo, 24*time.Hour, 30*24*time.Hour)
		logging.Logger().Info("Vault token cleanup routine started")

		// Phase 1.3: secret expiration sweeper (every 15m by default; honors
		// VAULT_EXPIRATION_SWEEP_INTERVAL if set). It also expires stale
		// break-glass grants as a side-effect.
		sweeperInterval := 15 * time.Minute
		if v := os.Getenv("VAULT_EXPIRATION_SWEEP_INTERVAL"); v != "" {
			if d, err := time.ParseDuration(v); err == nil && d >= time.Minute {
				sweeperInterval = d
			}
		}
		sweeper := vaultstorage.NewExpirationSweeper(s.vaultRepo, vaultstorage.ExpirationSweeperConfig{
			Interval:      sweeperInterval,
			WarningWindow: 7 * 24 * time.Hour,
			BatchSize:     200,
			Logger:        logging.Logger(),
		})
		go sweeper.Start(s.serverCtx)
		logging.Logger().WithField("interval", sweeperInterval.String()).Info("Vault expiration sweeper started")

		// Phase 2.2: dynamic-lease sweeper — drops DB users whose lease
		// has expired. Interval is configurable via env var.
		leaseInterval := 5 * time.Minute
		if v := os.Getenv("VAULT_DYNAMIC_LEASE_SWEEP_INTERVAL"); v != "" {
			if d, err := time.ParseDuration(v); err == nil && d >= time.Minute {
				leaseInterval = d
			}
		}
		leaseSweeper := vaultstorage.NewDynamicLeaseSweeper(s.vaultRepo, vaultstorage.DynamicLeaseSweeperConfig{
			Interval:  leaseInterval,
			BatchSize: 100,
			Logger:    logging.Logger(),
		})
		go leaseSweeper.Start(s.serverCtx)
		logging.Logger().WithField("interval", leaseInterval.String()).Info("Vault dynamic-lease sweeper started")

		// Phase 5.3: leader election. When running multiple API
		// instances, only the leader runs the background sweepers
		// above. Single-instance deployments are always leader.
		if s.redisClient != nil {
			leaderNS := os.Getenv("VAULT_LEADER_NAMESPACE")
			if leaderNS == "" {
				leaderNS = "vault-sweep"
			}
			elector := vaultstorage.NewLeaderElector(s.redisClient, vaultstorage.LeaderElectorConfig{
				Namespace:       leaderNS,
				TTL:             30 * time.Second,
				RenewInterval:   10 * time.Second,
				AcquireInterval: 5 * time.Second,
				Logger:          logging.Logger(),
			})
			// Wrap the sweepers in leader-gated funcs so only the
			// elected instance runs them.
			go vaultstorage.RunLeaderGatedWorkers(s.serverCtx, elector, logging.Logger(),
				func(ctx context.Context) { sweeper.Start(ctx) },
				func(ctx context.Context) { leaseSweeper.Start(ctx) },
			)
			logging.Logger().WithField("namespace", leaderNS).Info("Vault leader election started")
		}
	}

	// Function log retention (default 90 days; FUNCTION_LOG_RETENTION_DAYS=0 disables)
	retentionDays := 90
	if v := strings.TrimSpace(os.Getenv("FUNCTION_LOG_RETENTION_DAYS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			retentionDays = n
		}
	}
	cleanupInterval := 24 * time.Hour
	if v := strings.TrimSpace(os.Getenv("FUNCTION_LOG_CLEANUP_INTERVAL")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= time.Hour {
			cleanupInterval = d
		}
	}
	if retentionDays > 0 {
		go runFunctionLogRetention(s.serverCtx, s.postgresDB, cleanupInterval, retentionDays)
		logging.Logger().WithFields(logrus.Fields{
			"retention_days": retentionDays,
			"interval":       cleanupInterval.String(),
		}).Info("Function log retention cleanup started")
	} else {
		logging.Logger().Info("Function log retention cleanup disabled (FUNCTION_LOG_RETENTION_DAYS=0)")
	}

	// Start usage metrics aggregation service
	if s.usageMetricsAgg != nil && s.usageMetricsAgg.IsEnabled() {
		s.usageMetricsAgg.StartAggregationRoutine(s.serverCtx)
		logging.Logger().Info("Usage metrics aggregation service started")
	}

	// Start state usage aggregator for billing/quota integration
	if s.stateUsageAggregator != nil && s.stateUsageAggregator.IsEnabled() {
		s.stateUsageAggregator.Start(s.serverCtx)
		logging.Logger().Info("State usage aggregator started")
	}

	// Start unified analytics sync job (Phase 3: rollups from source tables)
	if s.unifiedSyncJob != nil {
		s.unifiedSyncJob.Start(s.serverCtx)
		logging.Logger().Info("Unified analytics sync job started")
	}

	// Wait for interrupt signal
	<-done
	logging.Logger().Info("Server is shutting down...")

	// On second signal, force exit immediately (in case graceful shutdown hangs)
	go func() {
		<-done
		logging.Logger().Warn("Second interrupt received, forcing immediate shutdown")
		os.Exit(1)
	}()

	// Stop health monitor first
	s.healthMonitor.Stop()

	// Stop trigger engine
	if s.triggerEngine != nil {
		s.triggerEngine.Stop()
		logging.Logger().Info("Trigger engine stopped")
	}

	// Create shutdown context with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	// Gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		logging.Logger().WithError(err).Error("Server forced to shutdown")
		return err
	}

	logging.Logger().Info("Server exited")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil // Server not started
	}

	// Cancel server context to signal all background workers to stop
	if s.serverCancel != nil {
		s.serverCancel()
		logging.Logger().Info("Server context cancelled, background workers will stop")
	}

	// Wait for all background worker goroutines to complete
	if s.wg != nil {
		s.wg.Wait()
		logging.Logger().Info("All background workers have stopped")
	}

	// Stop health monitor
	s.healthMonitor.Stop()

	// Stop notification service
	if s.notificationSvc != nil {
		s.notificationSvc.Stop()
		logging.Logger().Info("Notification service stopped")
	}

	// Stop WebSocket hub (must be before pool close to unblock LISTEN subscription)
	if s.notificationWSHandler != nil {
		s.notificationWSHandler.StopHub()
		logging.Logger().Info("WebSocket hub stopped")
	}

	// Close the notification LISTEN pool
	if s.notificationPool != nil {
		s.notificationPool.Close()
		logging.Logger().Info("Notification LISTEN pool closed")
	}

	// Stop usage metrics aggregation service
	if s.usageMetricsAgg != nil {
		s.usageMetricsAgg.Stop()
		logging.Logger().Info("Usage metrics aggregation service stopped")
	}

	// Stop state usage aggregator
	if s.stateUsageAggregator != nil {
		s.stateUsageAggregator.Stop()
		logging.Logger().Info("State usage aggregator stopped")
	}

	// Stop unified analytics sync job
	if s.unifiedSyncJob != nil {
		s.unifiedSyncJob.Stop()
		logging.Logger().Info("Unified analytics sync job stopped")
	}

	// Stop cleanup services
	if s.sessionCleanup != nil {
		s.sessionCleanup.Stop()
		logging.Logger().Info("Session cleanup service stopped")
	}
	if s.oauthStateCleanup != nil {
		s.oauthStateCleanup.Stop()
		logging.Logger().Info("OAuth state cleanup service stopped")
	}
	if s.loginAttemptCleanup != nil {
		s.loginAttemptCleanup.Stop()
		logging.Logger().Info("Login attempt cleanup service stopped")
	}
	if s.authEventCleanup != nil {
		s.authEventCleanup.Stop()
		logging.Logger().Info("Auth event cleanup service stopped")
	}
	if s.dunningManager != nil {
		s.dunningManager.Stop()
		logging.Logger().Info("Dunning manager stopped")
	}
	if s.billingSyncJob != nil {
		s.billingSyncJob.Stop()
		logging.Logger().Info("Billing sync job stopped")
	}
	if s.exportScheduler != nil {
		s.exportScheduler.Stop()
		logging.Logger().Info("Export scheduler stopped")
	}
	if s.consciousnessCleanupScheduler != nil {
		s.consciousnessCleanupScheduler.Stop()
		logging.Logger().Info("Consciousness cleanup scheduler stopped")
	}
	if s.consciousnessScheduler != nil {
		s.consciousnessScheduler.Stop()
		logging.Logger().Info("Consciousness scheduler stopped")
	}
	if s.consciousnessRetryScheduler != nil {
		s.consciousnessRetryScheduler.Stop()
		logging.Logger().Info("Consciousness retry scheduler stopped")
	}

	// Shutdown the HTTP server gracefully

	// Shutdown the persistent SandboxClient daemon
	regexec.ShutdownSandboxClient()

	return s.httpServer.Shutdown(ctx)
}

// GetShutdownTimeout returns the configured shutdown timeout
func (s *Server) GetShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}

// GetNotificationService returns the notification service
func (s *Server) GetNotificationService() *notification.Service {
	return s.notificationSvc
}

// GetNotificationRepository returns the notification repository
func (s *Server) GetNotificationRepository() notification.Repository {
	return s.notificationRepo
}

// Context returns the server lifecycle context for background workers.
// This context is cancelled when the server shuts down.
func (s *Server) Context() context.Context {
	if s.serverCtx == nil {
		return context.Background()
	}
	return s.serverCtx
}

// redisTLSConfig returns a *tls.Config if the address is an Upstash host (requires TLS).
// Returns nil for plain-text local Redis.
func redisTLSConfig(addr string) *tls.Config {
	if strings.Contains(addr, "upstash.io") {
		return &tls.Config{}
	}
	return nil
}

// initializeArtifactStore initializes the artifact store based on environment configuration
func initializeArtifactStore() (deployment.ArtifactStore, error) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil && parsed >= 0 {
			redisDB = parsed
		}
	}

	artifactTTL := 7 * 24 * time.Hour
	if ttlStr := os.Getenv("ARTIFACT_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil {
			artifactTTL = parsed
		}
	}

	// Create Redis client (TLS auto-detected for Upstash hosts)
	rdb := redis.NewClient(&redis.Options{
		Addr:      redisAddr,
		Password:  redisPassword,
		DB:        redisDB,
		TLSConfig: redisTLSConfig(redisAddr),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logging.Logger().WithFields(logrus.Fields{
		"addr": redisAddr,
		"db":   redisDB,
		"ttl":  artifactTTL.String(),
	}).Info("Initialized Redis artifact store")

	return deployment.NewRedisArtifactStoreFromClient(rdb, artifactTTL), nil
}

// initializeRedisClient creates a Redis client for caching.
// If UPSTASH_REDIS_REST_URL/TOKEN are set, returns nil (callers use Upstash HTTP client instead).
func initializeRedisClient() (*redis.Client, error) {
	upstashURL := os.Getenv("UPSTASH_REDIS_REST_URL")
	upstashToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if upstashURL != "" && upstashToken != "" {
		logging.Logger().Info("Using Upstash Redis REST API for caching (go-redis client not needed)")
		return nil, nil
	}

	// Direct Redis client (with TLS for Upstash)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		if parsed, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = parsed
		}
	}

	// Create Redis client (TLS auto-detected for Upstash hosts)
	rdb := redis.NewClient(&redis.Options{
		Addr:      redisAddr,
		Password:  redisPassword,
		DB:        redisDB,
		TLSConfig: redisTLSConfig(redisAddr),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logging.Logger().WithFields(logrus.Fields{
		"addr": redisAddr,
		"db":   redisDB,
	}).Info("Initialized Redis client for caching")

	return rdb, nil
}

// initializeUpstashRedis creates an Upstash Redis client
func initializeUpstashRedis() (*cache.UpstashRedisClient, error) {
	upstashURL := os.Getenv("UPSTASH_REDIS_REST_URL")
	upstashToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if upstashURL == "" || upstashToken == "" {
		return nil, nil // Not configured for Upstash
	}

	config := &cache.UpstashConfig{
		URL:       upstashURL,
		Token:     upstashToken,
		IsUpstash: true,
	}

	client := cache.NewUpstashRedisClient(config)
	if client == nil {
		return nil, fmt.Errorf("failed to create Upstash Redis client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Upstash Redis: %w", err)
	}

	logging.Logger().Info("Initialized Upstash Redis client")
	return client, nil
}

// serveSPAIndex serves the React SPA index.html for playground routes
func (s *Server) serveSPAIndex(w http.ResponseWriter, r *http.Request) {
	// Serve the SPA index.html from web/dashboard/dist/index.html
	indexPath := "web/dashboard/dist/index.html"

	// Check if file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		logging.Logger().WithError(err).Error("SPA index.html not found")
		http.Error(w, "SPA not available", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, indexPath)
}

// startDunningProcessors starts background goroutines for dunning management
// These processors handle scheduled payment retries and grace period expirations
func startDunningProcessors(ctx context.Context, wg *sync.WaitGroup, dunningManager *billingpkg.DunningManager) {
	// Retry processor - runs every hour to process scheduled retries
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logging.Logger().Info("Dunning retry processor stopping due to context cancellation")
				return
			case <-dunningManager.StopChan():
				logging.Logger().Info("Dunning retry processor stopping")
				return
			case <-ticker.C:
				if err := dunningManager.ProcessScheduledRetries(ctx); err != nil {
					logging.Logger().WithError(err).Error("Failed to process scheduled payment retries")
				}
			}
		}
	}()

	// Grace period expiration processor - runs daily to check for expired grace periods
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		if err := dunningManager.ProcessGracePeriodExpirations(ctx); err != nil {
			logging.Logger().WithError(err).Error("Failed to process grace period expirations")
		}

		for {
			select {
			case <-ctx.Done():
				logging.Logger().Info("Dunning grace period processor stopping due to context cancellation")
				return
			case <-dunningManager.StopChan():
				logging.Logger().Info("Dunning grace period processor stopping")
				return
			case <-ticker.C:
				if err := dunningManager.ProcessGracePeriodExpirations(ctx); err != nil {
					logging.Logger().WithError(err).Error("Failed to process grace period expirations")
				}
			}
		}
	}()

	logging.Logger().Info("Dunning processors started (retry: 1h, grace expiration: 24h)")
}
