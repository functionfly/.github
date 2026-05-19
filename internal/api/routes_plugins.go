package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/plugin"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerPluginRoutes(
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	pluginHandler *plugin.Handler,
) {
	protected.HandleFunc("/plugins", authMiddleware.RequireAuth(pluginHandler.HandleListPlugins)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/plugins", authMiddleware.RequireAuth(pluginHandler.HandleInstallPlugin)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}", authMiddleware.RequireAuth(pluginHandler.HandleGetPlugin)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/plugins/{id}", authMiddleware.RequireAuth(pluginHandler.HandleUpdatePlugin)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/plugins/{id}", authMiddleware.RequireAuth(pluginHandler.HandleUninstallPlugin)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/enable", authMiddleware.RequireAuth(pluginHandler.HandleEnablePlugin)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/disable", authMiddleware.RequireAuth(pluginHandler.HandleDisablePlugin)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/pause", authMiddleware.RequireAuth(pluginHandler.HandlePausePlugin)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/rollback", authMiddleware.RequireAuth(pluginHandler.HandleRollbackPlugin)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/config", authMiddleware.RequireAuth(pluginHandler.HandleConfigurePlugin)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/sandbox", authMiddleware.RequireAuth(pluginHandler.HandleGetSandbox)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/sandbox", authMiddleware.RequireAuth(pluginHandler.HandleUpdateSandbox)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/permissions", authMiddleware.RequireAuth(pluginHandler.HandleGetPermissions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/permissions", authMiddleware.RequireAuth(pluginHandler.HandleSetPermission)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/versions", authMiddleware.RequireAuth(pluginHandler.HandleListVersions)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/plugins/{id}/error", authMiddleware.RequireAuth(pluginHandler.HandleSetError)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/plugins/check-rate-limit", authMiddleware.RequireAuth(pluginHandler.HandleCheckRateLimit)).Methods("POST", "OPTIONS")
}