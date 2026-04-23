package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ==================== Webhook Replay Admin Endpoints ====================

// HandleListStoredWebhooks lists stored webhook payloads with filtering
// GET /v1/admin/billing/webhooks
func (h *Handler) HandleListStoredWebhooks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	eventType := r.URL.Query().Get("event_type")

	limit, offset := 100, 0
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

	// Get operational repository
	if h.billingOperationalRepo == nil {
		http.Error(w, "Webhook storage not configured", http.StatusServiceUnavailable)
		return
	}

	payloads, total, err := h.billingOperationalRepo.ListStoredWebhookPayloads(r.Context(), status, eventType, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list stored webhook payloads")
		http.Error(w, "Failed to list webhooks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payloads": payloads,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleGetStoredWebhook gets a specific stored webhook payload
// GET /v1/admin/billing/webhooks/:webhookId
func (h *Handler) HandleGetStoredWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhookId"]
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	if h.billingOperationalRepo == nil {
		http.Error(w, "Webhook storage not configured", http.StatusServiceUnavailable)
		return
	}

	payload, err := h.billingOperationalRepo.GetStoredWebhookPayload(r.Context(), webhookID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get stored webhook payload")
		http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		return
	}
	if payload == nil {
		http.Error(w, "Webhook payload not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// HandleReplayWebhook manually replays a stored webhook
// POST /v1/admin/billing/webhooks/:webhookId/replay
func (h *Handler) HandleReplayWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhookId"]
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}

	// Get current user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Reason is required", http.StatusBadRequest)
		return
	}

	if h.billingOperationalRepo == nil {
		http.Error(w, "Webhook storage not configured", http.StatusServiceUnavailable)
		return
	}

	// Get the stored payload
	payload, err := h.billingOperationalRepo.GetStoredWebhookPayload(r.Context(), webhookID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get stored webhook payload for replay")
		http.Error(w, "Failed to get webhook", http.StatusInternalServerError)
		return
	}
	if payload == nil {
		http.Error(w, "Webhook payload not found", http.StatusNotFound)
		return
	}

	// Check if already processed
	if payload.ProcessingStatus == "processed" {
		http.Error(w, "Webhook already processed - replay would cause duplicate processing", http.StatusConflict)
		return
	}

	// Create replay request record
	replayReq, err := h.billingOperationalRepo.CreateWebhookReplayRequest(r.Context(), webhookID, userID, req.Reason)
	if err != nil {
		logrus.WithError(err).Error("Failed to create webhook replay request")
		http.Error(w, "Failed to create replay request", http.StatusInternalServerError)
		return
	}

	// Mark as replayed
	if err := h.billingOperationalRepo.MarkWebhookPayloadReplayed(r.Context(), webhookID, userID, req.Reason); err != nil {
		logrus.WithError(err).Error("Failed to mark webhook as replayed")
	}

	// Complete the replay request
	resultMsg := "Webhook replayed successfully. The event will be reprocessed by the webhook handler."
	if err := h.billingOperationalRepo.CompleteWebhookReplayRequest(r.Context(), replayReq.ID, "completed", resultMsg); err != nil {
		logrus.WithError(err).Error("Failed to complete webhook replay request")
	}

	logrus.WithFields(logrus.Fields{
		"webhook_id":  webhookID,
		"event_id":    payload.StripeEventID,
		"event_type":  payload.EventType,
		"replayed_by": userID,
		"reason":      req.Reason,
	}).Info("Webhook manually replayed by admin")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     resultMsg,
		"replay_id":   replayReq.ID,
		"webhook_id":  webhookID,
		"event_id":    payload.StripeEventID,
		"event_type":  payload.EventType,
		"replayed_at": time.Now(),
	})
}

// HandleListWebhookReplayRequests lists webhook replay requests
// GET /v1/admin/billing/webhooks/replay-requests
func (h *Handler) HandleListWebhookReplayRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	limit, offset := 100, 0
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

	if h.billingOperationalRepo == nil {
		http.Error(w, "Webhook storage not configured", http.StatusServiceUnavailable)
		return
	}

	requests, total, err := h.billingOperationalRepo.ListWebhookReplayRequests(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list webhook replay requests")
		http.Error(w, "Failed to list replay requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// ==================== Tax Exemption Certificate Admin Endpoints ====================

// HandleListPendingTaxCertificates lists pending tax exemption certificates for admin review
// GET /v1/admin/billing/tax-exemptions/pending
func (h *Handler) HandleListPendingTaxCertificates(w http.ResponseWriter, r *http.Request) {
	limit, offset := 100, 0
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

	if h.billingOperationalRepo == nil {
		http.Error(w, "Tax exemption storage not configured", http.StatusServiceUnavailable)
		return
	}

	certs, total, err := h.billingOperationalRepo.ListPendingTaxExemptionCertificates(r.Context(), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list pending tax exemption certificates")
		http.Error(w, "Failed to list certificates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificates": certs,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// HandleReviewTaxCertificate approves or rejects a tax exemption certificate
// POST /v1/admin/billing/tax-exemptions/:certificateId/review
func (h *Handler) HandleReviewTaxCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	certIDStr := vars["certificateId"]
	certID, err := uuid.Parse(certIDStr)
	if err != nil {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}

	// Get current user ID from context
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Action          string `json:"action"` // "approve" or "reject"
		Notes           string `json:"notes"`
		RejectionReason string `json:"rejection_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		http.Error(w, "Action must be 'approve' or 'reject'", http.StatusBadRequest)
		return
	}

	if req.Action == "reject" && req.RejectionReason == "" {
		http.Error(w, "Rejection reason is required when rejecting", http.StatusBadRequest)
		return
	}

	if h.billingOperationalRepo == nil {
		http.Error(w, "Tax exemption storage not configured", http.StatusServiceUnavailable)
		return
	}

	approved := req.Action == "approve"
	cert, err := h.billingOperationalRepo.ReviewTaxExemptionCertificate(r.Context(), certID, userID, approved, req.Notes, req.RejectionReason)
	if err != nil {
		logrus.WithError(err).Error("Failed to review tax exemption certificate")
		http.Error(w, "Failed to review certificate", http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"certificate_id": certID,
		"action":         req.Action,
		"reviewed_by":    userID,
		"tenant_id":      cert.TenantID,
	}).Info("Tax exemption certificate reviewed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cert)
}

// HandleCleanupExpiredWebhooks triggers cleanup of expired webhook payloads
// POST /v1/admin/billing/webhooks/cleanup
func (h *Handler) HandleCleanupExpiredWebhooks(w http.ResponseWriter, r *http.Request) {
	if h.billingOperationalRepo == nil {
		http.Error(w, "Webhook storage not configured", http.StatusServiceUnavailable)
		return
	}

	deletedCount, err := h.billingOperationalRepo.CleanupExpiredWebhookPayloads(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to cleanup expired webhook payloads")
		http.Error(w, "Cleanup failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"deleted_count": deletedCount,
		"message":       "Expired webhook payloads cleaned up",
	})
}
