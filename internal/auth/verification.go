package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// generateVerificationToken generates a secure random verification token
func (a *AuthService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// VerifyEmail verifies a user's email address using a verification token
func (a *AuthService) VerifyEmail(token string) error {
	// Get user by verification token
	user, err := a.repo.GetUserByVerificationToken(token)
	if err != nil {
		return fmt.Errorf("failed to get user by verification token: %w", err)
	}
	if user == nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Check if token has expired
	if user.VerificationExpiresAt != nil && time.Now().After(*user.VerificationExpiresAt) {
		return fmt.Errorf("verification token has expired")
	}

	// Mark email as verified and clear verification token
	err = a.repo.UpdateUserEmailVerification(nil, user.ID, true, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email to an unverified user.
// Returns a generic error for non-existent and already-verified accounts to
// avoid leaking account existence information to callers.
func (a *AuthService) ResendVerificationEmail(email string) error {
	// Get user by email
	user, err := a.repo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	// Return a generic message so callers cannot enumerate accounts
	if user == nil {
		return fmt.Errorf("if that email is registered and unverified, a new verification email has been sent")
	}

	// Already verified — return a generic message to avoid distinguishing from non-existent
	if user.EmailVerified {
		return fmt.Errorf("if that email is registered and unverified, a new verification email has been sent")
	}

	// Generate new verification token
	verificationToken, err := a.generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

	// Update user with new verification token
	err = a.repo.UpdateUserEmailVerification(nil, user.ID, false, &verificationToken, &expiresAt)
	if err != nil {
		return fmt.Errorf("failed to update user verification: %w", err)
	}

	// Update user object for email service
	user.VerificationToken = &verificationToken
	user.VerificationExpiresAt = &expiresAt

	// Send verification email
	err = a.emailSvc.SendVerificationEmail(user, verificationToken)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}