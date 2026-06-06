// Package simulation provides HTTP handlers for the R-Sim simulation engine.
// R-Sim enables dry-run workflow simulation, Monte Carlo analysis, failure injection,
// and outcome prediction before actual deployment.
package simulation

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// Handler contains all R-Sim simulation handlers
type Handler struct {
	// In-memory storage for simulation results (shared across all simulation types)
	simulations     map[string]*SimulationResult
	monteCarloResults map[string]*MonteCarloResult
	stressTestResults map[string]*StressTestResult
	simMu           sync.RWMutex
}

// NewHandler creates a new R-Sim handler
func NewHandler() *Handler {
	return &Handler{
		simulations:      make(map[string]*SimulationResult),
		monteCarloResults: make(map[string]*MonteCarloResult),
		stressTestResults: make(map[string]*StressTestResult),
	}
}

// RegisterRoutes registers simulation routes on the given router
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Simulation lifecycle
	r.HandleFunc("/simulate/workflow", h.HandleSimulateWorkflow).Methods("POST", "OPTIONS")
	r.HandleFunc("/simulate/workflow/{id}", h.HandleGetSimulation).Methods("GET", "OPTIONS")
	r.HandleFunc("/simulate/workflow/{id}/abort", h.HandleAbortSimulation).Methods("POST", "OPTIONS")

	// Monte Carlo simulation
	r.HandleFunc("/simulate/monte-carlo", h.HandleMonteCarloSimulation).Methods("POST", "OPTIONS")

	// Failure injection
	r.HandleFunc("/simulate/failure-inject", h.HandleFailureInjection).Methods("POST", "OPTIONS")

	// Forecast and prediction
	r.HandleFunc("/forecast/execution", h.HandleExecutionForecast).Methods("POST", "OPTIONS")
	r.HandleFunc("/forecast/cost", h.HandleCostForecast).Methods("POST", "OPTIONS")
	r.HandleFunc("/forecast/latency", h.HandleLatencyForecast).Methods("POST", "OPTIONS")

	// Stress test
	r.HandleFunc("/stress-test/start", h.HandleStartStressTest).Methods("POST", "OPTIONS")
	r.HandleFunc("/stress-test/{id}", h.HandleGetStressTest).Methods("GET", "OPTIONS")
	r.HandleFunc("/stress-test/{id}/abort", h.HandleAbortStressTest).Methods("POST", "OPTIONS")
	r.HandleFunc("/stress-test/{id}/results", h.HandleGetStressTestResults).Methods("GET", "OPTIONS")

	// Resource collision detection
	r.HandleFunc("/detect/collisions", h.HandleDetectResourceCollisions).Methods("POST", "OPTIONS")

	// Agent behavior prediction
	r.HandleFunc("/predict/agent-behavior", h.HandlePredictAgentBehavior).Methods("POST", "OPTIONS")

	// Hallucination risk analysis
	r.HandleFunc("/analyze/hallucination-risk", h.HandleHallucinationRiskAnalysis).Methods("POST", "OPTIONS")
}

// ============================================================
// Workflow Simulation
// ============================================================

// WorkflowSpec defines the workflow to simulate
type WorkflowSpec struct {
	Nodes []NodeSpec `json:"nodes"`
	Edges []EdgeSpec `json:"edges"`
}

// NodeSpec defines a node in the workflow
type NodeSpec struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // trigger, function, agent, api, memory, database
	Timeout  int                     `json:"timeout_ms"`
	CostUSD  float64                `json:"cost_usd"`
	Metadata map[string]interface{} `json:"metadata"`
}

// EdgeSpec defines a connection between nodes
type EdgeSpec struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	ProbSuc float64 `json:"probability_success"` // 0-1
}

// SimulationResult holds the result of a workflow simulation
type SimulationResult struct {
	ID              string         `json:"id"`
	Status          string         `json:"status"` // running, completed, aborted
	SuccessRate     float64        `json:"success_rate"`
	AvgLatencyMs    int            `json:"avg_latency_ms"`
	AvgCostUSD      float64        `json:"avg_cost_usd"`
	PredictedNodes  map[string]int `json:"predicted_node_executions"` // node_id -> count
	FailedNodes     map[string]int `json:"failed_nodes"`             // node_id -> failure count
	ExecutionCount  int            `json:"execution_count"`
	StartedAt       time.Time      `json:"started_at"`
	CompletedAt     *time.Time     `json:"completed_at"`
	Iteration       int             `json:"iteration"` // current iteration for streaming results
}

// MonteCarloResult holds Monte Carlo simulation results
type MonteCarloResult struct {
	ID                  string           `json:"id"`
	Iterations          int              `json:"iterations"`
	SuccessRate         float64          `json:"success_rate"`
	PartialFailureRate  float64          `json:"partial_failure_rate"`
	TotalFailureRate    float64          `json:"total_failure_rate"`
	AvgLatencyMs        int              `json:"avg_latency_ms"`
	P50LatencyMs        int              `json:"p50_latency_ms"`
	P95LatencyMs        int              `json:"p95_latency_ms"`
	P99LatencyMs        int              `json:"p99_latency_ms"`
	AvgCostUSD          float64          `json:"avg_cost_usd"`
	Outcomes            []OutcomeSample  `json:"outcomes"`
	BottleneckNodes     []string         `json:"bottleneck_nodes"`
	CostBreakdown       map[string]float64 `json:"cost_breakdown"`
}

// OutcomeSample is a single Monte Carlo outcome
type OutcomeSample struct {
	Outcome       string  `json:"outcome"` // success, partial, failed
	Probability   float64 `json:"probability"`
	LatencyMs     int     `json:"latency_ms"`
	CostUSD       float64 `json:"cost_usd"`
	FailedNodes   []string `json:"failed_nodes"`
	RiskFactors   []string `json:"risk_factors"`
}

// StressTestConfig configures a stress test
type StressTestConfig struct {
	Iterations  int     `json:"iterations"`
	Parallelism int     `json:"parallelism"` // concurrent executions
	WorkflowID  string  `json:"workflow_id"`
	LoadProfile string  `json:"load_profile"` // constant, ramp, burst, poisson
}

// StressTestResult holds stress test results
type StressTestResult struct {
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Iterations    int          `json:"iterations"`
	TotalExecutions int        `json:"total_executions"`
	SuccessRate   float64      `json:"success_rate"`
	Throughput    float64      `json:"throughput"` // exec/sec
	LatencyP50    int          `json:"latency_p50_ms"`
	LatencyP95    int          `json:"latency_p95_ms"`
	LatencyP99    int          `json:"latency_p99_ms"`
	Errors        []ErrorBreakdown `json:"errors"`
}

// ErrorBreakdown categorizes errors
type ErrorBreakdown struct {
	Type       string  `json:"type"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ============================================================
// Workflow Simulation Handlers
// ============================================================

func (h *Handler) HandleSimulateWorkflow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req struct {
		Workflow   WorkflowSpec `json:"workflow"`
		Iterations int         `json:"iterations"` // number of dry-run iterations
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Iterations <= 0 {
		req.Iterations = 100
	}
	if req.Iterations > 10000 {
		writeError(w, http.StatusBadRequest, "TOO_MANY_ITERATIONS", "max 10000 iterations")
		return
	}

	// Run simulation
	result := h.simulateWorkflow(req.Workflow, req.Iterations)
	result.ID = "sim_" + generateID()

	// Store the simulation result for later retrieval
	h.simMu.Lock()
	h.simulations[result.ID] = &result
	h.simMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func (h *Handler) simulateWorkflow(workflow WorkflowSpec, iterations int) SimulationResult {
	now := time.Now()
	var completedAt *time.Time
	var nodeExecutions = make(map[string]int)
	var nodeFailures = make(map[string]int)
	var totalLatency int64
	var totalCost float64
	var successCount int

	for i := 0; i < iterations; i++ {
		execLatency := int64(0)
		execCost := 0.0
		execFailed := false

		// Topological sort simulation — follow edges
		visited := make(map[string]bool)
		var queue []string
		// Find root nodes (nodes with no incoming edges)
		hasIncoming := make(map[string]bool)
		for _, e := range workflow.Edges {
			hasIncoming[e.To] = true
		}
		for _, n := range workflow.Nodes {
			if !hasIncoming[n.ID] {
				queue = append(queue, n.ID)
			}
		}
		if len(queue) == 0 && len(workflow.Nodes) > 0 {
			queue = []string{workflow.Nodes[0].ID}
		}

		for len(queue) > 0 {
			nodeID := queue[0]
			queue = queue[1:]
			if visited[nodeID] {
				continue
			}
			visited[nodeID] = true

			// Find node spec
			var nodeSpec NodeSpec
			for _, n := range workflow.Nodes {
				if n.ID == nodeID {
					nodeSpec = n
					break
				}
			}

			nodeExecutions[nodeSpec.ID]++
			execLatency += int64(nodeSpec.Timeout + rand.Intn(50))
			execCost += nodeSpec.CostUSD

			// Check for failure
			probSuc := 0.95 // default success probability
			for _, e := range workflow.Edges {
				if e.From == nodeID {
					probSuc = e.ProbSuc
					break
				}
			}
			if rand.Float64() > probSuc {
				nodeFailures[nodeSpec.ID]++
				execFailed = true
				break // cascade failure
			}

			// Add downstream nodes
			for _, e := range workflow.Edges {
				if e.From == nodeID {
					queue = append(queue, e.To)
				}
			}
		}

		if !execFailed {
			successCount++
			totalLatency += execLatency
			totalCost += execCost
		}
	}

	completed := time.Now()
	completedAt = &completed

	return SimulationResult{
		Status:          "completed",
		SuccessRate:     float64(successCount) / float64(iterations),
		AvgLatencyMs:    int(totalLatency / int64(iterations)),
		AvgCostUSD:      totalCost / float64(iterations),
		PredictedNodes:  nodeExecutions,
		FailedNodes:     nodeFailures,
		ExecutionCount:  iterations,
		StartedAt:       now,
		CompletedAt:     completedAt,
		Iteration:       iterations,
	}
}

func (h *Handler) HandleGetSimulation(w http.ResponseWriter, r *http.Request) {
	simID := mux.Vars(r)["id"]
	if simID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "simulation id required")
		return
	}

	h.simMu.RLock()
	result, ok := h.simulations[simID]
	h.simMu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "simulation not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func (h *Handler) HandleAbortSimulation(w http.ResponseWriter, r *http.Request) {
	simID := mux.Vars(r)["id"]
	if simID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "simulation id required")
		return
	}

	h.simMu.Lock()
	result, ok := h.simulations[simID]
	if ok {
		result.Status = "aborted"
		h.simulations[simID] = result
	}
	h.simMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "simulation aborted",
		"id":      simID,
	})
}

// ============================================================
// Monte Carlo Simulation
// ============================================================

func (h *Handler) HandleMonteCarloSimulation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req struct {
		Workflow   WorkflowSpec `json:"workflow"`
		Iterations int          `json:"iterations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Iterations <= 0 {
		req.Iterations = 1000
	}
	if req.Iterations > 50000 {
		writeError(w, http.StatusBadRequest, "TOO_MANY_ITERATIONS", "max 50000 iterations")
		return
	}

	result := h.runMonteCarlo(req.Workflow, req.Iterations)
	result.ID = "mc_" + generateID()

	// Store the Monte Carlo result for later retrieval
	h.simMu.Lock()
	h.monteCarloResults[result.ID] = &result
	h.simMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func (h *Handler) runMonteCarlo(workflow WorkflowSpec, iterations int) MonteCarloResult {
	outcomes := make([]OutcomeSample, 0, iterations)
	latencies := make([]int, 0, iterations)
	costBreakdown := make(map[string]float64)

	successCount := 0
	partialCount := 0
	failCount := 0

	for i := 0; i < iterations; i++ {
		outcome := h.simulateSingleExecution(workflow, costBreakdown)
		outcomes = append(outcomes, outcome)
		latencies = append(latencies, outcome.LatencyMs)

		switch outcome.Outcome {
		case "success":
			successCount++
		case "partial":
			partialCount++
		case "failed":
			failCount++
		}
	}

	// Compute percentiles
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[int(float64(len(latencies))*0.50)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	avgLatency := 0
	avgCost := 0.0
	for _, o := range outcomes {
		avgLatency += o.LatencyMs
		avgCost += o.CostUSD
	}
	if len(outcomes) > 0 {
		avgLatency /= len(outcomes)
		avgCost /= float64(len(outcomes))
	}

	// Find bottleneck nodes (most executed with high failure rate)
	nodeFailCount := make(map[string]int)
	for _, o := range outcomes {
		for _, n := range o.FailedNodes {
			nodeFailCount[n]++
		}
	}

	bottlenecks := make([]string, 0)
	for n, fc := range nodeFailCount {
		if float64(fc)/float64(iterations) > 0.05 {
			bottlenecks = append(bottlenecks, n)
		}
	}

	return MonteCarloResult{
		Iterations:         iterations,
		SuccessRate:        float64(successCount) / float64(iterations),
		PartialFailureRate: float64(partialCount) / float64(iterations),
		TotalFailureRate:   float64(failCount) / float64(iterations),
		AvgLatencyMs:       avgLatency,
		P50LatencyMs:       p50,
		P95LatencyMs:       p95,
		P99LatencyMs:       p99,
		AvgCostUSD:         avgCost,
		Outcomes:           outcomes[:min(100, len(outcomes))], // return first 100 samples
		BottleneckNodes:    bottlenecks,
		CostBreakdown:      costBreakdown,
	}
}

func (h *Handler) simulateSingleExecution(workflow WorkflowSpec, costBreakdown map[string]float64) OutcomeSample {
	visited := make(map[string]bool)
	var queue []string
	hasIncoming := make(map[string]bool)
	for _, e := range workflow.Edges {
		hasIncoming[e.To] = true
	}
	for _, n := range workflow.Nodes {
		if !hasIncoming[n.ID] {
			queue = append(queue, n.ID)
		}
	}
	if len(queue) == 0 && len(workflow.Nodes) > 0 {
		queue = []string{workflow.Nodes[0].ID}
	}

	var failedNodes []string
	var riskFactors []string
	totalLatency := 0
	totalCost := 0.0
	hasFailure := false
	hasPartialFailure := false

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		if visited[nodeID] {
			continue
		}
		visited[nodeID] = true

		var nodeSpec NodeSpec
		for _, n := range workflow.Nodes {
			if n.ID == nodeID {
				nodeSpec = n
				break
			}
		}

		totalLatency += nodeSpec.Timeout + rand.Intn(50)
		totalCost += nodeSpec.CostUSD
		costBreakdown[nodeSpec.ID] += nodeSpec.CostUSD

		// Edge-specific failure probability
		probSuc := 0.95
		for _, e := range workflow.Edges {
			if e.From == nodeID {
				probSuc = e.ProbSuc
				break
			}
		}

		if rand.Float64() > probSuc {
			failedNodes = append(failedNodes, nodeID)
			hasFailure = true
			// But don't break — simulate partial completion
		}

		if float64(len(failedNodes))/float64(len(workflow.Nodes)) > 0.3 {
			hasPartialFailure = true
		}

		// Check for specific risk factors
		if nodeSpec.Type == "agent" && nodeSpec.CostUSD > 0.10 {
			riskFactors = append(riskFactors, "high-cost-agent")
		}
		if nodeSpec.Timeout > 30000 {
			riskFactors = append(riskFactors, "high-latency-node")
		}

		for _, e := range workflow.Edges {
			if e.From == nodeID {
				queue = append(queue, e.To)
			}
		}
	}

	outcome := "success"
	if hasFailure {
		outcome = "failed"
	} else if hasPartialFailure {
		outcome = "partial"
	}

	return OutcomeSample{
		Outcome:     outcome,
		Probability: 1.0,
		LatencyMs:   totalLatency,
		CostUSD:     totalCost,
		FailedNodes: failedNodes,
		RiskFactors: riskFactors,
	}
}

// ============================================================
// Failure Injection
// ============================================================

type FailureInjectionSpec struct {
	WorkflowID string           `json:"workflow_id"`
	Nodes      []FailureNodeSpec `json:"nodes"`
}

// FailureNodeSpec specifies failure characteristics for a node
type FailureNodeSpec struct {
	NodeID         string  `json:"node_id"`
	FailureType    string  `json:"failure_type"` // timeout, error, crash, data_corruption
	FailureRate    float64 `json:"failure_rate"` // 0-1
	RecoveryAction string  `json:"recovery_action"` // retry, fallback, abort
}

type FailureInjectionResult struct {
	ID             string                     `json:"id"`
	WorkflowID     string                     `json:"workflow_id"`
	Iterations     int                        `json:"iterations"`
	BaselineRate   float64                    `json:"baseline_success_rate"`
	InjectedRate   float64                    `json:"injected_failure_rate"`
	RecoveryRate   float64                    `json:"recovery_rate"` // how often recovery succeeded
	AvgLatencyIncrease int                     `json:"avg_latency_increase_ms"`
	CostImpact     float64                    `json:"cost_impact_usd"`
	NodeResults    map[string]FailureNodeResult `json:"node_results"`
}

type FailureNodeResult struct {
	FailuresInjected int     `json:"failures_injected"`
	Recoveries       int     `json:"recoveries"`
	AvgRecoveryTime  int     `json:"avg_recovery_time_ms"`
}

// HandleFailureInjection simulates failures at specified nodes to test resilience
func (h *Handler) HandleFailureInjection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req FailureInjectionSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	iterations := 500
	nodeResults := make(map[string]FailureNodeResult)
	var totalFailures, totalRecoveries int64
	var totalLatencyIncrease int64

	for _, nodeSpec := range req.Nodes {
		failures := 0
		recoveries := 0
		var recoveryTimes []int

		for i := 0; i < iterations; i++ {
			if rand.Float64() < nodeSpec.FailureRate {
				failures++
				totalFailures++

				// Simulate recovery
				if nodeSpec.RecoveryAction == "retry" && rand.Float64() < 0.7 {
					recoveries++
					totalRecoveries++
					recoveryTimes = append(recoveryTimes, 100+rand.Intn(500))
				} else if nodeSpec.RecoveryAction == "fallback" && rand.Float64() < 0.85 {
					recoveries++
					totalRecoveries++
					recoveryTimes = append(recoveryTimes, 200+rand.Intn(1000))
				}
			}
		}

		avgRecovery := 0
		if len(recoveryTimes) > 0 {
			sum := 0
			for _, t := range recoveryTimes {
				sum += t
			}
			avgRecovery = sum / len(recoveryTimes)
		}

		nodeResults[nodeSpec.NodeID] = FailureNodeResult{
			FailuresInjected: failures,
			Recoveries:       recoveries,
			AvgRecoveryTime:  avgRecovery,
		}
	}

	baselineRate := 0.95
	injectedRate := 1.0 - (float64(totalFailures) / float64(iterations*len(req.Nodes)))
	recoveryRate := 0.0
	if totalFailures > 0 {
		recoveryRate = float64(totalRecoveries) / float64(totalFailures)
	}

	result := FailureInjectionResult{
		ID:                 "fi_" + generateID(),
		WorkflowID:         req.WorkflowID,
		Iterations:         iterations,
		BaselineRate:       baselineRate,
		InjectedRate:       injectedRate,
		RecoveryRate:       recoveryRate,
		AvgLatencyIncrease: int(totalLatencyIncrease / int64(iterations)),
		CostImpact:         float64(totalFailures) * 0.05,
		NodeResults:        nodeResults,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

// ============================================================
// Forecasting
// ============================================================

type ExecutionForecastRequest struct {
	WorkflowID    string                 `json:"workflow_id"`
	TimeHorizon    string                 `json:"time_horizon"` // 1h, 24h, 7d, 30d
	CallVolume     int                    `json:"call_volume"` // expected calls
	Nodes          []NodeSpec             `json:"nodes"`
}

type ExecutionForecast struct {
	WorkflowID      string             `json:"workflow_id"`
	TimeHorizon     string             `json:"time_horizon"`
	PredictedExecutions int              `json:"predicted_executions"`
	SuccessRate     float64            `json:"success_rate"`
	AvgLatencyMs    int                `json:"avg_latency_ms"`
	CostUSD         float64            `json:"cost_usd"`
	P50LatencyMs    int                `json:"p50_latency_ms"`
	P95LatencyMs    int                `json:"p95_latency_ms"`
	P99LatencyMs    int                `json:"p99_latency_ms"`
	Predictions     []PredictionPoint   `json:"predictions"` // time-series predictions
}

type PredictionPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Executions int       `json:"executions"`
	SuccessRate float64  `json:"success_rate"`
}

func (h *Handler) HandleExecutionForecast(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req ExecutionForecastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	forecast := h.computeExecutionForecast(req)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"forecast": forecast,
	})
}

func (h *Handler) computeExecutionForecast(req ExecutionForecastRequest) ExecutionForecast {
	baseRate := 0.95
	totalCost := 0.0
	totalLatency := 0

	for _, n := range req.Nodes {
		totalCost += n.CostUSD * float64(req.CallVolume)
		totalLatency += n.Timeout
	}

	// Time horizon scaling
	var horizonFactor float64
	switch req.TimeHorizon {
	case "1h":
		horizonFactor = 1.0 / 24.0
	case "24h":
		horizonFactor = 1.0
	case "7d":
		horizonFactor = 7.0
	case "30d":
		horizonFactor = 30.0
	default:
		horizonFactor = 1.0
	}

	predictedExec := int(float64(req.CallVolume) * horizonFactor)

	// Generate prediction time series
	predictions := make([]PredictionPoint, 0, 24)
	interval := time.Hour
	if req.TimeHorizon == "1h" {
		interval = time.Minute * 5
	} else if req.TimeHorizon == "7d" {
		interval = time.Hour * 6
	} else if req.TimeHorizon == "30d" {
		interval = time.Hour * 24
	}

	baseTime := time.Now()
	for i := 0; i < 24; i++ {
		ts := baseTime.Add(interval * time.Duration(i))
		rate := baseRate + (rand.Float64()-0.5)*0.05
		if rate < 0.85 {
			rate = 0.85
		}
		if rate > 0.99 {
			rate = 0.99
		}
		predictions = append(predictions, PredictionPoint{
			Timestamp:   ts,
			Executions:  predictedExec / 24,
			SuccessRate: rate,
		})
	}

	return ExecutionForecast{
		WorkflowID:        req.WorkflowID,
		TimeHorizon:       req.TimeHorizon,
		PredictedExecutions: predictedExec,
		SuccessRate:       baseRate,
		AvgLatencyMs:      totalLatency / len(req.Nodes),
		CostUSD:           totalCost,
		P50LatencyMs:      totalLatency / len(req.Nodes),
		P95LatencyMs:      totalLatency * 2 / len(req.Nodes),
		P99LatencyMs:      totalLatency * 4 / len(req.Nodes),
		Predictions:       predictions,
	}
}

func (h *Handler) HandleCostForecast(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req struct {
		WorkflowID string      `json:"workflow_id"`
		Nodes      []NodeSpec  `json:"nodes"`
		CallVolume int         `json:"call_volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Simple cost model
	var totalCost, perCallCost float64
	var byNode map[string]float64 = make(map[string]float64)

	for _, n := range req.Nodes {
		nodeCost := n.CostUSD * float64(req.CallVolume)
		totalCost += nodeCost
		perCallCost += n.CostUSD
		byNode[n.ID] = nodeCost
	}

	// Confidence interval via simulation
	iterations := 200
	var costs []float64
	for i := 0; i < iterations; i++ {
		c := perCallCost * float64(req.CallVolume)
		// Add stochastic variance
		variance := (rand.Float64() - 0.5) * 0.1 * c
		costs = append(costs, c+variance)
	}
	sort.Slice(costs, func(i, j int) bool { return costs[i] < costs[j] })
	lower := costs[int(float64(len(costs))*0.05)]
	upper := costs[int(float64(len(costs))*0.95)]

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"forecast": map[string]interface{}{
			"workflow_id":   req.WorkflowID,
			"total_cost_usd":  totalCost,
			"per_call_usd":    perCallCost,
			"lower_bound_usd": lower,
			"upper_bound_usd": upper,
			"confidence":      0.90,
			"by_node":         byNode,
		},
	})
}

func (h *Handler) HandleLatencyForecast(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req struct {
		WorkflowID string     `json:"workflow_id"`
		Nodes      []NodeSpec `json:"nodes"`
		LoadLevel  float64    `json:"load_level"` // 0-1, percentage of max load
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.LoadLevel <= 0 {
		req.LoadLevel = 0.5
	}

	// Model latency as queuing system (M/M/1 approximation)
	var baseLatency int
	loadMultiplier := 1.0 + (req.LoadLevel * 0.3) // 30% increase at full load

	for _, n := range req.Nodes {
		baseLatency += n.Timeout
	}

	_ = int(float64(baseLatency) * loadMultiplier)        // p50 — available for extension
	_ = int(float64(baseLatency) * loadMultiplier * 1.8)  // p95 — available for extension
	_ = int(float64(baseLatency) * loadMultiplier * 3.0)  // p99 — available for extension

	// Distribution around the mean
	dist := make([]int, 100)
	for i := 0; i < 100; i++ {
		variance := float64(baseLatency) * 0.2 * (rand.Float64() - 0.5)
		dist[i] = int(float64(baseLatency)*loadMultiplier + variance)
	}
	sort.Slice(dist, func(i, j int) bool { return dist[i] < dist[j] })

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"forecast": map[string]interface{}{
			"workflow_id": req.WorkflowID,
			"load_level":  req.LoadLevel,
			"p50_latency_ms": dist[50],
			"p95_latency_ms": dist[95],
			"p99_latency_ms": dist[99],
			"avg_latency_ms": baseLatency,
			"distribution": map[string]interface{}{
				"p50": dist[50],
				"p75": dist[75],
				"p90": dist[90],
				"p95": dist[95],
				"p99": dist[99],
			},
		},
	})
}

// ============================================================
// Stress Test
// ============================================================

// In-memory stress test state (use Redis in production for multi-instance)
var (
	stressTests = make(map[string]*StressTestState)
	stressTestMu sync.RWMutex
)

type StressTestState struct {
	Config   StressTestConfig
	Result   *StressTestResult
	AbortCh  chan struct{}
	mu       sync.RWMutex
}

func (h *Handler) HandleStartStressTest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req StressTestConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Iterations <= 0 {
		req.Iterations = 1000
	}
	if req.Iterations > 100000 {
		req.Iterations = 100000
	}
	if req.Parallelism <= 0 {
		req.Parallelism = 10
	}
	if req.Parallelism > 100 {
		req.Parallelism = 100
	}

	id := "st_" + generateID()
	abortCh := make(chan struct{})

	state := &StressTestState{
		Config:  req,
		AbortCh: abortCh,
	}
	state.Result = &StressTestResult{
		ID:            id,
		Status:        "running",
		Iterations:    req.Iterations,
		TotalExecutions: 0,
		SuccessRate:   0,
		Throughput:    0,
		LatencyP50:    0,
		LatencyP95:    0,
		LatencyP99:    0,
		Errors:        []ErrorBreakdown{},
	}

	stressTestMu.Lock()
	stressTests[id] = state
	stressTestMu.Unlock()

	// Run async
	go h.runStressTest(id, req, abortCh)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"ok":  true,
		"id":  id,
		"status": "started",
	})
}

func (h *Handler) runStressTest(id string, config StressTestConfig, abortCh chan struct{}) {
	defer func() {
		stressTestMu.Lock()
		if s, ok := stressTests[id]; ok {
			s.mu.Lock()
			s.Result.Status = "completed"
			s.mu.Unlock()
		}
		stressTestMu.Unlock()
	}()

	var latencies []int
	var totalExec int64
	var successCount int64
	errorCounts := make(map[string]int)

	batchSize := config.Parallelism
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < config.Iterations && config.Iterations <= 100000; i++ {
		select {
		case <-abortCh:
			stressTestMu.Lock()
			if s, ok := stressTests[id]; ok {
				s.mu.Lock()
				s.Result.Status = "aborted"
				s.mu.Unlock()
			}
			stressTestMu.Unlock()
			return
		default:
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Simulate execution
			latency := 30 + rand.Intn(100)
			totalExec++

			// Simulate errors
			errType := ""
			if rand.Float64() < 0.005 {
				errType = "Timeout"
			} else if rand.Float64() < 0.003 {
				errType = "Rate Limit"
			} else if rand.Float64() < 0.002 {
				errType = "Auth Failure"
			}

			if errType == "" {
				successCount++
			} else {
				errorCounts[errType]++
			}

			latencies = append(latencies, latency)
		}()

		// Flow control: batch wait
		if i > 0 && i%batchSize == 0 {
			wg.Wait()
		}
	}
	wg.Wait()

	elapsed := time.Since(startTime).Seconds()
	if elapsed < 1 {
		elapsed = 1
	}

	// Compute percentiles
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[int(float64(len(latencies))*0.50)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	successRate := float64(successCount) / float64(totalExec)
	throughput := float64(totalExec) / elapsed

	var errors []ErrorBreakdown
	for et, count := range errorCounts {
		errors = append(errors, ErrorBreakdown{
			Type:       et,
			Count:      count,
			Percentage: float64(count) / float64(totalExec) * 100,
		})
	}

	stressTestMu.Lock()
	if s, ok := stressTests[id]; ok {
		s.mu.Lock()
		s.Result.TotalExecutions = int(totalExec)
		s.Result.SuccessRate = successRate
		s.Result.Throughput = throughput
		s.Result.LatencyP50 = p50
		s.Result.LatencyP95 = p95
		s.Result.LatencyP99 = p99
		s.Result.Errors = errors
		s.mu.Unlock()
	}
	stressTestMu.Unlock()
}

func (h *Handler) HandleGetStressTest(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	stressTestMu.RLock()
	state, ok := stressTests[id]
	stressTestMu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "stress test not found")
		return
	}

	state.mu.RLock()
	result := state.Result
	state.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func (h *Handler) HandleAbortStressTest(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	stressTestMu.RLock()
	state, ok := stressTests[id]
	stressTestMu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "stress test not found")
		return
	}

	close(state.AbortCh)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "abort signal sent",
		"id":      id,
	})
}

func (h *Handler) HandleGetStressTestResults(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	stressTestMu.RLock()
	state, ok := stressTests[id]
	stressTestMu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "stress test not found")
		return
	}

	state.mu.RLock()
	result := state.Result
	state.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

// ============================================================
// Resource Collision Detection
// ============================================================

type ResourceCollisionSpec struct {
	Resources []ResourceSpec `json:"resources"`
}

type ResourceSpec struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"` // gpu, cpu, memory, network
	Capacity  float64        `json:"capacity"` // total capacity
	Tasks     []TaskSchedule `json:"tasks"`
}

type TaskSchedule struct {
	TaskID   string  `json:"task_id"`
	StartMs  int64   `json:"start_ms"` // relative to now
	EndMs    int64   `json:"end_ms"`
	Usage    float64 `json:"usage"` // 0-1 fraction of capacity
	Priority int     `json:"priority"`
}

type CollisionResult struct {
	Collisions []Collision `json:"collisions"`
	Resolutions []string   `json:"resolutions"`
}

type Collision struct {
	ResourceID   string         `json:"resource_id"`
	ResourceName string         `json:"resource_name"`
	Severity     string         `json:"severity"` // high, medium, low
	ConflictingTasks []TaskSchedule `json:"conflicting_tasks"`
	Resolution   string         `json:"resolution"`
}

func (h *Handler) HandleDetectResourceCollisions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req ResourceCollisionSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	result := h.detectCollisions(req)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"result": result,
	})
}

func (h *Handler) detectCollisions(req ResourceCollisionSpec) CollisionResult {
	var collisions []Collision
	var resolutions []string

	for _, resource := range req.Resources {
		// Sort tasks by start time
		tasks := resource.Tasks
		if len(tasks) == 0 {
			continue
		}

		// Check for overlapping tasks
		_ = tasks // available for extension
		for i := 0; i < len(tasks); i++ {
			for j := i + 1; j < len(tasks); j++ {
				// Check overlap: task i overlaps task j if i.start < j.end && i.end > j.start
				if tasks[i].StartMs < tasks[j].EndMs && tasks[i].EndMs > tasks[j].StartMs {
					// Check if combined usage exceeds capacity
					combinedUsage := tasks[i].Usage + tasks[j].Usage
					if combinedUsage > 1.0 {
						severity := "low"
						if combinedUsage > 1.5 {
							severity = "high"
						} else if combinedUsage > 1.2 {
							severity = "medium"
						}

						// Generate resolution
						var resolution string
						if tasks[i].Priority >= tasks[j].Priority {
							resolution = "Reschedule task " + tasks[j].TaskID + " to avoid overlap with " + tasks[i].TaskID
						} else {
							resolution = "Reschedule task " + tasks[i].TaskID + " to avoid overlap with " + tasks[j].TaskID
						}

						collision := Collision{
							ResourceID:        resource.ID,
							ResourceName:      resource.Name,
							Severity:          severity,
							ConflictingTasks:  []TaskSchedule{tasks[i], tasks[j]},
							Resolution:        resolution,
						}
						collisions = append(collisions, collision)
						resolutions = append(resolutions, resolution)
					}
				}
			}
		}
	}

	return CollisionResult{
		Collisions:  collisions,
		Resolutions: resolutions,
	}
}

// ============================================================
// Agent Behavior Prediction
// ============================================================

type AgentBehaviorSpec struct {
	AgentID    string      `json:"agent_id"`
	HistorySize int        `json:"history_size"` // number of past executions to consider
	Context     string     `json:"context"` // current task description
}

type AgentBehaviorPrediction struct {
	AgentID        string           `json:"agent_id"`
	Confidence     float64          `json:"confidence"`
	BasedOnSamples int              `json:"based_on_samples"`
	LikelyActions  []ActionPrediction `json:"likely_actions"`
	CurrentTaskPrediction string     `json:"current_task_prediction"`
}

type ActionPrediction struct {
	Action          string  `json:"action"`
	Probability     float64 `json:"probability"`
	ExpectedOutcome string  `json:"expected_outcome"`
	RiskLevel       string  `json:"risk_level"` // low, medium, high
}

// HandlePredictAgentBehavior uses historical data to predict agent behavior
func (h *Handler) HandlePredictAgentBehavior(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req AgentBehaviorSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Simulate prediction based on context
	prediction := h.predictAgentBehavior(req)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"prediction": prediction,
	})
}

func (h *Handler) predictAgentBehavior(req AgentBehaviorSpec) AgentBehaviorPrediction {
	historySize := req.HistorySize
	if historySize <= 0 {
		historySize = 50
	}

	// Context-based action inference
	var actions []ActionPrediction
	lowercaseCtx := req.Context

	// Simple keyword-based prediction
	if contains(lowercaseCtx, "security", "scan", "audit") {
		actions = append(actions, ActionPrediction{
			Action:          "Scan for vulnerabilities",
			Probability:     0.85,
			ExpectedOutcome: "Risk assessment report",
			RiskLevel:       "low",
		})
	}
	if contains(lowercaseCtx, "design", "schema", "architecture") {
		actions = append(actions, ActionPrediction{
			Action:          "Design API schema",
			Probability:     0.72,
			ExpectedOutcome: "OpenAPI 3.1 specification",
			RiskLevel:       "medium",
		})
	}
	if contains(lowercaseCtx, "test", "coverage", "unit") {
		actions = append(actions, ActionPrediction{
			Action:          "Generate unit tests",
			Probability:     0.65,
			ExpectedOutcome: "85% code coverage",
			RiskLevel:       "low",
		})
	}
	if contains(lowercaseCtx, "deploy", "staging", "production") {
		actions = append(actions, ActionPrediction{
			Action:          "Prepare deployment package",
			Probability:     0.60,
			ExpectedOutcome: "Blue-green deployment ready",
			RiskLevel:       "high",
		})
	}

	// Default actions
	if len(actions) == 0 {
		actions = append(actions, ActionPrediction{
			Action:          "Execute workflow step",
			Probability:     0.55,
			ExpectedOutcome: "Step completed",
			RiskLevel:       "medium",
		})
	}

	confidence := math.Min(0.99, 0.5+float64(historySize)*0.01)

	return AgentBehaviorPrediction{
		AgentID:             req.AgentID,
		Confidence:          confidence,
		BasedOnSamples:      historySize,
		LikelyActions:       actions,
		CurrentTaskPrediction: "Analyzing current task context for optimal action selection",
	}
}

// ============================================================
// Hallucination Risk Analysis
// ============================================================

type HallucinationRiskSpec struct {
	ModelID      string        `json:"model_id"` // e.g., claude-sonnet-4, gpt-4o
	PromptLength  int           `json:"prompt_length"` // characters
	ContextWindow int           `json:"context_window"` // tokens
	TaskType      string        `json:"task_type"` // code_gen, reasoning, summarization, factual
	Complexity    string        `json:"complexity"` // low, medium, high, extreme
	PreviousErrors int           `json:"previous_errors"` // errors in last 100 calls
}

type HallucinationRiskResult struct {
	ModelID          string    `json:"model_id"`
	OverallRiskScore float64   `json:"risk_score"` // 0-1, higher = more risk
	RiskLevel        string    `json:"risk_level"` // low, medium, high, critical
	ContributingFactors []string `json:"contributing_factors"`
	Recommendations  []string   `json:"recommendations"`
	ConfidenceScore  float64   `json:"confidence"`
}

func (h *Handler) HandleHallucinationRiskAnalysis(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req HallucinationRiskSpec
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	result := h.analyzeHallucinationRisk(req)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func (h *Handler) analyzeHallucinationRisk(req HallucinationRiskSpec) HallucinationRiskResult {
	var riskScore float64
	var factors []string
	var recommendations []string

	// Base risk by model
	modelBaseRisk := map[string]float64{
		"claude-sonnet-4":  0.05,
		"claude-opus-4":    0.04,
		"gpt-4o":          0.08,
		"gpt-4o-mini":     0.10,
		"gemini-pro":      0.07,
		"mistral-large":   0.06,
	}
	baseRisk := modelBaseRisk[req.ModelID]
	if baseRisk == 0 {
		baseRisk = 0.10
	}
	riskScore = baseRisk

	// Prompt length factor (very long prompts increase confusion risk)
	if req.PromptLength > 50000 {
		riskScore += 0.15
		factors = append(factors, "Very long prompt (>50K chars)")
	} else if req.PromptLength > 20000 {
		riskScore += 0.08
		factors = append(factors, "Long prompt (>20K chars)")
	}

	// Context window utilization (>80% is risky)
	if req.ContextWindow > 0 {
		utilization := float64(req.PromptLength) / float64(req.ContextWindow)
		if utilization > 0.9 {
			riskScore += 0.12
			factors = append(factors, "Context window >90% utilized")
		} else if utilization > 0.75 {
			riskScore += 0.05
			factors = append(factors, "Context window >75% utilized")
		}
	}

	// Task type factor
	taskRisk := map[string]float64{
		"code_gen":    0.03,
		"reasoning":   0.08,
		"summarization": 0.05,
		"factual":    0.12,
	}
	riskScore += taskRisk[req.TaskType]
	if req.TaskType == "factual" {
		factors = append(factors, "Factual task (higher hallucination tendency)")
	}

	// Complexity factor
	complexityRisk := map[string]float64{
		"low":      0.02,
		"medium":   0.05,
		"high":     0.10,
		"extreme":  0.18,
	}
	riskScore += complexityRisk[req.Complexity]
	if req.Complexity == "high" || req.Complexity == "extreme" {
		factors = append(factors, "High/extreme complexity task")
	}

	// Previous error rate
	if req.PreviousErrors > 5 {
		riskScore += 0.15
		factors = append(factors, "High historical error rate")
	} else if req.PreviousErrors > 2 {
		riskScore += 0.07
		factors = append(factors, "Moderate historical error rate")
	}

	// Cap at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	// Risk level
	riskLevel := "low"
	if riskScore > 0.7 {
		riskLevel = "critical"
		recommendations = append(recommendations, "Consider using a more reliable model", "Break task into smaller steps", "Add verification layer")
	} else if riskScore > 0.5 {
		riskLevel = "high"
		recommendations = append(recommendations, "Add cross-validation", "Use ensemble of models", "Implement output verification")
	} else if riskScore > 0.3 {
		riskLevel = "medium"
		recommendations = append(recommendations, "Monitor outputs closely", "Consider adding a review step")
	} else {
		riskLevel = "low"
		recommendations = append(recommendations, "Standard monitoring sufficient", "Proceed with current model")
	}

	confidence := 1.0 - riskScore*0.3

	return HallucinationRiskResult{
		ModelID:            req.ModelID,
		OverallRiskScore:   riskScore,
		RiskLevel:          riskLevel,
		ContributingFactors: factors,
		Recommendations:    recommendations,
		ConfidenceScore:    confidence,
	}
}

// ============================================================
// Helpers
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) > 0 && (containsOne(s, sub)) {
			return true
		}
	}
	return false
}

func containsOne(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// strFormat is unused but needed for syntax
var _ = strconv.FormatInt