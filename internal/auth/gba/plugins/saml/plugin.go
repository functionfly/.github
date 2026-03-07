// Package saml provides SAML 2.0 SSO authentication support for GoBetterAuth
// This is Phase 4 of the Better Auth migration plan - Enterprise SSO
package saml

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SAMLPlugin provides SAML 2.0 SSO authentication functionality
type SAMLPlugin struct {
	db        *gorm.DB
	service   *SAMLService
	config    *SAMLPluginConfig
	logger    *logrus.Logger
	jwtSecret []byte
}

// SAMLPluginConfig holds SAML plugin configuration
type SAMLPluginConfig struct {
	Enabled          bool                  // Master switch for SAML
	EntityID         string                // SP Entity ID
	ACSURL           string                // SP Assertion Consumer Service URL
	AutoProvision    bool                  // Auto-provision users on first login
	SyncAttributes   bool                  // Sync user attributes from SAML on each login
	AttributeMapping *SAMLAttributeMapping // SAML attribute mapping
	JWTSecret        string                // JWT signing secret
	TokenExpiry      time.Duration         // Token expiry duration
}

// DefaultSAMLPluginConfig returns default SAML plugin configuration
func DefaultSAMLPluginConfig() *SAMLPluginConfig {
	return &SAMLPluginConfig{
		Enabled:          getEnvBool("GBA_SAML_ENABLED", false),
		EntityID:         getEnvOrDefault("SAML_SP_ENTITY_ID", "https://app.functionfly.com"),
		ACSURL:           getEnvOrDefault("SAML_ACS_URL", "https://app.functionfly.com/v1/auth/saml/acs"),
		AutoProvision:    true,
		SyncAttributes:   true,
		AttributeMapping: DefaultAttributeMapping(),
		JWTSecret:        getEnvOrDefault("JWT_SECRET", ""),
		TokenExpiry:      24 * time.Hour,
	}
}

// New creates a new SAML plugin instance
func New(db *gorm.DB, config *SAMLPluginConfig, logger *logrus.Logger) (*SAMLPlugin, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if config == nil {
		config = DefaultSAMLPluginConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Generate or use provided JWT secret
	jwtSecret := []byte(config.JWTSecret)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte(getEnvOrDefault("JWT_SECRET", "default-secret-change-in-production"))
		logger.Warn("Using default JWT secret - please set JWT_SECRET environment variable")
	}

	// Create SAML service
	service, err := NewSAMLService(db, logger, config.EntityID, config.ACSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML service: %w", err)
	}

	plugin := &SAMLPlugin{
		db:        db,
		service:   service,
		config:    config,
		logger:    logger,
		jwtSecret: jwtSecret,
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&SAMLConfig{}, &SAMLSession{}); err != nil {
		return nil, fmt.Errorf("failed to migrate SAML models: %w", err)
	}

	logger.Info("SAML plugin initialized")
	return plugin, nil
}

// IsEnabled returns true if SAML is enabled
func (p *SAMLPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// IsEnabledForTenant checks if SAML is enabled for a specific tenant
func (p *SAMLPlugin) IsEnabledForTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	if !p.IsEnabled() {
		return false, nil
	}

	config, err := p.service.GetConfig(ctx, tenantID)
	if err != nil {
		return false, nil // Not configured is not an error, just not enabled
	}

	return config.Enabled, nil
}

// GetStatus returns the SAML status for a tenant
func (p *SAMLPlugin) GetStatus(ctx context.Context, tenantID uuid.UUID) (*SAMLStatusResponse, error) {
	if !p.IsEnabled() {
		return &SAMLStatusResponse{
			Enabled:    false,
			Configured: false,
		}, nil
	}

	config, err := p.service.GetConfig(ctx, tenantID)
	if err != nil {
		return &SAMLStatusResponse{
			Enabled:    true,
			Configured: false,
		}, nil
	}

	return &SAMLStatusResponse{
		Enabled:     true,
		Configured:  true,
		IDPEntityID: config.IDPEntityID,
		SPEntityID:  config.SPEntityID,
	}, nil
}

// GetService returns the SAML service
func (p *SAMLPlugin) GetService() *SAMLService {
	return p.service
}

// GenerateToken generates a JWT token for a SAML-authenticated user
func (p *SAMLPlugin) GenerateToken(userID, tenantID uuid.UUID, authMethod string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         userID.String(),
		"tenant_id":   tenantID.String(),
		"auth_method": authMethod,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(p.config.TokenExpiry).Unix(),
		"type":        "saml",
	})

	tokenString, err := token.SignedString(p.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken verifies a JWT token
func (p *SAMLPlugin) VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// AuditLog logs a SAML-related audit event
func (p *SAMLPlugin) AuditLog(ctx context.Context, userID, tenantID uuid.UUID, action, result string, metadata map[string]interface{}) {
	// Create audit log entry
	auditEntry := map[string]interface{}{
		"id":         uuid.Must(uuid.NewRandom()),
		"user_id":    userID,
		"tenant_id":  tenantID,
		"action":     "saml_" + action,
		"result":     result,
		"metadata":   metadata,
		"created_at": time.Now(),
	}

	// Insert into audit log table (if exists)
	dbResult := p.db.WithContext(ctx).Table("gba_audit_logs").Create(auditEntry)
	if dbResult.Error != nil {
		p.logger.WithError(dbResult.Error).Warn("Failed to create audit log entry")
	}

	// Also log to application logs
	p.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
		"action":    action,
		"result":    result,
	}).Info("SAML audit event")
}

// CleanupExpiredSessions removes expired SAML sessions
func (p *SAMLPlugin) CleanupExpiredSessions(ctx context.Context) error {
	return p.service.CleanupExpiredSessions(ctx)
}

// GetHandler returns an HTTP handler for SAML endpoints
func (p *SAMLPlugin) GetHandler() *Handler {
	return NewHandler(p, p.logger)
}

// SetupRoutes registers SAML routes with the provided mux
func (p *SAMLPlugin) SetupRoutes(mux *http.ServeMux, basePath string) {
	handler := p.GetHandler()
	handler.SetupRoutes(mux, basePath)
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the boolean value of an environment variable or a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// SAMLClaims represents JWT claims for SAML authentication
type SAMLClaims struct {
	UserID     string `json:"sub"`
	TenantID   string `json:"tenant_id"`
	AuthMethod string `json:"auth_method"`
	Type       string `json:"type"`
	jwt.RegisteredClaims
}
