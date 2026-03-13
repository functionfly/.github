package apikeys

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// HandleCreate handles POST /api/v1/api-keys
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Parse request body
	var req apikey.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateCreateRequest(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Create the API key
	ctx := context.Background()
	apiKey, plaintext, err := h.repo.Create(ctx, claims.TenantID, claims.UserID, &req)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			h.writeError(w, http.StatusConflict, "duplicate_key", "An API key with this name already exists")
			return
		}
		errStr := err.Error()
		// Foreign key violation: user or tenant from session not found in DB
		if strings.Contains(errStr, "foreign key") || strings.Contains(errStr, "violates foreign key constraint") {
			logrus.WithError(err).WithField("tenant_id", claims.TenantID).WithField("user_id", claims.UserID).Warn("Create API key: user or tenant not found")
			h.writeError(w, http.StatusBadRequest, "invalid_claims", "Your session references a user or tenant that does not exist. Please sign out and sign in again.")
			return
		}
		// Unique constraint (e.g. duplicate name per tenant) — PostgreSQL 23505 or message substring
		if strings.Contains(errStr, "23505") ||
			strings.Contains(errStr, "unique constraint") ||
			strings.Contains(errStr, "duplicate key") {
			h.writeError(w, http.StatusConflict, "duplicate_key", "An API key with this name already exists")
			return
		}
		logrus.WithError(err).Error("Failed to create API key")
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create API key")
		return
	}

	// Build response with plaintext (only shown once!)
	resp := apikey.APIKeyCreateResponse{
		APIKeyResponse: *apiKey.ToResponse(),
		Plaintext:      plaintext,
	}

	// Set permissions and environments from the created key
	if len(apiKey.Permissions) > 0 {
		resp.Permissions = apiKey.Permissions
	}
	if len(apiKey.Environments) > 0 {
		resp.Environments = apiKey.Environments
	}

	h.writeSuccess(w, resp)
}

// validateCreateRequest validates the create API key request
func (h *Handler) validateCreateRequest(req *apikey.CreateAPIKeyRequest) error {
	var errors []string

	// Validate name (required)
	if req.Name == "" {
		errors = append(errors, "name is required")
	}

	// Validate key type (required)
	if req.KeyType == "" {
		errors = append(errors, "key_type is required")
	} else if !apikey.IsValidKeyType(string(req.KeyType)) {
		errors = append(errors, "invalid key_type")
	}

	// Validate rotation frequency
	if req.RotationFrequencyDays < 0 {
		errors = append(errors, "rotation_frequency_days must be non-negative")
	}

	// Validate rate limits if provided
	if req.RateLimit != nil {
		if req.RateLimit.RPM < 0 {
			errors = append(errors, "rate_limit.rpm must be non-negative")
		}
		if req.RateLimit.RPH < 0 {
			errors = append(errors, "rate_limit.rph must be non-negative")
		}
		if req.RateLimit.RPD < 0 {
			errors = append(errors, "rate_limit.rpd must be non-negative")
		}
	}

	// Validate permissions
	for i, perm := range req.Permissions {
		if perm.Permission == "" {
			errors = append(errors, "permission at index "+strconv.Itoa(i)+" must be specified")
		}
		if perm.ResourceType == "" {
			errors = append(errors, "permission resource_type at index "+strconv.Itoa(i)+" must be specified")
		}
	}

	if len(errors) > 0 {
		return &ValidationError{Message: joinErrors(errors)}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// joinErrors joins validation errors into a single string
func joinErrors(errors []string) string {
	result := ""
	for i, err := range errors {
		if i > 0 {
			result += "; "
		}
		result += err
	}
	return result
}

// init registers the route for HandleCreate
func init() {
	// This will be called from routes.go
}

// HandleCreateWithRouter creates the handler with router access
func HandleCreateWithRouter(repo *apikey.Repository) func(w http.ResponseWriter, r *http.Request) {
	h := NewHandler(repo)
	return func(w http.ResponseWriter, r *http.Request) {
		h.HandleCreate(w, r)
	}
}

// RegisterRoutes registers the API key routes
func RegisterRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo)

	// API Keys CRUD
	router.HandleFunc("/api-keys", h.HandleCreate).Methods("POST", "OPTIONS")
}
