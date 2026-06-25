package employee

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/handlers/chat"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (h *Handler) HandleListChatSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListAIChatSessionsOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if ct := q.Get("context_type"); ct != "" {
		opts.ContextType = &ct
	}

	sessions, total, err := h.repo.ListChatSessions(r.Context(), claims.UserID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list chat sessions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list chat sessions"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}

func (h *Handler) HandleGetChatSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid session ID"))
		return
	}

	session, err := h.repo.GetChatSessionByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get chat session")
		apierror.WriteError(w, apierror.NewInternal("Failed to get chat session"))
		return
	}
	if session == nil {
		apierror.WriteError(w, apierror.NewNotFound("Chat session not found"))
		return
	}
	if session.UserID != claims.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Not your chat session"))
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	messages, err := h.repo.ListChatMessages(r.Context(), id, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list chat messages")
		apierror.WriteError(w, apierror.NewInternal("Failed to list chat messages"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session":  session,
		"messages": messages,
	})
}

type createChatSessionRequest struct {
	Title           string  `json:"title"`
	ContextType     string  `json:"context_type"`
	ContextReference *string `json:"context_reference,omitempty"`
}

func (h *Handler) HandleCreateChatSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createChatSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Title == "" {
		req.Title = "New Chat"
	}
	if req.ContextType == "" {
		req.ContextType = "general"
	}

	sess := &storage.AIChatSession{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		Title:       req.Title,
		ContextType: req.ContextType,
	}
	if req.ContextReference != nil {
		ref, err := uuid.Parse(*req.ContextReference)
		if err == nil {
			sess.ContextReference = &ref
		}
	}

	created, err := h.repo.CreateChatSession(r.Context(), sess)
	if err != nil {
		h.log.WithError(err).Error("Failed to create chat session")
		apierror.WriteError(w, apierror.NewInternal("Failed to create chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": created,
	})
}

type sendMessageRequest struct {
	Content string `json:"content"`
	Model   string `json:"model,omitempty"`
}

func (h *Handler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	sessionID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid session ID"))
		return
	}

	session, err := h.repo.GetChatSessionByID(r.Context(), sessionID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get chat session")
		apierror.WriteError(w, apierror.NewInternal("Failed to get chat session"))
		return
	}
	if session == nil {
		apierror.WriteError(w, apierror.NewNotFound("Chat session not found"))
		return
	}
	if session.UserID != claims.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Not your chat session"))
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Content == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Message content is required"))
		return
	}

	userMsg := &storage.AIChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
	}
	if _, err := h.repo.CreateChatMessage(r.Context(), userMsg); err != nil {
		h.log.WithError(err).Error("Failed to save user message")
		apierror.WriteError(w, apierror.NewInternal("Failed to save message"))
		return
	}

	aiClient := h.getAIClient()

	var employeeContext map[string]string
	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err == nil && emp != nil {
		employeeContext = map[string]string{
			"role":            emp.EmploymentType,
			"clearance_level": emp.ClearanceLevel,
		}
		if emp.DepartmentID != nil {
			dept, _ := h.repo.GetDepartmentByID(r.Context(), *emp.DepartmentID)
			if dept != nil {
				employeeContext["department"] = dept.Name
			}
		}
	}

	aiReq := &chat.AIServiceRequest{
		SessionID: sessionID.String(),
		Message:   req.Content,
		TenantID:  claims.TenantID.String(),
		UserID:    claims.UserID.String(),
		Context:   employeeContext,
	}
	if req.Model != "" {
		aiReq.Model = req.Model
	}

	aiResp, err := aiClient.ChatMessage(r.Context(), aiReq)
	if err != nil {
		h.log.WithError(err).Warn("AI service error, using fallback")
		aiResp = &chat.AIServiceResponse{
			Message: "I'm having trouble connecting to my AI service. Please try again.",
		}
	}

	assistantMsg := &storage.AIChatMessage{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    aiResp.Message,
		TokensUsed: aiResp.Usage.TotalTokens,
		Model:      &aiResp.Model,
	}
	if _, err := h.repo.CreateChatMessage(r.Context(), assistantMsg); err != nil {
		h.log.WithError(err).Error("Failed to save assistant message")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": assistantMsg,
		"intent":  aiResp.Intent,
	})
}

func (h *Handler) HandleDeleteChatSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid session ID"))
		return
	}

	session, err := h.repo.GetChatSessionByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get chat session")
		apierror.WriteError(w, apierror.NewInternal("Failed to get chat session"))
		return
	}
	if session == nil {
		apierror.WriteError(w, apierror.NewNotFound("Chat session not found"))
		return
	}
	if session.UserID != claims.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Not your chat session"))
		return
	}

	if err := h.repo.DeleteChatSession(r.Context(), id); err != nil {
		h.log.WithError(err).Error("Failed to delete chat session")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) getAIClient() *chat.AIServiceClient {
	if h.aiClient != nil {
		return h.aiClient
	}
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:18081"
	}
	return chat.NewAIServiceClient(baseURL, "", logrus.New())
}
