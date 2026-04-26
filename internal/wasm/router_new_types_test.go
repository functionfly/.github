package wasm

import (
	"context"
	"testing"
)

// ─── RuntimeType Tests ─────────────────────────────────────────────────────────

func TestRuntimeTypeFromString_NewTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected RuntimeType
	}{
		// Go
		{"go", RuntimeGoWASM},
		{"go1.21", RuntimeGoWASM},
		{"go-wasm", RuntimeGoWASM},
		// C
		{"c", RuntimeC},
		{"c11", RuntimeC},
		// C++
		{"cpp", RuntimeCpp},
		{"cpp17", RuntimeCpp},
		{"c++", RuntimeCpp},
		// Ruby
		{"ruby", RuntimeRubyWASM},
		{"ruby3.3", RuntimeRubyWASM},
		{"ruby-wasm", RuntimeRubyWASM},
		{"mruby", RuntimeRubyWASM},
		// Kotlin
		{"kotlin", RuntimeKotlinWASM},
		{"kotlin1.9", RuntimeKotlinWASM},
		{"kotlin-wasm", RuntimeKotlinWASM},
		// Swift
		{"swift", RuntimeSwiftWASM},
		{"swift5.9", RuntimeSwiftWASM},
		{"swift-wasm", RuntimeSwiftWASM},
		// Existing types still work
		{"python", RuntimePython},
		{"rust", RuntimeRust},
		{"javascript", RuntimeJavaScript},
		{"browser-wasm", RuntimeBrowserNativeWASM},
		// Unknown
		{"unknown_lang", RuntimeUnknown},
		{"", RuntimeUnknown},
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

func TestRuntimeType_IsValid_NewTypes(t *testing.T) {
	t.Parallel()
	validTypes := []RuntimeType{
		RuntimeGoWASM,
		RuntimeC,
		RuntimeCpp,
		RuntimeRubyWASM,
		RuntimeKotlinWASM,
		RuntimeSwiftWASM,
	}

	for _, rt := range validTypes {
		if !rt.IsValid() {
			t.Errorf("RuntimeType %q should be valid", rt)
		}
	}

	invalidTypes := []RuntimeType{
		RuntimeUnknown,
		RuntimeType("java"),
		RuntimeType("scala"),
		RuntimeType(""),
	}

	for _, rt := range invalidTypes {
		if rt.IsValid() {
			t.Errorf("RuntimeType %q should be invalid", rt)
		}
	}
}

func TestRuntimeType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rt       RuntimeType
		expected string
	}{
		{RuntimeGoWASM, "go-wasm"},
		{RuntimeC, "c"},
		{RuntimeCpp, "cpp"},
		{RuntimeRubyWASM, "ruby-wasm"},
		{RuntimeKotlinWASM, "kotlin-wasm"},
		{RuntimeSwiftWASM, "swift-wasm"},
		{RuntimeRust, "rust"},
		{RuntimePython, "python"},
	}

	for _, tt := range tests {
		if got := tt.rt.String(); got != tt.expected {
			t.Errorf("RuntimeType(%q).String() = %q, want %q", tt.rt, got, tt.expected)
		}
	}
}

// ─── RuntimeRouter Tests ───────────────────────────────────────────────────────

func TestRuntimeRouter_NewRuntimeTypes(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)

	newTypes := []RuntimeType{
		RuntimeGoWASM,
		RuntimeC,
		RuntimeCpp,
		RuntimeRubyWASM,
		RuntimeKotlinWASM,
		RuntimeSwiftWASM,
	}

	dummyProvider := &dummyRuntimeProvider{}
	for _, rt := range newTypes {
		if err := router.RegisterRuntime(rt, dummyProvider); err != nil {
			t.Errorf("RegisterRuntime(%q) failed: %v", rt, err)
		}
	}
}

func TestRuntimeRouter_HasRuntime(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)

	dummyProvider := &dummyRuntimeProvider{}
	for _, rt := range []RuntimeType{RuntimeGoWASM, RuntimeC, RuntimeCpp, RuntimeRubyWASM, RuntimeKotlinWASM, RuntimeSwiftWASM} {
		router.RegisterRuntime(rt, dummyProvider)
		if !router.HasRuntime(rt) {
			t.Errorf("HasRuntime(%q) should be true after registration", rt)
		}
	}

	if router.HasRuntime(RuntimeType("nonexistent")) {
		t.Error("HasRuntime should return false for unregistered type")
	}
}

func TestRuntimeRouter_GetSupportedRuntimes(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	dummyProvider := &dummyRuntimeProvider{}

	registered := []RuntimeType{RuntimeGoWASM, RuntimeC, RuntimeRubyWASM}
	for _, rt := range registered {
		router.RegisterRuntime(rt, dummyProvider)
	}

	supported := router.GetSupportedRuntimes()
	if len(supported) != len(registered) {
		t.Errorf("expected %d supported runtimes, got %d", len(registered), len(supported))
	}
}

func TestRuntimeRouter_GetRuntime_NotRegistered(t *testing.T) {
	t.Parallel()
	router := NewRuntimeRouter(nil)
	_, err := router.GetRuntime(RuntimeGoWASM)
	if err == nil {
		t.Error("expected error for unregistered runtime")
	}
}

func TestRuntimeRouter_CircuitBreaker(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 1e9)
	if !cb.Allow() {
		t.Error("new circuit breaker should allow requests")
	}

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

func TestRuntimeRouter_CircuitBreaker_Recovery(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 0) // Immediate reset
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Error("should be open after failure")
	}

	// With 0 reset timeout, should transition to half-open
	cb.Allow()
	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected half-open, got %v", cb.State())
	}

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after success, got %v", cb.State())
	}
}

// ─── dummyRuntimeProvider for testing ──────────────────────────────────────────

type dummyRuntimeProvider struct{}

func (d *dummyRuntimeProvider) Execute(_ context.Context, _ []byte) ([]byte, error) {
	return []byte(`{"ok":true}`), nil
}

func (d *dummyRuntimeProvider) ExecuteWithConfig(_ context.Context, _ []byte, _ interface{}) ([]byte, error) {
	return []byte(`{"ok":true}`), nil
}

func (d *dummyRuntimeProvider) Close() error {
	return nil
}
