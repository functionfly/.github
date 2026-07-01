package aikeys

import (
	"encoding/json"
	"net/http"
)

// ConnectRequest is the request to connect a BYOK key.
type ConnectRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	Region   string `json:"region,omitempty"`
}

// RotateRequest is the request to rotate a BYOK key.
type RotateRequest struct {
	APIKey string `json:"apiKey"`
}

// KeyResponse is the response for a connected key (no secrets).
type KeyResponse struct {
	ID              string  `json:"id"`
	Provider        string  `json:"provider"`
	KeyLast4        string  `json:"key_last4"`
	Status          string  `json:"status"`
	HealthMessage   string  `json:"health_message,omitempty"`
	LastHealthCheck *string `json:"last_health_check,omitempty"`
	LastUsedAt      *string `json:"last_used_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	TokenPlanRegion string  `json:"token_plan_region,omitempty"`
}

// TestResponse is the response for a key test.
type TestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SupportedProvider describes a provider available for BYOK.
type SupportedProvider struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	KeyFormat      string `json:"key_format"`
	KeyPrefix      string `json:"key_prefix,omitempty"`
	IsMetaProvider bool   `json:"is_meta_provider,omitempty"`
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// DecodeJSON decodes a JSON request body.
func DecodeJSON(r *http.Request, v interface{}) error {
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(v)
}
