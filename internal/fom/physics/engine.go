package physics

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Engine struct {
	functionStats map[string]*FunctionPhysics
	cache         *redis.Client
	db            *sql.DB
	mu            sync.RWMutex
}

type FunctionPhysics struct {
	Name           string   `json:"name"`
	Cost           float64  `json:"cost"`
	LatencyMs      int      `json:"latency_ms"`
	LatencyP99Ms   int      `json:"latency_p99_ms"`
	SuccessRate    float64  `json:"success_rate"`
	Dependencies   []string `json:"dependencies"`
	Parallelizable bool     `json:"parallelizable"`
}

type WorkflowPrediction struct {
	EstimatedCost      float64 `json:"estimated_cost"`
	EstimatedTimeMs    int     `json:"estimated_time_ms"`
	SuccessProbability float64 `json:"success_probability"`
	Confidence         float64 `json:"confidence"`
}

type SimulationResult struct {
	SuccessRate float64   `json:"success_rate"`
	AvgCost     float64   `json:"avg_cost"`
	P50Cost     float64   `json:"p50_cost"`
	P95Cost     float64   `json:"p95_cost"`
	AvgTimeMs   int       `json:"avg_time_ms"`
	P50TimeMs   int       `json:"p50_time_ms"`
	P95TimeMs   int       `json:"p95_time_ms"`
	Confidence  float64   `json:"confidence"`
	CostDist    []float64 `json:"cost_distribution,omitempty"`
	TimeDist    []int     `json:"time_distribution,omitempty"`
}

func NewEngine(db *sql.DB, cache *redis.Client) *Engine {
	return &Engine{
		functionStats: make(map[string]*FunctionPhysics),
		cache:         cache,
		db:            db,
	}
}

func (e *Engine) LoadFunctionStats(ctx context.Context) error {
	rows, err := e.db.QueryContext(ctx, `
		SELECT function_name, avg_cost, avg_time_ms, success_rate, p99_time_ms, dependencies
		FROM fom_function_stats
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	e.mu.Lock()
	defer e.mu.Unlock()

	for rows.Next() {
		stats := &FunctionPhysics{}
		var dependencies []byte
		err := rows.Scan(&stats.Name, &stats.Cost, &stats.LatencyMs, &stats.SuccessRate, &stats.LatencyP99Ms, &dependencies)
		if err != nil {
			continue
		}
		if len(dependencies) > 0 {
			_ = json.Unmarshal(dependencies, &stats.Dependencies)
		}
		e.functionStats[stats.Name] = stats
	}

	return rows.Err()
}

func (e *Engine) RefreshFromExecutionData(ctx context.Context) error {
	rows, err := e.db.QueryContext(ctx, `
		SELECT
			function_name,
			AVG(actual_cost) as avg_cost,
			AVG(actual_time_ms) as avg_time_ms,
			PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY actual_time_ms) as p50,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY actual_time_ms) as p95,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY actual_time_ms) as p99,
			COUNT(*) as sample_count,
			SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate
		FROM fom_actions
		WHERE actual_cost IS NOT NULL
		GROUP BY function_name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	e.mu.Lock()
	defer e.mu.Unlock()

	for rows.Next() {
		var name string
		var avgCost, successRate float64
		var avgTimeMs, p50, p95, p99 int
		var sampleCount int

		err := rows.Scan(&name, &avgCost, &avgTimeMs, &p50, &p95, &p99, &sampleCount, &successRate)
		if err != nil {
			continue
		}

		existing := e.functionStats[name]
		if existing == nil {
			existing = &FunctionPhysics{Name: name}
		}

		existing.Cost = avgCost
		existing.LatencyMs = avgTimeMs
		existing.LatencyP99Ms = p99
		existing.SuccessRate = successRate
		e.functionStats[name] = existing
	}

	return rows.Err()
}

func (e *Engine) GetFunctionPhysics(name string) *FunctionPhysics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.functionStats[name]
}

func (e *Engine) SetFunctionPhysics(name string, physics *FunctionPhysics) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.functionStats[name] = physics
}

func (e *Engine) PredictWorkflowOutcome(workflow []string) (*WorkflowPrediction, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var totalCost float64
	var totalTimeMs int
	var successProb float64 = 1.0
	var confidenceSum float64
	var validCount int

	for i, fn := range workflow {
		physics := e.functionStats[fn]
		if physics == nil {
			confidenceSum += 0.3
			continue
		}

		validCount++
		totalTimeMs += physics.LatencyMs
		successProb *= physics.SuccessRate
		totalCost += physics.Cost

		for _, dep := range physics.Dependencies {
			if !e.dependencySatisfied(dep, workflow[:i]) {
				successProb *= 0.5
			}
		}
	}

	confidence := 0.5
	if validCount > 0 {
		confidence = confidenceSum / float64(validCount)
		if confidence < 0.3 {
			confidence = 0.3
		}
	}

	return &WorkflowPrediction{
		EstimatedCost:      totalCost,
		EstimatedTimeMs:    totalTimeMs,
		SuccessProbability: math.Max(0.01, successProb),
		Confidence:         confidence,
	}, nil
}

func (e *Engine) dependencySatisfied(dep string, executedFunctions []string) bool {
	for _, fn := range executedFunctions {
		if fn == dep {
			return true
		}
	}
	return false
}

func (e *Engine) SimulateWorkflow(workflow []string, iterations int) (*SimulationResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var successes int
	var totalCosts []float64
	var totalTimes []int

	for i := 0; i < iterations; i++ {
		cost, timeMs, success := e.simulateOnce(workflow)
		if success {
			successes++
		}
		totalCosts = append(totalCosts, cost)
		totalTimes = append(totalTimes, timeMs)
	}

	sort.Float64s(totalCosts)
	sort.Ints(totalTimes)

	return &SimulationResult{
		SuccessRate: float64(successes) / float64(iterations),
		AvgCost:     mean(totalCosts),
		P50Cost:     totalCosts[len(totalCosts)/2],
		P95Cost:     totalCosts[int(float64(len(totalCosts))*0.95)],
		AvgTimeMs:   meanInt(totalTimes),
		P50TimeMs:   totalTimes[len(totalTimes)/2],
		P95TimeMs:   totalTimes[int(float64(len(totalTimes))*0.95)],
		Confidence:  1.0 - (1.0 / float64(iterations)),
	}, nil
}

func (e *Engine) simulateOnce(workflow []string) (float64, int, bool) {
	var totalCost float64
	var totalTimeMs int
	allSucceeded := true

	for _, fn := range workflow {
		physics := e.functionStats[fn]
		if physics == nil {
			totalCost += 0.01
			totalTimeMs += 500
			allSucceeded = false
			continue
		}

		totalCost += physics.Cost
		totalTimeMs += physics.LatencyMs

		if rand.Float64() > physics.SuccessRate {
			allSucceeded = false
		}
	}

	return totalCost, totalTimeMs, allSucceeded
}

func (e *Engine) SimulateWorkflowDetailed(workflow []string, iterations int) (*SimulationResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var successes int
	var totalCosts []float64
	var totalTimes []int

	for i := 0; i < iterations; i++ {
		cost, timeMs, success := e.simulateOnce(workflow)
		if success {
			successes++
		}
		totalCosts = append(totalCosts, cost)
		totalTimes = append(totalTimes, timeMs)
	}

	sort.Float64s(totalCosts)
	sort.Ints(totalTimes)

	result := &SimulationResult{
		SuccessRate: float64(successes) / float64(iterations),
		AvgCost:     mean(totalCosts),
		P50Cost:     totalCosts[len(totalCosts)/2],
		P95Cost:     totalCosts[int(float64(len(totalCosts))*0.95)],
		AvgTimeMs:   meanInt(totalTimes),
		P50TimeMs:   totalTimes[len(totalTimes)/2],
		P95TimeMs:   totalTimes[int(float64(len(totalTimes))*0.95)],
		Confidence:  1.0 - (1.0 / float64(iterations)),
	}

	if iterations <= 1000 {
		result.CostDist = totalCosts
		result.TimeDist = totalTimes
	}

	return result, nil
}

func (e *Engine) CompareWorkflows(workflowA, workflowB []string, iterations int) (bool, float64, error) {
	simA, err := e.SimulateWorkflow(workflowA, iterations)
	if err != nil {
		return false, 0, err
	}

	simB, err := e.SimulateWorkflow(workflowB, iterations)
	if err != nil {
		return false, 0, err
	}

	scoreA := simA.SuccessRate*0.4 + (1.0-simA.AvgCost)*0.3 + (1.0/float64(simA.AvgTimeMs))*0.3
	scoreB := simB.SuccessRate*0.4 + (1.0-simB.AvgCost)*0.3 + (1.0/float64(simB.AvgTimeMs))*0.3

	return scoreA >= scoreB, scoreA - scoreB, nil
}

func (e *Engine) GenerateWorkflowVariants(workflow []string, count int) [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	variants := make([][]string, 0, count)
	variants = append(variants, workflow)

	optionalSteps := e.getOptionalSteps()
	if len(optionalSteps) == 0 {
		for i := 1; i < count && i < len(workflow); i++ {
			variant := make([]string, len(workflow))
			copy(variant, workflow)
			variant = append(variant[:i], variant[i+1:]...)
			variants = append(variants, variant)
		}
		return variants
	}

	for len(variants) < count {
		variant := e.mutateWorkflow(workflow, optionalSteps)
		variants = append(variants, variant)
	}

	return variants
}

func (e *Engine) getOptionalSteps() []string {
	optional := []string{}
	for name, physics := range e.functionStats {
		if len(physics.Dependencies) == 0 {
			optional = append(optional, name)
		}
	}
	return optional
}

func (e *Engine) mutateWorkflow(workflow []string, optionalSteps []string) []string {
	if len(workflow) == 0 || len(optionalSteps) == 0 {
		return workflow
	}

	mutation := rand.Intn(4)
	variant := make([]string, len(workflow))
	copy(variant, workflow)

	switch mutation {
	case 0:
		return variant
	case 1:
		if len(variant) > 2 {
			return variant[:len(variant)-1]
		}
	case 2:
		step := optionalSteps[rand.Intn(len(optionalSteps))]
		return append(variant, step)
	case 3:
		if len(variant) > 1 {
			i, j := rand.Intn(len(variant)), rand.Intn(len(variant))
			variant[i], variant[j] = variant[j], variant[i]
			return variant
		}
	}

	return variant
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func meanInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	var sum int
	for _, v := range values {
		sum += v
	}
	return sum / len(values)
}

type WorkflowRanker struct {
	engine *Engine
}

func NewWorkflowRanker(engine *Engine) *WorkflowRanker {
	return &WorkflowRanker{engine: engine}
}

type RankedWorkflow struct {
	Workflow   []string  `json:"workflow"`
	Score      float64   `json:"score"`
	SuccessRate float64  `json:"success_rate"`
	AvgCost    float64   `json:"avg_cost"`
	AvgTimeMs  int       `json:"avg_time_ms"`
}

func (r *WorkflowRanker) RankWorkflows(workflows [][]string, iterations int) ([]*RankedWorkflow, error) {
	results := make([]*RankedWorkflow, 0, len(workflows))

	for _, workflow := range workflows {
		sim, err := r.engine.SimulateWorkflow(workflow, iterations)
		if err != nil {
			continue
		}

		reliability := sim.SuccessRate
		efficiency := 1.0
		if sim.AvgCost > 0 {
			efficiency = 1.0 - math.Min(1.0, sim.AvgCost/1.0)
		}
		speed := 1.0
		if sim.AvgTimeMs > 0 {
			speed = 1.0 - math.Min(1.0, float64(sim.AvgTimeMs)/60000.0)
		}

		score := reliability*0.4 + efficiency*0.3 + speed*0.3

		results = append(results, &RankedWorkflow{
			Workflow:    workflow,
			Score:       score,
			SuccessRate: sim.SuccessRate,
			AvgCost:     sim.AvgCost,
			AvgTimeMs:   sim.AvgTimeMs,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}