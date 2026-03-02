package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/deno"
	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/adapters/vercel"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/deployment"
	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/health"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/routing"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Server struct {
	// Database
	postgresDB *storage.PostgresDB

	// Repository (unified interface)
	repo storage.Repository

	router          *mux.Router
	authSvc         *auth.AuthService
	routingSvc      *routing.Router
	deploySvc       *deployment.Orchestrator
	monitoringSvc   *monitoring.Service
	realtimeMonitor *monitoring.RealtimeMonitor
	storageService  *services.StorageService
	sessionCleanup  *storage.SessionCleanupService
	healthMonitor   *health.Monitor
	redisClient     *redis.Client
	httpServer      *http.Server
	shutdownTimeout time.Duration

	// Notification service
	notificationSvc *notification.Service
	notificationRepo notification.Repository
}

func NewServer(db *storage.PostgresDB) *Server {
	// Use PostgreSQL as the database backend
	repo := db.Repository()
	logrus.Info("Using PostgreSQL as database backend")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "functionfly-jwt-secret-key-2026" // must match generate_token.go default for local publish
	}

	// Initialize adapters
	adapters := map[string]common.DeploymentAdapter{
		"workers":     cloudflare.NewCloudflareAdapter(),
		"vercel":      vercel.NewVercelAdapter(),
		"fly":         fly.NewFlyDeploymentAdapter(),
		"deno-deploy": deno.NewDenoAdapter(),
	}

	// Initialize Redis client for caching and artifact store
	redisClient, err := initializeRedisClient()
	if err != nil {
		logrus.WithError(err).Warn("Failed to initialize Redis client, some features may not work")
		redisClient = nil
	}

	// Initialize artifact store (Redis for production, fallback to memory for development)
	artifactStore, err := initializeArtifactStore()
	if err != nil {
		logrus.WithError(err).Warn("Failed to initialize Redis artifact store, falling back to memory store")
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
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		emailConfig := email.Config{
			SMTPHost:     smtpHost,
			SMTPPort:     587, // Default SMTP port
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    os.Getenv("FROM_EMAIL"),
			FromName:     os.Getenv("FROM_NAME"),
			BaseURL:      os.Getenv("BASE_URL"),
		}
		if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
			if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
				emailConfig.SMTPPort = port
			}
		}
		emailSvc = email.NewSMTPService(emailConfig)
	}

	authSvc := auth.NewAuthService(repo, jwtSecret)
	authSvc.SetEmailService(emailSvc)

	// Set base URL for OAuth redirects
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		authSvc.SetBaseURL(baseURL)
	}

	// Initialize monitoring services
	monitoringSvc := monitoring.NewService(repo)
	realtimeMonitor := monitoring.NewRealtimeMonitor(monitoringSvc)

	// Initialize storage service (local filesystem)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	storageService := services.NewStorageService(baseURL)
	logrus.Info("Storage service initialized with local filesystem")

	// Initialize session cleanup service
	sessionCleanup := storage.NewSessionCleanupService(repo)

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
	notificationSvc := notification.NewService(notificationRepo, db, emailSvc, logrus.New())

	s := &Server{
		postgresDB:       db,
		repo:             repo,
		router:           router,
		authSvc:          authSvc,
		routingSvc:       routing.NewRouter(repo),
		deploySvc:        deploySvc,
		monitoringSvc:    monitoringSvc,
		realtimeMonitor:  realtimeMonitor,
		storageService:   storageService,
		sessionCleanup:   sessionCleanup,
		healthMonitor:    healthMonitor,
		redisClient:      redisClient,
		shutdownTimeout:  shutdownTimeout,
		notificationSvc:  notificationSvc,
		notificationRepo: notificationRepo,
		httpServer: &http.Server{
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	s.setupRoutes(s.realtimeMonitor)

	// Wrap router so every response (including 404) gets CORS when origin is localhost
	s.httpServer.Handler = localhostCORSWrapper(s.router)

	return s
}

// localhostCORSWrapper ensures CORS headers are set for localhost origins on every response.
// This guarantees the dashboard at :3000 can call the API at :8080 even when the route returns 404.
func localhostCORSWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		isLocalhost := origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:"))

		// Handle preflight first so the browser always gets CORS headers on OPTIONS.
		if r.Method == "OPTIONS" {
			if isLocalhost {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if isLocalhost {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe(addr string) error {
	// Set the server address
	s.httpServer.Addr = addr

	// Start session cleanup routine (runs every hour)
	s.sessionCleanup.StartCleanupRoutine(time.Hour)

	// Start local runtime cleanup routine (runs every 5 minutes)
	ctx := context.Background()
	go s.monitoringSvc.StartLocalRuntimeCleanup(ctx, 5*time.Minute, 10*time.Minute)

	// Start health monitor for backend health checks and circuit breaker (MVP gap fix)
	s.healthMonitor.Start()

	// Start notification service
	s.notificationSvc.Start(ctx)
	logrus.Info("Notification service started")

	// Channel to listen for interrupt signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logrus.WithField("addr", addr).Info("API server listening")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	<-done
	logrus.Info("Server is shutting down...")

	// Stop health monitor first
	s.healthMonitor.Stop()

	// Create shutdown context with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	// Gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
		return err
	}

	logrus.Info("Server exited")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil // Server not started
	}

	// Stop health monitor
	s.healthMonitor.Stop()

	// Stop notification service
	if s.notificationSvc != nil {
		s.notificationSvc.Stop()
		logrus.Info("Notification service stopped")
	}

	// Shutdown the HTTP server gracefully
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

// initializeArtifactStore initializes the artifact store based on environment configuration
func initializeArtifactStore() (deployment.ArtifactStore, error) {
	// Check if Redis is configured
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0 // Default database
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil && parsed >= 0 {
			redisDB = parsed
		}
	}

	// Configure artifact TTL (default: 7 days)
	artifactTTL := 7 * 24 * time.Hour
	if ttlStr := os.Getenv("ARTIFACT_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil {
			artifactTTL = parsed
		}
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"addr": redisAddr,
		"db":   redisDB,
		"ttl":  artifactTTL.String(),
	}).Info("Initialized Redis artifact store")

	return deployment.NewRedisArtifactStoreFromClient(rdb, artifactTTL), nil
}

// initializeRedisClient creates a Redis client for caching
func initializeRedisClient() (*redis.Client, error) {
	// Get Redis configuration from environment variables
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		if parsed, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = parsed
		}
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"addr": redisAddr,
		"db":   redisDB,
	}).Info("Initialized Redis client for caching")

	return rdb, nil
}

// serveSPAIndex serves the React SPA index.html for playground routes
func (s *Server) serveSPAIndex(w http.ResponseWriter, r *http.Request) {
	// Serve the SPA index.html from web/dashboard/dist/index.html
	indexPath := "web/dashboard/dist/index.html"

	// Check if file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		logrus.WithError(err).Error("SPA index.html not found")
		http.Error(w, "SPA not available", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, indexPath)
}
