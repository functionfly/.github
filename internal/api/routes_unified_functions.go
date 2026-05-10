package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	registryrepo "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
)

// registerUnifiedFunctionRoutes wires /fx/* routes that provide unified access
// to both public registry functions and tenant-private functions.
//
// URL Structure:
//   - /fx/{author}/{name}                    - Public registry functions
//   - /fx/{tenant_id}/{name}                 - Tenant-private functions
//   - /fx/{author}/{name}@{version}          - Versioned execution
//
// This consolidates the separate registry (HandleListFunctions, HandleExecute, etc.)
// and platform (HandleListFunctions, HandleDeployFunction, etc.) function handlers.
func registerUnifiedFunctionRoutes(
	s *Server,
	root *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	executionSecurityMW *middleware.ExecutionCoordinatorMiddleware,
	registryRepo *registryrepo.RegistryRepository,
	registryHandler *registry.Handler,
	advancedSecurityMiddleware *advanced_security.AdvancedSecurityMiddleware,
) {
	// ── Unified Function Execution ───────────────────────────────────────────
	// POST /fx/{author}/{name} - Execute public registry function
	// POST /fx/{author}/{name}@{version} - Execute specific version
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
	root.HandleFunc("/fx/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
	root.HandleFunc("/fx/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")

	// ── Unified Function Listing (public) ────────────────────────────────────
	// GET /fx/search - Search all public functions
	// GET /fx/trending - Trending functions
	root.HandleFunc("/fx/search", registryHandler.HandleSearchFunctions).Methods("GET", "OPTIONS")
	root.HandleFunc("/fx/trending", registryHandler.HandleGetTrendingFunctions).Methods("GET", "OPTIONS")

	// ── Unified Function Info ────────────────────────────────────────────────
	// GET /fx/{author}/{name} - Get function info and metadata
	root.HandleFunc("/fx/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET", "OPTIONS")
	root.HandleFunc("/fx/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET", "OPTIONS")
	root.HandleFunc("/fx/{author}/{name}/stats", registryHandler.HandleGetFunctionStats).Methods("GET", "OPTIONS")
	root.HandleFunc("/fx/{author}/{name}/trust", registryHandler.HandleGetFunctionTrustByAuthorName).Methods("GET", "OPTIONS")
	root.HandleFunc("/fx/{author}/{name}/source", registryHandler.HandleGetFunctionSource).Methods("GET", "OPTIONS")

	// ── Unified Function Publishing (authenticated) ──────────────────────────
	// POST /fx/functions/publish - Publish new function or version
	root.HandleFunc("/fx/functions/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST", "OPTIONS")

	// ── Unified Function Testing ──────────────────────────────────────────────
	// POST /fx/{author}/{name}/test - Test function with sample input
	root.HandleFunc("/fx/{author}/{name}/test", registryHandler.HandleTest).Methods("POST", "OPTIONS")

	// ── Unified Function Management (authenticated) ─────────────────────────
	// DELETE /fx/{author}/{name} - Delete function
	root.HandleFunc("/fx/{author}/{name}", authMiddleware.RequireAuth(registryHandler.HandleDeleteFunction)).Methods("DELETE", "OPTIONS")

	// ── Reviews & Ratings ─────────────────────────────────────────────────────
	// GET /fx/{author}/{name}/reviews - List reviews (public)
	root.HandleFunc("/fx/{author}/{name}/reviews", registryHandler.HandleListReviews).Methods("GET", "OPTIONS")
	// POST /fx/{author}/{name}/reviews - Submit review (authenticated)
	root.HandleFunc("/fx/{author}/{name}/reviews", authMiddleware.RequireAuth(registryHandler.HandleSubmitReview)).Methods("POST", "OPTIONS")
	// POST /fx/{author}/{name}/rating - Submit rating (authenticated)
	root.HandleFunc("/fx/{author}/{name}/rating", authMiddleware.RequireAuth(registryHandler.HandleSubmitRating)).Methods("POST", "OPTIONS")

	// ── Remix (fork) ──────────────────────────────────────────────────────────
	// GET /fx/{author}/{name}/remix/cost - Get remix cost
	root.HandleFunc("/fx/{author}/{name}/remix/cost", authMiddleware.RequireAuth(registryHandler.HandleGetRemixCost)).Methods("GET", "OPTIONS")
	// POST /fx/{author}/{name}/remix - Remix/fork function (authenticated)
	root.HandleFunc("/fx/{author}/{name}/remix", authMiddleware.RequireAuth(registryHandler.HandleRemix)).Methods("POST", "OPTIONS")
	// GET /fx/{author}/{name}/remix/history - Get remix history
	root.HandleFunc("/fx/{author}/{name}/remix/history", authMiddleware.RequireAuth(registryHandler.HandleGetRemixHistory)).Methods("GET", "OPTIONS")

	// ── Embed ─────────────────────────────────────────────────────────────────
	// GET /fx/{author}/{name}/embed/snippet - Get embed script
	root.HandleFunc("/fx/{author}/{name}/embed/snippet", registryHandler.HandleGetEmbedSnippet).Methods("GET", "OPTIONS")
	// GET /fx/{author}/{name}/embed - Get embed config
	root.HandleFunc("/fx/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleGetEmbedConfig)).Methods("GET", "OPTIONS")
	// PUT /fx/{author}/{name}/embed - Update embed config
	root.HandleFunc("/fx/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleUpdateEmbedConfig)).Methods("PUT", "OPTIONS")

	// ── Settings ──────────────────────────────────────────────────────────────
	// GET /fx/{author}/{name}/settings
	root.HandleFunc("/fx/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandleGetFunctionSettings)).Methods("GET", "OPTIONS")
	// PATCH /fx/{author}/{name}/settings
	root.HandleFunc("/fx/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandlePatchFunctionSettings)).Methods("PATCH", "OPTIONS")

	// ── Playground ─────────────────────────────────────────────────────────────
	// GET /fx/{author}/{name}/playground - Open playground UI
	// Note: Playground handler is set separately in registerRegistryRoutes
}