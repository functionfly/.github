// Package gba provides GoBetterAuth integration for FunctionFly
// This is Phase 1 of the Better Auth migration plan
package gba

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/email"
)

// Config holds the GoBetterAuth configuration
type Config struct {
	// Core settings
	AppName  string
	BasePath string
	Secret   string

	// Database
	DB *gorm.DB

	// OAuth providers
	OAuth OAuthConfig

	// Session settings
	Session SessionConfig

	// Feature flags for gradual migration
	Features FeatureFlags
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	GitHubEnabled bool
	GitHub        oauth2.Config

	GoogleEnabled bool
	Google        oauth2.Config
}

// SessionConfig holds session management settings
type SessionConfig struct {
	MaxAge         time.Duration
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
}

// FeatureFlags control gradual migration
type FeatureFlags struct {
	Enabled  bool // Master switch
	Login    bool // Use GoBetterAuth for login
	Register bool // Use GoBetterAuth for registration
	OAuth    bool // Use GoBetterAuth for OAuth
	Session  bool // Use GoBetterAuth for session validation
}

// Auth represents the GoBetterAuth instance
type Auth struct {
	config   *Config
	db       *gorm.DB
	oauth    *OAuthManager
	sessions *SessionManager
	hooks    *HookManager
	logger   *logrus.Logger
	emailSvc email.Service // optional; when set, verification emails are sent
}

// New creates a new GoBetterAuth instance
func New(cfg *Config) (*Auth, error) {
	if cfg.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if cfg.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	// Set defaults
	if cfg.AppName == "" {
		cfg.AppName = "FunctionFly"
	}
	if cfg.BasePath == "" {
		cfg.BasePath = "/v1/auth"
	}
	if cfg.Session.MaxAge == 0 {
		cfg.Session.MaxAge = 7 * 24 * time.Hour // 7 days
	}
	if cfg.Session.CookieName == "" {
		cfg.Session.CookieName = "ff_session"
	}

	auth := &Auth{
		config: cfg,
		db:     cfg.DB,
		logger: logrus.New(),
	}

	// Initialize OAuth manager
	auth.oauth = NewOAuthManager(&cfg.OAuth)

	// Initialize session manager
	auth.sessions = NewSessionManager(cfg.Secret, &cfg.Session)

	// Initialize hook manager
	auth.hooks = NewHookManager()

	// Register default hooks
	auth.registerDefaultHooks()

	return auth, nil
}

// SetEmailService sets the email service for sending verification and password-reset emails. Optional.
func (a *Auth) SetEmailService(svc email.Service) {
	a.emailSvc = svc
}

// EmailService returns the configured email service, or nil if not set.
func (a *Auth) EmailService() email.Service {
	return a.emailSvc
}

// ConfigFromEnv creates configuration from environment variables
func ConfigFromEnv(db *gorm.DB) (*Config, error) {
	baseURL := getEnv("BASE_URL", "http://localhost:8080")

	cfg := &Config{
		AppName:  getEnv("APP_NAME", "FunctionFly"),
		BasePath: getEnv("AUTH_BASE_PATH", "/v1/auth"),
		Secret:   getEnv("JWT_SECRET", ""),
		DB:       db,
		Session: SessionConfig{
			MaxAge:         parseDuration(getEnv("SESSION_MAX_AGE", "168h")),
			CookieName:     getEnv("SESSION_COOKIE_NAME", "ff_session"),
			CookieSecure:   getEnvBool("SESSION_COOKIE_SECURE", true),
			CookieHTTPOnly: getEnvBool("SESSION_COOKIE_HTTPONLY", true),
			CookieSameSite: getEnv("SESSION_COOKIE_SAMESITE", "Lax"),
		},
		Features: FeatureFlags{
			Enabled:  getEnvBool("GBA_ENABLED", false),
			Login:    getEnvBool("GBA_LOGIN", false),
			Register: getEnvBool("GBA_REGISTER", false),
			OAuth:    getEnvBool("GBA_OAUTH", false),
			Session:  getEnvBool("GBA_SESSION", false),
		},
	}

	// Configure OAuth providers
	if clientID := getEnv("GITHUB_CLIENT_ID", ""); clientID != "" {
		cfg.OAuth.GitHubEnabled = true
		cfg.OAuth.GitHub = oauth2.Config{
			ClientID:     clientID,
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			Endpoint:     github.Endpoint,
			RedirectURL:  baseURL + "/v1/auth/callback/github",
			Scopes:       []string{"user:email", "read:user"},
		}
	}

	if clientID := getEnv("GOOGLE_CLIENT_ID", ""); clientID != "" {
		cfg.OAuth.GoogleEnabled = true
		cfg.OAuth.Google = oauth2.Config{
			ClientID:     clientID,
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			Endpoint:     google.Endpoint,
			RedirectURL:  baseURL + "/v1/auth/callback/google",
			Scopes:       []string{"openid", "email", "profile"},
		}
	}

	return cfg, nil
}

// IsEnabled returns true if GoBetterAuth is enabled for a specific feature
func (a *Auth) IsEnabled(feature string) bool {
	if !a.config.Features.Enabled {
		return false
	}

	switch feature {
	case "login":
		return a.config.Features.Login
	case "register":
		return a.config.Features.Register
	case "oauth":
		return a.config.Features.OAuth
	case "session":
		return a.config.Features.Session
	default:
		return false
	}
}

// GetDB returns the database connection
func (a *Auth) GetDB() *gorm.DB {
	return a.db
}

// GetConfig returns the configuration
func (a *Auth) GetConfig() *Config {
	return a.config
}

// Logger returns the logger instance
func (a *Auth) Logger() *logrus.Logger {
	return a.logger
}

// ValidateToken validates a JWT token
func (a *Auth) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// getPermissionsForRole returns permissions for a given role using the unified role system
func (a *Auth) getPermissionsForRole(role string) []string {
	switch role {
	case "super_admin", "admin":
		return []string{
			"tenants.read", "tenants.write",
			"users.read", "users.write",
			"billing.read", "billing.write",
			"deployments.read", "deployments.write",
			"audit.read",
			"system.read", "system.write",
			"apps.read", "apps.write", "apps.delete",
			"functions.read", "functions.write", "functions.delete",
			"backends.read", "backends.write", "backends.delete",
			"teams.read", "teams.write", "teams.delete",
			"state.read", "state.write", "state.delete", "state.admin",
			"triggers.manage",
			"snapshots.create", "snapshots.restore",
			"replay.access",
			"memory.read", "memory.write",
			"registry.publish", "registry.verify", "registry.approve",
			"registry.sign", "registry.manage",
			"monitoring.alerts", "monitoring.manage", "monitoring.metrics",
			"monitoring.admin", "monitoring.health",
			"security.incidents", "security.scans", "security.audit", "security.admin",
			"content.create", "content.edit", "content.publish", "content.manage",
			"changelog.manage",
		}
	case "support":
		return []string{
			"users.read",
			"audit.read",
			"monitoring.health", "monitoring.metrics",
			"security.incidents",
		}
	case "billing_admin":
		return []string{
			"billing.read", "billing.write",
			"tenants.read",
			"users.read",
		}
	case "developer_admin":
		return []string{
			"deployments.read", "deployments.write", "deployments.delete",
			"apps.read", "apps.write", "apps.delete",
			"functions.read", "functions.write", "functions.delete",
			"backends.read", "backends.write", "backends.delete",
			"registry.publish", "registry.verify", "registry.approve",
			"registry.sign", "registry.manage",
			"monitoring.alerts", "monitoring.manage", "monitoring.metrics",
			"monitoring.admin", "monitoring.health",
		}
	case "read_only":
		return []string{
			"tenants.read",
			"users.read",
			"billing.read",
			"deployments.read",
			"audit.read",
			"system.read",
			"apps.read",
			"functions.read",
			"backends.read",
			"teams.read",
			"state.read",
			"memory.read",
			"monitoring.metrics", "monitoring.health",
			"security.audit",
			"content.create", "content.edit",
		}
	case "team_owner":
		return []string{
			"apps.read", "apps.write", "apps.delete",
			"functions.read", "functions.write", "functions.delete",
			"backends.read", "backends.write", "backends.delete",
			"deployments.read", "deployments.write", "deployments.delete",
			"teams.read", "teams.write", "teams.delete",
			"state.read", "state.write", "state.delete", "state.admin",
			"triggers.manage",
			"snapshots.create", "snapshots.restore",
			"replay.access",
			"memory.read", "memory.write",
			"registry.publish", "registry.verify",
			"monitoring.alerts", "monitoring.metrics", "monitoring.health",
			"security.scans", "security.audit",
			"content.create", "content.edit", "content.publish", "content.manage",
		}
	case "team_admin":
		return []string{
			"apps.read", "apps.write", "apps.delete",
			"functions.read", "functions.write", "functions.delete",
			"backends.read", "backends.write", "backends.delete",
			"deployments.read", "deployments.write", "deployments.delete",
			"teams.read", "teams.write",
			"state.read", "state.write", "state.delete",
			"triggers.manage",
			"snapshots.create", "snapshots.restore",
			"replay.access",
			"memory.read", "memory.write",
			"registry.publish", "registry.verify",
			"monitoring.alerts", "monitoring.metrics", "monitoring.health",
			"security.scans", "security.audit",
			"content.create", "content.edit", "content.publish",
		}
	case "team_member":
		return []string{
			"apps.read", "apps.write",
			"functions.read", "functions.write",
			"backends.read", "backends.write",
			"deployments.read", "deployments.write",
			"teams.read",
			"state.read", "state.write",
			"memory.read", "memory.write",
			"registry.publish",
			"monitoring.metrics", "monitoring.health",
			"security.scans",
			"content.create", "content.edit",
		}
	case "team_viewer":
		return []string{
			"apps.read",
			"functions.read",
			"backends.read",
			"deployments.read",
			"teams.read",
			"state.read",
			"memory.read",
			"monitoring.metrics", "monitoring.health",
			"security.audit",
			"content.create",
		}
	case "user":
		return []string{
			"apps.read", "apps.write",
			"functions.read", "functions.write",
			"backends.read", "backends.write",
			"deployments.read",
			"state.read", "state.write",
			"memory.read", "memory.write",
			"monitoring.health",
			"content.create",
		}
	default:
		return []string{}
	}
}

// registerDefaultHooks registers the default multi-tenancy hooks
func (a *Auth) registerDefaultHooks() {
	// Before sign-up hook: Extract and validate tenant
	a.hooks.Register("before:signup", func(ctx context.Context, req *HookRequest) error {
		tenantID := a.extractTenantID(req)
		if tenantID == uuid.Nil {
			return fmt.Errorf("tenant context is required")
		}

		// Validate tenant exists and is active
		var tenant Tenant
		if err := a.db.First(&tenant, "id = ? AND status = ?", tenantID, "active").Error; err != nil {
			return fmt.Errorf("invalid or inactive tenant: %w", err)
		}

		req.TenantID = tenantID
		return nil
	})

	// Before sign-in hook: Validate tenant context
	a.hooks.Register("before:signin", func(ctx context.Context, req *HookRequest) error {
		tenantID := a.extractTenantID(req)
		if tenantID == uuid.Nil {
			return fmt.Errorf("tenant context is required")
		}

		// Check IP allowlist if enabled
		if req.IPAddress != "" {
			var tenant Tenant
			if err := a.db.First(&tenant, "id = ?", tenantID).Error; err != nil {
				return fmt.Errorf("tenant not found: %w", err)
			}

			if tenant.IPAllowlistEnabled {
				allowed, err := a.checkIPAllowlist(tenantID, req.IPAddress)
				if err != nil {
					return fmt.Errorf("failed to check IP allowlist: %w", err)
				}
				if !allowed {
					return fmt.Errorf("access denied from this IP address")
				}
			}
		}

		req.TenantID = tenantID
		return nil
	})
}

// GenerateTokenWithRole generates a JWT token with role and permissions
func (a *Auth) GenerateTokenWithRole(userID, tenantID uuid.UUID, role string) (string, error) {
	if a.config.Secret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()
	permissions := a.getPermissionsForRole(role)

	claims := jwt.MapClaims{
		"user_id":     userID.String(),
		"email":       "", // Will be populated from user lookup if needed
		"tenant_id":   tenantID.String(),
		"role":        role,
		"permissions": permissions,
		"iss":         "functionfly-gba",
		"sub":         userID.String(),
		"exp":         now.Add(30 * time.Minute).Unix(),
		"iat":         now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.Secret))
}

// checkIPAllowlist checks if an IP is in the tenant's allowlist
func (a *Auth) checkIPAllowlist(tenantID uuid.UUID, ipAddress string) (bool, error) {
	var count int64
	err := a.db.Model(&TenantIPAllowlist{}).
		Where("tenant_id = ? AND ip_address = ? AND (expires_at IS NULL OR expires_at > ?)",
			tenantID, ipAddress, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 7 * 24 * time.Hour // Default 7 days
	}
	return d
}

// extractTenantID extracts tenant ID from the request context
// Priority: 1. X-Tenant-ID header, 2. Subdomain lookup (production), 3. Nil
func (a *Auth) extractTenantID(req *HookRequest) uuid.UUID {
	// Check header first
	if req.Headers["X-Tenant-ID"] != "" {
		if id, err := uuid.Parse(req.Headers["X-Tenant-ID"]); err == nil {
			return id
		}
	}

	// Check subdomain from host
	if req.Host != "" {
		parts := strings.Split(req.Host, ".")
		if len(parts) >= 3 {
			subdomain := parts[0]
			// Skip reserved subdomains
			reserved := []string{"www", "api", "auth", "admin", "app", "staging", "dev"}
			for _, r := range reserved {
				if strings.EqualFold(subdomain, r) {
					return uuid.Nil
				}
			}
			// Look up active tenant by subdomain
			var tenant Tenant
			if err := a.db.Where("subdomain = ? AND status = ?", subdomain, "active").First(&tenant).Error; err == nil {
				return tenant.ID
			}
		}
	}

	return uuid.Nil
}
