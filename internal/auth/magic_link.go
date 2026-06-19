package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MagicLinkRequest represents a request to send a magic link
// @Description Request body for magic link authentication
type MagicLinkRequest struct {
	Email        string `json:"email" example:"user@example.com" validate:"required,email"`
	RedirectPath string `json:"redirect_path,omitempty" example:"/dashboard"`
}

// MagicLinkResponse represents the response after requesting a magic link
// @Description Response confirming magic link has been sent
type MagicLinkResponse struct {
	Message   string `json:"message" example:"Magic link sent to your email"`
	EmailSent bool   `json:"email_sent" example:"true"`
}

// MagicLinkVerifyRequest represents a request to verify a magic link token
// @Description Request body for verifying a magic link token
type MagicLinkVerifyRequest struct {
	Token     string `json:"token" validate:"required"`
	IPAddress string `json:"-"` // Set from request context
	UserAgent string `json:"-"` // Set from request context
}

// MagicLinkVerifyResponse represents the response after verifying a magic link
// @Description Response containing authentication tokens and user info after successful magic link verification
type MagicLinkVerifyResponse struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         *LoginUser `json:"user"`
	NewUser      bool       `json:"new_user"`
}

// MagicLinkConfig holds configuration for magic link authentication
type MagicLinkConfig struct {
	// TokenExpiry is the duration for which a magic link is valid (default: 15 minutes)
	TokenExpiry time.Duration
	// MaxAttempts is the maximum number of magic links that can be requested per email per hour (default: 5)
	MaxAttempts int
	// AllowSignup controls whether new users can be created via magic link (default: true)
	AllowSignup bool
}

// DefaultMagicLinkConfig returns the default magic link configuration
// Values can be overridden with environment variables:
//   - MAGIC_LINK_TOKEN_EXPIRY_MINUTES: Token expiry in minutes (default: 15)
//   - MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR: Max attempts per email per hour (default: 5)
//   - MAGIC_LINK_ALLOW_SIGNUP: Allow new user signup via magic link (default: true)
func DefaultMagicLinkConfig() *MagicLinkConfig {
	config := &MagicLinkConfig{
		TokenExpiry: 15 * time.Minute,
		MaxAttempts: 5,
		AllowSignup: true,
	}

	// Override with environment variables if set
	if expiryMinutes := os.Getenv("MAGIC_LINK_TOKEN_EXPIRY_MINUTES"); expiryMinutes != "" {
		if minutes, err := strconv.Atoi(expiryMinutes); err == nil && minutes > 0 {
			config.TokenExpiry = time.Duration(minutes) * time.Minute
		}
	}

	if maxAttempts := os.Getenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR"); maxAttempts != "" {
		if attempts, err := strconv.Atoi(maxAttempts); err == nil && attempts > 0 {
			config.MaxAttempts = attempts
		}
	}

	if allowSignup := os.Getenv("MAGIC_LINK_ALLOW_SIGNUP"); allowSignup != "" {
		// Parse as bool - accepts "true", "false", "1", "0"
		config.AllowSignup = allowSignup == "true" || allowSignup == "1"
	}

	return config
}

// generateMagicLinkToken generates a cryptographically secure magic link token
func generateMagicLinkToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashMagicLinkToken creates an HMAC-SHA256 hash of a magic link token for secure storage
// Uses a secret key to prevent rainbow table attacks from database breaches
func hashMagicLinkToken(token string) (string, error) {
	secretKey := os.Getenv("MAGIC_LINK_TOKEN_SECRET")
	if secretKey == "" {
		return "", fmt.Errorf("MAGIC_LINK_TOKEN_SECRET environment variable is required")
	}
	if err := validateMagicLinkSecretStrength(secretKey); err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// validateMagicLinkSecretStrength checks that the magic link secret has sufficient entropy
// and is not a known weak value
func validateMagicLinkSecretStrength(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("MAGIC_LINK_TOKEN_SECRET must be at least 32 bytes; got %d bytes. Generate a secure secret with: openssl rand -hex 32", len(secret))
	}

	weakPatterns := []string{
		"password", "admin", "secret", "changeme",
	}

	secretLower := strings.ToLower(secret)
	for _, pattern := range weakPatterns {
		if strings.Contains(secretLower, pattern) && len(secret) < 64 {
			return fmt.Errorf("MAGIC_LINK_TOKEN_SECRET appears to be a weak secret. Please generate a secure secret with: openssl rand -hex 32")
		}
	}

	return nil
}

// normalizeEmail normalizes an email address for comparison
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// RequestMagicLink generates and sends a magic link to the user's email
// Returns nil error even if email doesn't exist to avoid account enumeration
func (a *AuthService) RequestMagicLink(req MagicLinkRequest, ipAddress, userAgent string) (*MagicLinkResponse, error) {
	config := DefaultMagicLinkConfig()

	// Normalize email
	email := normalizeEmail(req.Email)

	// Check rate limiting - count recent magic links for this email
	ctx := context.Background()
	recentLinks, err := a.repo.GetRecentMagicLinksByEmail(ctx, email, time.Now().Add(-1*time.Hour))
	if err != nil {
		logrus.WithError(err).WithField("email", email).Warn("Failed to get recent magic links")
		// Continue anyway - don't block on DB errors
	} else if len(recentLinks) >= config.MaxAttempts {
		// Too many attempts - but don't reveal this to the user
		logrus.WithField("email", email).WithField("attempts", len(recentLinks)).Warn("Magic link rate limit exceeded")
		return &MagicLinkResponse{
			Message:   "If that email is registered, a magic link has been sent.",
			EmailSent: false,
		}, nil
	}

	// Look up user by email
	user, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Warn("Failed to lookup user for magic link")
		// Don't reveal error - return success message
		return &MagicLinkResponse{
			Message:   "If that email is registered, a magic link has been sent.",
			EmailSent: false,
		}, nil
	}

	var userID *uuid.UUID
	if user != nil {
		userID = &user.ID
	}

	// Generate magic link token
	token, err := generateMagicLinkToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate magic link token: %w", err)
	}

	// Hash the token for secure database storage (prevents token theft from DB breaches)
	tokenHash, hashErr := hashMagicLinkToken(token)
	if hashErr != nil {
		logrus.Error(hashErr)
		return nil, fmt.Errorf("failed to configure magic link security: %w", hashErr)
	}

	// Calculate expiry
	expiresAt := time.Now().Add(config.TokenExpiry)

	// Create magic link record (store hash, not plaintext token)
	_, err = a.repo.CreateMagicLink(ctx, email, tokenHash, userID, ipAddress, userAgent, req.RedirectPath, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create magic link: %w", err)
	}

	// Send magic link email
	logrus.WithField("email", email).
		WithField("has_user", user != nil).
		WithField("allow_signup", config.AllowSignup).
		Info("Attempting to send magic link email")

	var sendErr error
	if user != nil {
		// Existing user - send magic link
		sendErr = a.emailSvc.SendMagicLinkEmail(user, token, config.TokenExpiry)
	} else if config.AllowSignup {
		// New user allowed - send signup magic link
		sendErr = a.emailSvc.SendMagicLinkSignupEmail(email, token, config.TokenExpiry)
	}

	if sendErr != nil {
		logrus.WithError(sendErr).
			WithField("email", email).
			Error("SendMagicLinkEmail returned error")
	}

	if sendErr != nil {
		// Log at ERROR level to ensure visibility in production logs
		logrus.WithError(sendErr).
			WithField("email", email).
			WithField("has_account", user != nil).
			WithField("allow_signup", config.AllowSignup).
			Error("Magic link email failed to send - check RESEND_API_KEY, FROM_EMAIL, and AUTH_URL configuration")
		// Don't fail the request to prevent account enumeration
		// but EmailSent will be false
	} else {
		logrus.WithField("email", email).
			WithField("has_account", user != nil).
			Info("Magic link email sent successfully")
	}

	return &MagicLinkResponse{
		Message:   "If that email is registered, a magic link has been sent.",
		EmailSent: sendErr == nil,
	}, nil
}

// VerifyMagicLink validates a magic link token and creates a user session
func (a *AuthService) VerifyMagicLink(req MagicLinkVerifyRequest) (*MagicLinkVerifyResponse, error) {
	ctx := context.Background()

	// Debug logging (always log at info for troubleshooting)
	tokenPreview := req.Token
	if len(tokenPreview) > 8 {
		tokenPreview = tokenPreview[:8] + "..."
	}
	logrus.WithField("token_preview", tokenPreview).
		WithField("token_len", len(req.Token)).
		Info("Magic link verification request received")

	// Hash the provided token to look up in database (tokens are stored as hashes for security)
	tokenHash, hashErr := hashMagicLinkToken(req.Token)
	if hashErr != nil {
		logrus.Error(hashErr)
		return nil, fmt.Errorf("magic link security not configured")
	}

	// Get magic link from database by hash
	magicLink, err := a.repo.GetMagicLinkByToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve magic link: %w", err)
	}

	if magicLink == nil {
		// Check if any magic links exist for debugging
		recent, _ := a.repo.GetRecentMagicLinksByEmail(ctx, "", time.Now().Add(-24*time.Hour))
		logrus.WithField("recent_count", len(recent)).
			WithField("token_preview", tokenPreview).
			WithField("token_full_len", len(req.Token)).
			Warn("Magic link not found - token mismatch or not in DB")
		return nil, fmt.Errorf("invalid magic link token")
	}

	logrus.WithField("token_preview", tokenPreview).
		WithField("email", magicLink.Email).
		WithField("used", magicLink.Used).
		WithField("expired", magicLink.IsExpired()).
		Info("Magic link found in database")

	// Check if already used
	if magicLink.Used {
		logrus.WithField("token_preview", tokenPreview).WithField("used_at", magicLink.UsedAt).Warn("Magic link already used")
		return nil, fmt.Errorf("magic link has already been used")
	}

	// Check if expired
	if magicLink.IsExpired() {
		logrus.WithField("token_preview", tokenPreview).
			WithField("expires_at", magicLink.ExpiresAt).
			WithField("now", time.Now()).
			Warn("Magic link expired")
		return nil, fmt.Errorf("magic link has expired")
	}

	// Mark as used
	if err := a.repo.MarkMagicLinkUsed(ctx, magicLink.ID); err != nil {
		return nil, fmt.Errorf("failed to mark magic link as used: %w", err)
	}

	var user *storage.User
	isNewUser := false

	if magicLink.UserID != nil {
		// Existing user
		user, err = a.repo.GetUserByID(ctx, *magicLink.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("user not found")
		}
	} else {
		// New user - create account
		config := DefaultMagicLinkConfig()
		if !config.AllowSignup {
			return nil, fmt.Errorf("signup via magic link is not enabled")
		}

		// Create a new tenant
		tenant, err := a.repo.CreateTenant(ctx, "Default Tenant")
		if err != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		// Generate a random password (user won't use it, they'll use magic links or set password later)
		passwordBytes := make([]byte, 32)
		rand.Read(passwordBytes)
		hashedPassword, _ := a.HashPassword(hex.EncodeToString(passwordBytes))

		// Create user with verified email
		user, err = a.repo.CreateUser(ctx, magicLink.Email, hashedPassword, tenant.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Mark email as verified
		updates := map[string]interface{}{
			"email_verified": true,
		}
		user, err = a.repo.UpdateUser(ctx, user.ID, updates)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		user.EmailVerified = true
		isNewUser = true

		// Send welcome notification
		a.sendWelcomeNotification(ctx, user.ID)
	}

	// Generate JWT token
	token, err := a.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, refreshTokenHash, err := a.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = a.repo.CreateRefreshToken(ctx, user.ID, refreshTokenHash, req.IPAddress, req.UserAgent, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Log successful magic link authentication
	authEvent := &storage.AuthEvent{
		UserID:    &user.ID,
		TenantID:  &user.TenantID,
		EventType: "magic_link_login",
		Success:   true,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Metadata: map[string]interface{}{
			"is_new_user": isNewUser,
		},
	}
	if logErr := a.repo.LogAuthEvent(ctx, authEvent); logErr != nil {
		logrus.WithError(logErr).WithField("userID", user.ID).Warn("Failed to log magic link auth event")
	}

	// Convert to login user
	loginUser := userToLoginUser(user)

	return &MagicLinkVerifyResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         loginUser,
		NewUser:      isNewUser,
	}, nil
}
