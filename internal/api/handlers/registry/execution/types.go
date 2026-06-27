package execution

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/privacy"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// DNARecorder records execution metrics for the DNA analysis pipeline.
type DNARecorder interface {
	RecordExecutionFromPipeline(ctx context.Context, functionID, functionType string, durationMs int, statusCode int, coldStart bool, region string)
}

// BillingController interface for wallet operations
type BillingController = billing.ControllerInterface

// AgentBillingControls holds the economic controls for an agent
type AgentBillingControls = billing.AgentBillingControls

// CreditBalanceUpdate captures a credit balance mutation.
type CreditBalanceUpdate = billing.CreditBalanceUpdate

// Handler contains dependencies for execution handlers
type Handler struct {
	Repo         *registry.RegistryRepository
	BackendRepo  storage.Repository
	CacheService *cache.CacheService
	EdgeCache    *cache.EdgeCacheService
	// UsageTracker provides real-time quota enforcement and usage tracking
	UsageTracker UsageTracker
	// PrivacyService provides privacy and compliance features
	PrivacyService PrivacyService
	// DNARecorder records execution metrics for DNA analysis (optional)
	DNARecorder DNARecorder
	// MicroVMRepo provides MicroVM execution tracking and billing (optional)
	MicroVMRepo *storage.MicroVMRepository
	// NodeID is the identifier of this execution node (used in MEG records and certificates)
	NodeID string
	// Region is the geographic region of this node
	Region string
	// NodeKey is the Ed25519 private key used to sign FXCERTs. If nil, certs are generated without a node signature (e.g. bootstrap).
	NodeKey ed25519.PrivateKey
	// PlatformKey is the optional Ed25519 platform key; when set, certs include a platform signature (Platform Key ID in UI).
	PlatformKey ed25519.PrivateKey
	// BillingController handles wallet operations for paid function execution (optional)
	BillingController BillingController
	// RuntimeRouter routes execution to the appropriate engine based on runtime + tier.
	// When nil, execution falls back to the legacy direct path.
	RuntimeRouter *RuntimeRouter
	// BundleService provides eager bundling at publish time (optional).
	BundleService *bundler.BundleService
	// ReceiptMilestoneHook is called after a successful public execution receipt is created.
	ReceiptMilestoneHook func(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string)
}

// PrivacyService interface for privacy and compliance features
type PrivacyService interface {
	GetPrivacySettings(userID uuid.UUID) (*privacy.PrivacySettings, error)
	AnonymizeExecutionData(ip, userAgent, embedOrigin string, settings *privacy.PrivacySettings) (string, string, string)
	ShouldLogGeoData(settings *privacy.PrivacySettings) bool
	ShouldLogEmbedOrigin(settings *privacy.PrivacySettings) bool
	ShouldStoreInputOutput(settings *privacy.PrivacySettings) bool
	ScanForPII(data interface{}, redact bool) (*privacy.PIIDetectionResult, error)
	SanitizeInputOutput(input, output []byte) ([]byte, []byte, bool)
}

// UsageTracker interface for real-time quota enforcement
type UsageTracker interface {
	RecordExecution(ctx context.Context, tenantID uuid.UUID, executionID string) (*services.QuotaCheckResult, error)
	RecordComputeUsage(ctx context.Context, tenantID uuid.UUID, cpuTimeMs int) error
	GetQuotaStatus(ctx context.Context, tenantID uuid.UUID) (*services.RealtimeQuotaStatus, error)
	IsEnabled() bool
}

// ResourceUsage tracks resource consumption during function execution
type ResourceUsage struct {
	MaxMemoryMB    int
	MaxCPUTimeMs   int
	MemoryUsedMB   int
	CPUTimeUsedMs  int
	WallTimeUsedMs int
	Region         string // Geographic region where execution occurred
}

// ExecutionError represents an error during function execution with resource context
type ExecutionError struct {
	Err           error
	ResourceUsage *ResourceUsage
	TerminatedBy  string
}

func (e *ExecutionError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "execution error"
}

// ReplayVerificationResult contains the outcome of a replay verification
type ReplayVerificationResult struct {
	VerifiedAt       time.Time
	Status           VerificationStatus
	OutputMatches    bool
	OriginalDuration int
	ReplayedDuration int
	ReplayedOutput   json.RawMessage
	Error            string
	OriginalOutput   json.RawMessage
	OriginalMEG      *drecrypto.MEGResult
	ReplayMEG        *drecrypto.MEGResult
	OriginalRootHash string
	ReplayRootHash   string
	DriftCategory    capsule.DriftCategory
	ComponentDiff    []string
}

// VerificationStatus represents the status of replay verification
type VerificationStatus string

const (
	VerificationPending    VerificationStatus = "pending"
	VerificationVerified   VerificationStatus = "verified"
	VerificationMismatched VerificationStatus = "mismatched"
	VerificationFailed     VerificationStatus = "failed"
	VerificationSkipped    VerificationStatus = "skipped"
)

// ExecutionMetadata contains metadata about an execution for DRE
type ExecutionMetadata struct {
	ExecutionID     string
	FunctionID      string
	OwnerID         string
	CallerID        string
	NodeID          string
	Region          string
	Nonce           string
	ProtocolVersion string
}

// ExecutionOutcome represents the outcome of a function execution
type ExecutionOutcome string

const (
	OutcomeSuccess   ExecutionOutcome = "success"
	OutcomeError     ExecutionOutcome = "error"
	OutcomeTimeout   ExecutionOutcome = "timeout"
	OutcomeOOM       ExecutionOutcome = "oom"
	OutcomeCancelled ExecutionOutcome = "cancelled"
)
