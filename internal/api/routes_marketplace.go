package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/marketplace"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerMarketplaceRoutes(
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	marketplaceHandler *marketplace.Handler,
) {
	listHandler := authMiddleware.OptionalAuth(http.HandlerFunc(marketplaceHandler.HandleListExtensions))
	protected.HandleFunc("/marketplace/extensions", func(w http.ResponseWriter, r *http.Request) {
		listHandler.ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions", authMiddleware.RequireAuth(marketplaceHandler.HandleCreateExtension)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}", authMiddleware.RequireAuth(marketplaceHandler.HandleGetExtension)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}", authMiddleware.RequireAuth(marketplaceHandler.HandleUpdateExtension)).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}", authMiddleware.RequireAuth(marketplaceHandler.HandleDeleteExtension)).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}/install", authMiddleware.RequireAuth(marketplaceHandler.HandleInstallExtension)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}/rate", authMiddleware.RequireAuth(marketplaceHandler.HandleRateExtension)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}/my-rating", authMiddleware.RequireAuth(marketplaceHandler.HandleGetMyRating)).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/extensions/{id}/ratings", marketplaceHandler.HandleListRatings).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/check-updates", authMiddleware.RequireAuth(marketplaceHandler.HandleCheckUpdates)).Methods("POST", "OPTIONS")
	protected.HandleFunc("/marketplace/install-counts", marketplaceHandler.HandleGetInstallCounts).Methods("GET", "OPTIONS")
	categoriesHandler := authMiddleware.OptionalAuth(http.HandlerFunc(marketplaceHandler.HandleGetCategories))
	protected.HandleFunc("/marketplace/categories", func(w http.ResponseWriter, r *http.Request) {
		categoriesHandler.ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")
}
