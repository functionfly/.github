package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	mfaSvc         *MFAService
}

// NewAuthService creates a new auth service
func NewAuthService(repo storage.Repository, jwtSecret string) *AuthService {
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

	service := &AuthService{
		repo:           repo,
		emailSvc:       email.NewMockService(emailConfig), // Default to mock service with config
		jwtSecret:      []byte(jwtSecret),
		jwtDuration:    30 * time.Minute, // 30 minutes - short lived access tokens
		oauthProviders: make(map[string]*OAuthProvider),
		baseURL:        "http://localhost:8080", // Default, can be overridden
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