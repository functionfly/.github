package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/a2a"
	a2astorage "github.com/functionfly/functionfly/internal/a2a/storage"
	"github.com/functionfly/functionfly/internal/api/handlers/mcp"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/gateway"
	gatewayreceipt "github.com/functionfly/functionfly/internal/gateway/receipt"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// registerA2ARoutes wires the A2A protocol endpoints:
//
//	GET  /.well-known/agent.json                — gateway Agent Card (public)
//	GET  /v1/a2a/agents/{agent_id}/card         — per-agent Agent Card
//	GET  /v1/a2a/agents/cards                   — list agent cards
//	POST /v1/a2a/agents/cards                   — publish agent card
//	DELETE /v1/a2a/agents/cards/{agent_id}      — unpublish agent card
//	POST /v1/a2a/{agent_id}/tasks/send          — send a task
//	GET  /v1/a2a/tasks/{task_id}                — poll task status
//	POST /v1/a2a/tasks/{task_id}/cancel         — cancel a task
//	POST /v1/a2a/tasks/{task_id}/subscribe      — SSE stream
func (s *Server) registerA2ARoutes(
	router *mux.Router,
	api *mux.Router,
	db *gorm.DB,
	registryRepo *registry.RegistryRepository,
	_ *mcp.Handler,
	authMiddleware *middleware.AuthMiddleware,
	authSvc *auth.AuthService,
) {
	// Only register A2A routes if the feature is enabled.
	if os.Getenv("A2A_ENABLED") == "false" {
		logrus.Info("A2A routes disabled via A2A_ENABLED=false")
		return
	}

	// Bearer auth for A2A task endpoints — validates JWT + API keys and
	// attaches a gateway.Caller to the request context.
	apikeyRepo := apikey.NewRepository(db)
	a2aBearerAuth := newA2ABearerAuthMiddleware(apikeyRepo, authSvc)

	// Initialize A2A storage.
	cardRepo := a2astorage.NewA2ACardRepository(db)
	taskStore := a2astorage.NewA2ATaskStore(db)

	// Initialize the receipt emitter for cross-protocol receipts.
	signingKey := []byte(os.Getenv("RECEIPT_SIGNING_KEY"))
	emitterCfg := gatewayreceipt.EmitterConfig{
		Signer: signingKey,
	}
	emitter := gatewayreceipt.NewEmitter(registryRepo, emitterCfg, logrus.New())

	// Initialize GatewayCore.
	core := gateway.NewCore(gateway.Deps{
		Auth:    nil, // Auth is handled by middleware
		Emitter: emitter,
		Logger:  logrus.New(),
	})

	// Initialize A2A task engine.
	engine := a2a.NewTaskEngine(taskStore, logrus.New())

	// Initialize A2A handler.
	a2aHandler := a2a.NewHandler(core, engine, cardRepo, logrus.New())

	// ── Well-known endpoints ────────────────────────────────────────────────

	// /.well-known/agent.json — A2A gateway Agent Card.
	router.HandleFunc("/.well-known/agent.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			gateway.SetCORSHeaders(w, r, gateway.CORSOptions{AllowMethods: "GET, OPTIONS"})
			w.WriteHeader(http.StatusNoContent)
			return
		}
		gateway.ServeGatewayCard(w, r)
	}).Methods("GET", "OPTIONS")

	// ── Agent Card endpoints ────────────────────────────────────────────────

	// GET /v1/a2a/agents/cards — list agent cards (paginated).
	api.HandleFunc("/a2a/agents/cards", a2aHandler.ListAgentCards).Methods("GET", "OPTIONS")

	// POST /v1/a2a/agents/cards — publish/update agent card.
	api.HandleFunc("/a2a/agents/cards", a2aHandler.PublishAgentCard).Methods("POST", "OPTIONS")

	// DELETE /v1/a2a/agents/cards/{agent_id} — unpublish agent card.
	api.HandleFunc("/a2a/agents/cards/{agent_id}", a2aHandler.DeleteAgentCard).Methods("DELETE", "OPTIONS")

	// GET /v1/a2a/agents/{agent_id}/card — single agent card.
	api.HandleFunc("/a2a/agents/{agent_id}/card", a2aHandler.ServeAgentCard).Methods("GET", "OPTIONS")

	// ── Task endpoints (auth required) ──────────────────────────────────────

	// POST /v1/a2a/{agent_id}/tasks/send — send a task.
	api.Handle("/a2a/{agent_id}/tasks/send", a2aBearerAuth(http.HandlerFunc(a2aHandler.SendTask))).Methods("POST", "OPTIONS")

	// GET /v1/a2a/tasks/{task_id} — poll task status.
	api.Handle("/a2a/tasks/{task_id}", a2aBearerAuth(http.HandlerFunc(a2aHandler.GetTask))).Methods("GET", "OPTIONS")

	// POST /v1/a2a/tasks/{task_id}/cancel — cancel a task.
	api.Handle("/a2a/tasks/{task_id}/cancel", a2aBearerAuth(http.HandlerFunc(a2aHandler.CancelTask))).Methods("POST", "OPTIONS")

	// POST /v1/a2a/tasks/{task_id}/subscribe — SSE stream for task updates.
	api.Handle("/a2a/tasks/{task_id}/subscribe", a2aBearerAuth(http.HandlerFunc(a2aHandler.SubscribeSSE))).Methods("POST", "OPTIONS")

	logrus.Info("A2A routes registered")
}

// registerExtendedWellKnown extends /.well-known/functionfly.json with the
// supported_protocols block for protocol negotiation.
func (s *Server) registerExtendedWellKnown(router *mux.Router) {
	base := strings.TrimRight(os.Getenv("PUBLIC_SITE_URL"), "/")
	if base == "" {
		base = strings.TrimRight(os.Getenv("BASE_URL"), "/")
	}
	if base == "" {
		base = "https://api.functionfly.com"
	}

	router.HandleFunc("/.well-known/functionfly-protocols.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			gateway.SetCORSHeaders(w, r, gateway.CORSOptions{AllowMethods: "GET, OPTIONS"})
			w.WriteHeader(http.StatusNoContent)
			return
		}
		gateway.ServeWellKnownExtended(w, r, base)
	}).Methods("GET", "OPTIONS")
}

// newA2ABearerAuthMiddleware returns an HTTP middleware that validates
// bearer tokens (JWT session tokens or API keys) and attaches a
// gateway.Caller to the request context. A2A task endpoints require
// authentication — anonymous requests are rejected with 401.
//
// This follows the same pattern as mcp.NewBearerAuthMiddleware but
// uses the gateway.Caller type directly so the A2A handler can read
// it via gateway.CallerFromContext.
func newA2ABearerAuthMiddleware(apikeyRepo *apikey.Repository, authSvc *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				writeA2AError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "bearer token required")
				return
			}

			caller := gateway.Caller{
				TokenHash: gateway.ShortHash(token),
			}

			credType := gateway.ClassifyCredential(token)
			switch credType {
			case gateway.CredentialJWT:
				if authSvc == nil {
					writeA2AError(w, http.StatusUnauthorized, "AUTH_UNAVAILABLE", "auth service unavailable")
					return
				}
				claims, err := authSvc.ValidateToken(r.Context(), token)
				if err != nil {
					writeA2AError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
					return
				}
				caller.UserID = claims.UserID.String()
				caller.TenantID = claims.TenantID.String()
				caller.AuthType = "session"

			case gateway.CredentialAPIKey, gateway.CredentialAgentKey:
				if apikeyRepo == nil {
					writeA2AError(w, http.StatusUnauthorized, "AUTH_UNAVAILABLE", "apikey service unavailable")
					return
				}
				key, err := apikeyRepo.ValidateAPIKey(token)
				if err != nil || key == nil {
					writeA2AError(w, http.StatusUnauthorized, "INVALID_API_KEY", "invalid API key")
					return
				}
				caller.APIKeyID = key.ID.String()
				if key.UserID != uuid.Nil {
					caller.UserID = key.UserID.String()
				}
				if key.TenantID != uuid.Nil {
					caller.TenantID = key.TenantID.String()
				}
				if credType == gateway.CredentialAgentKey {
					caller.AuthType = "agent"
				} else {
					caller.AuthType = "apikey"
				}

			default:
				writeA2AError(w, http.StatusUnauthorized, "INVALID_CREDENTIAL", "unrecognized credential format")
				return
			}

			ctx := gateway.WithCaller(r.Context(), caller)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// bearerToken extracts the token from the Authorization header.
func bearerToken(header string) string {
	if len(header) > 7 && (header[:7] == "Bearer " || header[:7] == "bearer ") {
		return header[7:]
	}
	return ""
}

// writeA2AError writes a JSON error response for A2A protocol errors.
func writeA2AError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":{"code":"%s","message":"%s"}}`, code, message)
}
