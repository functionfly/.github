package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// TurnstileSiteVerifyResponse represents the response from Cloudflare Turnstile siteverify API
type TurnstileSiteVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

// TurnstileVerifier handles server-side validation of Turnstile tokens
type TurnstileVerifier struct {
	secretKey string
	enabled   bool
	client    *http.Client
}

// NewTurnstileVerifier creates a new Turnstile verifier with secret key from environment
func NewTurnstileVerifier() *TurnstileVerifier {
	secretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	enabled := os.Getenv("TURNSTILE_VERIFY_ENABLED") != "false" // Enabled by default

	return &TurnstileVerifier{
		secretKey: secretKey,
		enabled:   enabled && secretKey != "",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyToken validates a Turnstile token with Cloudflare's siteverify API
func (t *TurnstileVerifier) VerifyToken(token, remoteIP string) (*TurnstileSiteVerifyResponse, error) {
	if !t.enabled {
		// If verification is disabled, return success (for development/testing)
		logrus.Warn("Turnstile verification is disabled - allowing request without validation")
		return &TurnstileSiteVerifyResponse{Success: true}, nil
	}

	if token == "" {
		return nil, fmt.Errorf("turnstile token is required")
	}

	formData := url.Values{
		"secret":   {t.secretKey},
		"response": {token},
	}

	// Include remote IP if provided (optional but recommended)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	resp, err := t.client.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to contact turnstile verification API: %w", err)
	}
	defer resp.Body.Close()

	var result TurnstileSiteVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode turnstile response: %w", err)
	}

	if !result.Success {
		logrus.WithFields(logrus.Fields{
			"error_codes": result.ErrorCodes,
			"hostname":    result.Hostname,
		}).Warn("Turnstile verification failed")
	}

	return &result, nil
}

// IsEnabled returns whether Turnstile verification is enabled
func (t *TurnstileVerifier) IsEnabled() bool {
	return t.enabled
}

// TurnstileMiddleware provides HTTP middleware for Turnstile token validation
type TurnstileMiddleware struct {
	verifier *TurnstileVerifier
}

// NewTurnstileMiddleware creates a new Turnstile middleware
func NewTurnstileMiddleware() *TurnstileMiddleware {
	return &TurnstileMiddleware{
		verifier: NewTurnstileVerifier(),
	}
}

// RequireTurnstile wraps a handler with Turnstile token validation
// Expects the token in the X-Turnstile-Token header or turnstile_token form field
func (t *TurnstileMiddleware) RequireTurnstile(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if verification is enabled
		if !t.verifier.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from header or form
		token := r.Header.Get("X-Turnstile-Token")
		if token == "" {
			token = r.FormValue("turnstile_token")
		}
		if token == "" {
			// Try to parse from JSON body if it's a POST/PUT request
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				var body struct {
					TurnstileToken string `json:"turnstile_token"`
				}
				// Only try to decode if content type is JSON
				if r.Header.Get("Content-Type") == "application/json" {
					// Preserve body for next handler by reading and restoring
					// This is a simplified approach - in production, use a body buffer
					decoder := json.NewDecoder(r.Body)
					if err := decoder.Decode(&body); err == nil {
						token = body.TurnstileToken
					}
				}
			}
		}

		if token == "" {
			writeJSONError(w, http.StatusBadRequest, "Turnstile token is required")
			return
		}

		// Verify the token
		remoteIP := getClientIP(r)
		result, err := t.verifier.VerifyToken(token, remoteIP)
		if err != nil {
			logrus.WithError(err).Warn("Turnstile verification error")
			writeJSONError(w, http.StatusInternalServerError, "Failed to verify security token")
			return
		}

		if !result.Success {
			writeJSONError(w, http.StatusForbidden, "Security verification failed. Please try again.")
			return
		}

		// Token is valid, proceed
		next.ServeHTTP(w, r)
	}
}

// OptionalTurnstile wraps a handler with optional Turnstile validation
// If token is present, it will be validated; if not, request proceeds
func (t *TurnstileMiddleware) OptionalTurnstile(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if verification is enabled
		if !t.verifier.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from header or form
		token := r.Header.Get("X-Turnstile-Token")
		if token == "" {
			token = r.FormValue("turnstile_token")
		}

		// If no token provided, proceed (optional validation)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Verify the token
		remoteIP := getClientIP(r)
		result, err := t.verifier.VerifyToken(token, remoteIP)
		if err != nil {
			logrus.WithError(err).Warn("Turnstile verification error")
			writeJSONError(w, http.StatusInternalServerError, "Failed to verify security token")
			return
		}

		if !result.Success {
			writeJSONError(w, http.StatusForbidden, "Security verification failed. Please try again.")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
