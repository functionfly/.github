package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/github"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerGitHubRoutes wires all GitHub integration endpoints:
// connection management, repo browsing, import operations, sync, webhooks, and templates.
func registerGitHubRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware, githubHandler *github.Handler) {
	gh := api.PathPrefix("/github").Subrouter()

	// ── Connection management ─────────────────────────────────────────────
	gh.HandleFunc("/connect", authMiddleware.RequireAuth(githubHandler.HandleConnect)).Methods("GET")
	gh.HandleFunc("/callback", githubHandler.HandleCallback).Methods("GET") // OAuth callback — no auth
	gh.HandleFunc("/connection", authMiddleware.RequireAuth(githubHandler.HandleGetConnection)).Methods("GET")
	gh.HandleFunc("/connection", authMiddleware.RequireAuth(githubHandler.HandleDisconnect)).Methods("DELETE")
	gh.HandleFunc("/connection/refresh", authMiddleware.RequireAuth(githubHandler.HandleRefreshToken)).Methods("POST")

	// ── Repository browsing & scanning ────────────────────────────────────
	gh.HandleFunc("/repos", authMiddleware.RequireAuth(githubHandler.HandleListRepos)).Methods("GET")
	gh.HandleFunc("/repos/refresh", authMiddleware.RequireAuth(githubHandler.HandleRefreshRepos)).Methods("POST")
	gh.HandleFunc("/repos/{repoId}", authMiddleware.RequireAuth(githubHandler.HandleGetRepo)).Methods("GET")
	gh.HandleFunc("/repos/{repoId}/scan", authMiddleware.RequireAuth(githubHandler.HandleScanRepo)).Methods("POST")
	gh.HandleFunc("/repos/{repoId}/branches", authMiddleware.RequireAuth(githubHandler.HandleListBranches)).Methods("GET")
	gh.HandleFunc("/repos/{repoId}/tree", authMiddleware.RequireAuth(githubHandler.HandleGetTree)).Methods("GET")

	// ── Import operations ─────────────────────────────────────────────────
	gh.HandleFunc("/imports", authMiddleware.RequireAuth(githubHandler.HandleImport)).Methods("POST")
	gh.HandleFunc("/imports/preview", authMiddleware.RequireAuth(githubHandler.HandlePreviewImport)).Methods("POST")
	gh.HandleFunc("/imports/bulk", authMiddleware.RequireAuth(githubHandler.HandleBulkImport)).Methods("POST")
	gh.HandleFunc("/imports", authMiddleware.RequireAuth(githubHandler.HandleListImports)).Methods("GET")
	gh.HandleFunc("/imports/{importId}", authMiddleware.RequireAuth(githubHandler.HandleGetImport)).Methods("GET")
	gh.HandleFunc("/imports/{importId}/cancel", authMiddleware.RequireAuth(githubHandler.HandleCancelImport)).Methods("POST")
	gh.HandleFunc("/imports/{importId}/retry", authMiddleware.RequireAuth(githubHandler.HandleRetryImport)).Methods("POST")
	gh.HandleFunc("/imports/{importId}/resync", authMiddleware.RequireAuth(githubHandler.HandleResyncImport)).Methods("POST")
	gh.HandleFunc("/imports/{importId}/progress", githubHandler.HandleImportProgress).Methods("GET") // No auth middleware — uses ?token= query param for SSE

	// ── Sync management ───────────────────────────────────────────────────
	gh.HandleFunc("/imports/{importId}/sync", authMiddleware.RequireAuth(githubHandler.HandleUpdateSync)).Methods("PUT", "PATCH")
	gh.HandleFunc("/imports/{importId}/sync-logs", authMiddleware.RequireAuth(githubHandler.HandleGetSyncLogs)).Methods("GET")

	// ── Webhook receiver (HMAC-verified, no auth) ─────────────────────────
	gh.HandleFunc("/webhook", githubHandler.HandleWebhook).Methods("POST")

	// ── Import templates ──────────────────────────────────────────────────
	gh.HandleFunc("/templates", authMiddleware.RequireAuth(githubHandler.HandleListTemplates)).Methods("GET")
	gh.HandleFunc("/templates", authMiddleware.RequireAuth(githubHandler.HandleCreateTemplate)).Methods("POST")
	gh.HandleFunc("/templates/{id}", authMiddleware.RequireAuth(githubHandler.HandleUpdateTemplate)).Methods("PUT")
	gh.HandleFunc("/templates/{id}", authMiddleware.RequireAuth(githubHandler.HandleDeleteTemplate)).Methods("DELETE")
}
