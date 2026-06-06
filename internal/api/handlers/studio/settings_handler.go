package studio

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

type SettingsHandler struct {
	repo *storage.StudioSettingsRepository
}

func NewSettingsHandler(repo *storage.StudioSettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

func (h *SettingsHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	settings, err := h.repo.GetSettings(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Warn("studio settings: failed to get settings")
		writeJSON(w, http.StatusOK, map[string]interface{}{"settings": getDefaultSettings()})
		return
	}

	if settings == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"settings": getDefaultSettings()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"settings": settings.Settings})
}

func (h *SettingsHandler) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	logrus.Infof("[StudioSettings] Save request - tenant: %s, user: %s, env: %s", tenantID, userID, environment)

	var req struct {
		Settings storage.Settings `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Warnf("[StudioSettings] Failed to decode request body: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logrus.Infof("[StudioSettings] Saving settings: %+v", req.Settings)

	saved, err := h.repo.SaveSettings(r.Context(), tenantID, userID, environment, req.Settings)
	if err != nil {
		logrus.WithError(err).Error("studio settings: failed to save settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save settings")
		return
	}

	logrus.Infof("[StudioSettings] Settings saved successfully: %+v", saved.Settings)
	writeJSON(w, http.StatusOK, map[string]interface{}{"settings": saved.Settings})
}

func (h *SettingsHandler) HandleResetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	if err := h.repo.DeleteSettings(r.Context(), tenantID, userID, environment); err != nil {
		logrus.WithError(err).Error("studio settings: failed to reset settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to reset settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"settings": getDefaultSettings()})
}

func getDefaultSettings() storage.Settings {
	return storage.Settings{
		Theme:               "dark",
		PrimaryColor:       "orange",
		FontSize:           14,
		SidebarPosition:    "left",
		CompactMode:        false,
		AnimationsEnabled:  true,
		Transparency:       false,
		NotificationLevel:  "all",
		SoundEnabled:       true,
		AutoSave:           true,
		AutoSaveInterval:   30,
		EditorFeatures: storage.EditorFeatures{
			BracketMatching: true,
			Minimap:         true,
			LineNumbers:     true,
			WordWrap:        false,
		},
		// Privacy
		UsageAnalyticsEnabled: false,
		CrashReportsEnabled:   true,
		// Shortcuts
		ShowShortcutHints: true,
		// Performance
		GPUAccelerationEnabled: true,
		DeveloperToolsEnabled:  false,
		MemoryLimitMB:          0,
		// Network
		ProxyEnabled: false,
		ProxyURL:     "",
		ProxyBypass:  "",
	}
}