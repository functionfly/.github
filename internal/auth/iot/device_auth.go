package iot

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DeviceAuthConfig struct {
	Enabled         bool
	CACertDir       string
	CACertPaths     []string
	JWTSecret       string
	JWTIssuer       string
	JWTExpiry       time.Duration
	SkipCertVerify  bool
	RateLimitWindow time.Duration
	RateLimitMax    int
}

func DefaultDeviceAuthConfig() *DeviceAuthConfig {
	return &DeviceAuthConfig{
		Enabled:         false,
		RateLimitWindow: 60 * time.Second,
		RateLimitMax:    10,
		JWTIssuer:      "functionfly-iot",
		JWTExpiry:      24 * time.Hour,
	}
}

type DeviceAuthConfigFromEnvFunc func(db *gorm.DB) (*DeviceAuthConfig, error)

func ConfigFromEnv(db *gorm.DB) (*DeviceAuthConfig, error) {
	cfg := &DeviceAuthConfig{
		Enabled:        getEnvBool("IOT_AUTH_ENABLED", false),
		CACertDir:      os.Getenv("IOT_CA_CERT_DIR"),
		JWTSecret:      os.Getenv("IOT_JWT_SECRET"),
		JWTIssuer:      getEnvOrDefault("IOT_JWT_ISSUER", "functionfly-iot"),
		SkipCertVerify: getEnvBool("IOT_SKIP_CERT_VERIFY", false),
		RateLimitMax:   getEnvInt("IOT_RATE_LIMIT_MAX", 10),
	}

	if paths := os.Getenv("IOT_CA_CERT_PATHS"); paths != "" {
		cfg.CACertPaths = splitPaths(paths)
	}

	return cfg, nil
}

type DeviceAuth struct {
	db           *gorm.DB
	config       *DeviceAuthConfig
	certValidator *CertValidator
	jwtValidator  *JWTValidator
	pskStore      *PSKStore
	logger       *logrus.Logger
	rateLimit    *RateLimiter
	mu           sync.RWMutex
}

func New(db *gorm.DB, config *DeviceAuthConfig, logger *logrus.Logger) (*DeviceAuth, error) {
	if config == nil {
		config = DefaultDeviceAuthConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	auth := &DeviceAuth{
		db:        db,
		config:    config,
		logger:    logger,
		rateLimit: NewRateLimiter(config.RateLimitMax, config.RateLimitWindow),
	}

	if config.Enabled {
		if err := auth.initializeValidators(); err != nil {
			return nil, fmt.Errorf("failed to initialize validators: %w", err)
		}
	}

	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return auth, nil
}

func (a *DeviceAuth) initializeValidators() error {
	certValidator, err := NewCertValidator(a.db, &CertValidatorConfig{
		CACertDir:  a.config.CACertDir,
		CACertPaths: a.config.CACertPaths,
		SkipVerify: a.config.SkipCertVerify,
	})
	if err != nil {
		return fmt.Errorf("failed to create cert validator: %w", err)
	}
	a.certValidator = certValidator

	jwtValidator, err := NewJWTValidator(a.db, &JWTValidatorConfig{
		JWTSecret: a.config.JWTSecret,
		JWTIssuer: a.config.JWTIssuer,
		JWTExpiry: a.config.JWTExpiry,
	})
	if err != nil {
		return fmt.Errorf("failed to create JWT validator: %w", err)
	}
	a.jwtValidator = jwtValidator

	a.pskStore = NewPSKStore(a.db, &PSKStoreConfig{HashPSK: true})

	return nil
}

func (a *DeviceAuth) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.Enabled
}

func (a *DeviceAuth) AuthenticateDevice(ctx context.Context, method AuthMethod, credential interface{}) (*DeviceClaims, error) {
	if !a.IsEnabled() {
		return nil, fmt.Errorf("IoT auth is disabled")
	}

	deviceIdentifier := getDeviceIdentifier(credential)
	if !a.rateLimit.Allow(deviceIdentifier) {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	switch method {
	case AuthMethodX509:
		if connState, ok := credential.(*tls.ConnectionState); ok {
			return a.certValidator.ValidateConnection(*connState)
		}
		return nil, fmt.Errorf("invalid credential type for X509")

	case AuthMethodJWT:
		if tokenString, ok := credential.(string); ok {
			return a.jwtValidator.ValidateToken(tokenString)
		}
		return nil, fmt.Errorf("invalid credential type for JWT")

	case AuthMethodPSK:
		if pskCred, ok := credential.(*PSKCredential); ok {
			return a.pskStore.ValidatePSK(pskCred.DeviceID, pskCred.PSK)
		}
		return nil, fmt.Errorf("invalid credential type for PSK")

	default:
		return nil, fmt.Errorf("unsupported auth method: %s", method)
	}
}

func (a *DeviceAuth) RegisterDevice(ctx context.Context, tenantID uuid.UUID, name string, deviceType DeviceType, authMethod AuthMethod, credential interface{}) (*Device, error) {
	if !a.IsEnabled() {
		return nil, fmt.Errorf("IoT auth is disabled")
	}

	switch authMethod {
	case AuthMethodX509:
		if cert, ok := credential.(*Certificate); ok {
			return a.certValidator.RegisterDevice(a.db, tenantID, name, deviceType, cert.Certificate)
		}
		return nil, fmt.Errorf("invalid certificate")

	case AuthMethodJWT:
		device := &Device{
			TenantID:   tenantID,
			Name:       name,
			DeviceType: deviceType,
			AuthMethod: AuthMethodJWT,
			Status:     DeviceStatusOffline,
		}
		if err := a.db.Create(device).Error; err != nil {
			return nil, fmt.Errorf("failed to register device: %w", err)
		}
		state := &DeviceState{DeviceID: device.ID, State: map[string]any{}}
		if err := a.db.Create(state).Error; err != nil {
			return nil, fmt.Errorf("failed to create state: %w", err)
		}
		return device, nil

	case AuthMethodPSK:
		if psk, ok := credential.(string); ok {
			return a.pskStore.RegisterDevicePSK(a.db, tenantID, name, deviceType, psk)
		}
		return nil, fmt.Errorf("invalid PSK")

	default:
		return nil, fmt.Errorf("unsupported auth method: %s", authMethod)
	}
}

func (a *DeviceAuth) IssueDeviceToken(ctx context.Context, deviceID uuid.UUID) (string, error) {
	var device Device
	if err := a.db.Where("id = ?", deviceID).First(&device).Error; err != nil {
		return "", fmt.Errorf("device not found: %w", err)
	}

	return a.jwtValidator.IssueToken(device.ID, device.TenantID, device.DeviceType, device.Name)
}

func (a *DeviceAuth) GetDevice(ctx context.Context, deviceID uuid.UUID) (*Device, error) {
	var device Device
	if err := a.db.Where("id = ?", deviceID).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

func (a *DeviceAuth) ListDevices(ctx context.Context, tenantID uuid.UUID, deviceType *DeviceType) ([]Device, error) {
	query := a.db.Where("tenant_id = ?", tenantID)
	if deviceType != nil {
		query = query.Where("device_type = ?", *deviceType)
	}

	var devices []Device
	if err := query.Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	return devices, nil
}

func (a *DeviceAuth) UpdateDeviceStatus(ctx context.Context, deviceID uuid.UUID, status DeviceStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == DeviceStatusOnline {
		now := time.Now()
		updates["last_seen"] = &now
	}
	return a.db.Model(&Device{}).Where("id = ?", deviceID).Updates(updates).Error
}

func (a *DeviceAuth) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	return a.db.Where("id = ?", deviceID).Delete(&Device{}).Error
}

func (a *DeviceAuth) GetDeviceState(ctx context.Context, deviceID uuid.UUID) (*DeviceState, error) {
	var state DeviceState
	if err := a.db.Where("device_id = ?", deviceID).First(&state).Error; err != nil {
		return nil, fmt.Errorf("device state not found: %w", err)
	}
	return &state, nil
}

func (a *DeviceAuth) UpdateDeviceState(ctx context.Context, deviceID uuid.UUID, state map[string]any) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	updates := map[string]interface{}{
		"state":      string(stateJSON),
		"updated_at": time.Now(),
	}
	return a.db.Model(&DeviceState{}).Where("device_id = ?", deviceID).Updates(updates).Error
}

func (a *DeviceAuth) CreateCommand(ctx context.Context, deviceID uuid.UUID, command map[string]any) (*DeviceCommand, error) {
	cmd := &DeviceCommand{
		DeviceID: deviceID,
		Command:  command,
		Status:   "pending",
	}
	if err := a.db.Create(cmd).Error; err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}
	return cmd, nil
}

func (a *DeviceAuth) AcknowledgeCommand(ctx context.Context, commandID uuid.UUID) error {
	now := time.Now()
	return a.db.Model(&DeviceCommand{}).Where("id = ?", commandID).Updates(map[string]interface{}{
		"status":          "acknowledged",
		"acknowledged_at": &now,
	}).Error
}

func (a *DeviceAuth) ListPendingCommands(ctx context.Context, deviceID uuid.UUID) ([]DeviceCommand, error) {
	var commands []DeviceCommand
	if err := a.db.Where("device_id = ? AND status = ?", deviceID, "pending").
		Order("created_at ASC").
		Find(&commands).Error; err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}
	return commands, nil
}

type PSKCredential struct {
	DeviceID uuid.UUID
	PSK      string
}

type Certificate struct {
	*x509.Certificate
}

func getDeviceIdentifier(credential interface{}) string {
	switch c := credential.(type) {
	case *tls.ConnectionState:
		if len(c.VerifiedChains) > 0 {
			return c.VerifiedChains[0][0].Subject.CommonName
		}
	case string:
		return c
	case *PSKCredential:
		return c.DeviceID.String()
	}
	return "unknown"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	switch os.Getenv(key) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvInt(key string, defaultValue int) int {
	var result int
	if _, err := fmt.Sscanf(os.Getenv(key), "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

func splitPaths(paths string) []string {
	var result []string
	for _, p := range splitString(paths, ",") {
		p = trimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for i := 0; i < len(s); {
		idx := indexString(s, sep, i)
		if idx < 0 {
			result = append(result, s[i:])
			break
		}
		result = append(result, s[i:idx])
		i = idx + len(sep)
	}
	return result
}

func indexString(s, substr string, start int) int {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
