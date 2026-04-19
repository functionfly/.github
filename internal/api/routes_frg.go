package api

import (
	"context"

	"github.com/functionfly/functionfly/internal/agent/graph"
	frghandler "github.com/functionfly/functionfly/internal/api/handlers/frg"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/dre"
	"github.com/functionfly/functionfly/internal/frg"
	frgapigen "github.com/functionfly/functionfly/internal/frg/api"
	"github.com/functionfly/functionfly/internal/frg/trigger"
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
) {
	// Initialize FRG components with Redis cache
	var frgCache *cache.RegistryRedisCache
	if s.redisClient != nil {
		frgCache = cache.NewRegistryRedisCache(s.redisClient, 0)
	}
	frgRepo := frg.NewRepository(s.postgresDB.GORM, frgCache)

	// Run FRG migrations before accessing the database
	if err := frgRepo.AutoMigrate(context.Background()); err != nil {
		logrus.WithError(err).Error("FRG: failed to run migrations; graph tables may be missing")
		// Continue anyway - tables may already exist
	}

	// Get graph service from existing agent package
	graphService := graph.NewService(s.postgresDB.GORM)

	// Load DRE configuration from environment (optional - FXCERTs will be unsigned if not configured)
	dreNodeKey, dreNodeID, dreRegion, err := dre.LoadNodeKeyFromEnv()
	if err != nil {
		logrus.WithError(err).Warn("FRG DRE: failed to load node key from env; FXCERTs will be unsigned")
	}
	drePlatformKey, _ := dre.LoadPlatformKeyFromEnv()

	// Initialize event bus (in-memory for Fly.io/Upstack stack)
	// For NATS support, build with: go build -tags nats
	eventBus := frg.NewInMemoryEventStream()
	logrus.Info("In-memory event stream initialized")

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
		logrus.WithError(err).Error("Failed to initialize FRG execution engine")
		return
	}

	// Initialize trigger router for reactive graphs
	triggerRouter := trigger.NewRouter(frgRepo, engine)
	triggerRouter.Start()

	// Load existing triggers on startup
	if err := triggerRouter.StartTriggerLoader(context.Background()); err != nil {
		logrus.WithError(err).Warn("Failed to load graph triggers on startup")
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
	)

	// ── Graph Definitions (CRUD) ───────────────────────────────────────────
	// Public read
	api.HandleFunc("/frg/graphs", frgHandler.ListGraphs).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", frgHandler.GetGraph).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/instances", frgHandler.ListInstances).Methods("GET", "OPTIONS")

	// AI Discovery (public search)
	api.HandleFunc("/frg/discover", frgHandler.SemanticSearch).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/optimizations", frgHandler.GetOptimizations).Methods("GET", "OPTIONS")

	// Protected write
	api.HandleFunc("/frg/graphs", authMiddleware.RequireAuth(frgHandler.CreateGraph)).Methods("POST", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", authMiddleware.RequireAuth(frgHandler.UpdateGraph)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}", authMiddleware.RequireAuth(frgHandler.DeleteGraph)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/frg/graphs/{author}/{name}/publish", authMiddleware.RequireAuth(frgHandler.PublishGraphVersion)).Methods("POST", "OPTIONS")

	// Remix/Fork
	api.HandleFunc("/frg/graphs/{author}/{name}/remix", authMiddleware.RequireAuth(frgHandler.RemixGraph)).Methods("POST", "OPTIONS")

	// ── Graph Execution ─────────────────────────────────────────────────────
	// Execute (public with basic auth check)
	api.HandleFunc("/gx/{author}/{name}", frgHandler.ExecuteGraph).Methods("POST", "OPTIONS")
	api.HandleFunc("/gx/{author}/{name}@{version}", frgHandler.ExecuteGraph).Methods("POST", "OPTIONS")

	// Instance management
	api.HandleFunc("/frg/instances/{instance_id}", frgHandler.GetInstanceStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/frg/instances/{instance_id}/stop", authMiddleware.RequireAuth(frgHandler.StopInstance)).Methods("POST", "OPTIONS")

	// ── AI Composition (protected) ──────────────────────────────────────────
	api.HandleFunc("/frg/compose", authMiddleware.RequireAuth(frgHandler.AICompose)).Methods("POST", "OPTIONS")

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
		logrus.WithError(err).Warn("Failed to register some auto-generated API routes")
	}
}
