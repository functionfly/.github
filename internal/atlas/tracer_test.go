package atlas

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestTracerEndToEnd(t *testing.T) {
	baseURL := os.Getenv("ATLAS_URL")
	apiKey := os.Getenv("ATLAS_API_KEY")
	if baseURL == "" {
		baseURL = "http://localhost:7447"
	}
	if apiKey == "" {
		t.Skip("ATLAS_API_KEY not set, skipping integration test")
	}

	os.Setenv("ATLAS_URL", baseURL)
	os.Setenv("ATLAS_API_KEY", apiKey)
	defer os.Unsetenv("ATLAS_URL")
	defer os.Unsetenv("ATLAS_API_KEY")

	tracer := NewTracer(nil)
	if !tracer.Enabled() {
		t.Fatal("tracer should be enabled")
	}

	ctx := context.Background()
	input, _ := json.Marshal(map[string]string{"name": "world"})

	// Start trace
	runID, err := tracer.StartExecutionTrace(ctx, &ExecutionTrace{
		FunctionID:   "test-fn-id",
		FunctionName: "hello-world",
		Author:       "testuser",
		Version:      "1.0.0",
		Runtime:      "node20",
		Tier:         "pro",
		StartTime:    time.Now(),
		InputPayload: input,
	})
	if err != nil {
		t.Fatalf("StartExecutionTrace: %v", err)
	}
	t.Logf("RunID: %s", runID)

	// Record a decision
	tracer.RecordDecision(ctx, runID, "select_engine", map[string]interface{}{
		"engine": "nodeEngine",
	})

	// Record an action
	tracer.RecordAction(ctx, runID, "execute", "node20", map[string]interface{}{
		"timeout_ms": 30000,
	})

	// Finish trace
	output, _ := json.Marshal(map[string]string{"greeting": "hello world"})
	tracer.FinishExecutionTrace(ctx, runID, &ExecutionResult{
		Output:     output,
		DurationMs: 42,
		StatusCode: 200,
		ResourceUsage: &ResourceUsageResult{
			MaxMemoryMB:    128,
			MemoryUsedMB:   64,
			CPUTimeUsedMs:  30,
			WallTimeUsedMs: 42,
		},
	})

	// Give flush goroutine time to complete
	time.Sleep(2 * time.Second)

	// Verify events were recorded
	client := NewClient(baseURL, apiKey)
	events, err := client.GetEvents(ctx, runID, 0, 100)
	if err != nil {
		t.Fatalf("GetEvents: %v", err)
	}
	t.Logf("Got %d events", len(events))

	if len(events) != 4 {
		t.Errorf("expected 4 events (input, decision, action, result), got %d", len(events))
	}

	for _, e := range events {
		payload := string(e.Payload)
		if len(payload) > 80 {
			payload = payload[:80]
		}
		t.Logf("  [%d] %s: %s", e.Sequence, e.Kind, payload)
	}
}
