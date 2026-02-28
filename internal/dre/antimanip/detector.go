// Package antimanip implements the DRE 2.0 anti-manipulation layer.
// It analyzes replay results, classifies drift categories, and computes
// trust penalties for functions that exhibit non-deterministic behavior.
package antimanip

import (
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

// DriftDetector analyzes replay results and classifies manipulation attempts.
type DriftDetector struct{}

// Analyze compares original and replay MEG results.
// Returns a DriftReport if the executions diverged, or nil if they match.
func (d *DriftDetector) Analyze(original, replay *drecrypto.MEGResult) (*capsule.DriftReport, error) {
	if original == nil || replay == nil {
		return nil, fmt.Errorf("antimanip: nil MEG result")
	}

	// If root hashes match, no drift
	if original.ExecutionRootHash == replay.ExecutionRootHash {
		return nil, nil
	}

	// Find which component hashes differ
	componentDiff := findComponentDiff(original, replay)

	// Classify the drift category from the component diff
	category := ClassifyDrift(componentDiff)

	// Compute trust penalty
	penalty := TrustPenalty(category)

	return &capsule.DriftReport{
		OriginalRoot:  original.ExecutionRootHash,
		ReplayRoot:    replay.ExecutionRootHash,
		Category:      category,
		ComponentDiff: componentDiff,
		DetectedAt:    time.Now(),
		TrustPenalty:  penalty,
	}, nil
}

// findComponentDiff returns the names of MEG components that differ between two results.
func findComponentDiff(original, replay *drecrypto.MEGResult) []string {
	var diff []string

	if original.InputHash != replay.InputHash {
		diff = append(diff, "input")
	}
	if original.EnvironmentHash != replay.EnvironmentHash {
		diff = append(diff, "environment")
	}
	if original.DependencyHash != replay.DependencyHash {
		diff = append(diff, "dependency")
	}
	if original.TraceHash != replay.TraceHash {
		diff = append(diff, "trace")
	}
	if original.ResourceHash != replay.ResourceHash {
		diff = append(diff, "resource")
	}
	if original.OutputHash != replay.OutputHash {
		diff = append(diff, "output")
	}
	if original.MetadataHash != replay.MetadataHash {
		diff = append(diff, "metadata")
	}

	return diff
}

// ClassifyDrift determines the most likely drift category from a component diff.
// The classification uses a priority ordering based on the severity and specificity
// of each component's role in determinism.
func ClassifyDrift(componentDiff []string) capsule.DriftCategory {
	if len(componentDiff) == 0 {
		return capsule.DriftUnknown
	}

	// Build a set for O(1) lookup
	diffSet := make(map[string]bool, len(componentDiff))
	for _, c := range componentDiff {
		diffSet[c] = true
	}

	// Priority classification:
	// 1. Dependency mutation is the most severe (indicates tampering)
	if diffSet["dependency"] {
		return capsule.DriftDependencyMutation
	}

	// 2. Syscall mismatch (environment + trace differ without dependency change)
	if diffSet["environment"] && diffSet["trace"] {
		return capsule.DriftSyscall
	}

	// 3. Memory access mismatch (resource + trace differ)
	if diffSet["resource"] && diffSet["trace"] {
		return capsule.DriftMemoryAccess
	}

	// 4. Instruction count mismatch (trace differs, output may differ)
	if diffSet["trace"] && diffSet["output"] {
		return capsule.DriftInstructionCount
	}

	// 5. Network mismatch (output differs, environment unchanged)
	if diffSet["output"] && !diffSet["environment"] && !diffSet["dependency"] {
		return capsule.DriftNetwork
	}

	// 6. Floating point mismatch (resource differs slightly)
	if diffSet["resource"] && !diffSet["trace"] {
		return capsule.DriftFloatingPoint
	}

	// 7. RNG divergence (output differs, everything else matches)
	if diffSet["output"] {
		return capsule.DriftRNG
	}

	return capsule.DriftUnknown
}

// TrustPenalty returns the trust score penalty (as a negative float64) for a given drift category.
// Penalties are applied to the function's trust score when drift is detected.
//
// Penalty table:
//
//	rng_divergence              -0.05
//	syscall_mismatch            -0.10
//	network_mismatch            -0.03
//	floating_point_mismatch     -0.02
//	instruction_count_mismatch  -0.08
//	memory_access_mismatch      -0.08
//	dependency_mutation         -0.15
//	unknown                     -0.05
func TrustPenalty(category capsule.DriftCategory) float64 {
	switch category {
	case capsule.DriftRNG:
		return 0.05
	case capsule.DriftSyscall:
		return 0.10
	case capsule.DriftNetwork:
		return 0.03
	case capsule.DriftFloatingPoint:
		return 0.02
	case capsule.DriftInstructionCount:
		return 0.08
	case capsule.DriftMemoryAccess:
		return 0.08
	case capsule.DriftDependencyMutation:
		return 0.15
	default:
		return 0.05
	}
}
