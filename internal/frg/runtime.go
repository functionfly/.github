package frg

import (
	"context"
	"encoding/json"
	"fmt"
)

// buildRuntime builds the runtime structure from a graph definition
func (e *ExecutionEngine) buildRuntime(ctx context.Context, def *GraphDefinition, instance *GraphInstance) (*GraphRuntime, error) {
	// Parse node refs
	var nodeRefs []GraphNodeRef
	if err := json.Unmarshal(def.NodeRefs, &nodeRefs); err != nil {
		return nil, fmt.Errorf("invalid node_refs: %w", err)
	}

	// Parse edges
	var edges []GraphEdge
	if err := json.Unmarshal(def.Edges, &edges); err != nil {
		return nil, fmt.Errorf("invalid edges: %w", err)
	}

	// Resolve all functions
	nodes := make(map[string]*RuntimeNode)
	for _, ref := range nodeRefs {
		fn, err := e.registryRepo.GetFunctionByAuthorName(ref.Author, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("function not found: %s/%s", ref.Author, ref.Name)
		}

		// Get specific version
		version, err := e.registryRepo.GetFunctionVersion(fn.ID, ref.Version)
		if err != nil {
			return nil, fmt.Errorf("version not found: %s/%s@%s", ref.Author, ref.Name, ref.Version)
		}

		node := &RuntimeNode{
			Ref:        &ref,
			Definition: version,
			State: &NodeState{
				Status: "pending",
			},
		}

		// Create channels for streaming nodes
		if def.ExecutionMode == ExecutionModeStreaming {
			node.InputBuffer = make(chan interface{}, 100)
			node.OutputBuffer = make(chan interface{}, 100)
		}

		nodes[ref.NodeID] = node
	}

	// Build runtime edges
	var runtimeEdges []*RuntimeEdge
	for _, edge := range edges {
		rtEdge := &RuntimeEdge{
			Definition: &edge,
		}

		// Create buffer for streaming edges
		if edge.Type == EdgeTypeStream {
			size := edge.BufferSize
			if size == 0 {
				size = 100
			}
			rtEdge.Buffer = make(chan interface{}, size)
		}

		runtimeEdges = append(runtimeEdges, rtEdge)
	}

	return &GraphRuntime{
		Instance:   instance,
		Definition: def,
		Nodes:      nodes,
		Edges:      runtimeEdges,
	}, nil
}

// computeExecutionOrder computes topological order for DAG execution
func (e *ExecutionEngine) computeExecutionOrder(runtime *GraphRuntime) ([]string, error) {
	// Build adjacency list
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for nodeID := range runtime.Nodes {
		inDegree[nodeID] = 0
	}

	for _, edge := range runtime.Edges {
		adjList[edge.Definition.SourceNodeID] = append(adjList[edge.Definition.SourceNodeID], edge.Definition.TargetNodeID)
		inDegree[edge.Definition.TargetNodeID]++
	}

	// Kahn's algorithm
	var queue []string
	for nodeID := range runtime.Nodes {
		if inDegree[nodeID] == 0 {
			queue = append(queue, nodeID)
		}
	}

	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(runtime.Nodes) {
		return nil, fmt.Errorf("graph contains cycles")
	}

	return order, nil
}

// collectNodeInputs collects inputs from predecessor edges
func (e *ExecutionEngine) collectNodeInputs(node *RuntimeNode, edges []*RuntimeEdge, results map[string]*NodeExecutionResult) map[string]interface{} {
	// Start with empty input
	input := make(map[string]interface{})

	// Collect from all incoming edges
	for _, edge := range edges {
		if edge.Definition.TargetNodeID == node.Ref.NodeID {
			// Get source node output
			sourceResult := results[edge.Definition.SourceNodeID]
			if sourceResult == nil || sourceResult.Output == nil {
				continue
			}

			var sourceOutput map[string]interface{}
			json.Unmarshal(sourceResult.Output, &sourceOutput)

			// Apply mapping
			e.applyDataMapping(input, sourceOutput, &edge.Definition.Mapping)
		}
	}

	return input
}
