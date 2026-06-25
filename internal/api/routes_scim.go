package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/functionfly/functionfly/internal/api/handlers/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
)

func registerSCIMRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware, scimHandler *auth.SCIMHandler, repo storage.Repository) {
	if os.Getenv("GBA_SCIM_ENABLED") != "true" {
		return
	}

	scim := api.PathPrefix("/scim").Subrouter()

	scim.HandleFunc("/Users", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleListUsers))).Methods("GET", "OPTIONS")
	scim.HandleFunc("/Users", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleCreateUser))).Methods("POST", "OPTIONS")
	scim.HandleFunc("/Users/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleGetUser))).Methods("GET", "OPTIONS")
	scim.HandleFunc("/Users/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleUpdateUser))).Methods("PUT", "OPTIONS")
	scim.HandleFunc("/Users/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandlePatchUser))).Methods("PATCH", "OPTIONS")
	scim.HandleFunc("/Users/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleDeleteUser))).Methods("DELETE", "OPTIONS")

	scim.HandleFunc("/Groups", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleListGroups))).Methods("GET", "OPTIONS")
	scim.HandleFunc("/Groups", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleCreateGroup))).Methods("POST", "OPTIONS")
	scim.HandleFunc("/Groups/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleGetGroup))).Methods("GET", "OPTIONS")
	scim.HandleFunc("/Groups/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleUpdateGroup))).Methods("PUT", "OPTIONS")
	scim.HandleFunc("/Groups/{id}", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleDeleteGroup))).Methods("DELETE", "OPTIONS")

	scim.HandleFunc("/Config", authMiddleware.RequireAuth(requireSCIMFeature(repo, scimHandler.HandleGetConfig))).Methods("GET", "OPTIONS")
}

func requireSCIMFeature(repo storage.Repository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUserFromContext(r)
		if user == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		tenant, err := repo.GetTenantByID(r.Context(), user.TenantID)
		if err != nil {
			http.Error(w, `{"error":"internal_error","message":"Failed to verify tenant"}`, http.StatusInternalServerError)
			return
		}

		if tenant == nil || !plans.HasFeature(tenant.Plan, plans.FeatureSCIM) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			resp := map[string]interface{}{
				"error":   "feature_not_available",
				"message": "SCIM Provisioning requires an Enterprise plan",
				"feature": plans.FeatureSCIM,
			}
			if tenant != nil {
				resp["current_plan"] = tenant.Plan
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		next(w, r)
	}
}
