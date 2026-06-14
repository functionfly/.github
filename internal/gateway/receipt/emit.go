// Package receipt (under gateway) provides the protocol-agnostic receipt
// emission pipeline. Both MCP and A2A calls flow through Emit() to
// create a row in registry_executions_public.
//
// This package does NOT import the parent gateway package to avoid
// import cycles. Types that both packages need are defined here or
// passed as primitive parameters.
package receipt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	registrystorage "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/jaevor/go-nanoid"
	"github.com/sirupsen/logrus"
)

// Emitter is the single entry point for receipt emission. It is
// protocol-agnostic: both MCP and A2A calls flow through here.
type Emitter struct {
	registryRepo *registrystorage.RegistryRepository
	signer       []byte
	milestone    *Milestone
	logger       *logrus.Logger
}

// EmitterConfig configures the receipt emitter.
type EmitterConfig struct {
	Signer    []byte
	Milestone *Milestone
}

// NewEmitter creates a receipt emitter.
func NewEmitter(
	registryRepo *registrystorage.RegistryRepository,
	cfg EmitterConfig,
	logger *logrus.Logger,
) *Emitter {
	if logger == nil {
		logger = logrus.New()
	}
	return &Emitter{
		registryRepo: registryRepo,
		signer:       cfg.Signer,
		milestone:    cfg.Milestone,
		logger:       logger,
	}
}

// EmitRequest describes a receipt to emit.
type EmitRequest struct {
	Protocol      string          // "mcp" or "a2a"
	State         string          // "completed", "submitted", "working", etc.
	FunctionID    uuid.UUID       // zero for pure A2A peer-to-peer
	Version       string
	Inputs        json.RawMessage
	Outputs       json.RawMessage
	DurationMs    int
	Cached        bool
	FallbackChain []string
	ParentTaskID  *uuid.UUID      // set for delegation (P2.5)
	CallerTenantID string         // tenant UUID for milestone scoping
}

// EmitResult is what Emit returns.
type EmitResult struct {
	PublicID   string
	ReceiptSig string
}

// Emit creates a receipt row in registry_executions_public and triggers
// async side-effects (milestone check, DRE anchoring).
func (e *Emitter) Emit(ctx context.Context, req EmitRequest) (*EmitResult, error) {
	if e == nil || e.registryRepo == nil {
		return nil, nil
	}

	// Generate a nanoid public ID.
	nano, err := nanoid.Canonic()
	if err != nil {
		return nil, err
	}
	publicID := nano()

	// Default protocol/state.
	if req.Protocol == "" {
		req.Protocol = "mcp"
	}
	if req.State == "" {
		req.State = "completed"
	}

	// Compute HMAC signature.
	sig := HMACSignLocal(e.signer, publicID)

	// Ensure input/output are never nil.
	inputs := req.Inputs
	if len(inputs) == 0 {
		inputs = []byte("null")
	}
	outputs := req.Outputs
	if len(outputs) == 0 {
		outputs = []byte("null")
	}

	// Build the execution row.
	exec := &registrystorage.RegistryExecutionPublic{
		PublicID:      publicID,
		FunctionID:    req.FunctionID,
		Version:       req.Version,
		InputJSON:     inputs,
		OutputJSON:    outputs,
		DurationMs:    req.DurationMs,
		Cached:        req.Cached,
		Shareable:     true,
		Protocol:      req.Protocol,
		State:         req.State,
		ParentTaskID:  req.ParentTaskID,
		FallbackChain: req.FallbackChain,
	}

	// Insert via the existing registry repository.
	if err := e.registryRepo.CreateExecutionPublic(ctx, exec); err != nil {
		e.logger.WithError(err).WithField("public_id", publicID).Error("receipt emit: insert failed")
		return nil, err
	}

	// Async: milestone check (fire-and-forget).
	if e.milestone != nil && req.FunctionID != uuid.Nil {
		go func(fid uuid.UUID, pid string) {
			mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var tenantID *uuid.UUID
			if req.CallerTenantID != "" {
				tid, err := uuid.Parse(req.CallerTenantID)
				if err == nil {
					tenantID = &tid
				}
			}
			e.milestone.OnExecution(mctx, fid, tenantID, pid)
		}(req.FunctionID, publicID)
	}

	return &EmitResult{
		PublicID:   publicID,
		ReceiptSig: sig,
	}, nil
}

// HMACSignLocal is a local copy of gateway.HMACSign to avoid import cycles.
func HMACSignLocal(key []byte, payload string) string {
	if len(key) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
