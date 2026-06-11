package vault

import (
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleGetAuditLog handles GET /v1/vault/audit
// Gets audit logs for the tenant with pagination
func (h *Handler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get total count for tenant
	total, err := h.repo.CountAuditLogsByTenant(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Get audit logs for tenant with pagination
	logs, err := h.repo.GetAuditLogsByTenant(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Convert to response
	responses := make([]AuditLogEntryResponse, len(logs))
	for i, log := range logs {
		responses[i] = auditLogToResponse(&log)
	}

	h.respondJSON(w, http.StatusOK, ListAuditLogResponse{
		Entries: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// HandleGetSecretAuditLog handles GET /v1/vault/secrets/{id}/audit
// Gets audit logs for a specific secret
func (h *Handler) HandleGetSecretAuditLog(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	// Verify secret exists and belongs to tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to verify secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get total count for secret
	total, err := h.repo.CountAuditLogsBySecret(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Get audit logs for secret
	logs, err := h.repo.GetAuditLogsBySecret(r.Context(), *secretID, claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Convert to response
	responses := make([]AuditLogEntryResponse, len(logs))
	for i, log := range logs {
		responses[i] = auditLogToResponse(&log)
	}

	h.respondJSON(w, http.StatusOK, ListAuditLogResponse{
		Entries: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// HandleGetAuditLogByAction handles GET /v1/vault/audit/action/{action}
// Gets audit logs filtered by action type
func (h *Handler) HandleGetAuditLogByAction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	actionStr := vars["action"]
	action := vault.AuditAction(actionStr)
	if !action.Valid() {
		h.respondError(w, http.StatusBadRequest, "INVALID_ACTION", "Invalid audit action")
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get total count for action
	total, err := h.repo.CountAuditLogsByAction(r.Context(), action, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Get audit logs by action
	logs, err := h.repo.GetAuditLogsByAction(r.Context(), action, claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Convert to response
	responses := make([]AuditLogEntryResponse, len(logs))
	for i, log := range logs {
		responses[i] = auditLogToResponse(&log)
	}

	h.respondJSON(w, http.StatusOK, ListAuditLogResponse{
		Entries: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// HandleGetAuditLogByActor handles GET /v1/vault/audit/actor/{actor_type}/{actor_id}
// Gets audit logs for a specific actor
func (h *Handler) HandleGetAuditLogByActor(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	actorTypeStr := vars["actor_type"]
	actorID := vars["actor_id"]

	actorType := vault.ActorType(actorTypeStr)
	if !actorType.Valid() {
		h.respondError(w, http.StatusBadRequest, "INVALID_ACTOR_TYPE", "Invalid actor type")
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get total count for actor
	total, err := h.repo.CountAuditLogsByActor(r.Context(), actorID, actorType, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Get audit logs by actor
	logs, err := h.repo.GetAuditLogsByActor(r.Context(), actorID, actorType, claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to get audit logs")
		return
	}

	// Convert to response
	responses := make([]AuditLogEntryResponse, len(logs))
	for i, log := range logs {
		responses[i] = auditLogToResponse(&log)
	}

	h.respondJSON(w, http.StatusOK, ListAuditLogResponse{
		Entries: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// AuditLogQueryParams represents query parameters for audit log filtering
type AuditLogQueryParams struct {
	SecretID  *uuid.UUID
	Action    *vault.AuditAction
	ActorID   string
	ActorType *vault.ActorType
	StartTime *string
	EndTime   *string
	Success   *bool
	Limit     int
	Offset    int
}

// parseAuditLogQueryParams parses query parameters from the request
func parseAuditLogQueryParams(r *http.Request) AuditLogQueryParams {
	params := AuditLogQueryParams{
		Limit:  20,
		Offset: 0,
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 && limit <= 100 {
		params.Limit = limit
	}

	if offset, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && offset >= 0 {
		params.Offset = offset
	}

	if secretID := r.URL.Query().Get("secret_id"); secretID != "" {
		if id, err := uuid.Parse(secretID); err == nil {
			params.SecretID = &id
		}
	}

	if actionStr := r.URL.Query().Get("action"); actionStr != "" {
		action := vault.AuditAction(actionStr)
		if action.Valid() {
			params.Action = &action
		}
	}

	if actorTypeStr := r.URL.Query().Get("actor_type"); actorTypeStr != "" {
		actorType := vault.ActorType(actorTypeStr)
		if actorType.Valid() {
			params.ActorType = &actorType
		}
	}

	params.ActorID = r.URL.Query().Get("actor_id")
	params.StartTime = strPtr(r.URL.Query().Get("start_time"))
	params.EndTime = strPtr(r.URL.Query().Get("end_time"))

	if successStr := r.URL.Query().Get("success"); successStr != "" {
		success := successStr == "true"
		params.Success = &success
	}

	return params
}

// strPtr returns a pointer to the string if non-empty, nil otherwise
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
