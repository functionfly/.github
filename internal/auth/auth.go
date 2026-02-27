package auth

import (
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// AuthService handles authentication operations
type AuthService struct {
	repo           storage.Repository
	emailSvc       email.Service
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
		FromEmail:    "noreply@functionfly.dev",
		FromName:     "FunctionFly",
		BaseURL:      "http://localhost:8080",
	}

	service := &AuthService{
		repo:           repo,
		emailSvc:       email.NewMockService(emailConfig), // Default to mock service with config
		jwtSecret:      []byte(jwtSecret),
		jwtDuration:    24 * time.Hour, // 24 hours
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

// SetEmailService sets the email service for the auth service
func (a *AuthService) SetEmailService(emailSvc email.Service) {
	a.emailSvc = emailSvc
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

// GenerateInviteToken generates a secure token for team invitations
func GenerateInviteToken() (string, time.Time) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	return token, expiresAt
}