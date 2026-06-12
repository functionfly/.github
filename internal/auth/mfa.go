package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// MFAService handles Multi-Factor Authentication operations
type MFAService struct {
	repo storage.Repository
}

// NewMFAService creates a new MFA service
func NewMFAService(repo storage.Repository) *MFAService {
	return &MFAService{
		repo: repo,
	}
}

// MFASetupRequest represents a request to set up MFA
type MFASetupRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// MFASetupResponse represents the response after setting up MFA
type MFASetupResponse struct {
	Secret      string `json:"secret"`
	QRCodeURL   string `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// MFAVerifyRequest represents a request to verify MFA
type MFAVerifyRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

// MFAVerifyResponse represents the response after verifying MFA
type MFAVerifyResponse struct {
	Verified bool `json:"verified"`
}

// MFADisableRequest represents a request to disable MFA
type MFADisableRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Code     string    `json:"code"` // Current MFA code for verification
	Password string    `json:"password"`
}

// SetupMFA generates a TOTP secret and backup codes for a user
func (m *MFAService) SetupMFA(req MFASetupRequest) (*MFASetupResponse, error) {
	// Get user
	user, err := m.repo.GetUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "FunctionFly",
		AccountName: user.Email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate backup codes (8 codes, each 8 characters)
	backupCodes := make([]string, 8)
	for i := 0; i < 8; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = code
	}

	// Hash backup codes with bcrypt before storage
	hashedBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedBackupCodes[i] = hashBackupCode(code)
	}

	// Update user with MFA setup
	secret := key.Secret()
	err = m.repo.UpdateUserMFA(req.UserID, &secret, false, hashedBackupCodes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update user MFA: %w", err)
	}

	return &MFASetupResponse{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyMFA verifies a TOTP code or backup code
func (m *MFAService) VerifyMFA(req MFAVerifyRequest) (*MFAVerifyResponse, error) {
	// Get user
	user, err := m.repo.GetUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.MFASecret == nil {
		return &MFAVerifyResponse{Verified: false}, nil
	}

	// Check if it's a backup code first
	if m.isValidBackupCode(user.MFABackupCodes, req.Code) {
		// Remove used backup code
		err = m.consumeBackupCode(req.UserID, req.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to consume backup code: %w", err)
		}

		// Update last used timestamp
		now := time.Now()
		err = m.repo.UpdateUserMFALastUsed(req.UserID, &now)
		if err != nil {
			return nil, fmt.Errorf("failed to update MFA last used: %w", err)
		}

		return &MFAVerifyResponse{Verified: true}, nil
	}

	// Verify TOTP code
	valid := totp.Validate(req.Code, *user.MFASecret)
	if !valid {
		return &MFAVerifyResponse{Verified: false}, nil
	}

	// Update last used timestamp
	now := time.Now()
	err = m.repo.UpdateUserMFALastUsed(req.UserID, &now)
	if err != nil {
		return nil, fmt.Errorf("failed to update MFA last used: %w", err)
	}

	return &MFAVerifyResponse{Verified: true}, nil
}

// EnableMFA enables MFA for a user after successful verification
func (m *MFAService) EnableMFA(userID uuid.UUID) error {
	return m.repo.UpdateUserMFAEnabled(userID, true)
}

// DisableMFA disables MFA for a user
func (m *MFAService) DisableMFA(req MFADisableRequest) error {
	// Get user
	user, err := m.repo.GetUserByID(req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Verify password
	valid, err := m.repo.VerifyPassword(req.UserID, req.Password)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid password")
	}

	// Verify MFA code if MFA is enabled
	if user.MFAEnabled {
		if user.MFASecret == nil {
			return fmt.Errorf("MFA secret not found")
		}

		validMFA := totp.Validate(req.Code, *user.MFASecret)
		if !validMFA {
			return fmt.Errorf("invalid MFA code")
		}
	}

	// Disable MFA
	return m.repo.UpdateUserMFA(req.UserID, nil, false, nil, nil)
}

// IsMFARequired checks if MFA is required for a user (admin users)
func (m *MFAService) IsMFARequired(userID uuid.UUID) (bool, error) {
	user, err := m.repo.GetUserByID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Require MFA for admin users
	return user.Role == "admin" || user.Role == "super_admin", nil
}

// GetMFAStatus returns the MFA status for a user
func (m *MFAService) GetMFAStatus(userID uuid.UUID) (*MFAStatus, error) {
	user, err := m.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &MFAStatus{
		Enabled:     user.MFAEnabled,
		LastUsed:    user.MFALastUsed,
		Required:    user.Role == "admin" || user.Role == "super_admin",
		BackupCodes: len(user.MFABackupCodes),
	}, nil
}

// MFAStatus represents the MFA status for a user
type MFAStatus struct {
	Enabled     bool       `json:"enabled"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	Required    bool       `json:"required"`
	BackupCodes int        `json:"backup_codes_remaining"`
}

// generateBackupCode generates a random 8-character backup code
func generateBackupCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// hashBackupCode hashes a backup code for storage using bcrypt
// Uses cost 14 for enhanced security since backup codes are high-entropy
// but used as a fallback when TOTP is unavailable.
func hashBackupCode(code string) string {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(code), 14)
	if err != nil {
		return "INVALID_HASH_ERROR"
	}
	return string(hashedBytes)
}

// isValidBackupCode checks if a code matches any stored backup code
func (m *MFAService) isValidBackupCode(storedCodes []string, code string) bool {
	for _, stored := range storedCodes {
		// Use bcrypt to verify the code against the stored hash
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(code))
		if err == nil {
			return true
		}
	}
	return false
}

// consumeBackupCode removes a used backup code
func (m *MFAService) consumeBackupCode(userID uuid.UUID, code string) error {
	user, err := m.repo.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	var remainingCodes []string
	for _, stored := range user.MFABackupCodes {
		// Check if this stored hash matches the input code
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(code))
		if err != nil {
			// This hash doesn't match the code, keep it
			remainingCodes = append(remainingCodes, stored)
		}
		// If err == nil, this code matches and we don't add it to remainingCodes (removing it)
	}

	return m.repo.UpdateUserMFABackupCodes(userID, remainingCodes)
}
