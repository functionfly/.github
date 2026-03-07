package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/sirupsen/logrus"
)

// BetterAuthProxyConfig holds configuration for the Better Auth proxy
type BetterAuthProxyConfig struct {
	AuthServiceURL string
	Enabled        bool
	// Feature flags for gradual migration
	UseBetterAuthForLogin    bool
	UseBetterAuthForRegister bool
	UseBetterAuthForOAuth    bool
	UseBetterAuthForSession  bool
}

// BetterAuthProxy middleware proxies auth requests to Better Auth service
type BetterAuthProxy struct {
	config     *BetterAuthProxyConfig
	proxy      *httputil.ReverseProxy
	authURL    *url.URL
	authSvc    *auth.AuthService
	httpClient *http.Client
}

// NewBetterAuthProxy creates a new Better Auth proxy middleware
func NewBetterAuthProxy(authSvc *auth.AuthService) (*BetterAuthProxy, error) {
	config := &BetterAuthProxyConfig{
		AuthServiceURL:           getProxyEnv("AUTH_SERVICE_URL", "http://localhost:3001"),
		Enabled:                  getProxyEnvBool("BETTER_AUTH_ENABLED", false),
		UseBetterAuthForLogin:    getProxyEnvBool("BETTER_AUTH_LOGIN", false),
		UseBetterAuthForRegister: getProxyEnvBool("BETTER_AUTH_REGISTER", false),
		UseBetterAuthForOAuth:    getProxyEnvBool("BETTER_AUTH_OAUTH", false),
		UseBetterAuthForSession:  getProxyEnvBool("BETTER_AUTH_SESSION", false),
	}

	authURL, err := url.Parse(config.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTH_SERVICE_URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(authURL)

	// Customize the director to add tenant headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Extract tenant ID from request context or headers
		tenantID := extractTenantID(req)
		if tenantID != "" {
			req.Header.Set("X-Tenant-ID", tenantID)
		}

		// Forward device fingerprint if present
		if deviceFP := req.Header.Get("X-Device-Fingerprint"); deviceFP != "" {
			req.Header.Set("X-Device-Fingerprint", deviceFP)
		}

		// Ensure proper content type
		if req.Body != nil && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Error handler for proxy failures
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		logrus.WithError(err).Error("Better Auth proxy error")
		http.Error(w, `{"message": "Authentication service unavailable"}`, http.StatusServiceUnavailable)
	}

	return &BetterAuthProxy{
		config:     config,
		proxy:      proxy,
		authURL:    authURL,
		authSvc:    authSvc,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Middleware returns the HTTP middleware handler
func (p *BetterAuthProxy) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if Better Auth is enabled
		if !p.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Determine if this request should be proxied
		if !p.shouldProxy(r) {
			next.ServeHTTP(w, r)
			return
		}

		logrus.WithFields(logrus.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Debug("Proxying request to Better Auth service")

		// Proxy the request
		p.proxy.ServeHTTP(w, r)
	})
}

// shouldProxy determines if a request should be proxied to Better Auth
func (p *BetterAuthProxy) shouldProxy(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	// Better Auth endpoints
	betterAuthPaths := map[string]bool{
		"/api/auth/sign-in":                 p.config.UseBetterAuthForLogin,
		"/api/auth/sign-up":                 p.config.UseBetterAuthForRegister,
		"/api/auth/sign-out":                p.config.UseBetterAuthForSession,
		"/api/auth/session":                 p.config.UseBetterAuthForSession,
		"/api/auth/callback":                p.config.UseBetterAuthForOAuth,
		"/api/auth/callback/github":         p.config.UseBetterAuthForOAuth,
		"/api/auth/callback/google":         p.config.UseBetterAuthForOAuth,
		"/api/auth/forgot-password":         true, // Always use Better Auth for password reset
		"/api/auth/reset-password":          true,
		"/api/auth/verify-email":            true,
		"/api/auth/send-verification-email": true,
	}

	// Check exact path match
	if useBetterAuth, exists := betterAuthPaths[path]; exists {
		return useBetterAuth
	}

	// Check path prefix matches
	if strings.HasPrefix(path, "/api/auth/callback/") && p.config.UseBetterAuthForOAuth {
		return true
	}

	// Session validation endpoint
	if path == "/api/auth/validate-session" && p.config.UseBetterAuthForSession {
		return true
	}

	// Allow passthrough for OPTIONS requests (CORS preflight)
	if method == "OPTIONS" {
		return false
	}

	return false
}

// ValidateSession validates a session with the Better Auth service
func (p *BetterAuthProxy) ValidateSession(token string, tenantID string) (*SessionValidation, error) {
	if !p.config.Enabled || !p.config.UseBetterAuthForSession {
		// Fall back to legacy auth service
		return p.validateSessionLegacy(token, tenantID)
	}

	url := fmt.Sprintf("%s/api/validate-session", p.config.AuthServiceURL)

	payload := map[string]string{
		"token":    token,
		"tenantId": tenantID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact auth service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("session validation failed: %s", string(body))
	}

	var result SessionValidation
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ValidateTenant validates a tenant with the Better Auth service
func (p *BetterAuthProxy) ValidateTenant(tenantID string) (*TenantValidation, error) {
	url := fmt.Sprintf("%s/api/validate-tenant/%s", p.config.AuthServiceURL, tenantID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact auth service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result TenantValidation
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// validateSessionLegacy falls back to the legacy auth service
func (p *BetterAuthProxy) validateSessionLegacy(token string, tenantID string) (*SessionValidation, error) {
	// Validate the JWT token
	claims, err := p.authSvc.ValidateToken(token)
	if err != nil {
		return &SessionValidation{
			Valid:    false,
			Error:    err.Error(),
			Fallback: true,
		}, nil
	}

	// Get user details from repository
	user, err := p.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil {
		return &SessionValidation{
			Valid:    false,
			Error:    fmt.Sprintf("failed to get user: %v", err),
			Fallback: true,
		}, nil
	}
	if user == nil {
		return &SessionValidation{
			Valid:    false,
			Error:    "user not found",
			Fallback: true,
		}, nil
	}

	// Validate tenant access if tenantID is provided
	if tenantID != "" && user.TenantID.String() != tenantID {
		return &SessionValidation{
			Valid:    false,
			Error:    "user does not have access to this tenant",
			Fallback: true,
		}, nil
	}

	// Build session response
	session := &Session{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID.String(),
		Email:     user.Email,
		Role:      user.Role,
		ExpiresAt: claims.ExpiresAt.Time.Format(time.RFC3339),
	}

	return &SessionValidation{
		Valid:    true,
		Session:  session,
		Fallback: true,
	}, nil
}

// SessionValidation represents a session validation response
type SessionValidation struct {
	Valid    bool     `json:"valid"`
	Session  *Session `json:"session,omitempty"`
	Error    string   `json:"error,omitempty"`
	Fallback bool     `json:"fallback,omitempty"`
}

// Session represents user session information
type Session struct {
	UserID    string `json:"userId"`
	TenantID  string `json:"tenantId"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
	ExpiresAt string `json:"expiresAt"`
}

// TenantValidation represents a tenant validation response
type TenantValidation struct {
	Valid  bool        `json:"valid"`
	Tenant *TenantInfo `json:"tenant,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// TenantInfo represents tenant information
type TenantInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// extractTenantID extracts tenant ID from request
func extractTenantID(r *http.Request) string {
	// Check header first
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID != "" {
		return tenantID
	}

	// Try to extract from subdomain
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		subdomain := parts[0]
		// Skip common non-tenant subdomains
		reserved := []string{"www", "api", "auth", "admin", "app", "staging", "dev"}
		for _, r := range reserved {
			if strings.EqualFold(subdomain, r) {
				return ""
			}
		}
		return subdomain
	}

	return ""
}

// Helper functions for proxy configuration
func getProxyEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getProxyEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
