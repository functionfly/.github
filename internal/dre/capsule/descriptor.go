// Package capsule implements the Deterministic Compute Capsule (DCC) protocol.
// A DCC is the sealed execution universe that guarantees deterministic replay.
package capsule

import (
	"fmt"

	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

// CapsuleDescriptor is the canonical object describing the sealed execution universe.
// It is canonicalized and hashed into the EnvironmentHash of the MEG.
// All fields are fixed at execution time and cannot change during replay.
type CapsuleDescriptor struct {
	ProtocolVersion  string           `json:"protocol_version"`   // "dcc/1.0"
	RuntimeVersion   string           `json:"runtime_version"`    // "wasm/1.12.3"
	EngineVersion    string           `json:"engine_version"`     // "fx-wasm/2.4.1"
	CapsuleVersion   string           `json:"capsule_version"`    // "dcc-core/1.0"
	CPUArch          string           `json:"cpu_arch"`           // "x86_64"
	MemoryLimit      int64            `json:"memory_limit"`       // bytes
	InstructionLimit int64            `json:"instruction_limit"`
	TimeSeed         string           `json:"time_seed"`          // H(execution_id)
	RNGSeed          string           `json:"rng_seed"`           // H(input_hash || env_hash)
	FSSnapshotHash   string           `json:"fs_snapshot_hash"`
	NetworkMode      string           `json:"network_mode"`       // "record"|"stub"|"disabled"
	SyscallProfile   string           `json:"syscall_profile"`    // "strict-v1"
	FloatMode        string           `json:"float_mode"`         // "ieee754-strict"
	DeterminismFlags DeterminismFlags `json:"determinism_flags"`
	DeterminismTier  string           `json:"determinism_tier"`   // "full"|"lite"
}

// DeterminismFlags controls scheduler and JIT behavior for deterministic execution.
type DeterminismFlags struct {
	LockScheduler      bool `json:"lock_scheduler"`
	DisableJITVariance bool `json:"disable_jit_variance"`
	FixedThreadCount   int  `json:"fixed_thread_count"`
}

// Hash returns the canonical hash of this descriptor.
// The hash is used as the capsule_descriptor_hash in MEG records.
func (d *CapsuleDescriptor) Hash() (string, error) {
	b, err := drecrypto.Canonicalize(d)
	if err != nil {
		return "", fmt.Errorf("capsule: hash descriptor: %w", err)
	}
	return drecrypto.HashString(drecrypto.TagEnv, b), nil
}

// Default returns a DCC v1.0 descriptor with safe defaults for a new execution.
// executionID, inputHash, and envHash are used to derive deterministic seeds.
func Default(executionID, inputHash, envHash string) *CapsuleDescriptor {
	// Derive time seed from execution ID (deterministic, not wall clock)
	timeSeed := drecrypto.HashString(drecrypto.TagMeta, []byte(executionID))

	// Derive RNG seed from input + env hashes
	rngSeed := drecrypto.HashString(drecrypto.TagMeta, []byte(inputHash+envHash))

	return &CapsuleDescriptor{
		ProtocolVersion:  "dcc/1.0",
		RuntimeVersion:   "wasm/1.0",
		EngineVersion:    "fx-wasm/1.0",
		CapsuleVersion:   "dcc-core/1.0",
		CPUArch:          "x86_64",
		MemoryLimit:      128 * 1024 * 1024, // 128 MB
		InstructionLimit: 1_000_000_000,     // 1B instructions
		TimeSeed:         timeSeed,
		RNGSeed:          rngSeed,
		FSSnapshotHash:   "",
		NetworkMode:      "disabled",
		SyscallProfile:   "strict-v1",
		FloatMode:        "ieee754-strict",
		DeterminismFlags: DeterminismFlags{
			LockScheduler:      true,
			DisableJITVariance: true,
			FixedThreadCount:   1,
		},
		DeterminismTier: "full",
	}
}
