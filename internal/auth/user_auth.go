package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

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

	// #region agent log
	payloadC := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "C", "location": "user_auth.go:Login entry", "message": "Login entered", "data": map[string]interface{}{"email": email}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadC)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadC).Info("agent log")
	// #endregion

	// Get user by email
	user, err := a.repo.GetUserByEmail(email)

	// #region agent log
	getUserErrStr := ""
	if err != nil {
		getUserErrStr = err.Error()
	}
	payloadD := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "D", "location": "user_auth.go:after GetUserByEmail", "message": "after GetUserByEmail", "data": map[string]interface{}{"err": getUserErrStr, "userNil": user == nil}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadD)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadD).Info("agent log")
	// #endregion

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

	// #region agent log
	tokenErrStr := ""
	if err != nil {
		tokenErrStr = err.Error()
	}
	payloadE := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "E", "location": "user_auth.go:after generateToken", "message": "after generateToken", "data": map[string]interface{}{"err": tokenErrStr, "hasToken": token != ""}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadE)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadE).Info("agent log")
	// #endregion

	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Build safe user for response (no password hash, MFA secrets, or verification tokens)
	loginUser := userToLoginUser(user)

	// #region agent log
	payloadF := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "F", "location": "user_auth.go:Login success", "message": "Login success path", "data": map[string]interface{}{}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadF)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadF).Info("agent log")
	// #endregion

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
