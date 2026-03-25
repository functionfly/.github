package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/trustapi"
	"github.com/functionfly/functionfly/internal/storage/registry"
	trustapirepo "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/gorilla/mux"
)

// registerTrustAPIRoutes wires all Trust API routes for external platform partners
func registerTrustAPIRoutes(
	s *Server,
	api *mux.Router,
	registryRepo *registry.RegistryRepository,
) {
	// Initialize Trust API repository
	trustRepo := trustapirepo.NewRepository(s.postgresDB.GORM)

	// Initialize handlers
	trustHandler := trustapi.NewHandler(trustRepo, registryRepo)

	// Initialize middleware
	authMiddleware := trustapi.NewAPIKeyAuthMiddleware(trustRepo)
	rateLimitMiddleware := trustapi.NewRateLimitMiddleware(trustRepo)
	usageTrackingMiddleware := trustapi.NewUsageTrackingMiddleware(trustRepo)

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
}
