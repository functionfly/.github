package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/agent_observability"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerAgentObservabilityRoutes(
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	agentObsHandler *agent_observability.Handler,
) {
	api.HandleFunc("/agent-observability/runs", agentObsHandler.HandleListRuns).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs", authMiddleware.RequireAuth(agentObsHandler.HandleCreateRun)).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}", authMiddleware.RequireAuth(agentObsHandler.HandleGetRun)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/events", authMiddleware.RequireAuth(agentObsHandler.HandleGetEvents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/replay", authMiddleware.RequireAuth(agentObsHandler.HandleReplay)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/stream", authMiddleware.RequireAuth(agentObsHandler.HandleStreamEvents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/stats", authMiddleware.RequireAuth(agentObsHandler.HandleGetStats)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/graph", authMiddleware.RequireAuth(agentObsHandler.HandleGetGraph)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/end", authMiddleware.RequireAuth(agentObsHandler.HandleEndRun)).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/span", authMiddleware.RequireAuth(agentObsHandler.HandleCreateSpan)).Methods("POST", "OPTIONS")
	api.HandleFunc("/agent-observability/runs/{id}/spans", authMiddleware.RequireAuth(agentObsHandler.HandleListSpans)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/config", authMiddleware.RequireAuth(agentObsHandler.HandleGetConfig)).Methods("GET", "OPTIONS")
	api.HandleFunc("/agent-observability/config", authMiddleware.RequireAuth(agentObsHandler.HandleUpdateConfig)).Methods("PUT", "OPTIONS")
}
