package iot

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JWTValidator struct {
	db     *gorm.DB
	secret []byte
	issuer string
	mu     sync.RWMutex
}

type JWTValidatorConfig struct {
	JWTSecret string
	JWTIssuer string
	JWTExpiry time.Duration
}

type DeviceTokenClaims struct {
	jwt.RegisteredClaims
	DeviceID   uuid.UUID  `json:"device_id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	DeviceType DeviceType `json:"device_type"`
	Name       string     `json:"name"`
	GatewayID  *uuid.UUID `json:"gateway_id,omitempty"`
}

func NewJWTValidator(db *gorm.DB, config *JWTValidatorConfig) (*JWTValidator, error) {
	secret := []byte(config.JWTSecret)
	if len(secret) == 0 {
		secret = []byte(getEnvOrDefault("IOT_JWT_SECRET", "change-me-in-production"))
	}

	issuer := config.JWTIssuer
	if issuer == "" {
		issuer = getEnvOrDefault("IOT_JWT_ISSUER", "functionfly-iot")
	}

	return &JWTValidator{
		db:     db,
		secret: secret,
		issuer: issuer,
	}, nil
}

func (v *JWTValidator) ValidateToken(tokenString string) (*DeviceClaims, error) {
	v.mu.RLock()
	secret := v.secret
	issuer := v.issuer
	v.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &DeviceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*DeviceTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.Issuer != issuer {
		return nil, fmt.Errorf("invalid token issuer: %s", claims.Issuer)
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	var device Device
	if err := v.db.Where("id = ? AND tenant_id = ?", claims.DeviceID, claims.TenantID).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not registered")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &DeviceClaims{
		DeviceID:    claims.DeviceID,
		TenantID:    claims.TenantID,
		DeviceType:  claims.DeviceType,
		Name:        claims.Name,
		AuthMethod:  AuthMethodJWT,
		Permissions: []string{"iot.telemetry.publish"},
	}, nil
}

func (v *JWTValidator) IssueToken(deviceID, tenantID uuid.UUID, deviceType DeviceType, name string) (string, error) {
	v.mu.RLock()
	secret := v.secret
	issuer := v.issuer
	v.mu.RUnlock()

	now := time.Now()
	claims := DeviceTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   deviceID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
		DeviceID:   deviceID,
		TenantID:   tenantID,
		DeviceType: deviceType,
		Name:       name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (v *JWTValidator) RefreshToken(tokenString string) (string, error) {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}

	return v.IssueToken(claims.DeviceID, claims.TenantID, claims.DeviceType, claims.Name)
}

func (v *JWTValidator) UpdateSecret(newSecret string) error {
	if len(newSecret) < 32 {
		return fmt.Errorf("secret must be at least 32 bytes")
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.secret = []byte(newSecret)
	return nil
}

type JWTValidatorSync struct {
	validator *JWTValidator
}

func NewJWTValidatorSync(db *gorm.DB, config *JWTValidatorConfig) (*JWTValidatorSync, error) {
	validator, err := NewJWTValidator(db, config)
	if err != nil {
		return nil, err
	}

	return &JWTValidatorSync{validator: validator}, nil
}

func (s *JWTValidatorSync) ValidateDeviceToken(tokenString string) (*DeviceClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	return s.validator.ValidateToken(tokenString)
}
