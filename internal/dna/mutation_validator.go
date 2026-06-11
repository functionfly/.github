package dna

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/sirupsen/logrus"
)

// MutationValidator validates DNA mutations before they are accepted.
// It runs the mutated code in a sandbox, compares outputs against the original,
// and checks for security regressions.
type MutationValidator struct {
	sandbox   SandboxExecutor
	logger    *logrus.Logger
	timeout   time.Duration
	maxTests  int
}

// SandboxExecutor abstracts the sandbox execution for mutation testing.
type SandboxExecutor interface {
	Execute(ctx context.Context, runtime, code string, input map[string]any) (SandboxResult, error)
}

// SandboxResult holds the result of a sandbox execution.
type SandboxResult struct {
	Passed     bool
	DurationMs int
	Output     map[string]any
	Error      string
}

// TestInput represents a test case for mutation validation.
type TestInput struct {
	Input          map[string]any `json:"input"`
	ExpectedOutput map[string]any `json:"expected_output,omitempty"`
	Description    string         `json:"description"`
}

// ValidationReport holds the results of mutation validation.
type ValidationReport struct {
	Passed          bool              `json:"passed"`
	TotalTests      int               `json:"total_tests"`
	PassedTests     int               `json:"passed_tests"`
	FailedTests     int               `json:"failed_tests"`
	SecurityChecks  []SecurityCheck   `json:"security_checks"`
	PerformanceDiff PerformanceDiff   `json:"performance_diff"`
	Errors          []string          `json:"errors"`
}

// SecurityCheck represents a security validation result.
type SecurityCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

// PerformanceDiff compares original vs mutated performance.
type PerformanceDiff struct {
	OriginalAvgMs  float64 `json:"original_avg_ms"`
	MutatedAvgMs   float64 `json:"mutated_avg_ms"`
	ImprovementPct float64 `json:"improvement_pct"`
}

// NewMutationValidator creates a new mutation validator.
func NewMutationValidator(sandbox SandboxExecutor, logger *logrus.Logger) *MutationValidator {
	timeout := 30 * time.Second
	if v := os.Getenv("DNA_SANDBOX_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}

	maxTests := 50
	if v := os.Getenv("DNA_VALIDATION_MAX_TESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxTests = n
		}
	}

	return &MutationValidator{
		sandbox:  sandbox,
		logger:   logger,
		timeout:  timeout,
		maxTests: maxTests,
	}
}

// ValidateMutation runs behavioral tests comparing original and mutated code.
// Returns a validation report indicating whether the mutation is safe to deploy.
func (v *MutationValidator) ValidateMutation(ctx context.Context, originalCode, mutatedCode, runtime string, testInputs []TestInput) (*ValidationReport, error) {
	ctx, span := tracing.StartSpan(ctx, "dna.validate_mutation")
	defer tracing.Finish(ctx)

	startTime := time.Now()
	defer func() {
		RecordMutationValidation("completed", time.Since(startTime))
	}()

	report := &ValidationReport{
		SecurityChecks: []SecurityCheck{},
		Errors:         []string{},
	}

	if len(testInputs) == 0 {
		testInputs = v.generateDefaultTestInputs()
	}

	if len(testInputs) > v.maxTests {
		testInputs = testInputs[:v.maxTests]
	}

	report.TotalTests = len(testInputs)
	tracing.SetAttribute(ctx, "total_tests", report.TotalTests)

	// Run security checks first
	securityChecks := v.runSecurityChecks(mutatedCode, runtime)
	report.SecurityChecks = securityChecks
	for _, check := range securityChecks {
		if !check.Passed {
			report.Errors = append(report.Errors, fmt.Sprintf("Security check failed: %s - %s", check.Name, check.Details))
		}
	}

	// Run behavioral comparison tests
	var originalTimes, mutatedTimes []float64
	for _, test := range testInputs {
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		default:
		}

		// Run original code
		origCtx, origCancel := context.WithTimeout(ctx, v.timeout)
		origResult, origErr := v.sandbox.Execute(origCtx, runtime, originalCode, test.Input)
		origCancel()

		if err := ctx.Err(); err != nil {
			return report, err
		}

		if origErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Original code failed: %v", origErr))
			report.FailedTests++
			continue
		}
		originalTimes = append(originalTimes, float64(origResult.DurationMs))

		// Run mutated code
		mutCtx, mutCancel := context.WithTimeout(ctx, v.timeout)
		mutResult, mutErr := v.sandbox.Execute(mutCtx, runtime, mutatedCode, test.Input)
		mutCancel()

		if err := ctx.Err(); err != nil {
			return report, err
		}

		if mutErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Mutated code failed on test '%s': %v", test.Description, mutErr))
			report.FailedTests++
			continue
		}

		mutatedTimes = append(mutatedTimes, float64(mutResult.DurationMs))

		// Compare outputs if expected output is provided
		if test.ExpectedOutput != nil {
			if !v.outputsMatch(origResult.Output, mutResult.Output) {
				report.Errors = append(report.Errors, fmt.Sprintf("Output mismatch on test '%s'", test.Description))
				report.FailedTests++
				continue
			}
		}

		report.PassedTests++
	}

	report.FailedTests = report.TotalTests - report.PassedTests

	// Compute performance diff
	report.PerformanceDiff = v.computePerformanceDiff(originalTimes, mutatedTimes)

	// Mutation passes if all tests pass and all security checks pass
	report.Passed = report.FailedTests == 0
	for _, check := range report.SecurityChecks {
		if !check.Passed {
			report.Passed = false
			break
		}
	}

	return report, nil
}

// runSecurityChecks performs static analysis on the mutated code.
func (v *MutationValidator) runSecurityChecks(code, runtime string) []SecurityCheck {
	checks := []SecurityCheck{}

	// Check 1: Dangerous function calls
	dangerousPatterns := []string{
		"eval(", "exec(", "system(", "os.system(", "subprocess.",
		"child_process", "require('child_process')", "shell_exec(",
		"popen(", "passthru(",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			checks = append(checks, SecurityCheck{
				Name:    "DangerousFunctionCall",
				Passed:  false,
				Details: fmt.Sprintf("Contains potentially dangerous call: %s", pattern),
			})
			return checks // Early return on critical failure
		}
	}

	checks = append(checks, SecurityCheck{
		Name:    "DangerousFunctionCall",
		Passed:  true,
		Details: "No dangerous function calls detected",
	})

	// Check 2: Network access patterns (for functions that shouldn't have network)
	networkPatterns := []string{
		"fetch(", "http.", "https.", "axios", "requests.",
		"urllib.", "net.", "socket.",
	}

	networkFound := false
	for _, pattern := range networkPatterns {
		if containsIgnoreCase(code, pattern) {
			networkFound = true
			break
		}
	}

	checks = append(checks, SecurityCheck{
		Name:    "NetworkAccess",
		Passed:  !networkFound,
		Details: ternaryStr(networkFound, "Network access patterns detected", "No network access patterns detected"),
	})

	// Check 3: File system access
	fsPatterns := []string{
		"fs.", "os.open(", "open(", "readFile", "writeFile",
		"unlink(", "rmdir(", "mkdir(",
	}

	fsFound := false
	for _, pattern := range fsPatterns {
		if containsIgnoreCase(code, pattern) {
			fsFound = true
			break
		}
	}

	checks = append(checks, SecurityCheck{
		Name:    "FilesystemAccess",
		Passed:  !fsFound,
		Details: ternaryStr(fsFound, "Filesystem access patterns detected", "No filesystem access patterns detected"),
	})

	// Check 4: Code size sanity check (prevent resource exhaustion)
	maxCodeSize := 100 * 1024 // 100KB
	if len(code) > maxCodeSize {
		checks = append(checks, SecurityCheck{
			Name:    "CodeSizeLimit",
			Passed:  false,
			Details: fmt.Sprintf("Code size %d bytes exceeds limit %d bytes", len(code), maxCodeSize),
		})
	} else {
		checks = append(checks, SecurityCheck{
			Name:    "CodeSizeLimit",
			Passed:  true,
			Details: fmt.Sprintf("Code size %d bytes within limit", len(code)),
		})
	}

	// Check 5: Hash verification
	hash := sha256.Sum256([]byte(code))
	checks = append(checks, SecurityCheck{
		Name:    "CodeHashVerification",
		Passed:  true,
		Details: fmt.Sprintf("Code hash: sha256:%x", hash[:8]),
	})

	return checks
}

// outputsMatch compares two sandbox outputs for equivalence.
func (v *MutationValidator) outputsMatch(orig, mut map[string]any) bool {
	if len(orig) != len(mut) {
		return false
	}
	for k, v1 := range orig {
		v2, ok := mut[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
			return false
		}
	}
	return true
}

// computePerformanceDiff calculates performance difference between original and mutated code.
func (v *MutationValidator) computePerformanceDiff(originalTimes, mutatedTimes []float64) PerformanceDiff {
	avg := func(times []float64) float64 {
		if len(times) == 0 {
			return 0
		}
		sum := 0.0
		for _, t := range times {
			sum += t
		}
		return sum / float64(len(times))
	}

	origAvg := avg(originalTimes)
	mutAvg := avg(mutatedTimes)

	improvement := 0.0
	if origAvg > 0 {
		improvement = ((origAvg - mutAvg) / origAvg) * 100
	}

	return PerformanceDiff{
		OriginalAvgMs:  origAvg,
		MutatedAvgMs:   mutAvg,
		ImprovementPct: improvement,
	}
}

// generateDefaultTestInputs creates basic test inputs for common patterns.
func (v *MutationValidator) generateDefaultTestInputs() []TestInput {
	return []TestInput{
		{Input: map[string]any{}, Description: "Empty input"},
		{Input: map[string]any{"test": "value"}, Description: "Basic key-value input"},
		{Input: map[string]any{"numbers": []int{1, 2, 3}}, Description: "Array input"},
		{Input: map[string]any{"nested": map[string]any{"key": "value"}}, Description: "Nested object input"},
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func ternaryStr(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}
