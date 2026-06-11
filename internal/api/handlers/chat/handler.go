package chat

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo    *Repository
	service *Service
	wsHub   *WebSocketHub
	logger  *logrus.Logger
}

func NewHandler(repo *Repository, svc *Service, wsHub *WebSocketHub, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		repo:    repo,
		service: svc,
		wsHub:   wsHub,
		logger:  logger,
	}
}

type CreateSessionRequest struct {
	Title        string   `json:"title"`
	Model        string   `json:"model"`
	ConnectorIDs []string `json:"connector_ids"`
}

type CreateSessionResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "New Chat"
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}

	session := &ChatSession{
		TenantID:      user.TenantID,
		UserID:        user.UserID,
		Title:         req.Title,
		Model:         req.Model,
		ConnectorIDs: req.ConnectorIDs,
	}

	if err := h.repo.CreateSession(r.Context(), session); err != nil {
		h.logger.WithError(err).Error("Create session failed")
		http.Error(w, `{"error":"Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateSessionResponse{
		ID:        session.ID,
		Title:     session.Title,
		Model:     session.Model,
		CreatedAt: session.CreatedAt,
	})
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions, err := h.repo.ListSessions(r.Context(), user.TenantID, user.UserID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("List sessions failed")
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sessions": sessions})
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	session, err := h.repo.GetSession(r.Context(), id, user.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Get session failed")
		http.Error(w, `{"error":"Failed to get session"}`, http.StatusInternalServerError)
		return
	}
	if session == nil || session.TenantID != user.TenantID {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	messages, _ := h.repo.ListMessages(r.Context(), id, 100, 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session":  session,
		"messages": messages,
	})
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteSession(r.Context(), id, user.TenantID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Delete session failed")
		http.Error(w, `{"error":"Failed to delete session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

type UpdateSessionRequest struct {
	Title  string `json:"title"`
	Model  string `json:"model"`
}

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}

	if len(updates) == 0 {
		http.Error(w, `{"error":"No fields to update"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateSession(r.Context(), id, user.TenantID, updates); err != nil {
		h.logger.WithError(err).Error("Update session failed")
		http.Error(w, `{"error":"Failed to update session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

type SendMessageRequest struct {
	Content      string   `json:"content"`
	SessionID    string   `json:"session_id"`
	Attachments  []string `json:"attachments"`
	Stream       bool     `json:"stream"`
}

type SendMessageResponse struct {
	Message    *ChatMessage `json:"message"`
	TokenUsage int          `json:"token_usage"`
	LatencyMS  int          `json:"latency_ms"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error":"Content is required"}`, http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	session, err := h.repo.GetSession(r.Context(), sessionID, user.TenantID)
	if err != nil || session == nil || session.TenantID != user.TenantID {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	start := time.Now()

	userMsg := &ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
	}
	if err := h.repo.CreateMessage(r.Context(), userMsg); err != nil {
		h.logger.WithError(err).Error("Create user message failed")
		http.Error(w, `{"error":"Failed to save message"}`, http.StatusInternalServerError)
		return
	}

	history, _ := h.repo.ListMessages(r.Context(), sessionID, 50, 0)
	connectors, _ := h.repo.GetConnectorsByIDs(r.Context(), user.TenantID, session.ConnectorIDs)

	aiResp, err := h.service.GenerateResponse(r.Context(), &ChatRequest{
		SessionID:  sessionID,
		UserID:     user.UserID,
		TenantID:   user.TenantID,
		Message:    req.Content,
		History:    history,
		Model:      session.Model,
		Connectors: connectors,
	})

	latencyMS := int(time.Since(start).Milliseconds())

	if err != nil {
		h.logger.WithError(err).Error("AI response failed")
		aiResp = &AIResponse{Message: "I apologize, but I encountered an error processing your request. Please try again."}
	}

	assistantMsg := &ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   aiResp.Message,
		Model:     session.Model,
		TokensUsed: aiResp.Tokens,
		LatencyMS: latencyMS,
	}
	if err := h.repo.CreateMessage(r.Context(), assistantMsg); err != nil {
		h.logger.WithError(err).Error("Create assistant message failed")
	}

	if err := h.repo.UpdateSessionTimestamp(r.Context(), sessionID); err != nil {
		h.logger.WithError(err).Warn("Update session timestamp failed")
	}

	h.repo.RecordBilling(r.Context(), user.TenantID, user.UserID, sessionID, assistantMsg.TokensUsed, session.Model)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SendMessageResponse{
		Message:    assistantMsg,
		TokenUsage: assistantMsg.TokensUsed,
		LatencyMS:  latencyMS,
	})
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	models := []map[string]string{
		{"id": "gpt-4o-mini", "name": "GPT-4o Mini", "provider": "openai"},
		{"id": "gpt-4o", "name": "GPT-4o", "provider": "openai"},
		{"id": "claude-sonnet-4-20250514", "name": "Claude Sonnet 4", "provider": "anthropic"},
		{"id": "claude-3-5-sonnet-latest", "name": "Claude 3.5 Sonnet", "provider": "anthropic"},
		{"id": "deepseek-chat", "name": "DeepSeek V3", "provider": "deepinfra"},
		{"id": "accounts/fireworks/models/llama-v3-70b-instruct", "name": "Llama 3 70B", "provider": "fireworks"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"models": models})
}

func (h *Handler) RegisterRoutes(router *mux.Router, authMw *middleware.AuthMiddleware) {
	router.HandleFunc("/chat/sessions", authMw.RequireAuth(h.CreateSession)).Methods("POST")
	router.HandleFunc("/chat/sessions", authMw.RequireAuth(h.ListSessions)).Methods("GET")
	router.HandleFunc("/chat/sessions/{id}", authMw.RequireAuth(h.GetSession)).Methods("GET")
	router.HandleFunc("/chat/sessions/{id}", authMw.RequireAuth(h.UpdateSession)).Methods("PATCH")
	router.HandleFunc("/chat/sessions/{id}", authMw.RequireAuth(h.DeleteSession)).Methods("DELETE")
	router.HandleFunc("/chat/messages", authMw.RequireAuth(h.SendMessage)).Methods("POST")
	router.HandleFunc("/chat/models", authMw.RequireAuth(h.ListModels)).Methods("GET")

	router.HandleFunc("/chat/ws", h.wsHub.HandleWebSocket).Methods("GET")
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.wsHub.HandleWebSocket(w, r)
}
