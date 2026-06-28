package atlas

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Tracer struct {
	client  *Client
	enabled bool
	logger  *logrus.Logger

	mu      sync.RWMutex
	runs    map[string]*traceRun
}

type traceRun struct {
	runID    string
	seq      uint64
	mu       sync.Mutex
	events   []Event
	started  time.Time
	systemID string
}

type ExecutionTrace struct {
	RunID          string
	FunctionID     string
	FunctionName   string
	Author         string
	Version        string
	Runtime        string
	Tier           string
	StartTime      time.Time
	InputPayload   json.RawMessage
}

type ExecutionResult struct {
	Output      json.RawMessage
	DurationMs  int
	Cached      bool
	ColdStart   bool
	StatusCode  int
	Error       string
	ResourceUsage *ResourceUsageResult
}

type ResourceUsageResult struct {
	MaxMemoryMB    int
	MemoryUsedMB   int
	CPUTimeUsedMs  int
	WallTimeUsedMs int
}

func NewTracer(client *Client) *Tracer {
	enabled := os.Getenv("ATLAS_ENABLED") == "true" || os.Getenv("ATLAS_URL") != ""

	var c *Client
	if client != nil {
		c = client
		enabled = true
	} else if url := os.Getenv("ATLAS_URL"); url != "" {
		apiKey := os.Getenv("ATLAS_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ATLAS_TOKEN")
		}
		c = NewClient(url, apiKey)
		enabled = true
	}

	return &Tracer{
		client:  c,
		enabled: enabled,
		logger:  logrus.StandardLogger(),
		runs:    make(map[string]*traceRun),
	}
}

func (t *Tracer) Enabled() bool {
	return t.enabled && t.client != nil
}

func (t *Tracer) Client() *Client {
	return t.client
}

func (t *Tracer) SetLogger(logger *logrus.Logger) {
	t.logger = logger
	if t.client != nil {
		t.client.SetLogger(logger)
	}
}

func (t *Tracer) StartExecutionTrace(ctx context.Context, trace *ExecutionTrace) (string, error) {
	if !t.Enabled() {
		return "", nil
	}

	labels := map[string]interface{}{
		"type":          "function_execution",
		"function_id":   trace.FunctionID,
		"function_name": trace.FunctionName,
		"author":        trace.Author,
		"version":       trace.Version,
		"runtime":       trace.Runtime,
		"tier":          trace.Tier,
	}

	runID, err := t.client.CreateRun(ctx, labels)
	if err != nil {
		t.logger.WithError(err).Warn("atlas: failed to create run")
		return "", err
	}

	tr := &traceRun{
		runID:    runID,
		started:  trace.StartTime,
		systemID: fmt.Sprintf("functionfly/%s/%s", trace.Author, trace.FunctionName),
	}

	t.mu.Lock()
	t.runs[runID] = tr
	t.mu.Unlock()

	inputEvent := Event{
		TimestampNs: uint64(trace.StartTime.UnixNano()),
		Sequence:    0,
		SystemID:    tr.systemID,
		Kind:        EventKindInput,
		Payload:     buildInputPayload(trace),
	}

	tr.events = append(tr.events, inputEvent)
	tr.seq = 1

	go t.flushEvents(runID, []Event{inputEvent})

	return runID, nil
}

func (t *Tracer) RecordDecision(ctx context.Context, runID string, decision string, metadata map[string]interface{}) {
	if !t.Enabled() || runID == "" {
		return
	}

	t.mu.RLock()
	tr, ok := t.runs[runID]
	t.mu.RUnlock()
	if !ok {
		return
	}

	tr.mu.Lock()
	payload, _ := json.Marshal(map[string]interface{}{
		"decision": decision,
		"metadata": metadata,
	})
	event := Event{
		TimestampNs: uint64(time.Now().UnixNano()),
		Sequence:    tr.seq,
		SystemID:    tr.systemID,
		Kind:        EventKindDecision,
		Payload:     payload,
	}
	tr.events = append(tr.events, event)
	tr.seq++
	tr.mu.Unlock()

	go t.flushEvents(runID, []Event{event})
}

func (t *Tracer) RecordAction(ctx context.Context, runID string, action string, target string, metadata map[string]interface{}) {
	if !t.Enabled() || runID == "" {
		return
	}

	t.mu.RLock()
	tr, ok := t.runs[runID]
	t.mu.RUnlock()
	if !ok {
		return
	}

	tr.mu.Lock()
	payload, _ := json.Marshal(map[string]interface{}{
		"action":   action,
		"target":   target,
		"metadata": metadata,
	})
	event := Event{
		TimestampNs: uint64(time.Now().UnixNano()),
		Sequence:    tr.seq,
		SystemID:    tr.systemID,
		Kind:        EventKindAction,
		Payload:     payload,
	}
	tr.events = append(tr.events, event)
	tr.seq++
	tr.mu.Unlock()

	go t.flushEvents(runID, []Event{event})
}

func (t *Tracer) FinishExecutionTrace(ctx context.Context, runID string, result *ExecutionResult) {
	if !t.Enabled() || runID == "" {
		return
	}

	t.mu.RLock()
	tr, ok := t.runs[runID]
	t.mu.RUnlock()
	if !ok {
		return
	}

	var kind EventKind
	var payload map[string]interface{}

	if result.Error != "" {
		kind = EventKindError
		payload = map[string]interface{}{
			"error":       result.Error,
			"status_code": result.StatusCode,
			"duration_ms": result.DurationMs,
		}
	} else {
		kind = EventKindResult
		payload = map[string]interface{}{
			"duration_ms": result.DurationMs,
			"cached":      result.Cached,
			"cold_start":  result.ColdStart,
			"status_code": result.StatusCode,
		}
		if result.Output != nil && len(result.Output) > 0 {
			payload["output_preview"] = truncateJSON(result.Output, 1024)
		}
	}

	if result.ResourceUsage != nil {
		payload["resource_usage"] = map[string]interface{}{
			"max_memory_mb":     result.ResourceUsage.MaxMemoryMB,
			"memory_used_mb":    result.ResourceUsage.MemoryUsedMB,
			"cpu_time_used_ms":  result.ResourceUsage.CPUTimeUsedMs,
			"wall_time_used_ms": result.ResourceUsage.WallTimeUsedMs,
		}
	}

	tr.mu.Lock()
	payloadBytes, _ := json.Marshal(payload)
	event := Event{
		TimestampNs: uint64(time.Now().UnixNano()),
		Sequence:    tr.seq,
		SystemID:    tr.systemID,
		Kind:        kind,
		Payload:     payloadBytes,
	}
	tr.events = append(tr.events, event)
	tr.seq++
	allEvents := make([]Event, len(tr.events))
	copy(allEvents, tr.events)
	tr.mu.Unlock()

	go func() {
		t.flushEvents(runID, []Event{event})
		t.mu.Lock()
		delete(t.runs, runID)
		t.mu.Unlock()
	}()
}

func (t *Tracer) flushEvents(runID string, events []Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	accepted, err := t.client.AppendEvents(ctx, runID, events)
	if err != nil {
		t.logger.WithError(err).WithField("run_id", runID).Warn("atlas: failed to flush events")
		return
	}

	t.logger.WithFields(logrus.Fields{
		"run_id":   runID,
		"accepted": accepted,
	}).Debug("atlas: flushed events")
}

func buildInputPayload(trace *ExecutionTrace) json.RawMessage {
	payload := map[string]interface{}{
		"function_id":   trace.FunctionID,
		"function_name": trace.FunctionName,
		"author":        trace.Author,
		"version":       trace.Version,
		"runtime":       trace.Runtime,
		"tier":          trace.Tier,
	}
	if trace.InputPayload != nil && len(trace.InputPayload) > 0 {
		payload["input_preview"] = truncateJSON(trace.InputPayload, 512)
	}
	b, _ := json.Marshal(payload)
	return b
}

func truncateJSON(data json.RawMessage, maxLen int) string {
	s := string(data)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
