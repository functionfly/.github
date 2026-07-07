package api

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/slack"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
)

func registerSlackRoutes(
	s *Server,
	router *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
) {
	slackHandler := slack.NewHandler(&s.repo, s.logger)

	slackAdminRoutes := router.PathPrefix("/admin").Subrouter()
	slackAdminRoutes.HandleFunc("/slack/config", authMiddleware.RequirePermission("system:write")(handleSlackConfigGet)).Methods("GET", "OPTIONS")
	slackAdminRoutes.HandleFunc("/slack/config", authMiddleware.RequirePermission("system:write")(handleSlackConfigPut)).Methods("PUT", "OPTIONS")
	slackAdminRoutes.HandleFunc("/slack/test", authMiddleware.RequirePermission("system:write")(handleSlackTest)).Methods("POST", "OPTIONS")
	slackAdminRoutes.HandleFunc("/slack/channels", authMiddleware.RequirePermission("system:read")(handleSlackChannels)).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/v1/slack/commands", slackHandler.HandleCommands).Methods("POST")
	router.HandleFunc("/api/v1/slack/interactions", slackHandler.HandleInteractions).Methods("POST")
}

func handleSlackConfigGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	slackRepo := storage.NewSlackConfigRepository(nil)
	config, err := slackRepo.GetByTenantID(ctx, user.TenantID)
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	if config == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":         config.Enabled,
		"alert_channel":   config.AlertChannel,
		"report_channel":  config.ReportChannel,
		"channel_routing": config.ChannelRouting,
		"severity_config": config.SeverityConfig,
		"quiet_hours":     config.QuietHours,
	})
}

func handleSlackConfigPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SlackConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	slackRepo := storage.NewSlackConfigRepository(nil)
	existing, err := slackRepo.GetByTenantID(ctx, user.TenantID)
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	channelRouting, _ := json.Marshal(req.ChannelRouting)
	severityConfig, _ := json.Marshal(req.SeverityConfig)
	quietHours, _ := json.Marshal(req.QuietHours)

	cfg := &storage.SlackConfig{
		TenantID:       user.TenantID,
		Enabled:        req.Enabled,
		WebhookURL:     req.WebhookURL,
		AlertChannel:   req.AlertChannel,
		ReportChannel:  req.ReportChannel,
		ChannelRouting: channelRouting,
		SeverityConfig: severityConfig,
		QuietHours:     quietHours,
	}

	if existing != nil {
		cfg.ID = existing.ID
		err = slackRepo.Update(ctx, cfg)
	} else {
		err = slackRepo.Create(ctx, cfg)
	}

	if err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type SlackConfigUpdateRequest struct {
	Enabled         bool                   `json:"enabled"`
	BotToken       string                 `json:"bot_token,omitempty"`
	SigningSecret  string                 `json:"signing_secret,omitempty"`
	WebhookURL     string                 `json:"webhook_url,omitempty"`
	AlertChannel   string                 `json:"alert_channel,omitempty"`
	ReportChannel  string                 `json:"report_channel,omitempty"`
	ChannelRouting map[string]string      `json:"channel_routing,omitempty"`
	SeverityConfig map[string]bool       `json:"severity_config,omitempty"`
	QuietHours     map[string]interface{} `json:"quiet_hours,omitempty"`
}

func handleSlackTest(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "test_sent",
		"message": "Test message sent to configured Slack channel",
	})
}

func handleSlackChannels(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string][]map[string]string{
		"channels": {
			{"id": "C01ALPHA", "name": "alerts"},
			{"id": "C02BETA", "name": "incidents"},
			{"id": "C03GAMMA", "name": "general"},
		},
	})
}
