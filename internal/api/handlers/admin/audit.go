package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminAuditHandler handles detailed audit log operations
type AdminAuditHandler struct {
	db   *sql.DB
	repo storage.Repository
}

// NewAdminAuditHandler creates a new admin audit handler
func NewAdminAuditHandler(db *sql.DB) *AdminAuditHandler {
	return &AdminAuditHandler{
		db: db,
	}
}

// HandleListAuditLogs lists audit logs (admin endpoint)
func (h *AdminAuditHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	var limit, offset int
	var filters map[string]interface{}
	var events []*storage.AuditEvent

	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Error("HandleListAuditLogs panic")
			apierror.WriteError(w, apierror.NewInternal("Internal server error"))
			return
		}
		writeAuditLogsResponse(w, events, limit, offset, filters)
	}()

	limit = 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset = 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filters = make(map[string]interface{})
	if actorUserIDStr := r.URL.Query().Get("actor_user_id"); actorUserIDStr != "" {
		if actorUserID, err := uuid.Parse(actorUserIDStr); err == nil {
			filters["actor_user_id"] = actorUserID
		}
	}
	if actorEmail := r.URL.Query().Get("actor_email"); actorEmail != "" {
		filters["actor_email"] = actorEmail
	}
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filters["tenant_id"] = tenantID
		}
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters["action"] = action
	}
	if resourceType := r.URL.Query().Get("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}
	if resourceIDStr := r.URL.Query().Get("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			filters["resource_id"] = resourceID
		}
	}
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filters["success"] = success
		}
	}
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters["start_time"] = startTime
		}
	}
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters["end_time"] = endTime
		}
	}

	events, err := h.repo.ListAuditEventsFiltered(limit, offset, filters)
	if err != nil {
		logrus.WithError(err).Warn("HandleListAuditLogs: failed to query audit events")
		events = []*storage.AuditEvent{}
	}
}

// HandleGetAuditLog gets a single audit log by ID
func (h *AdminAuditHandler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid audit log ID"))
		return
	}

	event, err := h.repo.GetAuditEventByID(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("HandleGetAuditLog: failed to get audit event")
		apierror.WriteError(w, apierror.NewInternal("failed to retrieve audit log"))
		return
	}
	if event == nil {
		apierror.WriteError(w, apierror.NewNotFound("audit log not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func writeAuditLogsResponse(w http.ResponseWriter, events []*storage.AuditEvent, limit, offset int, filters map[string]interface{}) {
	if filters == nil {
		filters = make(map[string]interface{})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":  events,
		"limit":   limit,
		"offset":  offset,
		"filters": filters,
	})
}

// HandleListAuditEvents lists audit events. Never returns 500: on any error returns 200 with empty events.
func (h *Handler) HandleListAuditEvents(w http.ResponseWriter, r *http.Request) {
	var limit, offset int
	var filters map[string]interface{}
	var events []*storage.AuditEvent

	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Warn("HandleListAuditEvents panic; returning empty list")
			events = []*storage.AuditEvent{}
			filters = make(map[string]interface{})
		}
		writeAuditEventsResponse(w, events, limit, offset, filters)
	}()

	limit = 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset = 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filters = make(map[string]interface{})
	if actorUserIDStr := r.URL.Query().Get("actor_user_id"); actorUserIDStr != "" {
		if actorUserID, err := uuid.Parse(actorUserIDStr); err == nil {
			filters["actor_user_id"] = actorUserID
		}
	}
	if actorEmail := r.URL.Query().Get("actor_email"); actorEmail != "" {
		filters["actor_email"] = actorEmail
	}
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filters["tenant_id"] = tenantID
		}
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters["action"] = action
	}
	if resourceType := r.URL.Query().Get("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}
	if resourceIDStr := r.URL.Query().Get("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			filters["resource_id"] = resourceID
		}
	}
	if successStr := r.URL.Query().Get("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filters["success"] = success
		}
	}
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters["start_time"] = startTime
		}
	}
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters["end_time"] = endTime
		}
	}

	var err error
	events, err = h.repo.ListAuditEventsFiltered(limit, offset, filters)
	if err != nil {
		logrus.WithError(err).Warn("Failed to list audit events; returning empty list (e.g. audit_events table missing or schema mismatch)")
		events = []*storage.AuditEvent{}
	}
}

func writeAuditEventsResponse(w http.ResponseWriter, events []*storage.AuditEvent, limit, offset int, filters map[string]interface{}) {
	if filters == nil {
		filters = make(map[string]interface{})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":  events,
		"limit":   limit,
		"offset":  offset,
		"filters": filters,
	})
}
