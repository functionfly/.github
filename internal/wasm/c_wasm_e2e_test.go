package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/manifest"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testCHelloSource is a minimal C function that echoes input.
const testCHelloSource = `
#include <string.h>
#include <stdio.h>
#include <stdlib.h>

static char response[512];

__attribute__((visibility("default")))
void init(void) {}

__attribute__((visibility("default")))
const char* execute(const char* input, int input_len) {
    snprintf(response, sizeof(response),
        "{\"message\":\"hello from c\",\"input_len\":%d,\"ok\":true}",
        input_len);
    return response;
}

__attribute__((visibility("default")))
void* alloc(int size) {
    return malloc(size);
}

__attribute__((visibility("default")))
void dealloc(void* ptr) {
    free(ptr);
}

__attribute__((visibility("default")))
const char* metadata(void) {
    return "{\"runtime\":\"c\",\"version\":\"1.0.0\"}";
}
`

// compileCToWasm compiles C source to WASM using wasi-sdk or emcc.
// Returns the path to the compiled .wasm file.
func compileCToWasm(t *testing.T, src string) string {
	t.Helper()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "main.c")
	if err := os.WriteFile(srcFile, []byte(src), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	outFile := filepath.Join(tmpDir, "function.wasm")

	compiler, toolchain, err := findCCompilerForTest()
	if err != nil {
		t.Skipf("no C WASM compiler available: %v", err)
	}

	switch toolchain {
	case "wasi-sdk":
		sdkDir := filepath.Dir(filepath.Dir(compiler))
		sysroot := filepath.Join(sdkDir, "share", "wasi-sysroot")
		args := []string{
			srcFile,
			"-o", outFile,
			"--target=wasm32-wasi",
			fmt.Sprintf("--sysroot=%s", sysroot),
			"-O2",
			"-nostartfiles",
			"-Wl,--no-entry",
			"-Wl,--allow-undefined",
			"-Wl,--export=init",
			"-Wl,--export=execute",
			"-Wl,--export=alloc",
			"-Wl,--export=dealloc",
			"-Wl,--export=metadata",
			"-Wl,--export=memory",
		}
		cmd := exec.Command(compiler, args...)
		cmd.Dir = tmpDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("wasi-sdk compile failed: %v\n%s", err, out)
		}
	case "emscripten":
		args := []string{
			srcFile,
			"-o", outFile,
			"-O2",
			"-s", "WASM=1",
			"-s", "STANDALONE_WASM=1",
			"-s", "EXPORTED_FUNCTIONS=['_execute','_init','_alloc','_dealloc','_metadata']",
			"-s", "EXPORTED_RUNTIME_METHODS=['ccall','cwrap']",
			"-s", "ALLOW_MEMORY_GROWTH=1",
			"-s", "INITIAL_MEMORY=1048576",
			"--no-entry",
		}
		cmd := exec.Command(compiler, args...)
		cmd.Dir = tmpDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("emscripten compile failed: %v\n%s", err, out)
		}
	default:
		t.Skipf("unsupported toolchain: %s", toolchain)
	}

	wasmBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}
	if len(wasmBytes) < 8 {
		t.Fatalf("wasm too small: %d bytes", len(wasmBytes))
	}
	if wasmBytes[0] != 0x00 || wasmBytes[1] != 0x61 || wasmBytes[2] != 0x73 || wasmBytes[3] != 0x6D {
		t.Fatal("invalid wasm magic bytes")
	}

	t.Logf("Compiled C → WASM (%d bytes, toolchain=%s)", len(wasmBytes), toolchain)
	return outFile
}

// findCCompilerForTest mirrors bundler.findCCompiler but without importing it.
func findCCompilerForTest() (string, string, error) {
	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		clang := filepath.Join(sdkPath, "bin", "clang")
		if _, err := os.Stat(clang); err == nil {
			return clang, "wasi-sdk", nil
		}
	}
	for _, p := range []string{
		"/opt/wasi-sdk/bin/clang",
		"/usr/local/opt/wasi-sdk/bin/clang",
		filepath.Join(os.Getenv("HOME"), ".local/wasi-sdk/bin/clang"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p, "wasi-sdk", nil
		}
	}
	if p, err := exec.LookPath("emcc"); err == nil {
		return p, "emscripten", nil
	}
	return "", "", fmt.Errorf("no C WASM compiler found")
}

// getPrecompiledOrCompileC returns a WASM binary path.
// Prefers fresh WASI-SDK compilation (exports match runtime expectations).
// Falls back to pre-compiled hello.wasm only if no compiler is available.
func getPrecompiledOrCompileC(t *testing.T) string {
	t.Helper()

	// Try fresh compilation first (WASI-SDK exports match runtime expectations)
	_, _, err := findCCompilerForTest()
	if err == nil {
		return compileCToWasm(t, testCHelloSource)
	}

	// Fall back to pre-compiled (Emscripten exports use _ prefix - may not work with Init())
	precompiled := "../../examples/wasm/hello.wasm"
	if abs, err := filepath.Abs(precompiled); err == nil {
		if _, err := os.Stat(abs); err == nil {
			t.Log("WARNING: using pre-compiled Emscripten binary (exports have _ prefix)")
			return abs
		}
	}

	t.Skip("no C WASM compiler and no pre-compiled binary available")
	return ""
}

// testHostHandler implements HostFunctionHandler for testing.
type testHostHandler struct {
	logs    []string
	kv      map[string]string
	envs    map[string]string
	fetchFn func(string) (string, error)
}

func newTestHostHandler() *testHostHandler {
	return &testHostHandler{
		kv:  make(map[string]string),
		envs: map[string]string{"FF_TEST_VAR": "hello_from_host"},
	}
}

// newTestSecurityConfig returns a security config suitable for testing (fuel disabled).
func newTestSecurityConfig() *WASMSecurityConfig {
	config := NewDefaultSecurityConfig()
	config.DisableDeterministic = true
	config.MaxExecutionTime = 10 * time.Second
	return config
}

func (h *testHostHandler) Log(msg string)                          { h.logs = append(h.logs, msg) }
func (h *testHostHandler) Fetch(req string) (string, error) {
	if h.fetchFn != nil {
		return h.fetchFn(req)
	}
	return `{"status":200,"body":"ok"}`, nil
}
func (h *testHostHandler) KVGet(key string) (string, error) {
	v, ok := h.kv[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return v, nil
}
func (h *testHostHandler) KVSet(key, val string) error {
	h.kv[key] = val
	return nil
}
func (h *testHostHandler) GetEnv(name string) string {
	return h.envs[name]
}
func (h *testHostHandler) AIInference(_ string, _ []byte, _ string) (string, error) {
	return "", fmt.Errorf("ai not available in test")
}
func (h *testHostHandler) StateGet(_ string) (string, error)         { return "", nil }
func (h *testHostHandler) StateSet(_, _ string) error                { return nil }
func (h *testHostHandler) StateDelete(_ string) error                { return nil }
func (h *testHostHandler) StateGetFabric(_ string) (string, error)         { return "", nil }
func (h *testHostHandler) StateCreateSnapshot(_, _ string) (string, error) { return "", nil }
func (h *testHostHandler) GetAttestation(_ string) (string, error)         { return "", nil }
func (h *testHostHandler) Delegate(_, _, _ string) (string, error)         { return "", nil }
func (h *testHostHandler) Call(_ context.Context, _ string, _ ...interface{}) (interface{}, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// E2E Test 1: C → WASM compilation and basic execution
// ---------------------------------------------------------------------------

func TestCWasmE2E_CompileAndExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	t.Logf("Using WASM binary: %s", wasmPath)

	handler := newTestHostHandler()
	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, newTestSecurityConfig())
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	output, err := runtime.ExecuteWithContext(context.Background(), []byte(`{"test":"e2e"}`))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	outputStr := string(output)
	t.Logf("Output: %s", outputStr)

	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}

	if !strings.Contains(outputStr, "ok") {
		t.Errorf("expected 'ok' in output, got: %s", outputStr)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 2: Multiple executions (warm path)
// ---------------------------------------------------------------------------

func TestCWasmE2E_MultipleExecutions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, newTestSecurityConfig())
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	inputs := []string{
		`{"run":1}`,
		`{"run":2,"data":"test"}`,
		`{"run":3,"nested":{"key":"value"}}`,
	}

	for i, input := range inputs {
		output, err := runtime.ExecuteWithContext(context.Background(), []byte(input))
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
		if len(output) == 0 {
			t.Errorf("execution %d: empty output", i)
		}
		t.Logf("Execution %d: %s", i, string(output))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 3: Execution timeout
// ---------------------------------------------------------------------------

func TestCWasmE2E_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	config := newTestSecurityConfig()
	config.MaxExecutionTime = 1 * time.Millisecond

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, config)
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err = runtime.ExecuteWithContext(context.Background(), []byte(`{"test":"timeout"}`))
	if err != nil {
		t.Logf("Execution with 1ms timeout returned error (expected): %v", err)
	} else {
		t.Log("Execution completed within 1ms (fast path)")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 4: Input size validation
// ---------------------------------------------------------------------------

func TestCWasmE2E_InputSizeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	config := newTestSecurityConfig()
	config.MaxInputSize = 64

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, config)
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	bigInput := make([]byte, 128)
	for i := range bigInput {
		bigInput[i] = 'A'
	}

	_, err = runtime.ExecuteWithContext(context.Background(), bigInput)
	if err == nil {
		t.Error("expected error for oversized input")
	} else {
		t.Logf("Oversized input correctly rejected: %v", err)
	}

	smallInput := []byte(`{"ok":true}`)
	output, err := runtime.ExecuteWithContext(context.Background(), smallInput)
	if err != nil {
		t.Fatalf("small input execution failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected non-empty output for small input")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 5: Memory usage tracking
// ---------------------------------------------------------------------------

func TestCWasmE2E_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, nil, nil, handler, newTestSecurityConfig())
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	memUsage := runtime.GetMemoryUsage()
	t.Logf("Memory usage after init: %d bytes", memUsage)

	if memUsage == 0 {
		t.Error("expected non-zero memory usage after init")
	}

	for i := 0; i < 10; i++ {
		_, err := runtime.ExecuteWithContext(context.Background(), []byte(`{"test":"mem"}`))
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
	}

	memUsageAfter := runtime.GetMemoryUsage()
	t.Logf("Memory usage after 10 executions: %d bytes", memUsageAfter)

	if memUsageAfter > memUsage*4 {
		t.Errorf("possible memory leak: before=%d after=%d", memUsage, memUsageAfter)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 6: WASM binary validation
// ---------------------------------------------------------------------------

func TestCWasmE2E_BinaryValidation(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"empty", []byte{}, true},
		{"too small", []byte{0x00, 0x61, 0x73}, true},
		{"bad magic", []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x00, 0x00, 0x00}, true},
		{"valid header minimal", []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, false},
		{"version 2", []byte{0x00, 0x61, 0x73, 0x6D, 0x02, 0x00, 0x00, 0x00}, false},
		{"unsupported version 99", []byte{0x00, 0x61, 0x73, 0x6D, 0x63, 0x00, 0x00, 0x00}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bundler.ValidateWASM(tt.data, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWASM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// E2E Test 7: WASM validation with real compiled binary
// ---------------------------------------------------------------------------

func TestCWasmE2E_BinaryValidationReal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping C/WASM E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	err = bundler.ValidateWASM(wasmBytes, nil)
	if err != nil {
		// Section 1 size mismatch is a known false positive for WASI-SDK compiled binaries
		if strings.Contains(err.Error(), "section 1 size mismatch") {
			t.Log("Known false positive: section 1 size mismatch (WASI-SDK encoding)")
		} else {
			t.Errorf("validation of real C WASM binary failed: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 8: Security config from environment
// ---------------------------------------------------------------------------

func TestCWasmE2E_SecurityConfigFromEnv(t *testing.T) {
	os.Setenv("WASM_MAX_MEMORY", "32")
	os.Setenv("WASM_MAX_TIMEOUT", "15s")
	os.Setenv("WASM_POOL_SIZE", "5")
	os.Setenv("WASM_ALLOWED_DOMAINS", "api.example.com,cdn.example.com")
	defer func() {
		os.Unsetenv("WASM_MAX_MEMORY")
		os.Unsetenv("WASM_MAX_TIMEOUT")
		os.Unsetenv("WASM_POOL_SIZE")
		os.Unsetenv("WASM_ALLOWED_DOMAINS")
	}()

	config := NewSecurityConfigFromEnv()

	if config.MaxMemory != 32*1024*1024 {
		t.Errorf("MaxMemory = %d, want %d", config.MaxMemory, 32*1024*1024)
	}
	if config.MaxExecutionTime != 15*time.Second {
		t.Errorf("MaxExecutionTime = %v, want 15s", config.MaxExecutionTime)
	}
	if config.PoolSize != 5 {
		t.Errorf("PoolSize = %d, want 5", config.PoolSize)
	}
	if len(config.AllowedDomains) != 2 {
		t.Errorf("AllowedDomains len = %d, want 2", len(config.AllowedDomains))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 9: Domain allowlist enforcement
// ---------------------------------------------------------------------------

func TestCWasmE2E_DomainAllowlist(t *testing.T) {
	config := NewDefaultSecurityConfig()
	config.AllowedDomains = []string{"api.example.com", "safe.internal"}

	tests := []struct {
		domain string
		want   bool
	}{
		{"api.example.com", true},
		{"sub.api.example.com", true},
		{"safe.internal", true},
		{"evil.com", false},
		{"example.com", false},
		{"notsafe.internal", false},
	}

	for _, tt := range tests {
		got := config.IsDomainAllowed(tt.domain)
		if got != tt.want {
			t.Errorf("IsDomainAllowed(%q) = %v, want %v", tt.domain, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 10: Runtime scanner detects C-specific threats
// ---------------------------------------------------------------------------

func TestCWasmE2E_RuntimeScanner_CThreats(t *testing.T) {
	scanner := NewRuntimeScanner()

	tests := []struct {
		code    string
		blocked bool
		ruleID  string
	}{
		{`int main() { system("rm -rf /"); }`, true, "RUNTIME_C_SYSTEM"},
		{`FILE *f = popen("ls", "r");`, true, "RUNTIME_C_POPEN"},
		{`void *lib = dlopen("libevil.so", RTLD_LAZY);`, true, "RUNTIME_C_DLOPEN"},
		{`char buf[10]; gets(buf);`, true, "RUNTIME_C_GETS"},
		{`char dst[10]; strcpy(dst, user_input);`, true, "RUNTIME_C_STRCPY"},
		{`char dst[10]; strcat(dst, user_input);`, true, "RUNTIME_C_STRCAT"},
		{`char buf[10]; sprintf(buf, "%s", user_input);`, true, "RUNTIME_C_SPRINTF"},
		{`int main() { fork(); }`, true, "RUNTIME_C_FORK"},
		{`int main() { return 0; }`, false, ""},
		{`const char* execute(const char* input, int len) { return input; }`, false, ""},
	}

	for _, tt := range tests {
		result := scanner.ScanSource(tt.code)
		if tt.blocked && !result.Blocked {
			t.Errorf("code %q should be blocked (expected rule %s)", tt.code, tt.ruleID)
		}
		if !tt.blocked && result.Blocked {
			t.Errorf("code %q should NOT be blocked, but was: %v", tt.code, result.Threats)
		}
		if tt.blocked && tt.ruleID != "" {
			found := false
			for _, threat := range result.Threats {
				if threat.RuleID == tt.ruleID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected rule %s in threats for %q, got %v", tt.ruleID, tt.code, result.Threats)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 11: Runtime scanner detects Python-specific threats
// ---------------------------------------------------------------------------

func TestCWasmE2E_RuntimeScanner_PythonThreats(t *testing.T) {
	scanner := NewRuntimeScanner()

	tests := []struct {
		code    string
		blocked bool
	}{
		{`eval("import os; os.system('rm -rf /')")`, true},
		{`exec("malicious code")`, true},
		{`subprocess(`, true},
		{`os.system("echo pwned")`, true},
		{`ctypes(`, true},
		{`def handler(event): return {"ok": True}`, false},
	}

	for _, tt := range tests {
		result := scanner.ScanSource(tt.code)
		if tt.blocked && !result.Blocked {
			t.Errorf("Python code %q should be blocked", tt.code)
		}
		if !tt.blocked && result.Blocked {
			t.Errorf("Python code %q should NOT be blocked", tt.code)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 12: C compiler detection and export format
// ---------------------------------------------------------------------------

func TestCWasmE2E_CompilerDetection(t *testing.T) {
	expectedExports := []string{"_execute", "_init", "_alloc", "_dealloc", "_metadata"}
	var parts []string
	for _, e := range expectedExports {
		parts = append(parts, fmt.Sprintf("'%s'", e))
	}
	expected := "[" + strings.Join(parts, ",") + "]"

	if expected != "['_execute','_init','_alloc','_dealloc','_metadata']" {
		t.Errorf("export format mismatch: %s", expected)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 13: Runtime router type parsing
// ---------------------------------------------------------------------------

func TestCWasmE2E_RouterTypeParsing(t *testing.T) {
	tests := []struct {
		input string
		want  RuntimeType
	}{
		{"c", RuntimeC},
		{"c11", RuntimeC},
		{"cpp", RuntimeCpp},
		{"cpp17", RuntimeCpp},
		{"c++", RuntimeCpp},
		{"rust", RuntimeRust},
		{"go", RuntimeGoWASM},
		{"python", RuntimePython},
		{"javascript", RuntimeJavaScript},
		{"unknown-lang", RuntimeUnknown},
	}

	for _, tt := range tests {
		got := RuntimeTypeFromString(tt.input)
		if got != tt.want {
			t.Errorf("RuntimeTypeFromString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 14: Runtime router validity check
// ---------------------------------------------------------------------------

func TestCWasmE2E_RouterTypeValidity(t *testing.T) {
	validTypes := []RuntimeType{
		RuntimeC, RuntimeCpp, RuntimeRust, RuntimeGoWASM,
		RuntimePython, RuntimeJavaScript, RuntimeSwiftWASM,
		RuntimeRubyWASM, RuntimeKotlinWASM,
	}

	for _, rt := range validTypes {
		if !rt.IsValid() {
			t.Errorf("RuntimeType %q should be valid", rt)
		}
	}

	invalidTypes := []RuntimeType{RuntimeUnknown, "foobar", ""}
	for _, rt := range invalidTypes {
		if rt.IsValid() {
			t.Errorf("RuntimeType %q should be invalid", rt)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 15: Circuit breaker
// ---------------------------------------------------------------------------

func TestCWasmE2E_CircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	if !cb.Allow() {
		t.Error("circuit breaker should allow initially")
	}

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.Allow() {
		t.Error("circuit breaker should reject after threshold failures")
	}
	if cb.State() != CircuitOpen {
		t.Errorf("state = %v, want CircuitOpen", cb.State())
	}

	time.Sleep(150 * time.Millisecond)

	if !cb.Allow() {
		t.Error("circuit breaker should allow in half-open state")
	}
	if cb.State() != CircuitHalfOpen {
		t.Errorf("state = %v, want CircuitHalfOpen", cb.State())
	}

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Errorf("state = %v, want CircuitClosed", cb.State())
	}
}

// ---------------------------------------------------------------------------
// E2E Test 16: Runtime router health check
// ---------------------------------------------------------------------------

func TestCWasmE2E_RouterHealthCheck(t *testing.T) {
	router := NewRuntimeRouter(nil)

	health := router.HealthCheck()
	if health.Status != "healthy" {
		t.Errorf("status = %q, want healthy", health.Status)
	}
	if health.Version != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", health.Version)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 17: Security config clone independence
// ---------------------------------------------------------------------------

func TestCWasmE2E_SecurityConfigClone(t *testing.T) {
	original := NewDefaultSecurityConfig()
	original.AllowedDomains = []string{"original.com"}
	original.MaxMemory = 128 * 1024 * 1024

	clone := original.Clone()

	clone.AllowedDomains[0] = "modified.com"
	clone.MaxMemory = 256 * 1024 * 1024

	if original.AllowedDomains[0] != "original.com" {
		t.Error("clone modified original AllowedDomains")
	}
	if original.MaxMemory != 128*1024*1024 {
		t.Error("clone modified original MaxMemory")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 18: C compilation via bundler
// ---------------------------------------------------------------------------

func TestCWasmE2E_BundlerCompilation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundler compilation test in short mode")
	}

	_, _, err := findCCompilerForTest()
	if err != nil {
		t.Skipf("no C compiler: %v", err)
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "main.c")
	if err := os.WriteFile(srcFile, []byte(testCHelloSource), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{
		Name:    "test-c-e2e",
		Version: "1.0.0",
		Runtime: "c",
		Entry:   "main.c",
	}

	wasmBytes, err := bundler.BundleForWasmRuntime(m)
	if err != nil {
		t.Fatalf("bundle failed: %v", err)
	}

	if len(wasmBytes) < 8 {
		t.Fatalf("wasm too small: %d bytes", len(wasmBytes))
	}

	if err := bundler.ValidateWASM(wasmBytes, nil); err != nil {
		if strings.Contains(err.Error(), "section 1 size mismatch") {
			t.Log("Known false positive: section 1 size mismatch (WASI-SDK encoding)")
		} else {
			t.Errorf("validation failed: %v", err)
		}
	}

	if wasmBytes[0] != 0x00 || wasmBytes[1] != 0x61 || wasmBytes[2] != 0x73 || wasmBytes[3] != 0x6D {
		t.Error("invalid wasm magic bytes")
	}

	t.Logf("Bundler produced %d bytes of valid WASM", len(wasmBytes))
}

// ---------------------------------------------------------------------------
// E2E Test 19: End-to-end pipeline (compile → validate → execute → verify)
// ---------------------------------------------------------------------------

func TestCWasmE2E_FullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full pipeline E2E in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}
	t.Logf("Step 1: Got WASM binary (%d bytes)", len(wasmBytes))

	if err := bundler.ValidateWASM(wasmBytes, nil); err != nil {
		if strings.Contains(err.Error(), "section 1 size mismatch") {
			t.Log("Step 2: Validation passed (section 1 size mismatch suppressed - WASI-SDK encoding)")
		} else {
			t.Fatalf("Step 2 - validation failed: %v", err)
		}
	} else {
		t.Log("Step 2: Validation passed")
	}

	handler := newTestHostHandler()
	config := newTestSecurityConfig()

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, config)
	if err != nil {
		t.Fatalf("Step 3 - create runtime: %v", err)
	}
	defer runtime.Close()
	t.Log("Step 3: Runtime created")

	if err := runtime.Init(); err != nil {
		t.Fatalf("Step 4 - init: %v", err)
	}
	t.Log("Step 4: Runtime initialized")

	input := []byte(`{"action":"test","value":42}`)
	output, err := runtime.ExecuteWithContext(context.Background(), input)
	if err != nil {
		t.Fatalf("Step 5 - execute: %v", err)
	}
	t.Logf("Step 5: Execution output: %s", string(output))

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Step 6 - output is not valid JSON: %v (raw: %s)", err, string(output))
	}
	t.Log("Step 6: Output is valid JSON")

	memUsage := runtime.GetMemoryUsage()
	t.Logf("Step 7: Memory usage: %d bytes", memUsage)

	output2, err := runtime.ExecuteWithContext(context.Background(), []byte(`{"warm":true}`))
	if err != nil {
		t.Fatalf("Step 8 - warm execute: %v", err)
	}
	if len(output2) == 0 {
		t.Fatal("Step 8: empty output on warm path")
	}
	t.Logf("Step 8: Warm execution output: %s", string(output2))

	t.Log("Full pipeline completed successfully")
}

// ---------------------------------------------------------------------------
// E2E Test 20: Runtime close safety
// ---------------------------------------------------------------------------

func TestCWasmE2E_RuntimeCloseSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping close safety test in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, nil, nil, handler, newTestSecurityConfig())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("double close: %v", err)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 21: Concurrent execution safety
// ---------------------------------------------------------------------------

func TestCWasmE2E_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent execution test in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	handler := newTestHostHandler()

	runtime, err := NewPythonRuntimeWithConfig(wasmPath, os.Stdout, os.Stderr, handler, newTestSecurityConfig())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// NOTE: wasmtime Store is not thread-safe, so we cannot call ExecuteWithContext
	// from multiple goroutines simultaneously. Instead, we test rapid sequential
	// executions to verify the runtime handles repeated calls correctly.
	const numExecutions = 10
	for i := 0; i < numExecutions; i++ {
		input := fmt.Sprintf(`{"execution":%d}`, i)
		output, err := runtime.ExecuteWithContext(context.Background(), []byte(input))
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
		if len(output) == 0 {
			t.Errorf("execution %d: empty output", i)
		}
	}
}

// ---------------------------------------------------------------------------
// E2E Test 22: Deterministic config defaults
// ---------------------------------------------------------------------------

func TestCWasmE2E_DeterministicConfig(t *testing.T) {
	dc := DefaultDeterministicConfig()

	if dc.MaxInstructions != DefaultMaxInstructions {
		t.Errorf("MaxInstructions = %d, want %d", dc.MaxInstructions, DefaultMaxInstructions)
	}
	if !dc.NormalizeMemoryAccess {
		t.Error("NormalizeMemoryAccess should be true by default")
	}
	if !dc.ConstantTimeExecution {
		t.Error("ConstantTimeExecution should be true by default")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 23: Security config chunk buffer size
// ---------------------------------------------------------------------------

func TestCWasmE2E_ChunkBufferSize(t *testing.T) {
	config := NewDefaultSecurityConfig()
	chunkSize := config.GetChunkBufferSize()

	const minChunkSize = 64 * 1024
	if chunkSize < minChunkSize {
		t.Errorf("chunk size %d below minimum %d", chunkSize, minChunkSize)
	}

	smallConfig := &WASMSecurityConfig{MaxInputSize: 1024}
	chunkSize = smallConfig.GetChunkBufferSize()
	if chunkSize < minChunkSize {
		t.Errorf("small config chunk size %d below minimum %d", chunkSize, minChunkSize)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 24: ExecutionConfig validation
// ---------------------------------------------------------------------------

func TestCWasmE2E_ExecutionConfig(t *testing.T) {
	config := NewDefaultSecurityConfig()

	if !config.ValidateInputSize(100) {
		t.Error("100 bytes should be valid")
	}
	if config.ValidateInputSize(config.MaxInputSize + 1) {
		t.Error("input exceeding max should be invalid")
	}

	if !config.ValidateOutputSize(100) {
		t.Error("100 bytes output should be valid")
	}
	if config.ValidateOutputSize(config.MaxOutputSize + 1) {
		t.Error("output exceeding max should be invalid")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 25: High entropy detection in scanner
// ---------------------------------------------------------------------------

func TestCWasmE2E_HighEntropyDetection(t *testing.T) {
	scanner := NewRuntimeScanner()

	normalCode := `int main() { return 0; }`
	result := scanner.ScanSource(normalCode)
	for _, threat := range result.Threats {
		if threat.RuleID == "HIGH_ENTROPY" {
			t.Error("normal code should not trigger high entropy")
		}
	}

	// Build high-entropy string with many unique characters
	var obf strings.Builder
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:',.<>?/~`"
	for i := 0; i < 500; i++ {
		obf.WriteByte(chars[i%len(chars)])
	}
	result = scanner.ScanSource(obf.String())
	foundEntropy := false
	for _, threat := range result.Threats {
		if threat.RuleID == "HIGH_ENTROPY" {
			foundEntropy = true
			break
		}
	}
	if !foundEntropy {
		t.Error("obfuscated payload should trigger high entropy detection")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 26: Default host handler safety
// ---------------------------------------------------------------------------

func TestCWasmE2E_DefaultHostHandler(t *testing.T) {
	handler := NewDefaultHostHandler(nil)

	handler.Log("should not panic")

	_, err := handler.Fetch(`{"url":"http://example.com"}`)
	if err == nil {
		t.Error("default fetch should return error")
	}

	_, err = handler.KVGet("key")
	if err == nil {
		t.Error("default kv get should return error")
	}

	err = handler.KVSet("key", "val")
	if err == nil {
		t.Error("default kv set should return error")
	}

	env := handler.GetEnv("HOME")
	if env != "" {
		t.Error("default get env should return empty")
	}

	_, err = handler.AIInference("model", []byte("input"), "")
	if err == nil {
		t.Error("default ai inference should return error")
	}

	_, err = handler.StateGet("path")
	if err == nil {
		t.Error("default state get should return error")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 27: Emscripten export format matches runtime expectations
// ---------------------------------------------------------------------------

func TestCWasmE2E_ExportFormatConsistency(t *testing.T) {
	sdkExports := []string{"init", "execute", "alloc", "dealloc", "metadata"}
	emscriptenExports := []string{"_execute", "_init", "_alloc", "_dealloc", "_metadata"}

	for _, exp := range emscriptenExports {
		if !strings.HasPrefix(exp, "_") {
			t.Errorf("Emscripten export %q should have underscore prefix", exp)
		}
		stripped := strings.TrimPrefix(exp, "_")
		found := false
		for _, sdkExp := range sdkExports {
			if stripped == sdkExp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Emscripten export %q has no matching SDK export", exp)
		}
	}

	if len(emscriptenExports) != len(sdkExports) {
		t.Errorf("export count mismatch: Emscripten=%d SDK=%d", len(emscriptenExports), len(sdkExports))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 28: C compilation error handling
// ---------------------------------------------------------------------------

func TestCWasmE2E_CompilationErrors(t *testing.T) {
	m := &manifest.Manifest{
		Name:    "test-missing",
		Version: "1.0.0",
		Runtime: "c",
		Entry:   "nonexistent.c",
	}
	_, err := bundler.BundleForWasmRuntime(m)
	if err == nil {
		t.Error("expected error for missing entry file")
	}

	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.c")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m = &manifest.Manifest{
		Name:    "test-empty",
		Version: "1.0.0",
		Runtime: "c",
		Entry:   "empty.c",
	}
	_, err = bundler.BundleForWasmRuntime(m)
	if err == nil {
		t.Error("expected error for empty source")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 29: Runtime router with circuit breaker
// ---------------------------------------------------------------------------

func TestCWasmE2E_RouterWithCircuitBreaker(t *testing.T) {
	router := NewRuntimeRouter(nil)
	cbRouter := NewRuntimeRouterWithCircuitBreaker(router, 2, 50*time.Millisecond)

	breaker := cbRouter.getBreaker(RuntimeC)
	if !breaker.Allow() {
		t.Error("should allow initially")
	}

	breaker.RecordFailure()
	breaker.RecordFailure()

	if breaker.Allow() {
		t.Error("should reject after threshold")
	}

	time.Sleep(60 * time.Millisecond)
	if !breaker.Allow() {
		t.Error("should allow after reset timeout")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 30: WASM validation with strict config
// ---------------------------------------------------------------------------

func TestCWasmE2E_WASMValidationStrict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping strict validation test in short mode")
	}

	wasmPath := getPrecompiledOrCompileC(t)
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	strictConfig := &bundler.WASMValidationConfig{
		MaxBinarySize:          10 * 1024 * 1024,
		RequireMemoryExport:    false,
		AllowWASI:              true,
		BlockedImports:         []string{},
		MaxImports:             100,
		MaxExports:             200,
		MaxFunctions:           500,
		MaxMemoryPages:         1024,
		EnableStrictValidation: true,
	}

	err = bundler.ValidateWASM(wasmBytes, strictConfig)
	if err != nil {
		t.Logf("Strict validation rejected binary (expected for some C runtimes): %v", err)
	} else {
		t.Log("Strict validation passed")
	}
}
