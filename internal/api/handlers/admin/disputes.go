package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DisputesHandler handles admin dispute management endpoints
type DisputesHandler struct {
	disputeRepo *storage.DisputeRepository
	refundRepo  *storage.RefundRepository
	userRepo    storage.Repository
}

// NewDisputesHandler creates a new disputes handler
func NewDisputesHandler(disputeRepo *storage.DisputeRepository, refundRepo *storage.RefundRepository, userRepo storage.Repository) *DisputesHandler {
	return &DisputesHandler{
		disputeRepo: disputeRepo,
		refundRepo:  refundRepo,
		userRepo:    userRepo,
	}
}

// HandleListDisputes lists all disputes with filtering
// GET /v1/admin/billing/disputes
func (h *DisputesHandler) HandleListDisputes(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	limit, offset := 50, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	// Build filter
	filter := &storage.DisputeFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if reason := r.URL.Query().Get("reason"); reason != "" {
		filter.Reason = reason
	}
	if isOpen := r.URL.Query().Get("is_open"); isOpen != "" {
		val := isOpen == "true"
		filter.IsOpen = &val
	}
	if requiresAction := r.URL.Query().Get("requires_action"); requiresAction != "" {
		val := requiresAction == "true"
		filter.RequiresAction = &val
	}
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tid
		}
	}
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &uid
		}
	}
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &t
		}
	}

	disputes, total, err := h.disputeRepo.ListDisputesWithRelations(r.Context(), filter, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list disputes")
		http.Error(w, "Failed to list disputes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"disputes": disputes,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleGetDispute retrieves a specific dispute by ID
// GET /v1/admin/billing/disputes/{disputeId}
func (h *DisputesHandler) HandleGetDispute(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	disputeIDStr := vars["disputeId"]

	disputeID, err := uuid.Parse(disputeIDStr)
	if err != nil {
		http.Error(w, "Invalid dispute ID", http.StatusBadRequest)
		return
	}

	dispute, err := h.disputeRepo.GetDisputeByID(r.Context(), disputeID)
	if err != nil {
		logrus.WithError(err).WithField("dispute_id", disputeID).Error("Failed to get dispute")
		http.Error(w, "Failed to get dispute", http.StatusInternalServerError)
		return
	}

	if dispute == nil {
		http.Error(w, "Dispute not found", http.StatusNotFound)
		return
	}

	// Populate relations
	if dispute.TenantID != nil {
		var tenantName string
		h.disputeRepo.DB().Raw("SELECT name FROM tenants WHERE id = ?", *dispute.TenantID).Scan(&tenantName)
		dispute.TenantName = tenantName
	}
	if dispute.UserID != nil {
		var userInfo struct {
			Email    string
			Username string
		}
		h.disputeRepo.DB().Raw("SELECT email, username FROM users WHERE id = ?", *dispute.UserID).Scan(&userInfo)
		dispute.UserEmail = userInfo.Email
		dispute.UserName = userInfo.Username
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dispute)
}

// HandleUpdateDisputeStatus manually updates a dispute status (for admin reconciliation)
// PATCH /v1/admin/billing/disputes/{disputeId}/status
func (h *DisputesHandler) HandleUpdateDisputeStatus(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	disputeIDStr := vars["disputeId"]

	disputeID, err := uuid.Parse(disputeIDStr)
	if err != nil {
		http.Error(w, "Invalid dispute ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status        string `json:"status"`
		Outcome       string `json:"outcome,omitempty"`
		OutcomeReason string `json:"outcome_reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{"needs_response", "warning_needs_response", "needs_review", "under_review", "won", "lost", "closed", "warning_closed"}
	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}
	if !isValid {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.disputeRepo.UpdateDisputeStatus(r.Context(), disputeID, req.Status, req.Outcome, req.OutcomeReason); err != nil {
		logrus.WithError(err).WithField("dispute_id", disputeID).Error("Failed to update dispute status")
		http.Error(w, "Failed to update dispute status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleGetDisputeStats returns aggregate statistics about disputes
// GET /v1/admin/billing/disputes/stats
func (h *DisputesHandler) HandleGetDisputeStats(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	stats, err := h.disputeRepo.GetDisputeStats(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get dispute stats")
		http.Error(w, "Failed to get dispute stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleListRefunds lists all refunds with filtering
// GET /v1/admin/billing/refunds
func (h *DisputesHandler) HandleListRefunds(w http.ResponseWriter, r *http.Request) {
	if h.refundRepo == nil {
		http.Error(w, "Refund service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	limit, offset := 50, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	// Build filter
	filter := &storage.RefundFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if reason := r.URL.Query().Get("reason"); reason != "" {
		filter.Reason = reason
	}
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tid
		}
	}
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &uid
		}
	}
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &t
		}
	}

	refunds, total, err := h.refundRepo.ListRefundsWithRelations(r.Context(), filter, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list refunds")
		http.Error(w, "Failed to list refunds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"refunds": refunds,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetRefund retrieves a specific refund by ID
// GET /v1/admin/billing/refunds/{refundId}
func (h *DisputesHandler) HandleGetRefund(w http.ResponseWriter, r *http.Request) {
	if h.refundRepo == nil {
		http.Error(w, "Refund service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	refundIDStr := vars["refundId"]

	refundID, err := uuid.Parse(refundIDStr)
	if err != nil {
		http.Error(w, "Invalid refund ID", http.StatusBadRequest)
		return
	}

	refund, err := h.refundRepo.GetRefundByID(r.Context(), refundID)
	if err != nil {
		logrus.WithError(err).WithField("refund_id", refundID).Error("Failed to get refund")
		http.Error(w, "Failed to get refund", http.StatusInternalServerError)
		return
	}

	if refund == nil {
		http.Error(w, "Refund not found", http.StatusNotFound)
		return
	}

	// Populate relations
	if refund.TenantID != nil {
		var tenantName string
		h.refundRepo.DB().Raw("SELECT name FROM tenants WHERE id = ?", *refund.TenantID).Scan(&tenantName)
		refund.TenantName = tenantName
	}
	if refund.UserID != nil {
		var userInfo struct {
			Email    string
			Username string
		}
		h.refundRepo.DB().Raw("SELECT email, username FROM users WHERE id = ?", *refund.UserID).Scan(&userInfo)
		refund.UserEmail = userInfo.Email
		refund.UserName = userInfo.Username
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refund)
}

// HandleGetRefundStats returns aggregate statistics about refunds
// GET /v1/admin/billing/refunds/stats
func (h *DisputesHandler) HandleGetRefundStats(w http.ResponseWriter, r *http.Request) {
	if h.refundRepo == nil {
		http.Error(w, "Refund service not available", http.StatusServiceUnavailable)
		return
	}

	stats, err := h.refundRepo.GetRefundStats(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get refund stats")
		http.Error(w, "Failed to get refund stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetChargebackReconciliation returns financial reconciliation for chargebacks
// GET /v1/admin/billing/chargebacks/reconciliation
func (h *DisputesHandler) HandleGetChargebackReconciliation(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse date range
	var startDate, endDate *time.Time
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
		}
	}

	recon, err := h.disputeRepo.GetChargebackReconciliation(r.Context(), startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to get chargeback reconciliation")
		http.Error(w, "Failed to get reconciliation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recon)
}

// HandleUpdateDisputeEvidence submits evidence for a dispute
// POST /v1/admin/billing/disputes/{disputeId}/evidence
func (h *DisputesHandler) HandleUpdateDisputeEvidence(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	disputeIDStr := vars["disputeId"]

	disputeID, err := uuid.Parse(disputeIDStr)
	if err != nil {
		http.Error(w, "Invalid dispute ID", http.StatusBadRequest)
		return
	}

	var req storage.EvidenceDetails
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.disputeRepo.UpdateDisputeEvidence(r.Context(), disputeID, &req); err != nil {
		logrus.WithError(err).WithField("dispute_id", disputeID).Error("Failed to update dispute evidence")
		http.Error(w, "Failed to update evidence", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "evidence_submitted"})
}

// HandleGetOpenDisputes returns only open disputes that require action
// GET /v1/admin/billing/disputes/open
func (h *DisputesHandler) HandleGetOpenDisputes(w http.ResponseWriter, r *http.Request) {
	if h.disputeRepo == nil {
		http.Error(w, "Dispute service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse pagination
	limit, offset := 50, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	// Build filter for open disputes
	filter := &storage.DisputeFilter{}
	isOpen := true
	filter.IsOpen = &isOpen

	disputes, total, err := h.disputeRepo.ListDisputesWithRelations(r.Context(), filter, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list open disputes")
		http.Error(w, "Failed to list disputes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"disputes":        disputes,
		"total":           total,
		"count":           len(disputes),
		"requires_action": h.countRequiresAction(disputes),
		"limit":           limit,
		"offset":          offset,
	})
}

// countRequiresAction counts disputes that require evidence submission
func (h *DisputesHandler) countRequiresAction(disputes []*storage.PaymentDispute) int {
	count := 0
	for _, d := range disputes {
		if strings.Contains(d.Status, "needs_response") || strings.Contains(d.Status, "warning") {
			count++
		}
	}
	return count
}

// DisputeRepository extends storage.DisputeRepository to expose DB access
type DisputeRepository interface {
	DB() interface{}
}
