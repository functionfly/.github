package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubSandbox struct {
	result SandboxExecutionResult
	err    error
}

func (s stubSandbox) Execute(_ context.Context, _, _ string, _ map[string]any) (SandboxExecutionResult, error) {
	return s.result, s.err
}

type stubBenchmark struct {
	result BenchmarkResult
	err    error
}

func (b stubBenchmark) Run(_ context.Context, _, _ string) (BenchmarkResult, error) {
	return b.result, b.err
}

func TestSecurityScannerDetectsCriticalPatternsCaseInsensitively(t *testing.T) {
	t.Parallel()

	passed, findings := NewSecurityScanner().Scan("result = EVAL(input())")

	assert.False(t, passed, "eval usage should fail the security scan")
	assert.Contains(t, findings, "critical:eval(", "critical finding should identify the matched pattern")
}

func TestRunTestsPersistsExpectedStagesAndFailures(t *testing.T) {
	t.Parallel()

	db := newTestingTestDB(t)
	svc := NewService(db, stubSandbox{result: SandboxExecutionResult{Passed: false, DurationMs: 12, Error: "sandbox failed"}}, stubBenchmark{err: errors.New("benchmark unavailable")})
	require.NoError(t, db.AutoMigrate(&TestResult{}))

	results, err := svc.RunTests(context.Background(), uuid.New(), "print('hello')", "python3.11")

	require.NoError(t, err)
	require.Len(t, results, 5, "the factory pipeline expects five staged test results")
	assert.Equal(t, StageSyntax, results[0].Stage)
	assert.False(t, results[0].Passed, "python code without a handler should fail syntax validation")
	assert.Equal(t, "sandbox failed", results[3].Error)
	assert.Equal(t, StatusFailed, results[4].Status, "benchmark runner errors should be surfaced as failed performance tests")

	var persisted int64
	require.NoError(t, db.Model(&TestResult{}).Count(&persisted).Error)
	assert.Equal(t, int64(5), persisted, "all generated test stages should be persisted for auditability")
}

func TestAggregateScoreAndAllPassedReflectStageOutcomes(t *testing.T) {
	t.Parallel()

	results := []TestResult{{Score: 90, Passed: true}, {Score: 80, Passed: true}, {Score: 40, Passed: false}}

	assert.InDelta(t, 70.0, AggregateScore(results), 0.001, "aggregate score should average all stage scores")
	assert.False(t, AllPassed(results), "a single failing stage should mark the run as not fully passed")
}

func newTestingTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
