package auth

import (
	"context"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
	"unicode"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestPasswordReset sends a password reset email to the given address.
// Returns nil even if the email doesn't exist to avoid account enumeration.
func (a *AuthService) RequestPasswordReset(email string) error {
	user, err := a.repo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to look up user: %w", err)
	}
	if user == nil {
		// Don't reveal that the email doesn't exist
		return nil
	}

	// Generate a secure reset token (reuse the verification token mechanism)
	resetToken, err := a.generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}
	expiresAt := time.Now().Add(1 * time.Hour) // 1 hour expiry for password resets

	// Store the token in the verification_token field (reuse existing column)
	if err := a.repo.UpdateUserEmailVerification(nil, user.ID, user.EmailVerified, &resetToken, &expiresAt); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send the reset email
	if err := a.emailSvc.SendPasswordResetEmail(user, resetToken); err != nil {
		logrus.WithError(err).WithField("email", email).Warn("Failed to send password reset email")
		// Don't fail — the token is stored; user can request again
	}

	return nil
}

// ConfirmPasswordReset validates the reset token and sets a new password.
func (a *AuthService) ConfirmPasswordReset(token, newPassword string) error {
	if token == "" || newPassword == "" {
		return fmt.Errorf("token and new password are required")
	}

	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Look up user by the reset token (stored in verification_token column)
	user, err := a.repo.GetUserByVerificationToken(token)
	if err != nil {
		return fmt.Errorf("failed to look up reset token: %w", err)
	}
	if user == nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check expiry
	if user.VerificationExpiresAt != nil && time.Now().After(*user.VerificationExpiresAt) {
		return fmt.Errorf("reset token has expired")
	}

	// Hash the new password
	hashedPassword, err := a.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear the reset token
	updates := map[string]interface{}{
		"password_hash":           hashedPassword,
		"verification_token":      nil,
		"verification_expires_at": nil,
	}
	if _, err := a.repo.UpdateUser(context.Background(), user.ID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// validatePasswordStrength checks that a password meets minimum security requirements:
//   - At least 8 characters
//   - At least one uppercase letter
//   - At least one lowercase letter
//   - At least one digit
//   - At least one special character
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	var missing []string
	if !hasUpper {
		missing = append(missing, "an uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "a lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "a digit")
	}
	if !hasSpecial {
		missing = append(missing, "a special character")
	}
	if len(missing) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(missing, ", "))
	}
	return nil
}

// Login authenticates a user and returns a JWT token
func (a *AuthService) Login(email, password, ipAddress, userAgent string) (res *LoginResponse, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).WithField("stack", string(debug.Stack())).Error("Login panic")
			res = nil
			err = fmt.Errorf("internal error")
		}
	}()

	if a.repo == nil {
		return nil, fmt.Errorf("internal error: auth not configured")
	}

	// Get user by email
	user, err := a.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return nil, fmt.Errorf("email not verified - please check your email and verify your account")
	}

	// Verify password (treat any verify error as invalid credentials to avoid 500 on bad hash format)
	valid, err := a.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Debug("Password verification error")
		return nil, fmt.Errorf("invalid credentials")
	}
	if !valid {
		return nil, fmt.Errorf("invalid credentials")
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

	// Store refresh token in database (expires in 30 days)
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = a.repo.CreateRefreshToken(user.ID, refreshTokenHash, ipAddress, userAgent, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Build safe user for response (no password hash, MFA secrets, or verification tokens)
	loginUser := userToLoginUser(user)

	// Set tenant plan for billing/UI
	if tenant, err := a.repo.GetTenantByID(user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
		loginUser.Plan = tenant.Plan
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         loginUser,
	}, nil
}

// userToLoginUser maps storage.User to safe LoginUser for API responses
func userToLoginUser(u *storage.User) *LoginUser {
	if u == nil {
		return nil
	}
	lu := &LoginUser{
		ID:            u.ID.String(),
		TenantID:      u.TenantID.String(),
		Email:         u.Email,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.Username != nil && *u.Username != "" {
		lu.Username = *u.Username
	} else if u.Email != "" {
		// Fallback: use email local part so login/CLI can show a name (e.g. admin@functionfly.local → admin)
		if at := strings.Index(u.Email, "@"); at > 0 {
			lu.Username = u.Email[:at]
		} else {
			lu.Username = u.Email
		}
	}
	if u.CompanyName != nil && *u.CompanyName != "" {
		lu.CompanyName = *u.CompanyName
	}
	// Prefer the explicit Name field; fall back to OAuth provider data
	if u.Name != "" {
		lu.Name = u.Name
	}
	if u.ProviderData != nil {
		lu.ProviderData = u.ProviderData
		if lu.Name == "" {
			if n, ok := u.ProviderData["name"].(string); ok {
				lu.Name = n
			}
		}
		if lu.Name == "" {
			if fn, ok := u.ProviderData["full_name"].(string); ok {
				lu.Name = fn
			}
		}
		if a, ok := u.ProviderData["avatar_url"].(string); ok {
			lu.Avatar = a
		}
		if p, ok := u.ProviderData["picture"].(string); ok && lu.Avatar == "" {
			lu.Avatar = p
		}
	}
	return lu
}

// Signup creates a new user account (unverified, requires email confirmation)
func (a *AuthService) Signup(req SignupRequest) (*SignupResponse, error) {
	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return nil, fmt.Errorf("email, password, and username are required")
	}

	if req.Password != req.ConfirmPassword {
		return nil, fmt.Errorf("passwords do not match")
	}

	if err := validatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	if !req.TermsAccepted {
		return nil, fmt.Errorf("terms must be accepted")
	}

	// Check if user already exists
	existingUser, err := a.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		if existingUser.EmailVerified {
			return nil, fmt.Errorf("user with this email already exists")
		}
		// If user exists but is not verified, we'll resend the verification email
		// Generate new verification token
		verificationToken, err := a.generateVerificationToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate verification token: %w", err)
		}
		expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

		// Update user with new verification token
		err = a.repo.UpdateUserEmailVerification(nil, existingUser.ID, false, &verificationToken, &expiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to update user verification: %w", err)
		}

		// Send verification email
		existingUser.VerificationToken = &verificationToken
		existingUser.VerificationExpiresAt = &expiresAt
		err = a.emailSvc.SendVerificationEmail(existingUser, verificationToken)
		if err != nil {
			return nil, fmt.Errorf("failed to send verification email: %w", err)
		}

		return &SignupResponse{
			Message:              "Verification email sent. Please check your email to complete registration.",
			EmailSent:            true,
			RequiresVerification: true,
		}, nil
	}

	// Hash password
	hashedPassword, err := a.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate verification token
	verificationToken, err := a.generateVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

	// Get or create default tenant
	tenants, err := a.repo.ListTenants()
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	var tenantID uuid.UUID
	if len(tenants) > 0 {
		// Use the first tenant as default
		tenantID = tenants[0].ID
	} else {
		// Create a default tenant
		tenant, err := a.repo.CreateTenant(context.Background(), "Default Tenant")
		if err != nil {
			return nil, fmt.Errorf("failed to create default tenant: %w", err)
		}
		tenantID = tenant.ID
	}

	// Create user with verification token
	user, err := a.repo.CreateUser(req.Email, hashedPassword, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Set optional username and company name if provided
	if req.Username != "" || req.CompanyName != "" {
		updates := make(map[string]interface{})
		if req.Username != "" {
			updates["username"] = req.Username
		}
		if req.CompanyName != "" {
			updates["company_name"] = req.CompanyName
		}
		if _, err := a.repo.UpdateUser(context.Background(), user.ID, updates); err != nil {
			// Log but don't fail signup - user can set these later
			fmt.Printf("Warning: failed to set username/company name: %v\n", err)
		} else {
			if req.Username != "" {
				user.Username = &req.Username
			}
			if req.CompanyName != "" {
				user.CompanyName = &req.CompanyName
			}
		}
	}

	// Set verification details
	err = a.repo.UpdateUserEmailVerification(nil, user.ID, false, &verificationToken, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to set verification details: %w", err)
	}

	// Update user object for email service
	user.VerificationToken = &verificationToken
	user.VerificationExpiresAt = &expiresAt

	// Send verification email
	err = a.emailSvc.SendVerificationEmail(user, verificationToken)
	emailSent := err == nil
	if err != nil {
		// Log error but don't fail signup - user can request resend
		fmt.Printf("Warning: failed to send verification email: %v\n", err)
	}

	// Welcome notification (in-app) so the user sees it when they open the platform
	a.sendWelcomeNotification(context.Background(), user.ID)

	return &SignupResponse{
		Message:              "Account created successfully. Please check your email to verify your account.",
		EmailSent:            emailSent,
		RequiresVerification: true,
	}, nil
}

// CheckUsernameAvailability checks if a username is available for registration
func (a *AuthService) CheckUsernameAvailability(username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username is required")
	}

	// Check if username matches validation pattern
	if !strings.Contains(username, "") { // This will be checked by the regex
		// Use the same validation as in the schema
		matched, err := regexp.MatchString("^[a-zA-Z0-9_-]*$", username)
		if err != nil {
			return false, fmt.Errorf("failed to validate username format: %w", err)
		}
		if !matched {
			return false, fmt.Errorf("username contains invalid characters")
		}
	}

	// Check length constraints
	if len(username) > 50 {
		return false, fmt.Errorf("username is too long")
	}

	// Check if username is already taken
	existingUser, err := a.repo.GetUserByUsername(username)
	if err != nil {
		return false, fmt.Errorf("failed to check username availability: %w", err)
	}

	// If user exists, username is not available
	if existingUser != nil {
		return false, nil
	}

	// Username is available
	return true, nil
}

// sendWelcomeNotification sends an in-app welcome notification if a notifier is configured.
func (a *AuthService) sendWelcomeNotification(ctx context.Context, userID uuid.UUID) {
	if a.notifySvc != nil {
		if err := a.notifySvc.SendWelcome(ctx, userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("Failed to send welcome notification")
		}
	}
}
