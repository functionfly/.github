package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/trustapi"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/registry"
	trustapirepo "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerTrustAPIRoutes wires all Trust API routes for external platform partners
func registerTrustAPIRoutes(
	s *Server,
	api *mux.Router,
	registryRepo *registry.RegistryRepository,
) {
	// Initialize Trust API repository
	trustRepo := trustapirepo.NewRepository(s.postgresDB.GORM)

	// Initialize revocation repository for extended functionality
	revocationRepo := trustapirepo.NewRevocationRepository(s.postgresDB.GORM)

	// Initialize webhook repository and service
	webhookRepo := trustapirepo.NewWebhookRepository(s.postgresDB.GORM)
	webhookService := trustapirepo.NewWebhookService(webhookRepo)
	webhookService.SetLogger(logrus.New())

	// Initialize handlers
	trustHandler := trustapi.NewHandler(trustRepo, registryRepo)
	webhookHandler := trustapi.NewWebhookHandler(webhookRepo)
	webhookHandler.SetLogger(logrus.New())

	// Initialize extended handler with revocation and webhook capabilities
	extendedHandler := trustapi.NewExtendedHandler(trustRepo, registryRepo, revocationRepo, webhookService)

	// Initialize streaming handler
	trustStreamer := trustapi.NewTrustScoreStreamer(registryRepo, logrus.New())
	go trustStreamer.Run()

	// Initialize middleware
	authMiddleware := trustapi.NewAPIKeyAuthMiddleware(trustRepo)
	rateLimitMiddleware := trustapi.NewRateLimitMiddleware(trustRepo)
	usageTrackingMiddleware := trustapi.NewUsageTrackingMiddleware(trustRepo)

	// Get internal auth middleware for JWT-protected routes
	internalAuthMiddleware := middleware.NewAuthMiddleware(s.authSvc)

	// Create a subrouter for Trust API
	trustAPI := api.PathPrefix("/v1").Subrouter()

	// Apply global Trust API middleware
	// Order: Usage Tracking -> Rate Limiting -> Authentication
	trustAPI.Use(usageTrackingMiddleware.Track())
	trustAPI.Use(rateLimitMiddleware.RateLimit())

	// Partner registration and management (public - no auth required for create/list)
	trustAPI.HandleFunc("/partners", trustHandler.HandleCreatePartner).Methods("POST")
	trustAPI.HandleFunc("/partners", trustHandler.HandleListPartners).Methods("GET")
	trustAPI.HandleFunc("/partners/{partner_id}", trustHandler.HandleGetPartner).Methods("GET")
	trustAPI.HandleFunc("/partners/{partner_id}", trustHandler.HandleUpdatePartner).Methods("PATCH")
	trustAPI.HandleFunc("/partners/{partner_id}/usage", trustHandler.HandleGetPartnerUsage).Methods("GET")

	// API Key management (requires partner auth)
	apiKeyAuthRouter := trustAPI.PathPrefix("/partners/{partner_id}/api-keys").Subrouter()
	apiKeyAuthRouter.Use(authMiddleware.Authenticate())
	apiKeyAuthRouter.HandleFunc("", trustHandler.HandleCreateAPIKey).Methods("POST")
	apiKeyAuthRouter.HandleFunc("", trustHandler.HandleListAPIKeys).Methods("GET")
	apiKeyAuthRouter.HandleFunc("/{key_id}", trustHandler.HandleRevokeAPIKey).Methods("DELETE")

	// Trust endpoints (requires API key authentication)
	trustAuthRouter := trustAPI.PathPrefix("/trust").Subrouter()
	trustAuthRouter.Use(authMiddleware.Authenticate())

	// Trust score endpoints
	trustAuthRouter.HandleFunc("/score/{function_id}", trustHandler.HandleGetTrustScore).Methods("GET")
	trustAuthRouter.HandleFunc("/batch", trustHandler.HandleBatchTrustScore).Methods("POST")
	trustAuthRouter.HandleFunc("/history/{function_id}", trustHandler.HandleGetTrustHistory).Methods("GET")

	// Trust verification endpoints (requires verification:request scope)
	trustAuthRouter.Handle("/verify",
		authMiddleware.RequireScope("verification:request")(http.HandlerFunc(trustHandler.HandleSubmitVerification))).Methods("POST")
	trustAuthRouter.HandleFunc("/verify/{verification_id}", trustHandler.HandleGetVerification).Methods("GET")

	// Trust reporting endpoints (requires reports:submit scope)
	trustAuthRouter.Handle("/report",
		authMiddleware.RequireScope("reports:submit")(http.HandlerFunc(trustHandler.HandleSubmitReport))).Methods("POST")
	trustAuthRouter.HandleFunc("/report/{report_id}", trustHandler.HandleGetReport).Methods("GET")

	// ============================================
	// Extended Trust API Routes (New Functionality)
	// ============================================

	// Attestation endpoints (API key auth)
	trustAuthRouter.HandleFunc("/attestations", extendedHandler.HandleGetAttestations).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/{attestation_id}", extendedHandler.HandleGetAttestation).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/{attestation_id}/verify", extendedHandler.HandleVerifyAttestation).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/chain/{function_id}", extendedHandler.HandleGetAttestationChain).Methods("GET")

	// Policy endpoints (public read/evaluate, JWT for write)
	trustAuthRouter.HandleFunc("/policies", extendedHandler.HandleListPolicies).Methods("GET")
	trustAuthRouter.HandleFunc("/policies/{policy_id}", extendedHandler.HandleGetPolicy).Methods("GET")
	trustAuthRouter.HandleFunc("/policies/evaluate", extendedHandler.HandleEvaluatePolicy).Methods("POST")
	trustAuthRouter.HandleFunc("/policies/evaluate/batch", extendedHandler.HandleBatchEvaluatePolicy).Methods("POST")

	// Policy management requires JWT authentication
	trustAuthRouter.Handle("/policies",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleCreatePolicy)).Methods("POST")
	trustAuthRouter.Handle("/policies/{policy_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleUpdatePolicy)).Methods("PUT")
	trustAuthRouter.Handle("/policies/{policy_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleDeletePolicy)).Methods("DELETE")

	// Revocation endpoints require admin/internal JWT authentication
	revokeRouter := trustAPI.PathPrefix("/trust/revoke").Subrouter()

	// Revocation listing/checking (read-only, requires auth)
	revokeRouter.Handle("/revoked",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleListRevocations)).Methods("GET")
	revokeRouter.Handle("/revoked/{function_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleCheckFunctionRevoked)).Methods("GET")
	revokeRouter.Handle("/{revocation_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleGetRevocation)).Methods("GET")

	// ============================================
	// Webhook Management Routes (JWT authenticated)
	// ============================================
	webhookRouter := api.PathPrefix("/v1/webhooks").Subrouter()

	// Webhook CRUD (all require auth)
	webhookRouter.Handle("",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleCreateWebhook)).Methods("POST")
	webhookRouter.Handle("",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleListWebhooks)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleGetWebhook)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleUpdateWebhook)).Methods("PUT")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleDeleteWebhook)).Methods("DELETE")

	// Webhook testing and stats
	webhookRouter.Handle("/{webhook_id}/test",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleTestWebhook)).Methods("POST")
	webhookRouter.Handle("/{webhook_id}/deliveries",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleListDeliveries)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}/stats",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleGetDeliveryStats)).Methods("GET")

	// ============================================
	// Real-time Trust Score Streaming Routes
	// ============================================
	// SSE streaming for trust score updates (requires JWT auth)

	// SSE endpoint for streaming trust score updates for watched functions
	api.HandleFunc("/v1/trust/stream/sse",
		internalAuthMiddleware.RequireAuth(trustStreamer.HandleSSE)).Methods("GET")

	// SSE endpoint for a specific function
	api.HandleFunc("/v1/trust/stream/functions/{function_id}/sse",
		internalAuthMiddleware.RequireAuth(trustStreamer.HandleSSEFunction)).Methods("GET")
}
