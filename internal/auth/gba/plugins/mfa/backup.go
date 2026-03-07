package mfa

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Number of backup codes to generate
	backupCodeCount = 10
	// Length of each backup code segment
	backupCodeSegmentLength = 4
	// Number of segments in a backup code (e.g., XXXX-XXXX-XXXX)
	backupCodeSegmentCount = 3
)

// generateBackupCodes generates a set of single-use backup codes for MFA recovery
// Returns the plain text codes (must be shown to user once) and stores hashes
func (p *MFAPlugin) generateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	codes := make([]string, 0, backupCodeCount)

	// Delete any existing backup codes for this user
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&MFABackupCode{}).Error; err != nil {
		return nil, fmt.Errorf("failed to clean up old backup codes: %w", err)
	}

	// Generate new backup codes
	for i := 0; i < backupCodeCount; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}

		// Hash the code for storage
		hash, err := hashBackupCode(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}

		// Store the hash
		backupCode := &MFABackupCode{
			UserID:   userID,
			CodeHash: hash,
			Used:     false,
		}

		if err := p.db.WithContext(ctx).Create(backupCode).Error; err != nil {
			return nil, fmt.Errorf("failed to store backup code: %w", err)
		}

		codes = append(codes, code)
	}

	p.logger.WithField("user_id", userID).Info("Backup codes generated")
	return codes, nil
}

// RegenerateBackupCodes regenerates all backup codes for a user
// Requires a valid TOTP code for verification
func (p *MFAPlugin) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, totpCode string) ([]string, error) {
	// Verify the TOTP code first
	if err := p.verifyTOTP(ctx, userID, totpCode); err != nil {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	// Generate new backup codes
	codes, err := p.generateBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	p.logger.WithField("user_id", userID).Info("Backup codes regenerated")
	return codes, nil
}

// verifyBackupCode verifies a backup code and marks it as used if valid
func (p *MFAPlugin) verifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	// Normalize the code (remove dashes, uppercase)
	normalizedCode := normalizeBackupCode(code)

	// Get all unused backup codes for this user
	var backupCodes []MFABackupCode
	if err := p.db.WithContext(ctx).Where("user_id = ? AND used = false", userID).Find(&backupCodes).Error; err != nil {
		return fmt.Errorf("failed to retrieve backup codes: %w", err)
	}

	// Check each code
	for i := range backupCodes {
		if verifyBackupCodeHash(normalizedCode, backupCodes[i].CodeHash) {
			// Mark as used
			backupCodes[i].MarkAsUsed()
			if err := p.db.WithContext(ctx).Save(&backupCodes[i]).Error; err != nil {
				p.logger.WithError(err).Error("Failed to mark backup code as used")
				// Don't fail the verification if we can't mark as used
			}

			p.logger.WithField("user_id", userID).Info("Backup code used for MFA verification")
			return nil
		}
	}

	return fmt.Errorf("invalid or already used backup code")
}

// GetBackupCodeCount returns the number of unused backup codes for a user
func (p *MFAPlugin) GetBackupCodeCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := p.db.WithContext(ctx).Model(&MFABackupCode{}).Where("user_id = ? AND used = false", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count backup codes: %w", err)
	}
	return count, nil
}

// HasBackupCodes returns true if the user has at least one unused backup code
func (p *MFAPlugin) HasBackupCodes(ctx context.Context, userID uuid.UUID) (bool, error) {
	count, err := p.GetBackupCodeCount(ctx, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Helper functions

// generateBackupCode generates a random backup code in format XXXX-XXXX-XXXX
func generateBackupCode() (string, error) {
	// Calculate total bytes needed
	// Each character in base32 is 5 bits, so we need (segmentLength * segmentCount * 5) / 8 bytes
	// For 4*3=12 characters, we need 8 bytes (64 bits / 5 = 12.8 -> 13 base32 chars)
	bytesNeeded := (backupCodeSegmentLength*backupCodeSegmentCount*5 + 7) / 8

	randomBytes := make([]byte, bytesNeeded)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Encode to base32 (uppercase alphanumeric without confusing characters)
	encoded := base32.StdEncoding.EncodeToString(randomBytes)

	// Take only the characters we need
	encoded = encoded[:backupCodeSegmentLength*backupCodeSegmentCount]

	// Format with dashes
	var result string
	for i := 0; i < backupCodeSegmentCount; i++ {
		if i > 0 {
			result += "-"
		}
		start := i * backupCodeSegmentLength
		end := start + backupCodeSegmentLength
		result += encoded[start:end]
	}

	return result, nil
}

// hashBackupCode hashes a backup code using bcrypt
func hashBackupCode(code string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyBackupCodeHash verifies a backup code against its bcrypt hash
func verifyBackupCodeHash(code, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil
}

// normalizeBackupCode removes dashes and converts to uppercase
func normalizeBackupCode(code string) string {
	result := make([]byte, 0, len(code))
	for i := 0; i < len(code); i++ {
		c := code[i]
		if c >= 'a' && c <= 'z' {
			result = append(result, c-'a'+'A')
		} else if c != '-' && c != ' ' {
			result = append(result, c)
		}
	}
	return string(result)
}

// ValidateBackupCodeFormat validates that a backup code has the correct format
func ValidateBackupCodeFormat(code string) bool {
	normalized := normalizeBackupCode(code)
	expectedLength := backupCodeSegmentLength * backupCodeSegmentCount
	if len(normalized) != expectedLength {
		return false
	}

	// Check that all characters are valid base32 characters
	for i := 0; i < len(normalized); i++ {
		c := normalized[i]
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')) {
			return false
		}
	}

	return true
}
