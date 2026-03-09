package testing

import (
	"context"
	"strings"
)

type HeuristicSandboxExecutor struct{}

func (HeuristicSandboxExecutor) Execute(_ context.Context, runtime, code string, input map[string]any) (SandboxExecutionResult, error) {
	passed := strings.TrimSpace(code) != "" && (strings.Contains(code, "handler") || strings.Contains(code, "def "))
	return SandboxExecutionResult{
		Passed:     passed,
		DurationMs: 25,
		Output: map[string]any{
			"runtime": runtime,
			"input":   input,
		},
		Error: ternaryError(passed, "", "handler entrypoint not detected"),
	}, nil
}
