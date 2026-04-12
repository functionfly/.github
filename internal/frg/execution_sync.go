package frg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecuteSync executes a graph synchronously (blocking)
func (e *ExecutionEngine) ExecuteSync(ctx context.Context, def *GraphDefinition, input map[string]interface{}) (*ExecutionResult, error) {
	logrus.WithFields(logrus.Fields{
		"graph": fmt.Sprintf("%s/%s@%s", def.Author, def.Name, def.Version),
	}).Info("Starting synchronous graph execution")

	// Create instance
	inputJSON, _ := json.Marshal(input)
	instance, err := e.repo.CreateInstance(ctx, def, inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Build runtime
	runtime, err := e.buildRuntime(ctx, def, instance)
	if err != nil {
		e.repo.UpdateInstanceStatus(ctx, instance.ID, InstanceStatusFailed)
		return nil, err
	}

	// Execute in topological order
	return e.executeDag(ctx, runtime, input)
}

// executeDag executes a DAG in topological order (synchronous)
func (e *ExecutionEngine) executeDag(ctx context.Context, runtime *GraphRuntime, input map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	// Get topological order (reuse existing graph service algorithm)
	order, err := e.computeExecutionOrder(runtime)
	if err != nil {
		e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusFailed)
		return nil, err
	}

	// Track execution state
	nodeResults := make(map[string]*NodeExecutionResult)
	currentData := input

	// Execute nodes in order
	for _, nodeID := range order {
		node := runtime.Nodes[nodeID]

		// Collect inputs from predecessor edges
		nodeInput := e.collectNodeInputs(node, runtime.Edges, nodeResults)

		// Execute node
		result, err := e.executeNode(ctx, runtime.Instance, node, nodeInput)
		if err != nil {
			// Check for fallback
			fallbackExecuted := false
			for _, edge := range runtime.Edges {
				if edge.Definition.SourceNodeID == nodeID && edge.Definition.FallbackNodeID != nil {
					fallbackNode := runtime.Nodes[*edge.Definition.FallbackNodeID]
					fallbackResult, err := e.executeNode(ctx, runtime.Instance, fallbackNode, nodeInput)
					if err == nil {
						result = fallbackResult
						fallbackExecuted = true
						break
					}
				}
			}

			if !fallbackExecuted {
				e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusFailed)
				return e.buildErrorResult(runtime.Instance, nodeResults, err, time.Since(startTime).Milliseconds()), nil
			}
		}

		nodeResults[nodeID] = result
		currentData = make(map[string]interface{})
		if result.Output != nil {
			json.Unmarshal(result.Output, &currentData)
		}

		// Update node state in database
		state := &NodeState{
			Status:       result.Status,
			Output:       result.Output,
			Error:        result.Error,
			AttemptCount: 1,
		}
		if result.Status == "completed" {
			state.ExecCertID = result.CertID
		}
		e.repo.UpdateInstanceNodeState(ctx, runtime.Instance.ID, nodeID, state)
	}

	duration := time.Since(startTime).Milliseconds()

	// Mark completed
	e.repo.UpdateInstanceStatus(ctx, runtime.Instance.ID, InstanceStatusCompleted)

	// Build final result
	return e.buildSuccessResult(runtime.Instance, nodeResults, duration), nil
}

func (e *ExecutionEngine) buildSuccessResult(instance *GraphInstance, nodeResults map[string]*NodeExecutionResult, durationMs int64) *ExecutionResult {
	// Get final output (from last node)
	var finalOutput json.RawMessage
	for _, result := range nodeResults {
		if result.Output != nil {
			finalOutput = result.Output
		}
	}

	return &ExecutionResult{
		InstanceID:  instance.ID,
		Status:      InstanceStatusCompleted,
		Output:      finalOutput,
		NodeResults: nodeResults,
		DurationMs:  int(durationMs),
	}
}

func (e *ExecutionEngine) buildErrorResult(instance *GraphInstance, nodeResults map[string]*NodeExecutionResult, err error, durationMs int64) *ExecutionResult {
	errStr := err.Error()
	return &ExecutionResult{
		InstanceID:  instance.ID,
		Status:      InstanceStatusFailed,
		Error:       &errStr,
		NodeResults: nodeResults,
		DurationMs:  int(durationMs),
	}
}
