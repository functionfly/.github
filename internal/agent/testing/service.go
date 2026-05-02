package testing

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JSONBMap map[string]any

func (m JSONBMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func (m *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONBMap)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for JSONB")
	}
	if len(b) == 0 {
		*m = make(JSONBMap)
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = make(map[string]any)
	}
	*m = JSONBMap(out)
	return nil
}

const (
	StageSyntax      = "syntax"
	StageSecurity    = "security"
	StageUnit        = "unit"
	StageExecution   = "execution"
	StagePerformance = "performance"

	StatusPassed = "passed"
	StatusFailed = "failed"
)

type TestResult struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID uuid.UUID      `json:"function_id" gorm:"type:uuid;not null;index"`
	Stage      string         `json:"stage" gorm:"not null;index"`
	Passed     bool           `json:"passed" gorm:"not null"`
	Status     string         `json:"status" gorm:"not null"`
	Score      float64        `json:"score" gorm:"type:decimal(5,2);default:0"`
	DurationMs int            `json:"duration_ms"`
	Error      string         `json:"error,omitempty" gorm:"type:text"`
	Details    JSONBMap     `json:"details" gorm:"type:jsonb;default:'{}'"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (TestResult) TableName() string { return "factory_test_results" }

type SandboxExecutor interface {
	Execute(ctx context.Context, runtime, code string, input map[string]any) (SandboxExecutionResult, error)
}

type BenchmarkRunner interface {
	Run(ctx context.Context, runtime, code string) (BenchmarkResult, error)
}

type SandboxExecutionResult struct {
	Passed     bool
	DurationMs int
	Output     map[string]any
	Error      string
}

type BenchmarkResult struct {
	Passed        bool
	DurationMs    int
	ColdStartMs   int
	AverageExecMs int
	MemoryBytes   int64
	Error         string
}

type Service struct {
	db        *gorm.DB
	security  *SecurityScanner
	sandbox   SandboxExecutor
	benchmark BenchmarkRunner
}

func NewService(db *gorm.DB, sandbox SandboxExecutor, benchmark BenchmarkRunner) *Service {
	return &Service{db: db, security: NewSecurityScanner(), sandbox: sandbox, benchmark: benchmark}
}

func (s *Service) RunTests(ctx context.Context, functionID uuid.UUID, code, runtime string) ([]TestResult, error) {
	results := []TestResult{
		s.runSyntaxValidation(functionID, code, runtime),
		s.runSecurityScan(functionID, code),
		s.runUnitPrimitive(functionID, code),
	}
	results = append(results, s.runSandboxExecution(ctx, functionID, code, runtime))
	results = append(results, s.runBenchmark(ctx, functionID, code, runtime))

	for i := range results {
		if err := s.db.WithContext(ctx).Create(&results[i]).Error; err != nil {
			return results, fmt.Errorf("persist test result: %w", err)
		}
	}
	return results, nil
}

func AggregateScore(results []TestResult) float64 {
	if len(results) == 0 {
		return 0
	}
	total := 0.0
	for _, result := range results {
		total += result.Score
	}
	return total / float64(len(results))
}

func AllPassed(results []TestResult) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

func (s *Service) runSyntaxValidation(functionID uuid.UUID, code, runtime string) TestResult {
	start := time.Now()
	passed := strings.TrimSpace(code) != ""
	status := StatusPassed
	errMsg := ""
	details := map[string]any{"runtime": runtime, "length": len(code)}
	if strings.Contains(strings.ToLower(runtime), "python") && !strings.Contains(code, "def ") {
		passed = false
		errMsg = "python handler function not detected"
	}
	if strings.Contains(strings.ToLower(runtime), "node") && !strings.Contains(code, "handler") {
		passed = false
		errMsg = "javascript handler function not detected"
	}
	if !passed {
		status = StatusFailed
	}
	return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StageSyntax, Passed: passed, Status: status, Score: boolScore(passed, 92, 25), DurationMs: int(time.Since(start).Milliseconds()), Error: errMsg, Details: details}
}

func (s *Service) runSecurityScan(functionID uuid.UUID, code string) TestResult {
	start := time.Now()
	passed, findings := s.security.Scan(code)
	status := StatusPassed
	errMsg := ""
	if !passed {
		status = StatusFailed
		errMsg = strings.Join(findings, "; ")
	}
	return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StageSecurity, Passed: passed, Status: status, Score: boolScore(passed, 88, 10), DurationMs: int(time.Since(start).Milliseconds()), Error: errMsg, Details: map[string]any{"findings": findings}}
}

func (s *Service) runUnitPrimitive(functionID uuid.UUID, code string) TestResult {
	start := time.Now()
	hasAssertions := strings.Contains(code, "assert") || strings.Contains(code, "expect(") || strings.Contains(code, "pytest")
	passed := hasAssertions || strings.Contains(code, "__main__") || strings.Contains(code, "if require.main")
	status := StatusPassed
	errMsg := ""
	if !passed {
		status = StatusFailed
		errMsg = "no test or executable validation primitive detected"
	}
	return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StageUnit, Passed: passed, Status: status, Score: boolScore(passed, 75, 30), DurationMs: int(time.Since(start).Milliseconds()), Error: errMsg, Details: map[string]any{"has_assertions": hasAssertions}}
}

func (s *Service) runSandboxExecution(ctx context.Context, functionID uuid.UUID, code, runtime string) TestResult {
	start := time.Now()
	if s.sandbox == nil {
		passed := strings.TrimSpace(code) != ""
		return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StageExecution, Passed: passed, Status: ternaryStatus(passed), Score: boolScore(passed, 80, 35), DurationMs: int(time.Since(start).Milliseconds()), Error: ternaryError(passed, "", "sandbox executor not configured"), Details: map[string]any{"mode": "heuristic"}}
	}
	result, err := s.sandbox.Execute(ctx, runtime, code, map[string]any{"sample": true})
	passed := err == nil && result.Passed
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = result.Error
	}
	return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StageExecution, Passed: passed, Status: ternaryStatus(passed), Score: boolScore(passed, 85, 20), DurationMs: maxDuration(result.DurationMs, int(time.Since(start).Milliseconds())), Error: errMsg, Details: map[string]any{"output": result.Output}}
}

func (s *Service) runBenchmark(ctx context.Context, functionID uuid.UUID, code, runtime string) TestResult {
	start := time.Now()
	if s.benchmark == nil {
		latencyMs := estimateLatency(code)
		passed := latencyMs <= 500
		return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StagePerformance, Passed: passed, Status: ternaryStatus(passed), Score: boolScore(passed, 82, 40), DurationMs: int(time.Since(start).Milliseconds()), Error: ternaryError(passed, "", "benchmark runner not configured"), Details: map[string]any{"estimated_avg_exec_ms": latencyMs, "estimated_cold_start_ms": estimateColdStart(runtime)}}
	}
	result, err := s.benchmark.Run(ctx, runtime, code)
	passed := err == nil && result.Passed
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = result.Error
	}
	return TestResult{ID: uuid.New(), FunctionID: functionID, Stage: StagePerformance, Passed: passed, Status: ternaryStatus(passed), Score: boolScore(passed, 90, 25), DurationMs: maxDuration(result.DurationMs, int(time.Since(start).Milliseconds())), Error: errMsg, Details: map[string]any{"cold_start_ms": result.ColdStartMs, "average_exec_ms": result.AverageExecMs, "memory_bytes": result.MemoryBytes}}
}

func boolScore(passed bool, passScore, failScore float64) float64 {
	if passed {
		return passScore
	}
	return failScore
}

func ternaryStatus(passed bool) string {
	if passed {
		return StatusPassed
	}
	return StatusFailed
}

func ternaryError(passed bool, ok string, failed string) string {
	if passed {
		return ok
	}
	return failed
}

func maxDuration(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func estimateLatency(code string) int {
	latency := 120 + len(code)/50
	if strings.Contains(strings.ToLower(code), "requests") || strings.Contains(strings.ToLower(code), "fetch(") {
		latency += 250
	}
	return latency
}

func estimateColdStart(runtime string) int {
	if strings.Contains(strings.ToLower(runtime), "node") {
		return 450
	}
	return 700
}
