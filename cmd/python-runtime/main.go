// Package main provides the Python WASM Runtime Service.
//
// This service executes Python code in WASM using wasmtime (with CGO).
// It is designed to run as a separate process from the main orchestrator,
// allowing the orchestrator to be built with CGO_ENABLED=0 while still
// supporting full Python execution.
//
// Usage:
//   ./python-runtime [--port 8083] [--pool-size 4] [--max-memory-mb 512]
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/cmd/python-runtime/api"
	"github.com/functionfly/functionfly/cmd/python-runtime/executor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := getEnvInt("PORT", 8083)
	poolSize := getEnvInt("POOL_SIZE", 4)
	maxMemoryMB := getEnvInt("MAX_MEMORY_MB", 512)
	cpythonPath := getEnv("CPYTHON_WASM_PATH", "./runtimes/cpython-wasi/python.wasm")
	prewarm := os.Getenv("PREWARM") == "true" || os.Getenv("PREWARM") == "1"

	// Initialize executor
	exec, err := executor.NewPythonExecutor(cpythonPath, poolSize, maxMemoryMB)
	if err != nil {
		log.Fatalf("Failed to create executor: %v", err)
	}
	defer exec.Close()

	if prewarm {
		log.Println("Prewarming runtime pool...")
		if err := exec.Prewarm(); err != nil {
			log.Printf("Warning: prewarm failed: %v", err)
		}
	}

	// Setup HTTP handlers
	mux := http.NewServeMux()

	authToken := os.Getenv("AUTH_TOKEN")

	// Health check
	mux.HandleFunc("GET /health", api.HandleHealth(exec))

	// Execution endpoints
	mux.HandleFunc("POST /execute", api.HandleExecute(exec, authToken))
	mux.HandleFunc("POST /execute/stream", api.HandleExecuteStream(exec))

	// Metrics
	mux.Handle("/metrics", promhttp.Handler())

	// Pool management
	mux.HandleFunc("POST /pool/maintain", api.HandlePoolMaintain(exec))
	mux.HandleFunc("GET /pool/stats", api.HandlePoolStats(exec))

	// Graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Long timeout for streaming
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Python WASM Runtime starting on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown failed: %v", err)
	}
	log.Println("Shutdown complete")
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var intVal int
		if _, err := fmt.Sscanf(val, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultVal
}
