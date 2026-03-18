// Package cert implements the FXCERT execution certificate protocol.
// This file implements replay certification for execution verification.
package cert

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

// ReplayRequest represents a request for replay certification.
type ReplayRequest struct {
	CertificateID    string                 `json:"certificate_id"`
	ExecutionID       string                 `json:"execution_id"`
	FunctionID        string                 `json:"function_id"`
	InputPayload      json.RawMessage        `json:"input_payload"`
	CapsuleDescriptor string                 `json:"capsule_descriptor"`
	RequestingNodeID  string                 `json:"requesting_node_id"`
	RequestingRegion  string                 `json:"requesting_region"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	RequestedAt       time.Time              `json:"requested_at"`
}

// ReplayResult contains the results of a replay execution.
type ReplayResult struct {
	ReplayID          string                 `json:"replay_id"`
	OriginalRootHash  string                 `json:"original_root_hash"`
	ReplayRootHash    string                 `json:"replay_root_hash"`
	RootsMatch        bool                   `json:"roots_match"`
	ReplayNodeID      string                 `json:"replay_node_id"`
	ReplayRegion      string                 `json:"replay_region"`
	ReplayTimestamp   time.Time              `json:"replay_timestamp"`
	ExecutionDuration time.Duration          `json:"execution_duration"`
	DriftReport       *ReplayDriftReport     `json:"drift_report,omitempty"`
	Signature         *Signature             `json:"signature,omitempty"`
}

// ReplayDriftReport contains details about any drift detected during replay.
type ReplayDriftReport struct {
	DriftDetected     bool     `json:"drift_detected"`
	DriftType         string   `json:"drift_type"` // "output", "timing", "environment", "deterministic"
	OriginalValue     string   `json:"original_value"`
	ReplayValue       string   `json:"replay_value"`
	Difference        string   `json:"difference"`
	Component         string   `json:"component"`
	Severity          string   `json:"severity"` // "minor", "major", "critical"
	Description       string   `json:"description"`
}

// ReplayService defines the interface for replay certification operations.
type ReplayService interface {
	// RequestReplay initiates a replay verification request.
	RequestReplay(ctx context.Context, req *ReplayRequest) (*ReplayRequest, error)

	// ExecuteReplay performs the replay execution and returns the result.
	ExecuteReplay(ctx context.Context, req *ReplayRequest) (*ReplayResult, error)

	// CertifyReplay adds replay certification to a certificate.
	CertifyReplay(ctx context.Context, cert *FXCert, result *ReplayResult, nodeKey ed25519.PrivateKey) (*FXCert, error)

	// VerifyReplayCertification verifies replay certification on a certificate.
	VerifyReplayCertification(ctx context.Context, cert *FXCert, nodePublicKey ed25519.PublicKey) (bool, error)
}

// DefaultReplayService implements ReplayService.
type DefaultReplayService struct {
	// Function executor for running replays
	executor FunctionExecutor
	// Drift detector for comparing executions
	driftDetector DriftDetector
	// Node key for signing replay results
	nodeKey ed25519.PrivateKey
	// Optional: retrieves original MEG from storage for comparison and drift analysis
	megProvider OriginalMEGProvider
}

// FunctionExecutor defines the interface for executing functions.
type FunctionExecutor interface {
	// Execute runs a function with the given input and returns the output.
	Execute(ctx context.Context, functionID string, input json.RawMessage, capsule string) (json.RawMessage, error)
}

// DriftDetector defines the interface for detecting execution drift.
type DriftDetector interface {
	// Analyze compares two MEG results and returns a drift report.
	Analyze(original, replay *drecrypto.MEGResult) (*ReplayDriftReport, error)
}

// OriginalMEGProvider supplies the stored MEG for an execution (e.g. from registry storage).
// Implementations typically call registry.GetMEGByExecutionID and convert MEGRecord to MEGResult.
type OriginalMEGProvider interface {
	GetOriginalMEG(ctx context.Context, executionID string) (*drecrypto.MEGResult, error)
}

// NewDefaultReplayService creates a new default replay service.
// megProvider may be nil; if set, ExecuteReplay will retrieve the original MEG from storage for comparison and drift analysis.
func NewDefaultReplayService(executor FunctionExecutor, driftDetector DriftDetector, nodeKey ed25519.PrivateKey, megProvider OriginalMEGProvider) *DefaultReplayService {
	return &DefaultReplayService{
		executor:      executor,
		driftDetector: driftDetector,
		nodeKey:       nodeKey,
		megProvider:   megProvider,
	}
}

// RequestReplay initiates a replay verification request.
func (s *DefaultReplayService) RequestReplay(ctx context.Context, req *ReplayRequest) (*ReplayRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("cert: nil replay request")
	}

	if req.ExecutionID == "" {
		return nil, fmt.Errorf("cert: execution_id is required")
	}

	if req.FunctionID == "" {
		return nil, fmt.Errorf("cert: function_id is required")
	}

	// Set request timestamp if not already set
	if req.RequestedAt.IsZero() {
		req.RequestedAt = time.Now().UTC()
	}

	return req, nil
}

// ExecuteReplay performs the replay execution.
func (s *DefaultReplayService) ExecuteReplay(ctx context.Context, req *ReplayRequest) (*ReplayResult, error) {
	if req == nil {
		return nil, fmt.Errorf("cert: nil replay request")
	}

	startTime := time.Now()

	// Execute the function replay
	replayOutput, err := s.executor.Execute(ctx, req.FunctionID, req.InputPayload, req.CapsuleDescriptor)
	if err != nil {
		return nil, fmt.Errorf("cert: replay execution failed: %w", err)
	}

	executionDuration := time.Since(startTime)

	// Build MEG from replay using full DRE/1.0 component set so root comparison and drift analysis match production.
	replayMEG, err := buildReplayMEG(req, replayOutput)
	if err != nil {
		return nil, fmt.Errorf("cert: build replay MEG: %w", err)
	}

	// Compare with original execution root hash: retrieve original MEG from storage when available
	var originalMEG *drecrypto.MEGResult
	if s.megProvider != nil && req.ExecutionID != "" {
		originalMEG, err = s.megProvider.GetOriginalMEG(ctx, req.ExecutionID)
		if err != nil {
			// Log but continue; we can still fall back to metadata
			originalMEG = nil
		}
	}

	var originalRoot string
	if originalMEG != nil {
		originalRoot = originalMEG.ExecutionRootHash
	} else if req.Metadata != nil {
		if s, ok := req.Metadata["original_root_hash"].(string); ok && s != "" {
			originalRoot = s
		}
	}

	rootsMatch := originalRoot != "" && originalRoot == replayMEG.ExecutionRootHash

	// Perform drift analysis if there's a mismatch
	var driftReport *ReplayDriftReport
	if !rootsMatch && s.driftDetector != nil {
		// Use stored original MEG when available for accurate drift detection; otherwise minimal stub
		if originalMEG == nil {
			originalMEG = &drecrypto.MEGResult{ExecutionRootHash: originalRoot}
		}
		drift, err := s.driftDetector.Analyze(originalMEG, replayMEG)
		if err == nil && drift != nil {
			driftReport = &ReplayDriftReport{
				DriftDetected: true,
				DriftType:     "output",
				Severity:      "major",
				Description:  "Execution produced different output during replay",
			}
		}
	}

	// Sign the replay result if node key is available
	var signature *Signature
	if s.nodeKey != nil {
		sigData := []byte(replayMEG.ExecutionRootHash + req.RequestingNodeID)
		sig := ed25519.Sign(s.nodeKey, sigData)
		pubKey := s.nodeKey.Public().(ed25519.PublicKey)

		signature = &Signature{
			Algorithm: "Ed25519",
			PublicKey: base64.StdEncoding.EncodeToString(pubKey),
			Signature: base64.StdEncoding.EncodeToString(sig),
		}
	}

	result := &ReplayResult{
		ReplayID:          fmt.Sprintf("replay_%d", time.Now().UnixNano()),
		OriginalRootHash:  originalRoot,
		ReplayRootHash:    replayMEG.ExecutionRootHash,
		RootsMatch:        rootsMatch,
		ReplayNodeID:      req.RequestingNodeID,
		ReplayRegion:      req.RequestingRegion,
		ReplayTimestamp:   time.Now().UTC(),
		ExecutionDuration: executionDuration,
		DriftReport:       driftReport,
		Signature:         signature,
	}

	return result, nil
}

// CertifyReplay adds replay certification to a certificate.
func (s *DefaultReplayService) CertifyReplay(ctx context.Context, cert *FXCert, result *ReplayResult, nodeKey ed25519.PrivateKey) (*FXCert, error) {
	if cert == nil {
		return nil, fmt.Errorf("cert: nil certificate")
	}

	if result == nil {
		return nil, fmt.Errorf("cert: nil replay result")
	}

	// Update the replay certification section
	cert.ReplayCert = &ReplayCertSection{
		ReplayRootHash:  result.ReplayRootHash,
		ReplayNodeID:    result.ReplayNodeID,
		ReplayTimestamp: result.ReplayTimestamp.Format(time.RFC3339),
		RootsMatch:      result.RootsMatch,
	}

	// Sign the certificate with replay node key if provided
	if nodeKey != nil {
		certHash := cert.Integrity.CertificateHash
		sigData := []byte(certHash + result.ReplayRootHash)
		sig := ed25519.Sign(nodeKey, sigData)
		pubKey := nodeKey.Public().(ed25519.PublicKey)

		// Add or update platform signature
		cert.Signatures.PlatformSignature = &Signature{
			Algorithm: "Ed25519",
			PublicKey: base64.StdEncoding.EncodeToString(pubKey),
			Signature: base64.StdEncoding.EncodeToString(sig),
		}
	}

	return cert, nil
}

// VerifyReplayCertification verifies replay certification on a certificate.
func (s *DefaultReplayService) VerifyReplayCertification(ctx context.Context, cert *FXCert, nodePublicKey ed25519.PublicKey) (bool, error) {
	if cert == nil {
		return false, fmt.Errorf("cert: nil certificate")
	}

	if cert.ReplayCert == nil {
		return false, fmt.Errorf("cert: no replay certification present")
	}

	// Verify roots match
	if !cert.ReplayCert.RootsMatch {
		return false, fmt.Errorf("cert: replay root hash does not match original")
	}

	// If there's a platform signature, verify it
	if cert.Signatures.PlatformSignature != nil && nodePublicKey != nil {
		sig, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.Signature)
		if err != nil {
			return false, fmt.Errorf("cert: decode platform signature: %w", err)
		}

		sigData := []byte(cert.Integrity.CertificateHash + cert.ReplayCert.ReplayRootHash)
		if !ed25519.Verify(nodePublicKey, sigData, sig) {
			return false, fmt.Errorf("cert: platform signature verification failed")
		}
	}

	return true, nil
}

// AddReplayCertification adds replay certification to a certificate directly.
func AddReplayCertification(cert *FXCert, result *ReplayResult) error {
	if cert == nil {
		return fmt.Errorf("cert: nil certificate")
	}

	if result == nil {
		return fmt.Errorf("cert: nil replay result")
	}

	cert.ReplayCert = &ReplayCertSection{
		ReplayRootHash:  result.ReplayRootHash,
		ReplayNodeID:    result.ReplayNodeID,
		ReplayTimestamp: result.ReplayTimestamp.Format(time.RFC3339),
		RootsMatch:      result.RootsMatch,
	}

	return nil
}

// HasReplayCertification returns true if the certificate has replay certification.
func HasReplayCertification(cert *FXCert) bool {
	return cert != nil && cert.ReplayCert != nil
}

// GetReplayInfo returns replay certification information for a certificate.
func GetReplayInfo(cert *FXCert) map[string]interface{} {
	if cert == nil || cert.ReplayCert == nil {
		return nil
	}

	info := map[string]interface{}{
		"has_replay_cert":    true,
		"replay_root_hash":   cert.ReplayCert.ReplayRootHash,
		"replay_node_id":     cert.ReplayCert.ReplayNodeID,
		"replay_timestamp":   cert.ReplayCert.ReplayTimestamp,
		"roots_match":        cert.ReplayCert.RootsMatch,
	}

	// Include signature info if present
	if cert.Signatures.PlatformSignature != nil {
		info["has_platform_signature"] = true
		info["signature_algorithm"] = cert.Signatures.PlatformSignature.Algorithm
	}

	return info
}

// buildReplayMEG constructs a full MEG from replay request and output using DRE/1.0 component ordering.
// Uses the same canonical shapes as the execution path so root and drift comparison are meaningful.
func buildReplayMEG(req *ReplayRequest, replayOutput json.RawMessage) (*drecrypto.MEGResult, error) {
	meta := req.Metadata
	getStr := func(key string) string {
		if meta == nil {
			return ""
		}
		if v, ok := meta[key]; ok {
			if s, _ := v.(string); s != "" {
				return s
			}
		}
		return ""
	}

	// Input: same shape as execution megInputPayload (args, caller_id, fx_uri, seed)
	inputPayload := map[string]interface{}{
		"args":      req.InputPayload,
		"caller_id": getStr("caller_id"),
		"fx_uri":    "fx://" + req.FunctionID,
		"seed":      getStr("nonce"),
	}

	// Environment: runtime_version + capsule (string or parsed JSON)
	var envCapsule interface{} = req.CapsuleDescriptor
	if len(req.CapsuleDescriptor) > 0 && (req.CapsuleDescriptor[0] == '{' || req.CapsuleDescriptor[0] == '[') {
		var parsed interface{}
		if err := json.Unmarshal([]byte(req.CapsuleDescriptor), &parsed); err == nil {
			envCapsule = parsed
		}
	}
	environmentData := map[string]interface{}{
		"runtime_version": getStr("runtime_version"),
		"capsule":        envCapsule,
	}

	// Output: return_value + exit_code
	outputPayload := map[string]interface{}{
		"return_value": replayOutput,
		"exit_code":    0,
	}

	// Metadata: execution_id, function_id, and standard fields
	timestamp := req.RequestedAt.Format(time.RFC3339)
	if timestamp == "" || req.RequestedAt.IsZero() {
		timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	metadataPayload := map[string]interface{}{
		"execution_id":     req.ExecutionID,
		"function_id":      req.FunctionID,
		"owner_id":         getStr("owner_id"),
		"caller_id":        getStr("caller_id"),
		"node_id":          req.RequestingNodeID,
		"region":           req.RequestingRegion,
		"nonce":            getStr("nonce"),
		"protocol_version": getStr("protocol_version"),
		"timestamp":        timestamp,
	}

	components := drecrypto.MEGComponents{
		InputPayload:    inputPayload,
		EnvironmentData: environmentData,
		Dependencies:    nil,
		TraceChunks:     nil,
		ResourceUsage:   nil,
		OutputPayload:   outputPayload,
		Metadata:        metadataPayload,
	}

	return drecrypto.BuildMEG(components)
}
