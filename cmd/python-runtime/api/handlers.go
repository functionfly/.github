package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/cmd/python-runtime/executor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	maxCodeSizeBytes    = 1 * 1024 * 1024  // 1MB max code size
	maxBodySizeBytes    = 10 * 1024 * 1024 // 10MB max body size
	authTokenHeader     = "X-Auth-Token"
	rateLimitWindow     = 1 * time.Minute
	rateLimitMaxRequests = 100
)

var (
	executionLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "python_runtime_execution_latency_ms",
		Help:    "Python execution latency in milliseconds",
		Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
	})
	executionErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "python_runtime_execution_errors_total",
		Help: "Total number of Python execution errors",
	})
	pythonRuntimeUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "python_runtime_up",
		Help: "Whether the Python runtime is healthy (1=up, 0=down)",
	})
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "python_runtime_request_duration_ms",
		Help:    "HTTP request duration in milliseconds",
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"method", "endpoint", "status"})
)

type ipRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	blocked  map[string]bool
}

func newIPRateLimiter() *ipRateLimiter {
	return &ipRateLimiter{
		requests: make(map[string][]time.Time),
		blocked:  make(map[string]bool),
	}
}

func (r *ipRateLimiter) allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.blocked[ip] {
		return false
	}

	now := time.Now()
	windowStart := now.Add(-rateLimitWindow)

	if requests, ok := r.requests[ip]; ok {
		var validRequests []time.Time
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		r.requests[ip] = validRequests

		if len(validRequests) >= rateLimitMaxRequests {
			r.blocked[ip] = true
			go func() {
				time.Sleep(rateLimitWindow)
				r.mu.Lock()
				delete(r.blocked, ip)
				r.requests[ip] = nil
				r.mu.Unlock()
			}()
			return false
		}
		r.requests[ip] = append(r.requests[ip], now)
	} else {
		r.requests[ip] = []time.Time{now}
	}

	return true
}

var authTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]{16,128}$`)

func validateAuthToken(token string) bool {
	return authTokenRegex.MatchString(token)
}

// Response types
type ExecuteRequest struct {
	Code      string          `json:"code"`
	Input     json.RawMessage `json:"input"`
	TimeoutMs int             `json:"timeout_ms,omitempty"`
}

type ExecuteResponse struct {
	Output      string `json:"output"`
	Error       string `json:"error,omitempty"`
	LatencyMs   int64  `json:"latency_ms"`
	MemoryBytes uint64 `json:"memory_bytes,omitempty"`
}

type HealthResponse struct {
	Status     string `json:"status"`
	Pooled     int32  `json:"pooled"`
	Active     int32  `json:"active"`
	Executions int64  `json:"executions"`
	Errors     int64  `json:"errors"`
}

type PoolStatsResponse struct {
	Pooled    int32 `json:"pooled"`
	Idle     int32 `json:"idle"`
	Active   int32 `json:"active"`
	Hits     int64 `json:"hits"`
	Misses   int64 `json:"misses"`
	Evictions int64 `json:"evictions"`
}

// Handler creators - exported for use by main.go

// HandleHealth returns a handler for health checks.
func HandleHealth(exec *executor.PythonExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		healthy := exec.Healthy(ctx)
		stats := exec.Stats()

		testExecHealthy := testExecution(exec)

		status := "healthy"
		httpStatus := http.StatusOK
		if !healthy || !testExecHealthy {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			pythonRuntimeUp.Set(0)
		} else {
			pythonRuntimeUp.Set(1)
		}

		resp := HealthResponse{
			Status:     status,
			Pooled:     stats.Pooled,
			Active:     stats.Active,
			Executions: stats.Executions,
			Errors:     stats.Errors,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		json.NewEncoder(w).Encode(resp)
	}
}

func testExecution(exec *executor.PythonExecutor) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := exec.Execute(ctx, "print(1)", nil, 5000)
	return err == nil
}

// HandleExecute returns a handler for execute requests.
func HandleExecute(exec *executor.PythonExecutor, authToken string) http.HandlerFunc {
	limiter := newIPRateLimiter()

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			requestDuration.WithLabelValues(r.Method, "/execute", fmt.Sprintf("%d", http.StatusOK)).Observe(float64(time.Since(start).Milliseconds()))
		}()

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		clientIP := strings.Split(r.RemoteAddr, ":")[0]
		if !limiter.allow(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		if authToken != "" {
			token := r.Header.Get(authTokenHeader)
			if !validateAuthToken(token) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySizeBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		var req ExecuteRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if req.Code == "" {
			http.Error(w, "code is required", http.StatusBadRequest)
			return
		}

		if len(req.Code) > maxCodeSizeBytes {
			http.Error(w, fmt.Sprintf("Code size exceeds maximum of %d bytes", maxCodeSizeBytes), http.StatusBadRequest)
			return
		}

		timeoutMs := req.TimeoutMs
		if timeoutMs <= 0 {
			timeoutMs = 30000
		}
		if timeoutMs > 300000 {
			timeoutMs = 300000
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutMs+5000)*time.Millisecond)
		defer cancel()

		inputBytes, _ := json.Marshal(req.Input)
		if req.Input == nil {
			inputBytes = []byte("{}")
		}

		result, err := exec.Execute(ctx, req.Code, inputBytes, timeoutMs)
		if err != nil {
			executionErrorsTotal.Inc()
			http.Error(w, fmt.Sprintf("Execution failed: %v", err), http.StatusInternalServerError)
			return
		}

		resp := ExecuteResponse{
			LatencyMs:   result.LatencyMs,
			MemoryBytes: result.MemoryBytes,
		}

		if result.Error != "" {
			resp.Error = result.Error
		} else {
			resp.Output = string(result.Output)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleExecuteStream returns a handler for streaming execution.
func HandleExecuteStream(exec *executor.PythonExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For streaming, we'll implement chunked responses
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 50*1024*1024)) // 50MB limit
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		var req ExecuteRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Set headers for streaming
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)

		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
		defer cancel()

		// Parse input to bytes
		inputBytes, _ := json.Marshal(req.Input)
		if req.Input == nil {
			inputBytes = []byte("{}")
		}

		flusher, canFlush := w.(http.Flusher)

		result, err := exec.Execute(ctx, req.Code, inputBytes, req.TimeoutMs)
		if err != nil {
			resp := ExecuteResponse{Error: err.Error()}
			json.NewEncoder(w).Encode(resp)
			if canFlush {
				flusher.Flush()
			}
			return
		}

		resp := ExecuteResponse{
			Output:      string(result.Output),
			LatencyMs:   result.LatencyMs,
			MemoryBytes: result.MemoryBytes,
		}
		json.NewEncoder(w).Encode(resp)
		if canFlush {
			flusher.Flush()
		}
	}
}

// HandlePoolMaintain returns a handler for pool maintenance.
func HandlePoolMaintain(exec *executor.PythonExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Run maintenance in background
		go func() {
			if err := exec.Prewarm(); err != nil {
				log.Printf("Pool maintenance failed: %v", err)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "maintenance started"})
	}
}

// HandlePoolStats returns a handler for pool statistics.
func HandlePoolStats(exec *executor.PythonExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := exec.Stats()

		resp := PoolStatsResponse{
			Pooled: stats.Pooled,
			Idle:   stats.Pooled - stats.Active,
			Active: stats.Active,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
