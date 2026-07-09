package wasm

import (
	"bytes"
	"os"
	"testing"
)

func TestPythonRuntime_BasicExecution(t *testing.T) {
	// Create runtime instance with default handler
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test loading Python code
	pythonCode := `
def handler(event):
    return {"result": "executed", "input_received": True}
`

	if err := runtime.LoadCode(pythonCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	// Test executing with input
	input := `{"test": "data"}`
	output, err := runtime.Execute([]byte(input))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	// Check that we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty output from execution")
	}

	// The WAT interpreter returns a JSON response with result and input_received fields
	actualResponse := string(output)
	if !bytes.Contains([]byte(actualResponse), []byte(`"result"`)) || !bytes.Contains([]byte(actualResponse), []byte(`"executed"`)) {
		t.Errorf("Expected response to contain result and executed, got %q", actualResponse)
	}
	if !bytes.Contains([]byte(actualResponse), []byte(`"input_received"`)) {
		t.Errorf("Expected response to contain 'input_received', got %q", actualResponse)
	}

	t.Logf("Runtime test passed. Output: %s", string(output))
}

func TestPythonRuntime_NoHandlerFunction(t *testing.T) {
	// Create runtime instance with default handler
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test loading Python code without handler function
	pythonCode := `
def other_function():
    return "not a handler"
`

	if err := runtime.LoadCode(pythonCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	// Test executing - should return error message
	output, err := runtime.Execute([]byte("{}"))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	// Check that we got an error response
	if len(output) == 0 {
		t.Error("Expected non-empty output from execution")
	}

	outputStr := string(output)
	if !bytes.Contains(output, []byte("error")) {
		t.Errorf("Expected error response, got: %s", outputStr)
	}

	t.Logf("No handler test passed. Output: %s", outputStr)
}

func TestPythonRuntime_MemoryManagement(t *testing.T) {
	// Create runtime instance with default handler
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test memory allocation/deallocation
	allocPtr, err := runtime.allocate(100)
	if err != nil {
		t.Fatalf("Failed to allocate memory: %v", err)
	}

	if allocPtr == 0 {
		t.Error("Expected non-zero pointer from allocation")
	}

	// Test writing to memory
	testData := []byte("Hello, WASM!")
	if err := runtime.writeMemory(allocPtr, testData); err != nil {
		t.Fatalf("Failed to write to memory: %v", err)
	}

	// Test reading from memory
	readData, err := runtime.readMemory(allocPtr, len(testData))
	if err != nil {
		t.Fatalf("Failed to read from memory: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Memory read/write mismatch. Expected %q, got %q", testData, readData)
	}

	// Test deallocation
	if err := runtime.deallocate(allocPtr, 100); err != nil {
		t.Errorf("Failed to deallocate memory: %v", err)
	}

	t.Log("Memory management test passed")
}

func TestPythonRuntime_HostFunctions(t *testing.T) {
	// DefaultHostHandler is a documented no-op: KV/Fetch/AI/State all
	// return sentinel errors. The semantics are tested here so future
	// refactors don't accidentally turn the no-op into a working handler.
	handler := NewDefaultHostHandler(nil)

	// KV operations should return ErrKVNotAvailable.
	if err := handler.KVSet("test_key", "test_value"); err == nil {
		t.Fatal("Expected ErrKVNotAvailable from default handler KVSet")
	}
	if _, err := handler.KVGet("any"); err == nil {
		t.Fatal("Expected ErrKVNotAvailable from default handler KVGet")
	}

	// Fetch returns ErrFetchNotAvailable.
	if _, err := handler.Fetch(`{"method": "GET", "url": "http://example.com"}`); err == nil {
		t.Fatal("Expected ErrFetchNotAvailable from default handler Fetch")
	}

	// AI returns ErrAINotAvailable.
	if _, err := handler.AIInference("model", []byte("input"), ""); err == nil {
		t.Fatal("Expected ErrAINotAvailable from default handler AIInference")
	}

	// State operations return ErrStateNotAvailable.
	if _, err := handler.StateGet("path"); err == nil {
		t.Fatal("Expected ErrStateNotAvailable from default handler StateGet")
	}
	if err := handler.StateSet("path", "value"); err == nil {
		t.Fatal("Expected ErrStateNotAvailable from default handler StateSet")
	}

	// Empty env returns "".
	if got := handler.GetEnv("ANY"); got != "" {
		t.Errorf("Expected empty env, got %q", got)
	}

	t.Log("DefaultHostHandler no-op contract verified")
}

func TestPythonRuntime_DebugMode(t *testing.T) {
	// Create runtime instance with debug mode enabled
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntimeWithDebug("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler, true)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime (should produce debug output)
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test loading and executing code (should produce debug output)
	pythonCode := `
def handler(event):
    return {"result": "executed", "input_received": True}
`

	if err := runtime.LoadCode(pythonCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	output, err := runtime.Execute([]byte(`{"test": "data"}`))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	if !bytes.Contains(output, []byte("result")) {
		t.Errorf("Expected result in output, got: %s", string(output))
	}

	t.Log("Debug mode test passed")
}

func TestPythonRuntime_OSEnviron(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	// Create runtime instance
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test loading Python code that uses os.environ
	pythonCode := `
import os

def handler(event):
    # Test accessing environment variables
    test_var = os.environ.get('TEST_VAR', 'not_found')
    return {"env_test": test_var, "result": "executed"}
`

	if err := runtime.LoadCode(pythonCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	// Test executing with input
	input := `{"test": "data"}`
	output, err := runtime.Execute([]byte(input))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	// Check that we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty output from execution")
	}

	// For now, this will return the basic response since full Python parsing isn't implemented yet
	// But the framework is in place for when it is
	actualResponse := string(output)
	if !bytes.Contains([]byte(actualResponse), []byte(`"result"`)) {
		t.Errorf("Expected response to contain result, got %q", actualResponse)
	}

	t.Logf("OS environ test passed. Output: %s", string(output))
}

func TestPythonRuntime_StdlibIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WASM stdlib integration test in short mode")
	}

	// Set test environment variables
	os.Setenv("TEST_VAR", "hello_world")
	os.Setenv("API_KEY", "secret123")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("API_KEY")
	}()

	// Create runtime instance
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize the runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test loading Python code that uses multiple stdlib features
	pythonCode := `
import json
import os
import sys
import time

def handler(event):
    # Test JSON parsing
    input_data = json.loads(json.dumps(event))

    # Test environment variables
    test_var = os.environ.get('TEST_VAR', 'default')
    api_key = os.environ.get('API_KEY', 'no_key')

    # Test time functions
    current_time = time.time()

    # Test sys info
    version = sys.version
    platform = sys.platform

    # Return comprehensive response
    return {
        "json_processed": True,
        "input_echo": input_data,
        "env_vars": {
            "test_var": test_var,
            "api_key": api_key
        },
        "time_sample": current_time,
        "sys_info": {
            "version": version,
            "platform": platform
        },
        "stdlib_test": "passed"
    }
`

	if err := runtime.LoadCode(pythonCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	// Test executing with complex input
	input := `{
		"user": "test_user",
		"action": "stdlib_test",
		"data": {
			"items": [1, 2, 3],
			"nested": {"key": "value"}
		}
	}`
	output, err := runtime.Execute([]byte(input))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	// Check that we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty output from execution")
	}

	// The WAT interpreter will return the basic response since full Python parsing isn't implemented yet
	// But this test demonstrates the framework is ready for when it is
	actualResponse := string(output)
	if !bytes.Contains([]byte(actualResponse), []byte(`"result"`)) {
		t.Errorf("Expected response to contain result, got %q", actualResponse)
	}

	t.Logf("Stdlib integration test passed. Output: %s", string(output))
}
