// Package frg provides control flow primitives for graph execution
// These enable complex workflows: loops, conditionals, parallel execution, error handling
package frg

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ControlFlowNode types extend the standard graph execution with advanced primitives
type ControlFlowNodeType string

const (
	ControlFlowLoop       ControlFlowNodeType = "loop"           // Iterate over collection
	ControlFlowParallel   ControlFlowNodeType = "parallel"       // Execute branches in parallel
	ControlFlowCondition  ControlFlowNodeType = "condition"      // If/then/else branching
	ControlFlowErrorBound ControlFlowNodeType = "error_boundary" // Try/catch
	ControlFlowDelay      ControlFlowNodeType = "delay"          // Add delay/timers
	ControlFlowMerge      ControlFlowNodeType = "merge"          // Merge parallel branches
	ControlFlowWaitFor    ControlFlowNodeType = "wait_for"       // Wait for external event
)

// LoopNodeConfig configures a loop node
type LoopNodeConfig struct {
	CollectionInput string `json:"collection_input"`  // Input key containing collection
	MaxIterations   int    `json:"max_iterations"`    // Safety limit (default 1000)
	BatchSize       int    `json:"batch_size"`        // Batch for parallel iteration
	ContinueOnError bool   `json:"continue_on_error"` // Continue loop on item error
}

// ParallelNodeConfig configures parallel execution
type ParallelNodeConfig struct {
	Branches       []string `json:"branches"`        // Node IDs of branch entry points
	MaxConcurrency int      `json:"max_concurrency"` // Max parallel (default 10)
	WaitAll        bool     `json:"wait_all"`        // Wait for all (vs first)
	FailFast       bool     `json:"fail_fast"`       // Fail on first error
}

// ConditionNodeConfig configures conditional branching
type ConditionNodeConfig struct {
	Conditions []ConditionBranch `json:"conditions"`  // Ordered conditions, first match wins
	ElseBranch string            `json:"else_branch"` // Fallback branch
}

// ConditionBranch is a single if/then pair
type ConditionBranch struct {
	Condition Condition `json:"condition"`
	Target    string    `json:"target"` // Node ID to execute if condition true
}

// ErrorBoundaryConfig configures error handling
type ErrorBoundaryConfig struct {
	TryNodes       []string `json:"try_nodes"`        // Nodes to try
	CatchNode      string   `json:"catch_node"`       // Node on error
	FinallyNode    string   `json:"finally_node"`     // Node always executed
	RetryAttempts  int      `json:"retry_attempts"`   // Auto-retry count
	RetryBackoffMs int      `json:"retry_backoff_ms"` // Backoff between retries
}

// DelayNodeConfig configures a delay
type DelayNodeConfig struct {
	DelayMs        int    `json:"delay_ms"`        // Fixed delay
	DelayInputKey  string `json:"delay_input_key"` // Use input value
	UntilTimestamp *int64 `json:"until_timestamp"` // Delay until specific time
	CronExpression string `json:"cron_expression"` // Wait until next cron match
}

// executeControlFlowNode executes a control flow node
func (e *ExecutionEngine) executeControlFlowNode(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeID string,
	nodeType ControlFlowNodeType,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	switch nodeType {
	case ControlFlowLoop:
		return e.executeLoopNode(ctx, instance, runtime, nodeID, config, input)
	case ControlFlowParallel:
		return e.executeParallelNode(ctx, instance, runtime, nodeID, config, input)
	case ControlFlowCondition:
		return e.executeConditionNode(ctx, instance, runtime, nodeID, config, input)
	case ControlFlowErrorBound:
		return e.executeErrorBoundary(ctx, instance, runtime, nodeID, config, input)
	case ControlFlowDelay:
		return e.executeDelayNode(ctx, instance, nodeID, config, input)
	default:
		return nil, fmt.Errorf("unknown control flow type: %s", nodeType)
	}
}

// executeLoopNode iterates over a collection
func (e *ExecutionEngine) executeLoopNode(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeID string,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	var cfg LoopNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		cfg = LoopNodeConfig{MaxIterations: 100}
	}
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 1000
	}

	// Get collection from input
	collection, ok := input[cfg.CollectionInput]
	if !ok {
		return nil, fmt.Errorf("loop node missing collection input: %s", cfg.CollectionInput)
	}

	// Convert to slice
	items := toSlice(collection)
	if len(items) == 0 {
		return &NodeExecutionResult{
			Status: "completed",
			Output: json.RawMessage(`{"results": [], "count": 0}`),
		}, nil
	}

	// Limit iterations
	if len(items) > cfg.MaxIterations {
		items = items[:cfg.MaxIterations]
	}

	// Execute iterations
	results := make([]interface{}, 0, len(items))
	var loopErrors []error

	for i, item := range items {
		iterationInput := map[string]interface{}{
			"_loop_index": i,
			"_loop_item":  item,
			"_loop_count": len(items),
			"_is_last":    i == len(items)-1,
			"_original":   input,
		}

		// Find the next node after this loop node
		var nextNodeID string
		for _, edge := range runtime.Edges {
			if edge.Definition.SourceNodeID == nodeID {
				nextNodeID = edge.Definition.TargetNodeID
				break
			}
		}

		if nextNodeID == "" {
			loopErrors = append(loopErrors, fmt.Errorf("no successor node for loop body"))
			if !cfg.ContinueOnError {
				break
			}
			continue
		}

		// Execute body node
		nextNode := runtime.Nodes[nextNodeID]
		result, err := e.executeNode(ctx, instance, nextNode, iterationInput)
		if err != nil {
			loopErrors = append(loopErrors, err)
			if !cfg.ContinueOnError {
				break
			}
			results = append(results, map[string]interface{}{
				"index": i,
				"error": err.Error(),
			})
		} else {
			var output interface{}
			json.Unmarshal(result.Output, &output)
			results = append(results, map[string]interface{}{
				"index":  i,
				"output": output,
			})
		}
	}

	// Build result
	output := map[string]interface{}{
		"results":    results,
		"count":      len(results),
		"iterations": len(items),
	}

	if len(loopErrors) > 0 {
		output["errors"] = len(loopErrors)
	}

	outputJSON, _ := json.Marshal(output)
	return &NodeExecutionResult{
		Status: "completed",
		Output: outputJSON,
	}, nil
}

// executeParallelNode executes branches in parallel
func (e *ExecutionEngine) executeParallelNode(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeID string,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	var cfg ParallelNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid parallel config: %w", err)
	}

	if cfg.MaxConcurrency == 0 {
		cfg.MaxConcurrency = 10
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, cfg.MaxConcurrency)

	type branchResult struct {
		index  int
		nodeID string
		result *NodeExecutionResult
		err    error
	}

	resultChan := make(chan branchResult, len(cfg.Branches))

	// Launch all branches
	var wg sync.WaitGroup
	for i, branchNodeID := range cfg.Branches {
		wg.Add(1)
		go func(idx int, nid string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			node := runtime.Nodes[nid]
			if node == nil {
				resultChan <- branchResult{index: idx, nodeID: nid, err: fmt.Errorf("branch node not found: %s", nid)}
				return
			}

			result, err := e.executeNode(ctx, instance, node, input)
			resultChan <- branchResult{index: idx, nodeID: nid, result: result, err: err}
		}(i, branchNodeID)

		// Fail fast check
		if cfg.FailFast && i > 0 {
			select {
			case r := <-resultChan:
				if r.err != nil {
					wg.Wait()
					return nil, fmt.Errorf("parallel branch %d failed: %w", r.index, r.err)
				}
			default:
			}
		}
	}

	// Wait for all to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	results := make([]branchResult, 0, len(cfg.Branches))
	for r := range resultChan {
		results = append(results, r)
	}

	// Build output
	output := map[string]interface{}{
		"branches": make(map[string]interface{}),
	}

	branchesMap := output["branches"].(map[string]interface{})
	var hasError bool

	for _, r := range results {
		if r.err != nil {
			hasError = true
			branchesMap[r.nodeID] = map[string]interface{}{
				"success": false,
				"error":   r.err.Error(),
			}
		} else {
			var res interface{}
			json.Unmarshal(r.result.Output, &res)
			branchesMap[r.nodeID] = map[string]interface{}{
				"success": true,
				"output":  res,
			}
		}
	}

	outputJSON, _ := json.Marshal(output)

	if hasError {
		return &NodeExecutionResult{
			Status: "completed_with_errors",
			Output: outputJSON,
		}, nil
	}

	return &NodeExecutionResult{
		Status: "completed",
		Output: outputJSON,
	}, nil
}

// executeConditionNode evaluates conditions and routes
func (e *ExecutionEngine) executeConditionNode(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeID string,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	var cfg ConditionNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid condition config: %w", err)
	}

	// Evaluate conditions in order
	var selectedBranch string
	for _, branch := range cfg.Conditions {
		if evaluateCondition(branch.Condition, input) {
			selectedBranch = branch.Target
			break
		}
	}

	// Use else branch if no condition matched
	if selectedBranch == "" {
		selectedBranch = cfg.ElseBranch
	}

	// Execute selected branch
	if selectedBranch == "" {
		// No branch selected, just return input
		outputJSON, _ := json.Marshal(input)
		return &NodeExecutionResult{
			Status: "completed",
			Output: outputJSON,
		}, nil
	}

	node := runtime.Nodes[selectedBranch]
	if node == nil {
		return nil, fmt.Errorf("condition target node not found: %s", selectedBranch)
	}

	result, err := e.executeNode(ctx, instance, node, input)
	if err != nil {
		return nil, err
	}

	// Wrap with branch info
	var output interface{}
	json.Unmarshal(result.Output, &output)

	wrapped := map[string]interface{}{
		"_selected_branch": selectedBranch,
		"output":           output,
	}

	outputJSON, _ := json.Marshal(wrapped)
	return &NodeExecutionResult{
		Status: result.Status,
		Output: outputJSON,
		CertID: result.CertID,
	}, nil
}

// executeErrorBoundary implements try/catch/finally
func (e *ExecutionEngine) executeErrorBoundary(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeID string,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	var cfg ErrorBoundaryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		cfg = ErrorBoundaryConfig{RetryAttempts: 0}
	}

	// Try block with optional retries
	var lastError error
	var lastResult *NodeExecutionResult

	for attempt := 0; attempt <= cfg.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(cfg.RetryBackoffMs) * time.Millisecond)
			logrus.WithField("attempt", attempt).Info("Retrying after error")
		}

		// Execute try nodes sequentially
		tryResult, tryErr := e.executeNodesSequential(ctx, instance, runtime, cfg.TryNodes, input)

		if tryErr == nil {
			lastResult = tryResult
			break
		}

		lastError = tryErr
	}

	// If still failed, execute catch block
	if lastError != nil && cfg.CatchNode != "" {
		catchNode := runtime.Nodes[cfg.CatchNode]
		if catchNode != nil {
			catchInput := map[string]interface{}{
				"_error":       lastError.Error(),
				"_original":    input,
				"_retry_count": cfg.RetryAttempts,
			}

			catchResult, err := e.executeNode(ctx, instance, catchNode, catchInput)
			if err != nil {
				// Catch block also failed
				return nil, fmt.Errorf("error boundary catch failed: %w", err)
			}

			var catchOutput interface{}
			json.Unmarshal(catchResult.Output, &catchOutput)

			lastResult = &NodeExecutionResult{
				Status: "completed_caught",
				Output: json.RawMessage(`{"caught": true}`),
			}
		}
	}

	// Always execute finally block if defined
	if cfg.FinallyNode != "" {
		finallyNode := runtime.Nodes[cfg.FinallyNode]
		if finallyNode != nil {
			finallyInput := map[string]interface{}{
				"_original":   input,
				"_had_error":  lastError != nil,
				"_try_result": lastResult,
			}
			e.executeNode(ctx, instance, finallyNode, finallyInput)
		}
	}

	if lastError != nil && cfg.CatchNode == "" {
		return nil, lastError
	}

	return lastResult, nil
}

// executeDelayNode adds a delay
func (e *ExecutionEngine) executeDelayNode(
	ctx context.Context,
	instance *GraphInstance,
	nodeID string,
	config json.RawMessage,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	var cfg DelayNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		cfg = DelayNodeConfig{DelayMs: 1000}
	}

	delay := time.Duration(cfg.DelayMs) * time.Millisecond

	if cfg.DelayInputKey != "" {
		if val, ok := input[cfg.DelayInputKey]; ok {
			switch v := val.(type) {
			case float64:
				delay = time.Duration(v) * time.Millisecond
			case int:
				delay = time.Duration(v) * time.Millisecond
			}
		}
	}

	// Wait for delay or context cancellation
	select {
	case <-time.After(delay):
		outputJSON, _ := json.Marshal(map[string]interface{}{
			"delayed_ms": int(delay.Milliseconds()),
			"resumed_at": time.Now().Format(time.RFC3339),
		})
		return &NodeExecutionResult{
			Status: "completed",
			Output: outputJSON,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// executeNodesSequential executes multiple nodes in sequence
func (e *ExecutionEngine) executeNodesSequential(
	ctx context.Context,
	instance *GraphInstance,
	runtime *GraphRuntime,
	nodeIDs []string,
	input map[string]interface{},
) (*NodeExecutionResult, error) {
	currentData := input

	for _, nodeID := range nodeIDs {
		node := runtime.Nodes[nodeID]
		if node == nil {
			return nil, fmt.Errorf("node not found in try block: %s", nodeID)
		}

		result, err := e.executeNode(ctx, instance, node, currentData)
		if err != nil {
			return nil, err
		}

		var output interface{}
		json.Unmarshal(result.Output, &output)

		// Next node gets this output
		currentData = map[string]interface{}{
			"_previous_output": output,
			"_original":        input,
		}

		lastResult = result
	}

	return lastResult, nil
}

// evaluateCondition evaluates a condition against input
type evaluationContext struct {
	Input map[string]interface{} `json:"input"`
}

func evaluateCondition(cond Condition, input map[string]interface{}) bool {
	// Get value from input using the field path (JSONPath style)
	value := getValueFromPath(input, cond.Field)

	switch cond.Operator {
	case "eq":
		return compareValues(value, cond.Value) == 0
	case "ne":
		return compareValues(value, cond.Value) != 0
	case "gt":
		return compareValues(value, cond.Value) > 0
	case "lt":
		return compareValues(value, cond.Value) < 0
	case "gte":
		return compareValues(value, cond.Value) >= 0
	case "lte":
		return compareValues(value, cond.Value) <= 0
	case "contains":
		return containsValue(value, cond.Value)
	case "exists":
		return value != nil
	case "not_exists":
		return value == nil
	case "regex":
		return matchRegex(value, cond.Value)
	default:
		return false
	}
}

func getValueFromPath(input map[string]interface{}, path string) interface{} {
	if path == "" || path == "*" {
		return input
	}

	// Simple JSONPath-like navigation
	if path[0] == '$' {
		path = path[1:]
	}
	if path[0] == '.' {
		path = path[1:]
	}

	parts := splitPath(path)
	current := interface{}(input)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			if idx, err := parseIndex(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

func splitPath(path string) []string {
	// Simple split by dots and brackets
	var parts []string
	current := ""
	inBracket := false

	for _, char := range path {
		switch char {
		case '.':
			if !inBracket && current != "" {
				parts = append(parts, current)
				current = ""
			}
		case '[':
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			inBracket = true
		case ']':
			inBracket = false
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func parseIndex(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Numeric comparison
	fa, aok := toFloat64(a)
	fb, bok := toFloat64(b)
	if aok && bok {
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0
	}

	// String comparison
	sa, aok := a.(string)
	sb, bok := b.(string)
	if aok && bok {
		if sa < sb {
			return -1
		}
		if sa > sb {
			return 1
		}
		return 0
	}

	// Boolean comparison
	ba, aok := a.(bool)
	bb, bok := b.(bool)
	if aok && bok {
		if !ba && bb {
			return -1
		}
		if ba && !bb {
			return 1
		}
		return 0
	}

	return 0
}

func containsValue(container, item interface{}) bool {
	if container == nil || item == nil {
		return false
	}

	// String contains
	cs, cok := container.(string)
	is, iok := item.(string)
	if cok && iok {
		return len(is) > 0 && len(cs) > 0 && containsSubstring(cs, is)
	}

	// Array contains
	arr, ok := container.([]interface{})
	if ok {
		for _, v := range arr {
			if compareValues(v, item) == 0 {
				return true
			}
		}
	}

	return false
}

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

func matchRegex(value, pattern interface{}) bool {
	// Use Go's standard regexp library for production regex matching
	s, ok := value.(string)
	if !ok {
		return false
	}
	p, ok := pattern.(string)
	if !ok {
		return false
	}
	re, err := regexp.Compile(p)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}

	// Already a slice
	if slice, ok := v.([]interface{}); ok {
		return slice
	}

	// Array
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result
	}

	// Single item as slice
	return []interface{}{v}
}

// lastResult tracks the result in sequential execution
var lastResult *NodeExecutionResult
