package statefabric

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	repo "github.com/functionfly/functionfly/internal/storage/statefabric"
)

type Handler struct {
	repo         *repo.Repository
	sfAddons     *statefabricaddons.Repository
	cleanupSvc   *repo.CleanupService
	planResolver PlanResolver
}

type PlanResolver interface {
	GetTenantPlan(ctx context.Context, tenantID uuid.UUID) string
}

func NewHandler(r *repo.Repository, sfAddons *statefabricaddons.Repository) *Handler {
	return &Handler{repo: r, sfAddons: sfAddons}
}

func NewHandlerWithCleanup(r *repo.Repository, sfAddons *statefabricaddons.Repository, cleanupSvc *repo.CleanupService) *Handler {
	return &Handler{repo: r, sfAddons: sfAddons, cleanupSvc: cleanupSvc}
}

func NewHandlerWithPlanResolver(r *repo.Repository, sfAddons *statefabricaddons.Repository, pr PlanResolver) *Handler {
	return &Handler{repo: r, sfAddons: sfAddons, planResolver: pr}
}

type createFabricRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Settings    map[string]interface{} `json:"settings"`
}

type updateFabricRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
}

type createStoreRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	MaxSize int64  `json:"maxSize"`
	Region  string `json:"region"`
}

type createPipelineRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Steps       []map[string]interface{} `json:"steps"`
}

type updatePipelineRequest struct {
	Name        *string                  `json:"name"`
	Description *string                  `json:"description"`
	Steps       []map[string]interface{} `json:"steps"`
	Status      *string                  `json:"status"`
}

type createSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StoreID     string `json:"storeId"`
}

type createReplayRequest struct {
	SnapshotID   string `json:"snapshotId"`
	StartEventID string `json:"startEventId"`
	EndEventID   string `json:"endEventId"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("statefabric: failed to marshal JSON response")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal encoding error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func writeErr(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"request_uri": r.RequestURI,
		"method":      r.Method,
	}).Error("internal server error")
	apierror.WriteError(w, apierror.NewInternal("an internal error occurred"))
}

func clientError(w http.ResponseWriter, r *http.Request, err error) {
	apierror.LogAndBadRequest(w, r, err, "statefabric handler")
}

type auditInfo struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	ResourceID  uuid.UUID
	Action      string
	Success     bool
	Duration    time.Duration
	Description string
}

func (h *Handler) auditLog(r *http.Request, info auditInfo) {
	logrus.WithFields(logrus.Fields{
		"service":       "statefabric",
		"actor_user_id": info.UserID.String(),
		"tenant_id":     info.TenantID.String(),
		"resource_id":   info.ResourceID.String(),
		"action":        info.Action,
		"success":       info.Success,
		"duration_ms":   info.Duration.Milliseconds(),
		"ip_address":    getIPAddress(r),
		"user_agent":    r.UserAgent(),
		"request_id":    r.Header.Get("X-Request-ID"),
	}).Info("state_fabric_audit")
}

func getIPAddress(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func (h *Handler) fabricToAPI(f repo.Fabric, stores []repo.FabricStore, pipelines []repo.Pipeline) map[string]interface{} {
	data, err := json.Marshal(f)
	out := make(map[string]interface{})
	if err == nil {
		_ = json.Unmarshal(data, &out)
	}
	if out == nil {
		out = make(map[string]interface{})
	}
	out["stores"] = stores
	out["pipelines"] = pipelines
	return out
}

func getClaims(r *http.Request) (*middleware.AuthMiddleware, *uuid.UUID, *uuid.UUID, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return nil, nil, nil, false
	}
	userID := claims.UserID
	tenantID := claims.TenantID
	return nil, &userID, &tenantID, true
}

func tenantAndUser(r *http.Request, w http.ResponseWriter) (uuid.UUID, uuid.UUID, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return uuid.Nil, uuid.Nil, false
	}
	return claims.TenantID, claims.UserID, true
}

func parseID(w http.ResponseWriter, value string, label string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		http.Error(w, "invalid "+label, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) requireAddon(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID, addonID string) bool {
	if h.sfAddons == nil {
		writeErr(w, http.StatusForbidden, "This feature requires a paid add-on")
		return false
	}
	ok, err := h.sfAddons.HasActiveAddon(r.Context(), tenantID, addonID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to verify add-on entitlement")
		return false
	}
	if !ok {
		writeErr(w, http.StatusPaymentRequired, "add-on required")
		return false
	}
	return true
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, _, err := h.repo.ListFabrics(r.Context(), repo.ListOptions{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
		Status:   r.URL.Query().Get("status"),
		Search:   r.URL.Query().Get("search"),
	})
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

var allowedFabricTypes = map[string]bool{
	"session":  true,
	"catalog":  true,
	"cache":    true,
	"workflow": true,
	"custom":   true,
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		monitoring.RecordStateFabricOperation("", "", "create", "unauthorized")
		return
	}
	if !h.requireFabricQuota(w, r, tenantID) {
		return
	}
	var req createFabricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), "", "create", "bad_request")
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	if req.Type == "" {
		req.Type = "custom"
	}
	if !allowedFabricTypes[req.Type] {
		apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("invalid fabric type: %s (allowed: session, catalog, cache, workflow, custom)", req.Type)))
		return
	}
	item, err := h.repo.CreateFabric(r.Context(), tenantID, req.Name, req.Description, req.Type, req.Settings)
	if err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), "", "create", "error")
		serverError(w, r, err)
		return
	}

	fabricID := ""
	var userID uuid.UUID
	if item != nil {
		fabricID = item.ID.String()
	}
	_, userID, _ = tenantAndUser(r, w)
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID, "create", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID, "create", time.Since(start))
	monitoring.UpdateStateFabricActiveCount(tenantID.String(), 1)

	if item != nil {
		h.auditLog(r, auditInfo{
			TenantID:   tenantID,
			UserID:     userID,
			ResourceID: item.ID,
			Action:     "state_fabric.create",
			Success:    true,
			Duration:   time.Since(start),
		})
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		monitoring.RecordStateFabricOperation("", "", "read", "unauthorized")
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	item, err := h.repo.GetFabric(r.Context(), tenantID, fabricID)
	if err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "read", "not_found")
		apierror.WriteError(w, apierror.NewNotFound("state fabric not found"))
		return
	}
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "read", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID.String(), "read", time.Since(start))

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		monitoring.RecordStateFabricOperation("", "", "update", "unauthorized")
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	var req updateFabricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "update", "bad_request")
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Settings != nil {
		updates["settings"] = req.Settings
	}
	item, err := h.repo.UpdateFabric(r.Context(), tenantID, fabricID, updates)
	if err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "update", "error")
		serverError(w, r, err)
		return
	}
	_, userID, _ := tenantAndUser(r, w)
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "update", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID.String(), "update", time.Since(start))

	h.auditLog(r, auditInfo{
		TenantID:   tenantID,
		UserID:     userID,
		ResourceID: fabricID,
		Action:     "state_fabric.update",
		Success:    true,
		Duration:   time.Since(start),
	})

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		monitoring.RecordStateFabricOperation("", "", "delete", "unauthorized")
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	if err := h.repo.DeleteFabric(r.Context(), tenantID, fabricID); err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "delete", "error")
		serverError(w, r, err)
		return
	}
	_, userID, _ := tenantAndUser(r, w)
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "delete", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID.String(), "delete", time.Since(start))

	h.auditLog(r, auditInfo{
		TenantID:   tenantID,
		UserID:     userID,
		ResourceID: fabricID,
		Action:     "state_fabric.delete",
		Success:    true,
		Duration:   time.Since(start),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "advanced_insights") {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	item, err := h.repo.GetFabric(r.Context(), tenantID, fabricID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state fabric not found"))
		return
	}
	writeJSON(w, http.StatusOK, item.Metrics)
}

func (h *Handler) HandleListStores(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	stores, err := h.repo.ListStores(r.Context(), tenantID, fabricID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state fabric not found"))
		return
	}
	writeJSON(w, http.StatusOK, stores)
}

func (h *Handler) HandleCreateStore(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	var req createStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	switch req.Type {
	case "vector", "embedding", "ai-memory":
		if !h.requireAddon(w, r, tenantID, "ai_memory_pack") {
			return
		}
	}
	store, err := h.repo.CreateStore(r.Context(), tenantID, fabricID, req.Name, req.Type, req.MaxSize, req.Region)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, store)
}

func (h *Handler) HandleDeleteStore(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	if err := h.repo.DeleteStore(r.Context(), tenantID, fabricID, mux.Vars(r)["storeId"]); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleListPipelines(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	item, err := h.repo.GetFabric(r.Context(), tenantID, fabricID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state fabric not found"))
		return
	}
	writeJSON(w, http.StatusOK, item.Pipelines)
}

func (h *Handler) HandleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	var req createPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	pipeline, err := h.repo.CreatePipeline(r.Context(), tenantID, fabricID, req.Name, req.Description, req.Steps)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, pipeline)
}

func (h *Handler) HandleUpdatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	pipelineID, parsed := parseID(w, vars["pipelineId"], "pipeline id")
	if !parsed {
		return
	}
	var req updatePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Steps != nil {
		updates["steps"] = req.Steps
	}
	pipeline, err := h.repo.UpdatePipeline(r.Context(), tenantID, fabricID, pipelineID, updates)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, pipeline)
}

func (h *Handler) HandleDeletePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	pipelineID, parsed := parseID(w, vars["pipelineId"], "pipeline id")
	if !parsed {
		return
	}
	if err := h.repo.DeletePipeline(r.Context(), tenantID, fabricID, pipelineID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleExecutePipeline(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, userID, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	pipelineID, parsed := parseID(w, vars["pipelineId"], "pipeline id")
	if !parsed {
		return
	}
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	result, err := h.repo.ExecutePipeline(r.Context(), tenantID, fabricID, pipelineID, input)
	if err != nil {
		h.auditLog(r, auditInfo{
			TenantID:    tenantID,
			UserID:      userID,
			ResourceID:  pipelineID,
			Action:      "state_fabric.pipeline.execute",
			Success:     false,
			Duration:    time.Since(start),
			Description: err.Error(),
		})
		serverError(w, r, err)
		return
	}

	h.auditLog(r, auditInfo{
		TenantID:   tenantID,
		UserID:     userID,
		ResourceID: pipelineID,
		Action:     "state_fabric.pipeline.execute",
		Success:    true,
		Duration:   time.Since(start),
	})

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "advanced_security_pack") {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var startTime *time.Time
	if raw := r.URL.Query().Get("startTime"); raw != "" {
		if parsedTime, err := time.Parse(time.RFC3339, raw); err == nil {
			startTime = &parsedTime
		}
	}
	var endTime *time.Time
	if raw := r.URL.Query().Get("endTime"); raw != "" {
		if parsedTime, err := time.Parse(time.RFC3339, raw); err == nil {
			endTime = &parsedTime
		}
	}
	events, total, err := h.repo.ListEvents(r.Context(), tenantID, fabricID, repo.EventListOptions{
		StoreID:   r.URL.Query().Get("storeId"),
		EventType: r.URL.Query().Get("eventType"),
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"events": events, "total": total})
}

func (h *Handler) HandleListSnapshots(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	items, err := h.repo.ListSnapshots(r.Context(), tenantID, fabricID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	var req createSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	item, err := h.repo.CreateSnapshot(r.Context(), tenantID, fabricID, req.Name)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) HandleDeleteSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	snapshotID, parsed := parseID(w, vars["snapshotId"], "snapshot id")
	if !parsed {
		return
	}
	if err := h.repo.DeleteSnapshot(r.Context(), tenantID, fabricID, snapshotID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleListReplays(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "hot_cache_booster") {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	items, err := h.repo.ListReplays(r.Context(), tenantID, fabricID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleCreateReplay(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "hot_cache_booster") {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	var req createReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	item, err := h.repo.CreateReplay(r.Context(), tenantID, fabricID, repo.ReplayCreateRequest(req))
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "hot_cache_booster") {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	item, err := h.repo.GetReplay(r.Context(), tenantID, fabricID, vars["replayId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("replay not found"))
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// HandleResumeReplay handles POST /state-fabrics/{id}/replays/{replayId}/resume - Resume a paused or failed replay
// @Summary Resume a replay session
// @Description Resumes a paused or failed replay session
// @Tags StateFabric
// @Accept json
// @Param id path string true "State Fabric ID"
// @Param replayId path string true "Replay ID"
// @Success 200 {object} ReplaySession
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /state-fabrics/{id}/replays/{replayId}/resume [post]
func (h *Handler) HandleResumeReplay(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !h.requireAddon(w, r, tenantID, "hot_cache_booster") {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}

	replay, err := h.repo.ResumeReplay(r.Context(), tenantID, fabricID, vars["replayId"])
	if err != nil {
		if err.Error() == "replay not found" {
			apierror.WriteError(w, apierror.NewNotFound("replay not found"))
			return
		}
		if strings.Contains(err.Error(), "can only resume") {
			apierror.LogAndBadRequest(w, r, err, "statefabric handler")
			return
		}
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, replay)
}

// HandleHealth handles GET /state-fabrics/health - liveness probe
// @Summary Health check (liveness)
// @Description Returns OK if the StateFabric service is alive
// @Tags StateFabric
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /state-fabrics/health [get]
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

// FeatureFlags returns the current feature flags for StateFabric
// @Summary Get feature flags
// @Description Returns the current feature flags for StateFabric
// @Tags StateFabric
// @Produce json
// @Success 200 {object} StateFabricFeatureFlags
// @Router /state-fabrics/feature-flags [get]
func (h *Handler) HandleGetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	flags := StateFabricFeatureFlags{
		ReplayProgressStreaming: isFeatureEnabled("statefabric_replay_streaming"),
		PipelineCircuitBreaker:  isFeatureEnabled("statefabric_pipeline_circuit_breaker"),
		R2StorageOffload:        isFeatureEnabled("statefabric_r2_offload"),
		AdvancedSecurityPack:    isFeatureEnabled("statefabric_advanced_security"),
		HotCacheBooster:         isFeatureEnabled("statefabric_hot_cache"),
		AIMemoryPack:            isFeatureEnabled("statefabric_ai_memory"),
		VectorSearch:            config.IsVectorSearchEnabled(),
	}
	writeJSON(w, http.StatusOK, flags)
}

// StateFabricFeatureFlags represents the feature flags for StateFabric
type StateFabricFeatureFlags struct {
	ReplayProgressStreaming bool `json:"replay_progress_streaming"`
	PipelineCircuitBreaker  bool `json:"pipeline_circuit_breaker"`
	R2StorageOffload        bool `json:"r2_storage_offload"`
	AdvancedSecurityPack    bool `json:"advanced_security_pack"`
	HotCacheBooster         bool `json:"hot_cache_booster"`
	AIMemoryPack            bool `json:"ai_memory_pack"`
	VectorSearch            bool `json:"vector_search"`
}

// isFeatureEnabled checks if a feature flag is enabled via environment variable
func isFeatureEnabled(flag string) bool {
	return os.Getenv("STATEFABRIC_FEATURE_"+strings.ToUpper(flag)) == "true"
}

// HandleReplayProgress handles GET /state-fabrics/{id}/replays/{replayId}/progress
// Returns streaming replay progress via Server-Sent Events (SSE)
// @Summary Get replay progress (SSE)
// @Description Streams replay progress via Server-Sent Events
// @Tags StateFabric
// @Produce text/event-stream
// @Param id path string true "State Fabric ID"
// @Param replayId path string true "Replay ID"
// @Success 200 {object} ReplayProgress
// @Failure 404 {object} map[string]string
// @Router /state-fabrics/{id}/replays/{replayId}/progress [get]
func (h *Handler) HandleReplayProgress(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	if !isFeatureEnabled("statefabric_replay_streaming") {
		apierror.WriteError(w, apierror.NewForbidden("replay progress streaming not enabled"))
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	replayIDStr := vars["replayId"]

	replay, err := h.repo.GetReplay(r.Context(), tenantID, fabricID, replayIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("replay not found"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		apierror.WriteError(w, apierror.NewInternal("streaming not supported"))
		return
	}

	monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_started")

	const (
		maxStreamingDuration = 30 * time.Minute
		pollInterval         = 500 * time.Millisecond
		idleTimeout         = 5 * time.Minute
	)

	streamStart := time.Now()
	lastProgress := -1
	lastActivity := time.Now()

	for {
		select {
		case <-r.Context().Done():
			monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_cancelled")
			return
		case <-time.After(pollInterval):
			if time.Since(streamStart) > maxStreamingDuration {
				data := fmt.Sprintf("data: {\"error\": \"streaming timeout\", \"code\": \"STREAM_TIMEOUT\"}\n\n")
				w.Write([]byte(data))
				flusher.Flush()
				monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_timeout")
				return
			}

			currentReplay, err := h.repo.GetReplay(r.Context(), tenantID, fabricID, replayIDStr)
			if err != nil {
				data := fmt.Sprintf("data: {\"error\": \"failed to fetch progress\", \"code\": \"FETCH_ERROR\"}\n\n")
				w.Write([]byte(data))
				flusher.Flush()
				monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_error")
				return
			}

			if currentReplay.Progress != lastProgress {
				lastProgress = currentReplay.Progress
				lastActivity = time.Now()
				data := fmt.Sprintf("data: {\"progress\": %d, \"eventsReplayed\": %d, \"status\": \"%s\"}\n\n",
					currentReplay.Progress, currentReplay.EventsReplayed, currentReplay.Status)
				w.Write([]byte(data))
				flusher.Flush()

				if currentReplay.Status == "completed" {
					data := fmt.Sprintf("data: {\"progress\": 100, \"eventsReplayed\": %d, \"status\": \"completed\", \"completed\": true}\n\n",
						currentReplay.EventsReplayed)
					w.Write([]byte(data))
					flusher.Flush()
					monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_completed")
					return
				}
				if currentReplay.Status == "failed" || currentReplay.Status == "cancelled" {
					data := fmt.Sprintf("data: {\"progress\": %d, \"eventsReplayed\": %d, \"status\": \"%s\", \"error\": \"%s\"}\n\n",
						currentReplay.Progress, currentReplay.EventsReplayed, currentReplay.Status,
						escapeErrorMessage(currentReplay.Error))
					w.Write([]byte(data))
					flusher.Flush()
					monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_"+currentReplay.Status)
					return
				}
			} else if time.Since(lastActivity) > idleTimeout && replay.Status != "running" && replay.Status != "pending" {
				monitoring.RecordStateFabricReplayOperation(tenantID.String(), fabricID.String(), "streaming_idle_timeout")
				return
			}
		}
	}
}

func escapeErrorMessage(msg string) string {
	if msg == "" {
		return ""
	}
	escaped := strings.ReplaceAll(msg, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	return escaped
}

// HandleReady handles GET /state-fabrics/ready - readiness probe
// @Summary Health check (readiness)
// @Description Returns OK if StateFabric is ready, including R2 connectivity and pipeline execution status
// @Tags StateFabric
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /state-fabrics/ready [get]
func (h *Handler) HandleReady(w http.ResponseWriter, r *http.Request) {
	health := h.repo.HealthCheck(r.Context())

	r2Storage := health["r2_storage"].(map[string]interface{})
	r2Available, _ := r2Storage["available"].(bool)
	if !r2Available {
		apierror.WriteErrorWithStatus(w, http.StatusServiceUnavailable, apierror.ErrCodeServiceUnavailable, "R2 storage unavailable")
		return
	}

	pipelineExec := health["pipeline_exec"].(map[string]interface{})
	pipelineConfigured, _ := pipelineExec["configured"].(bool)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":              "ready",
		"r2_storage":          r2Storage,
		"pipeline_execution":  pipelineExec,
		"pipeline_configured": pipelineConfigured,
	})
}
