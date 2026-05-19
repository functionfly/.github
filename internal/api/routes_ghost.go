package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/ghost"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerGhostRoutes wires Ghost Mode autonomous building engine endpoints.
func registerGhostRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	ghostHandler *ghost.Handler,
) {
	// Build lifecycle
	api.HandleFunc("/v1/ghost/builds", authMiddleware.RequireAuth(ghostHandler.HandleCreateBuild)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds", authMiddleware.RequireAuth(ghostHandler.HandleListBuilds)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}", authMiddleware.RequireAuth(ghostHandler.HandleGetBuild)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}", authMiddleware.RequireAuth(ghostHandler.HandleUpdateBuild)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}", authMiddleware.RequireAuth(ghostHandler.HandleDeleteBuild)).Methods("DELETE", "OPTIONS")

	// Task management
	api.HandleFunc("/v1/ghost/builds/{id}/tasks", authMiddleware.RequireAuth(ghostHandler.HandleListTasks)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/tasks/{task_id}/start", authMiddleware.RequireAuth(ghostHandler.HandleStartTask)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/tasks/{task_id}/complete", authMiddleware.RequireAuth(ghostHandler.HandleCompleteTask)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/tasks/{task_id}/fail", authMiddleware.RequireAuth(ghostHandler.HandleFailTask)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/tasks/{task_id}/logs", authMiddleware.RequireAuth(ghostHandler.HandleAddTaskLog)).Methods("POST", "OPTIONS")

	// Human approval
	api.HandleFunc("/v1/ghost/builds/{id}/approve", authMiddleware.RequireAuth(ghostHandler.HandleApproval)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/pause", authMiddleware.RequireAuth(ghostHandler.HandlePauseBuild)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/resume", authMiddleware.RequireAuth(ghostHandler.HandleResumeBuild)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/builds/{id}/cancel", authMiddleware.RequireAuth(ghostHandler.HandleCancelBuild)).Methods("POST", "OPTIONS")

	// Architecture planning
	api.HandleFunc("/v1/ghost/plan/architecture", authMiddleware.RequireAuth(ghostHandler.HandlePlanArchitecture)).Methods("POST", "OPTIONS")

	// Infra provisioning
	api.HandleFunc("/v1/ghost/provision/database", authMiddleware.RequireAuth(ghostHandler.HandleProvisionDatabase)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/provision/api", authMiddleware.RequireAuth(ghostHandler.HandleProvisionAPI)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/provision/docker", authMiddleware.RequireAuth(ghostHandler.HandleProvisionDocker)).Methods("POST", "OPTIONS")

	// Code generation
	api.HandleFunc("/v1/ghost/generate/schema", authMiddleware.RequireAuth(ghostHandler.HandleGenerateSchema)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/generate/backend", authMiddleware.RequireAuth(ghostHandler.HandleGenerateBackend)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/generate/frontend", authMiddleware.RequireAuth(ghostHandler.HandleGenerateFrontend)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/generate/tests", authMiddleware.RequireAuth(ghostHandler.HandleGenerateTests)).Methods("POST", "OPTIONS")

	// Deployment
	api.HandleFunc("/v1/ghost/deploy/staging", authMiddleware.RequireAuth(ghostHandler.HandleDeployStaging)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/ghost/deploy/production", authMiddleware.RequireAuth(ghostHandler.HandleDeployProduction)).Methods("POST", "OPTIONS")

	// Monitoring
	api.HandleFunc("/v1/ghost/monitor/setup", authMiddleware.RequireAuth(ghostHandler.HandleSetupMonitoring)).Methods("POST", "OPTIONS")
}

// suppressUnused silences unused import warnings
var _ http.HandlerFunc