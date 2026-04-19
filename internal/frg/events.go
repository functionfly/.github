package frg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// handleTriggerEvent processes trigger events (node execution start)
func (e *ExecutionEngine) handleTriggerEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	// Parse input
	var input map[string]interface{}
	json.Unmarshal(event.Payload, &input)

	// Start DAG execution
	order, err := e.computeExecutionOrder(runtime)
	if err != nil {
		logrus.WithError(err).Error("Failed to compute execution order")
		return
	}

	// Execute in order (async variant)
	for _, nodeID := range order {
		node := runtime.Nodes[nodeID]

		// Fire off node execution asynchronously
		go func(n *RuntimeNode, id string) {
			nodeInput := e.collectNodeInputs(n, runtime.Edges, nil) // Empty for now
			result, err := e.executeNode(ctx, runtime.Instance, n, nodeInput)

			if err != nil {
				errEvent := &GraphEvent{
					InstanceID: runtime.Instance.ID,
					EventType:  "error",
					NodeID:     &id,
					Payload:    []byte(`{"error": "` + err.Error() + `"}`),
				}
				runtime.InputChannel <- errEvent
			} else {
				resultJSON, _ := json.Marshal(result)
				completeEvent := &GraphEvent{
					InstanceID: runtime.Instance.ID,
					EventType:  "complete",
					NodeID:     &id,
					Payload:    resultJSON,
				}
				runtime.InputChannel <- completeEvent
			}
		}(node, nodeID)
	}
}

// handleStreamEvent processes streaming data events
func (e *ExecutionEngine) handleStreamEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	if event.NodeID == nil {
		return
	}

	sourceNodeID := *event.NodeID

	// Parse stream payload
	var streamData interface{}
	if err := json.Unmarshal(event.Payload, &streamData); err != nil {
		logrus.WithError(err).WithField("instance_id", runtime.Instance.ID).Warn("Failed to unmarshal stream payload")
		return
	}

	// Find edges originating from this node
	for _, edge := range runtime.Edges {
		if edge.Definition.SourceNodeID != sourceNodeID {
			continue
		}

		targetNode := runtime.Nodes[edge.Definition.TargetNodeID]
		if targetNode == nil {
			continue
		}

		// Evaluate condition if present (applies to all edge types)
		if edge.Definition.Condition != nil {
			if !e.evaluateCondition(streamData, edge.Definition.Condition) {
				continue // Skip this edge - condition not met
			}
		}

		switch edge.Definition.Type {
		case EdgeTypeStream:
			// Streaming edge: send to edge buffer (continuous flow)
			if edge.Buffer != nil {
				select {
				case edge.Buffer <- streamData:
					// Successfully queued to edge buffer
				default:
					// Buffer full - drop with warning
					logrus.WithFields(logrus.Fields{
						"instance_id": runtime.Instance.ID,
						"edge_id":     edge.Definition.ID,
					}).Warn("Streaming edge buffer full, dropping data")
				}
			}

		default:
			// Standard edge: accumulate for batch processing
			if targetNode.InputBuffer != nil {
				select {
				case targetNode.InputBuffer <- streamData:
				default:
					logrus.WithFields(logrus.Fields{
						"instance_id": runtime.Instance.ID,
						"node_id":     targetNode.Ref.NodeID,
					}).Warn("Node input buffer full")
				}
			}
		}

		// Update edge execution metrics asynchronously
		go func(edgeID string, payloadLen int) {
			// Create or update edge execution record
			edgeExec := &GraphEdgeExecution{
				InstanceID:         runtime.Instance.ID,
				SourceNodeID:       sourceNodeID,
				TargetNodeID:       edge.Definition.TargetNodeID,
				Status:             "active",
				EdgeType:           edge.Definition.Type,
				RecordsTransferred: 1,
				BytesTransferred:   int64(payloadLen),
			}
			if err := e.repo.CreateEdgeExecution(ctx, edgeExec); err != nil {
				logrus.WithError(err).WithField("edge_id", edgeID).Debug("Failed to record edge execution")
			}
		}(edge.Definition.ID, len(event.Payload))
	}
}

// handleNodeCompleteEvent processes node completion and propagates to downstream nodes
func (e *ExecutionEngine) handleNodeCompleteEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	if event.NodeID == nil {
		return
	}

	completedNodeID := *event.NodeID

	// Parse the completed node's output
	var nodeOutput map[string]interface{}
	if err := json.Unmarshal(event.Payload, &nodeOutput); err != nil {
		logrus.WithError(err).WithField("node_id", completedNodeID).Warn("Failed to unmarshal node completion output")
		return
	}

	// Mark source node as completed in runtime state
	if completedNode := runtime.Nodes[completedNodeID]; completedNode != nil && completedNode.State != nil {
		completedNode.State.Status = "completed"
		outputJSON, _ := json.Marshal(nodeOutput)
		completedNode.State.Output = outputJSON
	}

	// Find all downstream edges from this node and propagate
	for _, edge := range runtime.Edges {
		if edge.Definition.SourceNodeID != completedNodeID {
			continue
		}

		targetNode := runtime.Nodes[edge.Definition.TargetNodeID]
		if targetNode == nil {
			continue
		}

		// Check if target node is already completed or failed (don't propagate to finished nodes)
		if targetNode.State != nil && (targetNode.State.Status == "completed" || targetNode.State.Status == "failed") {
			continue
		}

		// Apply data mapping from source output to target input buffer
		targetInput := make(map[string]interface{})

		// Drain existing accumulated input from buffer if present
		if targetNode.InputBuffer != nil {
			select {
			case existing := <-targetNode.InputBuffer:
				if existingMap, ok := existing.(map[string]interface{}); ok {
					targetInput = existingMap
				}
			default:
				// No existing input
			}
		}

		// Apply the edge's data mapping
		e.applyDataMapping(targetInput, nodeOutput, &edge.Definition.Mapping)

		// Store accumulated input back to buffer
		if targetNode.InputBuffer != nil {
			select {
			case targetNode.InputBuffer <- targetInput:
			default:
				logrus.WithFields(logrus.Fields{
					"instance_id": runtime.Instance.ID,
					"node_id":     targetNode.Ref.NodeID,
				}).Warn("Target node input buffer full during propagation")
			}
		}

		// Check if target node is ready to execute (all dependencies satisfied)
		if e.isNodeReadyToExecute(runtime, targetNode) {
			// Update state to executing
			if targetNode.State != nil {
				targetNode.State.Status = "executing"
			}

			// Trigger execution based on edge type
			switch edge.Definition.Type {
			case EdgeTypeAsync:
				// Fire-and-forget: execute in goroutine
				go func(node *RuntimeNode, input map[string]interface{}) {
					_, err := e.executeNode(ctx, runtime.Instance, node, input)
					if err != nil {
						errorEvent := &GraphEvent{
							InstanceID: runtime.Instance.ID,
							EventType:  "error",
							NodeID:     &node.Ref.NodeID,
							Payload:    []byte(fmt.Sprintf(`{"error":%q}`, err.Error())),
						}
						runtime.InputChannel <- errorEvent
					}
				}(targetNode, targetInput)

			default:
				// Sync/stream: queue trigger event for ordered execution
				triggerEvent := &GraphEvent{
					InstanceID: runtime.Instance.ID,
					EventType:  "trigger",
					NodeID:     &targetNode.Ref.NodeID,
					Payload:    []byte(fmt.Sprintf(`{"input_ready":true,"source":"%s"}`, completedNodeID)),
				}
				runtime.InputChannel <- triggerEvent
			}
		} else {
			// Node not ready yet - mark as waiting for other dependencies
			if targetNode.State != nil {
				targetNode.State.Status = "waiting"
			}
		}
	}
}

// handleNodeErrorEvent processes node errors (retry/fallback logic)
func (e *ExecutionEngine) handleNodeErrorEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	if event.NodeID == nil {
		return
	}
	nodeID := *event.NodeID

	logger := logrus.WithFields(logrus.Fields{
		"instance_id": runtime.Instance.ID,
		"node_id":     nodeID,
	})

	// Get the node runtime
	node := runtime.Nodes[nodeID]
	if node == nil {
		logger.Error("Node not found in runtime")
		return
	}

	// Parse error from event payload
	var errorPayload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &errorPayload); err != nil {
		logger.WithError(err).Warn("Failed to parse error payload")
		errorPayload = map[string]interface{}{"error": string(event.Payload)}
	}

	errorMsg := ""
	if err, ok := errorPayload["error"].(string); ok {
		errorMsg = err
	}

	// Get current node state
	if node.State == nil {
		node.State = &NodeState{
			Status:       "failed",
			Error:        &errorMsg,
			AttemptCount: 1,
		}
	} else {
		node.State.AttemptCount++
		node.State.Error = &errorMsg
	}

	currentAttempt := node.State.AttemptCount

	// Find retry policy from incoming edges
	retryPolicy := e.findRetryPolicy(runtime, nodeID)
	fallbackNodeID := e.findFallbackNodeID(runtime, nodeID)

	logger = logger.WithFields(logrus.Fields{
		"attempt":       currentAttempt,
		"max_attempts":  retryPolicy.MaxAttempts,
		"retryable":     retryPolicy.MaxAttempts > 1,
		"fallback_node": fallbackNodeID,
	})

	// Check if error is retryable
	isRetryable := e.isErrorRetryable(errorMsg, retryPolicy)
	shouldRetry := isRetryable && currentAttempt < retryPolicy.MaxAttempts

	if shouldRetry {
		// Calculate backoff delay
		backoff := e.calculateBackoff(currentAttempt, retryPolicy)
		logger.WithField("backoff_ms", backoff.Milliseconds()).Info("Scheduling node retry")

		// Update node state
		node.State.Status = "retrying"

		// Schedule retry with backoff
		go func() {
			time.Sleep(backoff)

			retryEvent := &GraphEvent{
				InstanceID: runtime.Instance.ID,
				EventType:  "retry",
				NodeID:     &nodeID,
				Payload:    []byte(fmt.Sprintf(`{"retry_attempt":%d,"previous_error":"%s"}`, currentAttempt, errorMsg)),
			}

			select {
			case runtime.InputChannel <- retryEvent:
				logger.WithField("retry_attempt", currentAttempt+1).Info("Retry event queued")
			case <-ctx.Done():
				logger.Warn("Context cancelled, retry event not queued")
			}
		}()

		return
	}

	// Max retries exceeded or not retryable - try fallback
	if fallbackNodeID != "" {
		logger.Info("Triggering fallback node")

		fallbackNode := runtime.Nodes[fallbackNodeID]
		if fallbackNode == nil {
			logger.Error("Fallback node not found, marking as failed")
			node.State.Status = "failed"
			return
		}

		// Mark original node as failed with fallback triggered
		node.State.Status = "failed_with_fallback"
		fallbackTriggered := true
		node.State.FallbackTriggered = &fallbackTriggered
		node.State.FallbackNodeID = &fallbackNodeID

		// Trigger fallback node with error context
		fallbackInput := map[string]interface{}{
			"original_node_id": nodeID,
			"original_error":   errorMsg,
			"fallback_input":   errorPayload,
		}
		fallbackInputJSON, _ := json.Marshal(fallbackInput)

		triggerEvent := &GraphEvent{
			InstanceID: runtime.Instance.ID,
			EventType:  "trigger",
			NodeID:     &fallbackNodeID,
			Payload:    fallbackInputJSON,
		}

		select {
		case runtime.InputChannel <- triggerEvent:
			logger.Info("Fallback node triggered")
		case <-ctx.Done():
			logger.Warn("Context cancelled, fallback trigger not queued")
		}

		return
	}

	// No retry, no fallback - node is failed
	logger.Info("Node failed permanently (no retry or fallback available)")
	node.State.Status = "failed"
}

// handleCheckpointEvent processes state snapshot checkpoints
func (e *ExecutionEngine) handleCheckpointEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	logger := logrus.WithField("instance_id", runtime.Instance.ID)

	// Get or create state manager for this instance
	// Use TenantID from definition (pointer) and graph name from definition
	tenantID := uuid.Nil
	if runtime.Definition != nil && runtime.Definition.TenantID != nil {
		tenantID = *runtime.Definition.TenantID
	}
	graphName := ""
	if runtime.Definition != nil {
		graphName = runtime.Definition.Name
	}
	stateManager, err := e.GetStateManager(ctx, runtime.Instance.ID, tenantID, graphName)
	if err != nil {
		logger.WithError(err).Error("Failed to get state manager for checkpoint")
		return
	}

	// Collect current node states for the checkpoint
	nodeStates := make(map[string]interface{}, len(runtime.Nodes))
	for nodeID, node := range runtime.Nodes {
		if node.State != nil {
			nodeStates[nodeID] = node.State
		}
	}

	// Create the checkpoint/snapshot
	snapshot, err := stateManager.CreateCheckpoint(ctx, runtime.Instance.ID, nodeStates)
	if err != nil {
		logger.WithError(err).Error("Failed to create state checkpoint")
		return
	}

	logger.WithFields(logrus.Fields{
		"snapshot_id": snapshot.SnapshotID,
		"node_count":  len(nodeStates),
		"state_keys":  len(snapshot.State),
	}).Info("Created state checkpoint")
}

// isNodeReadyToExecute checks if a node has received input from all required upstream nodes
func (e *ExecutionEngine) isNodeReadyToExecute(runtime *GraphRuntime, node *RuntimeNode) bool {
	// Count dependencies (incoming edges)
	dependencyCount := 0
	completedDependencies := 0

	for _, edge := range runtime.Edges {
		if edge.Definition.TargetNodeID != node.Ref.NodeID {
			continue
		}
		dependencyCount++

		sourceNode := runtime.Nodes[edge.Definition.SourceNodeID]
		if sourceNode != nil && sourceNode.State != nil && sourceNode.State.Status == "completed" {
			completedDependencies++
		}
	}

	// If no dependencies, it's ready (shouldn't happen in practice for downstream nodes)
	if dependencyCount == 0 {
		return true
	}

	// Node is ready when all dependencies are completed
	return completedDependencies >= dependencyCount
}

// findRetryPolicy finds the retry policy from edges targeting this node
// Returns the most restrictive (highest max attempts) policy found
func (e *ExecutionEngine) findRetryPolicy(runtime *GraphRuntime, nodeID string) *RetryPolicy {
	var policy *RetryPolicy
	maxAttempts := 1 // Default: no retries

	for _, edge := range runtime.Edges {
		if edge.Definition.TargetNodeID != nodeID {
			continue
		}

		if edge.Definition.RetryPolicy != nil {
			if policy == nil || edge.Definition.RetryPolicy.MaxAttempts > maxAttempts {
				policy = edge.Definition.RetryPolicy
				maxAttempts = edge.Definition.RetryPolicy.MaxAttempts
			}
		}
	}

	if policy == nil {
		// Return default retry policy (no retries)
		return &RetryPolicy{
			MaxAttempts:    1,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		}
	}

	return policy
}

// findFallbackNodeID finds the fallback node ID from edges targeting this node
func (e *ExecutionEngine) findFallbackNodeID(runtime *GraphRuntime, nodeID string) string {
	for _, edge := range runtime.Edges {
		if edge.Definition.TargetNodeID != nodeID {
			continue
		}

		if edge.Definition.FallbackNodeID != nil && *edge.Definition.FallbackNodeID != "" {
			return *edge.Definition.FallbackNodeID
		}
	}

	return ""
}

// isErrorRetryable checks if an error message indicates a retryable error
func (e *ExecutionEngine) isErrorRetryable(errorMsg string, policy *RetryPolicy) bool {
	if policy == nil || policy.MaxAttempts <= 1 {
		return false
	}

	// If specific retryable errors are configured, check them
	if len(policy.RetryableErrors) > 0 {
		errorLower := strings.ToLower(errorMsg)
		for _, code := range policy.RetryableErrors {
			if strings.Contains(errorLower, strings.ToLower(code)) {
				return true
			}
		}
		// Error not in retryable list
		return false
	}

	// Default: retry on common transient errors
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"temporary",
		"transient",
		"rate limit",
		"too many requests",
		"service unavailable",
		"503",
		"502",
		"504",
		"429",
	}

	errorLower := strings.ToLower(errorMsg)
	for _, pattern := range retryablePatterns {
		if strings.Contains(errorLower, pattern) {
			return true
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for a retry attempt
// Uses exponential backoff with jitter
func (e *ExecutionEngine) calculateBackoff(attempt int, policy *RetryPolicy) time.Duration {
	if policy == nil {
		return 100 * time.Millisecond
	}

	// Calculate exponential backoff
	backoff := float64(policy.InitialBackoff)
	if attempt > 1 {
		backoff = backoff * math.Pow(policy.BackoffFactor, float64(attempt-1))
	}

	// Cap at max backoff
	if backoff > float64(policy.MaxBackoff) {
		backoff = float64(policy.MaxBackoff)
	}

	// Add jitter (±20%) to prevent thundering herd
	// Simple pseudo-random based on attempt number
	pseudoRand := float64((attempt*6364136223846793005+1442695040888963407)%1000) / 1000.0
	jitter := backoff * 0.2 * (2*pseudoRand - 1)
	backoff = backoff + jitter

	return time.Duration(backoff)
}
