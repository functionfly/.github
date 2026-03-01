// Package localruntime provides a local execution environment for FunctionFly functions
// that mirrors production behavior.
package localruntime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Runtime represents a local FunctionFly function runtime
type Runtime struct {
	manifest        *manifest.Manifest
	function        func(interface{}) (interface{}, error)
	server          *http.Server
	functionJS      string
	watchFiles      []string
	metrics         *monitoring.LocalRuntimeMetricsCollector
	port            int
	activeRequests  int64

	// Registry functionality
	runtimeID       string
	repo            storage.Repository
	instanceID      uuid.UUID
	registered      bool
	heartbeatTicker *time.Ticker
	metricsTicker   *time.Ticker
	shutdownChan    chan struct{}
}

// Option is a functional option for Runtime
type Option func(*Runtime)

// WithFunction sets the function handler
func WithFunction(fn func(interface{}) (interface{}, error)) Option {
	return func(r *Runtime) {
		r.function = fn
	}
}

// WithRepository sets the repository for registry functionality
func WithRepository(repo storage.Repository) Option {
	return func(r *Runtime) {
		r.repo = repo
	}
}

// New creates a new local runtime
func New(m *manifest.Manifest, opts ...Option) (*Runtime, error) {
	r := &Runtime{
		manifest:     m,
		shutdownChan: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	// Generate unique runtime ID
	runtimeID, err := generateRuntimeID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate runtime ID: %w", err)
	}
	r.runtimeID = runtimeID

	// Initialize metrics collector
	metrics, err := monitoring.NewLocalRuntimeMetricsCollector(r.manifest.Runtime)
	if err != nil {
		log.Printf("Warning: Failed to initialize metrics collector: %v", err)
	} else {
		r.metrics = metrics
	}

	// If no function provided, load from file
	if r.function == nil {
		if err := r.loadFunction(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// loadFunction loads the function from the source file
func (r *Runtime) loadFunction() error {
	// Determine source file based on runtime
	var sourceFile string
	switch r.manifest.Runtime {
	case "node18", "node20":
		sourceFile = "index.js"
	case "python3.11":
		sourceFile = "main.py"
	default:
		sourceFile = "index.js"
	}

	// Check if file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("function source file not found: %s", sourceFile)
	}

	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read function source: %w", err)
	}

	r.functionJS = string(data)

	// For now, we'll use a simple eval-like approach
	// In production, this would use a proper JS runtime likeotto or goja
	r.function = r.createSimpleHandler()

	return nil
}

// createSimpleHandler creates a simple handler based on the runtime
func (r *Runtime) createSimpleHandler() func(interface{}) (interface{}, error) {
	switch r.manifest.Runtime {
	case "python3.11":
		return r.pythonHandler
	default:
		return r.jsHandler
	}
}

// jsHandler handles JavaScript functions
func (r *Runtime) jsHandler(input interface{}) (interface{}, error) {
	// Simple passthrough for now - in production would use JS runtime
	inputStr, ok := input.(string)
	if !ok {
		return input, nil
	}
	return inputStr, nil
}

// pythonHandler handles Python functions
func (r *Runtime) pythonHandler(input interface{}) (interface{}, error) {
	// Simple passthrough for now
	inputStr, ok := input.(string)
	if !ok {
		return input, nil
	}
	return inputStr, nil
}

// ServeHTTP handles HTTP requests
func (r *Runtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Record request start
	atomic.AddInt64(&r.activeRequests, 1)
	if r.metrics != nil {
		r.metrics.RecordRequestStart(r.port)
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle OPTIONS
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		atomic.AddInt64(&r.activeRequests, -1)
		if r.metrics != nil {
			r.metrics.RecordRequestEnd(r.port, true, time.Since(start), "cors_preflight")
		}
		return
	}

	// Handle /metrics endpoint for Prometheus scraping
	if req.URL.Path == "/metrics" && req.Method == http.MethodGet {
		r.handleMetrics(w, req)
		atomic.AddInt64(&r.activeRequests, -1)
		if r.metrics != nil {
			r.metrics.RecordRequestEnd(r.port, true, time.Since(start), "metrics")
		}
		return
	}

	// Handle /health endpoint for health checks
	if req.URL.Path == "/health" && req.Method == http.MethodGet {
		r.handleHealth(w, req)
		atomic.AddInt64(&r.activeRequests, -1)
		if r.metrics != nil {
			r.metrics.RecordRequestEnd(r.port, true, time.Since(start), "health")
		}
		return
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		atomic.AddInt64(&r.activeRequests, -1)
		if r.metrics != nil {
			r.metrics.RecordError("read_body_error", r.manifest.Name)
			r.metrics.RecordRequestEnd(r.port, false, time.Since(start), r.manifest.Name)
		}
		return
	}

	input := string(body)
	if len(body) == 0 {
		// Try to get input from query params
		input = req.URL.Query().Get("input")
	}

	// Execute function
	result, err := r.function(input)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		atomic.AddInt64(&r.activeRequests, -1)
		if r.metrics != nil {
			r.metrics.RecordError("function_execution_error", r.manifest.Name)
			r.metrics.RecordRequestEnd(r.port, false, time.Since(start), r.manifest.Name)
		}
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Execution-Time", time.Since(start).String())
	w.Header().Set("X-Functionfly-Runtime", "local")

	// Return result
	response := map[string]interface{}{
		"result":   result,
		"input":    input,
		"runtime":  r.manifest.Runtime,
		"execTime": time.Since(start).Milliseconds(),
	}

	json.NewEncoder(w).Encode(response)

	// Record successful request end
	atomic.AddInt64(&r.activeRequests, -1)
	if r.metrics != nil {
		r.metrics.RecordRequestEnd(r.port, true, time.Since(start), r.manifest.Name)
	}
}

// handleMetrics serves Prometheus metrics
func (r *Runtime) handleMetrics(w http.ResponseWriter, req *http.Request) {
	if r.metrics == nil {
		http.Error(w, "Metrics collection not available", http.StatusServiceUnavailable)
		return
	}

	// Get current runtime metrics
	runtimeMetrics := r.metrics.GetCurrentMetrics()

	// Get system info
	systemInfo, err := r.metrics.GetSystemInfo()
	if err != nil {
		log.Printf("Failed to get system info: %v", err)
	}

	// Format as Prometheus metrics
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_info Local runtime information\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_info gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_info{runtime=\"%s\",function=\"%s\"} 1\n",
		runtimeMetrics.Runtime, r.manifest.Name)

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_memory_usage_bytes Memory usage in bytes\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_memory_usage_bytes{runtime=\"%s\",type=\"heap\"} %d\n",
		runtimeMetrics.Runtime, runtimeMetrics.MemoryUsage.Heap)
	fmt.Fprintf(w, "functionfly_local_runtime_memory_usage_bytes{runtime=\"%s\",type=\"stack\"} %d\n",
		runtimeMetrics.Runtime, runtimeMetrics.MemoryUsage.Stack)
	fmt.Fprintf(w, "functionfly_local_runtime_memory_usage_bytes{runtime=\"%s\",type=\"system\"} %d\n",
		runtimeMetrics.Runtime, runtimeMetrics.MemoryUsage.System)

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_cpu_usage_percent CPU usage percentage\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_cpu_usage_percent gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_cpu_usage_percent{runtime=\"%s\"} %f\n",
		runtimeMetrics.Runtime, runtimeMetrics.CPUUsage)

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_active_connections Number of active connections\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_active_connections gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_active_connections{runtime=\"%s\",port=\"%d\"} %d\n",
		runtimeMetrics.Runtime, r.port, runtimeMetrics.ActiveConnections)

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_request_throughput_per_second Request throughput\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_request_throughput_per_second gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_request_throughput_per_second{runtime=\"%s\"} %f\n",
		runtimeMetrics.Runtime, runtimeMetrics.RequestThroughput)

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_uptime_seconds Runtime uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_uptime_seconds gauge\n")
	fmt.Fprintf(w, "functionfly_local_runtime_uptime_seconds{runtime=\"%s\"} %f\n",
		runtimeMetrics.Runtime, runtimeMetrics.Uptime.Seconds())

	fmt.Fprintf(w, "# HELP functionfly_local_runtime_total_requests Total number of requests\n")
	fmt.Fprintf(w, "# TYPE functionfly_local_runtime_total_requests counter\n")
	fmt.Fprintf(w, "functionfly_local_runtime_total_requests{runtime=\"%s\"} %d\n",
		runtimeMetrics.Runtime, runtimeMetrics.TotalRequests)

	// System metrics
	if systemInfo != nil {
		fmt.Fprintf(w, "# HELP functionfly_system_memory_total_bytes Total system memory\n")
		fmt.Fprintf(w, "# TYPE functionfly_system_memory_total_bytes gauge\n")
		fmt.Fprintf(w, "functionfly_system_memory_total_bytes %v\n", systemInfo["system_memory_total"])

		fmt.Fprintf(w, "# HELP functionfly_system_memory_used_bytes Used system memory\n")
		fmt.Fprintf(w, "# TYPE functionfly_system_memory_used_bytes gauge\n")
		fmt.Fprintf(w, "functionfly_system_memory_used_bytes %v\n", systemInfo["system_memory_used"])

		fmt.Fprintf(w, "# HELP functionfly_system_cpu_cores Number of CPU cores\n")
		fmt.Fprintf(w, "# TYPE functionfly_system_cpu_cores gauge\n")
		fmt.Fprintf(w, "functionfly_system_cpu_cores %v\n", systemInfo["cpu_cores"])

		fmt.Fprintf(w, "# HELP functionfly_go_routines Number of goroutines\n")
		fmt.Fprintf(w, "# TYPE functionfly_go_routines gauge\n")
		fmt.Fprintf(w, "functionfly_go_routines %v\n", systemInfo["go_routines"])
	}
}

// handleHealth serves runtime health check information
func (r *Runtime) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Basic health checks
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"runtime":   r.manifest.Runtime,
		"function":  r.manifest.Name,
		"uptime":    r.metrics.GetCurrentMetrics().Uptime.String(),
	}

	checks := map[string]interface{}{
		"function_loaded": map[string]interface{}{
			"status": "healthy",
			"detail": "Function code is loaded and available",
		},
		"server_running": map[string]interface{}{
			"status": "healthy",
			"detail": fmt.Sprintf("HTTP server listening on port %d", r.port),
		},
		"metrics_collector": map[string]interface{}{
			"status": "healthy",
			"detail": "Metrics collection is active",
		},
	}

	// Check if function file still exists
	if _, err := os.Stat(r.manifest.Name); os.IsNotExist(err) {
		checks["function_loaded"] = map[string]interface{}{
			"status": "unhealthy",
			"detail": "Function file no longer exists",
		}
		health["status"] = "unhealthy"
	}

	// Check if metrics collector is working
	if r.metrics == nil {
		checks["metrics_collector"] = map[string]interface{}{
			"status": "unhealthy",
			"detail": "Metrics collector not initialized",
		}
		health["status"] = "unhealthy"
	} else {
		// Get current metrics to verify collection is working
		currentMetrics := r.metrics.GetCurrentMetrics()
		if currentMetrics.TotalRequests < 0 {
			checks["metrics_collector"] = map[string]interface{}{
				"status": "unhealthy",
				"detail": "Metrics collection appears to be malfunctioning",
			}
			health["status"] = "unhealthy"
		} else {
			checks["metrics_collector"].(map[string]interface{})["detail"] = fmt.Sprintf(
				"Metrics collection active - %d total requests processed",
				currentMetrics.TotalRequests,
			)
		}
	}

	health["checks"] = checks

	// Set appropriate HTTP status code
	if health["status"] == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(health)
}

// Start starts the local runtime server
func (r *Runtime) Start(port int) error {
	r.port = port
	addr := fmt.Sprintf(":%d", port)
	r.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Register with registry if repository is available
	if r.repo != nil {
		if err := r.registerWithRegistry(port); err != nil {
			log.Printf("Warning: Failed to register with registry: %v", err)
		} else {
			r.registered = true
			log.Printf("📋 Registered with runtime registry (ID: %s)", r.runtimeID)
		}
	}

	// Start metrics collection if available
	if r.metrics != nil {
		ctx, cancel := context.WithCancel(context.Background())
		go r.metrics.Start(ctx, 10*time.Second) // Collect every 10 seconds

		// Start metrics reporting if registered
		if r.registered {
			r.startMetricsReporting()
		}

		// Clean up on shutdown
		defer func() {
			cancel()
			r.stopMetricsReporting()
		}()
	}

	// Start heartbeat if registered
	if r.registered {
		r.startHeartbeat()
		defer r.stopHeartbeat()
	}

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down local runtime...")

		// Deregister from registry
		if r.registered {
			if err := r.deregisterFromRegistry(); err != nil {
				log.Printf("Warning: Failed to deregister from registry: %v", err)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.server.Shutdown(ctx)
	}()

	log.Printf("🚀 Local FunctionFly runtime started\n")
	log.Printf("   http://localhost:%d\n", port)
	log.Printf("   Health: http://localhost:%d/health\n", port)
	log.Printf("   Metrics: http://localhost:%d/metrics\n", port)
	if r.registered {
		log.Printf("   Registry ID: %s\n", r.runtimeID)
	}
	log.Printf("   Function: %s (%s)\n", r.manifest.Name, r.manifest.Runtime)
	log.Printf("\nPress Ctrl+C to stop\n")

	return r.server.ListenAndServe()
}

// Run starts the runtime and blocks until stopped
func Run(manifestPath string, port int, watch bool) error {
	// Load manifest
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate
	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	// Create runtime
	rt, err := New(m)
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	// Handle file watching if enabled
	if watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}
		defer watcher.Close()

		watchDir := filepath.Dir(manifestPath)
		if err := watcher.Add(watchDir); err != nil {
			return fmt.Errorf("failed to watch %s: %w", watchDir, err)
		}

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						logrus.Infof("File modified: %s, reloading...", event.Name)
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					logrus.WithError(err).Error("File watcher error")
				}
			}
		}()
	}

	// Start server
	return rt.Start(port)
}

// Response represents a function execution response
type Response struct {
	Result   interface{} `json:"result"`
	Input    interface{} `json:"input"`
	Runtime  string      `json:"runtime"`
	ExecTime int64       `json:"execTime"`
	Error    string      `json:"error,omitempty"`
}

// generateRuntimeID generates a unique identifier for this runtime instance
func generateRuntimeID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("runtime-%s", hex.EncodeToString(bytes)[:16]), nil
}

// registerWithRegistry registers this runtime instance with the central registry
func (r *Runtime) registerWithRegistry(port int) error {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	wd, _ := os.Getwd()
	manifestPath := r.manifest.Name
	if wd != "" {
		manifestPath = fmt.Sprintf("%s/%s", wd, r.manifest.Name)
	}

	instance := &storage.LocalRuntimeInstance{
		RuntimeID:     r.runtimeID,
		RuntimeType:   r.manifest.Runtime,
		FunctionName:  r.manifest.Name,
		ManifestPath:  manifestPath,
		Host:          hostname,
		Port:          port,
		PID:           os.Getpid(),
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        0,
	}

	registered, err := r.repo.RegisterLocalRuntime(context.Background(), instance)
	if err != nil {
		return err
	}

	r.instanceID = registered.ID
	return nil
}

// deregisterFromRegistry removes this runtime instance from the registry
func (r *Runtime) deregisterFromRegistry() error {
	return r.repo.DeregisterLocalRuntime(context.Background(), r.runtimeID)
}

// startHeartbeat starts the periodic heartbeat to the registry
func (r *Runtime) startHeartbeat() {
	r.heartbeatTicker = time.NewTicker(30 * time.Second) // Heartbeat every 30 seconds

	go func() {
		for {
			select {
			case <-r.heartbeatTicker.C:
				if err := r.repo.UpdateLocalRuntimeHeartbeat(context.Background(), r.runtimeID); err != nil {
					log.Printf("Warning: Failed to send heartbeat: %v", err)
				}
			case <-r.shutdownChan:
				return
			}
		}
	}()
}

// stopHeartbeat stops the heartbeat ticker
func (r *Runtime) stopHeartbeat() {
	if r.heartbeatTicker != nil {
		r.heartbeatTicker.Stop()
		r.heartbeatTicker = nil
	}
}

// startMetricsReporting starts periodic metrics reporting to the registry
func (r *Runtime) startMetricsReporting() {
	r.metricsTicker = time.NewTicker(60 * time.Second) // Report metrics every 60 seconds

	go func() {
		for {
			select {
			case <-r.metricsTicker.C:
				r.reportMetrics()
			case <-r.shutdownChan:
				return
			}
		}
	}()
}

// stopMetricsReporting stops the metrics reporting ticker
func (r *Runtime) stopMetricsReporting() {
	if r.metricsTicker != nil {
		r.metricsTicker.Stop()
		r.metricsTicker = nil
	}
	close(r.shutdownChan)
}

// reportMetrics sends current metrics to the registry
func (r *Runtime) reportMetrics() {
	if r.metrics == nil {
		return
	}

	currentMetrics := r.metrics.GetCurrentMetrics()

	metrics := &storage.LocalRuntimeMetric{
		RuntimeInstanceID: r.instanceID,
		Timestamp:         time.Now(),
		MemoryUsage: storage.MemoryStats{
			Heap:   currentMetrics.MemoryUsage.Heap,
			Stack:  currentMetrics.MemoryUsage.Stack,
			System: currentMetrics.MemoryUsage.System,
		},
		CPUUsage:          currentMetrics.CPUUsage,
		ActiveConnections: currentMetrics.ActiveConnections,
		RequestThroughput: currentMetrics.RequestThroughput,
		TotalRequests:     currentMetrics.TotalRequests,
		ErrorRate:         currentMetrics.ErrorRate,
		ExecutionCount:    currentMetrics.TotalRequests, // Simplified - could be enhanced
		AverageLatency:    time.Duration(currentMetrics.Uptime.Milliseconds()) * time.Millisecond, // Placeholder
		ErrorCount:        int64(currentMetrics.ErrorRate * float64(currentMetrics.TotalRequests) / 100), // Estimate
	}

	if err := r.repo.RecordLocalRuntimeMetrics(context.Background(), metrics); err != nil {
		log.Printf("Warning: Failed to report metrics: %v", err)
	}
}
