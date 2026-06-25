package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/cityranking"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

func registerCityReviewRoutes(
	admin *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	csrfMiddleware *middleware.CSRFMiddleware,
	h *cityranking.Handler,
) {
	admin.HandleFunc("/cities", h.HandleListAllCitiesAdmin).Methods("GET", "OPTIONS")
	admin.HandleFunc("/cities/pending", h.HandleListPendingCityReviews).Methods("GET", "OPTIONS")
	admin.HandleFunc("/cities/{id}/review", h.HandleGetCityReview).Methods("GET", "OPTIONS")
	admin.HandleFunc("/cities/{id}/review", authMiddleware.RequireAuth(csrfMiddleware.RequireCSRF(h.HandleReviewCity))).Methods("POST", "OPTIONS")
}
