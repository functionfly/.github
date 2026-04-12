package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListIncidents lists all incidents
func (h *Handler) HandleListIncidents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")

	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	incidents, err := h.repo.ListIncidents(r.Context(), limit, offset, statusPtr)
	if err != nil {
		logrus.WithError(err).Error("Failed to list incidents")
		http.Error(w, "Failed to list incidents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incidents": incidents,
	})
}

// HandleCreateIncident creates a new incident
func (h *Handler) HandleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Severity == "" || req.Description == "" {
		http.Error(w, "Title, severity, and description are required", http.StatusBadRequest)
		return
	}

	// Validate severity
	validSeverities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}
	if !validSeverities[req.Severity] {
		http.Error(w, "Invalid severity. Must be one of: critical, high, medium, low", http.StatusBadRequest)
		return
	}

	incident := &storage.Incident{
		Title:       req.Title,
		Severity:    req.Severity,
		Status:      "investigating", // Default status for new incidents
		Description: req.Description,
	}

	createdIncident, err := h.repo.CreateIncident(r.Context(), incident)
	if err != nil {
		logrus.WithError(err).Error("Failed to create incident")
		http.Error(w, "Failed to create incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdIncident)
}

// HandleGetIncident gets a specific incident
func (h *Handler) HandleGetIncident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentIDStr := vars["incidentId"]

	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	incident, err := h.repo.GetIncidentByID(r.Context(), incidentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Incident not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get incident")
		http.Error(w, "Failed to get incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

// HandleUpdateIncident updates an incident
func (h *Handler) HandleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentIDStr := vars["incidentId"]

	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate status if provided
	if status, ok := updates["status"].(string); ok {
		validStatuses := map[string]bool{
			"resolved":      true,
			"investigating": true,
			"monitoring":    true,
		}
		if !validStatuses[status] {
			http.Error(w, "Invalid status. Must be one of: resolved, investigating, monitoring", http.StatusBadRequest)
			return
		}
	}

	// Validate severity if provided
	if severity, ok := updates["severity"].(string); ok {
		validSeverities := map[string]bool{
			"critical": true,
			"high":     true,
			"medium":   true,
			"low":      true,
		}
		if !validSeverities[severity] {
			http.Error(w, "Invalid severity. Must be one of: critical, high, medium, low", http.StatusBadRequest)
			return
		}
	}

	updatedIncident, err := h.repo.UpdateIncident(r.Context(), incidentID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Incident not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to update incident")
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedIncident)
}

// HandleResolveIncident marks an incident as resolved
func (h *Handler) HandleResolveIncident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentIDStr := vars["incidentId"]

	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	resolvedIncident, err := h.repo.ResolveIncident(r.Context(), incidentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Incident not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to resolve incident")
		http.Error(w, "Failed to resolve incident", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resolvedIncident)
}
