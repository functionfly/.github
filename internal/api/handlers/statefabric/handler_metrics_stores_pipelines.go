package statefabric

import (
	"encoding/json"
	"net/http"
	"time"

	statefabricstorage "github.com/functionfly/functionfly/internal/storage/statefabric"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetMetrics GET /v1/state-fabrics/{id}/metrics
func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	// Return aggregated metrics; could be computed from events/stores in future
	metrics := map[string]interface{}{
		"totalOperations":     f.Throughput,
		"operationsPerSecond": 0,
		"averageLatency":      f.LatencyMs,
		"errorRate":           0,
		"storageUsed":         0,
		"lastCalculatedAt":    time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, metrics)
}

// HandleListStores GET /v1/state-fabrics/{id}/stores
func (h *Handler) HandleListStores(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	list, err := h.repo.ListStoresByFabric(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("list stores")
		writeErr(w, http.StatusInternalServerError, "failed to list stores")
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		result = append(result, storeToAPI(s))
	}
	writeJSON(w, http.StatusOK, result)
}

// HandleCreateStore POST /v1/state-fabrics/{id}/stores
func (h *Handler) HandleCreateStore(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	var body struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		MaxSize int64  `json:"maxSize"`
		Region  string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.Type == "" {
		body.Type = "persistent"
	}
	if body.Region == "" {
		body.Region = "default"
	}
	s := &statefabricstorage.StateFabricStore{
		FabricID: fabricID,
		Name:     body.Name,
		Type:     body.Type,
		Status:   "active",
		MaxSize:  body.MaxSize,
		Region:   body.Region,
		Provider: "local",
	}
	if err := h.repo.CreateStore(r.Context(), s); err != nil {
		logrus.WithError(err).Error("create store")
		writeErr(w, http.StatusInternalServerError, "failed to create store")
		return
	}
	writeJSON(w, http.StatusCreated, storeToAPI(s))
}

// HandleDeleteStore DELETE /v1/state-fabrics/{fabricId}/stores/{storeId}
func (h *Handler) HandleDeleteStore(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	storeID, err := uuid.Parse(vars["storeId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid storeId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	if err := h.repo.EnsureStoreBelongsToFabric(r.Context(), storeID, fabricID); err != nil {
		writeErr(w, http.StatusNotFound, "store not found")
		return
	}
	if err := h.repo.DeleteStore(r.Context(), storeID); err != nil {
		logrus.WithError(err).Error("delete store")
		writeErr(w, http.StatusInternalServerError, "failed to delete store")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleListPipelines GET /v1/state-fabrics/{id}/pipelines
func (h *Handler) HandleListPipelines(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), id, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	list, err := h.repo.ListPipelinesByFabric(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("list pipelines")
		writeErr(w, http.StatusInternalServerError, "failed to list pipelines")
		return
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, p := range list {
		result = append(result, pipelineToAPI(p))
	}
	writeJSON(w, http.StatusOK, result)
}

// HandleCreatePipeline POST /v1/state-fabrics/{id}/pipelines
func (h *Handler) HandleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	var body struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Steps       interface{} `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	steps := statefabricstorage.JSONMap{}
	if body.Steps != nil {
		if m, ok := body.Steps.(map[string]interface{}); ok {
			steps = m
		} else if arr, ok := body.Steps.([]interface{}); ok {
			steps["steps"] = arr
		}
	}
	p := &statefabricstorage.StateFabricPipeline{
		FabricID:    fabricID,
		Name:        body.Name,
		Description: body.Description,
		Status:      "draft",
		Steps:       steps,
	}
	if err := h.repo.CreatePipeline(r.Context(), p); err != nil {
		logrus.WithError(err).Error("create pipeline")
		writeErr(w, http.StatusInternalServerError, "failed to create pipeline")
		return
	}
	writeJSON(w, http.StatusCreated, pipelineToAPI(p))
}

// HandleUpdatePipeline PATCH /v1/state-fabrics/{fabricId}/pipelines/{pipelineId}
func (h *Handler) HandleUpdatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	pipelineID, err := uuid.Parse(vars["pipelineId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid pipelineId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	p, err := h.repo.GetPipelineByID(r.Context(), pipelineID)
	if err != nil || p == nil || p.FabricID != fabricID {
		writeErr(w, http.StatusNotFound, "pipeline not found")
		return
	}
	var body struct {
		Name        *string     `json:"name"`
		Description *string     `json:"description"`
		Steps       interface{} `json:"steps"`
		Status      *string     `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Name != nil {
		p.Name = *body.Name
	}
	if body.Description != nil {
		p.Description = *body.Description
	}
	if body.Status != nil {
		p.Status = *body.Status
	}
	if body.Steps != nil {
		if m, ok := body.Steps.(map[string]interface{}); ok {
			p.Steps = m
		} else if arr, ok := body.Steps.([]interface{}); ok {
			p.Steps = statefabricstorage.JSONMap{"steps": arr}
		}
	}
	if err := h.repo.UpdatePipeline(r.Context(), p); err != nil {
		logrus.WithError(err).Error("update pipeline")
		writeErr(w, http.StatusInternalServerError, "failed to update pipeline")
		return
	}
	writeJSON(w, http.StatusOK, pipelineToAPI(p))
}

// HandleDeletePipeline DELETE /v1/state-fabrics/{fabricId}/pipelines/{pipelineId}
func (h *Handler) HandleDeletePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	pipelineID, err := uuid.Parse(vars["pipelineId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid pipelineId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	if err := h.repo.EnsurePipelineBelongsToFabric(r.Context(), pipelineID, fabricID); err != nil {
		writeErr(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err := h.repo.DeletePipeline(r.Context(), pipelineID); err != nil {
		logrus.WithError(err).Error("delete pipeline")
		writeErr(w, http.StatusInternalServerError, "failed to delete pipeline")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleExecutePipeline POST /v1/state-fabrics/{fabricId}/pipelines/{pipelineId}/execute
func (h *Handler) HandleExecutePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mustTenant(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	fabricID, _ := uuid.Parse(vars["id"])
	pipelineID, err := uuid.Parse(vars["pipelineId"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid pipelineId")
		return
	}
	f, err := h.repo.GetFabricByIDAndTenant(r.Context(), fabricID, tenantID)
	if err != nil || f == nil {
		writeErr(w, http.StatusNotFound, "state fabric not found")
		return
	}
	if err := h.repo.EnsurePipelineBelongsToFabric(r.Context(), pipelineID, fabricID); err != nil {
		writeErr(w, http.StatusNotFound, "pipeline not found")
		return
	}
	var input map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if input == nil {
		input = map[string]interface{}{}
	}
	exec := &statefabricstorage.StateFabricPipelineExecution{
		PipelineID: pipelineID,
		Status:     "completed",
		InputData:  statefabricstorage.JSONMap(input),
		OutputData: statefabricstorage.JSONMap{"ok": true},
	}
	if err := h.repo.CreatePipelineExecution(r.Context(), exec); err != nil {
		logrus.WithError(err).Error("execute pipeline")
		writeErr(w, http.StatusInternalServerError, "failed to execute pipeline")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"executionId": exec.ID.String(),
		"status":      exec.Status,
	})
}
