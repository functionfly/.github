package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/workflow"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerWorkflowRoutes(
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	workflowHandler *workflow.Handler,
) {
	protected.HandleFunc("/workflow/graph", authMiddleware.RequireAuth(workflowHandler.HandleGetGraph)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workflow/graph", authMiddleware.RequireAuth(workflowHandler.HandleCreateGraph)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/workflow/graph/{id}", authMiddleware.RequireAuth(workflowHandler.HandleUpdateGraph)).Methods("PUT", "OPTIONS")

	protected.HandleFunc("/workflow/nodes", authMiddleware.RequireAuth(workflowHandler.HandleListNodes)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workflow/nodes", authMiddleware.RequireAuth(workflowHandler.HandleCreateNode)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/workflow/nodes/{nodeId}", authMiddleware.RequireAuth(workflowHandler.HandleUpdateNode)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/workflow/nodes/{nodeId}", authMiddleware.RequireAuth(workflowHandler.HandleDeleteNode)).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/workflow/edges", authMiddleware.RequireAuth(workflowHandler.HandleListEdges)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workflow/edges", authMiddleware.RequireAuth(workflowHandler.HandleCreateEdge)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/workflow/edges/{edgeId}", authMiddleware.RequireAuth(workflowHandler.HandleUpdateEdge)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/workflow/edges/{edgeId}", authMiddleware.RequireAuth(workflowHandler.HandleDeleteEdge)).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/workflow/executions", authMiddleware.RequireAuth(workflowHandler.HandleListExecutions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workflow/executions/{id}", authMiddleware.RequireAuth(workflowHandler.HandleGetExecution)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workflow/executions/{id}/cancel", authMiddleware.RequireAuth(workflowHandler.HandleCancelExecution)).Methods("POST", "OPTIONS")

	protected.HandleFunc("/workflow/execute", authMiddleware.RequireAuth(workflowHandler.HandleExecuteWorkflow)).Methods("POST", "OPTIONS")

	protected.HandleFunc("/workflow/{graphId}/timeline", authMiddleware.RequireAuth(workflowHandler.HandleGetTimeline)).Methods("GET", "OPTIONS")
}