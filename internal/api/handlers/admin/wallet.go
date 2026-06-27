package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WalletHandler handles admin wallet management operations
type WalletHandler struct {
	walletService         *wallet.Service
	payoutApprovalService *wallet.PayoutApprovalService
	reconciliationService *wallet.ReconciliationService
	logger                *logrus.Logger
}

// NewWalletHandler creates a new wallet admin handler
func NewWalletHandler(walletService *wallet.Service, payoutApprovalSvc *wallet.PayoutApprovalService, reconciliationSvc *wallet.ReconciliationService) *WalletHandler {
	return &WalletHandler{
		walletService:         walletService,
		payoutApprovalService: payoutApprovalSvc,
		reconciliationService: reconciliationSvc,
		logger:                logrus.New(),
	}
}

// SetLogger sets the logger
func (h *WalletHandler) SetLogger(logger *logrus.Logger) {
	h.logger = logger
}

// ============================================
// Wallet Freeze/Suspend Operations
// ============================================

// HandleFreezeWallet freezes a wallet for security investigation
// POST /v1/admin/wallets/{walletId}/freeze
func (h *WalletHandler) HandleFreezeWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	// Get admin user from context
	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Reason   string `json:"reason"`
		Duration *int   `json:"duration_hours,omitempty"` // Optional auto-unfreeze duration
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Reason == "" {
		writeJSONError(w, http.StatusBadRequest, "Reason is required")
		return
	}

	ctx := r.Context()

	// Freeze the wallet
	freezeReason := fmt.Sprintf("Frozen by admin %s: %s", adminUser.UserID, req.Reason)
	if err := h.walletService.SuspendWallet(ctx, walletID, freezeReason); err != nil {
		h.logger.WithError(err).WithField("wallet_id", walletID).Error("Failed to freeze wallet")
		writeJSONError(w, http.StatusInternalServerError, "Failed to freeze wallet")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"wallet_id": walletID,
		"admin_id":  adminUser.UserID,
		"reason":    req.Reason,
	}).Info("Wallet frozen by admin")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":                true,
		"wallet_id":              walletID.String(),
		"status":                 "suspended",
		"reason":                 req.Reason,
		"frozen_by":              adminUser.UserID,
		"frozen_at":              time.Now().Format(time.RFC3339),
		"auto_unfreeze_in_hours": req.Duration,
	})
}

// HandleUnfreezeWallet reactivates a suspended wallet
// POST /v1/admin/wallets/{walletId}/unfreeze
func (h *WalletHandler) HandleUnfreezeWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ctx := r.Context()

	if err := h.walletService.ReactivateWallet(ctx, walletID); err != nil {
		h.logger.WithError(err).WithField("wallet_id", walletID).Error("Failed to unfreeze wallet")
		writeJSONError(w, http.StatusInternalServerError, "Failed to unfreeze wallet")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"wallet_id": walletID,
		"admin_id":  adminUser.UserID,
		"reason":    req.Reason,
	}).Info("Wallet unfrozen by admin")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"wallet_id":   walletID.String(),
		"status":      "active",
		"unfrozen_by": adminUser.UserID,
		"unfrozen_at": time.Now().Format(time.RFC3339),
	})
}

// HandleCloseWallet permanently closes a wallet
// POST /v1/admin/wallets/{walletId}/close
func (h *WalletHandler) HandleCloseWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Reason == "" {
		writeJSONError(w, http.StatusBadRequest, "Reason is required")
		return
	}

	ctx := r.Context()

	closeReason := fmt.Sprintf("Closed by admin %s: %s", adminUser.UserID, req.Reason)
	if err := h.walletService.CloseWallet(ctx, walletID, closeReason); err != nil {
		h.logger.WithError(err).WithField("wallet_id", walletID).Error("Failed to close wallet")
		writeJSONError(w, http.StatusInternalServerError, "Failed to close wallet")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"wallet_id": walletID,
		"admin_id":  adminUser.UserID,
		"reason":    req.Reason,
	}).Warn("Wallet closed by admin")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"wallet_id": walletID.String(),
		"status":    "closed",
		"reason":    req.Reason,
		"closed_by": adminUser.UserID,
		"closed_at": time.Now().Format(time.RFC3339),
	})
}

// ============================================
// Wallet Listing and Search
// ============================================

// HandleListWallets lists all wallets with filtering
// GET /v1/admin/wallets
func (h *WalletHandler) HandleListWallets(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	ownerType := r.URL.Query().Get("owner_type")
	status := r.URL.Query().Get("status")
	minBalance := r.URL.Query().Get("min_balance")
	maxBalance := r.URL.Query().Get("max_balance")

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	ctx := r.Context()

	// Build filter
	filter := wallet.WalletFilter{
		OwnerType: ownerType,
		Status:    status,
		Limit:     limit,
		Offset:    offset,
	}

	if minBalance != "" {
		if min, err := strconv.ParseFloat(minBalance, 64); err == nil {
			filter.MinBalance = &min
		}
	}

	if maxBalance != "" {
		if max, err := strconv.ParseFloat(maxBalance, 64); err == nil {
			filter.MaxBalance = &max
		}
	}

	result, err := h.walletService.ListWallets(ctx, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list wallets")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list wallets")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallets": result.Wallets,
		"total":   result.Total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetWallet retrieves detailed wallet information
// GET /v1/admin/wallets/{walletId}
func (h *WalletHandler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	ctx := r.Context()

	wlt, err := h.walletService.GetWallet(ctx, walletID)
	if err != nil {
		h.logger.WithError(err).WithField("wallet_id", walletID).Error("Failed to get wallet")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get wallet")
		return
	}

	if wlt == nil {
		writeJSONError(w, http.StatusNotFound, "Wallet not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wlt)
}

// ============================================
// Wallet Balance Adjustment (Admin)
// ============================================

// HandleAdjustWalletBalance manually adjusts a wallet balance (emergency use)
// POST /v1/admin/wallets/{walletId}/adjust
func (h *WalletHandler) HandleAdjustWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Amount    float64 `json:"amount_usd"` // Positive for credit, negative for debit
		Reason    string  `json:"reason"`
		Reference string  `json:"reference,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Amount == 0 {
		writeJSONError(w, http.StatusBadRequest, "Amount cannot be zero")
		return
	}

	if req.Reason == "" {
		writeJSONError(w, http.StatusBadRequest, "Reason is required")
		return
	}

	ctx := r.Context()

	// Get wallet to determine owner
	wlt, err := h.walletService.GetWallet(ctx, walletID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get wallet")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get wallet")
		return
	}
	if wlt == nil {
		writeJSONError(w, http.StatusNotFound, "Wallet not found")
		return
	}

	// Check adjustment limits
	amount := req.Amount
	if amount < 0 {
		amount = -amount // Use absolute value for limit checks
	}

	limits := wallet.GetAdminAdjustmentLimits()
	if amount > limits.SingleOperationMax {
		// Check if secondary approval is required
		if amount > limits.RequiresSecondaryApprovalAbove {
			writeJSONError(w, http.StatusForbidden, fmt.Sprintf("Amount $%.2f exceeds maximum allowed and requires secondary approval above $%.2f",
				amount, limits.RequiresSecondaryApprovalAbove))
			return
		}
		// Log warning for large adjustment requiring secondary approval
		h.logger.WithFields(logrus.Fields{
			"wallet_id": walletID,
			"admin_id":  adminUser.UserID,
			"amount":    req.Amount,
			"limit":     limits.SingleOperationMax,
			"threshold": limits.RequiresSecondaryApprovalAbove,
		}).Warn("Large admin adjustment requires secondary approval")
	}

	// Check if this exceeds alert threshold and should trigger monitoring
	if amount > limits.AlertThreshold {
		h.logger.WithFields(logrus.Fields{
			"wallet_id":  walletID,
			"admin_id":   adminUser.UserID,
			"amount":     req.Amount,
			"threshold":  limits.AlertThreshold,
			"alert_type": "large_admin_adjustment",
		}).Warn("ALERT: Large admin balance adjustment detected")
	}

	// Perform adjustment using admin credit/debit
	var update *wallet.BalanceUpdate
	if req.Amount > 0 {
		update, err = h.walletService.AdminCredit(ctx, walletID, req.Amount, req.Reference, req.Reason, adminUser.UserID)
	} else {
		update, err = h.walletService.AdminDebit(ctx, walletID, -req.Amount, req.Reference, req.Reason, adminUser.UserID)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to adjust wallet balance")
		writeJSONError(w, http.StatusInternalServerError, "Failed to adjust wallet balance")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"wallet_id":      walletID,
		"admin_id":       adminUser.UserID,
		"amount":         req.Amount,
		"new_balance":    update.CurrentBalance,
		"transaction_id": update.TransactionID,
	}).Warn("Wallet balance adjusted by admin")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":              true,
		"wallet_id":            walletID.String(),
		"adjustment_usd":       req.Amount,
		"new_balance_usd":      update.CurrentBalance,
		"previous_balance_usd": update.PreviousBalance,
		"transaction_id":       update.TransactionID.String(),
		"adjusted_by":          adminUser.UserID,
		"adjusted_at":          time.Now().Format(time.RFC3339),
	})
}

// ============================================
// Wallet Reconciliation Admin Operations
// ============================================

// HandleTriggerReconciliation manually triggers a reconciliation run
// POST /v1/admin/wallets/reconciliation
func (h *WalletHandler) HandleTriggerReconciliation(w http.ResponseWriter, r *http.Request) {
	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.reconciliationService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Reconciliation service not available")
		return
	}

	ctx := r.Context()

	// Run reconciliation
	run, err := h.reconciliationService.RunFullReconciliation(ctx, "manual", &adminUser.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to run reconciliation")
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Reconciliation failed: %v", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"run_id":   run.ID,
		"admin_id": adminUser.UserID,
		"status":   run.Status,
	}).Info("Manual reconciliation triggered")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"run_id":       run.ID.String(),
		"status":       run.Status,
		"triggered_by": adminUser.UserID.String(),
		"triggered_at": run.CreatedAt.Format(time.RFC3339),
		"message":      "Reconciliation run started successfully",
	})
}

// HandleGetReconciliationRuns lists reconciliation runs
// GET /v1/admin/wallets/reconciliation/runs
func (h *WalletHandler) HandleGetReconciliationRuns(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Reconciliation service not available")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	ctx := r.Context()

	runs, total, err := h.reconciliationService.ListReconciliationRuns(ctx, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list reconciliation runs")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list reconciliation runs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"runs":   runs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ============================================
// Payout Approval Operations
// ============================================

// HandleListPendingPayouts lists payout requests awaiting approval
// GET /v1/admin/payouts/pending
func (h *WalletHandler) HandleListPendingPayouts(w http.ResponseWriter, r *http.Request) {
	if h.payoutApprovalService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Payout approval service not available")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	ctx := r.Context()

	// Get pending approvals
	records, total, err := h.payoutApprovalService.GetPendingApprovals(ctx, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list pending payouts")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list pending payouts")
		return
	}

	// Also get summary
	summary, err := h.payoutApprovalService.GetApprovalSummary(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get approval summary")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payouts": records,
		"summary": summary,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleApprovePayout approves a payout request
// POST /v1/admin/payouts/{payoutId}/approve
func (h *WalletHandler) HandleApprovePayout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["payoutId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.payoutApprovalService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Payout approval service not available")
		return
	}

	var req struct {
		Notes string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional, so this is not a fatal error
		req.Notes = ""
	}

	ctx := r.Context()

	// Get client IP and user agent for audit
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	userAgent := r.Header.Get("User-Agent")

	// Get the approval record for this payout
	// We need to look up by payout request ID
	record, err := h.payoutApprovalService.GetApprovalRecordByPayoutRequestID(ctx, payoutID)
	if err != nil {
		h.logger.WithError(err).WithField("payout_id", payoutID).Error("Failed to get approval record")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get approval record")
		return
	}

	if record == nil {
		writeJSONError(w, http.StatusNotFound, "Payout approval record not found")
		return
	}

	// Check if this is first or second approval
	var approveErr error
	if record.Status == wallet.PayoutApprovalStatusPending {
		approveErr = h.payoutApprovalService.FirstApprove(ctx, record.ID, adminUser.UserID, req.Notes, &clientIP, &userAgent)
	} else if record.Status == wallet.PayoutApprovalStatusFirstApproved {
		approveErr = h.payoutApprovalService.SecondApprove(ctx, record.ID, adminUser.UserID, req.Notes, &clientIP, &userAgent)
	} else {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Payout cannot be approved - current status: %s", record.Status))
		return
	}

	if approveErr != nil {
		h.logger.WithError(approveErr).WithField("payout_id", payoutID).Error("Failed to approve payout")
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to approve payout: %v", approveErr))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"payout_id": payoutID,
		"admin_id":  adminUser.UserID,
		"record_id": record.ID,
	}).Info("Payout approved by admin")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"payout_id":   payoutID.String(),
		"approved_by": adminUser.UserID.String(),
		"approved_at": time.Now().Format(time.RFC3339),
		"notes":       req.Notes,
	})
}

// HandleRejectPayout rejects a payout request
// POST /v1/admin/payouts/{payoutId}/reject
func (h *WalletHandler) HandleRejectPayout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["payoutId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.payoutApprovalService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Payout approval service not available")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Reason == "" {
		writeJSONError(w, http.StatusBadRequest, "Rejection reason is required")
		return
	}

	ctx := r.Context()

	// Get client IP and user agent for audit
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	userAgent := r.Header.Get("User-Agent")

	// Get the approval record for this payout
	record, err := h.payoutApprovalService.GetApprovalRecordByPayoutRequestID(ctx, payoutID)
	if err != nil {
		h.logger.WithError(err).WithField("payout_id", payoutID).Error("Failed to get approval record")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get approval record")
		return
	}

	if record == nil {
		writeJSONError(w, http.StatusNotFound, "Payout approval record not found")
		return
	}

	// Reject the payout
	if err := h.payoutApprovalService.Reject(ctx, record.ID, adminUser.UserID, req.Reason, &clientIP, &userAgent); err != nil {
		h.logger.WithError(err).WithField("payout_id", payoutID).Error("Failed to reject payout")
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reject payout: %v", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"payout_id": payoutID,
		"admin_id":  adminUser.UserID,
		"record_id": record.ID,
		"reason":    req.Reason,
	}).Warn("Payout rejected by admin")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"payout_id":   payoutID.String(),
		"rejected_by": adminUser.UserID.String(),
		"rejected_at": time.Now().Format(time.RFC3339),
		"reason":      req.Reason,
	})
}

// HandleListPayoutApprovalRules lists payout approval rules
// GET /v1/admin/payouts/approval-rules
func (h *WalletHandler) HandleListPayoutApprovalRules(w http.ResponseWriter, r *http.Request) {
	if h.payoutApprovalService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Payout approval service not available")
		return
	}

	ctx := r.Context()

	rules, err := h.payoutApprovalService.ListApprovalRules(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list approval rules")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list approval rules")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
	})
}

// HandleCreatePayoutApprovalRule creates a new payout approval rule
// POST /v1/admin/payouts/approval-rules
func (h *WalletHandler) HandleCreatePayoutApprovalRule(w http.ResponseWriter, r *http.Request) {
	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.payoutApprovalService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Payout approval service not available")
		return
	}

	var req wallet.PayoutApprovalRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "Rule name is required")
		return
	}

	ctx := r.Context()

	rule, err := h.payoutApprovalService.CreateApprovalRule(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create approval rule")
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create rule: %v", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"rule_id":  rule.ID,
		"admin_id": adminUser.UserID,
	}).Info("Payout approval rule created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    rule,
	})
}

// ============================================
// Helper Functions
// ============================================

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(statusCode),
			"message": message,
		},
	})
}

// Permission helpers for wallet operations
func WalletFreezePermission(authMW *middleware.AuthMiddleware) func(http.HandlerFunc) http.HandlerFunc {
	return authMW.RequirePermission(auth.PermBillingWrite)
}

func WalletReadPermission(authMW *middleware.AuthMiddleware) func(http.HandlerFunc) http.HandlerFunc {
	return authMW.RequirePermission(auth.PermBillingRead)
}
