package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewServiceUnavailable("Webhook storage not configured"))
		return
	}

	payloads, total, err := h.billingOperationalRepo.ListStoredWebhookPayloads(r.Context(), status, eventType, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list stored webhook payloads")
		apierror.WriteError(w, apierror.NewInternal("Failed to list webhooks"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid webhook ID"))
		return
	}

	if h.billingOperationalRepo == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Webhook storage not configured"))
		return
	}

	payload, err := h.billingOperationalRepo.GetStoredWebhookPayload(r.Context(), webhookID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get stored webhook payload")
		apierror.WriteError(w, apierror.NewInternal("Failed to get webhook"))
		return
	}
	if payload == nil {
		apierror.WriteError(w, apierror.NewNotFound("Webhook payload not found"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid webhook ID"))
		return
	}

	// Get current user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if req.Reason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Reason is required"))
		return
	}

	if h.billingOperationalRepo == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Webhook storage not configured"))
		return
	}

	// Get the stored payload
	payload, err := h.billingOperationalRepo.GetStoredWebhookPayload(r.Context(), webhookID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get stored webhook payload for replay")
		apierror.WriteError(w, apierror.NewInternal("Failed to get webhook"))
		return
	}
	if payload == nil {
		apierror.WriteError(w, apierror.NewNotFound("Webhook payload not found"))
		return
	}

	// Check if already processed
	if payload.ProcessingStatus == "processed" {
		apierror.WriteError(w, apierror.NewConflict("Webhook already processed - replay would cause duplicate processing"))
		return
	}

	// Create replay request record
	replayReq, err := h.billingOperationalRepo.CreateWebhookReplayRequest(r.Context(), webhookID, userID, req.Reason)
	if err != nil {
		logrus.WithError(err).Error("Failed to create webhook replay request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create replay request"))
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
		apierror.WriteError(w, apierror.NewServiceUnavailable("Webhook storage not configured"))
		return
	}

	requests, total, err := h.billingOperationalRepo.ListWebhookReplayRequests(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list webhook replay requests")
		apierror.WriteError(w, apierror.NewInternal("Failed to list replay requests"))
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
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tax exemption storage not configured"))
		return
	}

	certs, total, err := h.billingOperationalRepo.ListPendingTaxExemptionCertificates(r.Context(), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list pending tax exemption certificates")
		apierror.WriteError(w, apierror.NewInternal("Failed to list certificates"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid certificate ID"))
		return
	}

	// Get current user ID from context
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req struct {
		Action          string `json:"action"` // "approve" or "reject"
		Notes           string `json:"notes"`
		RejectionReason string `json:"rejection_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		apierror.WriteError(w, apierror.NewBadRequest("Action must be 'approve' or 'reject'"))
		return
	}

	if req.Action == "reject" && req.RejectionReason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Rejection reason is required when rejecting"))
		return
	}

	if h.billingOperationalRepo == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tax exemption storage not configured"))
		return
	}

	approved := req.Action == "approve"
	cert, err := h.billingOperationalRepo.ReviewTaxExemptionCertificate(r.Context(), certID, userID, approved, req.Notes, req.RejectionReason)
	if err != nil {
		logrus.WithError(err).Error("Failed to review tax exemption certificate")
		apierror.WriteError(w, apierror.NewInternal("Failed to review certificate"))
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
		apierror.WriteError(w, apierror.NewServiceUnavailable("Webhook storage not configured"))
		return
	}

	deletedCount, err := h.billingOperationalRepo.CleanupExpiredWebhookPayloads(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to cleanup expired webhook payloads")
		apierror.WriteError(w, apierror.NewInternal("Cleanup failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"deleted_count": deletedCount,
		"message":       "Expired webhook payloads cleaned up",
	})
}
