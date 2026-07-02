package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/circuitbreaker"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// proxyCircuitRecorder bridges the utils.CircuitRecorder interface to the shared breaker manager.
type proxyCircuitRecorder struct {
	breakerMgr *circuitbreaker.Manager
}

func (r *proxyCircuitRecorder) RecordFailure(backendID uuid.UUID) {
	r.breakerMgr.ForBackend(backendID).RecordFailure()
}

// handlePublicRoute handles public routing to deployed applications
func (s *Server) handlePublicRoute(w http.ResponseWriter, r *http.Request) {
	// Start tracing span for public route handling
	ctx, _ := tracing.StartSpan(r.Context(), "handle-public-route")
	defer tracing.Finish(ctx)
	r = r.WithContext(ctx)

	logger := middleware.GetRequestLogger(r)

	// Extract appSlug from route vars (path-based) or first path segment (edge rewrite).
	// The MatcherFunc-based route doesn't set mux vars, so fall back to path parsing.
	appSlug := mux.Vars(r)["appSlug"]
	if appSlug == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 1 {
			appSlug = parts[0]
		}
	}

	logger = logger.WithField("app_slug", appSlug)
	tracing.SetAttribute(ctx, "app_slug", appSlug)

	if appSlug == "" {
		logger.Warn("Invalid app slug")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid app slug"))
		return
	}

	// Get app by slug
	app, err := s.repo.GetAppBySlug(r.Context(), appSlug)
	if err != nil {
		logger.WithError(err).Error("Failed to get app by slug")
		apierror.WriteError(w, apierror.NewInternal("Failed to get app"))
		return
	}
	if app == nil {
		logger.Warn("App not found")
		apierror.WriteError(w, apierror.NewNotFound("App not found"))
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"app_id":    app.ID,
		"tenant_id": app.TenantID,
	})
	tracing.SetAttribute(ctx, "app_id", app.ID.String())
	tracing.SetAttribute(ctx, "tenant_id", app.TenantID.String())

	// Get tenant and check request limits
	tenant, err := s.repo.GetTenantByID(r.Context(), app.TenantID)
	if err != nil {
		logger.WithError(err).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		logger.Error("Tenant not found")
		apierror.WriteError(w, apierror.NewInternal("Tenant not found"))
		return
	}

	logger = logger.WithField("tenant_plan", tenant.Plan)

	// Check monthly request limit
	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	count, err := s.repo.CountRoutingEventsForTenantSince(r.Context(), app.TenantID, startOfMonth)
	if err != nil {
		logger.WithError(err).Error("Failed to count routing events")
		apierror.WriteError(w, apierror.NewInternal("Failed to check request limits"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to get routing decision"))
		return
	}

	if decision.SelectedBackend == nil {
		logger.WithField("reason", decision.Reason).Warn("No backend available for routing")
		apierror.WriteError(w, apierror.NewServiceUnavailable("No backend available"))
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"selected_backend_id": decision.SelectedBackend.ID,
		"backend_provider":    decision.SelectedBackend.Provider,
		"backend_region":      decision.SelectedBackend.Region,
		"routing_reason":      decision.Reason,
	})
	tracing.SetAttribute(ctx, "selected_backend_id", decision.SelectedBackend.ID.String())
	tracing.SetAttribute(ctx, "backend_provider", decision.SelectedBackend.Provider)
	tracing.SetAttribute(ctx, "backend_region", decision.SelectedBackend.Region)
	tracing.SetAttribute(ctx, "routing_reason", decision.Reason)

	logger.Info("Routing request to backend")

	// Record edge traffic for monitoring when routing to FunctionFly Edge
	if decision.SelectedBackend.Provider == "functionfly-edge" || decision.SelectedBackend.Provider == "functionfly" {
		monitoring.RecordEdgeRequestAndMetric()
	}

	// Proxy to selected backend and capture the result.
	// Pass a circuit recorder for immediate failure feedback to the circuit breaker.
	result := utils.ProxyToBackend(w, r, decision.SelectedBackend, decision.FailoverBackends, requestID, &proxyCircuitRecorder{
		breakerMgr: s.healthMonitor.GetBreakerManager(),
	})

	// Record the actual routing result
	err = s.routingSvc.RecordRoutingResult(app.ID, result.BackendID, int(result.LatencyMs), result.Outcome, requestID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"backend_id": result.BackendID,
			"outcome":    result.Outcome,
			"latency_ms": result.LatencyMs,
		}).Error("Failed to record routing result")
	}

	// Record tracing attributes for the proxy result
	tracing.SetAttribute(ctx, "proxy_outcome", result.Outcome)
	tracing.SetAttribute(ctx, "proxy_status_code", result.StatusCode)
	tracing.SetAttribute(ctx, "proxy_latency_ms", result.LatencyMs)
	if result.Outcome != "success" {
		tracing.RecordError(ctx, fmt.Errorf("proxy outcome: %s", result.Outcome))
	}
}
