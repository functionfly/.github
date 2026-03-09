// Package mfa provides Multi-Factor Authentication support for GoBetterAuth
// This is Phase 2 of the Better Auth migration plan - TOTP implementation
package mfa

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MFAPlugin provides Multi-Factor Authentication functionality
type MFAPlugin struct {
	db     *gorm.DB
	config *MFAConfig
	logger *logrus.Logger
}

// MFAConfig holds MFA configuration
type MFAConfig struct {
	Enabled        bool   // Master switch for MFA
	RequireOnSetup bool   // Require MFA during account setup
	TOTPIssuer     string // TOTP issuer name (displayed in authenticator apps)
	TOTPPeriod     uint   // TOTP period in seconds (default: 30)
	TOTPDigits     uint   // TOTP code digits (default: 6)
	SkewPeriods    uint   // Clock skew tolerance periods (default: 1)
}

// DefaultMFAConfig returns default MFA configuration
func DefaultMFAConfig() *MFAConfig {
	return &MFAConfig{
		Enabled:        false,
		RequireOnSetup: false,
		TOTPIssuer:     getEnvOrDefault("TOTP_ISSUER", "FunctionFly"),
		TOTPPeriod:     30,
		TOTPDigits:     6,
		SkewPeriods:    1,
	}
}

// New creates a new MFA plugin instance
func New(db *gorm.DB, config *MFAConfig, logger *logrus.Logger) (*MFAPlugin, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if config == nil {
		config = DefaultMFAConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&MFATOTP{}, &MFABackupCode{}); err != nil {
		return nil, fmt.Errorf("failed to migrate MFA models: %w", err)
	}

	return &MFAPlugin{
		db:     db,
		config: config,
		logger: logger,
	}, nil
}

// IsEnabled returns true if MFA is enabled globally
func (p *MFAPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// IsEnabledForUser checks if MFA is enabled for a specific user
func (p *MFAPlugin) IsEnabledForUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	var totp MFATOTP
	result := p.db.WithContext(ctx).Where("user_id = ? AND enabled = true", userID).First(&totp)
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if result.Error != nil {
		return false, fmt.Errorf("failed to check MFA status: %w", result.Error)
	}
	return true, nil
}

// GetStatus returns the MFA status for a user
func (p *MFAPlugin) GetStatus(ctx context.Context, userID uuid.UUID) (*MFAStatus, error) {
	var totp MFATOTP
	result := p.db.WithContext(ctx).Where("user_id = ?", userID).First(&totp)

	status := &MFAStatus{
		Enabled:        false,
		Verified:       false,
		HasBackupCodes: false,
	}

	if result.Error == gorm.ErrRecordNotFound {
		return status, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get MFA status: %w", result.Error)
	}

	status.Enabled = totp.Enabled
	status.Verified = totp.Verified

	// Check if user has backup codes
	var backupCount int64
	if err := p.db.WithContext(ctx).Model(&MFABackupCode{}).Where("user_id = ? AND used = false", userID).Count(&backupCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count backup codes: %w", err)
	}
	status.HasBackupCodes = backupCount > 0
	status.BackupCodeCount = int(backupCount)

	return status, nil
}

// MFAStatus represents the MFA status for a user
type MFAStatus struct {
	Enabled         bool `json:"enabled"`
	Verified        bool `json:"verified"`
	HasBackupCodes  bool `json:"has_backup_codes"`
	BackupCodeCount int  `json:"backup_code_count,omitempty"`
}

// VerifyCode verifies a TOTP code or backup code for a user
// Returns nil on success, error on failure
func (p *MFAPlugin) VerifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	// Try TOTP verification first
	if err := p.verifyTOTP(ctx, userID, code); err == nil {
		return nil
	}

	// Try backup code verification
	if err := p.verifyBackupCode(ctx, userID, code); err == nil {
		return nil
	}

	return fmt.Errorf("invalid MFA code")
}

// Disable disables MFA for a user
// Requires either a valid TOTP code or backup code for verification
func (p *MFAPlugin) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	// Verify the code first
	if err := p.VerifyCode(ctx, userID, code); err != nil {
		return fmt.Errorf("invalid verification code")
	}

	// Delete TOTP configuration
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&MFATOTP{}).Error; err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Delete all backup codes
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&MFABackupCode{}).Error; err != nil {
		p.logger.WithError(err).Warn("Failed to delete backup codes when disabling MFA")
	}

	// Update user's MFA flag
	if err := p.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("mfa_enabled", false).Error; err != nil {
		p.logger.WithError(err).Warn("Failed to update user MFA flag")
	}

	p.logger.WithField("user_id", userID).Info("MFA disabled")
	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// User is a minimal interface for the user model
// This avoids circular imports
type User struct {
	ID         uuid.UUID
	MFAEnabled bool `gorm:"default:false"`
}

// TableName returns the table name
func (User) TableName() string {
	return "gba_users"
}
