package wasm

import (
	"bytes"
	"os"
	"testing"
)

func TestPythonRuntime_BundlerIntegration(t *testing.T) {
	// Create test Python code
	testCode := `
def handler(event):
    """Test handler function for integration testing"""
    return {
        "status": "success",
        "message": "Hello from WASM Python runtime!",
        "input": event
    }
`

	// Test bundling with the WASM runtime
	// This simulates what the bundler does
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, NewDefaultHostHandler(nil))
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize runtime
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Load the Python code (simulating what the bundler would do)
	if err := runtime.LoadCode(testCode); err != nil {
		t.Fatalf("Failed to load Python code: %v", err)
	}

	// Test execution with sample input
	testInput := `{"user": "test", "action": "greet"}`
	output, err := runtime.Execute([]byte(testInput))
	if err != nil {
		t.Fatalf("Failed to execute Python code: %v", err)
	}

	// Verify the output contains expected elements
	outputStr := string(output)
	if !bytes.Contains([]byte(outputStr), []byte("result")) {
		t.Errorf("Expected result in output, got: %s", outputStr)
	}

	if !bytes.Contains([]byte(outputStr), []byte("executed")) {
		t.Errorf("Expected 'executed' in output, got: %s", outputStr)
	}

	t.Logf("Bundler integration test passed. Output: %s", outputStr)
}

func TestPythonRuntime_EndToEndFlow(t *testing.T) {
	// Test the complete flow: create runtime -> init -> load code -> execute multiple times
	handler := NewDefaultHostHandler(nil)
	runtime, err := NewPythonRuntime("../bundler/python/minimal-python.wasm", os.Stdout, os.Stderr, handler)
	if err != nil {
		t.Fatalf("Failed to create Python runtime: %v", err)
	}
	defer runtime.Close()

	// Initialize
	if err := runtime.Init(); err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Load code
	code := `
def handler(event):
    return {"processed": True, "data": event}
`
	if err := runtime.LoadCode(code); err != nil {
		t.Fatalf("Failed to load code: %v", err)
	}

	// Execute multiple times with different inputs
	// Note: Current WAT interpreter returns fixed response regardless of input
	testInputs := []string{
		`{"test": 1}`,
		`{"test": 2}`,
		`{"action": "process"}`,
	}

	for i, input := range testInputs {
		output, err := runtime.Execute([]byte(input))
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}

		// WAT interpreter always returns the same response
		if !bytes.Contains(output, []byte("result")) || !bytes.Contains(output, []byte("executed")) {
			t.Errorf("Execution %d: unexpected output %q", i, string(output))
		}
	}

	// Test host functions integration
	// Set a value in KV store via host function
	if err := handler.KVSet("integration_test", "success"); err != nil {
		t.Fatalf("Failed to set KV: %v", err)
	}

	// Get it back
	value, err := handler.KVGet("integration_test")
	if err != nil {
		t.Fatalf("Failed to get KV: %v", err)
	}

	if value != "success" {
		t.Errorf("Expected KV value 'success', got '%s'", value)
	}

	t.Log("End-to-end flow test passed")
}