package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// handlePublicRoute handles public routing to deployed applications
func (s *Server) handlePublicRoute(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetRequestLogger(r)

	// Extract appSlug from path
	vars := mux.Vars(r)
	appSlug := vars["appSlug"]

	logger = logger.WithField("app_slug", appSlug)

	if appSlug == "" {
		logger.Warn("Invalid app slug")
		http.Error(w, "Invalid app slug", http.StatusBadRequest)
		return
	}

	// Get app by slug
	app, err := s.repo.GetAppBySlug(r.Context(), appSlug)
	if err != nil {
		logger.WithError(err).Error("Failed to get app by slug")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil {
		logger.Warn("App not found")
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"app_id":    app.ID,
		"tenant_id": app.TenantID,
	})

	// Get tenant and check request limits
	tenant, err := s.repo.GetTenantByID(r.Context(), app.TenantID)
	if err != nil {
		logger.WithError(err).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		logger.Error("Tenant not found")
		http.Error(w, "Tenant not found", http.StatusInternalServerError)
		return
	}

	logger = logger.WithField("tenant_plan", tenant.Plan)

	// Check monthly request limit
	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	count, err := s.repo.CountRoutingEventsForTenantSince(r.Context(), app.TenantID, startOfMonth)
	if err != nil {
		logger.WithError(err).Error("Failed to count routing events")
		http.Error(w, "Failed to check request limits", http.StatusInternalServerError)
		return
	}

	maxRequests := plans.MaxRequestsPerMonth(tenant.Plan)
	logger = logger.WithFields(logrus.Fields{
		"current_requests": count,
		"max_requests":     maxRequests,
	})

	if count >= maxRequests {
		logger.Warn("Monthly request limit exceeded")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "Monthly request limit exceeded",
			"current_count": count,
			"max_requests":  maxRequests,
			"plan":          tenant.Plan,
			"reset_date":    startOfMonth.AddDate(0, 1, 0).Format(time.RFC3339),
		})
		return
	}

	// Get request ID from header (should be set by logging middleware)
	requestID := r.Header.Get("X-Request-Id")
	logger = logger.WithField("routing_request_id", requestID)

	// Get routing decision (pass tenant plan)
	decision, err := s.routingSvc.SelectBackend(app.ID, r.Method, requestID, tenant.Plan)
	if err != nil {
		logger.WithError(err).WithField("method", r.Method).Error("Failed to get routing decision")
		http.Error(w, "Failed to get routing decision", http.StatusInternalServerError)
		return
	}

	if decision.SelectedBackend == nil {
		logger.WithField("reason", decision.Reason).Warn("No backend available for routing")
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"selected_backend_id": decision.SelectedBackend.ID,
		"backend_provider":    decision.SelectedBackend.Provider,
		"backend_region":      decision.SelectedBackend.Region,
		"routing_reason":      decision.Reason,
	})

	logger.Info("Routing request to backend")

	// Record edge traffic for monitoring when routing to FunctionFly Edge
	if decision.SelectedBackend.Provider == "functionfly-edge" || decision.SelectedBackend.Provider == "functionfly" {
		monitoring.RecordEdgeRequestAndMetric()
	}

	// Proxy to selected backend and capture the result
	result := utils.ProxyToBackend(w, r, decision.SelectedBackend, decision.FailoverBackends, requestID)

	// Record the actual routing result
	err = s.routingSvc.RecordRoutingResult(app.ID, result.BackendID, int(result.LatencyMs), result.Outcome, requestID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"backend_id": result.BackendID,
			"outcome":    result.Outcome,
			"latency_ms": result.LatencyMs,
		}).Error("Failed to record routing result")
	}
}
