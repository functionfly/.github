package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

func TestServer_ShutdownTimeout(t *testing.T) {
	// Test default shutdown timeout
	server := NewServer(&storage.PostgresDB{})
	defaultTimeout := 30 * time.Second

	if server.GetShutdownTimeout() != defaultTimeout {
		t.Errorf("Expected default shutdown timeout %v, got %v", defaultTimeout, server.GetShutdownTimeout())
	}
}

func TestServer_ConfigurableShutdownTimeout(t *testing.T) {
	// Test configurable shutdown timeout via environment variable
	originalValue := os.Getenv("SHUTDOWN_TIMEOUT")
	defer func() {
		if originalValue != "" {
			os.Setenv("SHUTDOWN_TIMEOUT", originalValue)
		} else {
			os.Unsetenv("SHUTDOWN_TIMEOUT")
		}
	}()

	// Set custom timeout
	os.Setenv("SHUTDOWN_TIMEOUT", "45s")
	server := NewServer(&storage.PostgresDB{})
	expectedTimeout := 45 * time.Second

	if server.GetShutdownTimeout() != expectedTimeout {
		t.Errorf("Expected shutdown timeout %v, got %v", expectedTimeout, server.GetShutdownTimeout())
	}
}

func TestServer_InvalidShutdownTimeout(t *testing.T) {
	// Test that invalid timeout falls back to default
	originalValue := os.Getenv("SHUTDOWN_TIMEOUT")
	defer func() {
		if originalValue != "" {
			os.Setenv("SHUTDOWN_TIMEOUT", originalValue)
		} else {
			os.Unsetenv("SHUTDOWN_TIMEOUT")
		}
	}()

	// Set invalid timeout
	os.Setenv("SHUTDOWN_TIMEOUT", "invalid")
	server := NewServer(&storage.PostgresDB{})
	defaultTimeout := 30 * time.Second

	if server.GetShutdownTimeout() != defaultTimeout {
		t.Errorf("Expected default timeout %v for invalid input, got %v", defaultTimeout, server.GetShutdownTimeout())
	}
}

func TestServer_Shutdown(t *testing.T) {
	server := NewServer(&storage.PostgresDB{})

	// Test shutdown without starting server
	ctx := context.Background()
	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown without server should not error, got: %v", err)
	}

	// Test shutdown with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown should not error, got: %v", err)
	}
}

func TestServer_HTTPTimeouts(t *testing.T) {
	server := NewServer(&storage.PostgresDB{})

	if server.httpServer.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", server.httpServer.ReadTimeout)
	}

	if server.httpServer.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", server.httpServer.WriteTimeout)
	}

	if server.httpServer.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", server.httpServer.IdleTimeout)
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	server := NewServer(&storage.PostgresDB{})

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the health handler directly
	server.handleHealth(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := `{"status": "ok"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected response %q, got %q", expected, w.Body.String())
	}
}

func TestServer_RouterSetup(t *testing.T) {
	server := NewServer(&storage.PostgresDB{})

	// Check that router is initialized
	if server.router == nil {
		t.Error("Router should be initialized")
	}

	// Check that HTTP server is configured
	if server.httpServer == nil {
		t.Error("HTTP server should be initialized")
	}

	if server.httpServer.Handler != server.router {
		t.Error("HTTP server should use the router as handler")
	}
}