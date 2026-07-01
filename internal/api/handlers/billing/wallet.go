package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// HandleGetWallet returns the authenticated user's registry wallet (platform fee balance).
// GET /v1/billing/wallet
// After unified wallet migration, this reads from the wallets table via walletService.
func (h *Handler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Try unified wallet service first, fall back to legacy platformFees
	if h.walletService != nil {
		walletInfo, err := h.walletService.GetUserWallet(r.Context(), claims.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: get wallet failed")
			writeJSONError(w, http.StatusInternalServerError, "Failed to load wallet")
			return
		}
		if walletInfo == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user_id":               claims.UserID,
				"balance_usd":           float64(0),
				"lifetime_earnings_usd": float64(0),
				"lifetime_fees_usd":     float64(0),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":               walletInfo.UserID,
			"balance_usd":           walletInfo.BalanceUSD,
			"lifetime_earnings_usd": walletInfo.LifetimeEarningsUSD,
			"lifetime_fees_usd":     walletInfo.LifetimeSpentUSD,
		})
		return
	}

	// Legacy fallback: use platformFees repository
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

// HandleListWalletTransactions returns paginated wallet transaction history for the current user.
// GET /v1/billing/wallet/transactions
func (h *Handler) HandleListWalletTransactions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet is unavailable")
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

	walletObj, err := h.walletService.GetUserWallet(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: get wallet for transactions failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load wallet")
		return
	}
	if walletObj == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"transactions": []interface{}{},
			"total":        0,
		})
		return
	}

	transactions, total, err := h.walletService.GetTransactionHistory(r.Context(), walletObj.ID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("wallet_id", walletObj.ID).Warn("billing: list wallet transactions failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load transactions")
		return
	}

	type transactionJSON struct {
		ID              string     `json:"id"`
		Type            string     `json:"type"`
		AmountUSD       float64    `json:"amount"`
		Reference       string     `json:"description"`
		CreatedAt       time.Time  `json:"timestamp"`
		Status          string     `json:"status"`
		TransactionType string     `json:"transaction_type,omitempty"`
		BalanceBefore   float64    `json:"balance_before,omitempty"`
		BalanceAfter    float64    `json:"balance_after,omitempty"`
		FeeType         *string    `json:"fee_type,omitempty"`
		CompletedAt     *time.Time `json:"completed_at,omitempty"`
	}

	out := make([]transactionJSON, 0, len(transactions))
	for _, tx := range transactions {
		desc := ""
		if tx.Reference != nil {
			desc = *tx.Reference
		}
		out = append(out, transactionJSON{
			ID:              tx.ID.String(),
			Type:            tx.TransactionType,
			AmountUSD:       tx.AmountUSD,
			Reference:       desc,
			CreatedAt:       tx.CreatedAt,
			Status:          tx.Status,
			TransactionType: tx.TransactionType,
			BalanceBefore:   tx.BalanceBeforeUSD,
			BalanceAfter:    tx.BalanceAfterUSD,
			FeeType:         tx.FeeType,
			CompletedAt:     tx.CompletedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": out,
		"total":        total,
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
	AmountUSD  float64 `json:"amount_usd"`
	SuccessURL string  `json:"success_url"`
	CancelURL  string  `json:"cancel_url"`
}

// HandleWalletTopUp creates a Stripe Checkout session to add funds to the registry wallet.
// POST /v1/billing/wallet/top-up
// Rate limited: Max 5 top-up attempts per hour per user
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

	// REDIS RATE LIMITING: Check if user has exceeded top-up limits
	// Max 5 attempts per hour per user with exponential backoff on failures
	if h.redisClient != nil {
		ctx := r.Context()
		userKey := fmt.Sprintf("wallet_topup:user:%s", claims.UserID.String())
		tenantUserKey := fmt.Sprintf("wallet_topup:tenant:%s:user:%s", claims.TenantID.String(), claims.UserID.String())

		// Check hourly limit (5 attempts per hour)
		allowed, retryAfter, err := h.checkRateLimit(ctx, userKey, 5, time.Hour)
		if err != nil {
			logrus.WithError(err).WithField("user_id", claims.UserID).Warn("wallet top-up rate limit check failed")
			// Continue anyway - don't block on Redis errors
		} else if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"code":    "WALLET_TOPUP_RATE_LIMIT",
				"message": "Too many wallet top-up attempts. Maximum 5 attempts per hour allowed. Please try again later.",
			})
			logrus.WithFields(logrus.Fields{
				"user_id":     claims.UserID,
				"retry_after": retryAfter.Seconds(),
			}).Warn("Wallet top-up rate limit exceeded")
			return
		}

		// Check for exponential backoff on failed attempts
		backoffAllowed, backoffWait, err := h.checkExponentialBackoff(ctx, tenantUserKey)
		if err != nil {
			logrus.WithError(err).WithField("user_id", claims.UserID).Warn("wallet top-up backoff check failed")
		} else if !backoffAllowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(backoffWait.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "backoff_required",
				"code":    "WALLET_TOPUP_BACKOFF",
				"message": "Too many failed top-up attempts. Please wait before trying again.",
			})
			return
		}
	}

	var req walletTopUpRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AmountUSD < payment.MinRegistryWalletTopUpUSD {
		writeJSONError(w, http.StatusBadRequest, "Invalid amount. Minimum top-up is $1.00.")
		return
	}
	if req.AmountUSD > payment.MaxRegistryWalletTopUpUSD {
		writeJSONError(w, http.StatusBadRequest, "Invalid amount. Maximum top-up is $10,000.00.")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
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

// checkRateLimit checks if the user has exceeded the rate limit for wallet top-ups
// Returns (allowed bool, retryAfter duration, error)
func (h *Handler) checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	pipe := h.redisClient.Pipeline()
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count entries in current window
	countCmd := pipe.ZCard(ctx, key)

	// Add current attempt
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// Set expiration on the key
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, 0, err // Allow on error
	}

	count := countCmd.Val()
	if count >= int64(limit) {
		// Get the oldest entry to calculate retry after
		oldestCmd := h.redisClient.ZRangeWithScores(ctx, key, 0, 0)
		if len(oldestCmd.Val()) > 0 {
			oldest := oldestCmd.Val()[0].Score
			retryAfter := time.Duration(int64(oldest)+int64(window.Seconds())-now) * time.Second
			if retryAfter < 0 {
				retryAfter = window
			}
			return false, retryAfter, nil
		}
		return false, window, nil
	}

	return true, 0, nil
}

// checkExponentialBackoff checks if the user should wait due to failed attempts
// Returns (allowed bool, wait duration, error)
func (h *Handler) checkExponentialBackoff(ctx context.Context, key string) (bool, time.Duration, error) {
	// Key for tracking failed attempts
	failureKey := key + ":failures"

	// Get recent failure count
	failures, err := h.redisClient.Get(ctx, failureKey).Int()
	if err == redis.Nil {
		return true, 0, nil // No failures, allowed
	}
	if err != nil {
		return true, 0, err // Allow on error
	}

	if failures == 0 {
		return true, 0, nil
	}

	// Calculate backoff: 2^n minutes (2, 4, 8, 16, 32... minutes)
	backoffMinutes := 1 << uint(failures)
	if backoffMinutes > 60 {
		backoffMinutes = 60 // Max 1 hour
	}

	// Check last failure timestamp
	lastFailureKey := key + ":last_failure"
	lastFailureStr, err := h.redisClient.Get(ctx, lastFailureKey).Result()
	if err == redis.Nil {
		return true, 0, nil // No last failure recorded
	}
	if err != nil {
		return true, 0, err
	}

	lastFailure, err := strconv.ParseInt(lastFailureStr, 10, 64)
	if err != nil {
		return true, 0, err
	}

	elapsed := time.Since(time.Unix(lastFailure, 0))
	requiredWait := time.Duration(backoffMinutes) * time.Minute

	if elapsed < requiredWait {
		return false, requiredWait - elapsed, nil
	}

	return true, 0, nil
}

// recordFailedAttempt records a failed top-up attempt for exponential backoff
func (h *Handler) recordFailedAttempt(ctx context.Context, key string) error {
	failureKey := key + ":failures"
	lastFailureKey := key + ":last_failure"

	pipe := h.redisClient.Pipeline()
	pipe.Incr(ctx, failureKey)
	pipe.Set(ctx, lastFailureKey, fmt.Sprintf("%d", time.Now().Unix()), 24*time.Hour)
	pipe.Expire(ctx, failureKey, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}
