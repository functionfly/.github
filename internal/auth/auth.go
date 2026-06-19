package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// WelcomeNotifier is implemented by the notification service to send welcome notifications to new users.
type WelcomeNotifier interface {
	SendWelcome(ctx context.Context, userID uuid.UUID) error
}

// AuthService handles authentication operations
type AuthService struct {
	repo           storage.Repository
	emailSvc       email.Service
	notifySvc      WelcomeNotifier // optional; when set, welcome notification is sent on signup/OAuth signup
	jwtSecret      []byte
	jwtDuration    time.Duration
	oauthProviders map[string]*OAuthProvider
	baseURL        string
	authURL        string
	mfaSvc         *MFAService
}

// NewAuthService creates a new auth service
// JWT secret must be at least 32 bytes (256 bits) for HS256 security
func NewAuthService(repo storage.Repository, jwtSecret string) (*AuthService, error) {
	// Validate JWT secret minimum length (32 bytes = 256 bits for HS256)
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes (256 bits) for HS256 security, got %d bytes. Generate a secure secret with: openssl rand -hex 32", len(jwtSecret))
	}

	// Validate JWT secret strength to prevent weak secrets
	if err := validateJWTSecretStrength(jwtSecret); err != nil {
		return nil, err
	}

	// Default email config for testing/development
	emailConfig := email.Config{
		SMTPHost:     "localhost",
		SMTPPort:     1025, // Mailpit default port
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "noreply@functionfly.com",
		FromName:     "FunctionFly",
		BaseURL:      getEnvOrDefault("BASE_URL", "https://api.functionfly.com"),
	}

	// In production, require Resend or SMTP (mock is not allowed)
	var emailSvc email.Service
	if os.Getenv("PRODUCTION_ENV") == "true" {
		svc, ok := email.NewServiceFromEnv()
		if !ok {
			return nil, fmt.Errorf("PRODUCTION_ENV=true requires RESEND_API_KEY or SMTP_HOST. Mock email service is not allowed in production")
		}
		emailSvc = svc
	} else {
		emailSvc = email.NewMockService(emailConfig)
	}

	jwtDuration := 4 * time.Hour
	if d, err := time.ParseDuration(getEnvOrDefault("JWT_EXPIRATION", "4h")); err == nil {
		jwtDuration = d
	}
	service := &AuthService{
		repo:           repo,
		emailSvc:       emailSvc,
		jwtSecret:      []byte(jwtSecret),
		jwtDuration:    jwtDuration,
		oauthProviders: make(map[string]*OAuthProvider),
		baseURL:        getEnvOrDefault("BASE_URL", "https://api.functionfly.com"),
		authURL:        getEnvOrDefault("AUTH_FRONTEND_URL", "https://auth.functionfly.com"),
	}

	// Initialize MFA service (Redis optional - rate limiting will be no-op without it)
	service.mfaSvc = NewMFAService(repo, nil)

	// Initialize OAuth providers (will be configured with environment variables)
	service.initOAuthProviders()

	return service, nil
}

// GenerateToken creates a JWT token for a user (for debugging)
func (a *AuthService) GenerateToken(user *storage.User) (string, error) {
	return a.generateToken(user)
}

// GenerateRefreshToken creates a new refresh token (for public API)
func (a *AuthService) GenerateRefreshToken() (token, hash string, err error) {
	return a.generateRefreshToken()
}

// SetEmailService sets the email service for the auth service
func (a *AuthService) SetEmailService(emailSvc email.Service) {
	a.emailSvc = emailSvc
}

// SetNotificationService sets the optional notifier for welcome notifications (e.g. notification.Service).
func (a *AuthService) SetNotificationService(n WelcomeNotifier) {
	a.notifySvc = n
}

// Repo returns the repository interface
func (a *AuthService) Repo() storage.Repository {
	return a.repo
}

// EmailService returns the email service
func (a *AuthService) EmailService() email.Service {
	return a.emailSvc
}

// AuthURL returns the auth frontend base URL used for email links.
func (a *AuthService) AuthURL() string {
	return a.authURL
}

// SetAuthURL sets the auth frontend base URL.
func (a *AuthService) SetAuthURL(url string) {
	a.authURL = url
}

// MFA methods
func (a *AuthService) SetupMFA(ctx context.Context, req MFASetupRequest) (*MFASetupResponse, error) {
	return a.mfaSvc.SetupMFA(ctx, req)
}

func (a *AuthService) VerifyMFA(ctx context.Context, req MFAVerifyRequest) (*MFAVerifyResponse, error) {
	return a.mfaSvc.VerifyMFA(ctx, req)
}

func (a *AuthService) EnableMFA(ctx context.Context, userID uuid.UUID) error {
	return a.mfaSvc.EnableMFA(ctx, userID)
}

func (a *AuthService) DisableMFA(ctx context.Context, req MFADisableRequest) error {
	return a.mfaSvc.DisableMFA(ctx, req)
}

func (a *AuthService) GetMFAStatus(ctx context.Context, userID uuid.UUID) (*MFAStatus, error) {
	return a.mfaSvc.GetMFAStatus(ctx, userID)
}

func (a *AuthService) IsMFARequired(ctx context.Context, userID uuid.UUID) (bool, error) {
	return a.mfaSvc.IsMFARequired(ctx, userID)
}

// GetTenantOAuthProviders returns the list of enabled OAuth provider names for a tenant
func (a *AuthService) GetTenantOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	providers, err := a.repo.GetEnabledOAuthProviders(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Provider
	}
	return names, nil
}

// GenerateInviteToken generates a cryptographically secure token for team invitations.
// Uses 32 bytes of crypto/rand (256 bits) — significantly stronger than a UUID v4.
func GenerateInviteToken() (string, time.Time) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fall back to UUID if crypto/rand is unavailable (should never happen in practice)
		return uuid.New().String(), time.Now().Add(7 * 24 * time.Hour)
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	return token, expiresAt
}

// validateJWTSecretStrength checks that the JWT secret has sufficient entropy
// and is not a known weak value
func validateJWTSecretStrength(secret string) error {
	// Check for common weak patterns
	weakPatterns := []string{
		"12345678901234567890123456789012", // Sequential numeric
		"password", "admin", "secret",      // Common words
		"abcdefghijklmnopqrstuvwxyz12",    // Sequential alpha
	}

	secretLower := strings.ToLower(secret)
	for _, pattern := range weakPatterns {
		if secretLower == strings.ToLower(pattern) || strings.Contains(secretLower, pattern) {
			// Allow if it's long enough and contains mixed characters
			if len(secret) >= 64 {
				break // Long enough random strings are ok even if they contain these chars
			}
			return fmt.Errorf("JWT_SECRET appears to be a weak or known secret. Please generate a secure secret with: openssl rand -hex 32")
		}
	}

	// Check for low entropy (all same character, sequential, etc.)
	if isLowEntropy(secret) {
		return fmt.Errorf("JWT_SECRET has low entropy and appears to be a weak secret. Please generate a secure secret with: openssl rand -hex 32")
	}

	return nil
}

// isLowEntropy checks if a string has low entropy (all same, sequential, etc.)
func isLowEntropy(s string) bool {
	if len(s) < 32 {
		return true
	}

	// Check if all characters are the same
	if len(s) > 0 {
		first := s[0]
		allSame := true
		for i := 1; i < len(s); i++ {
			if s[i] != first {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	}

	// Check for mostly sequential patterns (simple check)
	sequentialCount := 0
	for i := 1; i < len(s); i++ {
		if int(s[i])-int(s[i-1]) == 1 || int(s[i])-int(s[i-1]) == -1 {
			sequentialCount++
		}
	}
	if sequentialCount > len(s)/3 {
		return true // More than 1/3 sequential is suspicious
	}

	// Check character distribution - too uniform is suspicious
	seen := make(map[rune]int)
	for _, c := range s {
		seen[c]++
	}
	// If more than 50% of characters are the same character, it's low entropy
	for _, count := range seen {
		if float64(count)/float64(len(s)) > 0.5 {
			return true
		}
	}

	return false
}
