package api

import (
	citywarhandler "github.com/functionfly/functionfly/internal/api/handlers/citywar"
	"github.com/gorilla/mux"
)

// registerCityWarRoutes wires the public city-wars endpoints.
func registerCityWarRoutes(
	api *mux.Router,
	h *citywarhandler.Handler,
) {
	api.HandleFunc("/city-wars", h.HandleListWars).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-wars/latest", h.HandleGetLatestWar).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-wars/champions", h.HandleGetWarChampions).Methods("GET", "OPTIONS")
	api.HandleFunc("/city-wars/{slug}", h.HandleGetWar).Methods("GET", "OPTIONS")
}
