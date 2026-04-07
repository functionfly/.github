package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
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

// Account lockout configuration functions
func getMaxLoginAttempts() int {
	if val := os.Getenv("MAX_LOGIN_ATTEMPTS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 1 {
			return i
		}
	}
	return 5 // Default to 5 failed attempts before lockout
}

func getAccountLockoutDuration() time.Duration {
	if val := os.Getenv("ACCOUNT_LOCKOUT_DURATION_MINUTES"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 1 {
			return time.Duration(i) * time.Minute
		}
	}
	return 15 * time.Minute // Default to 15 minutes lockout
}

// recordFailedLoginAttempt records a failed login attempt and locks the account if threshold is exceeded
func (a *AuthService) recordFailedLoginAttempt(userID uuid.UUID, email, ipAddress, userAgent string) {
	maxAttempts := getMaxLoginAttempts()
	lockoutDuration := getAccountLockoutDuration()

	// Record the failed attempt
	_, err := a.repo.CreateLoginAttempt(userID, ipAddress, userAgent, false, nil)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to record failed login attempt")
		return
	}

	// Count recent failed attempts
	recentFailures, err := a.repo.GetRecentFailedLoginAttempts(userID, time.Now().Add(-lockoutDuration))
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get recent failed login attempts")
		return
	}

	// Check if we should lock the account
	if recentFailures >= maxAttempts {
		lockoutUntil := time.Now().Add(lockoutDuration)
		// Create a login attempt with the lockout
		_, err := a.repo.CreateLoginAttempt(userID, ipAddress, userAgent, false, &lockoutUntil)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to record account lockout")
			return
		}

		logrus.WithFields(logrus.Fields{
			"user_id":        userID,
			"email":          email,
			"failed_attempts": recentFailures,
			"lockout_until":  lockoutUntil,
		}).Warn("Account locked due to too many failed login attempts")
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"email":      email,
		"reason":     "invalid_credentials",
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}).Info("Failed login attempt recorded")
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

	// SECURITY FIX: Check if account is locked due to too many failed login attempts
	// Lockout status is checked via login_attempts table
	if lockoutUntil, lockoutErr := a.repo.GetUserLockoutStatus(user.ID); lockoutErr == nil && lockoutUntil != nil {
		if time.Now().Before(*lockoutUntil) {
			logrus.WithFields(logrus.Fields{
				"user_id":       user.ID,
				"lockout_until": lockoutUntil,
			}).Warn("Login attempt on locked account")
			return nil, fmt.Errorf("account temporarily locked due to too many failed login attempts. Please try again later")
		}
		// Lockout expired, continue with authentication
	}

	// Verify password (treat any verify error as invalid credentials to avoid 500 on bad hash format)
	valid, err := a.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Debug("Password verification error")
		// SECURITY FIX: Record failed login attempt
		a.recordFailedLoginAttempt(user.ID, email, ipAddress, userAgent)
		return nil, fmt.Errorf("invalid credentials")
	}
	if !valid {
		// SECURITY FIX: Record failed login attempt
		a.recordFailedLoginAttempt(user.ID, email, ipAddress, userAgent)
		return nil, fmt.Errorf("invalid credentials")
	}

	// SECURITY FIX: Successful login - clear any failed login attempts
	if err := a.repo.ClearUserLockout(user.ID); err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to clear user lockout on successful login")
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

// minimumSignupAge is the minimum age required to create an account (COPPA-aligned threshold).
const minimumSignupAge = 13

func ageFromBirthdateUTC(birth time.Time, now time.Time) int {
	birth = birth.UTC()
	now = now.UTC()
	age := now.Year() - birth.Year()
	anniversary := time.Date(now.Year(), birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
	if now.Before(anniversary) {
		age--
	}
	return age
}

func parseDateOfBirthForSignup(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("date of birth is required")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date of birth")
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if t.After(todayDate) {
		return time.Time{}, fmt.Errorf("date of birth cannot be in the future")
	}
	if ageFromBirthdateUTC(t, today) < minimumSignupAge {
		return time.Time{}, fmt.Errorf("you must be at least %d years old to register", minimumSignupAge)
	}
	return t, nil
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

	var reservedInviteID uuid.UUID
	signupFailed := true
	defer func() {
		if reservedInviteID != uuid.Nil && signupFailed {
			if err := a.repo.ReleaseSignupInviteReservation(context.Background(), reservedInviteID); err != nil {
				logrus.WithError(err).Warn("failed to release signup invite reservation after error")
			}
		}
	}()

	if SignupInviteRequired() {
		if strings.TrimSpace(req.InviteCode) == "" {
			return nil, fmt.Errorf("a valid invite code is required to sign up")
		}
		id, err := a.repo.ReserveSignupInvite(context.Background(), req.InviteCode)
		if err != nil {
			if errors.Is(err, storage.ErrSignupInviteInvalid) {
				return nil, fmt.Errorf("invalid or expired invite code")
			}
			if errors.Is(err, storage.ErrSignupInviteExhausted) {
				return nil, fmt.Errorf("this invite code has no uses remaining")
			}
			if errors.Is(err, storage.ErrSignupInviteRevoked) {
				return nil, fmt.Errorf("this invite code is no longer valid")
			}
			return nil, fmt.Errorf("invite code could not be validated: %w", err)
		}
		reservedInviteID = id
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

	// Create a new tenant for each user (don't reuse existing tenants)
	// This ensures each user gets their own tenant with the correct plan
	tenant, err := a.repo.CreateTenant(context.Background(), "Default Tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to create default tenant: %w", err)
	}
	tenantID := tenant.ID

	dob, err := parseDateOfBirthForSignup(req.DateOfBirth)
	if err != nil {
		return nil, err
	}

	// Create user with verification token
	user, err := a.repo.CreateUser(req.Email, hashedPassword, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Profile fields: username, company, display name, date of birth
	updates := map[string]interface{}{
		"date_of_birth": &dob,
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.CompanyName != "" {
		updates["company_name"] = req.CompanyName
	}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if _, err := a.repo.UpdateUser(context.Background(), user.ID, updates); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}
	if req.Username != "" {
		user.Username = &req.Username
	}
	if req.CompanyName != "" {
		user.CompanyName = &req.CompanyName
	}
	if strings.TrimSpace(req.Name) != "" {
		user.Name = strings.TrimSpace(req.Name)
	}
	user.DateOfBirth = &dob

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

	signupFailed = false

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
