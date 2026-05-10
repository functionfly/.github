package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/dre/antimanip"
	"github.com/functionfly/functionfly/internal/dre/capsule"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// verifyReplay re-executes a function with the same input and verifies the output matches.
// DRE 2.0: Uses MEG root hash comparison instead of raw output byte comparison.
func (h *Handler) verifyReplay(fnVersion *storage.RegistryFunctionVersion, originalInput json.RawMessage, originalOutput json.RawMessage, originalDuration int) *ReplayVerificationResult {
	result := &ReplayVerificationResult{
		Status:           VerificationPending,
		OriginalOutput:   originalOutput,
		OriginalDuration: originalDuration,
		VerifiedAt:       time.Now(),
	}

	// Define the execution function (same as in HandleExecute)
	executeFn := func() (json.RawMessage, error) {
		// Check if function has WASM binary (local execution)
		if len(fnVersion.WasmBinary) > 0 {
			// Execute using sandbox - all runtimes use the same path
			return executeLocally(fnVersion, originalInput)
		} else if fnVersion.BackendID != nil {
			// Execute via backend
			backend, err := h.BackendRepo.GetBackendByID(*fnVersion.BackendID)
			if err != nil {
				return nil, fmt.Errorf("backend not found: %w", err)
			}

			// Create execution URL
			execURL := fmt.Sprintf("%s/execute", strings.TrimSuffix(backend.URL, "/"))
			return executeOnBackend(execURL, string(originalInput), fnVersion.TimeoutMs)
		} else if fnVersion.DeploymentID != nil {
			// Execute via deployment (similar to playground)
			deployment, err := h.BackendRepo.GetActiveDeploymentForFunction(context.Background(), fnVersion.FunctionID)
			if err != nil || deployment == nil || deployment.DeployedURL == nil {
				return nil, fmt.Errorf("function is not deployed")
			}

			// Forward request to deployment
			execURL := *deployment.DeployedURL + "/execute"
			client := &http.Client{Timeout: time.Duration(fnVersion.TimeoutMs) * time.Millisecond}

			reqBody := map[string]interface{}{"input": originalInput}
			jsonBody, _ := json.Marshal(reqBody)

			req, err := http.NewRequest("POST", execURL, bytes.NewReader(jsonBody))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("deployment execution failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, fmt.Errorf("deployment returned status %d: %s", resp.StatusCode, string(body))
			}

			return json.RawMessage(body), nil
		} else {
			// Function uses lazy bundling or has no executable path - verification cannot be performed
			return nil, fmt.Errorf("function uses lazy bundling or has no executable path")
		}
	}

	// Execute the function for verification
	startTime := time.Now()
	replayedOutput, err := executeFn()
	result.ReplayedDuration = int(time.Since(startTime).Milliseconds())

	if err != nil {
		// Check if this is a non-executable environment issue (lazy bundling, etc.)
		// These are not actual verification failures - they're skipped verifications
		errStr := err.Error()
		if strings.Contains(errStr, "lazy bundling") || strings.Contains(errStr, "no executable path") || strings.Contains(errStr, "not executable") {
			result.Status = VerificationSkipped
			result.Error = errStr
			result.OutputMatches = false
		} else {
			result.Status = VerificationFailed
			result.Error = errStr
			result.OutputMatches = false
		}
		return result
	}

	result.ReplayedOutput = replayedOutput

	// DRE 2.0: Build MEG for both original and replay executions, compare root hashes
	nonce := fmt.Sprintf("replay-%d", time.Now().UnixNano())
	execMeta := ExecutionMetadata{
		ExecutionID:     uuid.New().String(),
		FunctionID:      fnVersion.FunctionID.String(),
		NodeID:          h.NodeID,
		Region:          h.Region,
		Nonce:           nonce,
		ProtocolVersion: "dre/1.0",
	}

	capsuleDesc := capsule.Default(execMeta.ExecutionID, "", "")

	// Build MEG for original output
	originalMEG, origErr := BuildMEGFromExecution(fnVersion, originalInput, originalOutput, nil, capsuleDesc, execMeta)
	// Build MEG for replay output (same capsule = same deterministic environment)
	replayMEG, replayErr := BuildMEGFromExecution(fnVersion, originalInput, replayedOutput, nil, capsuleDesc, execMeta)

	if origErr != nil || replayErr != nil {
		// Fall back to output byte comparison if MEG construction fails
		result.OutputMatches = outputsEqual(originalOutput, replayedOutput)
		if result.OutputMatches {
			result.Status = VerificationVerified
		} else {
			result.Status = VerificationFailed
			result.Error = "output mismatch: replay produced different result"
		}
		return result
	}

	result.OriginalMEG = originalMEG
	result.ReplayMEG = replayMEG
	result.OriginalRootHash = originalMEG.ExecutionRootHash
	result.ReplayRootHash = replayMEG.ExecutionRootHash

	// Compare MEG root hashes (DRE 2.0 verification)
	if originalMEG.ExecutionRootHash == replayMEG.ExecutionRootHash {
		result.OutputMatches = true
		result.Status = VerificationVerified

		// Update passport to mark as verified (async)
		go h.updatePassportVerified(fnVersion.FunctionID, replayMEG.ResourceHash)
	} else {
		result.OutputMatches = false
		result.Status = VerificationFailed

		// Classify the drift using the anti-manipulation detector
		detector := &antimanip.DriftDetector{}
		driftReport, _ := detector.Analyze(originalMEG, replayMEG)
		if driftReport != nil {
			result.DriftCategory = driftReport.Category
			result.ComponentDiff = driftReport.ComponentDiff
			result.Error = fmt.Sprintf("MEG root hash mismatch: drift category=%s", driftReport.Category)

			// Store drift report asynchronously
			go h.storeDriftReport(fnVersion, originalMEG.ExecutionRootHash, replayMEG.ExecutionRootHash, driftReport)
		} else {
			result.DriftCategory = capsule.DriftUnknown
			result.Error = fmt.Sprintf("MEG root hash mismatch: original=%s replay=%s",
				originalMEG.ExecutionRootHash[:16], replayMEG.ExecutionRootHash[:16])
		}
	}

	return result
}

// updatePassportVerified updates the passport after successful replay verification.
func (h *Handler) updatePassportVerified(functionID uuid.UUID, resourceHash string) {
	now := time.Now()
	update := storage.PassportUpdate{
		IncrementVerified: true,
		ResourceHash:      resourceHash,
		LastVerifiedAt:   &now,
	}
	if err := h.Repo.UpdatePassport(functionID, update); err != nil {
		fmt.Printf("DRE: failed to update passport verified: %v\n", err)
	}
}

// storeDriftReport persists a drift report and updates the function passport.
func (h *Handler) storeDriftReport(
	fnVersion *storage.RegistryFunctionVersion,
	originalRoot, replayRoot string,
	driftReport *capsule.DriftReport,
) {
	if driftReport == nil {
		return
	}

	componentDiffJSON, _ := json.Marshal(driftReport.ComponentDiff)

	record := &storage.DriftReportRecord{
		FunctionID:       fnVersion.FunctionID,
		Version:          fnVersion.Version,
		OriginalRootHash: originalRoot,
		ReplayRootHash:   replayRoot,
		DriftCategory:    string(driftReport.Category),
		ComponentDiff:    componentDiffJSON,
		TrustPenalty:     driftReport.TrustPenalty,
	}

	if err := h.Repo.StoreDriftReport(record); err != nil {
		// Log but don't fail — this is async
		fmt.Printf("DRE: failed to store drift report: %v\n", err)
	}
}
