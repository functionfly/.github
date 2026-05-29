package studio

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

func (h *SettingsHandler) HandleGetAccountSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	prefs, err := h.repo.GetAccountPreferences(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to get preferences")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get account settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"account_preferences": prefs,
	})
}

func (h *SettingsHandler) HandleSaveAccountSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		AccountPreferences storage.AccountPreferences `json:"account_preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	saved, err := h.repo.SaveAccountPreferences(r.Context(), tenantID, userID, environment, req.AccountPreferences)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to save preferences")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save account settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"account_preferences": saved,
	})
}

func (h *SettingsHandler) HandlePatchAccountSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var patch storage.AccountPreferencesPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	saved, err := h.repo.PatchAccountPreferences(r.Context(), tenantID, userID, environment, patch)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to patch preferences")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update account settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"account_preferences": saved,
	})
}

func (h *SettingsHandler) HandleResetAccountSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	saved, err := h.repo.ResetAccountPreferences(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to reset preferences")
		writeJSONError(w, http.StatusInternalServerError, "Failed to reset account settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"account_preferences": saved,
	})
}

func (h *SettingsHandler) HandleGetLastWorkspace(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	workspace, err := h.repo.GetLastWorkspace(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to get last workspace")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get last workspace")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"last_workspace": workspace,
	})
}

func (h *SettingsHandler) HandleSaveLastWorkspace(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		LastWorkspace storage.LastWorkspaceState `json:"last_workspace"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	route := strings.TrimSpace(req.LastWorkspace.Route)
	if route == "" || !strings.HasPrefix(route, "/") {
		writeJSONError(w, http.StatusBadRequest, "last_workspace.route must be a relative path")
		return
	}
	if !strings.HasPrefix(route, "/studio") {
		writeJSONError(w, http.StatusBadRequest, "last_workspace.route must be under /studio")
		return
	}

	req.LastWorkspace.Route = route
	if req.LastWorkspace.UpdatedAt.IsZero() {
		req.LastWorkspace.UpdatedAt = time.Now().UTC()
	}

	saved, err := h.repo.SaveLastWorkspace(r.Context(), tenantID, userID, environment, req.LastWorkspace)
	if err != nil {
		logrus.WithError(err).Error("studio account settings: failed to save last workspace")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save last workspace")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"last_workspace": saved,
	})
}
