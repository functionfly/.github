package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminEmailEventsHandler handles admin operations for email events
type AdminEmailEventsHandler struct {
	repo storage.Repository
}

// NewAdminEmailEventsHandler creates a new admin email events handler
func NewAdminEmailEventsHandler(repo storage.Repository) *AdminEmailEventsHandler {
	return &AdminEmailEventsHandler{
		repo: repo,
	}
}

// RegisterRoutes registers admin email event routes
func (h *AdminEmailEventsHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/admin/email-events", h.ListEmailEvents).Methods("GET")
	router.HandleFunc("/admin/email-events/bounces", h.ListPendingBounces).Methods("GET")
	router.HandleFunc("/admin/email-events/{id}/review", h.MarkEventReviewed).Methods("POST")
	router.HandleFunc("/admin/email-events/stats", h.GetEmailStats).Methods("GET")
}

// ListEmailEvents returns a paginated list of email events with optional filters
func (h *AdminEmailEventsHandler) ListEmailEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Build filters
	filters := make(map[string]interface{})
	if eventType := query.Get("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}
	if userEmail := query.Get("user_email"); userEmail != "" {
		filters["user_email"] = userEmail
	}
	if reviewed := query.Get("reviewed"); reviewed != "" {
		filters["reviewed"] = reviewed == "true"
	}
	if userID := query.Get("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filters["user_id"] = uid
		}
	}

	// Parse pagination
	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get events from database
	events, err := h.repo.GetEmailEvents(r.Context(), filters, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve email events")
		http.Error(w, `{"error": "Failed to retrieve email events"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"limit":  limit,
		"offset": offset,
	}); err != nil {
		logrus.WithError(err).Error("Failed to encode email events response")
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// ListPendingBounces returns email bounce/complaint events that need admin review
func (h *AdminEmailEventsHandler) ListPendingBounces(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	query := r.URL.Query()
	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get pending bounce reviews from database
	events, err := h.repo.GetPendingBounceReviews(r.Context(), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve pending bounce reviews")
		http.Error(w, `{"error": "Failed to retrieve pending bounce reviews"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"limit":  limit,
		"offset": offset,
	}); err != nil {
		logrus.WithError(err).Error("Failed to encode pending bounces response")
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// MarkEventReviewed marks an email event as reviewed by an admin
func (h *AdminEmailEventsHandler) MarkEventReviewed(w http.ResponseWriter, r *http.Request) {
	// Get event ID from URL parameters
	vars := mux.Vars(r)
	eventIDStr := vars["id"]
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error": "Invalid event ID"}`, http.StatusBadRequest)
		return
	}

	// Get admin user ID from context (should be set by auth middleware)
	// For now, we'll accept it from the request body as a fallback
	type ReviewRequest struct {
		AdminID string `json:"admin_id"`
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	adminID, err := uuid.Parse(req.AdminID)
	if err != nil {
		http.Error(w, `{"error": "Invalid admin ID"}`, http.StatusBadRequest)
		return
	}

	// Mark event as reviewed
	if err := h.repo.MarkEmailEventReviewed(r.Context(), eventID, adminID); err != nil {
		logrus.WithError(err).WithField("event_id", eventID).Error("Failed to mark event as reviewed")
		http.Error(w, `{"error": "Failed to mark event as reviewed"}`, http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"event_id": eventID,
		"message":  "Event marked as reviewed",
	})
}

// GetEmailStats returns aggregated email statistics
func (h *AdminEmailEventsHandler) GetEmailStats(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filters
	query := r.URL.Query()
	filters := make(map[string]interface{})

	if userEmail := query.Get("user_email"); userEmail != "" {
		filters["user_email"] = userEmail
	}
	if userID := query.Get("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filters["user_id"] = uid
		}
	}

	// Parse date range if provided
	if startDate := query.Get("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filters["start_date"] = t
		}
	}
	if endDate := query.Get("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filters["end_date"] = t
		}
	}

	// Get statistics from database
	stats, err := h.repo.GetEmailEventStats(r.Context(), filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve email statistics")
		http.Error(w, `{"error": "Failed to retrieve email statistics"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logrus.WithError(err).Error("Failed to encode email stats response")
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}
