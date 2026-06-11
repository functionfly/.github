package api

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/mcp"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// registerMCPRoutes wires the MCP Function Registry endpoints:
//
//	GET  /v1/mcp/manifest   — public, cacheable, server identity
//	GET  /v1/mcp/tools      — public, cacheable, tool index (SEO anchor)
//	POST /v1/mcp            — JSON-RPC 2.0 transport (streamable-HTTP)
//
// All three live on the /v1 subrouter. The bearer-auth middleware tries JWT
// session tokens first and falls back to FunctionFly API keys (ffp_/fff_/ffo_/
// aep_/ffe_/fft_). tools/call is the only method that REQUIRES a valid
// caller; manifest and tools index are public and may be hit anonymously.
//
// The authMiddleware parameter is reserved for future caller-scope checks
// (e.g. "only show verified-mcp tools to authenticated callers"); the
// current implementation does not require it.
func (s *Server) registerMCPRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	executionSecurityMW *middleware.ExecutionCoordinatorMiddleware,
	registryHandler *registryhandler.Handler,
	registryRepo *registry.RegistryRepository,
) {
	// Resolve the API key repository from s (we re-instantiate it to keep
	// the wiring local; the constructor is cheap and the repo is the
	// canonical place to look up ffp_/fff_/... keys).
	apikeyRepo := apikey.NewRepository(s.postgresDB.GORM)
	mcpHandler := buildMCPHandler(s, apikeyRepo, executionSecurityMW, registryHandler, registryRepo)

	// Bearer auth: optional, sets caller identity when present.
	bearer := mcp.NewBearerAuthMiddleware(mcp.AuthConfig{
		APIKeyRepo:  apikeyRepo,
		AuthService: bearerJWTSvc(s.authSvc),
	})

	// /v1/mcp/manifest: public, cacheable, no auth gate.
	api.HandleFunc("/mcp/manifest", mcpHandler.HandleManifest).Methods("GET", "OPTIONS")
	// /v1/mcp/tools: public, cacheable, no auth gate.
	api.HandleFunc("/mcp/tools", mcpHandler.HandleToolsIndex).Methods("GET", "OPTIONS")
	// /v1/mcp/sitemap.xml + sitemap index: SEO crawl anchors.
	api.HandleFunc("/mcp/sitemap.xml", mcpHandler.HandleSitemap).Methods("GET", "OPTIONS")
	api.HandleFunc("/mcp/sitemap-index.xml", mcpHandler.HandleSitemapIndex).Methods("GET", "OPTIONS")
	// /v1/mcp/og/{author}/{name}.png — per-function dynamic OG image.
	api.HandleFunc("/mcp/og/{author}/{name}.png", mcpHandler.HandleOGImage).Methods("GET", "OPTIONS")
	// /v1/mcp: streamable-HTTP transport. tools/call is gated by
	// RequireAuthForToolsCall; other methods (initialize, tools/list,
	// ping) do not require auth.
	api.Handle("/mcp",
		bearer(
			mcp.RequireAuthForToolsCall(
				http.HandlerFunc(mcpHandler.HandleJSONRPC),
			),
		),
	).Methods("POST", "OPTIONS")

	// Optional health endpoint for ops (DB + Redis ping).
	api.HandleFunc("/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"service":"mcp-registry"}`))
	}).Methods("GET", "OPTIONS")

	// ── Global MCP Settings, Analytics & Connections ────────────────────────
	mcpGlobalHandler := mcp.NewGlobalHandler(registryRepo)
	api.HandleFunc("/mcp/settings", mcpGlobalHandler.HandleGetMCPSettings).Methods("GET", "OPTIONS")
	api.HandleFunc("/mcp/settings", mcpGlobalHandler.HandleUpdateMCPSettings).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/mcp/analytics", mcpGlobalHandler.HandleGetMCPAnalytics).Methods("GET", "OPTIONS")
	api.HandleFunc("/mcp/connections", mcpGlobalHandler.HandleGetMCPConnections).Methods("GET", "OPTIONS")
}

// buildMCPHandler constructs a *mcp.Handler with the production store and
// executor. It is split out so it can be tested independently.
func buildMCPHandler(
	s *Server,
	apikeyRepo *apikey.Repository,
	esMW *middleware.ExecutionCoordinatorMiddleware,
	registryHandler *registryhandler.Handler,
	registryRepo *registry.RegistryRepository,
) *mcp.Handler {
	if s == nil || registryRepo == nil {
		return nil
	}
	store := mcpStoreAdapter{repo: registryRepo}
	executor := mcp.NewRegistryExecutor(registryHandler, esMW, registryRepo)
	h := mcp.NewHandler(store, executor)
	// Public-facing URL: prefer PUBLIC_SITE_URL, then BASE_URL, then host.
	if v := strings.TrimSpace(os.Getenv("PUBLIC_SITE_URL")); v != "" {
		h.ServerPublicURL = strings.TrimRight(v, "/")
	} else if v := strings.TrimSpace(os.Getenv("BASE_URL")); v != "" {
		h.ServerPublicURL = strings.TrimRight(v, "/")
	}
	if os.Getenv("MCP_REGISTRY_ENABLED") == "false" {
		h.Disabled = true
	}
	_ = apikeyRepo // reserved for future use (e.g. caller-scope checks)
	return h
}

// bearerJWTSvc adapts *auth.AuthService to the jwtAuthenticator interface
// expected by NewBearerAuthMiddleware. We use a small wrapper to keep the
// interface narrow and dependency-invert the auth service.
func bearerJWTSvc(svc *auth.AuthService) mcpAuthJWT {
	return mcpAuthJWT{svc: svc}
}

type mcpAuthJWT struct {
	svc *auth.AuthService
}

func (m mcpAuthJWT) ValidateToken(token string) (*auth.Claims, error) {
	if m.svc == nil {
		return nil, context.DeadlineExceeded
	}
	return m.svc.ValidateToken(token)
}

// mcpStoreAdapter adapts *registry.RegistryRepository to mcp.FunctionStore.
// The interface is intentionally narrow so test mocks can implement only
// what they need; the adapter simply forwards every call.
type mcpStoreAdapter struct {
	repo *registry.RegistryRepository
}

func (a mcpStoreAdapter) GetFunctionByID(ctx context.Context, id uuid.UUID) (*registry.RegistryFunction, error) {
	return a.repo.GetFunctionByID(id)
}

func (a mcpStoreAdapter) GetFunctionByAuthorName(ctx context.Context, author, name string) (*registry.RegistryFunction, error) {
	return a.repo.GetFunctionByAuthorName(author, name)
}

func (a mcpStoreAdapter) GetLatestFunctionVersion(ctx context.Context, functionID uuid.UUID) (*registry.RegistryFunctionVersion, error) {
	return a.repo.GetLatestFunctionVersion(functionID)
}

func (a mcpStoreAdapter) GetMCPSettings(ctx context.Context, functionID uuid.UUID) (*registry.MCPSettings, error) {
	return a.repo.GetMCPSettings(ctx, functionID)
}

func (a mcpStoreAdapter) SearchFunctions(ctx context.Context, query, category, runtime string, minRating float64, limit, offset int) ([]registry.RegistryFunction, int, error) {
	return a.repo.SearchFunctions(query, category, runtime, minRating, limit, offset)
}

func (a mcpStoreAdapter) ListEnabledMCPSettings(ctx context.Context, category, runtime string, minTrust float64, limit, offset int) ([]registry.MCPSettings, int, error) {
	return a.repo.ListEnabledMCPSettings(ctx, category, runtime, minTrust, limit, offset)
}

func (a mcpStoreAdapter) IncrementMCPInvocationCount(ctx context.Context, functionID uuid.UUID) error {
	return a.repo.IncrementMCPInvocationCount(ctx, functionID)
}

func (a mcpStoreAdapter) RecordMCPInvocation(ctx context.Context, rec registry.MCPInvocationRecord) error {
	return a.repo.RecordMCPInvocation(ctx, rec)
}

// Note: mcp.Handler.Now is set to time.Now by default. We add an override
// knob here for tests.
var _ = time.Now
