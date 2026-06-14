package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns an HTTP handler for Prometheus metrics
func Handler() http.Handler {
	return promhttp.Handler()
}

// HealthHandler provides a comprehensive health check endpoint
func HealthHandler(collector *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := map[string]interface{}{
			"status":  "healthy",
			"service": "monitoring",
			"checks":  map[string]interface{}{},
		}

		checks := health["checks"].(map[string]interface{})

		// Database connectivity check
		dbHealthy := true
		dbError := ""
		if collector.db != nil {
			if _, err := collector.db.GetDatabaseHealthMetrics(ctx); err != nil {
				dbHealthy = false
				dbError = err.Error()
			}
		} else {
			dbHealthy = false
			dbError = "database connection not available"
		}
		checks["database"] = map[string]interface{}{
			"healthy": dbHealthy,
			"error":   dbError,
		}

		// Backend health check - check if we have healthy backends
		backendHealthy := true
		backendError := ""
		if collector.db != nil {
			backends, err := collector.db.GetAllEnabledBackends(r.Context())
			if err != nil {
				backendHealthy = false
				backendError = fmt.Sprintf("failed to get backends: %v", err)
			} else if len(backends) == 0 {
				backendHealthy = false
				backendError = "no enabled backends found"
			} else {
				healthyCount := 0
				for _, backend := range backends {
					recentChecks, err := collector.db.GetRecentHealthChecks(r.Context(), backend.ID, 1)
					if err == nil && len(recentChecks) > 0 && recentChecks[0].OK {
						healthyCount++
					}
				}
				if healthyCount == 0 {
					backendHealthy = false
					backendError = "no healthy backends available"
				}
			}
		} else {
			backendHealthy = false
			backendError = "repository not available"
		}
		checks["backends"] = map[string]interface{}{
			"healthy": backendHealthy,
			"error":   backendError,
		}

		// Overall health determination
		overallHealthy := dbHealthy && backendHealthy
		if !overallHealthy {
			health["status"] = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(health); err != nil {
			// Fallback to basic response if JSON encoding fails
			w.Write([]byte(`{"status": "error", "message": "failed to encode health response"}`))
		}
	}
}