package atlas

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestClientEndToEnd(t *testing.T) {
	baseURL := os.Getenv("ATLAS_URL")
	apiKey := os.Getenv("ATLAS_API_KEY")
	if baseURL == "" {
		baseURL = "http://localhost:7447"
	}
	if apiKey == "" {
		t.Skip("ATLAS_API_KEY not set, skipping integration test")
	}

	client := NewClient(baseURL, apiKey)
	ctx := context.Background()

	// Health check
	status, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	t.Logf("Health: %s", status)

	// Create run
	labels := map[string]interface{}{
		"type":          "function_execution",
		"function_name": "test-func",
		"author":        "testuser",
		"runtime":       "node20",
	}
	runID, err := client.CreateRun(ctx, labels)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	t.Logf("RunID: %s", runID)

	// Append events
	inputPayload, _ := json.Marshal(map[string]interface{}{
		"function_name": "test-func",
		"input_preview": `{"name": "world"}`,
	})
	actionPayload, _ := json.Marshal(map[string]interface{}{
		"action": "execute",
		"target": "node20",
	})
	resultPayload, _ := json.Marshal(map[string]interface{}{
		"duration_ms": 42,
		"cached":      false,
		"status_code": 200,
	})

	events := []Event{
		{TimestampNs: 1000000000, Sequence: 0, SystemID: "functionfly/testuser/test-func", Kind: EventKindInput, Payload: inputPayload},
		{TimestampNs: 2000000000, Sequence: 1, SystemID: "functionfly/testuser/test-func", Kind: EventKindAction, Payload: actionPayload},
		{TimestampNs: 3000000000, Sequence: 2, SystemID: "functionfly/testuser/test-func", Kind: EventKindResult, Payload: resultPayload},
	}

	accepted, err := client.AppendEvents(ctx, runID, events)
	if err != nil {
		t.Fatalf("AppendEvents: %v", err)
	}
	t.Logf("Accepted: %d", accepted)

	if accepted != 3 {
		t.Errorf("expected 3 accepted, got %d", accepted)
	}

	// Get events back
	gotEvents, err := client.GetEvents(ctx, runID, 0, 100)
	if err != nil {
		t.Fatalf("GetEvents: %v", err)
	}
	t.Logf("Got %d events", len(gotEvents))

	if len(gotEvents) != 3 {
		t.Errorf("expected 3 events, got %d", len(gotEvents))
	}

	// Get run
	run, err := client.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run == nil {
		t.Fatal("GetRun returned nil")
	}
	t.Logf("Run: event_count=%d", run.EventCount)

	// Get graph
	graph, err := client.GetGraph(ctx, runID)
	if err != nil {
		t.Fatalf("GetGraph: %v", err)
	}
	t.Logf("Graph: %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))

	// List runs
	runs, err := client.ListRuns(ctx, 10, "")
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	t.Logf("Listed %d runs", len(runs))

	// Search
	kind := EventKindResult
	searchResult, err := client.Search(ctx, SearchRequest{Kind: &kind, Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	t.Logf("Search found %d events (scanned %d runs)", len(searchResult.Events), searchResult.ScannedRuns)
}
