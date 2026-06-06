package api

import (
	brainhandler "github.com/functionfly/functionfly/internal/api/handlers/brain"
	connectorhandler "github.com/functionfly/functionfly/internal/api/handlers/connectors"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerConnectorRoutes(
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	connectorHandler *connectorhandler.Handler,
) {
	// Public: catalog
	api.HandleFunc("/connectors/catalog", connectorHandler.HandleListCatalog).Methods("GET", "OPTIONS")

	// Protected: user connector CRUD
	protected.HandleFunc("/connectors",
		authMiddleware.RequireAuth(connectorHandler.HandleListUserConnectors),
	).Methods("GET", "OPTIONS")

	protected.HandleFunc("/connectors/link",
		authMiddleware.RequireAuth(connectorHandler.HandleLinkConnector),
	).Methods("POST", "OPTIONS")

	protected.HandleFunc("/connectors/callback",
		authMiddleware.RequireAuth(connectorHandler.HandleOAuthCallback),
	).Methods("POST", "OPTIONS")

	protected.HandleFunc("/connectors/{connector_id}",
		authMiddleware.RequireAuth(connectorHandler.HandleUnlinkConnector),
	).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/connectors/{connector_id}/sync",
		authMiddleware.RequireAuth(connectorHandler.HandleTriggerSync),
	).Methods("POST", "OPTIONS")
}

func registerBrainRoutes(
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	brainHandler *brainhandler.Handler,
) {
	// Signal CRUD
	protected.HandleFunc("/brain/signals",
		authMiddleware.RequireAuth(brainHandler.HandleListSignals),
	).Methods("GET", "OPTIONS")

	protected.HandleFunc("/brain/signals/search",
		authMiddleware.RequireAuth(brainHandler.HandleSearchSignals),
	).Methods("GET", "OPTIONS")

	protected.HandleFunc("/brain/signals/{signal_id}",
		authMiddleware.RequireAuth(brainHandler.HandleDeleteSignal),
	).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/brain/signals/purge",
		authMiddleware.RequireAuth(brainHandler.HandlePurgeSignals),
	).Methods("POST", "OPTIONS")

	// Stats
	protected.HandleFunc("/brain/stats",
		authMiddleware.RequireAuth(brainHandler.HandleGetStats),
	).Methods("GET", "OPTIONS")

	// Feedback
	protected.HandleFunc("/brain/feedback",
		authMiddleware.RequireAuth(brainHandler.HandleSubmitFeedback),
	).Methods("POST", "OPTIONS")

	// Composers
	protected.HandleFunc("/brain/composers",
		authMiddleware.RequireAuth(brainHandler.HandleListComposers),
	).Methods("GET", "OPTIONS")

	protected.HandleFunc("/brain/composers",
		authMiddleware.RequireAuth(brainHandler.HandleCreateComposer),
	).Methods("POST", "OPTIONS")

	protected.HandleFunc("/brain/composers/{composer_id}",
		authMiddleware.RequireAuth(brainHandler.HandleDeleteComposer),
	).Methods("DELETE", "OPTIONS")

	// Triggers
	protected.HandleFunc("/brain/triggers",
		authMiddleware.RequireAuth(brainHandler.HandleListTriggers),
	).Methods("GET", "OPTIONS")

	protected.HandleFunc("/brain/triggers",
		authMiddleware.RequireAuth(brainHandler.HandleCreateTrigger),
	).Methods("POST", "OPTIONS")

	protected.HandleFunc("/brain/triggers/{trigger_id}",
		authMiddleware.RequireAuth(brainHandler.HandleDeleteTrigger),
	).Methods("DELETE", "OPTIONS")
}
