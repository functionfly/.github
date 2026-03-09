package testing

import "context"

type HeuristicBenchmarkRunner struct{}

func (HeuristicBenchmarkRunner) Run(_ context.Context, runtime, code string) (BenchmarkResult, error) {
	avg := estimateLatency(code)
	cold := estimateColdStart(runtime)
	passed := avg <= 500 && cold <= 2000
	return BenchmarkResult{
		Passed:        passed,
		DurationMs:    30,
		ColdStartMs:   cold,
		AverageExecMs: avg,
		MemoryBytes:   int64(len(code) * 32),
		Error:         ternaryError(passed, "", "performance thresholds exceeded"),
	}, nil
}
