package trustapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WebhookHandler handles webhook management API endpoints
type WebhookHandler struct {
	repo    *trustapi.WebhookRepository
	service *trustapi.WebhookService
	logger  *logrus.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(repo *trustapi.WebhookRepository) *WebhookHandler {
	service := trustapi.NewWebhookService(repo)
	return &WebhookHandler{
		repo:    repo,
		service: service,
		logger:  logrus.New(),
	}
}

// SetLogger sets the logger for the handler
func (h *WebhookHandler) SetLogger(logger *logrus.Logger) {
	h.logger = logger
	h.service.SetLogger(logger)
}

// ============================================
// Webhook CRUD Handlers
// ============================================

// HandleCreateWebhook handles POST /v1/webhooks
// Creates a new webhook configuration
func (h *WebhookHandler) HandleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req trustapi.WebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate URL
	if err := h.service.ValidateWebhookURL(req.URL); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error(), "invalid_url")
		return
	}

	// Get user from context
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Marshal arrays to JSON
	eventsJSON, _ := json.Marshal(req.Events)
	filterJSON, _ := json.Marshal(req.FunctionFilter)
	headersJSON, _ := json.Marshal(req.CustomHeaders)

	// Create webhook
	webhook := &trustapi.TrustWebhook{
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Events:         eventsJSON,
		EventFilter:    "specific",
		FunctionFilter: filterJSON,
		Status:         string(trustapi.WebhookStatusActive),
		MaxRetries:     req.MaxRetries,
		OwnerID:        claims.UserID,
		OwnerType:      "user",
		CustomHeaders:  headersJSON,
	}

	if req.Secret != "" {
		webhook.Secret = req.Secret
	}

	if err := h.repo.CreateWebhook(webhook); err != nil {
		h.logger.WithError(err).Error("Failed to create webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to create webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"webhook_id": webhook.WebhookID,
		"user_id":    claims.UserID,
		"url":        req.URL,
	}).Info("Webhook created")

	// Convert events for response
	var events []string
	json.Unmarshal(webhook.Events, &events)

	var headers map[string]string
	json.Unmarshal(webhook.CustomHeaders, &headers)

	response := trustapi.WebhookResponse{
		ID:            webhook.ID,
		WebhookID:     webhook.WebhookID,
		Name:          webhook.Name,
		Description:   webhook.Description,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        events,
		EventFilter:   webhook.EventFilter,
		Status:        webhook.Status,
		MaxRetries:    webhook.MaxRetries,
		CustomHeaders: headers,
		CreatedAt:     webhook.CreatedAt,
		UpdatedAt:     webhook.UpdatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleListWebhooks handles GET /v1/webhooks
// Lists all webhooks for the authenticated user
func (h *WebhookHandler) HandleListWebhooks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	status := r.URL.Query().Get("status")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	webhooks, total, err := h.repo.ListWebhooksForOwner(claims.UserID, "user", status, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list webhooks")
		h.writeError(w, http.StatusInternalServerError, "Failed to list webhooks", "internal_error")
		return
	}

	responseList := make([]trustapi.WebhookResponse, len(webhooks))
	for i, webhook := range webhooks {
		var events []string
		json.Unmarshal(webhook.Events, &events)

		var headers map[string]string
		json.Unmarshal(webhook.CustomHeaders, &headers)

		responseList[i] = trustapi.WebhookResponse{
			ID:            webhook.ID,
			WebhookID:     webhook.WebhookID,
			Name:          webhook.Name,
			Description:   webhook.Description,
			URL:           webhook.URL,
			Method:        webhook.Method,
			Events:        events,
			EventFilter:   webhook.EventFilter,
			Status:        webhook.Status,
			FailCount:     webhook.FailCount,
			LastFailure:   webhook.LastFailure,
			LastSuccess:   webhook.LastSuccess,
			MaxRetries:    webhook.MaxRetries,
			CustomHeaders: headers,
			CreatedAt:     webhook.CreatedAt,
			UpdatedAt:     webhook.UpdatedAt,
		}
	}

	response := trustapi.WebhookListResponse{
		Webhooks:   responseList,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetWebhook handles GET /v1/webhooks/{webhook_id}
// Gets a specific webhook
func (h *WebhookHandler) HandleGetWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	// Verify ownership
	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to access this webhook", "forbidden")
		return
	}

	var events []string
	json.Unmarshal(webhook.Events, &events)

	var headers map[string]string
	json.Unmarshal(webhook.CustomHeaders, &headers)

	response := trustapi.WebhookResponse{
		ID:            webhook.ID,
		WebhookID:     webhook.WebhookID,
		Name:          webhook.Name,
		Description:   webhook.Description,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        events,
		EventFilter:   webhook.EventFilter,
		Status:        webhook.Status,
		FailCount:     webhook.FailCount,
		LastFailure:   webhook.LastFailure,
		LastSuccess:   webhook.LastSuccess,
		MaxRetries:    webhook.MaxRetries,
		CustomHeaders: headers,
		CreatedAt:     webhook.CreatedAt,
		UpdatedAt:     webhook.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleUpdateWebhook handles PUT /v1/webhooks/{webhook_id}
// Updates a webhook
func (h *WebhookHandler) HandleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to update this webhook", "forbidden")
		return
	}

	var req trustapi.WebhookUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Update fields
	if req.Name != "" {
		webhook.Name = req.Name
	}
	if req.Description != "" {
		webhook.Description = req.Description
	}
	if req.URL != "" {
		if err := h.service.ValidateWebhookURL(req.URL); err != nil {
			h.writeError(w, http.StatusBadRequest, err.Error(), "invalid_url")
			return
		}
		webhook.URL = req.URL
	}
	if len(req.Events) > 0 {
		eventsJSON, _ := json.Marshal(req.Events)
		webhook.Events = eventsJSON
	}
	if req.Secret != "" {
		webhook.Secret = req.Secret
	}
	if req.Status != "" {
		webhook.Status = req.Status
	}
	if req.MaxRetries > 0 {
		webhook.MaxRetries = req.MaxRetries
	}
	if req.FunctionFilter != nil {
		filterJSON, _ := json.Marshal(req.FunctionFilter)
		webhook.FunctionFilter = filterJSON
	}
	if req.CustomHeaders != nil {
		headersJSON, _ := json.Marshal(req.CustomHeaders)
		webhook.CustomHeaders = headersJSON
	}

	if err := h.repo.UpdateWebhook(webhook); err != nil {
		h.logger.WithError(err).Error("Failed to update webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to update webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"webhook_id": webhook.WebhookID,
		"user_id":    claims.UserID,
	}).Info("Webhook updated")

	var events []string
	json.Unmarshal(webhook.Events, &events)

	var headers map[string]string
	json.Unmarshal(webhook.CustomHeaders, &headers)

	response := trustapi.WebhookResponse{
		ID:            webhook.ID,
		WebhookID:     webhook.WebhookID,
		Name:          webhook.Name,
		Description:   webhook.Description,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        events,
		EventFilter:   webhook.EventFilter,
		Status:        webhook.Status,
		FailCount:     webhook.FailCount,
		LastFailure:   webhook.LastFailure,
		LastSuccess:   webhook.LastSuccess,
		MaxRetries:    webhook.MaxRetries,
		CustomHeaders: headers,
		CreatedAt:     webhook.CreatedAt,
		UpdatedAt:     webhook.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleDeleteWebhook handles DELETE /v1/webhooks/{webhook_id}
// Deletes a webhook
func (h *WebhookHandler) HandleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to delete this webhook", "forbidden")
		return
	}

	if err := h.repo.DeleteWebhook(webhook.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"webhook_id": webhook.WebhookID,
		"user_id":    claims.UserID,
	}).Info("Webhook deleted")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Webhook deleted successfully",
		"webhook_id": webhookIDStr,
	})
}

// ============================================
// Webhook Delivery Handlers
// ============================================

// HandleListDeliveries handles GET /v1/webhooks/{webhook_id}/deliveries
// Lists delivery attempts for a webhook
func (h *WebhookHandler) HandleListDeliveries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized", "forbidden")
		return
	}

	status := r.URL.Query().Get("status")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	deliveries, total, err := h.repo.ListDeliveriesForWebhook(webhook.ID, status, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list deliveries")
		h.writeError(w, http.StatusInternalServerError, "Failed to list deliveries", "internal_error")
		return
	}

	responseList := make([]trustapi.WebhookDeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		responseList[i] = trustapi.WebhookDeliveryResponse{
			ID:                 delivery.ID,
			DeliveryID:         delivery.DeliveryID,
			WebhookID:          delivery.WebhookID,
			EventType:          delivery.EventType,
			EntityID:           delivery.EntityID,
			AttemptNumber:      delivery.AttemptNumber,
			MaxAttempts:        delivery.MaxAttempts,
			Status:             delivery.Status,
			ResponseStatusCode: delivery.ResponseStatusCode,
			ResponseTimeMs:     delivery.ResponseTimeMs,
			ErrorMessage:       delivery.ErrorMessage,
			SentAt:             delivery.SentAt,
			DeliveredAt:        delivery.DeliveredAt,
			CreatedAt:          delivery.CreatedAt,
		}
	}

	response := trustapi.WebhookDeliveryListResponse{
		Deliveries: responseList,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleTestWebhook handles POST /v1/webhooks/{webhook_id}/test
// Tests a webhook with a test payload
func (h *WebhookHandler) HandleTestWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized", "forbidden")
		return
	}

	var req trustapi.WebhookTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.EventType = "trust.score.updated"
		req.TestData = map[string]interface{}{
			"message": "This is a test webhook payload",
			"test_id": uuid.New().String(),
		}
	}

	result, err := h.service.TestWebhook(webhook, req.EventType, req.TestData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to test webhook", "test_failed")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// HandleGetDeliveryStats handles GET /v1/webhooks/{webhook_id}/stats
// Gets delivery statistics for a webhook
func (h *WebhookHandler) HandleGetDeliveryStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookIDStr := vars["webhook_id"]

	webhook, err := h.repo.GetWebhookByWebhookID(webhookIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Webhook not found", "webhook_not_found")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || webhook.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized", "forbidden")
		return
	}

	stats, err := h.repo.GetDeliveryStats(webhook.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get delivery stats")
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", "internal_error")
		return
	}

	stats["webhook_id"] = webhook.WebhookID

	h.writeJSON(w, http.StatusOK, stats)
}

// ============================================
// Helper Methods
// ============================================

func (h *WebhookHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *WebhookHandler) writeError(w http.ResponseWriter, status int, err string, code string) {
	h.writeJSON(w, status, trustapi.ErrorResponse{
		Error: err,
		Code:  code,
	})
}
