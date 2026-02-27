package execution

import (
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage/registry"
)

// ReplayVerificationStatus represents the status of replay verification
type ReplayVerificationStatus string

const (
	VerificationPending  ReplayVerificationStatus = "pending"
	VerificationVerified ReplayVerificationStatus = "verified"
	VerificationFailed   ReplayVerificationStatus = "failed"
)

// ReplayVerificationResult represents the result of a replay verification.
// DRE 2.0 adds MEG-based root hash comparison fields.
type ReplayVerificationResult struct {
	Status           ReplayVerificationStatus
	OriginalOutput   json.RawMessage
	ReplayedOutput   json.RawMessage
	OriginalDuration int
	ReplayedDuration int
	Error            string
	VerifiedAt       time.Time
	OutputMatches    bool

	// DRE 2.0 fields — MEG-based verification
	OriginalRootHash string
	ReplayRootHash   string
	DriftCategory    capsule.DriftCategory
	ComponentDiff    []string
	MEGRecord        *registry.MEGRecord
	OriginalMEG      *drecrypto.MEGResult
	ReplayMEG        *drecrypto.MEGResult
}

// ResourceUsage represents resource usage for an execution
type ResourceUsage struct {
	MaxMemoryMB    int
	MaxCPUTimeMs   int
	MemoryUsedMB   float64
	CPUTimeUsedMs  int
	WallTimeUsedMs int
}

// ExecutionError wraps execution errors with resource usage information
type ExecutionError struct {
	Err           error
	ResourceUsage *ResourceUsage
	TerminatedBy  string
}

func (ee *ExecutionError) Error() string {
	return ee.Err.Error()
}

func (ee *ExecutionError) Unwrap() error {
	return ee.Err
}