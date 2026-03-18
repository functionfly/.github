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
func NewAPIKeyAuthHandler(repo *apikey.Repository, jwtSecret string) *APIKeyAuthHandler {
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
