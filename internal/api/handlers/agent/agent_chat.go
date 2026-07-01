package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/aikeys"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type agentChatRequest struct {
	Message string `json:"message"`
}

type agentChatResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Model   string `json:"model"`
}

// agentSessionID returns a deterministic UUID for an agent's chat session.
func agentSessionID(agentID uuid.UUID) uuid.UUID {
	h := sha256.Sum256([]byte("agent-chat-session:" + agentID.String()))
	s := hex.EncodeToString(h[:16])
	u, _ := uuid.Parse(s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:])
	return u
}

// HandleAgentChat sends a message to the agent via the FlyMind AI service
// using the agent's configured model. Persists messages to chat history.
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

	sessionID := agentSessionID(agent.ID)

	// Execute tools if the message looks like it needs them
	toolResults := h.executeAgentTools(r.Context(), req.Message, agent, claims)

	messageWithContext := req.Message
	if toolResults != "" {
		messageWithContext = fmt.Sprintf("%s\n\n[Tool Results]\n%s", req.Message, toolResults)
		systemPrompt += "\n\nYou have access to tools. When tool results are provided in [Tool Results], use them to answer the user's question. Cite sources when available."
	}

	aiReq := map[string]interface{}{
		"session_id": sessionID.String(),
		"message":    messageWithContext,
		"model":      model,
		"tenant_id":  claims.TenantID.String(),
		"user_id":    claims.UserID.String(),
		"context": map[string]string{
			"agent_id":      agent.AgentID,
			"agent_name":    agent.Name,
			"system_prompt": systemPrompt,
		},
	}

	// Persist user message
	modelCopy := model
	h.saveChatMessage(r.Context(), sessionID, "user", req.Message, &modelCopy)

	body, _ := json.Marshal(aiReq)
	httpReq, err := http.NewRequestWithContext(r.Context(), "POST", aiBaseURL+"/api/chat/message", bytes.NewBuffer(body))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "failed to create AI request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		httpReq.Header.Set("X-API-Key", aiAPIKey)
	}

	// Inject BYOK key if tenant has one for this model's provider
	if h.byokRepo != nil {
		provider := providerFromModel(model)
		if provider != "" {
			tid, err := uuid.Parse(claims.TenantID.String())
			if err == nil {
				key, err := h.byokRepo.GetByTenantAndProvider(r.Context(), tid, provider)
				if err != nil || key == nil || key.Status != "active" {
					if provider == "mimo" {
						key, err = h.byokRepo.GetByTenantAndProvider(r.Context(), tid, "mimo-token-plan")
						if err != nil || key == nil || key.Status != "active" {
							key = nil
						}
					} else if provider == "minimax" {
						key, err = h.byokRepo.GetByTenantAndProvider(r.Context(), tid, "minimax-token-plan")
						if err != nil || key == nil || key.Status != "active" {
							key = nil
						}
					} else {
						key = nil
					}
				}
				if key != nil {
					plaintext, err := aikeys.DecryptKey(key.EncryptedKey, key.KeyNonce, tid)
					if err == nil {
						httpReq.Header.Set("X-BYOK-Key", string(plaintext))
						httpReq.Header.Set("X-BYOK-Provider", provider)
						httpReq.Header.Set("X-Key-Source", "byok")
						if key.Provider == "mimo-token-plan" {
							httpReq.Header.Set("X-BYOK-Provider", "mimo")
							httpReq.Header.Set("X-Key-Source", "token-plan")
							region := extractRegionFromHealthMessage(key.HealthMessage)
							if base, ok := aikeys.TokenPlanRegionURLs[region]; ok {
								httpReq.Header.Set("X-BYOK-Base-URL", base)
							}
						} else if key.Provider == "minimax-token-plan" {
							httpReq.Header.Set("X-BYOK-Provider", "minimax")
							httpReq.Header.Set("X-Key-Source", "token-plan")
						}
					}
				}
			}
		}
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
	reply := string(respBody)
	if err := json.Unmarshal(respBody, &aiResp); err == nil && aiResp.Message != "" {
		reply = aiResp.Message
	}

	// Persist assistant reply
	h.saveChatMessage(r.Context(), sessionID, "assistant", reply, &modelCopy)

	writeJSON(w, http.StatusOK, agentChatResponse{
		OK:      true,
		Message: reply,
		Model:   model,
	})
}

// HandleAgentChatClear deletes all chat messages for an agent session.
// DELETE /v1/agent/{agent_id}/chat
func (h *Handler) HandleAgentChatClear(w http.ResponseWriter, r *http.Request) {
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

	if h.rawDB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}

	sessionID := agentSessionID(agent.ID)
	if _, err := h.rawDB.ExecContext(r.Context(),
		`DELETE FROM ai_chat_messages WHERE session_id = $1`, sessionID,
	); err != nil {
		logrus.WithError(err).WithField("session_id", sessionID).Error("HandleAgentChatClear: failed to delete messages")
		writeError(w, http.StatusInternalServerError, "CLEAR_FAILED", "failed to clear chat")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// HandleAgentChatHistory returns chat history for an agent.
// GET /v1/agent/{agent_id}/chat/history?limit=50&offset=0
func (h *Handler) HandleAgentChatHistory(w http.ResponseWriter, r *http.Request) {
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

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	sessionID := agentSessionID(agent.ID)

	db := h.rawDB
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "messages": []map[string]interface{}{}})
		return
	}

	rows, err := db.QueryContext(r.Context(),
		`SELECT role, content, COALESCE(model, ''), created_at FROM ai_chat_messages WHERE session_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		sessionID, limit, offset,
	)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "messages": []interface{}{}})
		return
	}
	defer rows.Close()

	type chatMsg struct {
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		Model     string    `json:"model,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	var messages []chatMsg
	for rows.Next() {
		var m chatMsg
		if err := rows.Scan(&m.Role, &m.Content, &m.Model, &m.CreatedAt); err == nil {
			messages = append(messages, m)
		}
	}

	if messages == nil {
		messages = []chatMsg{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"messages": messages,
	})
}

// saveChatMessage persists a chat message to the database.
func (h *Handler) saveChatMessage(ctx context.Context, sessionID uuid.UUID, role, content string, model *string) {
	if h.rawDB == nil {
		logrus.Warn("saveChatMessage: rawDB is nil, cannot persist chat message")
		return
	}

	// Ensure session exists
	if _, err := h.rawDB.ExecContext(ctx,
		`INSERT INTO ai_chat_sessions (id, user_id, tenant_id, title, context_type, context_reference, is_active, created_at, updated_at)
		 VALUES ($1, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 'Agent Chat', 'agent', NULL, true, NOW(), NOW())
		 ON CONFLICT (id) DO NOTHING`,
		sessionID,
	); err != nil {
		logrus.WithError(err).WithField("session_id", sessionID).Error("saveChatMessage: failed to upsert session")
		return
	}

	if _, err := h.rawDB.ExecContext(ctx,
		`INSERT INTO ai_chat_messages (id, session_id, role, content, tokens_used, model, metadata, created_at)
		 VALUES ($1, $2, $3, $4, 0, $5, '{}', NOW())`,
		uuid.New(), sessionID, role, content, model,
	); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"role":       role,
		}).Error("saveChatMessage: failed to insert message")
	}
}

// executeAgentTools detects tool-worthy intents and executes them, returning results as text.
func (h *Handler) executeAgentTools(ctx context.Context, message string, agent *identity.AgentIdentity, claims *auth.Claims) string {
	msg := strings.ToLower(message)

	// Detect search intent
	searchTriggers := []string{"search", "find", "look up", "what is", "what are", "best", "compare", "who is", "where", "how to", "latest", "news", "price", "buy"}
	needsSearch := false
	for _, trigger := range searchTriggers {
		if strings.Contains(msg, trigger) {
			needsSearch = true
			break
		}
	}

	if !needsSearch {
		return ""
	}

	// Call the AI service's tool execution endpoint
	aiBaseURL := os.Getenv("AI_SERVICE_URL")
	if aiBaseURL == "" {
		aiBaseURL = "http://localhost:18081"
	}
	aiAPIKey := os.Getenv("AI_SERVICE_API_KEY")

	toolReq := map[string]interface{}{
		"tool": "web_search",
		"params": map[string]interface{}{
			"query":        message,
			"max_results":  5,
		},
		"agent_id":  agent.AgentID,
		"tenant_id": claims.TenantID.String(),
	}

	body, _ := json.Marshal(toolReq)
	req, err := http.NewRequestWithContext(ctx, "POST", aiBaseURL+"/api/tools/execute", bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	if aiAPIKey != "" {
		req.Header.Set("X-API-Key", aiAPIKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback: execute search directly
		return h.executeSearchDirect(ctx, message)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK   bool        `json:"ok"`
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.OK {
		return h.executeSearchDirect(ctx, message)
	}

	data, _ := json.MarshalIndent(result.Data, "", "  ")
	return string(data)
}

// executeSearchDirect runs a web search directly without the AI service.
func (h *Handler) executeSearchDirect(ctx context.Context, query string) string {
	searchURL := os.Getenv("AGENT_WEB_SEARCH_URL")
	if searchURL == "" {
		return ""
	}

	reqBody := fmt.Sprintf(`{"query":%q,"max_results":5}`, query)
	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, strings.NewReader(reqBody))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	data, _ := json.MarshalIndent(parsed, "", "  ")
	return string(data)
}

// providerFromModel maps a model name to its primary BYOK provider.
func providerFromModel(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4"):
		return "openai"
	case strings.HasPrefix(m, "claude-"):
		return "anthropic"
	case strings.HasPrefix(m, "groq/") || strings.Contains(m, "llama") || strings.Contains(m, "mixtral"):
		return "groq"
	case strings.HasPrefix(m, "minimax"):
		return "minimax"
	case strings.HasPrefix(m, "mimo") || strings.HasPrefix(m, "xiaomi"):
		return "mimo"
	case strings.HasPrefix(m, "step-") || strings.HasPrefix(m, "stepfun"):
		return "stepfun"
	default:
		return ""
	}
}

func extractRegionFromHealthMessage(healthMessage string) string {
	const prefix = "region:"
	if len(healthMessage) > len(prefix) && healthMessage[:len(prefix)] == prefix {
		return healthMessage[len(prefix):]
	}
	return ""
}
