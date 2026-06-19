package api

import (
	"context"

	"github.com/functionfly/functionfly/internal/agent/graph"
	frghandler "github.com/functionfly/functionfly/internal/api/handlers/frg"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/dre"
	"github.com/functionfly/functionfly/internal/frg"
	frgapigen "github.com/functionfly/functionfly/internal/frg/api"
	"github.com/functionfly/functionfly/internal/frg/trigger"
	"github.com/functionfly/functionfly/internal/logging"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/storage/state"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerFRGRoutes wires the Function Registry + Live Runtime Graph (FRG) endpoints.
// FRG provides versioned, composable function graphs with streaming execution.
func registerFRGRoutes(
	s *Server,
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	_ *middleware.ExecutionCoordinatorMiddleware,
	registryRepo *registry.RegistryRepository,
	advancedSecurityMiddleware *advanced_security.AdvancedSecurityMiddleware,
	realtimeUsageTracker services.RealtimeUsageTrackerInterface,
) {
	// Initialize FRG components with Redis cache
	var frgCache *cache.RegistryRedisCache
	if s.redisClient != nil {
		frgCache = cache.NewRegistryRedisCache(s.redisClient, 0)
	}
	frgRepo := frg.NewRepository(s.postgresDB.GORM, frgCache)

	// Run FRG migrations before accessing the database
	if err := frgRepo.AutoMigrate(context.Background()); err != nil {
		logging.Logger().WithError(err).Error("FRG: failed to run migrations; graph tables may be missing")
		// Continue anyway - tables may already exist
	}

	// Get graph service from existing agent package
	graphService := graph.NewService(s.postgresDB.GORM)

	// Load DRE configuration from environment (optional - FXCERTs will be unsigned if not configured)
	dreNodeKey, dreNodeID, dreRegion, err := dre.LoadNodeKeyFromEnv()
	if err != nil {
		logging.Logger().WithError(err).Warn("FRG DRE: failed to load node key from env; FXCERTs will be unsigned")
	}
	drePlatformKey, _ := dre.LoadPlatformKeyFromEnv()

	// Initialize event bus — try NATS first, fall back to in-memory
	var eventBus frg.EventStream
	natsStream, err := frg.TryCreateNATSEventStream()
	if err != nil {
		logging.Logger().WithError(err).Warn("NATS unavailable, using in-memory event stream")
	}
	if natsStream != nil {
		eventBus = natsStream
		logging.Logger().Info("NATS event stream active")
	} else {
		eventBus = frg.NewInMemoryEventStream()
		logging.Logger().Info("In-memory event stream initialized")
	}

	// Start NATS runtime subscriber (listens for runtime registrations,
	// heartbeats, and execution results from Prism/SAR/Kotlin runtimes)
	_ = frg.TryCreateRuntimeSubscriber(frg.RuntimeEventHandlers{
		OnRegistration: func(reg frg.RuntimeRegistration) {
			logging.Logger().WithFields(logrus.Fields{
				"cell_id": reg.CellID,
				"name":    reg.Name,
			}).Info("Runtime registered via NATS")
		},
		OnHeartbeat: func(hb frg.RuntimeHeartbeat) {
			logging.Logger().WithFields(logrus.Fields{
				"cell_id":           hb.CellID,
				"status":            hb.Status,
				"active_executions": hb.ActiveExecutions,
			}).Debug("Runtime heartbeat via NATS")
		},
		OnExecutionResult: func(result frg.RuntimeExecutionResult) {
			logging.Logger().WithFields(logrus.Fields{
				"execution_id": result.ExecutionID,
				"cell_id":      result.CellID,
				"status":       result.Status,
			}).Info("Execution result received via NATS")
		},
	})

	// Initialize execution engine
	engine, err := frg.NewExecutionEngine(
		frgRepo,
		registryRepo,
		storage.NewFunctionRepository(s.postgresDB.DB),
		state.NewStateRepository(s.postgresDB.GORM),
		graphService,
		eventBus,
		dreNodeID,
		dreRegion,
		dreNodeKey,
		drePlatformKey,
	)
	if err != nil {
		logging.Logger().WithError(err).Error("Failed to initialize FRG execution engine")
		return
	}

	// Initialize trigger router for reactive graphs
	triggerRouter := trigger.NewRouter(frgRepo, engine)
	triggerRouter.Start()

	// Load existing triggers on startup
	if err := triggerRouter.StartTriggerLoader(context.Background()); err != nil {
		logging.Logger().WithError(err).Warn("Failed to load graph triggers on startup")
	}

	// Initialize handler
	frgHandler := frghandler.NewHandler(
		frgRepo,
		registryRepo,
		storage.NewFunctionRepository(s.postgresDB.DB),
		engine,
		graphService,
		nil, // Cache service
		frghandler.NewAICompositionClient(),
		frghandler.NewEmbeddingServiceClient(),
		realtimeUsageTracker,
	)

	// ── Graph Definitions (CRUD) ───────────────────────────────────────────
	// Public read
	api.HandleFunc("/frg/graphs", frgHandler.ListGraphs).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", frgHandler.GetGraph).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/instances", frgHandler.ListInstances).Methods("GET", "OPTIONS")

	// AI Discovery (public search)
	api.HandleFunc("/frg/discover", frgHandler.SemanticSearch).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/optimizations", frgHandler.GetOptimizations).Methods("GET", "OPTIONS")

	// Protected write (with rate limiting)
	api.HandleFunc("/frg/graphs", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.CreateGraph))).Methods("POST", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.UpdateGraph))).Methods("PUT", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.DeleteGraph))).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/publish", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.PublishGraphVersion))).Methods("POST", "OPTIONS")

	// Remix/Fork (with rate limiting)
	api.HandleFunc("/frg/graphs/{author}/{name}/remix", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.RemixGraph))).Methods("POST", "OPTIONS")

	// ── Graph Execution ─────────────────────────────────────────────────────
	// Execute (auth required + rate limiting - tracks usage and prevents abuse)
	api.HandleFunc("/gx/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ExecuteGraph))).Methods("POST", "OPTIONS")
	api.HandleFunc("/gx/{author}/{name}@{version}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ExecuteGraph))).Methods("POST", "OPTIONS")

	// Instance management
	api.HandleFunc("/frg/instances/{instance_id}", frgHandler.GetInstanceStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/instances/{instance_id}/stop", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.StopInstance))).Methods("POST", "OPTIONS")
	api.HandleFunc("/frg/instances/{instance_id}/resume", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ResumeInstance))).Methods("POST", "OPTIONS")

	// ── AI Composition (protected + rate limiting) ──────────────────────────────────────────
	api.HandleFunc("/frg/compose", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.AICompose))).Methods("POST", "OPTIONS")
	api.HandleFunc("/frg/functions/generate", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.GenerateFunction))).Methods("POST", "OPTIONS")

	// ── Webhook Triggers (public endpoints that trigger graphs) ───────────────
	// Dynamic webhooks - graphs can register custom paths
	api.HandleFunc("/webhook/{path:.*}", triggerRouter.GetWebhookHandler()).Methods("POST", "GET", "OPTIONS")
	// Fixed path webhooks for specific graphs
	api.HandleFunc("/api/webhooks/graph/{graph_id}", triggerRouter.GetWebhookHandler()).Methods("POST", "OPTIONS")

	// ── Auto-Generated APIs (Backend as a Graph) ────────────────────────────────
	// REST/GraphQL APIs auto-generated from published graphs
	autoGenAPIHandler := frgapigen.NewAutoGeneratedAPIHandler(
		frgRepo,
		state.NewStateRepository(s.postgresDB.GORM),
		engine,
		"/api/graphs",
	)

	// Register dynamic routes for all published graphs
	if err := autoGenAPIHandler.RouteRegistrar(api); err != nil {
		logging.Logger().WithError(err).Warn("Failed to register some auto-generated API routes")
	}
}

// registerFRGRoutesOnRoot registers FRG routes directly on the root router (for /frg/* paths)
func registerFRGRoutesOnRoot(
	s *Server,
	root *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	_ *middleware.ExecutionCoordinatorMiddleware,
	registryRepo *registry.RegistryRepository,
	advancedSecurityMiddleware *advanced_security.AdvancedSecurityMiddleware,
	realtimeUsageTracker services.RealtimeUsageTrackerInterface,
) {
	// Initialize FRG components with Redis cache
	var frgCache *cache.RegistryRedisCache
	if s.redisClient != nil {
		frgCache = cache.NewRegistryRedisCache(s.redisClient, 0)
	}
	frgRepo := frg.NewRepository(s.postgresDB.GORM, frgCache)

	// Run FRG migrations before accessing the database
	if err := frgRepo.AutoMigrate(context.Background()); err != nil {
		logging.Logger().WithError(err).Error("FRG: failed to run migrations; graph tables may be missing")
	}

	// Get graph service from existing agent package
	graphService := graph.NewService(s.postgresDB.GORM)

	// Load DRE configuration
	dreNodeKey, dreNodeID, dreRegion, err := dre.LoadNodeKeyFromEnv()
	if err != nil {
		logging.Logger().WithError(err).Warn("FRG DRE: failed to load node key from env; FXCERTs will be unsigned")
	}
	drePlatformKey, _ := dre.LoadPlatformKeyFromEnv()

	// Initialize event bus — try NATS first, fall back to in-memory
	var eventBus frg.EventStream
	natsStream2, err := frg.TryCreateNATSEventStream()
	if err != nil {
		logging.Logger().WithError(err).Warn("NATS unavailable (root routes), using in-memory event stream")
	}
	if natsStream2 != nil {
		eventBus = natsStream2
	} else {
		eventBus = frg.NewInMemoryEventStream()
	}

	// Initialize execution engine
	engine, err := frg.NewExecutionEngine(
		frgRepo,
		registryRepo,
		storage.NewFunctionRepository(s.postgresDB.DB),
		state.NewStateRepository(s.postgresDB.GORM),
		graphService,
		eventBus,
		dreNodeID,
		dreRegion,
		dreNodeKey,
		drePlatformKey,
	)
	if err != nil {
		logging.Logger().WithError(err).Error("Failed to initialize FRG execution engine")
		return
	}

	// Initialize trigger router
	triggerRouter := trigger.NewRouter(frgRepo, engine)
	triggerRouter.Start()
	_ = triggerRouter.StartTriggerLoader(context.Background())

	// Initialize handler
	frgHandler := frghandler.NewHandler(
		frgRepo,
		registryRepo,
		storage.NewFunctionRepository(s.postgresDB.DB),
		engine,
		graphService,
		nil,
		frghandler.NewAICompositionClient(),
		frghandler.NewEmbeddingServiceClient(),
		realtimeUsageTracker,
	)

	// ── FRG Graph Definition Routes (on root, no /v1 prefix) ──────────────────
	root.HandleFunc("/frg/graphs", frgHandler.ListGraphs).Methods("GET", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}", frgHandler.GetGraph).Methods("GET", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}/instances", frgHandler.ListInstances).Methods("GET", "OPTIONS")
	root.HandleFunc("/frg/discover", frgHandler.SemanticSearch).Methods("GET", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}/optimizations", frgHandler.GetOptimizations).Methods("GET", "OPTIONS")

	// Protected write operations (with rate limiting)
	root.HandleFunc("/frg/graphs", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.CreateGraph))).Methods("POST", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.UpdateGraph))).Methods("PUT", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.DeleteGraph))).Methods("DELETE", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}/publish", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.PublishGraphVersion))).Methods("POST", "OPTIONS")
	root.HandleFunc("/frg/graphs/{author}/{name}/remix", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.RemixGraph))).Methods("POST", "OPTIONS")

	// Graph execution (auth required + rate limiting)
	root.HandleFunc("/gx/{author}/{name}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ExecuteGraph))).Methods("POST", "OPTIONS")
	root.HandleFunc("/gx/{author}/{name}@{version}", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ExecuteGraph))).Methods("POST", "OPTIONS")

	// Instance management
	root.HandleFunc("/frg/instances/{instance_id}", frgHandler.GetInstanceStatus).Methods("GET", "OPTIONS")
	root.HandleFunc("/frg/instances/{instance_id}/stop", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.StopInstance))).Methods("POST", "OPTIONS")

	// AI Composition (with rate limiting)
	root.HandleFunc("/frg/compose", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.AICompose))).Methods("POST", "OPTIONS")
	root.HandleFunc("/frg/functions/generate", advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.GenerateFunction))).Methods("POST", "OPTIONS")

	// Webhooks
	root.HandleFunc("/webhook/{path:.*}", triggerRouter.GetWebhookHandler()).Methods("POST", "GET", "OPTIONS")
	root.HandleFunc("/api/webhooks/graph/{graph_id}", triggerRouter.GetWebhookHandler()).Methods("POST", "OPTIONS")

	// Auto-generated APIs
	autoGenAPIHandler := frgapigen.NewAutoGeneratedAPIHandler(
		frgRepo,
		state.NewStateRepository(s.postgresDB.GORM),
		engine,
		"/api/graphs",
	)
	if err := autoGenAPIHandler.RouteRegistrar(root); err != nil {
		logging.Logger().WithError(err).Warn("Failed to register some auto-generated API routes")
	}
}
