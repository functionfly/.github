package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetUsernameChangeEligibility returns GET /v1/users/me/username/eligibility
// Returns information about the user's ability to change their username
func (h *Handler) HandleGetUsernameChangeEligibility(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eligibility, err := h.authSvc.CheckUsernameChangeEligibility(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to check username change eligibility")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check eligibility")
		return
	}

	writeJSON(w, http.StatusOK, eligibility)
}

// HandleGetUsernameChangeEligibilityByUsername returns GET /v1/users/{username}/username/eligibility
func (h *Handler) HandleGetUsernameChangeEligibilityByUsername(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	_, ok := h.requireSelfUsername(w, r, pathUsername)
	if !ok {
		return
	}

	claims := middleware.GetUserFromContext(r)
	eligibility, err := h.authSvc.CheckUsernameChangeEligibility(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to check username change eligibility")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check eligibility")
		return
	}

	writeJSON(w, http.StatusOK, eligibility)
}

// HandleChangeUsernameMe handles POST /v1/users/me/username/change
func (h *Handler) HandleChangeUsernameMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req auth.ChangeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get IP and User Agent
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	resp, err := h.authSvc.ChangeUsername(r.Context(), claims.UserID, req, ipAddress, userAgent)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Username change failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleChangeUsernameByUsername handles POST /v1/users/{username}/username/change
func (h *Handler) HandleChangeUsernameByUsername(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	_, ok := h.requireSelfUsername(w, r, pathUsername)
	if !ok {
		return
	}

	claims := middleware.GetUserFromContext(r)

	var req auth.ChangeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get IP and User Agent
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	resp, err := h.authSvc.ChangeUsername(r.Context(), claims.UserID, req, ipAddress, userAgent)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Username change failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetUsernameChangeHistoryMe returns GET /v1/users/me/username/history
func (h *Handler) HandleGetUsernameChangeHistoryMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	history, err := h.repo.GetUsernameChangeHistory(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to get username change history")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

// CreateUsernameChangeCheckoutRequest is the request for creating a checkout session
type CreateUsernameChangeCheckoutRequest struct {
	NewUsername string `json:"new_username"`
	SuccessURL  string `json:"success_url,omitempty"`
	CancelURL   string `json:"cancel_url,omitempty"`
}

// CreateUsernameChangeCheckoutResponse contains the checkout URL
type CreateUsernameChangeCheckoutResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	PendingID string `json:"pending_id"`
	Message   string `json:"message"`
}

// HandleCreateUsernameChangeCheckout handles POST /v1/users/me/username/checkout
// Creates a Stripe checkout session for paid username changes
func (h *Handler) HandleCreateUsernameChangeCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateUsernameChangeCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate username
	if req.NewUsername == "" {
		writeJSONError(w, http.StatusBadRequest, "new_username is required")
		return
	}

	// Get user details
	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	oldUsername := ""
	if user.Username != nil {
		oldUsername = *user.Username
	}

	// Check eligibility
	eligibility, err := h.authSvc.CheckUsernameChangeEligibility(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to check eligibility")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check eligibility")
		return
	}

	// If can change freely, no need for checkout
	if eligibility.CanChangeFreely {
		writeJSONError(w, http.StatusBadRequest, "No payment required - you have free changes available")
		return
	}

	// If cannot change even with fee, reject
	if !eligibility.CanChangeWithFee {
		writeJSONError(w, http.StatusBadRequest, eligibility.Message)
		return
	}

	// Get IP and User Agent
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	// Create pending username change record
	pending := &storage.PendingUsernameChange{
		ID:          uuid.New(),
		UserID:      claims.UserID,
		OldUsername: oldUsername,
		NewUsername: req.NewUsername,
		Status:      "pending",
		FeeCents:    eligibility.EarlyChangeFeeCents,
		FeeCurrency: "USD",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	if err := h.repo.CreatePendingUsernameChange(r.Context(), pending); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to create pending username change")
		writeJSONError(w, http.StatusInternalServerError, "Failed to initiate username change")
		return
	}

	// Get user's name for Stripe
	name := user.Name
	if name == "" {
		name = oldUsername
	}

	// Create checkout session
	checkoutResp, err := payment.CreateUsernameChangeCheckoutSession(
		r.Context(),
		h.repo,
		user.Email,
		name,
		payment.CreateUsernameChangeCheckoutSessionRequest{
			SuccessURL:      req.SuccessURL,
			CancelURL:       req.CancelURL,
			TenantID:        user.TenantID,
			UserID:          claims.UserID,
			PendingChangeID: pending.ID,
			NewUsername:     req.NewUsername,
			FeeCents:        eligibility.EarlyChangeFeeCents,
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to create checkout session")
		// Mark pending change as failed
		_ = h.repo.UpdatePendingUsernameChangeStatus(r.Context(), pending.ID, "failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	// Update pending change with checkout session ID
	pending.CheckoutSessionID = checkoutResp.SessionID
	// Note: We need to update this, but our current method doesn't support partial updates
	// For now, the webhook will look up by session ID from Stripe

	writeJSON(w, http.StatusOK, CreateUsernameChangeCheckoutResponse{
		SessionID: checkoutResp.SessionID,
		URL:       checkoutResp.URL,
		PendingID: pending.ID.String(),
		Message:   "Redirect to Stripe checkout to complete payment",
	})
}
