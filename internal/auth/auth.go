package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

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
func NewAuthService(repo storage.Repository, jwtSecret string) *AuthService {
	// Validate JWT secret minimum length (32 bytes = 256 bits for HS256)
	if len(jwtSecret) < 32 {
		panic(fmt.Sprintf("JWT_SECRET must be at least 32 bytes (256 bits) for HS256 security, got %d bytes. Generate a secure secret with: openssl rand -hex 32", len(jwtSecret)))
	}

	// Default email config for testing/development
	emailConfig := email.Config{
		SMTPHost:     "localhost",
		SMTPPort:     1025, // Mailpit default port
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "noreply@functionfly.com",
		FromName:     "FunctionFly",
		BaseURL:      "http://localhost:8080",
	}

	// In production, require Resend or SMTP (mock is not allowed)
	var emailSvc email.Service
	if os.Getenv("PRODUCTION_ENV") == "true" {
		svc, ok := email.NewServiceFromEnv()
		if !ok {
			panic("PRODUCTION_ENV=true requires RESEND_API_KEY or SMTP_HOST. Mock email service is not allowed in production.")
		}
		emailSvc = svc
	} else {
		emailSvc = email.NewMockService(emailConfig)
	}

	service := &AuthService{
		repo:           repo,
		emailSvc:       emailSvc,
		jwtSecret:      []byte(jwtSecret),
		jwtDuration:    30 * time.Minute, // 30 minutes - short lived access tokens
		oauthProviders: make(map[string]*OAuthProvider),
		baseURL:        "http://localhost:8080", // Default, can be overridden
		authURL:        "http://localhost:4321", // Auth frontend URL for waitlist/emails
	}

	// Initialize MFA service
	service.mfaSvc = NewMFAService(repo)

	// Initialize OAuth providers (will be configured with environment variables)
	service.initOAuthProviders()

	return service
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
func (a *AuthService) SetupMFA(req MFASetupRequest) (*MFASetupResponse, error) {
	return a.mfaSvc.SetupMFA(req)
}

func (a *AuthService) VerifyMFA(req MFAVerifyRequest) (*MFAVerifyResponse, error) {
	return a.mfaSvc.VerifyMFA(req)
}

func (a *AuthService) EnableMFA(userID uuid.UUID) error {
	return a.mfaSvc.EnableMFA(userID)
}

func (a *AuthService) DisableMFA(req MFADisableRequest) error {
	return a.mfaSvc.DisableMFA(req)
}

func (a *AuthService) GetMFAStatus(userID uuid.UUID) (*MFAStatus, error) {
	return a.mfaSvc.GetMFAStatus(userID)
}

func (a *AuthService) IsMFARequired(userID uuid.UUID) (bool, error) {
	return a.mfaSvc.IsMFARequired(userID)
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
