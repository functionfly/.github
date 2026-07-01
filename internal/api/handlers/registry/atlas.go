package registry

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/atlas"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AtlasHandler struct {
	tracer *atlas.Tracer
}

func NewAtlasHandler(tracer *atlas.Tracer) *AtlasHandler {
	return &AtlasHandler{tracer: tracer}
}

func (h *AtlasHandler) HandleGetTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	if h.tracer == nil || !h.tracer.Enabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas tracing is not configured"))
		return
	}

	client := h.tracer.Client()
	if client == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas client not available"))
		return
	}

	run, err := client.GetRun(r.Context(), runID)
	if err != nil {
		logrus.WithError(err).WithField("run_id", runID).Error("Failed to get atlas run")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve trace"))
		return
	}
	if run == nil {
		apierror.WriteError(w, apierror.NewNotFound("Trace not found"))
		return
	}

	events, err := client.GetEvents(r.Context(), runID, 0, 10000)
	if err != nil {
		logrus.WithError(err).WithField("run_id", runID).Error("Failed to get atlas events")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve trace events"))
		return
	}

	graph, err := client.GetGraph(r.Context(), runID)
	if err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("Failed to get atlas graph, returning events only")
	}

	response := map[string]interface{}{
		"run":    run,
		"events": events,
	}
	if graph != nil {
		response["graph"] = graph
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *AtlasHandler) HandleListTraces(w http.ResponseWriter, r *http.Request) {
	if h.tracer == nil || !h.tracer.Enabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas tracing is not configured"))
		return
	}

	client := h.tracer.Client()
	if client == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas client not available"))
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	after := r.URL.Query().Get("after")

	runs, err := client.ListRuns(r.Context(), limit, after)
	if err != nil {
		logrus.WithError(err).Error("Failed to list atlas runs")
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas service unavailable"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"runs":  runs,
		"count": len(runs),
	})
}

func (h *AtlasHandler) HandleSearchTraces(w http.ResponseWriter, r *http.Request) {
	if h.tracer == nil || !h.tracer.Enabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas tracing is not configured"))
		return
	}

	client := h.tracer.Client()
	if client == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas client not available"))
		return
	}

	var req atlas.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	result, err := client.Search(r.Context(), req)
	if err != nil {
		logrus.WithError(err).Error("Failed to search atlas traces")
		apierror.WriteError(w, apierror.NewInternal("Search failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *AtlasHandler) HandleGetGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	if h.tracer == nil || !h.tracer.Enabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas tracing is not configured"))
		return
	}

	client := h.tracer.Client()
	if client == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Atlas client not available"))
		return
	}

	graph, err := client.GetGraph(r.Context(), runID)
	if err != nil {
		logrus.WithError(err).WithField("run_id", runID).Error("Failed to get atlas graph")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve decision graph"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(graph)
}

func (h *AtlasHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if h.tracer == nil || !h.tracer.Enabled() {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "disabled",
			"message": "Atlas tracing is not configured",
		})
		return
	}

	client := h.tracer.Client()
	if client == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unavailable",
			"message": "Atlas client not initialized",
		})
		return
	}

	status, err := client.HealthCheck(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
	})
}
