package function_webhooks

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo    *storage.FunctionWebhookRepository
	service *storage.FunctionWebhookService
	logger  *logrus.Logger
}

func NewHandler(repo *storage.FunctionWebhookRepository, service *storage.FunctionWebhookService) *Handler {
	return &Handler{
		repo:    repo,
		service: service,
		logger:  logrus.New(),
	}
}

func (h *Handler) SetLogger(logger *logrus.Logger) {
	h.logger = logger
	if h.service != nil {
		h.service.SetLogger(logger)
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errMsg string, code string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": errMsg,
		"code":  code,
	})
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	var req storage.FunctionWebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	if err := h.service.ValidateURL(req.URL); err != nil {
		logrus.WithError(err).WithField("url", req.URL).Info("function_webhooks: invalid URL")
		h.writeError(w, http.StatusBadRequest, "Invalid webhook URL", "invalid_url")
		return
	}

	secret := req.Secret
	if secret == "" {
		secret = h.repo.GenerateSecret()
	}

	var functionID *uuid.UUID
	if req.FunctionID != nil {
		fid, err := uuid.Parse(*req.FunctionID)
		if err == nil {
			functionID = &fid
		}
	}

	sub := &storage.FunctionWebhookSubscription{
		TenantID:   claims.TenantID,
		FunctionID: functionID,
		URL:        req.URL,
		Secret:     secret,
		EventTypes: pq.StringArray(req.EventTypes),
		Active:     true,
		CreatedBy:  &claims.UserID,
	}

	if err := h.repo.Create(r.Context(), sub); err != nil {
		h.logger.WithError(err).Error("Failed to create function webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to create function webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": sub.ID,
		"tenant_id":       claims.TenantID,
		"url":             req.URL,
	}).Info("Function webhook created")

	h.writeJSON(w, http.StatusCreated, storage.FunctionWebhookResponse{
		ID:         sub.ID,
		TenantID:   sub.TenantID,
		FunctionID: sub.FunctionID,
		URL:        sub.URL,
		EventTypes: sub.EventTypes,
		Active:     sub.Active,
		CreatedAt:  sub.CreatedAt,
		CreatedBy:  sub.CreatedBy,
	})
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var functionID *uuid.UUID
	if fidStr := r.URL.Query().Get("function_id"); fidStr != "" {
		if fid, err := uuid.Parse(fidStr); err == nil {
			functionID = &fid
		}
	}

	subs, total, err := h.repo.List(r.Context(), claims.TenantID, functionID, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list function webhooks")
		h.writeError(w, http.StatusInternalServerError, "Failed to list function webhooks", "internal_error")
		return
	}

	responseList := make([]storage.FunctionWebhookResponse, len(subs))
	for i, sub := range subs {
		responseList[i] = storage.FunctionWebhookResponse{
			ID:         sub.ID,
			TenantID:   sub.TenantID,
			FunctionID: sub.FunctionID,
			URL:        sub.URL,
			EventTypes: sub.EventTypes,
			Active:     sub.Active,
			CreatedAt:  sub.CreatedAt,
			CreatedBy:  sub.CreatedBy,
		}
	}

	h.writeJSON(w, http.StatusOK, storage.FunctionWebhookListResponse{
		Subscriptions: responseList,
		TotalCount:    total,
		Page:          page,
		PageSize:      pageSize,
	})
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	subID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid subscription ID", "invalid_id")
		return
	}

	sub, err := h.repo.GetByID(r.Context(), subID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function webhook not found", "not_found")
		return
	}

	h.writeJSON(w, http.StatusOK, storage.FunctionWebhookResponse{
		ID:         sub.ID,
		TenantID:   sub.TenantID,
		FunctionID: sub.FunctionID,
		URL:        sub.URL,
		EventTypes: sub.EventTypes,
		Active:     sub.Active,
		CreatedAt:  sub.CreatedAt,
		CreatedBy:  sub.CreatedBy,
	})
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	subID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid subscription ID", "invalid_id")
		return
	}

	sub, err := h.repo.GetByID(r.Context(), subID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function webhook not found", "not_found")
		return
	}

	var req storage.FunctionWebhookUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	if req.URL != "" {
		if err := h.service.ValidateURL(req.URL); err != nil {
			logrus.WithError(err).WithField("url", req.URL).Info("function_webhooks: invalid URL on update")
			h.writeError(w, http.StatusBadRequest, "Invalid webhook URL", "invalid_url")
			return
		}
		sub.URL = req.URL
	}

	if req.EventTypes != nil {
		sub.EventTypes = pq.StringArray(req.EventTypes)
	}

	if req.Active != nil {
		sub.Active = *req.Active
	}

	if err := h.repo.Update(r.Context(), sub); err != nil {
		h.logger.WithError(err).Error("Failed to update function webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to update function webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": sub.ID,
		"tenant_id":       claims.TenantID,
	}).Info("Function webhook updated")

	h.writeJSON(w, http.StatusOK, storage.FunctionWebhookResponse{
		ID:         sub.ID,
		TenantID:   sub.TenantID,
		FunctionID: sub.FunctionID,
		URL:        sub.URL,
		EventTypes: sub.EventTypes,
		Active:     sub.Active,
		CreatedAt:  sub.CreatedAt,
		CreatedBy:  sub.CreatedBy,
	})
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	subID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid subscription ID", "invalid_id")
		return
	}

	if err := h.repo.Delete(r.Context(), subID, claims.TenantID); err != nil {
		h.logger.WithError(err).Error("Failed to delete function webhook")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete function webhook", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": subID,
		"tenant_id":       claims.TenantID,
	}).Info("Function webhook deleted")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Function webhook deleted successfully",
	})
}

func (h *Handler) HandleListDeliveries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	subID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid subscription ID", "invalid_id")
		return
	}

	_, err = h.repo.GetByID(r.Context(), subID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function webhook not found", "not_found")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	deliveries, total, err := h.repo.ListDeliveries(r.Context(), subID, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list deliveries")
		h.writeError(w, http.StatusInternalServerError, "Failed to list deliveries", "internal_error")
		return
	}

	responseList := make([]storage.FunctionWebhookDeliveryResponse, len(deliveries))
	for i, d := range deliveries {
		responseList[i] = storage.FunctionWebhookDeliveryResponse{
			ID:             d.ID,
			SubscriptionID: d.SubscriptionID,
			EventType:      d.EventType,
			Payload:        d.Payload,
			ResponseStatus: d.ResponseStatus,
			ResponseBody:   d.ResponseBody,
			AttemptedAt:    d.AttemptedAt,
			Success:        d.Success,
			ErrorMessage:   d.ErrorMessage,
		}
	}

	h.writeJSON(w, http.StatusOK, storage.FunctionWebhookDeliveryListResponse{
		Deliveries: responseList,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	subID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid subscription ID", "invalid_id")
		return
	}

	sub, err := h.repo.GetByID(r.Context(), subID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function webhook not found", "not_found")
		return
	}

	var req storage.FunctionWebhookTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.EventType = "function.deployed"
		req.TestData = map[string]interface{}{
			"message": "This is a test webhook payload",
			"test_id": uuid.New().String(),
		}
	}

	result, err := h.service.TestWebhook(sub, req.EventType, req.TestData)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to test webhook", "test_failed")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}
