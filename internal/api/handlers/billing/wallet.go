package billing

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleGetWallet returns the authenticated user's registry wallet (platform fee balance).
// GET /v1/billing/wallet
func (h *Handler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.platformFees == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet is unavailable")
		return
	}

	wallet, err := h.platformFees.GetOrCreateWallet(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: get wallet failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load wallet")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":               wallet.UserID,
		"balance_usd":           wallet.BalanceUSD,
		"lifetime_earnings_usd": wallet.LifetimeEarningsUSD,
		"lifetime_fees_usd":     wallet.LifetimeFeesUSD,
	})
}

// HandleListPlatformFees returns paginated registry platform fee history for the current user.
// GET /v1/billing/fees
func (h *Handler) HandleListPlatformFees(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.platformFees == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Fee history is unavailable")
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	fees, total, err := h.platformFees.ListPlatformFeesByUserPaged(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: list platform fees failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load fee history")
		return
	}

	out := make([]map[string]interface{}, 0, len(fees))
	for _, f := range fees {
		out = append(out, map[string]interface{}{
			"id":          f.ID.String(),
			"user_id":     f.UserID.String(),
			"fee_type":    f.FeeType,
			"amount_usd":  f.AmountUSD,
			"description": platformFeeDescription(f.FeeType),
			"created_at":  f.ChargedAt.UTC().Format(time.RFC3339Nano),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"fees":   out,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}

func platformFeeDescription(feeType string) string {
	switch feeType {
	case storageregistry.FeeTypePublish:
		return "Registry publish fee"
	case storageregistry.FeeTypeVersionUpdate:
		return "Registry version update fee"
	case storageregistry.FeeTypeCommission:
		return "Platform commission"
	default:
		if feeType == "" {
			return "Platform fee"
		}
		return feeType
	}
}

type walletTopUpRequest struct {
	AmountUSD   float64 `json:"amount_usd"`
	SuccessURL  string  `json:"success_url"`
	CancelURL   string  `json:"cancel_url"`
}

// HandleWalletTopUp creates a Stripe Checkout session to add funds to the registry wallet.
// POST /v1/billing/wallet/top-up
func (h *Handler) HandleWalletTopUp(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.platformFees == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet is unavailable")
		return
	}
	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	var req walletTopUpRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AmountUSD < payment.MinRegistryWalletTopUpUSD {
		writeJSONError(w, http.StatusBadRequest, "Invalid amount. Minimum top-up is $1.00.")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing wallet top-up: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	name := user.Email
	if user.Username != nil && *user.Username != "" {
		name = *user.Username
	}

	result, err := payment.CreateRegistryWalletCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		claims.UserID,
		user.Email,
		name,
		req.AmountUSD,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: registry wallet checkout failed")
		msg := "Could not create checkout session"
		if err.Error() != "" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusBadRequest, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"checkout_url": result.URL,
		"session_id":   result.SessionID,
	})
}
