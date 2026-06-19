// Package paperclip provides the webhook adapter so Paperclip heartbeat
// invocations can trigger FunctionFly agent executions and report results back.
package paperclip

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

const (
	webhookSecretHeader = "X-Paperclip-Webhook-Secret"
)

// Adapter receives Paperclip heartbeat webhooks, runs FunctionFly agent execution, and reports back.
type Adapter struct {
	client                *http.Client
	log                   *logrus.Logger
	webhookSecret         string
	paperclipBaseURL      string
	paperclipAPIKey       string
	functionflyBaseURL    string
	functionflyAgentAPIKey string
}

// NewAdapter creates an adapter with config from environment.
func NewAdapter(log *logrus.Logger) *Adapter {
	if log == nil {
		log = logrus.New()
	}
	return &Adapter{
		client: &http.Client{Timeout: 60 * time.Second},
		log:    log,
		// Webhook auth: only accept if header matches
		webhookSecret:         os.Getenv("PAPERCLIP_WEBHOOK_SECRET"),
		paperclipBaseURL:      os.Getenv("PAPERCLIP_BASE_URL"),
		paperclipAPIKey:       os.Getenv("PAPERCLIP_API_KEY"),
		functionflyBaseURL:    os.Getenv("FUNCTIONFLY_BASE_URL"),
		functionflyAgentAPIKey: os.Getenv("FUNCTIONFLY_AGENT_API_KEY"),
	}
}

// HeartbeatWebhookRequest is the payload Paperclip (or our adapter client) sends.
type HeartbeatWebhookRequest struct {
	PaperclipAgentID  string          `json:"paperclip_agent_id"`
	PaperclipIssueID  string          `json:"paperclip_issue_id"`
	CompanyID         string          `json:"company_id"`
	FunctionAuthor    string          `json:"function_author"`
	FunctionName      string          `json:"function_name"`
	FunctionVersion   string          `json:"function_version,omitempty"`
	Input             json.RawMessage `json:"input,omitempty"`
}

// HeartbeatWebhookResponse is returned to the webhook caller.
type HeartbeatWebhookResponse struct {
	OK           bool   `json:"ok"`
	ExecutionID  string `json:"execution_id,omitempty"`
	DurationMs   int    `json:"duration_ms,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
	CommentPosted bool   `json:"comment_posted"`
	Error        string `json:"error,omitempty"`
}

// HandleHeartbeatWebhook handles POST /v1/integrations/paperclip/heartbeat.
// It runs the requested FunctionFly function and optionally posts the result as a Paperclip issue comment.
func (a *Adapter) HandleHeartbeatWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		w.Header().Set("Allow", "POST, OPTIONS")
		apierror.WriteError(w, apierror.NewBadRequest("method not allowed"))
		return
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if a.webhookSecret != "" && r.Header.Get(webhookSecretHeader) != a.webhookSecret {
		a.log.Warn("paperclip webhook: missing or invalid secret")
		writeJSON(w, http.StatusUnauthorized, HeartbeatWebhookResponse{OK: false, Error: "unauthorized"})
		return
	}

	var req HeartbeatWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, HeartbeatWebhookResponse{OK: false, Error: "invalid request body"})
		return
	}
	if req.FunctionAuthor == "" || req.FunctionName == "" {
		writeJSON(w, http.StatusBadRequest, HeartbeatWebhookResponse{OK: false, Error: "function_author and function_name required"})
		return
	}

	execID, durationMs, costUSD, execErr := a.runFunctionFlyExecution(r.Context(), &req)
	commentPosted := false
	if a.paperclipBaseURL != "" && a.paperclipAPIKey != "" && req.PaperclipIssueID != "" {
		commentPosted = a.postResultToPaperclip(r.Context(), req.PaperclipIssueID, req.PaperclipAgentID, execID, durationMs, costUSD, execErr)
	}

	resp := HeartbeatWebhookResponse{
		OK:            execErr == nil,
		ExecutionID:   execID,
		DurationMs:    durationMs,
		CostUSD:       costUSD,
		CommentPosted: commentPosted,
	}
	if execErr != nil {
		resp.Error = execErr.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *Adapter) runFunctionFlyExecution(ctx context.Context, req *HeartbeatWebhookRequest) (execID string, durationMs int, costUSD float64, err error) {
	base := a.functionflyBaseURL
	if base == "" {
		base = "http://localhost:8080"
	}
	path := fmt.Sprintf("%s/v1/agent/execute/%s/%s", base, req.FunctionAuthor, req.FunctionName)
	if req.FunctionVersion != "" {
		path = fmt.Sprintf("%s/v1/agent/execute/%s/%s/%s", base, req.FunctionAuthor, req.FunctionName, req.FunctionVersion)
	}

	input := req.Input
	if input == nil {
		input = []byte("{}")
	}
	body := map[string]interface{}{
		"input": input,
	}
	bodyBytes, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", 0, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.functionflyAgentAPIKey != "" {
		httpReq.Header.Set("X-Agent-API-Key", a.functionflyAgentAPIKey)
	}

	start := time.Now()
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	durationMs = int(time.Since(start).Milliseconds())
	var result struct {
		OK          bool    `json:"ok"`
		ExecutionID string  `json:"execution_id"`
		DurationMs  int     `json:"duration_ms"`
		CostUSD     float64 `json:"cost_usd"`
		Error       *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	if result.DurationMs != 0 {
		durationMs = result.DurationMs
	}
	if result.CostUSD > 0 {
		costUSD = result.CostUSD
	}

	if resp.StatusCode != http.StatusOK {
		msg := ""
		if result.Error != nil {
			msg = result.Error.Message
		}
		return result.ExecutionID, durationMs, costUSD,
			fmt.Errorf("functionfly execute: %d %s", resp.StatusCode, msg)
	}
	return result.ExecutionID, durationMs, costUSD, nil
}

func (a *Adapter) postResultToPaperclip(ctx context.Context, issueID, agentID, execID string, durationMs int, costUSD float64, execErr error) bool {
	url := fmt.Sprintf("%s/api/issues/%s/comments", a.paperclipBaseURL, issueID)
	body := fmt.Sprintf("[FunctionFly execution] execution_id=%s duration_ms=%d cost_usd=%.4f", execID, durationMs, costUSD)
	if execErr != nil {
		body = fmt.Sprintf("[FunctionFly execution failed] %s", execErr.Error())
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"body":             body,
		"author_agent_id":  agentID,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		a.log.WithError(err).Warn("paperclip post comment: new request")
		return false
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.paperclipAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.paperclipAPIKey)
	}
	resp, err := a.client.Do(httpReq)
	if err != nil {
		a.log.WithError(err).Warn("paperclip post comment: do")
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.log.WithField("status", resp.StatusCode).Warn("paperclip post comment: non-2xx")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// RegisterRoutes adds Paperclip integration routes to the given router.
func RegisterRoutes(r *mux.Router, adapter *Adapter) {
	if adapter == nil {
		adapter = NewAdapter(logrus.New())
	}
	r.HandleFunc("/integrations/paperclip/heartbeat", adapter.HandleHeartbeatWebhook).Methods("POST", "OPTIONS")
}
