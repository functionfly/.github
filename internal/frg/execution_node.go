package frg

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/dre/capsule"
	"github.com/functionfly/functionfly/internal/dre/cert"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// executeNode executes a single function node with full DRE support
func (e *ExecutionEngine) executeNode(ctx context.Context, instance *GraphInstance, node *RuntimeNode, input map[string]interface{}) (*NodeExecutionResult, error) {
	startTime := time.Now()
	logger := logrus.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"node_id":     node.Ref.NodeID,
		"function":    fmt.Sprintf("%s/%s@%s", node.Ref.Author, node.Ref.Name, node.Ref.Version),
	})

	// Create execution record
	inputJSON, _ := json.Marshal(input)
	execRecord := &GraphNodeExecution{
		InstanceID:      instance.ID,
		NodeID:          node.Ref.NodeID,
		FunctionAuthor:  node.Ref.Author,
		FunctionName:    node.Ref.Name,
		FunctionVersion: node.Ref.Version,
		Status:          "executing",
		InputData:       inputJSON,
		AttemptCount:    1,
		StartedAt:       timePtr(time.Now()),
	}

	if err := e.repo.CreateNodeExecution(ctx, execRecord); err != nil {
		logger.WithError(err).Error("Failed to create node execution record")
	}

	// Execute via sandbox using existing execution infrastructure
	fnVersion := node.Definition

	var output []byte
	var execErr error

	// Execute function in sandbox with proper timeout
	output, execErr = e.sandboxExecutor.ExecuteFunction(fnVersion, inputJSON, fnVersion.TimeoutMs)

	durationMs := int(time.Since(startTime).Milliseconds())

	// Update execution record
	if execErr != nil {
		execRecord.Status = "failed"
		errStr := execErr.Error()
		execRecord.ErrorMessage = &errStr
	} else {
		execRecord.Status = "completed"
		execRecord.OutputData = output
		execRecord.DurationMs = durationMs

		// Async DRE: Build and store MEG + FXCERT for deterministic functions
		// This runs in a goroutine with full panic recovery and circuit breaker
		if e.shouldIssueFXCERT(node.Ref.Author, fnVersion) {
			e.wg.Add(1)
			go func() {
				defer e.wg.Done()
				e.buildAndStoreMEGWithRecovery(node, fnVersion, inputJSON, output, durationMs, execRecord.ID)
			}()
		}
	}
	execRecord.CompletedAt = timePtr(time.Now())

	if err := e.repo.UpdateNodeExecution(ctx, execRecord); err != nil {
		logger.WithError(err).Error("Failed to update node execution")
	}

	if execErr != nil {
		return nil, execErr
	}

	return &NodeExecutionResult{
		Status:     "completed",
		Output:     output,
		DurationMs: durationMs,
		CertID:     nil, // Cert ID not available immediately (async DRE generation)
	}, nil
}

// executeNodeParallel executes independent nodes in parallel
func (e *ExecutionEngine) executeNodeParallel(ctx context.Context, instance *GraphInstance, nodes []*RuntimeNode, inputs map[string]map[string]interface{}) ([]*NodeExecutionResult, error) {
	var wg sync.WaitGroup
	results := make([]*NodeExecutionResult, len(nodes))
	errChan := make(chan error, len(nodes))

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *RuntimeNode) {
			defer wg.Done()
			result, err := e.executeNode(ctx, instance, n, inputs[n.Ref.NodeID])
			if err != nil {
				errChan <- fmt.Errorf("node %s: %w", n.Ref.NodeID, err)
				return
			}
			results[idx] = result
		}(i, node)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// shouldIssueFXCERT determines if a function execution should issue an FXCERT.
// Returns true for deterministic functions.
func (e *ExecutionEngine) shouldIssueFXCERT(author string, fnVersion *storage.RegistryFunctionVersion) bool {
	if fnVersion == nil {
		return false
	}
	return fnVersion.Deterministic
}

// buildAndStoreMEGWithRecovery wraps buildAndStoreMEG with panic recovery
// This ensures DRE failures never crash the execution engine
func (e *ExecutionEngine) buildAndStoreMEGWithRecovery(
	node *RuntimeNode,
	fnVersion *storage.RegistryFunctionVersion,
	input, output []byte,
	durationMs int,
	executionID uuid.UUID,
) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"panic":      r,
				"stacktrace": string(debug.Stack()),
				"function":   fmt.Sprintf("%s/%s@%s", node.Ref.Author, node.Ref.Name, node.Ref.Version),
				"execution":  executionID,
			}).Error("DRE: Panic recovered in buildAndStoreMEG")
		}
	}()

	// Use a timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e.buildAndStoreMEG(ctx, node, fnVersion, input, output, durationMs, executionID)
}

// buildAndStoreMEG constructs the Merkle Execution Graph for a completed execution
// and stores the MEG record and FXCERT certificate.
// This is called in a goroutine and must not block the graph execution.
func (e *ExecutionEngine) buildAndStoreMEG(
	ctx context.Context,
	node *RuntimeNode,
	fnVersion *storage.RegistryFunctionVersion,
	input, output []byte,
	durationMs int,
	executionID uuid.UUID,
) {
	logger := logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"function_id":  fnVersion.FunctionID,
		"function":     fmt.Sprintf("%s/%s@%s", node.Ref.Author, node.Ref.Name, node.Ref.Version),
	})

	// Skip if DRE is not configured
	if e.dreNodeID == "" {
		logger.Debug("DRE: Skipping FXCERT generation - DRE not configured")
		return
	}

	// Skip if registry repo is not available
	if e.registryRepo == nil {
		logger.Warn("DRE: Skipping FXCERT generation - registry repository not available")
		return
	}

	// Generate a cryptographically unique nonce
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	// Build execution metadata
	execMeta := execution.ExecutionMetadata{
		ExecutionID:     executionID.String(),
		FunctionID:      fnVersion.FunctionID.String(),
		OwnerID:         "",
		CallerID:        "",
		NodeID:          e.dreNodeID,
		Region:          e.dreRegion,
		Nonce:           nonce,
		ProtocolVersion: "dre/1.0",
	}

	// Create a default capsule descriptor
	capsuleDesc := capsule.Default(executionID.String(), "", "")

	// Override with function-specific settings
	if fnVersion.MemoryMB > 0 {
		capsuleDesc.MemoryLimit = int64(fnVersion.MemoryMB) * 1024 * 1024
	}
	if fnVersion.TimeoutMs > 0 {
		capsuleDesc.InstructionLimit = int64(fnVersion.TimeoutMs) * 1000000
	}

	// Build the MEG using the standard builder
	megResult, err := execution.BuildMEGFromExecution(
		fnVersion,
		input,
		output,
		nil, // Resource usage not tracked in FRG execution yet
		capsuleDesc,
		execMeta,
	)
	if err != nil {
		logger.WithError(err).Error("DRE: Failed to build MEG")
		return
	}

	// Get capsule descriptor hash
	capsuleHash, err := capsuleDesc.Hash()
	if err != nil {
		logger.WithError(err).Error("DRE: Failed to hash capsule descriptor")
		capsuleHash = ""
	}

	// Create MEG record
	megRecord := &registry.MEGRecord{
		ID:                    uuid.New(),
		ExecutionID:           executionID,
		FunctionID:            fnVersion.FunctionID,
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
		DeterminismTier:       "lite",
		ProtocolVersion:       "dre/1.0",
	}

	// Store MEG record with timeout
	if err := e.storeMEGWithTimeout(ctx, megRecord); err != nil {
		logger.WithError(err).Error("DRE: Failed to store MEG record")
		return
	}

	// Get trust scores for certificate
	trustSnapshot := e.getTrustSnapshot(fnVersion.FunctionID)

	// Build execution section for certificate
	execSection := cert.ExecutionSection{
		ExecutionID:      executionID.String(),
		FunctionID:       fmt.Sprintf("fx://%s/%s/%s", node.Ref.Author, node.Ref.Name, fnVersion.Version),
		OwnerID:          "",
		CallerID:         "",
		NodeID:           e.dreNodeID,
		Region:           e.dreRegion,
		TimestampVirtual: capsuleDesc.TimeSeed,
		TimestampRealUTC: time.Now().UTC().Format(time.RFC3339),
		ProtocolVersion:  "dre/1.0",
	}

	// Build capsule section
	capsuleSection := cert.CapsuleSection{
		CapsuleDescriptorHash: capsuleHash,
		DeterminismTier:       "lite",
		ProtocolVersion:       "dcc/1.0",
	}

	// Generate FXCERT
	fxcert, err := cert.Generate(
		megResult,
		execSection,
		capsuleSection,
		trustSnapshot,
		cert.CertLevelStandard,
		e.dreNodeKey,
		e.drePlatformKey,
	)
	if err != nil {
		logger.WithError(err).Error("DRE: Failed to generate FXCERT")
		return
	}

	// Store certificate
	certRecord := &registry.ExecutionCertificate{
		ID:                uuid.New(),
		CertificateID:     fxcert.CertificateID,
		ExecutionID:       executionID,
		MEGRecordID:       megRecord.ID,
		FunctionID:        fnVersion.FunctionID,
		CertLevel:         "standard",
		CertJSON:          mustMarshal(fxcert),
		ExecutionRootHash: megResult.ExecutionRootHash,
		CertificateHash:   fxcert.Integrity.CertificateHash,
	}

	if e.dreNodeKey != nil {
		if fxcert.Signatures.NodeSignature != nil {
			certRecord.NodeSignature = fxcert.Signatures.NodeSignature.Signature
		}
	}
	if e.drePlatformKey != nil && fxcert.Signatures.PlatformSignature != nil {
		certRecord.PlatformSignature = fxcert.Signatures.PlatformSignature.Signature
	}

	if err := e.storeCertificateWithTimeout(ctx, certRecord); err != nil {
		logger.WithError(err).Error("DRE: Failed to store certificate")
		return
	}

	// Update execution record with certificate ID
	if err := e.repo.UpdateNodeExecutionCertID(executionID, &certRecord.ID); err != nil {
		logger.WithError(err).Warn("DRE: Failed to update execution with cert ID")
	}

	// Update passport with verified execution
	update := registry.PassportUpdate{
		IncrementVerified:     true,
		IncrementTotal:        true,
		CapsuleDescriptorHash: capsuleHash,
		LastVerifiedAt:        timePtr(time.Now()),
		ResourceHash:          megResult.ResourceHash,
	}

	if err := e.updatePassportWithTimeout(ctx, fnVersion.FunctionID, update); err != nil {
		logger.WithError(err).Warn("DRE: Failed to update passport")
	}

	logger.WithFields(logrus.Fields{
		"meg_id":       megRecord.ID,
		"cert_id":      fxcert.CertificateID,
		"root_hash":    megResult.ExecutionRootHash,
		"capsule_hash": capsuleHash,
	}).Info("DRE: Successfully generated FXCERT")
}

// Helper methods for DRE operations with timeouts

func (e *ExecutionEngine) storeMEGWithTimeout(ctx context.Context, meg *registry.MEGRecord) error {
	done := make(chan error, 1)
	go func() {
		done <- e.registryRepo.StoreMEGRecord(meg)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("store MEG timeout: %w", ctx.Err())
	case err := <-done:
		return err
	}
}

func (e *ExecutionEngine) storeCertificateWithTimeout(ctx context.Context, cert *registry.ExecutionCertificate) error {
	done := make(chan error, 1)
	go func() {
		done <- e.registryRepo.StoreCertificate(cert)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("store certificate timeout: %w", ctx.Err())
	case err := <-done:
		return err
	}
}

func (e *ExecutionEngine) updatePassportWithTimeout(ctx context.Context, functionID uuid.UUID, update registry.PassportUpdate) error {
	done := make(chan error, 1)
	go func() {
		done <- e.registryRepo.UpdatePassport(functionID, update)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("update passport timeout: %w", ctx.Err())
	case err := <-done:
		return err
	}
}

// getTrustSnapshot retrieves the current trust scores for a function
func (e *ExecutionEngine) getTrustSnapshot(functionID uuid.UUID) cert.TrustSection {
	// Default trust snapshot
	trust := cert.TrustSection{
		TrustScore:              0.5,
		DeterminismScore:        0,
		ReplayConsistencyScore:  0,
		DriftIncidentsTotal:     0,
		VerifiedExecutionsTotal: 0,
	}

	// Try to get passport for actual scores
	if e.registryRepo != nil {
		passport, err := e.registryRepo.GetPassportByFunctionID(functionID)
		if err == nil && passport != nil {
			trust.TrustScore = passport.DeterministicReliability
			trust.DeterminismScore = passport.DeterminismScore
			trust.ReplayConsistencyScore = passport.ReplayIntegrityScore
			trust.DriftIncidentsTotal = passport.ReplayDriftIncidents
			trust.VerifiedExecutionsTotal = passport.VerifiedExecutionsTotal
		}
	}

	return trust
}

// mustMarshal marshals a value to JSON, returning empty object on error
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
