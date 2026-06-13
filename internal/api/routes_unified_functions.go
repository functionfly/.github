package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	registryrepo "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
)

// registerUnifiedFunctionRoutes wires /v1/fx/* routes that provide unified access
// to both public registry functions and tenant-private functions.
//
// URL Structure:
//   - /v1/fx/{author}/{name}                    - Public registry functions
//   - /v1/fx/{tenant_id}/{name}                 - Tenant-private functions
//   - /v1/fx/{author}/{name}@{version}          - Versioned execution
//
// This consolidates the separate registry (HandleListFunctions, HandleExecute, etc.)
// and platform (HandleListFunctions, HandleDeployFunction, etc.) function handlers.
func registerUnifiedFunctionRoutes(
	s *Server,
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	executionSecurityMW *middleware.ExecutionCoordinatorMiddleware,
	registryRepo *registryrepo.RegistryRepository,
	registryHandler *registry.Handler,
	advancedSecurityMiddleware *advanced_security.AdvancedSecurityMiddleware,
) {
	// ── Unified Function Execution ───────────────────────────────────────────
	// POST /v1/fx/{author}/{name} - Execute public registry function
	// POST /v1/fx/{author}/{name}@{version} - Execute specific version
	secureExecuteHandler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fn, err := registryRepo.GetFunctionByAuthorName(vars["author"], vars["name"])
		if err != nil {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		version := vars["version"]
		executionSecurityMW.SecureExecution(fn.ID, version)(registryHandler.HandleExecute).ServeHTTP(w, r)
	}

	// Execute without version verification (latest)
	api.HandleFunc("/fx/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")

	// ── Unified Function Listing (public) ────────────────────────────────────
	// GET /v1/fx/search - Search all public functions
	// GET /v1/fx/trending - Trending functions
	api.HandleFunc("/fx/search", registryHandler.HandleSearchFunctions).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/trending", registryHandler.HandleGetTrendingFunctions).Methods("GET", "OPTIONS")

	// ── Unified Function Info ────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name} - Get function info and metadata
	api.HandleFunc("/fx/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/stats", registryHandler.HandleGetFunctionStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/trust", registryHandler.HandleGetFunctionTrustByAuthorName).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/source", registryHandler.HandleGetFunctionSource).Methods("GET", "OPTIONS")

	// ── Unified Function Publishing (authenticated) ──────────────────────────
	// POST /v1/fx/publish - Publish new function or version
	api.HandleFunc("/fx/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST", "OPTIONS")

	// ── Unified Function Testing ──────────────────────────────────────────────
	// POST /v1/fx/{author}/{name}/test - Test function with sample input
	api.HandleFunc("/fx/{author}/{name}/test", registryHandler.HandleTest).Methods("POST", "OPTIONS")

	// ── Unified Function Management (authenticated) ─────────────────────────
	// DELETE /v1/fx/{author}/{name} - Delete function
	api.HandleFunc("/fx/{author}/{name}", authMiddleware.RequireAuth(registryHandler.HandleDeleteFunction)).Methods("DELETE", "OPTIONS")

	// ── Reviews & Ratings ─────────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name}/reviews - List reviews (public)
	api.HandleFunc("/fx/{author}/{name}/reviews", registryHandler.HandleListReviews).Methods("GET", "OPTIONS")
	// POST /v1/fx/{author}/{name}/reviews - Submit review (authenticated)
	api.HandleFunc("/fx/{author}/{name}/reviews", authMiddleware.RequireAuth(registryHandler.HandleSubmitReview)).Methods("POST", "OPTIONS")
	// POST /v1/fx/{author}/{name}/rating - Submit rating (authenticated)
	api.HandleFunc("/fx/{author}/{name}/rating", authMiddleware.RequireAuth(registryHandler.HandleSubmitRating)).Methods("POST", "OPTIONS")

	// ── Remix (fork) ──────────────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name}/remix/cost - Get remix cost
	api.HandleFunc("/fx/{author}/{name}/remix/cost", authMiddleware.RequireAuth(registryHandler.HandleGetRemixCost)).Methods("GET", "OPTIONS")
	// POST /v1/fx/{author}/{name}/remix - Remix/fork function (authenticated)
	api.HandleFunc("/fx/{author}/{name}/remix", authMiddleware.RequireAuth(registryHandler.HandleRemix)).Methods("POST", "OPTIONS")
	// GET /v1/fx/{author}/{name}/remix/history - Get remix history
	api.HandleFunc("/fx/{author}/{name}/remix/history", authMiddleware.RequireAuth(registryHandler.HandleGetRemixHistory)).Methods("GET", "OPTIONS")

	// ── Embed ─────────────────────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name}/embed/snippet - Get embed script
	api.HandleFunc("/fx/{author}/{name}/embed/snippet", registryHandler.HandleGetEmbedSnippet).Methods("GET", "OPTIONS")
	// GET /v1/fx/{author}/{name}/embed - Get embed config
	api.HandleFunc("/fx/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleGetEmbedConfig)).Methods("GET", "OPTIONS")
	// PUT /v1/fx/{author}/{name}/embed - Update embed config
	api.HandleFunc("/fx/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleUpdateEmbedConfig)).Methods("PUT", "OPTIONS")

	// ── Settings ──────────────────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name}/settings
	api.HandleFunc("/fx/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandleGetFunctionSettings)).Methods("GET", "OPTIONS")
	// PATCH /v1/fx/{author}/{name}/settings
	api.HandleFunc("/fx/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandlePatchFunctionSettings)).Methods("PATCH", "OPTIONS")

	// ── Playground ─────────────────────────────────────────────────────────────
	// GET /v1/fx/{author}/{name}/playground - Open playground UI
	// Note: Playground handler is set separately in registerRegistryRoutes
}