package admin

import (
	"github.com/functionfly/functionfly/internal/analytics/unified"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/services/membership"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wallet"
	"net/http"
)

// Handler contains admin handlers
type Handler struct {
	repo                   storage.Repository
	loginAttemptRepo       *storage.LoginAttemptRepository
	analyticsRepo          *storage.AnalyticsRepository
	authSvc                *auth.AuthService
	unifiedAnalytics       *unified.Service
	sfAddons               *statefabricaddons.Repository
	membershipSvc          *membership.Service
	walletService          *wallet.Service
	payoutApprovalService  *wallet.PayoutApprovalService
	reconciliationService  *wallet.ReconciliationService
	billingOperationalRepo *storage.BillingOperationalRepository
}

// NewHandler creates a new admin handler. unifiedAnalytics may be nil (tenant metrics will be placeholders).
func NewHandler(
	repo storage.Repository,
	loginAttemptRepo *storage.LoginAttemptRepository,
	analyticsRepo *storage.AnalyticsRepository,
	authSvc *auth.AuthService,
	unifiedAnalytics *unified.Service,
	sfAddons *statefabricaddons.Repository,
) *Handler {
	return &Handler{
		repo:             repo,
		loginAttemptRepo: loginAttemptRepo,
		analyticsRepo:    analyticsRepo,
		authSvc:          authSvc,
		unifiedAnalytics: unifiedAnalytics,
		sfAddons:         sfAddons,
		membershipSvc:    membership.NewService(repo),
	}
}

// SetWalletService sets the wallet service for admin wallet operations
func (h *Handler) SetWalletService(walletService *wallet.Service) {
	h.walletService = walletService
}

// SetPayoutApprovalService sets the payout approval service for admin wallet operations
func (h *Handler) SetPayoutApprovalService(svc *wallet.PayoutApprovalService) {
	h.payoutApprovalService = svc
}

// SetReconciliationService sets the reconciliation service for admin wallet operations
func (h *Handler) SetReconciliationService(svc *wallet.ReconciliationService) {
	h.reconciliationService = svc
}

// SetBillingOperationalRepository sets the billing operational repository
func (h *Handler) SetBillingOperationalRepository(repo *storage.BillingOperationalRepository) {
	h.billingOperationalRepo = repo
}

// getWalletHandler returns a WalletHandler with all services
func (h *Handler) getWalletHandler() *WalletHandler {
	// Use services from handler if available, otherwise the handler's stored services
	return NewWalletHandler(h.walletService, h.payoutApprovalService, h.reconciliationService)
}

// Wallet admin handler methods - delegated to WalletHandler

// HandleListWallets lists all wallets with filtering
func (h *Handler) HandleListWallets(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleListWallets(w, r)
}

// HandleGetWallet retrieves detailed wallet information
func (h *Handler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleGetWallet(w, r)
}

// HandleFreezeWallet freezes a wallet for security investigation
func (h *Handler) HandleFreezeWallet(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleFreezeWallet(w, r)
}

// HandleUnfreezeWallet reactivates a suspended wallet
func (h *Handler) HandleUnfreezeWallet(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleUnfreezeWallet(w, r)
}

// HandleCloseWallet permanently closes a wallet
func (h *Handler) HandleCloseWallet(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleCloseWallet(w, r)
}

// HandleAdjustWalletBalance manually adjusts a wallet balance
func (h *Handler) HandleAdjustWalletBalance(w http.ResponseWriter, r *http.Request) {
	if h.walletService == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Wallet service not available")
		return
	}
	h.getWalletHandler().HandleAdjustWalletBalance(w, r)
}

// HandleTriggerReconciliation manually triggers a reconciliation run
func (h *Handler) HandleTriggerReconciliation(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleTriggerReconciliation(w, r)
}

// HandleGetReconciliationRuns lists reconciliation runs
func (h *Handler) HandleGetReconciliationRuns(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleGetReconciliationRuns(w, r)
}

// HandleListPendingPayouts lists payout requests awaiting approval
func (h *Handler) HandleListPendingPayouts(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleListPendingPayouts(w, r)
}

// HandleApprovePayout approves a payout request
func (h *Handler) HandleApprovePayout(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleApprovePayout(w, r)
}

// HandleRejectPayout rejects a payout request
func (h *Handler) HandleRejectPayout(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleRejectPayout(w, r)
}

// HandleListPayoutApprovalRules lists payout approval rules
func (h *Handler) HandleListPayoutApprovalRules(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleListPayoutApprovalRules(w, r)
}

// HandleCreatePayoutApprovalRule creates a new payout approval rule
func (h *Handler) HandleCreatePayoutApprovalRule(w http.ResponseWriter, r *http.Request) {
	h.getWalletHandler().HandleCreatePayoutApprovalRule(w, r)
}
