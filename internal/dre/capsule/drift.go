package capsule

import "time"

// DriftCategory classifies why a replay diverged from the original execution.
// Each category maps to a specific trust penalty (see antimanip package).
type DriftCategory string

const (
	// DriftRNG indicates the random number generator produced different output.
	DriftRNG DriftCategory = "rng_divergence"
	// DriftSyscall indicates a system call produced a different result.
	DriftSyscall DriftCategory = "syscall_mismatch"
	// DriftNetwork indicates a network call returned different data.
	DriftNetwork DriftCategory = "network_mismatch"
	// DriftFloatingPoint indicates floating-point arithmetic produced different results.
	DriftFloatingPoint DriftCategory = "floating_point_mismatch"
	// DriftInstructionCount indicates the instruction count differed between executions.
	DriftInstructionCount DriftCategory = "instruction_count_mismatch"
	// DriftMemoryAccess indicates memory access patterns differed.
	DriftMemoryAccess DriftCategory = "memory_access_mismatch"
	// DriftDependencyMutation indicates a dependency was mutated between executions.
	DriftDependencyMutation DriftCategory = "dependency_mutation"
	// DriftUnknown is used when the category cannot be determined from component diffs.
	DriftUnknown DriftCategory = "unknown"
)

// DriftReport is the structured output when a replay diverges from the original execution.
// It is persisted to the drift_reports table and used to apply trust penalties.
type DriftReport struct {
	ExecutionID   string        `json:"execution_id"`
	OriginalRoot  string        `json:"original_root"`
	ReplayRoot    string        `json:"replay_root"`
	Category      DriftCategory `json:"category"`
	ComponentDiff []string      `json:"component_diff"` // which component hashes differ
	DetectedAt    time.Time     `json:"detected_at"`
	TrustPenalty  float64       `json:"trust_penalty"`
}
