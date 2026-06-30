package wasm

import (
	"context"
	"testing"
)

// ─── Swift WASM Runtime Registration ──────────────────────────────────────────

func TestSwiftWASM_Registration(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	if err := router.RegisterRuntime(RuntimeSwiftWASM, provider); err != nil {
		t.Fatalf("RegisterRuntime(RuntimeSwiftWASM) failed: %v", err)
	}

	if !router.HasRuntime(RuntimeSwiftWASM) {
		t.Fatal("HasRuntime(RuntimeSwiftWASM) should be true after registration")
	}
}

func TestSwiftWASM_RegistrationRejectsInvalid(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	err := router.RegisterRuntime(RuntimeType("not-a-runtime"), provider)
	if err == nil {
		t.Fatal("expected error for invalid runtime type")
	}
}

func TestSwiftWASM_GetRuntime(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	router.RegisterRuntime(RuntimeSwiftWASM, provider)

	got, err := router.GetRuntime(RuntimeSwiftWASM)
	if err != nil {
		t.Fatalf("GetRuntime(RuntimeSwiftWASM) failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetRuntime(RuntimeSwiftWASM) returned nil provider")
	}
}

func TestSwiftWASM_GetRuntime_NotRegistered(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)

	_, err := router.GetRuntime(RuntimeSwiftWASM)
	if err == nil {
		t.Fatal("expected error for unregistered swift-wasm runtime")
	}
}

func TestSwiftWASM_Execute(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	router.RegisterRuntime(RuntimeSwiftWASM, provider)

	output, err := router.Execute(context.Background(), RuntimeSwiftWASM, []byte(`{"test": true}`))
	if err != nil {
		t.Fatalf("Execute(RuntimeSwiftWASM) failed: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSwiftWASM_ExecuteWithTimeout(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	router.RegisterRuntime(RuntimeSwiftWASM, provider)

	output, err := router.ExecuteWithTimeout(
		context.Background(),
		RuntimeSwiftWASM,
		[]byte(`{"test": true}`),
		5_000_000_000, // 5 seconds
	)
	if err != nil {
		t.Fatalf("ExecuteWithTimeout(RuntimeSwiftWASM) failed: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSwiftWASM_ExecuteUnregistered(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)

	_, err := router.Execute(context.Background(), RuntimeSwiftWASM, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error when executing unregistered swift-wasm runtime")
	}
}

// ─── Swift Runtime Type Parsing ───────────────────────────────────────────────

func TestSwiftWASM_RuntimeTypeParsing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected RuntimeType
	}{
		{"swift", RuntimeSwiftWASM},
		{"swift5.9", RuntimeSwiftWASM},
		{"swift-wasm", RuntimeSwiftWASM},
		{"SWIFT", RuntimeUnknown},     // case-sensitive
		{"Swift", RuntimeUnknown},     // case-sensitive
		{"swift-wasm-extra", RuntimeUnknown}, // exact match only
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RuntimeTypeFromString(tt.input)
			if result != tt.expected {
				t.Errorf("RuntimeTypeFromString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSwiftWASM_IsValid(t *testing.T) {
	t.Parallel()
	if !RuntimeSwiftWASM.IsValid() {
		t.Error("RuntimeSwiftWASM should be valid")
	}
}

func TestSwiftWASM_String(t *testing.T) {
	t.Parallel()
	if got := RuntimeSwiftWASM.String(); got != "swift-wasm" {
		t.Errorf("RuntimeSwiftWASM.String() = %q, want %q", got, "swift-wasm")
	}
}

// ─── Swift in Supported Runtimes ──────────────────────────────────────────────

func TestSwiftWASM_InSupportedRuntimes(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	router.RegisterRuntime(RuntimeSwiftWASM, provider)

	supported := router.GetSupportedRuntimes()
	found := false
	for _, rt := range supported {
		if rt == RuntimeSwiftWASM {
			found = true
			break
		}
	}
	if !found {
		t.Error("RuntimeSwiftWASM should appear in GetSupportedRuntimes()")
	}
}

// ─── Swift in Health Check ────────────────────────────────────────────────────

func TestSwiftWASM_InHealthCheck(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	provider := &dummyRuntimeProvider{}

	router.RegisterRuntime(RuntimeSwiftWASM, provider)

	health := router.HealthCheck()
	if health.Status != "healthy" {
		t.Errorf("expected healthy status, got %q", health.Status)
	}
	if !health.Runtimes["swift-wasm"] {
		t.Error("swift-wasm should appear in health check runtimes")
	}
}

// ─── Swift with Circuit Breaker ───────────────────────────────────────────────

func TestSwiftWASM_CircuitBreaker(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 1e9)
	if !cb.Allow() {
		t.Fatal("new circuit breaker should allow requests")
	}

	// Simulate success
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Errorf("expected CircuitClosed after success, got %v", cb.State())
	}

	// Simulate failures
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.Allow() {
		t.Error("circuit breaker should be open after 3 failures")
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected CircuitOpen, got %v", cb.State())
	}
}

// ─── Swift WASM Security Config ───────────────────────────────────────────────

func TestSwiftWASM_SecurityConfig(t *testing.T) {
	t.Parallel()
	config := NewDefaultSecurityConfig()

	// Verify defaults are reasonable for Swift
	if config.MaxMemory < 64*1024*1024 {
		t.Errorf("MaxMemory too low for Swift: %d", config.MaxMemory)
	}
	if config.MaxExecutionTime.Seconds() < 10 {
		t.Errorf("MaxExecutionTime too low for Swift: %v", config.MaxExecutionTime)
	}
	if config.MaxInputSize == 0 {
		t.Error("MaxInputSize should not be 0")
	}
	if config.MaxOutputSize == 0 {
		t.Error("MaxOutputSize should not be 0")
	}
}

func TestSwiftWASM_InputValidation(t *testing.T) {
	t.Parallel()
	config := NewDefaultSecurityConfig()

	// Normal input should pass
	if !config.ValidateInputSize(1024) {
		t.Error("1024 bytes should pass default input validation")
	}

	// Oversized input should fail
	if config.ValidateInputSize(config.MaxInputSize + 1) {
		t.Error("input exceeding MaxInputSize should fail validation")
	}
}

func TestSwiftWASM_DomainAllowlist(t *testing.T) {
	t.Parallel()

	// Empty allowlist = default-deny (no domains allowed)
	config := NewDefaultSecurityConfig()
	if config.IsDomainAllowed("example.com") {
		t.Error("empty allowlist should deny all domains (default-deny)")
	}

	// Explicit wildcard = allow all
	config.AllowedDomains = []string{"*"}
	if !config.IsDomainAllowed("example.com") {
		t.Error("wildcard should allow all domains")
	}

	// With allowlist
	config.AllowedDomains = []string{"api.example.com", "cdn.jsdelivr.net"}
	if !config.IsDomainAllowed("api.example.com") {
		t.Error("should allow listed domain")
	}
	if config.IsDomainAllowed("evil.com") {
		t.Error("should block unlisted domain")
	}
	if !config.IsDomainAllowed("sub.api.example.com") {
		t.Error("should allow subdomain of listed domain")
	}
}
