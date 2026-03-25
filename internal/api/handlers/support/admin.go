package support

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/support"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminHandler handles admin operations for support system
type AdminHandler struct {
	repo   *support.PostgresRepository
	logger *logrus.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo *support.PostgresRepository, logger *logrus.Logger) *AdminHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &AdminHandler{
		repo:   repo,
		logger: logger,
	}
}

// SupportMetrics represents support system metrics
type SupportMetrics struct {
	TotalConversations   int64 `json:"total_conversations"`
	ActiveConversations  int64 `json:"active_conversations"`
	PendingConversations int64 `json:"pending_conversations"`
	ResolvedConversations int64 `json:"resolved_conversations"`
	EscalatedConversations int64 `json:"escalated_conversations"`
	EmergencyRequests    int64 `json:"emergency_requests"`
	PendingEmergencies   int64 `json:"pending_emergencies"`
	AverageResolutionTime float64 `json:"average_resolution_time_minutes"`
	TotalMessages        int64 `json:"total_messages"`
	OnlineStaffCount     int   `json:"online_staff_count"`
}

// GetSupportMetrics returns support system metrics
// GET /admin/support/metrics
func (h *AdminHandler) GetSupportMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics := &SupportMetrics{}

	// Get conversation counts by status
	conversations, err := h.repo.ListConversations(ctx, uuid.Nil, 1000, 0)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list conversations for metrics")
		http.Error(w, "Failed to get metrics", http.StatusInternalServerError)
		return
	}

	metrics.TotalConversations = int64(len(conversations))
	var totalResolutionTime float64
	resolvedCount := 0

	for _, conv := range conversations {
		switch conv.Status {
		case support.StatusActive:
			metrics.ActiveConversations++
		case support.StatusPending:
			metrics.PendingConversations++
		case support.StatusResolved:
			metrics.ResolvedConversations++
			if conv.ResolvedAt != nil {
				resolvedCount++
				totalResolutionTime += conv.ResolvedAt.Sub(conv.CreatedAt).Minutes()
			}
		case support.StatusEscalated:
			metrics.EscalatedConversations++
		}

		if conv.IsEmergency {
			metrics.EmergencyRequests++
		}
	}

	if resolvedCount > 0 {
		metrics.AverageResolutionTime = totalResolutionTime / float64(resolvedCount)
	}

	// Get pending emergencies
	emergencies, err := h.repo.ListPendingEmergencies(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to list pending emergencies")
	} else {
		metrics.PendingEmergencies = int64(len(emergencies))
	}

	// Get online staff
	onlineStaff, err := h.repo.ListOnlineStaff(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to list online staff")
	} else {
		metrics.OnlineStaffCount = len(onlineStaff)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// ConversationSummary represents a conversation summary for admin view
type ConversationSummary struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Type              string    `json:"type"`
	Status            string    `json:"status"`
	Priority          string    `json:"priority"`
	Title             string    `json:"title"`
	AIHandled         bool      `json:"ai_handled"`
	IsEmergency       bool      `json:"is_emergency"`
	StaffID          *uuid.UUID `json:"staff_id,omitempty"`
	MessageCount      int       `json:"message_count"`
	CreatedAt         time.Time `json:"created_at"`
	LastMessageAt     time.Time `json:"last_message_at"`
	ResolutionTimeMin float64   `json:"resolution_time_minutes,omitempty"`
}

// ListAllConversations returns all conversations for admin view
// GET /admin/support/conversations
func (h *AdminHandler) ListAllConversations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all conversations (admin sees all)
	conversations, err := h.repo.ListConversations(ctx, uuid.Nil, 100, 0)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list conversations")
		http.Error(w, "Failed to list conversations", http.StatusInternalServerError)
		return
	}

	summaries := make([]*ConversationSummary, len(conversations))
	for i, conv := range conversations {
		summaries[i] = &ConversationSummary{
			ID:          conv.ID,
			UserID:      conv.UserID,
			Type:        string(conv.Type),
			Status:      string(conv.Status),
			Priority:    string(conv.Priority),
			Title:       conv.Title,
			AIHandled:   conv.AIHandled,
			IsEmergency:  conv.IsEmergency,
			StaffID:     conv.StaffID,
			CreatedAt:    conv.CreatedAt,
		}

		if conv.ResolvedAt != nil {
			summaries[i].ResolutionTimeMin = conv.ResolvedAt.Sub(conv.CreatedAt).Minutes()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summaries)
}

// StaffStatus represents staff availability status
type StaffStatus struct {
	StaffID      uuid.UUID `json:"staff_id"`
	IsOnline     bool      `json:"is_online"`
	ActiveChats  int       `json:"active_chats"`
	MaxChats     int       `json:"max_chats"`
	LastSeen     time.Time `json:"last_seen"`
}

// ListStaffStatus returns all staff status
// GET /admin/support/staff
func (h *AdminHandler) ListStaffStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	onlineStaff, err := h.repo.ListOnlineStaff(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list online staff")
		http.Error(w, "Failed to list staff", http.StatusInternalServerError)
		return
	}

	statuses := make([]*StaffStatus, len(onlineStaff))
	for i, staff := range onlineStaff {
		statuses[i] = &StaffStatus{
			StaffID:     staff.StaffID,
			IsOnline:    staff.IsOnline,
			ActiveChats: staff.ActiveChats,
			MaxChats:    staff.MaxChats,
			LastSeen:    staff.LastSeen,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(statuses)
}

// EmergencyRequest represents an emergency fix request
type EmergencyRequest struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	UserID        uuid.UUID  `json:"user_id"`
	FunctionID    uuid.UUID  `json:"function_id"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	StaffID       *uuid.UUID `json:"staff_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ListEmergencyRequests returns all emergency requests
// GET /admin/support/emergencies
func (h *AdminHandler) ListEmergencyRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	emergencies, err := h.repo.ListPendingEmergencies(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list emergencies")
		http.Error(w, "Failed to list emergencies", http.StatusInternalServerError)
		return
	}

	requests := make([]*EmergencyRequest, len(emergencies))
	for i, e := range emergencies {
		requests[i] = &EmergencyRequest{
			ID:             e.ID,
			ConversationID: e.ConversationID,
			UserID:         e.UserID,
			FunctionID:     e.FunctionID,
			Reason:        e.Reason,
			Status:        e.Status,
			StaffID:       e.StaffID,
			CreatedAt:     e.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
}

// RegisterRoutes registers admin routes for support
func (h *AdminHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/admin/support/metrics", h.GetSupportMetrics).Methods("GET")
	router.HandleFunc("/admin/support/conversations", h.ListAllConversations).Methods("GET")
	router.HandleFunc("/admin/support/staff", h.ListStaffStatus).Methods("GET")
	router.HandleFunc("/admin/support/emergencies", h.ListEmergencyRequests).Methods("GET")
}
