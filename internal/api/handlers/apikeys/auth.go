package apikeys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AuthAPIKeyRequest represents a request to authenticate with an API key
type AuthAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// AuthAPIKeyResponse represents the response after successful API key authentication
type AuthAPIKeyResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// APIKeyAuthHandler handles API key authentication
type APIKeyAuthHandler struct {
	repo      *apikey.Repository
	hasher    *apikey.Hasher
	jwtSecret string
}

// NewAPIKeyAuthHandler creates a new API key auth handler
// JWT secret must be at least 32 bytes (256 bits) for HS256 security
func NewAPIKeyAuthHandler(repo *apikey.Repository, jwtSecret string) *APIKeyAuthHandler {
	if len(jwtSecret) < 32 {
		panic(fmt.Sprintf("JWT_SECRET must be at least 32 bytes (256 bits) for HS256 security, got %d bytes", len(jwtSecret)))
	}
	return &APIKeyAuthHandler{
		repo:      repo,
		hasher:    apikey.NewHasher(),
		jwtSecret: jwtSecret,
	}
}

// HandleAPIKeyAuth returns an http.HandlerFunc for API key authentication
func HandleAPIKeyAuth(repo *apikey.Repository) http.HandlerFunc {
	h := NewAPIKeyAuthHandler(repo, os.Getenv("JWT_SECRET"))
	return h.HandleAuthenticate
}

// HandleAuthenticate handles POST /api/v1/auth/api-key
// Authenticates using an API key and returns a JWT token
func (h *APIKeyAuthHandler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req AuthAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate API key presence
	if req.APIKey == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key is required")
		return
	}

	// Hash the provided API key
	keyHash := h.hasher.Hash(req.APIKey)

	// Look up the API key
	ctx := context.Background()
	apiKey, err := h.repo.GetByHash(ctx, keyHash)
	if err != nil {
		logrus.WithError(err).Warn("API key authentication failed - invalid key")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	// Check if key is active
	if !apiKey.IsActive {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "API key is revoked")
		return
	}

	// Check if key has expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "API key has expired")
		return
	}

	// Update last used timestamp
	_ = h.repo.UpdateLastUsed(ctx, apiKey.ID)

	// Generate JWT (HS256, exp/iat/iss + key/tenant/user claims)
	token, expiresAt, err := h.generateJWT(apiKey)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT token")
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate authentication token")
		return
	}

	// Return success response
	h.writeSuccess(w, AuthAPIKeyResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// generateJWT generates a JWT token for the authenticated API key
func (h *APIKeyAuthHandler) generateJWT(apiKey *apikey.APIKey) (string, time.Time, error) {
	if len(h.jwtSecret) == 0 {
		return "", time.Time{}, fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	// Create API key-specific claims
	claims := jwt.MapClaims{
		"tenant_id": apiKey.TenantID.String(),
		"user_id":   apiKey.UserID.String(),
		"key_id":    apiKey.ID.String(),
		"key_name":  apiKey.Name,
		"key_type":  string(apiKey.KeyType),
		"exp":       expiresAt.Unix(),
		"iat":       now.Unix(),
		"iss":       "functionfly",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// writeJSON writes a JSON response
func (h *APIKeyAuthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeSuccess writes a success JSON response with the standard format
func (h *APIKeyAuthHandler) writeSuccess(w http.ResponseWriter, data interface{}) {
	resp := map[string]interface{}{
		"data": data,
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// writeError writes an error JSON response
func (h *APIKeyAuthHandler) writeError(w http.ResponseWriter, status int, code string, message string) {
	resp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	h.writeJSON(w, status, resp)
}

// RegisterAuthRoutes registers the API key auth routes
func RegisterAuthRoutes(router *mux.Router, repo *apikey.Repository, jwtSecret string) {
	h := NewAPIKeyAuthHandler(repo, jwtSecret)
	router.HandleFunc("/auth/api-key", h.HandleAuthenticate).Methods("POST", "OPTIONS")
}

// ValidateAPIKeyRequest represents a request to validate an API key (for AI service integration)
type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateAPIKeyResponse represents the response after validating an API key
type ValidateAPIKeyResponse struct {
	Valid       bool      `json:"valid"`
	TenantID    string    `json:"tenant_id,omitempty"`
	KeyID       string    `json:"key_id,omitempty"`
	Name        string    `json:"name,omitempty"`
	KeyType     string    `json:"key_type,omitempty"`
	Scopes      []string  `json:"scopes,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsRevoked   bool      `json:"is_revoked"`
	ExpiresAt   *string   `json:"expires_at,omitempty"`
	LastUsedAt  *string   `json:"last_used_at,omitempty"`
	RateLimitRPM int      `json:"rate_limit_rpm"`
	RateLimitRPH int      `json:"rate_limit_rph"`
	RateLimitRPD int      `json:"rate_limit_rpd"`
}

// HandleValidateAPIKey handles POST /api/v1/auth/validate-key
// Validates an API key and returns metadata (used by AI service for key validation)
// This endpoint does NOT generate a JWT - it returns raw key metadata
func (h *APIKeyAuthHandler) HandleValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req ValidateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate API key presence
	if req.APIKey == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key is required")
		return
	}

	// Hash the provided API key (using bcrypt, same as Go storage)
	keyHash := h.hasher.Hash(req.APIKey)

	// Look up the API key
	ctx := context.Background()
	apiKey, err := h.repo.GetByHash(ctx, keyHash)
	if err != nil {
		logrus.WithError(err).Debug("API key validation failed - key not found")
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": ValidateAPIKeyResponse{
				Valid: false,
			},
		})
		return
	}

	// Build scopes list from JSONB map
	var scopes []string
	if apiKey.Scopes != nil {
		for scope := range apiKey.Scopes {
			scopes = append(scopes, scope)
		}
	}

	// Format times
	var expiresAt *string
	if apiKey.ExpiresAt != nil {
		t := apiKey.ExpiresAt.Format(time.RFC3339)
		expiresAt = &t
	}
	var lastUsedAt *string
	if apiKey.LastUsedAt != nil {
		t := apiKey.LastUsedAt.Format(time.RFC3339)
		lastUsedAt = &t
	}

	// Update last used timestamp (fire and forget)
	go func() {
		_ = h.repo.UpdateLastUsed(ctx, apiKey.ID)
	}()

	// Return key metadata
	h.writeSuccess(w, ValidateAPIKeyResponse{
		Valid:        true,
		TenantID:     apiKey.TenantID.String(),
		KeyID:        apiKey.ID.String(),
		Name:         apiKey.Name,
		KeyType:      string(apiKey.KeyType),
		Scopes:       scopes,
		IsActive:     apiKey.IsActive,
		IsRevoked:    apiKey.IsRevoked,
		ExpiresAt:    expiresAt,
		LastUsedAt:   lastUsedAt,
		RateLimitRPM: apiKey.RateLimitRPM,
		RateLimitRPH: apiKey.RateLimitRPH,
		RateLimitRPD: apiKey.RateLimitRPD,
	})
}

// RegisterValidateRoutes registers the API key validation route for AI service
func RegisterValidateRoutes(router *mux.Router, repo *apikey.Repository, jwtSecret string) {
	h := NewAPIKeyAuthHandler(repo, jwtSecret)
	router.HandleFunc("/auth/validate-key", h.HandleValidateAPIKey).Methods("POST", "OPTIONS")
}
