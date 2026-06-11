package statefabric

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListAll (admin) GET /v1/admin/state-fabrics
func (h *Handler) HandleListAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsRead) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	status := q.Get("status")
	var tenantID *uuid.UUID
	if t := q.Get("tenantId"); t != "" {
		parsed, err := uuid.Parse(t)
		if err == nil {
			tenantID = &parsed
		}
	}
	list, total, err := h.repo.ListFabricsAdmin(r.Context(), tenantID, status, limit, offset)
	if err != nil {
		logrus.WithError(err).Warn("admin list state fabrics failed, returning empty list")
		writeJSON(w, http.StatusOK, map[string]interface{}{"fabrics": []interface{}{}, "total": 0})
		return
	}
	fabrics := make([]map[string]interface{}, 0, len(list))
	for _, f := range list {
		stores, _ := h.repo.ListStoresByFabric(r.Context(), f.ID)
		pipelines, _ := h.repo.ListPipelinesByFabric(r.Context(), f.ID)
		fabrics = append(fabrics, h.fabricToAPI(f, stores, pipelines))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"fabrics": fabrics, "total": total})
}

// HandleGetStats (admin) GET /v1/admin/state-fabrics/stats
func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsRead) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	totalFabrics, activeFabrics, totalStores, totalPipelines, totalEvents, storageUsed, err := h.repo.GetAdminStats(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("admin state fabric stats failed, returning zeros")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"totalFabrics": 0, "activeFabrics": 0, "totalStores": 0, "totalPipelines": 0, "totalEvents": 0, "storageUsed": 0,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totalFabrics":   totalFabrics,
		"activeFabrics":  activeFabrics,
		"totalStores":    totalStores,
		"totalPipelines": totalPipelines,
		"totalEvents":    totalEvents,
		"storageUsed":    storageUsed,
	})
}

// HandleSuspendFabric (admin) POST /v1/admin/state-fabrics/{id}/suspend
func (h *Handler) HandleSuspendFabric(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsWrite) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	f, err := h.repo.GetFabricByID(r.Context(), id)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	if err := h.repo.SetFabricSuspended(r.Context(), id, true, body.Reason); err != nil {
		logrus.WithError(err).Error("suspend state fabric")
		writeErr(w, http.StatusInternalServerError, "failed to suspend state fabric")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleResumeFabric (admin) POST /v1/admin/state-fabrics/{id}/resume
func (h *Handler) HandleResumeFabric(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsWrite) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByID(r.Context(), id)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	if err := h.repo.SetFabricSuspended(r.Context(), id, false, ""); err != nil {
		logrus.WithError(err).Error("resume state fabric")
		writeErr(w, http.StatusInternalServerError, "failed to resume state fabric")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleGetSettings (admin) GET /v1/admin/state-fabrics/settings
func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsRead) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	config, err := h.repo.GetPlatformSettings(r.Context())
	if err != nil {
		logrus.WithError(err).Error("get state fabric platform settings")
		writeErr(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	writeJSON(w, http.StatusOK, config)
}

// HandleUpdateSettings (admin) PATCH /v1/admin/state-fabrics/settings
func (h *Handler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermTenantsWrite) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.repo.UpdatePlatformSettings(r.Context(), body); err != nil {
		logrus.WithError(err).Error("update state fabric platform settings")
		writeErr(w, http.StatusInternalServerError, "failed to update settings")
		return
	}
	config, _ := h.repo.GetPlatformSettings(r.Context())
	writeJSON(w, http.StatusOK, config)
}

func hasAdminPermission(claims *auth.Claims, permission string) bool {
	if claims.Role == auth.RoleSuperAdmin || claims.Role == auth.RoleAdmin {
		return true
	}
	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HandleRunTTLCleanup (admin) POST /v1/admin/state-fabrics/cleanup
func (h *Handler) HandleRunTTLCleanup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermSystemWrite) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}

	if h.cleanupSvc == nil {
		writeErr(w, http.StatusServiceUnavailable, "cleanup service not available")
		return
	}

	start := time.Now()
	result, err := h.cleanupSvc.RunManualCleanup(r.Context())
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).Error("state fabric TTL cleanup failed")
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success":    false,
			"message":    "Cleanup failed",
			"error":      err.Error(),
			"durationMs": duration.Milliseconds(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"durationMs":              duration.Milliseconds(),
		"expiredSnapshotsDeleted": result.ExpiredSnapshotsDeleted,
	}).Info("State fabric TTL cleanup completed")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":                 true,
		"message":                 "Cleanup completed successfully",
		"expiredSnapshotsDeleted": result.ExpiredSnapshotsDeleted,
		"snapshotsBefore":         result.SnapshotsBefore,
		"snapshotsAfter":          result.SnapshotsAfter,
		"durationMs":              duration.Milliseconds(),
		"startedAt":               result.StartedAt,
		"completedAt":             result.CompletedAt,
	})
}

// HandleGetTTLCleanupStats (admin) GET /v1/admin/state-fabrics/cleanup/stats
func (h *Handler) HandleGetTTLCleanupStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermSystemRead) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}

	if h.cleanupSvc == nil {
		writeErr(w, http.StatusServiceUnavailable, "cleanup service not available")
		return
	}

	stats, err := h.cleanupSvc.GetStats(r.Context())
	if err != nil {
		logrus.WithError(err).Error("failed to get state fabric cleanup stats")
		writeErr(w, http.StatusInternalServerError, "failed to get cleanup stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"expiredSnapshotsPending": stats["expiredSnapshotsPending"],
		"totalSnapshots":          stats["totalSnapshots"],
		"snapshotsWithExpiration": stats["snapshotsWithExpiration"],
	})
}

// HandleStateFabricHealth returns health status for StateFabric components
func (h *Handler) HandleStateFabricHealth(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !hasAdminPermission(claims, auth.PermSystemRead) {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}

	ctx := r.Context()
	health := map[string]interface{}{
		"status":     "healthy",
		"components": map[string]interface{}{},
		"timestamp":  time.Now(),
	}

	// Check database connectivity
	dbStatus := map[string]interface{}{
		"status":    "healthy",
		"latencyMs": 0,
	}
	dbStart := time.Now()
	if err := h.repo.Ping(ctx); err != nil {
		dbStatus["status"] = "unhealthy"
		dbStatus["error"] = err.Error()
		health["status"] = "degraded"
	}
	dbStatus["latencyMs"] = time.Since(dbStart).Milliseconds()
	health["components"].(map[string]interface{})["database"] = dbStatus

	// Check Redis cache connectivity
	cacheStatus := map[string]interface{}{
		"status":  "unknown",
		"enabled": h.repo.IsCacheEnabled(),
	}
	if h.repo.IsCacheEnabled() {
		cacheStatus["status"] = "healthy"
		if err := h.repo.PingCache(ctx); err != nil {
			cacheStatus["status"] = "unhealthy"
			cacheStatus["error"] = err.Error()
		}
	} else {
		cacheStatus["status"] = "disabled"
	}
	health["components"].(map[string]interface{})["cache"] = cacheStatus

	// Check R2 backend connectivity
	r2Status := map[string]interface{}{
		"status":  "unknown",
		"enabled": h.repo.IsR2Enabled(),
	}
	if h.repo.IsR2Enabled() {
		r2Status["status"] = "healthy"
		if err := h.repo.PingR2(ctx); err != nil {
			r2Status["status"] = "unhealthy"
			r2Status["error"] = err.Error()
			health["status"] = "degraded"
		}
	} else {
		r2Status["status"] = "disabled"
	}
	health["components"].(map[string]interface{})["r2_storage"] = r2Status

	// Check cleanup service
	cleanupStatus := map[string]interface{}{
		"status":  "unknown",
		"enabled": h.cleanupSvc != nil,
	}
	if h.cleanupSvc != nil {
		cleanupStatus["status"] = "healthy"
		if !h.cleanupSvc.IsRunning() {
			cleanupStatus["status"] = "stopped"
		}
	} else {
		cleanupStatus["status"] = "disabled"
	}
	health["components"].(map[string]interface{})["cleanup_service"] = cleanupStatus

	// Check dead letter queue stats
	dlqStatus := map[string]interface{}{
		"status": "healthy",
	}
	totalDLQ, err := h.repo.CountDeadLetters(ctx)
	if err != nil {
		dlqStatus["status"] = "unknown"
		dlqStatus["error"] = err.Error()
	} else {
		dlqStatus["total_pending"] = totalDLQ
		if totalDLQ > 100 {
			dlqStatus["status"] = "warning"
			dlqStatus["message"] = "High number of pending dead letters"
		}
	}
	health["components"].(map[string]interface{})["dead_letter_queue"] = dlqStatus

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, health)
}

// HandleGetFabricHealth returns health status for a specific fabric
func (h *Handler) HandleGetFabricHealth(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid fabric id")
		return
	}

	ctx := r.Context()
	fabric, err := h.repo.GetFabricByID(ctx, fabricID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}

	health := map[string]interface{}{
		"fabric_id":  fabricID.String(),
		"name":       fabric.Name,
		"status":     fabric.Status,
		"timestamp":  time.Now(),
		"components": map[string]interface{}{},
	}

	// Check fabric-specific metrics
	metrics, _ := h.repo.GetMetrics(ctx, fabricID, "")
	health["components"].(map[string]interface{})["metrics"] = map[string]interface{}{
		"status":             "healthy",
		"total_operations":   metrics.TotalOperations,
		"operations_per_sec": metrics.OperationsPerSecond,
		"average_latency_ms": metrics.AverageLatency,
		"error_rate":         metrics.ErrorRate,
	}

	// Check store status
	stores, _ := h.repo.ListStoresByFabric(ctx, fabricID)
	health["components"].(map[string]interface{})["stores"] = map[string]interface{}{
		"status": "healthy",
		"count":  len(stores),
	}

	// Check pipeline status
	pipelines, _ := h.repo.ListPipelinesByFabric(ctx, fabricID)
	activePipelines := 0
	for _, p := range pipelines {
		if p.Status == "active" {
			activePipelines++
		}
	}
	health["components"].(map[string]interface{})["pipelines"] = map[string]interface{}{
		"status": "healthy",
		"total":  len(pipelines),
		"active": activePipelines,
	}

	writeJSON(w, http.StatusOK, health)
}
