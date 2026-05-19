package studio

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles studio collaboration HTTP requests.
type Handler struct {
	repo *CollabRepository
}

// NewHandler creates a new studio collab handler.
func NewHandler(repo *CollabRepository) *Handler {
	return &Handler{repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getTenantID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.TenantID.String()
}

func getUserID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.UserID.String()
}

func getEnvironment(r *http.Request) string {
	return middleware.GetEnvironmentFromContext(r)
}

// HandleListEvents handles GET /v1/studio/collab/events?type=&limit=&offset=
func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventType := r.URL.Query().Get("type")
	environment := getEnvironment(r)
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	events, err := h.repo.ListEvents(r.Context(), tenantID, eventType, environment, limit, offset)
	if err != nil {
		logrus.WithError(err).Warn("studio collab: failed to list events")
		writeJSON(w, http.StatusOK, map[string]interface{}{"events": []CollabEvent{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"events": events})
}

// HandleGetEvent handles GET /v1/studio/collab/events/{id}
func (h *Handler) HandleGetEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	environment := getEnvironment(r)

	ev, err := h.repo.GetEvent(r.Context(), tenantID, id, environment)
	if err != nil {
		logrus.WithError(err).Warn("studio collab: failed to get event")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get event")
		return
	}
	if ev == nil {
		writeJSONError(w, http.StatusNotFound, "Event not found")
		return
	}

	writeJSON(w, http.StatusOK, ev)
}

// HandleCreateEvent handles POST /v1/studio/collab/events
func (h *Handler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		EventType string                 `json:"event_type"`
		Metadata  map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.EventType == "" {
		writeJSONError(w, http.StatusBadRequest, "event_type is required")
		return
	}

	ev, err := h.repo.CreateEvent(r.Context(), tenantID, req.EventType, userID, environment, req.Metadata)
	if err != nil {
		logrus.WithError(err).Error("studio collab: failed to create event")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	writeJSON(w, http.StatusCreated, ev)
}

// HandleUpdateEvent handles PATCH /v1/studio/collab/events/{id}
func (h *Handler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ev, err := h.repo.UpdateEvent(r.Context(), tenantID, id, environment, req.Metadata)
	if err != nil {
		logrus.WithError(err).Error("studio collab: failed to update event")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}
	if ev == nil {
		writeJSONError(w, http.StatusNotFound, "Event not found")
		return
	}

	writeJSON(w, http.StatusOK, ev)
}

// HandleDeleteEvent handles DELETE /v1/studio/collab/events/{id}
func (h *Handler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	environment := getEnvironment(r)

	if err := h.repo.DeleteEvent(r.Context(), tenantID, id, environment); err != nil {
		logrus.WithError(err).Error("studio collab: failed to delete event")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete event")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Event deleted"})
}

// HandleGetActivityFeed synthesizes a tenant activity feed from dashboard activity + collab events.
// GET /v1/studio/collab/activity
func (h *Handler) HandleGetActivityFeed(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	environment := getEnvironment(r)
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	// Return collab events with type 'activity' as the feed
	events, err := h.repo.ListEvents(r.Context(), tenantID, "activity", environment, limit, 0)
	if err != nil {
		logrus.WithError(err).Warn("studio collab: failed to list activity events")
		writeJSON(w, http.StatusOK, map[string]interface{}{"activities": []CollabEvent{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"activities": events})
}

// HandleCreateActivity handles POST /v1/studio/collab/activity
func (h *Handler) HandleCreateActivity(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		Action   string `json:"action"`
		Target   string `json:"target"`
		Icon     string `json:"icon"`
		UserName string `json:"user_name"`
		UserColor string `json:"user_color"`
		IsAI     bool   `json:"is_ai"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	metadata := map[string]interface{}{
		"action":     req.Action,
		"target":     req.Target,
		"icon":       req.Icon,
		"user_name":  req.UserName,
		"user_color": req.UserColor,
		"is_ai":      req.IsAI,
	}

	ev, err := h.repo.CreateEvent(r.Context(), tenantID, "activity", userID, environment, metadata)
	if err != nil {
		logrus.WithError(err).Error("studio collab: failed to create activity")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	writeJSON(w, http.StatusCreated, ev)
}

// HandleGetTelemetry handles GET /v1/studio/telemetry?hours=24
func (h *Handler) HandleGetTelemetry(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	environment := getEnvironment(r)
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 168 {
			hours = n
		}
	}

	metrics, err := h.repo.GetTelemetryMetrics(r.Context(), tenantID, environment, hours)
	if err != nil {
		logrus.WithError(err).Warn("studio collab: failed to get telemetry")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"metrics": []TelemetryMetric{},
			"hours":   hours,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"hours":   hours,
	})
}
