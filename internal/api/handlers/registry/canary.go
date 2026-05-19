package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// CanaryHandler handles canary deployment endpoints
type CanaryHandler struct {
	canaryRepo   *registry.CanaryConfigRepository
	functionRepo *registry.RegistryRepository
}

// NewCanaryHandler creates a new canary handler
func NewCanaryHandler(canaryRepo *registry.CanaryConfigRepository, functionRepo *registry.RegistryRepository) *CanaryHandler {
	return &CanaryHandler{
		canaryRepo:   canaryRepo,
		functionRepo: functionRepo,
	}
}

// CanaryCreateRequest represents a canary deployment creation request
type CanaryCreateRequest struct {
	Version        string  `json:"version"`
	TrafficPercent int     `json:"traffic_percent"`
	AutoPromote   bool    `json:"auto_promote"`
	PromoteThreshold float64 `json:"promote_threshold"`
	PromoteWindow  int     `json:"promote_window"`
}

// CanaryUpdateRequest represents a canary deployment update request
type CanaryUpdateRequest struct {
	TrafficPercent *int     `json:"traffic_percent,omitempty"`
	AutoPromote   *bool    `json:"auto_promote,omitempty"`
	PromoteThreshold *float64 `json:"promote_threshold,omitempty"`
	PromoteWindow  *int     `json:"promote_window,omitempty"`
}

// HandleCreateCanary creates a new canary deployment
func (h *CanaryHandler) HandleCreateCanary(w http.ResponseWriter, r *http.Request) {
	// Extract author and name from URL
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get function from repository
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		logrus.WithError(err).Warn("Function not found for canary deployment")
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if there's already an active canary
	existingCanary, _ := h.canaryRepo.GetByFunctionID(fn.ID)
	if existingCanary != nil {
		http.Error(w, "An active canary deployment already exists for this function. Cancel it first.", http.StatusConflict)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CanaryCreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Version == "" {
		http.Error(w, "version is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.TrafficPercent <= 0 || req.TrafficPercent > 100 {
		req.TrafficPercent = 10 // Default 10%
	}
	if req.PromoteWindow <= 0 {
		req.PromoteWindow = 300 // Default 5 minutes
	}
	if req.PromoteThreshold <= 0 {
		req.PromoteThreshold = 0.01 // Default 1% error rate
	}

	// Create canary config
	canary := &registry.CanaryConfig{
		FunctionID:       fn.ID,

		Version:          req.Version,
		TrafficPercent:  req.TrafficPercent,
		AutoPromote:     req.AutoPromote,
		PromoteThreshold: req.PromoteThreshold,
		PromoteWindow:    req.PromoteWindow,
		Status:           "active",
	}

	if err := h.canaryRepo.Create(canary); err != nil {
		logrus.WithError(err).Error("Failed to create canary config")
		http.Error(w, "Failed to create canary deployment", http.StatusInternalServerError)
		return
	}

	// Return the created canary config
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(canary)
}

// HandleGetCanary returns the canary configuration for a function
func (h *CanaryHandler) HandleGetCanary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Get canary config
	canary, err := h.canaryRepo.GetByFunctionID(fn.ID)
	if err != nil {
		// No active canary - return empty response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
			"message": "No active canary deployment",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(canary)
}

// HandleUpdateCanary updates an existing canary deployment
func (h *CanaryHandler) HandleUpdateCanary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get canary config
	canary, err := h.canaryRepo.GetByFunctionID(fn.ID)
	if err != nil {
		http.Error(w, "No active canary deployment", http.StatusNotFound)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CanaryUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.TrafficPercent != nil {
		if *req.TrafficPercent < 0 || *req.TrafficPercent > 100 {
			http.Error(w, "traffic_percent must be between 0 and 100", http.StatusBadRequest)
			return
		}
		canary.TrafficPercent = *req.TrafficPercent
	}
	if req.AutoPromote != nil {
		canary.AutoPromote = *req.AutoPromote
	}
	if req.PromoteThreshold != nil {
		canary.PromoteThreshold = *req.PromoteThreshold
	}
	if req.PromoteWindow != nil {
		canary.PromoteWindow = *req.PromoteWindow
	}

	if err := h.canaryRepo.Update(canary); err != nil {
		logrus.WithError(err).Error("Failed to update canary config")
		http.Error(w, "Failed to update canary deployment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(canary)
}

// HandleCancelCanary cancels an active canary deployment
func (h *CanaryHandler) HandleCancelCanary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get canary config
	canary, err := h.canaryRepo.GetByFunctionID(fn.ID)
	if err != nil {
		http.Error(w, "No active canary deployment", http.StatusNotFound)
		return
	}

	// Update status to cancelled
	if err := h.canaryRepo.UpdateStatus(canary.ID, "cancelled"); err != nil {
		logrus.WithError(err).Error("Failed to cancel canary")
		http.Error(w, "Failed to cancel canary deployment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Canary deployment cancelled",
		"status":  "cancelled",
	})
}

// HandlePromoteCanary promotes a canary to stable
func (h *CanaryHandler) HandlePromoteCanary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get canary config
	canary, err := h.canaryRepo.GetByFunctionID(fn.ID)
	if err != nil {
		http.Error(w, "No active canary deployment", http.StatusNotFound)
		return
	}

	// Update function's latest version to canary version
	if err := h.functionRepo.UpdateFunctionLatestVersion(fn.ID, canary.Version); err != nil {
		logrus.WithError(err).Error("Failed to update function version")
		http.Error(w, "Failed to promote canary", http.StatusInternalServerError)
		return
	}

	// Mark canary as promoted
	if err := h.canaryRepo.UpdateStatus(canary.ID, "promoted"); err != nil {
		logrus.WithError(err).Error("Failed to update canary status")
	}

	// Add deprecation warning header
	middleware.AddVersionHeaders(w, "v1", "", nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Canary promoted to stable version",
		"new_version":     canary.Version,
		"previous_status": "active",
		"new_status":      "promoted",
	})
}

// HandleRollbackCanary rolls back a canary deployment
func (h *CanaryHandler) HandleRollbackCanary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get canary config
	canary, err := h.canaryRepo.GetByFunctionID(fn.ID)
	if err != nil {
		http.Error(w, "No active canary deployment", http.StatusNotFound)
		return
	}

	// Mark canary as rolled_back
	if err := h.canaryRepo.UpdateStatus(canary.ID, "rolled_back"); err != nil {
		logrus.WithError(err).Error("Failed to update canary status")
		http.Error(w, "Failed to rollback canary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Canary rolled back",
		"canary_version": canary.Version,
		"status":     "rolled_back",
	})
}

// HandleGetCanaryHistory returns the canary history for a function
func (h *CanaryHandler) HandleGetCanaryHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	// Get function
	fn, err := h.functionRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Get all canary configs
	canaries, err := h.canaryRepo.GetAllByFunctionID(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get canary history")
		http.Error(w, "Failed to get canary history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(canaries)
}

// GetCanaryVersion determines which version to route to based on canary config
func (h *CanaryHandler) GetCanaryVersion(functionID uuid.UUID, stableVersion string, hash string) (string, bool) {
	canary, err := h.canaryRepo.GetByFunctionID(functionID)
	if err != nil {
		// No canary, use stable version
		return stableVersion, false
	}

	// Check if canary is still active
	if canary.Status != "active" {
		return stableVersion, false
	}

	// Determine routing based on traffic percentage
	// Use hash to make routing deterministic for the same request
	hashVal := hashString(hash)
	percent := hashVal % 100

	if percent < canary.TrafficPercent {
		return canary.Version, true // Route to canary
	}

	return stableVersion, false
}

// hashString creates a simple hash for deterministic routing
func hashString(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = h*31 + int(s[i])
	}
	if h < 0 {
		h = -h
	}
	return h % 10000
}

// AutoPromoteCanary checks and auto-promotes canaries that meet the criteria
func (h *CanaryHandler) AutoPromoteCanary(canaryID uuid.UUID) error {
	shouldPromote, err := h.canaryRepo.AutoPromoteCheck(canaryID)
	if err != nil {
		return err
	}

	if shouldPromote {
		canary, err := h.canaryRepo.GetByID(canaryID)
		if err != nil {
			return err
		}

		// Get function and update latest version
		fn, err := h.functionRepo.GetFunctionByID(canary.FunctionID)
		if err != nil {
			return err
		}

		if err := h.functionRepo.UpdateFunctionLatestVersion(fn.ID, canary.Version); err != nil {
			return err
		}

		// Mark canary as promoted
		return h.canaryRepo.UpdateStatus(canaryID, "promoted")
	}

	return nil
}

// GetCanaryConfigFromRequest extracts canary configuration from request
func GetCanaryConfigFromRequest(r *http.Request) (bool, int) {
	// Check for canary percentage in query params
	canaryStr := r.URL.Query().Get("canary")
	if canaryStr == "" {
		return false, 0
	}

	// Check for explicit canary routing
	if canaryStr == "true" || canaryStr == "1" {
		return true, 10 // Default 10%
	}

	// Try to parse as percentage
	if percent, err := strconv.Atoi(canaryStr); err == nil {
		if percent > 0 && percent <= 100 {
			return true, percent
		}
	}

	return false, 0
}

// GetCanaryHash generates a hash for canary routing decisions
func GetCanaryHash(r *http.Request) string {
	// Use a combination of headers and IP for consistent routing
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")

	// Combine for hash
	return fmt.Sprintf("%s:%s:%s", r.URL.Path, ip, userAgent)
}
