package execution

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage"
)

// megInputPayload is the canonical input structure for MEG hashing.
type megInputPayload struct {
	Args     json.RawMessage `json:"args"`
	CallerID string          `json:"caller_id"`
	FxURI    string          `json:"fx_uri"`
	Seed     string          `json:"seed"`
}

// megEnvironmentPayload is the canonical environment structure for MEG hashing.
type megEnvironmentPayload struct {
	RuntimeVersion string                    `json:"runtime_version"`
	Capsule        *capsule.CapsuleDescriptor `json:"capsule"`
}

// megResourcePayload is the canonical resource usage structure for MEG hashing.
type megResourcePayload struct {
	CPUTimeMs      int     `json:"cpu_time_ms"`
	MemoryPeakMB   float64 `json:"memory_peak_mb"`
	WallTimeMs     int     `json:"wall_time_ms"`
}

// megOutputPayload is the canonical output structure for MEG hashing.
type megOutputPayload struct {
	ReturnValue json.RawMessage `json:"return_value"`
	ExitCode    int             `json:"exit_code"`
}

// megMetadataPayload is the canonical metadata structure for MEG hashing.
type megMetadataPayload struct {
	ExecutionID     string `json:"execution_id"`
	FunctionID      string `json:"function_id"`
	OwnerID         string `json:"owner_id"`
	CallerID        string `json:"caller_id"`
	NodeID          string `json:"node_id"`
	Region          string `json:"region"`
	Nonce           string `json:"nonce"`
	ProtocolVersion string `json:"protocol_version"`
	Timestamp       string `json:"timestamp"` // ISO-8601 UTC (virtual, from capsule seed)
}

// BuildMEGFromExecution constructs MEGComponents from execution context and builds the MEG.
// This is called after a successful function execution to create the cryptographic proof.
func BuildMEGFromExecution(
	fnVersion *storage.RegistryFunctionVersion,
	input json.RawMessage,
	output json.RawMessage,
	resourceUsage *ResourceUsage,
	capsuleDesc *capsule.CapsuleDescriptor,
	execMeta ExecutionMetadata,
) (*drecrypto.MEGResult, error) {
	if fnVersion == nil {
		return nil, fmt.Errorf("meg_builder: fnVersion is nil")
	}

	// Build input payload
	inputPayload := megInputPayload{
		Args:     input,
		CallerID: execMeta.CallerID,
		FxURI:    fmt.Sprintf("fx://%s/%s", fnVersion.FunctionID, fnVersion.Version),
		Seed:     execMeta.Nonce,
	}

	// Build environment payload
	envPayload := megEnvironmentPayload{
		RuntimeVersion: fnVersion.Runtime,
		Capsule:        capsuleDesc,
	}

	// Build dependencies from function version
	deps := buildDependencies(fnVersion)

	// Build resource payload
	var resPayload megResourcePayload
	if resourceUsage != nil {
		resPayload = megResourcePayload{
			CPUTimeMs:    resourceUsage.CPUTimeUsedMs,
			MemoryPeakMB: float64(resourceUsage.MemoryUsedMB),
			WallTimeMs:   resourceUsage.WallTimeUsedMs,
		}
	}

	// Build output payload
	outputPayload := megOutputPayload{
		ReturnValue: output,
		ExitCode:    0, // 0 = success
	}

	// Build metadata payload
	// Use a deterministic virtual timestamp derived from the capsule seed
	virtualTimestamp := time.Unix(0, 0).UTC().Format(time.RFC3339)
	if capsuleDesc != nil && capsuleDesc.TimeSeed != "" {
		// Use the capsule time seed as a deterministic timestamp marker
		virtualTimestamp = capsuleDesc.TimeSeed
	}

	metaPayload := megMetadataPayload{
		ExecutionID:     execMeta.ExecutionID,
		FunctionID:      execMeta.FunctionID,
		OwnerID:         execMeta.OwnerID,
		CallerID:        execMeta.CallerID,
		NodeID:          execMeta.NodeID,
		Region:          execMeta.Region,
		Nonce:           execMeta.Nonce,
		ProtocolVersion: execMeta.ProtocolVersion,
		Timestamp:       virtualTimestamp,
	}

	// Assemble MEG components
	components := drecrypto.MEGComponents{
		InputPayload:    inputPayload,
		EnvironmentData: envPayload,
		Dependencies:    deps,
		TraceChunks:     nil, // Lite tier by default; full tier would include trace
		ResourceUsage:   resPayload,
		OutputPayload:   outputPayload,
		Metadata:        metaPayload,
	}

	return drecrypto.BuildMEG(components)
}

// buildDependencies extracts dependency information from a function version.
// For WASM functions, the binary itself is the primary dependency.
func buildDependencies(fnVersion *storage.RegistryFunctionVersion) []drecrypto.Dependency {
	var deps []drecrypto.Dependency

	// Add the function binary as a dependency
	if fnVersion.ContentHash.Valid && fnVersion.ContentHash.String != "" {
		deps = append(deps, drecrypto.Dependency{
			Name:        fmt.Sprintf("fx://%s", fnVersion.FunctionID),
			Version:     fnVersion.Version,
			ContentHash: fnVersion.ContentHash.String,
		})
	}

	// Add source hash if available
	if fnVersion.SourceHash.Valid && fnVersion.SourceHash.String != "" {
		deps = append(deps, drecrypto.Dependency{
			Name:        fmt.Sprintf("fx://%s/source", fnVersion.FunctionID),
			Version:     fnVersion.Version,
			ContentHash: fnVersion.SourceHash.String,
		})
	}

	return deps
}
