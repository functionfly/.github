package statefabric

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	repo "github.com/functionfly/functionfly/internal/storage/statefabric"
)

type Handler struct {
	repo       *repo.Repository
	sfAddons   *statefabricaddons.Repository
	cleanupSvc *repo.CleanupService
}

func NewHandler(r *repo.Repository, sfAddons *statefabricaddons.Repository) *Handler {
	return &Handler{repo: r, sfAddons: sfAddons}
}

func NewHandlerWithCleanup(r *repo.Repository, sfAddons *statefabricaddons.Repository, cleanupSvc *repo.CleanupService) *Handler {
	return &Handler{repo: r, sfAddons: sfAddons, cleanupSvc: cleanupSvc}
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErr(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) fabricToAPI(f repo.Fabric, stores []repo.FabricStore, pipelines []repo.Pipeline) map[string]interface{} {
	data, _ := json.Marshal(f)
	var out map[string]interface{}
	_ = json.Unmarshal(data, &out)
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID, _, ok := tenantAndUser(r, w)
	if !ok {
		monitoring.RecordStateFabricOperation("", "", "create", "unauthorized")
		return
	}
	var req createFabricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), "", "create", "bad_request")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.repo.CreateFabric(r.Context(), tenantID, req.Name, req.Description, req.Type, req.Settings)
	if err != nil {
		monitoring.RecordStateFabricOperation(tenantID.String(), "", "create", "error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fabricID := ""
	if item != nil {
		fabricID = item.ID.String()
	}
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID, "create", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID, "create", time.Since(start))
	monitoring.UpdateStateFabricActiveCount(tenantID.String(), 1)

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
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "update", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID.String(), "update", time.Since(start))

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "delete", "success")
	monitoring.RecordStateFabricOperationDuration(tenantID.String(), fabricID.String(), "delete", time.Since(start))

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
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	pipeline, err := h.repo.CreatePipeline(r.Context(), tenantID, fabricID, req.Name, req.Description, req.Steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleExecutePipeline(w http.ResponseWriter, r *http.Request) {
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
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.repo.ExecutePipeline(r.Context(), tenantID, fabricID, pipelineID, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.repo.CreateSnapshot(r.Context(), tenantID, fabricID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.repo.CreateReplay(r.Context(), tenantID, fabricID, repo.ReplayCreateRequest(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
