package auth

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
	"unicode"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

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
func (a *AuthService) Login(email, password string) (res *LoginResponse, err error) {
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

	// Build safe user for response (no password hash, MFA secrets, or verification tokens)
	loginUser := userToLoginUser(user)

	return &LoginResponse{
		Token: token,
		User:  loginUser,
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
	}
	if u.CompanyName != nil && *u.CompanyName != "" {
		lu.CompanyName = *u.CompanyName
	}
	if u.ProviderData != nil {
		lu.ProviderData = u.ProviderData
		if n, ok := u.ProviderData["name"].(string); ok {
			lu.Name = n
		}
		if fn, ok := u.ProviderData["full_name"].(string); ok && lu.Name == "" {
			lu.Name = fn
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
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
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

	return &SignupResponse{
		Message:              "Account created successfully. Please check your email to verify your account.",
		EmailSent:            emailSent,
		RequiresVerification: true,
	}, nil
}
