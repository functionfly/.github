package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/reputation"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// registerReputationRoutes wires all reputation API routes
func registerReputationRoutes(
	s *Server,
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
) {
	reputationHandler := reputation.NewHandler(s.postgresDB, logrus.New())

	// Adapter: authMiddleware.RequireAuth is func(HandlerFunc) HandlerFunc,
	// but mux.Router.Use expects func(http.Handler) http.Handler.
	requireAuth := func(next http.Handler) http.Handler {
		return authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}))
	}

	// Adapter for RequirePermission to work with mux.Router.Use
	requirePermission := func(permission string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return authMiddleware.RequirePermission(permission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}))
		}
	}

	reputationRouter := api.PathPrefix("/v1/reputation").Subrouter()

	// Public routes
	reputationRouter.HandleFunc("/leaderboard", reputationHandler.HandleGetLeaderboard).Methods("GET")
	reputationRouter.HandleFunc("/profile/{userId}", reputationHandler.HandleGetProfileByUserID).Methods("GET")
	reputationRouter.HandleFunc("/stats", reputationHandler.HandleGetStats).Methods("GET")

	// Protected routes (require authentication)
	protected := reputationRouter.PathPrefix("").Subrouter()
	protected.Use(requireAuth)

	protected.HandleFunc("/profile", reputationHandler.HandleGetProfile).Methods("GET")
	protected.HandleFunc("/events", reputationHandler.HandleGetReputationEvents).Methods("GET")
	protected.HandleFunc("/score", reputationHandler.HandleAddScore).Methods("POST")

	// Trust weights routes
	reputationRouter.HandleFunc("/trust-weights", reputationHandler.HandleGetTrustWeights).Methods("GET")
	reputationRouter.HandleFunc("/trust-weights/history", reputationHandler.HandleGetTrustWeightsHistory).Methods("GET")

	// Admin-only routes
	adminRouter := reputationRouter.PathPrefix("").Subrouter()
	adminRouter.Use(requireAuth)
	adminRouter.Use(requirePermission(auth.PermSystemRead))

	adminRouter.HandleFunc("/trust-weights", reputationHandler.HandleUpdateTrustWeights).Methods("PUT")
	adminRouter.HandleFunc("/alerts", reputationHandler.HandleGetReputationFarmingAlerts).Methods("GET")
	adminRouter.HandleFunc("/detect-farming", reputationHandler.HandleDetectReputationFarming).Methods("POST")
	adminRouter.HandleFunc("/cleanup-trust-history", reputationHandler.HandleCleanupTrustHistory).Methods("POST")

	// Admin routes that require write permission
	adminWriteRouter := reputationRouter.PathPrefix("").Subrouter()
	adminWriteRouter.Use(requireAuth)
	adminWriteRouter.Use(requirePermission(auth.PermSystemWrite))

	adminWriteRouter.HandleFunc("/alerts/{alertId}/resolve", reputationHandler.HandleResolveReputationFarmingAlert).Methods("POST")
	adminWriteRouter.HandleFunc("/alerts/{alertId}/dismiss", reputationHandler.HandleDismissReputationFarmingAlert).Methods("POST")
}
