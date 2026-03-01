package crypto

import (
	"encoding/json"
	"fmt"
)

// PerfSig is the Resource Usage Signature 2.0.
// It's a cryptographic fingerprint of execution performance characteristics
// that enables:
// - Detection of replay manipulation
// - Anti-optimization fraud detection
// - Trust Score integration
// - Marketplace fairness
type PerfSig struct {
	// CPU cycles consumed during execution
	CPUCycles uint64 `json:"cpu_cycles"`
	// Peak memory usage in bytes
	MemoryPeak uint64 `json:"memory_peak"`
	// Number of system calls made
	SyscallCount uint32 `json:"syscall_count"`
	// Hash of I/O patterns (files read/written, network calls)
	IOPatterns string `json:"io_patterns"`
	// WASM operation counts (for Wasm execution)
	WasmOpCounts map[string]uint64 `json:"wasm_op_counts"`
	// Hash of scheduling trace (for parallel execution)
	SchedulingTrace string `json:"scheduling_trace"`
	// Wall clock time in microseconds
	WallTimeUS uint64 `json:"wall_time_us"`
	// Peak concurrent goroutines (if applicable)
	PeakGoroutines uint32 `json:"peak_goroutines"`
}

// ResourceUsage contains raw performance data from execution.
// This is provided by the runtime during execution.
type ResourceUsage struct {
	CPUCycles      uint64            `json:"cpu_cycles"`
	MemoryPeak     uint64            `json:"memory_peak"`
	SyscallCount   uint32            `json:"syscall_count"`
	IOPatterns     []string          `json:"io_patterns"` // List of I/O operations
	WasmOpCounts   map[string]uint64 `json:"wasm_op_counts"`
	WallTimeUS     uint64            `json:"wall_time_us"`
	PeakGoroutines uint32            `json:"peak_goroutines"`
}

// ComputePerfSig generates a cryptographic performance signature from raw resource usage.
// The signature is deterministic and can be used for replay verification.
func ComputePerfSig(usage ResourceUsage) (*PerfSig, error) {
	if usage.CPUCycles == 0 && usage.MemoryPeak == 0 && usage.WallTimeUS == 0 {
		return nil, fmt.Errorf("perfsig: empty resource usage")
	}

	// Hash I/O patterns
	ioHash := HashString(TagResource, []byte(joinStrings(usage.IOPatterns)))

	// Hash WASM operation counts if present
	var wasmHash string
	if len(usage.WasmOpCounts) > 0 {
		wasmBytes, err := json.Marshal(usage.WasmOpCounts)
		if err != nil {
			return nil, fmt.Errorf("perfsig: marshal wasm ops: %w", err)
		}
		wasmHash = HashString(TagResource, wasmBytes)
	} else {
		wasmHash = HashString(TagResource, []byte(""))
	}

	// Scheduling trace is not currently tracked - use empty hash
	schedulingHash := HashString(TagResource, []byte(""))

	return &PerfSig{
		CPUCycles:       usage.CPUCycles,
		MemoryPeak:      usage.MemoryPeak,
		SyscallCount:    usage.SyscallCount,
		IOPatterns:      ioHash,
		WasmOpCounts:    usage.WasmOpCounts,
		SchedulingTrace: schedulingHash,
		WallTimeUS:      usage.WallTimeUS,
		PeakGoroutines:  usage.PeakGoroutines,
	}, nil
}

// Hash returns the cryptographic hash of this PerfSig.
// This hash is included in the MEG ResourceHash component.
func (p *PerfSig) Hash() (string, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("perfsig: marshal: %w", err)
	}
	return HashString(TagResource, bytes), nil
}

// Equals compares two PerfSig for equality.
// Used for replay verification.
func (p *PerfSig) Equals(other *PerfSig) bool {
	if p == nil || other == nil {
		return p == other
	}
	return p.CPUCycles == other.CPUCycles &&
		p.MemoryPeak == other.MemoryPeak &&
		p.SyscallCount == other.SyscallCount &&
		p.IOPatterns == other.IOPatterns &&
		p.WallTimeUS == other.WallTimeUS &&
		p.PeakGoroutines == other.PeakGoroutines
}

// ComputeResourceHash computes the MEG ResourceHash from performance data.
// This is a convenience function that combines ComputePerfSig and PerfSig.Hash.
func ComputeResourceHash(usage ResourceUsage) (string, error) {
	perfSig, err := ComputePerfSig(usage)
	if err != nil {
		return "", err
	}
	return perfSig.Hash()
}

// joinStrings joins a slice of strings deterministically.
// Used for I/O pattern hashing.
func joinStrings(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s + "|"
	}
	return result
}

// PerfSigFromHash reconstructs a PerfSig from its hash and raw data.
// This is used for verification where we only store the hash.
func PerfSigFromHash(hash string, usage ResourceUsage) (*PerfSig, error) {
	perfSig, err := ComputePerfSig(usage)
	if err != nil {
		return nil, err
	}
	computedHash, err := perfSig.Hash()
	if err != nil {
		return nil, err
	}
	if computedHash != hash {
		return nil, fmt.Errorf("perfsig: hash mismatch: expected %s, got %s", hash, computedHash)
	}
	return perfSig, nil
}

// DetectManipulation analyzes two PerfSig results to detect potential manipulation.
// Returns true if the performance signatures differ in ways that suggest manipulation.
func DetectManipulation(original, replay *PerfSig) (bool, string) {
	if original == nil || replay == nil {
		return false, ""
	}

	// Check for impossible optimizations
	// If replay is significantly faster, it might be fake
	if original.WallTimeUS > 0 && replay.WallTimeUS > 0 {
		speedup := float64(original.WallTimeUS) / float64(replay.WallTimeUS)
		if speedup > 10 {
			return true, fmt.Sprintf("suspicious speedup: %.1fx faster", speedup)
		}
	}

	// Check for memory anomalies
	if replay.MemoryPeak > original.MemoryPeak*2 {
		return true, "suspicious memory increase"
	}

	// Check for syscall count anomalies
	if replay.SyscallCount > original.SyscallCount*2 {
		return true, "suspicious syscall count increase"
	}

	return false, ""
}
