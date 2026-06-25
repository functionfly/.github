package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/cityranking"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerCityRankingRoutes wires the public + auth-protected city-ranking
// endpoints. The cron job that produces the data lives in
// internal/jobs/cityranking and is started in routes.go alongside other
// schedulers. Specific routes (`/me`, `/movers`, `/resolve`, `/states`, `/map`)
// MUST be registered before the `/{slug}` catch-all.
func registerCityRankingRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	csrfMiddleware *middleware.CSRFMiddleware,
	h *cityranking.Handler,
) {
	// Specific public routes first.
	api.HandleFunc("/city-rankings", h.HandleListLeaderboard).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/categories", h.HandleListCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/movers", h.HandleListMovers).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/states", h.HandleListStates).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/states/{code}", h.HandleGetState).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/map", h.HandleListMapPoints).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/cities", h.HandleListCities).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/countries", h.HandleListCountries).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/ambassadors", h.HandleListAmbassadors).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/resolve-by-ip", h.HandleResolveByIP).Methods("GET", "OPTIONS")

	// Auth required (must precede /{slug}).
	api.HandleFunc("/city-rankings/me", authMiddleware.RequireAuth(h.HandleGetMyCity)).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/me", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleSetMyCity))).Methods("POST", "OPTIONS")
	api.HandleFunc("/city-rankings/resolve", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleResolveCity))).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/city-ranking-opt-out", authMiddleware.RequireAuth(h.HandleGetOptOut)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/city-ranking-opt-out", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleSetOptOut))).Methods("POST", "OPTIONS")

	// Public catch-all last. The /builders and /ambassador and /universities sub-paths are
	// registered before the {slug} catch-all so they can use different
	// handlers.
	api.HandleFunc("/city-rankings/{slug}/builders", h.HandleListBuilders).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/{slug}/ambassador", h.HandleGetAmbassador).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/{slug}/universities", h.HandleListUniversities).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-rankings/{slug}", h.HandleGetMetro).Methods("GET", "OPTIONS")
}
