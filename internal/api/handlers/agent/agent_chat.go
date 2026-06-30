package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

type agentChatRequest struct {
	Message string `json:"message"`
}

type agentChatResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Model   string `json:"model"`
}

// HandleAgentChat sends a message to the agent via the FlyMind AI service
// using the agent's configured model.
// POST /v1/agent/{agent_id}/chat
func (h *Handler) HandleAgentChat(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req agentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "message is required")
		return
	}

	model := agent.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	aiBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiBaseURL == "" {
		aiBaseURL = "http://localhost:18081"
	}
	aiAPIKey := os.Getenv("AI_SERVICE_API_KEY")

	systemPrompt := fmt.Sprintf(
		"You are %s, an AI agent. %s Respond concisely and helpfully.",
		agent.Name, agent.Description,
	)

	aiReq := map[string]interface{}{
		"session_id": fmt.Sprintf("agent-console-%s-%d", agent.ID.String(), time.Now().UnixMilli()),
		"message":    req.Message,
		"model":      model,
		"tenant_id":  claims.TenantID.String(),
		"user_id":    claims.UserID.String(),
		"context": map[string]string{
			"agent_id":      agent.AgentID,
			"agent_name":    agent.Name,
			"system_prompt": systemPrompt,
		},
	}

	body, _ := json.Marshal(aiReq)
	httpReq, err := http.NewRequestWithContext(r.Context(), "POST", aiBaseURL+"/api/chat/message", bytes.NewBuffer(body))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "failed to create AI request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+aiAPIKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_UNREACHABLE", "AI service is unreachable")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		writeJSON(w, resp.StatusCode, map[string]interface{}{
			"ok":      false,
			"message": fmt.Sprintf("AI service returned status %d", resp.StatusCode),
			"raw":     string(respBody),
		})
		return
	}

	var aiResp struct {
		Message string `json:"message"`
		Model   string `json:"model"`
	}
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		writeJSON(w, http.StatusOK, agentChatResponse{
			OK:      true,
			Message: string(respBody),
			Model:   model,
		})
		return
	}

	writeJSON(w, http.StatusOK, agentChatResponse{
		OK:      true,
		Message: aiResp.Message,
		Model:   aiResp.Model,
	})
}
