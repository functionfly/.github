package agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type SwarmControllerHandler struct {
	platformController *swarm.PlatformController
	metricsCollector   *swarm.MetricsCollector
	workerService     *swarm.WorkerService
}

type SwarmControllerHandlerInterface interface {
	RegisterRoutes(router *mux.Router, basePath string)
	HandleGetSwarmStatus(w http.ResponseWriter, r *http.Request)
	HandleTriggerDiscovery(w http.ResponseWriter, r *http.Request)
	HandleTriggerGeneration(w http.ResponseWriter, r *http.Request)
	HandleTriggerScan(w http.ResponseWriter, r *http.Request)
	HandleGetDailyMetrics(w http.ResponseWriter, r *http.Request)
	HandleGetWeeklyMetrics(w http.ResponseWriter, r *http.Request)
	HandleGetTopFunctions(w http.ResponseWriter, r *http.Request)
	HandleGetProductivityScore(w http.ResponseWriter, r *http.Request)
	HandleRecordSnapshot(w http.ResponseWriter, r *http.Request)
	HandleGetChildren(w http.ResponseWriter, r *http.Request)
	HandleChildHeartbeat(w http.ResponseWriter, r *http.Request)
	HandleGetWorkerLogs(w http.ResponseWriter, r *http.Request)
}

func NewSwarmControllerHandler(platformController *swarm.PlatformController, metricsCollector *swarm.MetricsCollector, workerService *swarm.WorkerService) *SwarmControllerHandler {
	return &SwarmControllerHandler{
		platformController: platformController,
		metricsCollector:   metricsCollector,
		workerService:     workerService,
	}
}

var _ SwarmControllerHandlerInterface = (*SwarmControllerHandler)(nil)

func (h *SwarmControllerHandler) RegisterRoutes(router *mux.Router, basePath string) {
	logrus.Info("SwarmControllerHandler: Registering platform controller routes")

	platform := router.PathPrefix(basePath + "/platform").Subrouter()

	platform.HandleFunc("/swarm/status", h.HandleGetSwarmStatus).Methods("GET", "OPTIONS")
	platform.HandleFunc("/swarm/discover", h.HandleTriggerDiscovery).Methods("POST", "OPTIONS")
	platform.HandleFunc("/swarm/generate", h.HandleTriggerGeneration).Methods("POST", "OPTIONS")
	platform.HandleFunc("/swarm/scan", h.HandleTriggerScan).Methods("POST", "OPTIONS")

	platform.HandleFunc("/metrics/daily", h.HandleGetDailyMetrics).Methods("GET", "OPTIONS")
	platform.HandleFunc("/metrics/weekly", h.HandleGetWeeklyMetrics).Methods("GET", "OPTIONS")
	platform.HandleFunc("/metrics/top-functions", h.HandleGetTopFunctions).Methods("GET", "OPTIONS")
	platform.HandleFunc("/metrics/productivity", h.HandleGetProductivityScore).Methods("GET", "OPTIONS")
	platform.HandleFunc("/metrics/snapshot", h.HandleRecordSnapshot).Methods("POST", "OPTIONS")

	platform.HandleFunc("/swarm/children", h.HandleGetChildren).Methods("GET", "OPTIONS")
	platform.HandleFunc("/swarm/child/{agent_id}/heartbeat", h.HandleChildHeartbeat).Methods("POST", "OPTIONS")
	platform.HandleFunc("/swarm/workers/logs", h.HandleGetWorkerLogs).Methods("GET", "OPTIONS")

	logrus.Info("SwarmControllerHandler: All platform controller routes registered")
}

func (h *SwarmControllerHandler) HandleGetSwarmStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.platformController.GetStatus(r.Context())
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"status":       status,
		"retrieved_at": time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleTriggerDiscovery(w http.ResponseWriter, r *http.Request) {
	if err := h.platformController.TriggerDiscoveryScan(r.Context()); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":          true,
		"message":     "Discovery scan triggered",
		"triggered_at": time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleTriggerGeneration(w http.ResponseWriter, r *http.Request) {
	if err := h.platformController.TriggerGeneration(r.Context()); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":          true,
		"message":     "Generation triggered",
		"triggered_at": time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleTriggerScan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source string `json:"source"`
	}

	if err := decodeBody(r, &req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm controller request", err)
		return
	}

	scannerChildren := h.getChildrenByRole("scanner")
	if len(scannerChildren) == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No scanner workers available")
		return
	}

	for _, scanner := range scannerChildren {
		if req.Source == "" || h.hasSourceCapability(scanner, req.Source) {
			if err := h.platformController.DispatchTask(r.Context(), scanner.AgentID, "scan_source", map[string]any{
				"source": req.Source,
			}); err != nil {
				logrus.WithError(err).Warnf("Failed to dispatch scan task to %s", scanner.AgentID)
				continue
			}
		}
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":           true,
		"message":      "Scan triggered",
		"dispatched_to": len(scannerChildren),
		"triggered_at": time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleGetDailyMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsCollector.CollectDailyMetrics(r.Context())
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"metrics": metrics,
	})
}

func (h *SwarmControllerHandler) HandleGetWeeklyMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsCollector.GetWeeklyMetrics(r.Context())
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"metrics": metrics,
	})
}

func (h *SwarmControllerHandler) HandleGetTopFunctions(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n := parseInt(l); n > 0 && n <= 100 {
			limit = n
		}
	}

	topFunctions, err := h.metricsCollector.GetTopPerformingFunctions(r.Context(), limit)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"top_functions": topFunctions,
		"limit":        limit,
	})
}

func (h *SwarmControllerHandler) HandleGetProductivityScore(w http.ResponseWriter, r *http.Request) {
	score, err := h.metricsCollector.GetProductivityScore(r.Context())
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                true,
		"productivity_score": score,
		"calculated_at":      time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleRecordSnapshot(w http.ResponseWriter, r *http.Request) {
	if err := h.metricsCollector.RecordMetricsSnapshot(r.Context()); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"message": "Metrics snapshot recorded",
		"at":      time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleGetChildren(w http.ResponseWriter, r *http.Request) {
	status, err := h.platformController.GetStatus(r.Context())
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"children": status.Children,
		"total":    status.TotalChildren,
	})
}

func (h *SwarmControllerHandler) HandleChildHeartbeat(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id required")
		return
	}

	if err := h.platformController.SendHeartbeat(r.Context(), agentID); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm controller handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"agent_id": agentID,
		"at":       time.Now(),
	})
}

func (h *SwarmControllerHandler) HandleGetWorkerLogs(w http.ResponseWriter, r *http.Request) {
	if h.workerService == nil {
		writeError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "worker service not initialized")
		return
	}

	logs := h.workerService.GetWorkerLogs()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"logs": logs,
	})
}

func (h *SwarmControllerHandler) getChildrenByRole(role string) []*swarm.ChildAgentStatus {
	status, err := h.platformController.GetStatus(nil)
	if err != nil {
		return nil
	}

	var children []*swarm.ChildAgentStatus
	for _, child := range status.Children {
		if cap, ok := child.Capabilities["role"].(string); ok && cap == role {
			children = append(children, &child)
		}
	}
	return children
}

func (h *SwarmControllerHandler) hasSourceCapability(child *swarm.ChildAgentStatus, source string) bool {
	if cap, ok := child.Capabilities["source"].(string); ok && cap == source {
		return true
	}
	return false
}

func decodeBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}