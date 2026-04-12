package execution

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// updateTrustScoreV2 calculates and updates the Trust Score v2 for a function.
// This includes DRE 2.0 sub-scores from the execution passport.
// This is called asynchronously after successful replay verification.
func (h *Handler) updateTrustScoreV2(functionID uuid.UUID) {
	// Get DRE scores from the passport
	dreScores, err := h.Repo.GetDREScoresForTrust(functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Warn("Failed to get DRE scores for trust calculation")
		return
	}

	// If no passport exists yet, use default scores
	if dreScores == nil {
		dreScores = &registry.DREScores{
			DeterminismScore:          0,
			ReplayIntegrityScore:      0,
			PerformanceStabilityScore: 0,
			DriftScore:                1.0,
		}
	}

	// Convert to TrustMetricsV2 for the calculator
	metrics := &functionregistry.TrustMetricsV2{
		TrustMetrics: functionregistry.TrustMetrics{
			// These would be populated from the function rating in a full implementation
			SuccessRate:  1.0,
			P50LatencyMs: 0,
			P95LatencyMs: 0,
		},
		DeterminismScore:          dreScores.DeterminismScore,
		ReplayIntegrityScore:      dreScores.ReplayIntegrityScore,
		PerformanceStabilityScore: dreScores.PerformanceStabilityScore,
		DriftScore:                dreScores.DriftScore,
	}

	// Calculate Trust Score v2
	calc := functionregistry.NewTrustScoreCalculator()
	result := calc.CalculateV2(metrics)

	// Update the trust score in the database
	if err := h.Repo.UpdateTrustScoreV2(functionID, dreScores, result.TrustScoreV2); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Warn("Failed to update Trust Score v2")
		return
	}

	logrus.WithFields(logrus.Fields{
		"function_id":      functionID,
		"trust_score_v2":   result.TrustScoreV2,
		"determinism":      dreScores.DeterminismScore,
		"replay_integrity": dreScores.ReplayIntegrityScore,
		"performance":      dreScores.PerformanceStabilityScore,
		"drift":            dreScores.DriftScore,
	}).Info("Updated Trust Score v2")
}

// buildAndStoreMEG constructs the Merkle Execution Graph for a completed execution
// and stores the MEG record and FXCERT certificate asynchronously.
// This is called in a goroutine and must not block the HTTP response.
func (h *Handler) buildAndStoreMEG(
	fn *storage.RegistryFunction,
	fnVersion *storage.RegistryFunctionVersion,
	input json.RawMessage,
	output json.RawMessage,
	resourceUsage *ResourceUsage,
	durationMs int,
) {
	// Generate a nonce for this MEG construction
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	// Build execution metadata
	execMeta := ExecutionMetadata{
		ExecutionID:     uuid.New().String(),
		FunctionID:      fn.ID.String(),
		OwnerID:         "",
		CallerID:        "",
		NodeID:          h.NodeID,
		Region:          h.Region,
		Nonce:           nonce,
		ProtocolVersion: "dre/1.0",
	}
	if fn.OwnerUserID != nil {
		execMeta.OwnerID = fn.OwnerUserID.String()
	}

	// Create a default capsule descriptor
	capsuleDesc := capsule.Default(execMeta.ExecutionID, "", "")

	// Build the MEG
	megResult, err := BuildMEGFromExecution(fnVersion, input, output, resourceUsage, capsuleDesc, execMeta)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fn.ID,
			"version":     fnVersion.Version,
		}).Warn("DRE: Failed to build MEG for execution")
		return
	}

	// Get capsule descriptor hash
	capsuleHash, err := capsuleDesc.Hash()
	if err != nil {
		logrus.WithError(err).Warn("DRE: Failed to hash capsule descriptor")
		capsuleHash = ""
	}

	// Parse execution ID as UUID (generated above)
	execUUID, err := uuid.Parse(execMeta.ExecutionID)
	if err != nil {
		execUUID = uuid.New()
	}

	// Store MEG record
	megRecord := &storage.MEGRecord{
		ID:                    uuid.New(),
		ExecutionID:           execUUID,
		FunctionID:            fn.ID,
		Version:               fnVersion.Version,
		ExecutionRootHash:     megResult.ExecutionRootHash,
		InputHash:             megResult.InputHash,
		EnvironmentHash:       megResult.EnvironmentHash,
		DependencyHash:        megResult.DependencyHash,
		TraceHash:             megResult.TraceHash,
		ResourceHash:          megResult.ResourceHash,
		OutputHash:            megResult.OutputHash,
		MetadataHash:          megResult.MetadataHash,
		CapsuleDescriptorHash: capsuleHash,
		DeterminismTier:       capsuleDesc.DeterminismTier,
		ProtocolVersion:       "dre/1.0",
	}

	if err := h.Repo.StoreMEGRecord(megRecord); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id":         fn.ID,
			"execution_root_hash": megResult.ExecutionRootHash,
		}).Warn("DRE: Failed to store MEG record")
		return
	}

	// Generate FXCERT (standard level)
	certExec := drecert.ExecutionSection{
		ExecutionID:      execMeta.ExecutionID,
		FunctionID:       fmt.Sprintf("fx://%s/%s/%s", fn.Author, fn.Name, fnVersion.Version),
		OwnerID:          execMeta.OwnerID,
		CallerID:         execMeta.CallerID,
		NodeID:           h.NodeID,
		Region:           h.Region,
		TimestampVirtual: capsuleDesc.TimeSeed,
		TimestampRealUTC: time.Now().UTC().Format(time.RFC3339),
		ProtocolVersion:  "dre/1.0",
	}

	certCapsule := drecert.CapsuleSection{
		CapsuleDescriptorHash: capsuleHash,
		DeterminismTier:       capsuleDesc.DeterminismTier,
		ProtocolVersion:       capsuleDesc.ProtocolVersion,
	}

	certTrust := drecert.TrustSection{
		TrustScore:       0,
		DeterminismScore: 0,
	}

	// Generate certificate; sign with node key and optional platform key when configured
	cert, err := drecert.Generate(megResult, certExec, certCapsule, certTrust, drecert.CertLevelStandard, h.NodeKey, h.PlatformKey)
	if err != nil {
		logrus.WithError(err).Warn("DRE: Failed to generate FXCERT")
		return
	}

	// Marshal certificate to JSON
	certJSON, err := json.Marshal(cert)
	if err != nil {
		logrus.WithError(err).Warn("DRE: Failed to marshal FXCERT")
		return
	}

	// Store certificate
	execCert := &storage.ExecutionCertificate{
		ID:                uuid.New(),
		CertificateID:     cert.CertificateID,
		ExecutionID:       megRecord.ID, // Use MEG record ID as proxy for execution ID
		MEGRecordID:       megRecord.ID,
		FunctionID:          fn.ID,
		CertLevel:         string(drecert.CertLevelStandard),
		CertJSON:          certJSON,
		ExecutionRootHash: megResult.ExecutionRootHash,
		CertificateHash:   cert.Integrity.CertificateHash,
	}

	if err := h.Repo.StoreCertificate(execCert); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"certificate_id": cert.CertificateID,
		}).Warn("DRE: Failed to store FXCERT")
		return
	}

	// Update execution passport
	now := time.Now()
	passportUpdate := storage.PassportUpdate{
		IncrementTotal:        true,
		IncrementVerified:     false, // Will be set to true after replay verification
		CapsuleDescriptorHash: capsuleHash,
		LastVerifiedAt:        &now,
		ResourceHash:          megResult.ResourceHash, // For performance stability tracking
	}
	if err := h.Repo.UpdatePassport(fn.ID, passportUpdate); err != nil {
		logrus.WithError(err).WithField("function_id", fn.ID).Warn("DRE: Failed to update execution passport")
	}

	logrus.WithFields(logrus.Fields{
		"function_id":         fn.ID,
		"execution_root_hash": megResult.ExecutionRootHash,
		"certificate_id":      cert.CertificateID,
	}).Debug("DRE: MEG and certificate stored successfully")
}

// issueFXCERT determines if a function execution should issue an FXCERT.
// Returns true for deterministic functions or functionfly-authored functions.
func shouldIssueFXCERT(author string, fnVersion *storage.RegistryFunctionVersion) bool {
	return fnVersion.Deterministic || strings.EqualFold(author, "functionfly")
}
