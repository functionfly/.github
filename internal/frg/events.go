package frg

import (
	"context"
	"encoding/json"
	"fmt"

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
	// TODO: Implement retry/fallback logic
	if event.NodeID == nil {
		return
	}
	nodeID := *event.NodeID
	logrus.WithFields(logrus.Fields{
		"instance_id": runtime.Instance.ID,
		"node_id":     nodeID,
		"payload":     string(event.Payload),
	}).Warn("Node error event received (retry/fallback not yet implemented)")
}

// handleCheckpointEvent processes state snapshot checkpoints
func (e *ExecutionEngine) handleCheckpointEvent(ctx context.Context, runtime *GraphRuntime, event *GraphEvent) {
	// TODO: Save state snapshot
	logrus.WithField("instance_id", runtime.Instance.ID).Info("Checkpoint event received (state snapshot not yet implemented)")
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
