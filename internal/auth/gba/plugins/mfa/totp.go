package mfa

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GenerateOpts contains options for TOTP generation
type GenerateOpts struct {
	Issuer      string
	AccountName string
	Period      uint
	Digits      uint
}

// DefaultGenerateOpts returns default TOTP generation options
func (p *MFAPlugin) DefaultGenerateOpts(accountName string) GenerateOpts {
	return GenerateOpts{
		Issuer:      p.config.TOTPIssuer,
		AccountName: accountName,
		Period:      p.config.TOTPPeriod,
		Digits:      p.config.TOTPDigits,
	}
}

// GenerateTOTP generates a new TOTP secret and QR code URL for a user
// This is the first step in MFA setup
func (p *MFAPlugin) GenerateTOTP(ctx context.Context, userID uuid.UUID, accountName string) (*TOTPSetup, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("MFA is not enabled")
	}

	// Generate a random secret
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// Encrypt the secret for storage
	encryptedSecret, err := encryptSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Generate QR code URL
	opts := p.DefaultGenerateOpts(accountName)
	qrCodeURL, err := totp.Generate(totp.GenerateOpts{
		Issuer:      opts.Issuer,
		AccountName: opts.AccountName,
		Secret:      []byte(secret),
		Period:      opts.Period,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code URL: %w", err)
	}

	// Store unverified TOTP configuration
	totpConfig := &MFATOTP{
		UserID:   userID,
		Secret:   encryptedSecret,
		Enabled:  false,
		Verified: false,
	}

	// Delete any existing unverified configuration
	if err := p.db.WithContext(ctx).Where("user_id = ? AND verified = false", userID).Delete(&MFATOTP{}).Error; err != nil {
		p.logger.WithError(err).Warn("Failed to clean up unverified TOTP config")
	}

	// Create new configuration
	if err := p.db.WithContext(ctx).Create(totpConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to store TOTP configuration: %w", err)
	}

	// Generate backup codes
	backupCodes, err := p.generateBackupCodes(ctx, userID)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to generate backup codes")
		backupCodes = []string{}
	}

	p.logger.WithField("user_id", userID).Info("TOTP setup initiated")

	return &TOTPSetup{
		Secret:      secret,
		QRCodeURL:   qrCodeURL.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndEnableTOTP verifies a TOTP code and enables MFA for the user
// This is the second step in MFA setup - user must provide a valid code from their authenticator app
func (p *MFAPlugin) VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	if !p.config.Enabled {
		return fmt.Errorf("MFA is not enabled")
	}

	// Find the unverified TOTP configuration
	var totpConfig MFATOTP
	result := p.db.WithContext(ctx).Where("user_id = ? AND verified = false", userID).First(&totpConfig)
	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("no pending TOTP setup found")
	}
	if result.Error != nil {
		return fmt.Errorf("failed to find TOTP configuration: %w", result.Error)
	}

	// Decrypt the secret
	secret, err := decryptSecret(totpConfig.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Verify the code
	if !validateTOTPCode(secret, code, p.config.TOTPPeriod, p.config.SkewPeriods) {
		return fmt.Errorf("invalid TOTP code")
	}

	// Mark as verified and enabled
	totpConfig.Verified = true
	totpConfig.Enabled = true
	totpConfig.UpdatedAt = time.Now()

	if err := p.db.WithContext(ctx).Save(&totpConfig).Error; err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	// Update user's MFA flag
	if err := p.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("mfa_enabled", true).Error; err != nil {
		p.logger.WithError(err).Warn("Failed to update user MFA flag")
	}

	p.logger.WithField("user_id", userID).Info("TOTP enabled successfully")
	return nil
}

// verifyTOTP verifies a TOTP code for an enabled user
// This is used during login MFA challenge
func (p *MFAPlugin) verifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	// Find the enabled TOTP configuration
	var totpConfig MFATOTP
	result := p.db.WithContext(ctx).Where("user_id = ? AND enabled = true", userID).First(&totpConfig)
	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("MFA not enabled for user")
	}
	if result.Error != nil {
		return fmt.Errorf("failed to find TOTP configuration: %w", result.Error)
	}

	// Decrypt the secret
	secret, err := decryptSecret(totpConfig.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Verify the code
	if !validateTOTPCode(secret, code, p.config.TOTPPeriod, p.config.SkewPeriods) {
		return fmt.Errorf("invalid TOTP code")
	}

	return nil
}

// RegenerateTOTPSecret regenerates the TOTP secret for a user
// Requires verification with existing MFA method
func (p *MFAPlugin) RegenerateTOTPSecret(ctx context.Context, userID uuid.UUID, verificationCode string, accountName string) (*TOTPSetup, error) {
	// Verify the user has MFA enabled and can provide valid code
	if err := p.VerifyCode(ctx, userID, verificationCode); err != nil {
		return nil, fmt.Errorf("invalid verification code")
	}

	// Delete existing TOTP configuration
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&MFATOTP{}).Error; err != nil {
		return nil, fmt.Errorf("failed to reset TOTP: %w", err)
	}

	// Delete existing backup codes
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&MFABackupCode{}).Error; err != nil {
		p.logger.WithError(err).Warn("Failed to delete old backup codes")
	}

	// Generate new TOTP
	return p.GenerateTOTP(ctx, userID, accountName)
}

// Helper functions

// generateSecret generates a random base32-encoded secret
func generateSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

const (
	secretPrefixEnc  = "enc:"
	secretPrefixPlain = "plain:"
)

// getMFAEncryptionKey returns the 32-byte key for MFA secret encryption from env.
// Use MFA_ENCRYPTION_KEY or GBA_MFA_ENCRYPTION_KEY (base64-encoded, 32 bytes decoded).
func getMFAEncryptionKey() []byte {
	for _, name := range []string{"MFA_ENCRYPTION_KEY", "GBA_MFA_ENCRYPTION_KEY"} {
		if b64 := os.Getenv(name); b64 != "" {
			key, err := base64.StdEncoding.DecodeString(b64)
			if err != nil || len(key) != 32 {
				continue
			}
			return key
		}
	}
	return nil
}

func isMFAEncryptionAvailable() bool {
	key := getMFAEncryptionKey()
	return len(key) == 32
}

func isProduction() bool {
	return os.Getenv("PRODUCTION") == "true" || os.Getenv("ENVIRONMENT") == "production"
}

// encryptSecret encrypts a TOTP secret for storage using AES-256-GCM when a key is configured.
// SECURITY: If no encryption key is set, plain storage is NOT allowed - returns error in production.
// Only in development mode is plaintext storage permitted (with warning).
func encryptSecret(secret string) (string, error) {
	key := getMFAEncryptionKey()
	if len(key) == 32 {
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return "", err
		}
		ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
		return secretPrefixEnc + base64.StdEncoding.EncodeToString(ciphertext), nil
	}
	if isProduction() {
		return "", fmt.Errorf("MFA_ENCRYPTION_KEY not configured - TOTP secrets cannot be stored in production without encryption")
	}
	logrus.Warn("MFA_ENCRYPTION_KEY not set - TOTP secret stored without encryption (DEV ONLY). Set MFA_ENCRYPTION_KEY in production")
	return secretPrefixPlain + base64.StdEncoding.EncodeToString([]byte(secret)), nil
}

// decryptSecret decrypts a stored TOTP secret (AES-256-GCM or legacy plain).
func decryptSecret(encrypted string) (string, error) {
	if strings.HasPrefix(encrypted, secretPrefixEnc) {
		key := getMFAEncryptionKey()
		if len(key) != 32 {
			return "", fmt.Errorf("MFA_ENCRYPTION_KEY not set or invalid; cannot decrypt")
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encrypted, secretPrefixEnc))
		if err != nil {
			return "", err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}
		nonceSize := gcm.NonceSize()
		if len(data) < nonceSize {
			return "", fmt.Errorf("ciphertext too short")
		}
		plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
		if err != nil {
			return "", err
		}
		return string(plaintext), nil
	}
	if strings.HasPrefix(encrypted, secretPrefixPlain) {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encrypted, secretPrefixPlain))
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}
	return "", fmt.Errorf("invalid or legacy secret format; re-enroll MFA")
}

// validateTOTPCode validates a TOTP code against a secret
func validateTOTPCode(secret, code string, period, skew uint) bool {
	// Use totp.Validate with skew support
	valid := totp.Validate(code, secret)
	if valid {
		return true
	}

	// If strict validation failed, try with skew tolerance
	// Generate codes for previous, current, and next periods
	now := time.Now()
	skewInt := int(skew)

	for i := -skewInt; i <= skewInt; i++ {
		targetTime := now.Add(time.Duration(i) * time.Duration(period) * time.Second)
		expectedCode, err := totp.GenerateCode(secret, targetTime)
		if err != nil {
			continue
		}
		if expectedCode == code {
			return true
		}
	}

	return false
}
