package frg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ExecuteAsync creates a persistent instance for async/streaming execution
func (e *ExecutionEngine) ExecuteAsync(ctx context.Context, def *GraphDefinition, input map[string]interface{}) (*GraphInstance, error) {
	logrus.WithFields(logrus.Fields{
		"graph": fmt.Sprintf("%s/%s@%s", def.Author, def.Name, def.Version),
		"mode":  def.ExecutionMode,
	}).Info("Starting asynchronous graph execution")

	// Create instance
	inputJSON, _ := json.Marshal(input)
	instance, err := e.repo.CreateInstance(ctx, def, inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Build runtime with channels
	runtime, err := e.buildRuntime(ctx, def, instance)
	if err != nil {
		e.repo.UpdateInstanceStatus(ctx, instance.ID, InstanceStatusFailed)
		return nil, err
	}

	runtime.InputChannel = make(chan *GraphEvent, 100)
	runtime.OutputChannel = make(chan *ExecutionResult, 10)

	// Store runtime
	e.instancesMu.Lock()
	e.instances[instance.ID] = runtime
	e.instancesMu.Unlock()

	// Start background execution
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.runInstance(runtime)
	}()

	// Initial trigger event
	go func() {
		initialEvent := &GraphEvent{
			InstanceID:  instance.ID,
			EventType:   "trigger",
			Payload:     inputJSON,
			ContentType: "json",
			SequenceNum: 1,
		}
		runtime.InputChannel <- initialEvent
	}()

	return instance, nil
}

// runInstance runs an async/streaming instance
func (e *ExecutionEngine) runInstance(runtime *GraphRuntime) {
	ctx := context.Background()
	defer func() {
		e.instancesMu.Lock()
		delete(e.instances, runtime.Instance.ID)
		e.instancesMu.Unlock()
	}()

	// Mark as running
	e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusRunning)
	runtime.Instance.Status = InstanceStatusRunning

	if runtime.Definition.ExecutionMode == ExecutionModeStreaming {
		e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusStreaming)
		runtime.Instance.Status = InstanceStatusStreaming
	}

	// Event-driven execution loop
	sequenceNum := int64(1)
	for event := range runtime.InputChannel {
		// Process event
		sequenceNum++
		event.SequenceNum = sequenceNum

		switch event.EventType {
		case "trigger":
			e.handleTriggerEvent(ctx, runtime, event)
		case "stream":
			e.handleStreamEvent(ctx, runtime, event)
		case "complete":
			e.handleNodeCompleteEvent(ctx, runtime, event)
		case "error":
			e.handleNodeErrorEvent(ctx, runtime, event)
		case "checkpoint":
			e.handleCheckpointEvent(ctx, runtime, event)
		}

		// Append to event log
		e.repo.AppendEvent(ctx, event)
	}

	// Instance stopped
	if runtime.Instance.Status != InstanceStatusFailed {
		e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusCompleted)
	}
}
