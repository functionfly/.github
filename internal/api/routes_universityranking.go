package api

import (
	universityrankinghandler "github.com/functionfly/functionfly/internal/api/handlers/universityranking"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerUniversityRankingRoutes wires the public + authed routes for the
// university leaderboard. Specific routes (`/me`, `/resolve`, `/opt-out`)
// MUST be registered before the `/{slug}` catch-all.
func registerUniversityRankingRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	csrfMiddleware *middleware.CSRFMiddleware,
	h *universityrankinghandler.Handler,
) {
	// Public list.
	api.HandleFunc("/university-rankings", h.HandleListLeaderboard).Methods("GET", "OPTIONS")

	// Auth (must precede /{slug}).
	api.HandleFunc("/university-rankings/me", authMiddleware.RequireAuth(h.HandleGetMyUniversity)).Methods("GET", "OPTIONS")
	api.HandleFunc("/university-rankings/resolve", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleResolveUniversity))).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/university-ranking-opt-out", authMiddleware.RequireAuth(h.HandleGetOptOut)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/university-ranking-opt-out", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleSetOptOut))).Methods("POST", "OPTIONS")

	// Public detail (catch-all last so the above specifics win).
	api.HandleFunc("/university-rankings/{slug}", h.HandleGetUniversity).Methods("GET", "OPTIONS")
}
