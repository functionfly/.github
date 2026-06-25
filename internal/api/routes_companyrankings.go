package api

import (
	companyrankinghandler "github.com/functionfly/functionfly/internal/api/handlers/companyranking"
	"github.com/gorilla/mux"
)

func registerCompanyRankingRoutes(api *mux.Router, h *companyrankinghandler.Handler) {
	api.HandleFunc("/companies", h.HandleListLeaderboard).Methods("GET", "OPTIONS")
	api.HandleFunc("/companies/{slug}", h.HandleGetCompany).Methods("GET", "OPTIONS")
	api.HandleFunc("/company-rankings", h.HandleListLeaderboard).Methods("GET", "OPTIONS")
	api.HandleFunc("/company-rankings/{slug}", h.HandleGetCompany).Methods("GET", "OPTIONS")
}
