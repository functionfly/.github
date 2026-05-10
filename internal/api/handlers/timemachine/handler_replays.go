package timemachine

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage/registry"
	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type createReplayRequest struct {
	FunctionID         string `json:"function_id"`
	WindowStart        string `json:"window_start"`
	WindowEnd          string `json:"window_end"`
	TargetVersionID    string `json:"target_version_id"`
	Reason             string `json:"reason"`
	IncidentURL        string `json:"incident_url"`
	ReconciliationMode string `json:"reconciliation_mode"`
	MaxExecutions      int    `json:"max_executions"`
}

type reconciliationRequest struct {
	DryRun bool `json:"dry_run"`
}

func (h *Handler) HandleCreateReplay(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	userID := h.getUserID(r)

	var req createReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if req.Reason == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_REASON", "reason is required")
		return
	}
	if req.FunctionID == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FUNCTION_ID", "function_id is required")
		return
	}
	if req.WindowStart == "" || req.WindowEnd == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_WINDOW", "window_start and window_end are required")
		return
	}
	if req.TargetVersionID == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_TARGET_VERSION", "target_version_id is required")
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_FUNCTION_ID", "Invalid function_id format")
		return
	}

	targetVersionID, err := uuid.Parse(req.TargetVersionID)
	if err != nil {
		if req.TargetVersionID != "latest" && req.TargetVersionID != "stable" && req.TargetVersionID != "previous" {
			h.writeError(w, http.StatusBadRequest, "INVALID_TARGET_VERSION", "Invalid target_version_id format")
			return
		}
		targetVersionID = uuid.Nil
	}

	var targetVersion *registry.RegistryFunctionVersion
	if targetVersionID != uuid.Nil {
		targetVersion, err = h.regRepo.GetFunctionVersionByID(targetVersionID)
		if err != nil || targetVersion == nil {
			h.writeError(w, http.StatusBadRequest, "TARGET_VERSION_NOT_FOUND", "Target version not found")
			return
		}
	} else {
		switch req.TargetVersionID {
		case "latest":
			targetVersion, err = h.regRepo.GetLatestFunctionVersion(functionID)
		case "stable":
			targetVersion, err = h.getStableVersion(functionID)
		case "previous":
			targetVersion, err = h.getPreviousVersion(functionID)
		}
		if err != nil || targetVersion == nil {
			h.writeError(w, http.StatusBadRequest, "TARGET_VERSION_NOT_FOUND", "Target version not found")
			return
		}
		targetVersionID = targetVersion.ID
	}

	windowStart, err := time.Parse(time.RFC3339, req.WindowStart)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_WINDOW_START", "window_start must be RFC3339 format")
		return
	}

	windowEnd, err := time.Parse(time.RFC3339, req.WindowEnd)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_WINDOW_END", "window_end must be RFC3339 format")
		return
	}

	if windowEnd.Before(windowStart) {
		h.writeError(w, http.StatusBadRequest, "INVALID_WINDOW", "window_end must be after window_start")
		return
	}

	plan := h.getPlan(r)
	if plan == "" {
		plan = plans.PlanFree
	}

	fn, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to look up function")
		return
	}
	if fn == nil {
		h.writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", "Function not found")
		return
	}
	if fn.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", "Function not found")
		return
	}

	replayWindowHours := plans.GetReplayWindowHours(plan)
	if replayWindowHours > 0 {
		maxWindow := time.Duration(replayWindowHours) * time.Hour
		if windowEnd.Sub(windowStart) > maxWindow {
			h.writeError(w, http.StatusBadRequest, "WINDOW_EXCEEDS_PLAN_LIMIT",
				"Replay window exceeds plan limit. Upgrade your plan for a longer window.")
			return
		}
	}

	maxExecutions := plans.GetMaxExecutionsPerReplay(plan)
	if req.MaxExecutions > 0 && maxExecutions > 0 && req.MaxExecutions > maxExecutions {
		h.writeError(w, http.StatusBadRequest, "MAX_EXECUTIONS_EXCEEDS_LIMIT",
			"max_executions exceeds plan limit")
		return
	}

	maxConcurrent := plans.GetMaxConcurrentReplays(plan)
	if maxConcurrent > 0 {
		activeReplays, err := h.tmRepo.CountActiveReplaysByTenant(tenantID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check active replays")
			return
		}
		if activeReplays >= int64(maxConcurrent) {
			h.writeError(w, http.StatusConflict, "CONCURRENT_REPLAY_LIMIT",
				"Maximum concurrent replay jobs reached for your plan")
			return
		}
	}

	reconciliationMode := req.ReconciliationMode
	if reconciliationMode == "" {
		reconciliationMode = "dry_run"
	}
	if reconciliationMode == "live" && !plans.SupportsLiveReconciliation(plan) {
		h.writeError(w, http.StatusForbidden, "FEATURE_NOT_AVAILABLE",
			"Live reconciliation requires Enterprise plan or higher")
		return
	}

	execLimit := req.MaxExecutions
	if execLimit == 0 {
		execLimit = maxExecutions
		if execLimit < 0 {
			execLimit = 0
		}
	}

	replay := &tmstorage.Replay{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		UserID:             userID,
		FunctionID:         functionID,
		WindowStart:        windowStart,
		WindowEnd:          windowEnd,
		TargetVersionID:    targetVersionID,
		TargetVersion:      targetVersion.Version,
		MaxExecutions:      execLimit,
		ReconciliationMode: reconciliationMode,
		Status:             "pending",
		Reason:             req.Reason,
	}

	if req.IncidentURL != "" {
		replay.IncidentURL = sql.NullString{String: req.IncidentURL, Valid: true}
	}

	if err := h.tmRepo.CreateReplay(replay); err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create replay")
		return
	}

	h.engine.StartReplay(replay.ID)

	h.writeJSON(w, http.StatusCreated, replay)
}

func (h *Handler) HandleListReplays(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit, offset := h.getPagination(r)

	functionIDStr := r.URL.Query().Get("function_id")
	status := r.URL.Query().Get("status")

	var replays []tmstorage.Replay
	var total int64
	var err error

	if functionIDStr != "" {
		functionID, parseErr := uuid.Parse(functionIDStr)
		if parseErr != nil {
			h.writeError(w, http.StatusBadRequest, "INVALID_FUNCTION_ID", "Invalid function_id format")
			return
		}
		replays, total, err = h.tmRepo.ListReplaysByFunction(functionID, limit, offset)
	} else {
		replays, total, err = h.tmRepo.ListReplaysByTenant(tenantID, limit, offset)
	}

	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list replays")
		return
	}

	if status != "" {
		filtered := make([]tmstorage.Replay, 0, len(replays))
		for _, replay := range replays {
			if replay.Status == status {
				filtered = append(filtered, replay)
			}
		}
		replays = filtered
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":  replays,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	plan := h.getPlan(r)
	if plan == "" {
		plan = plans.PlanFree
	}
	limits := plans.GetTimeMachineLimits(plan)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"replay": replay,
		"limits": limits,
	})
}

func (h *Handler) HandleCancelReplay(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	cancellable := map[string]bool{
		"pending":   true,
		"scanning":  true,
		"replaying": true,
		"diffing":   true,
	}
	if !cancellable[replay.Status] {
		h.writeError(w, http.StatusBadRequest, "CANNOT_CANCEL",
			"Replay cannot be cancelled in its current status")
		return
	}

	if err := h.tmRepo.UpdateReplayStatus(id, "cancelled", replay.ProgressPercent, ""); err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel replay")
		return
	}

	h.engine.CancelReplay(id)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": "cancelled",
	})
}

func (h *Handler) HandleReplayProgress(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	phase := ""
	if replay.CurrentPhase.Valid {
		phase = replay.CurrentPhase.String
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"phase":                    phase,
		"percent":                  replay.ProgressPercent,
		"executions_found":         replay.TotalExecutionsFound,
		"executions_replayed":      replay.TotalExecutionsReplayed,
		"executions_changed":       replay.TotalExecutionsChanged,
		"executions_failed":        replay.TotalExecutionsFailed,
		"status":                   replay.Status,
		"error_message":            nullStringToPtr(replay.ErrorMessage),
	})
}

func (h *Handler) HandleListReplayItems(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	limit, offset := h.getPagination(r)
	diffType := r.URL.Query().Get("diff_type")

	items, total, err := h.tmRepo.ListReplayItems(id, limit, offset, diffType)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list replay items")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetReplayItem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	replayID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	itemID, err := uuid.Parse(vars["itemId"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", "Invalid item ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(replayID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	item, err := h.tmRepo.GetReplayItem(itemID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay item")
		return
	}
	if item == nil || item.ReplayID != replayID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay item not found")
		return
	}

	h.writeJSON(w, http.StatusOK, item)
}

func (h *Handler) HandleDiffSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	diffTypes := []string{"identical", "minor", "major", "breaking", "error"}
	breakdown := make(map[string]int64)
	for _, dt := range diffTypes {
		_, count, listErr := h.tmRepo.ListReplayItems(id, 1, 0, dt)
		if listErr != nil {
			h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute diff summary")
			return
		}
		breakdown[dt] = count
	}

	identical := breakdown["identical"]
	changed := breakdown["minor"] + breakdown["major"] + breakdown["breaking"]
	failed := breakdown["error"]

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_executions": replay.TotalExecutionsFound,
		"identical":        identical,
		"changed":          changed,
		"failed":           failed,
		"breakdown":        breakdown,
	})
}

func (h *Handler) HandleStartReconciliation(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	if replay.Status != "completed" {
		h.writeError(w, http.StatusBadRequest, "REPLAY_NOT_COMPLETED",
			"Replay must be completed before reconciliation")
		return
	}

	plan := h.getPlan(r)
	if plan == "" {
		plan = plans.PlanFree
	}

	var req reconciliationRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	if req.DryRun && !plans.HasFeature(plan, plans.FeatureTimeMachinePro) {
		h.writeError(w, http.StatusForbidden, "FEATURE_NOT_AVAILABLE",
			"Dry-run reconciliation requires Pro plan or higher")
		return
	}

	if !req.DryRun && !plans.SupportsLiveReconciliation(plan) {
		h.writeError(w, http.StatusForbidden, "FEATURE_NOT_AVAILABLE",
			"Live reconciliation requires Enterprise plan or higher")
		return
	}

	isDryRun := req.DryRun || replay.ReconciliationMode == "dry_run"
	if !isDryRun && !plans.SupportsLiveReconciliation(plan) {
		isDryRun = true
	}

	reconPlan, err := h.recEngine.GeneratePlan(id, isDryRun)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate reconciliation plan")
		return
	}

	h.writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"replay_id":           id,
		"reconciliation_mode": replay.ReconciliationMode,
		"dry_run":             isDryRun,
		"plan":                reconPlan,
		"status":              "completed",
	})
}

func (h *Handler) HandleListReconciliations(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	limit, offset := h.getPagination(r)

	recs, total, err := h.tmRepo.ListReconciliations(id, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reconciliations")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":  recs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetAuditCertificate(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	plan := h.getPlan(r)
	if plan == "" {
		plan = plans.PlanFree
	}
	if !plans.SupportsAuditCertificates(plan) {
		h.writeError(w, http.StatusForbidden, "FEATURE_NOT_AVAILABLE",
			"Audit certificates require Enterprise plan or higher")
		return
	}

	cert, err := h.tmRepo.GetAuditCertificateByReplayID(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get audit certificate")
		return
	}
	if cert == nil {
		if replay.Status != "completed" {
			h.writeError(w, http.StatusBadRequest, "REPLAY_NOT_COMPLETED",
				"Replay must be completed before generating an audit certificate")
			return
		}

		cert, err = h.auditGen.Generate(id)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate audit certificate")
			return
		}
	}

	h.writeJSON(w, http.StatusOK, cert)
}

func (h *Handler) HandleGetLimits(w http.ResponseWriter, r *http.Request) {
	plan := h.getPlan(r)
	if plan == "" {
		plan = plans.PlanFree
	}

	limits := plans.GetTimeMachineLimits(plan)
	h.writeJSON(w, http.StatusOK, limits)
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func (h *Handler) HandleReplayStream(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuthOrToken(w, r)
	if claims == nil {
		return
	}
	tenantID := claims.TenantID

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid replay ID")
		return
	}

	replay, err := h.tmRepo.GetReplay(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get replay")
		return
	}
	if replay == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}
	if replay.TenantID != tenantID {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Replay not found")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, http.StatusInternalServerError, "SSE_ERROR", "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	terminalStates := map[string]bool{
		"completed": true,
		"failed":    true,
		"cancelled": true,
	}

	if terminalStates[replay.Status] {
		data, _ := json.Marshal(map[string]interface{}{
			"status":   replay.Status,
			"phase":    nullStringToPtr(replay.CurrentPhase),
			"percent":  replay.ProgressPercent,
			"found":    replay.TotalExecutionsFound,
			"replayed": replay.TotalExecutionsReplayed,
			"changed":  replay.TotalExecutionsChanged,
			"failed":   replay.TotalExecutionsFailed,
		})
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", replay.Status, string(data))
		flusher.Flush()
		return
	}

	channel := fmt.Sprintf("timemachine:progress:%s", id.String())
	sub := h.redis.Subscribe(r.Context(), channel)
	defer sub.Close()

	ch := sub.Channel()

	initialData, _ := json.Marshal(map[string]interface{}{
		"status":  replay.Status,
		"percent": replay.ProgressPercent,
	})
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", replay.Status, string(initialData))
	flusher.Flush()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: progress\ndata: %s\n\n", msg.Payload)
			flusher.Flush()

			var status map[string]interface{}
			if json.Unmarshal([]byte(msg.Payload), &status) == nil {
				if s, ok := status["status"].(string); ok && terminalStates[s] {
					fmt.Fprintf(w, "event: %s\ndata: %s\n\n", s, msg.Payload)
					flusher.Flush()
					return
				}
			}
		case <-ticker.C:
			replay, err := h.tmRepo.GetReplay(id)
			if err != nil || replay == nil {
				return
			}
			data, _ := json.Marshal(map[string]interface{}{
				"status":  replay.Status,
				"percent": replay.ProgressPercent,
			})
			fmt.Fprintf(w, "event: heartbeat\ndata: %s\n\n", string(data))
			flusher.Flush()

			if terminalStates[replay.Status] {
				return
			}
		}
	}
}
