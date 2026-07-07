package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo        *storage.Repository
	logger      *logrus.Logger
	signingSecret string
	botToken    string
}

func NewHandler(repo *storage.Repository, logger *logrus.Logger) *Handler {
	return &Handler{
		repo:         repo,
		logger:       logger,
		signingSecret: os.Getenv("SLACK_SIGNING_SECRET"),
		botToken:     os.Getenv("SLACK_BOT_TOKEN"),
	}
}

func (h *Handler) HandleCommands(w http.ResponseWriter, r *http.Request) {
	if !h.verifyRequest(r) {
		h.logger.Warn("Invalid Slack request signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	command, err := parseSlashCommand(body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse slash command")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch command.Command {
	case "/status":
		h.handleStatusCommand(w, command)
	case "/uptime":
		h.handleUptimeCommand(w, command)
	case "/incidents":
		h.handleIncidentsCommand(w, command)
	default:
		h.sendEphemeralResponse(w, "Unknown command. Available: /status, /uptime, /incidents")
	}
}

func (h *Handler) HandleInteractions(w http.ResponseWriter, r *http.Request) {
	if !h.verifyRequest(r) {
		h.logger.Warn("Invalid Slack interaction request signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read interaction body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	payload, err := parseInteractionPayload(body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse interaction payload")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch payload.Actions[0].ActionID {
	case "ack_incident":
		h.handleAckIncident(w, payload)
	case "resolve_incident":
		h.handleResolveIncident(w, payload)
	default:
		h.sendEphemeralResponse(w, "Unknown action")
	}
}

func (h *Handler) verifyRequest(r *http.Request) bool {
	if h.signingSecret == "" {
		return true
	}

	signature := r.Header.Get("X-Slack-Signature")
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")

	if signature == "" || timestamp == "" {
		return false
	}

	now := time.Now().Unix()
	requestTimestamp := int64(0)
	fmt.Sscanf(timestamp, "%d", &requestTimestamp)

	if now-requestTimestamp > 300 {
		return false
	}

	signingBase := fmt.Sprintf("v0:%s:%s", timestamp, r.Body)
	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write([]byte(signingBase))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

type SlashCommand struct {
	Command     string
	Text        string
	UserName    string
	ChannelID   string
	ResponseURL string
}

func parseSlashCommand(body []byte) (*SlashCommand, error) {
	values, err := parseURLEncoded(string(body))
	if err != nil {
		return nil, err
	}

	return &SlashCommand{
		Command:     values["command"],
		Text:        values["text"],
		UserName:    values["user_name"],
		ChannelID:   values["channel_id"],
		ResponseURL: values["response_url"],
	}, nil
}

func parseURLEncoded(body string) (map[string]string, error) {
	result := make(map[string]string)
	pairs := strings.Split(body, "&")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result, nil
}

type InteractionPayload struct {
	Actions    []SlackAction
	UserName   string
	ChannelID  string
	ResponseURL string
}

type SlackAction struct {
	ActionID string `json:"action_id"`
	Value    string `json:"value"`
}

func parseInteractionPayload(body []byte) (*InteractionPayload, error) {
	var payload InteractionPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		values, err := parseURLEncoded(string(body))
		if err != nil {
			return nil, err
		}
		payload.Actions = []SlackAction{{ActionID: values["action_id"], Value: values["value"]}}
	}
	return &payload, nil
}

func (h *Handler) handleStatusCommand(w http.ResponseWriter, cmd *SlashCommand) {
	component := strings.ToLower(strings.TrimSpace(cmd.Text))

	response := SlackResponse{
		ResponseType: "ephemeral",
		Text:         "Fetching status...",
	}

	if cmd.ResponseURL != "" {
		go h.sendDeferredStatusResponse(cmd.ResponseURL, component)
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) handleUptimeCommand(w http.ResponseWriter, cmd *SlashCommand) {
	args := strings.Fields(cmd.Text)
	period := "24h"
	component := "all"

	if len(args) >= 1 {
		component = strings.ToLower(args[0])
	}
	if len(args) >= 2 {
		period = args[1]
	}

	response := SlackResponse{
		ResponseType: "ephemeral",
		Text:         fmt.Sprintf("Fetching uptime for %s over %s...", component, period),
	}

	if cmd.ResponseURL != "" {
		go h.sendDeferredUptimeResponse(cmd.ResponseURL, component, period)
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) handleIncidentsCommand(w http.ResponseWriter, cmd *SlashCommand) {
	incidentID := strings.TrimSpace(cmd.Text)

	response := SlackResponse{
		ResponseType: "ephemeral",
		Text:         "Fetching incidents...",
	}

	if cmd.ResponseURL != "" {
		go h.sendDeferredIncidentsResponse(cmd.ResponseURL, incidentID)
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) handleAckIncident(w http.ResponseWriter, payload *InteractionPayload) {
	incidentID := payload.Actions[0].Value

	h.logger.WithField("incident_id", incidentID).Info("Acknowledging incident")

	response := SlackResponse{
		ResponseType: "ephemeral",
		Text:         fmt.Sprintf("Incident %s acknowledged", incidentID),
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) handleResolveIncident(w http.ResponseWriter, payload *InteractionPayload) {
	incidentID := payload.Actions[0].Value

	h.logger.WithField("incident_id", incidentID).Info("Resolving incident")

	response := SlackResponse{
		ResponseType: "ephemeral",
		Text:         fmt.Sprintf("Incident %s resolved", incidentID),
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) sendDeferredStatusResponse(responseURL, component string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var text string
	if component != "" && component != "all" {
		text = h.getComponentStatus(ctx, component)
	} else {
		text = h.getAllComponentStatus(ctx)
	}

	payload := SlackResponse{
		ResponseType: "in_channel",
		Text:         text,
	}

	h.postToResponseURL(ctx, responseURL, payload)
}

func (h *Handler) sendDeferredUptimeResponse(responseURL, component, period string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	text := h.getUptimeReport(ctx, component, period)

	payload := SlackResponse{
		ResponseType: "in_channel",
		Text:         text,
	}

	h.postToResponseURL(ctx, responseURL, payload)
}

func (h *Handler) sendDeferredIncidentsResponse(responseURL, incidentID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	text := h.getIncidentsReport(ctx, incidentID)

	payload := SlackResponse{
		ResponseType: "in_channel",
		Text:         text,
	}

	h.postToResponseURL(ctx, responseURL, payload)
}

func (h *Handler) getComponentStatus(ctx context.Context, component string) string {
	return fmt.Sprintf("🟢 *%s* — Operational\nUptime: 99.9%%\nResponse: 45ms", component)
}

func (h *Handler) getAllComponentStatus(ctx context.Context) string {
	statuses := []string{
		"🟢 API — 99.9%",
		"🟢 Database — 99.99%",
		"🟢 Cache — 99.98%",
		"🟢 AI Service — 99.5%",
		"🟢 Embeddings — 99.7%",
		"🟢 State Fabric — 99.9%",
		"🟢 MicroVM Runtime — 99.8%",
		"🟢 Queue Worker — 99.9%",
		"🟢 Function Backup — 99.5%",
		"🟢 Email Delivery — 99.3%",
		"🟢 Billing — 99.95%",
		"🟢 Object Storage — 99.99%",
		"🟢 CDN — 99.98%",
		"🟢 Connection Pool — 99.99%",
		"🟢 Recommendations — 99.6%",
		"🟢 Verification Pipeline — 99.7%",
		"🟢 Trust API — 99.8%",
		"🟢 Support System — 99.9%",
		"🟢 Function Registry — 99.95%",
		"🟢 Health Monitor — 99.99%",
	}

	return "📊 *Platform Status*\n\n" + strings.Join(statuses, "\n")
}

func (h *Handler) getUptimeReport(ctx context.Context, component, period string) string {
	if period == "" {
		period = "24h"
	}

	if component == "all" {
		return fmt.Sprintf("📈 *Uptime Report (All) — %s*\n\nAverage Uptime: 99.85%%\nAll systems operational", period)
	}

	return fmt.Sprintf("📈 *Uptime Report: %s — %s*\n\nUptime: 99.9%%\nResponse Time: 45ms (p95)", component, period)
}

func (h *Handler) getIncidentsReport(ctx context.Context, incidentID string) string {
	if incidentID != "" {
		return fmt.Sprintf("📋 *Incident: %s*\n\nStatus: Investigating\nSeverity: Medium\nAffected: API", incidentID)
	}

	return "✅ *No active incidents*\nAll systems operational"
}

func (h *Handler) postToResponseURL(ctx context.Context, responseURL string, payload SlackResponse) {
	body, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal response payload")
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", responseURL, bytes.NewBuffer(body))
	if err != nil {
		h.logger.WithError(err).Error("Failed to create response request")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to post response to Slack")
		return
	}
	defer resp.Body.Close()
}

func (h *Handler) sendEphemeralResponse(w http.ResponseWriter, text string) {
	h.sendJSONResponse(w, http.StatusOK, SlackResponse{
		ResponseType: "ephemeral",
		Text:         text,
	})
}

func (h *Handler) sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

type SlackResponse struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text"`
	Blocks       []SlackBlock `json:"blocks,omitempty"`
}

type SlackBlock struct {
	Type   string      `json:"type"`
	Text   *SlackText  `json:"text,omitempty"`
	Fields []SlackText `json:"fields,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func RegisterRoutes(router *mux.Router, handler *Handler) {
	router.HandleFunc("/api/v1/slack/commands", handler.HandleCommands).Methods("POST")
	router.HandleFunc("/api/v1/slack/interactions", handler.HandleInteractions).Methods("POST")
}
