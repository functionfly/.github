package payouts

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// Handler handles payout-related API endpoints.
type Handler struct {
	payoutService *payment.PayoutService
	payoutRepo    *storage.PayoutRepository
	repo          storage.Repository
}

// NewHandler creates a new payout handler.
func NewHandler(payoutService *payment.PayoutService, payoutRepo *storage.PayoutRepository, repo storage.Repository) *Handler {
	return &Handler{
		payoutService: payoutService,
		payoutRepo:    payoutRepo,
		repo:          repo,
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSONErrorWithCode(w, status, msg, "")
}

func writeJSONErrorWithCode(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]string{"error": msg}
	if code != "" {
		body["code"] = code
	}
	_ = json.NewEncoder(w).Encode(body)
}

// HandleGetConnectAccount returns the user's Stripe Connect account status.
// GET /v1/payouts/connect-account
func (h *Handler) HandleGetConnectAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status, err := h.payoutService.GetConnectAccountStatus(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to get connect account status")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve account status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

// StartOnboardingRequest is the request body for starting onboarding.
type StartOnboardingRequest struct {
	// Empty — email is pulled from the authenticated user.
}

// HandleStartOnboarding creates a Stripe Express connected account and returns an onboarding link.
// POST /v1/payouts/connect-account
func (h *Handler) HandleStartOnboarding(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Payouts are not configured")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("payouts: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	result, err := h.payoutService.StartOnboarding(r.Context(), claims.UserID, user.Email)
	if err != nil {
		if errors.Is(err, payment.ErrConnectPlatformNotReady) {
			logrus.WithField("user_id", claims.UserID).Warn("payouts: Stripe Connect not enabled on platform account")
			writeJSONErrorWithCode(
				w,
				http.StatusServiceUnavailable,
				"Publisher payouts are not available yet. Please try again later or contact support if this continues.",
				"CONNECT_PLATFORM_NOT_READY",
			)
			return
		}
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to start onboarding")
		writeJSONError(w, http.StatusBadRequest, "Failed to start payout onboarding")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// RefreshConnectAccount syncs the account status from Stripe.
// POST /v1/payouts/connect-account/refresh
func (h *Handler) HandleRefreshConnectAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Payouts are not configured")
		return
	}

	account, err := h.payoutRepo.GetConnectAccountByUserID(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to get connect account")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}
	if account == nil {
		writeJSONError(w, http.StatusNotFound, "No connected account found")
		return
	}

	if err := h.payoutService.RefreshAccountStatus(r.Context(), account.StripeAccountID); err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to refresh account status")
		writeJSONError(w, http.StatusInternalServerError, "Failed to refresh account status")
		return
	}

	// Return updated status
	status, err := h.payoutService.GetConnectAccountStatus(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("payouts: failed to get updated status")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve updated status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

// PayoutRequestPayload is the request body for requesting a payout.
type PayoutRequestPayload struct {
	AmountCents    int    `json:"amount_cents"`
	IdempotencyKey string `json:"idempotency_key"`
}

// HandleRequestPayout creates a payout request.
// POST /v1/payouts/request
func (h *Handler) HandleRequestPayout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Payouts are not configured")
		return
	}

	var payload PayoutRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.AmountCents <= 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_cents must be positive")
		return
	}

	if payload.IdempotencyKey == "" {
		writeJSONError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}

	result, err := h.payoutService.RequestPayout(r.Context(), claims.UserID, payload.AmountCents, payload.IdempotencyKey)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":      claims.UserID,
			"amount_cents": payload.AmountCents,
		}).Error("payouts: failed to request payout")
		msg := "Payout request failed"
		if err.Error() != "" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusBadRequest, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// HandleGetBalance returns the user's payout balance.
// GET /v1/payouts/balance
func (h *Handler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	balance, err := h.payoutRepo.GetPayoutBalance(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to get balance")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve balance")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(balance)
}

// HandleListPayoutRequests returns the user's payout requests.
// GET /v1/payouts/requests
func (h *Handler) HandleListPayoutRequests(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	requests, total, err := h.payoutRepo.GetPayoutRequestsByUserID(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to list requests")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve payout requests")
		return
	}

	out := make([]map[string]interface{}, 0, len(requests))
	for _, req := range requests {
		item := map[string]interface{}{
			"id":              req.ID.String(),
			"amount_cents":    req.AmountCents,
			"currency":        req.Currency,
			"status":          req.Status,
			"idempotency_key": req.IdempotencyKey,
			"created_at":      req.CreatedAt,
			"updated_at":      req.UpdatedAt,
		}
		if req.StripeTransferID != nil {
			item["stripe_transfer_id"] = *req.StripeTransferID
		}
		if req.StripePayoutID != nil {
			item["stripe_payout_id"] = *req.StripePayoutID
		}
		if req.FailureReason != nil {
			item["failure_reason"] = *req.FailureReason
		}
		out = append(out, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": out,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleListLedgerEntries returns the user's payout ledger.
// GET /v1/payouts/ledger
func (h *Handler) HandleListLedgerEntries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	entries, total, err := h.payoutRepo.ListLedgerEntries(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to list ledger entries")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve ledger")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
