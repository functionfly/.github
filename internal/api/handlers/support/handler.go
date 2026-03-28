package support

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/support"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles support API requests
type Handler struct {
	service *support.Service
	logger  *logrus.Logger
}

// NewHandler creates a new support handler
func NewHandler(svc *support.Service, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{service: svc, logger: logger}
}

// isStaff checks if the user has a staff role (super_admin, admin, or support)
func (h *Handler) isStaff(claims *auth.Claims) bool {
	if claims == nil {
		return false
	}
	// Check for staff roles
	switch claims.Role {
	case "super_admin", "admin", "support":
		return true
	}
	return false
}

// CreateConversationRequest is the request body for creating a support conversation
type CreateConversationRequest struct {
	Type             string            `json:"type"`
	Priority         string            `json:"priority"`
	Title            string            `json:"title"`
	FunctionAuthor   string            `json:"function_author,omitempty"`
	FunctionName     string            `json:"function_name,omitempty"`
	FunctionVersion  string            `json:"function_version,omitempty"`
	DeploymentID     string            `json:"deployment_id,omitempty"`
	DeploymentLogs   string            `json:"deployment_logs,omitempty"`
	DeploymentError  string            `json:"deployment_error,omitempty"`
	IsEmergency      bool              `json:"is_emergency"`
}

// CreateConversation handles POST /v1/support/conversations
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	conversationType := support.ConversationType(req.Type)
	if conversationType == "" {
		conversationType = support.TypeSupportAI
	}

	priority := support.Priority(req.Priority)
	if priority == "" {
		priority = support.PriorityNormal
	}

	createReq := &support.CreateConversationRequest{
		Type:     conversationType,
		Priority: priority,
		Title:    req.Title,
	}

	if req.FunctionAuthor != "" && req.FunctionName != "" {
		createReq.FunctionRef = &support.FunctionRef{
			Author:  req.FunctionAuthor,
			Name:    req.FunctionName,
			Version: req.FunctionVersion,
		}
	}

	if req.DeploymentID != "" {
		id, err := uuid.Parse(req.DeploymentID)
		if err == nil {
			createReq.DeploymentID = &id
		}
	}

	createReq.DeploymentLogs = req.DeploymentLogs
	createReq.DeploymentError = req.DeploymentError
	createReq.IsEmergency = req.IsEmergency

	conversation, err := h.service.CreateConversation(r.Context(), user.UserID, createReq)
	if err != nil {
		h.logger.WithError(err).Error("Create conversation failed")
		msg := "Failed to create conversation"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = fmt.Sprintf("Failed to create conversation: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conversation)
}

// ListConversations handles GET /v1/support/conversations
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	list, err := h.service.ListConversations(r.Context(), user.UserID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("List conversations failed")
		http.Error(w, `{"error":"Failed to list conversations"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"conversations": list})
}

// ListActiveConversations handles GET /v1/support/conversations/active (staff only)
func (h *Handler) ListActiveConversations(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// SECURITY: Require staff role
	if !h.isStaff(user) {
		http.Error(w, `{"error":"Staff access required"}`, http.StatusForbidden)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	list, err := h.service.ListActiveConversations(r.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("List active conversations failed")
		http.Error(w, `{"error":"Failed to list active conversations"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"conversations": list})
}

// GetConversation handles GET /v1/support/conversations/{id}
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}

	conversation, err := h.service.GetConversation(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Get conversation failed")
		http.Error(w, `{"error":"Failed to get conversation"}`, http.StatusInternalServerError)
		return
	}
	if conversation == nil {
		http.Error(w, `{"error":"Conversation not found"}`, http.StatusNotFound)
		return
	}

	// Verify user has access
	if conversation.UserID != user.UserID && !h.isStaff(user) {
		http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// GetMessages handles GET /v1/support/conversations/{id}/messages
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.service.GetMessages(r.Context(), id, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Get messages failed")
		http.Error(w, `{"error":"Failed to get messages"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"messages": messages})
}

// SendMessageRequest is the request body for sending a message
type SendMessageRequest struct {
	Content string `json:"content"`
}

// SendMessage handles POST /v1/support/conversations/{id}/messages
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
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

	message, err := h.service.SendMessage(r.Context(), id, user.UserID, support.AuthorUser, req.Content)
	if err != nil {
		h.logger.WithError(err).Error("Send message failed")
		http.Error(w, `{"error":"Failed to send message"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// EscalateConversation handles POST /v1/support/conversations/{id}/escalate
func (h *Handler) EscalateConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.EscalateToHuman(r.Context(), id); err != nil {
		h.logger.WithError(err).Error("Escalate conversation failed")
		http.Error(w, `{"error":"Failed to escalate conversation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "escalated"})
}

// ResolveConversationRequest is the request body for resolving a conversation
type ResolveConversationRequest struct {
	Note string `json:"note"`
}

// ResolveConversation handles POST /v1/support/conversations/{id}/resolve
func (h *Handler) ResolveConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}

	var req ResolveConversationRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.ResolveConversation(r.Context(), id, user.UserID, req.Note); err != nil {
		h.logger.WithError(err).Error("Resolve conversation failed")
		http.Error(w, `{"error":"Failed to resolve conversation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

// EmergencyFixRequest is the request body for emergency fix
type EmergencyFixRequest struct {
	FunctionID string `json:"function_id"`
	Reason    string `json:"reason"`
}

// CreateEmergencyFix handles POST /v1/support/conversations/{id}/emergency
func (h *Handler) CreateEmergencyFix(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}

	var req EmergencyFixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		http.Error(w, `{"error":"Invalid function ID"}`, http.StatusBadRequest)
		return
	}

	emergency, err := h.service.CreateEmergencyFixRequest(r.Context(), &support.EmergencyFixRequestInput{
		ConversationID: conversationID,
		UserID:         user.UserID,
		FunctionID:     functionID,
		Reason:         req.Reason,
	})
	if err != nil {
		h.logger.WithError(err).Error("Create emergency fix request failed")
		http.Error(w, `{"error":"Failed to create emergency request"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(emergency)
}

// ListEmergencies handles GET /v1/support/emergencies (staff only)
func (h *Handler) ListEmergencies(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// SECURITY: Require staff role
	if !h.isStaff(user) {
		http.Error(w, `{"error":"Staff access required"}`, http.StatusForbidden)
		return
	}

	emergencies, err := h.service.ListPendingEmergencies(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("List emergencies failed")
		http.Error(w, `{"error":"Failed to list emergencies"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"emergencies": emergencies})
}

// AcceptEmergency handles POST /v1/support/emergencies/{id}/accept
func (h *Handler) AcceptEmergency(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	emergencyID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid emergency ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.AcceptEmergency(r.Context(), emergencyID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Accept emergency failed")
		http.Error(w, `{"error":"Failed to accept emergency"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

// StaffStatusRequest is the request body for updating staff status
type StaffStatusRequest struct {
	Online bool `json:"online"`
}

// SetStaffStatus handles POST /v1/support/staff/status
func (h *Handler) SetStaffStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req StaffStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.SetStaffOnline(r.Context(), user.UserID, req.Online); err != nil {
		h.logger.WithError(err).Error("Set staff status failed")
		http.Error(w, `{"error":"Failed to set staff status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"online": req.Online})
}

// GetStaffStatus handles GET /v1/support/staff/status
func (h *Handler) GetStaffStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	status, err := h.service.GetStaffAvailability(r.Context(), user.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Get staff status failed")
		http.Error(w, `{"error":"Failed to get staff status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if status == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"online": false})
		return
	}
	json.NewEncoder(w).Encode(status)
}

// ListOnlineStaff handles GET /v1/support/staff/online
func (h *Handler) ListOnlineStaff(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	_ = user

	staff, err := h.service.ListOnlineStaff(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("List online staff failed")
		http.Error(w, `{"error":"Failed to list online staff"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"staff": staff})
}

// RegisterRoutes registers support routes (protected is under /v1, so paths omit /v1)
func (h *Handler) RegisterRoutes(router *mux.Router, auth *middleware.AuthMiddleware) {
	wrap := auth.RequireAuth
	// Conversation routes
	router.HandleFunc("/support/conversations", wrap(h.CreateConversation)).Methods("POST")
	router.HandleFunc("/support/conversations", wrap(h.ListConversations)).Methods("GET")
	router.HandleFunc("/support/conversations/active", wrap(h.ListActiveConversations)).Methods("GET")
	router.HandleFunc("/support/conversations/{id}", wrap(h.GetConversation)).Methods("GET")
	router.HandleFunc("/support/conversations/{id}/messages", wrap(h.GetMessages)).Methods("GET")
	router.HandleFunc("/support/conversations/{id}/messages", wrap(h.SendMessage)).Methods("POST")
	router.HandleFunc("/support/conversations/{id}/escalate", wrap(h.EscalateConversation)).Methods("POST")
	router.HandleFunc("/support/conversations/{id}/resolve", wrap(h.ResolveConversation)).Methods("POST")
	router.HandleFunc("/support/conversations/{id}/emergency", wrap(h.CreateEmergencyFix)).Methods("POST")

	// Emergency routes
	router.HandleFunc("/support/emergencies", wrap(h.ListEmergencies)).Methods("GET")
	router.HandleFunc("/support/emergencies/{id}/accept", wrap(h.AcceptEmergency)).Methods("POST")

	// Staff routes
	router.HandleFunc("/support/staff/status", wrap(h.GetStaffStatus)).Methods("GET")
	router.HandleFunc("/support/staff/status", wrap(h.SetStaffStatus)).Methods("POST")
	router.HandleFunc("/support/staff/online", wrap(h.ListOnlineStaff)).Methods("GET")
}
