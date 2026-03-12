package statefabric

import (
	"encoding/json"
	"net/http"
	"strconv"

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
