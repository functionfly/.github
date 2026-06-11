package e2e

import (
	"context"
	"errors"
	"testing"

	agenttesting "github.com/functionfly/functionfly/internal/agent/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// Security Scanner Tests
// ============================================================

func TestSecurityScanner(t *testing.T) {
	t.Run("should detect critical eval pattern", func(t *testing.T) {
		scanner := agenttesting.NewSecurityScanner()

		code := "result = eval(user_input)"
		passed, findings := scanner.Scan(code)

		assert.False(t, passed)
		assert.NotEmpty(t, findings)
		assert.Contains(t, findings, "critical:eval(")
	})

	t.Run("should detect eval in various forms", func(t *testing.T) {
		scanner := agenttesting.NewSecurityScanner()

		dangerousCode := []string{
			"eval(input())",
			"result = EVAL(request.body)",
			"exec('print(1)')",
			"compile(code, '', 'exec')",
		}

		for _, code := range dangerousCode {
			passed, findings := scanner.Scan(code)
			assert.False(t, passed, "code '%s' should fail security scan", code)
			assert.NotEmpty(t, findings, "code '%s' should have findings", code)
		}
	})

	t.Run("should pass safe code", func(t *testing.T) {
		scanner := agenttesting.NewSecurityScanner()

		safeCode := []string{
			"def hello(): return 'world'",
			"x = [1, 2, 3]",
			"print('hello')",
			"result = x + y",
		}

		for _, code := range safeCode {
			passed, findings := scanner.Scan(code)
			assert.True(t, passed, "code '%s' should pass security scan", code)
			assert.Empty(t, findings, "code '%s' should have no findings", code)
		}
	})

	t.Run("should detect dangerous imports", func(t *testing.T) {
		scanner := agenttesting.NewSecurityScanner()

		code := "import os; import sys; eval('ls')"
		passed, findings := scanner.Scan(code)

		assert.False(t, passed)
		assert.NotEmpty(t, findings)
	})
}

// ============================================================
// Testing Service Tests
// ============================================================

func newTestingDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestTestingService(t *testing.T) {
	t.Run("should run tests and persist all stages", func(t *testing.T) {
		db := newTestingDB(t)
		require.NoError(t, db.AutoMigrate(&agenttesting.TestResult{}))

		stubSandbox := &stubSandboxForTesting{
			result: agenttesting.SandboxExecutionResult{Passed: false, DurationMs: 10, Error: "sandbox error"},
		}
		stubBench := &stubBenchmarkForTesting{
			err: errors.New("benchmark unavailable"),
		}

		svc := agenttesting.NewService(db, stubSandbox, stubBench)

		code := "print('hello')"
		runtime := "python3.11"
		runID := uuid.New()

		results, err := svc.RunTests(context.Background(), runID, code, runtime)
		require.NoError(t, err)
		require.Len(t, results, 5, "expected 5 test stages")

		// First stage should be syntax
		assert.Equal(t, agenttesting.StageSyntax, results[0].Stage)
		assert.False(t, results[0].Passed, "python without handler should fail syntax")

		// Last stage is performance which should fail
		assert.Equal(t, agenttesting.StagePerformance, results[4].Stage)
		assert.Equal(t, agenttesting.StatusFailed, results[4].Status)

		// Verify persistence
		var count int64
		db.Model(&agenttesting.TestResult{}).Count(&count)
		assert.Equal(t, int64(5), count)
	})

	t.Run("should calculate aggregate score", func(t *testing.T) {
		results := []agenttesting.TestResult{
			{Score: 100, Passed: true},
			{Score: 80, Passed: true},
			{Score: 50, Passed: false},
			{Score: 90, Passed: true},
			{Score: 70, Passed: false},
		}

		score := agenttesting.AggregateScore(results)
		assert.InDelta(t, 78.0, score, 0.1, "aggregate should be weighted average")
	})

	t.Run("should return false for AllPassed when any fails", func(t *testing.T) {
		results := []agenttesting.TestResult{
			{Score: 100, Passed: true},
			{Score: 80, Passed: true},
			{Score: 50, Passed: false},
		}

		assert.False(t, agenttesting.AllPassed(results))
	})

	t.Run("should return true for AllPassed when all pass", func(t *testing.T) {
		results := []agenttesting.TestResult{
			{Score: 100, Passed: true},
			{Score: 80, Passed: true},
			{Score: 90, Passed: true},
		}

		assert.True(t, agenttesting.AllPassed(results))
	})

	t.Run("should handle empty results for aggregate", func(t *testing.T) {
		results := []agenttesting.TestResult{}

		score := agenttesting.AggregateScore(results)
		assert.Equal(t, 0.0, score)
	})
}

// Stubs for testing service

type stubSandboxForTesting struct {
	result agenttesting.SandboxExecutionResult
	err    error
}

func (s *stubSandboxForTesting) Execute(ctx context.Context, code, runtime string, inputs map[string]any) (agenttesting.SandboxExecutionResult, error) {
	return s.result, s.err
}

type stubBenchmarkForTesting struct {
	result agenttesting.BenchmarkResult
	err    error
}

func (b *stubBenchmarkForTesting) Run(ctx context.Context, code, runtime string) (agenttesting.BenchmarkResult, error) {
	return b.result, b.err
}
