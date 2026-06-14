package auth

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// HandleSignup handles user registration
func (h *Handler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req auth.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		writeJSONError(w, http.StatusBadRequest, "Email, password, and username are required")
		return
	}

	if req.Password != req.ConfirmPassword {
		writeJSONError(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	if !req.TermsAccepted {
		writeJSONError(w, http.StatusBadRequest, "Terms must be accepted")
		return
	}

	response, err := h.authSvc.Signup(r.Context(), req)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Signup failed")

		authEvent := &storage.AuthEvent{
			EventType:     "signup",
			Success:       false,
			FailureReason: stringPtr("signup_failed"),
			IPAddress:     getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
			Metadata: map[string]interface{}{
				"email":    req.Email,
				"username": req.Username,
			},
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(r.Context(), authEvent); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log signup failure")
		}

		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	authEvent := &storage.AuthEvent{
		EventType: "signup",
		Success:   true,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Metadata: map[string]interface{}{
			"email":                 req.Email,
			"username":              req.Username,
			"email_sent":            response.EmailSent,
			"requires_verification": response.RequiresVerification,
		},
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(r.Context(), authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log signup success")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// HandleSignupConfig returns public signup UI flags (e.g. invite-only mode).
func (h *Handler) HandleSignupConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(auth.SignupConfigResponse{
		InviteRequired: auth.SignupInviteRequired(),
	})
}

// HandleWaitlist handles POST /waitlist (public — no auth required).
func (h *Handler) HandleWaitlist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	var req struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Company string `json:"company"`
		UseCase string `json:"useCase"`
		Source  string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeJSONError(w, http.StatusBadRequest, "A valid email address is required")
		return
	}

	ipAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ipAddr = host
	}

	entry, err := h.authSvc.Repo().CreateWaitlistEntry(r.Context(), req.Email, req.Name, req.Company, req.UseCase, req.Source, ipAddr, r.Header.Get("User-Agent"))
	if err != nil {
		if err == storage.ErrWaitlistEntryExists {
			writeJSONError(w, http.StatusConflict, "This email is already on the waitlist")
			return
		}
		logrus.WithError(err).Error("HandleWaitlist: CreateWaitlistEntry failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to join waitlist")
		return
	}

	_ = h.authSvc.EmailService().SendWaitlistConfirmationEmail(entry.Email)

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Successfully joined the waitlist",
	})
}

// HandleCheckInviteCode validates an invite code without consuming a use.
func (h *Handler) HandleCheckInviteCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))

	if auth.SignupInviteRequired() && code == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
			"error": "invite code is required",
		})
		return
	}

	if !auth.SignupInviteRequired() && code == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": true,
		})
		return
	}

	err := h.authSvc.Repo().ValidateSignupInviteReadOnly(r.Context(), code)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": true,
		})
		return
	}

	var msg string
	switch {
	case errors.Is(err, storage.ErrSignupInviteInvalid):
		msg = "invalid or expired invite code"
	case errors.Is(err, storage.ErrSignupInviteExhausted):
		msg = "this invite code has no uses remaining"
	case errors.Is(err, storage.ErrSignupInviteRevoked):
		msg = "this invite code is no longer valid"
	default:
		msg = "could not validate invite code"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid": false,
		"error": msg,
	})
}

// HandleVerifyEmail handles email verification
func (h *Handler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	user, userErr := h.authSvc.Repo().GetUserByVerificationToken(r.Context(), token)
	if userErr != nil {
		logrus.WithError(userErr).Warn("Failed to look up user by verification token")
	}

	err := h.authSvc.VerifyEmail(r.Context(), token)
	if err != nil {
		logrus.WithError(err).Warn("Email verification failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var accessToken string
	if user != nil {
		if t, err := h.authSvc.GenerateToken(user); err == nil {
			accessToken = t
		}
	}

	response := map[string]interface{}{
		"message": "Email verified successfully! You can now log in to your account.",
		"token":   accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleResendVerification handles resending verification emails
func (h *Handler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	err := h.authSvc.ResendVerificationEmail(r.Context(), req.Email)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Resend verification failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "If that email address is registered and unverified, a verification email has been sent.",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleCheckUsernameAvailability checks if a username is available for registration
func (h *Handler) HandleCheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "Username parameter is required")
		return
	}

	available, err := h.authSvc.CheckUsernameAvailability(r.Context(), username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Warn("Username availability check failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]interface{}{
		"available": available,
		"username":  username,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
