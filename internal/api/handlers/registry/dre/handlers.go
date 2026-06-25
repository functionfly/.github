// Package dre implements the DRE 2.0 API handlers for execution certificates,
// replay verification, execution passports, and divergence simulation.
package dre

import (
	"context"
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"unsafe"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DREStatsResponse is the response shape for the DRE summary endpoint.
type DREStatsResponse struct {
	FunctionID uuid.UUID       `json:"function_id"`
	Summary    DREStatsSummary `json:"summary"`
}

type DREStatsSummary struct {
	DeterminismScore        float64 `json:"determinism_score"`
	ReplayIntegrityScore    float64 `json:"replay_integrity_score"`
	VerifiedExecutionsTotal int64   `json:"verified_executions_total"`
	TotalExecutions         int64   `json:"total_executions"`
	ReplayDriftIncidents    int     `json:"replay_drift_incidents"`
	DriftScore              float64 `json:"drift_score"`
	DeterminismTier         string  `json:"determinism_tier"`
}

// DRERepository defines the subset of the registry repository used by DRE handlers.
type DRERepository interface {
	GetCertificateByID(certID string) (*registry.ExecutionCertificate, error)
	GetCertificateByExecutionID(executionID uuid.UUID) (*registry.ExecutionCertificate, error)
	GetCertificatesByFunctionID(functionID uuid.UUID, limit, offset int) ([]*registry.ExecutionCertificate, error)
	GetMEGByExecutionID(executionID uuid.UUID) (*registry.MEGRecord, error)
	GetMEGByExecutionRootHash(hash string) (*registry.MEGRecord, error)
	GetMEGRecordsByFunctionID(functionID uuid.UUID, limit, offset int, filters registry.MEGRecordFilters) ([]*registry.MEGRecord, int64, error)
	GetPassportByFunctionID(functionID uuid.UUID) (*registry.ExecutionPassport, error)
	GetOrCreatePassport(functionID uuid.UUID) (*registry.ExecutionPassport, error)
	GetDriftReportsByFunctionID(functionID uuid.UUID, limit, offset int) ([]*registry.DriftReportRecord, error)
	GetFunctionByAuthorName(ctx context.Context, author, name string) (*registry.RegistryFunction, error)
	GetFunctionByID(ctx context.Context, id uuid.UUID) (*registry.RegistryFunction, error)
	GetLatestFunctionVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error)
	GetFunctionVersion(functionID uuid.UUID, version string) (*registry.RegistryFunctionVersion, error)
	GetExecutionTimelineBuckets(functionID uuid.UUID, from, to time.Time, metric string) ([]registry.ExecutionTimelineBucket, error)
	UpdateCertificateAnchored(certID string, anchored bool, anchorChain, anchorTxHash, anchorMerkleRoot string, anchorBlockNumber int64, anchoredAt *time.Time) error
}

// Handler contains dependencies for DRE API handlers.
type Handler struct {
	repo             DRERepository
	anchoringService AnchorServicer
}

// AnchorServicer defines the interface for blockchain anchoring operations.
type AnchorServicer interface {
	Anchor(ctx context.Context, req *drecert.AnchorRequest) (*drecert.AnchorReceipt, error)
	IsConfigured() bool
	// Chains returns the chains that have both an RPC endpoint and a
	// contract address configured.
	Chains() []string
}

// EthereumAnchoringService exposes the Ethereum anchoring service for route initialization.
type EthereumAnchoringService = drecert.EthereumAnchoringService

// NewEthereumAnchoringService creates a new Ethereum anchoring service.
func NewEthereumAnchoringService() *EthereumAnchoringService {
	return drecert.NewEthereumAnchoringService(nil)
}

// NewHandler creates a new DRE handler.
func NewHandler(repo *registry.RegistryRepository) *Handler {
	return &Handler{repo: repo}
}

// NewHandlerWithAnchoring creates a new DRE handler with an anchoring service.
func NewHandlerWithAnchoring(repo DRERepository, anchoring AnchorServicer) *Handler {
	return &Handler{repo: repo, anchoringService: anchoring}
}

// NewHandlerFromRepo creates a new DRE handler from a DRERepository (for testing).
func NewHandlerFromRepo(repo DRERepository) *Handler {
	return &Handler{repo: repo}
}

// HandleGetCertificate returns the FXCERT for a specific execution certificate.
//
// GET /registry/{author}/{name}/cert/{cert_id}
func (h *Handler) HandleGetCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	certID := vars["cert_id"]

	if certID == "" {
		writeError(w, http.StatusBadRequest, "cert_id is required")
		return
	}

	cert, err := h.repo.GetCertificateByID(certID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get certificate: %v", err))
		return
	}
	if cert == nil {
		writeError(w, http.StatusNotFound, "certificate not found")
		return
	}

	// Parse the stored FXCERT JSON
	var fxcert drecert.FXCert
	if err := json.Unmarshal(cert.CertJSON, &fxcert); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to parse certificate")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"certificate_id":      cert.CertificateID,
		"cert_level":          cert.CertLevel,
		"execution_root_hash": cert.ExecutionRootHash,
		"certificate_hash":    cert.CertificateHash,
		"created_at":          cert.CreatedAt,
		"anchored":            cert.Anchored,
		"cert":                fxcert,
	})
}

// HandleVerifyCertificate verifies the authenticity and integrity of an FXCERT.
// It checks the certificate hash, signature chain, and anchoring status.
//
// POST /registry/{author}/{name}/cert/{cert_id}/verify
func (h *Handler) HandleVerifyCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	certID := vars["cert_id"]

	if certID == "" {
		writeError(w, http.StatusBadRequest, "cert_id is required")
		return
	}

	cert, err := h.repo.GetCertificateByID(certID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get certificate: %v", err))
		return
	}
	if cert == nil {
		writeError(w, http.StatusNotFound, "certificate not found")
		return
	}

	var fxcert drecert.FXCert
	if err := json.Unmarshal(cert.CertJSON, &fxcert); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to parse certificate")
		return
	}

	hashValid := false
	signaturesValid := false
	certValid := false
	verificationErrors := []string{}

	if cert.CertificateHash == "" {
		verificationErrors = append(verificationErrors, "certificate hash is empty")
	} else {
		hashValid = true
	}

	if fxcert.Signatures.NodeSignature == nil {
		verificationErrors = append(verificationErrors, "node signature is missing")
	} else {
		pubKeyBytes, err := base64.StdEncoding.DecodeString(fxcert.Signatures.NodeSignature.PublicKey)
		if err != nil {
			verificationErrors = append(verificationErrors, fmt.Sprintf("failed to decode node public key: %v", err))
		} else if len(pubKeyBytes) != ed25519.PublicKeySize {
			verificationErrors = append(verificationErrors, "invalid node public key size")
		} else {
			nodePublicKey := ed25519.PublicKey(pubKeyBytes)
			if err := drecert.Verify(&fxcert, nodePublicKey); err != nil {
				verificationErrors = append(verificationErrors, fmt.Sprintf("signature verification failed: %v", err))
			} else {
				signaturesValid = true
				certValid = true
			}
		}
	}

	if fxcert.Trust.VerifiedExecutionsTotal == 0 && fxcert.Trust.DriftIncidentsTotal > 0 {
		certValid = false
		verificationErrors = append(verificationErrors, "certificate has no verified executions but has drift incidents")
	}

	anchored := cert.Anchored

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"certificate_id":   cert.CertificateID,
		"cert_level":       cert.CertLevel,
		"certificate_hash": cert.CertificateHash,
		"anchored":         anchored,
		"verification": map[string]interface{}{
			"hash_valid":        hashValid,
			"signatures_valid":  signaturesValid,
			"anchored":          anchored,
			"cert_valid":        certValid,
			"verified_at":       time.Now().UTC().Format(time.RFC3339),
			"verification_errors": verificationErrors,
		},
		"trust": map[string]interface{}{
			"trust_score":               fxcert.Trust.TrustScore,
			"determinism_score":         fxcert.Trust.DeterminismScore,
			"replay_consistency_score":  fxcert.Trust.ReplayConsistencyScore,
			"drift_incidents_total":     fxcert.Trust.DriftIncidentsTotal,
			"verified_executions_total": fxcert.Trust.VerifiedExecutionsTotal,
		},
	})
}


// HandleGetAnchoringStatus returns the operational status of the DRE
// anchoring service so operators can verify configuration without inspecting
// environment variables. It is safe to call publicly — it exposes no
// secrets, only which chains are ready and whether the signing key is set.
//
// GET /dre/anchoring/status
func (h *Handler) HandleGetAnchoringStatus(w http.ResponseWriter, r *http.Request) {
	enabled := h.anchoringService != nil && h.anchoringService.IsConfigured()
	chains := []string{}
	if h.anchoringService != nil {
		chains = h.anchoringService.Chains()
	}
	defaultChain := ""
	if len(chains) > 0 {
		defaultChain = chains[0]
	} else {
		defaultChain = drecert.DefaultChain
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":        enabled,
		"chains":         chains,
		"default_chain":  defaultChain,
		"supported_chains": drecert.SupportedChains,
		"message":        anchoringStatusMessage(enabled, len(chains)),
	})
}

func anchoringStatusMessage(enabled bool, chainCount int) string {
	if enabled {
		return fmt.Sprintf("DRE anchoring is enabled for %d chain(s)", chainCount)
	}
	return "DRE anchoring is not configured. Set ANCHOR_SIGNING_KEY, ANCHOR_RPC_<CHAIN>, and ANCHOR_CONTRACT_<CHAIN> to enable."
}

// HandleAnchorCertificate anchors a certificate to the blockchain.
//
// POST /registry/{author}/{name}/cert/{cert_id}/anchor
// Body: {"chain": "base"} (optional, defaults to "base")
func (h *Handler) HandleAnchorCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	certID := vars["cert_id"]

	if certID == "" {
		writeError(w, http.StatusBadRequest, "cert_id is required")
		return
	}

	cert, err := h.repo.GetCertificateByID(certID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get certificate: %v", err))
		return
	}
	if cert == nil {
		writeError(w, http.StatusNotFound, "certificate not found")
		return
	}

	if cert.Anchored {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"certificate_id": cert.CertificateID,
			"anchored":       true,
			"message":        "certificate is already anchored",
		})
		return
	}

	// Check if DRE blockchain anchoring is enabled via env var
	if os.Getenv("DRE_BLOCKCHAIN_ANCHORING_ENABLED") != "true" {
		writeError(w, http.StatusServiceUnavailable, "DRE blockchain anchoring is not enabled. Set DRE_BLOCKCHAIN_ANCHORING_ENABLED=true to enable.")
		return
	}

	// Check if anchoring service is configured
	if h.anchoringService == nil || !h.anchoringService.IsConfigured() {
		writeError(w, http.StatusServiceUnavailable, "anchoring requires an HSM (Hardware Security Module) with a configured signing key. To enable blockchain anchoring, set the following environment variables: ANCHOR_SIGNING_KEY, ANCHOR_RPC_<CHAIN>, ANCHOR_CONTRACT_<CHAIN>. See documentation for setup instructions.")
		return
	}

	// Parse chain from request body (optional, defaults to "base")
	chain := drecert.DefaultChain
	if r.ContentLength > 0 {
		var reqBody struct {
			Chain string `json:"chain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil && reqBody.Chain != "" {
			chain = reqBody.Chain
		}
	}

	// Validate chain
	if !drecert.IsChainSupported(chain) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported chain: %s", chain))
		return
	}

	// Submit to blockchain
	ctx := r.Context()
	receipt, err := h.anchoringService.Anchor(ctx, &drecert.AnchorRequest{
		CertificateID:     certID,
		ExecutionRootHash: cert.ExecutionRootHash,
		Chain:             chain,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("anchoring failed: %v", err))
		return
	}

	// Persist anchoring state to DB
	now := time.Now()
	upErr := h.repo.UpdateCertificateAnchored(
		certID,
		true,
		receipt.Chain,
		receipt.TxHash,
		receipt.MerkleRoot,
		receipt.BlockNumber,
		&now,
	)
	if upErr != nil {
		logrus.WithError(upErr).Warn("dre: failed to persist anchoring state")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"certificate_id":     cert.CertificateID,
		"anchored":           true,
		"anchor_chain":       receipt.Chain,
		"anchor_tx_hash":     receipt.TxHash,
		"anchor_block":       receipt.BlockNumber,
		"anchor_merkle_root": receipt.MerkleRoot,
		"anchored_at":        receipt.AnchoredAt.Format(time.RFC3339),
	})
}

// HandleListCertificates lists certificates for a function (paginated).
//
// GET /registry/{author}/{name}/certs
func (h *Handler) HandleListCertificates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Parse pagination params
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	certs, err := h.repo.GetCertificatesByFunctionID(fn.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list certificates: %v", err))
		return
	}

	// Build response (omit full cert JSON for list view)
	items := make([]map[string]interface{}, len(certs))
	for i, cert := range certs {
		items[i] = map[string]interface{}{
			"certificate_id":      cert.CertificateID,
			"cert_level":          cert.CertLevel,
			"execution_root_hash": cert.ExecutionRootHash,
			"certificate_hash":    cert.CertificateHash,
			"anchored":            cert.Anchored,
			"created_at":          cert.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function": fmt.Sprintf("fx://%s/%s", author, name),
		"certs":    items,
		"limit":    limit,
		"offset":   offset,
	})
}

// rootsMatch performs a timing-safe comparison of two root hashes.
// Returns true if both hashes are non-empty and equal.
func rootsMatch(hash1, hash2 string) bool {
	if hash1 == "" || hash2 == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) == 1
}

// HandleReplay returns the replay verification status for a specific execution.
// This endpoint provides the stored MEG record with verification status.
// Note: This is a status-check endpoint, not an actual replay trigger.
//
// POST /registry/{author}/{name}/replay/{execution_id}
func (h *Handler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	executionIDStr := vars["execution_id"]

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	// Get MEG record for this execution
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid execution_id")
		return
	}

	megRecord, err := h.repo.GetMEGByExecutionID(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get MEG record: %v", err))
		return
	}
	if megRecord == nil {
		writeError(w, http.StatusNotFound, "no MEG record found for this execution")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":            fmt.Sprintf("fx://%s/%s", author, name),
		"function_id":         fn.ID,
		"execution_id":        executionIDStr,
		"execution_root_hash": megRecord.ExecutionRootHash,
		"replay_root_hash":    megRecord.ReplayRootHash,
		"replay_verified_at":  megRecord.ReplayVerifiedAt,
		"determinism_tier":   megRecord.DeterminismTier,
		"protocol_version":    megRecord.ProtocolVersion,
		"component_hashes": map[string]string{
			"input":       megRecord.InputHash,
			"environment": megRecord.EnvironmentHash,
			"dependency":  megRecord.DependencyHash,
			"trace":       megRecord.TraceHash,
			"resource":    megRecord.ResourceHash,
			"output":      megRecord.OutputHash,
			"metadata":    megRecord.MetadataHash,
		},
		"roots_match":      rootsMatch(megRecord.ReplayRootHash, megRecord.ExecutionRootHash),
		"replay_status":    "status_check",
		"replay_note":      "This endpoint returns stored MEG record status. Use POST /replay/{execution_id}/trigger to trigger actual replay.",
	})
}

// HandleGetPassport returns the Execution Passport for a function.
// The passport contains public determinism statistics for the marketplace.
//
// GET /registry/{author}/{name}/passport
func (h *Handler) HandleGetPassport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	passport, err := h.repo.GetPassportByFunctionID(fn.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get passport: %v", err))
		return
	}

	if passport == nil {
		// Return empty passport for functions with no DRE data yet
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"function": fmt.Sprintf("fx://%s/%s", author, name),
			"passport": map[string]interface{}{
				"deterministic_reliability":   0,
				"replay_drift_incidents":      0,
				"verified_executions_total":   0,
				"total_executions":            0,
				"determinism_score":           0,
				"replay_integrity_score":      0,
				"performance_stability_score": 0,
				"drift_score":                 1.0,
				"capsule_version":             "dcc/1.0",
				"determinism_tier":            "full",
				"last_verified_at":            nil,
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function": fmt.Sprintf("fx://%s/%s", author, name),
		"passport": map[string]interface{}{
			"deterministic_reliability":   passport.DeterministicReliability,
			"replay_drift_incidents":      passport.ReplayDriftIncidents,
			"verified_executions_total":   passport.VerifiedExecutionsTotal,
			"total_executions":            passport.TotalExecutions,
			"determinism_score":           passport.DeterminismScore,
			"replay_integrity_score":      passport.ReplayIntegrityScore,
			"performance_stability_score": passport.PerformanceStabilityScore,
			"drift_score":                 passport.DriftScore,
			"capsule_version":             "dcc/1.0",
			"determinism_tier":            "full",
			"last_verified_at":            passport.LastVerifiedAt,
		},
	})
}

// HandleGetPassportPublic returns a limited public subset of the Execution Passport
// suitable for marketplace display without exposing sensitive internal details.
//
// GET /registry/{author}/{name}/passport/public
func (h *Handler) HandleGetPassportPublic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	passport, err := h.repo.GetPassportByFunctionID(fn.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get passport: %v", err))
		return
	}

	if passport == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"function": fmt.Sprintf("fx://%s/%s", author, name),
			"passport": map[string]interface{}{
				"determinism_score":         0,
				"replay_integrity_score":    0,
				"verified_executions_total": 0,
				"determinism_tier":          "unknown",
				"capsule_version":           "dcc/1.0",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function": fmt.Sprintf("fx://%s/%s", author, name),
		"passport": map[string]interface{}{
			"determinism_score":           passport.DeterminismScore,
			"replay_integrity_score":      passport.ReplayIntegrityScore,
			"performance_stability_score": passport.PerformanceStabilityScore,
			"verified_executions_total":   passport.VerifiedExecutionsTotal,
			"total_executions":            passport.TotalExecutions,
			"determinism_tier":            "full",
			"capsule_version":             "dcc/1.0",
		},
	})
}

// DivergenceSimulationRequest is the request body for divergence simulation.
type DivergenceSimulationRequest struct {
	MemoryLimit    int64           `json:"memory_limit"`
	RuntimeVersion string          `json:"runtime_version"`
	Region         string          `json:"region"`
	Input          json.RawMessage `json:"input"`   // optional; default {} for re-execution
	Version        string          `json:"version"` // optional; default latest
}

// HandleDivergenceSimulation simulates replay under modified constraints.
// This is used to test how a function behaves under different execution environments.
//
// POST /registry/{author}/{name}/diverge
func (h *Handler) HandleDivergenceSimulation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	var req DivergenceSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	// Get function version (latest or requested)
	var fnVersion *registry.RegistryFunctionVersion
	if req.Version != "" {
		fnVersion, err = h.repo.GetFunctionVersion(fn.ID, req.Version)
	} else {
		fnVersion, err = h.repo.GetLatestFunctionVersion(fn.ID)
	}
	if err != nil || fnVersion == nil {
		writeError(w, http.StatusNotFound, "function version not found")
		return
	}

	// Default input for re-execution
	input := req.Input
	if input == nil {
		input = json.RawMessage(`{}`)
	}

	// Build a modified capsule descriptor with the requested constraints
	baseExecID := uuid.New().String()
	modifiedCapsule := capsule.Default(baseExecID, "", "")

	if req.MemoryLimit > 0 {
		modifiedCapsule.MemoryLimit = req.MemoryLimit
	}
	if req.RuntimeVersion != "" {
		modifiedCapsule.RuntimeVersion = req.RuntimeVersion
	}

	// Compute the modified capsule hash
	modifiedHash, err := modifiedCapsule.Hash()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash modified capsule")
		return
	}

	memoryMB := fnVersion.MemoryMB
	if req.MemoryLimit > 0 {
		memoryMB = int(req.MemoryLimit)
	}
	maxCPUTimeMs := fnVersion.TimeoutMs

	// Re-execute the function with the modified constraints when the function has WASM
	var simulatedMEG *drecrypto.MEGResult
	reExecuted := false
	reExecError := ""
	simulationMode := "simulated"

	if len(fnVersion.WasmBinary) > 0 {
		// storage.RegistryFunctionVersion is an alias for registry.RegistryFunctionVersion; same underlying type
		fnVersionStorage := (*storage.RegistryFunctionVersion)(unsafe.Pointer(fnVersion))
		output, execErr := execution.ExecuteLocally(fnVersionStorage, input, memoryMB, maxCPUTimeMs)
		if execErr == nil {
			reExecuted = true
			simulationMode = "computed"
			inputBytes := input
			if len(inputBytes) == 0 {
				inputBytes = []byte("{}")
			}
			outputBytes := output
			if outputBytes == nil {
				outputBytes = []byte("null")
			}
			simulatedMEG = &drecrypto.MEGResult{
				ExecutionRootHash: drecrypto.HashString(drecrypto.TagMeta, []byte(modifiedHash+baseExecID)),
				InputHash:         drecrypto.HashString(drecrypto.TagInput, inputBytes),
				EnvironmentHash:   modifiedHash,
				DependencyHash:    drecrypto.HashString(drecrypto.TagDeps, []byte("[]")),
				ResourceHash:      drecrypto.HashString(drecrypto.TagResource, []byte("{}")),
				OutputHash:        drecrypto.HashString(drecrypto.TagOutput, outputBytes),
				MetadataHash:      drecrypto.HashString(drecrypto.TagMeta, []byte(baseExecID)),
			}
		} else {
			reExecError = execErr.Error()
		}
	}

	// Fallback to simulated MEG when re-execution was not run or failed
	if simulatedMEG == nil {
		simulationMode = "simulated"
		simulatedMEG = &drecrypto.MEGResult{
			ExecutionRootHash: drecrypto.HashString(drecrypto.TagMeta, []byte(modifiedHash+baseExecID)),
			InputHash:         drecrypto.HashString(drecrypto.TagInput, []byte("simulated")),
			EnvironmentHash:   modifiedHash,
			DependencyHash:    drecrypto.HashString(drecrypto.TagDeps, []byte("[]")),
			ResourceHash:      drecrypto.HashString(drecrypto.TagResource, []byte("{}")),
			OutputHash:        drecrypto.HashString(drecrypto.TagOutput, []byte("simulated")),
			MetadataHash:      drecrypto.HashString(drecrypto.TagMeta, []byte(baseExecID)),
		}
	}

	simulationPayload := map[string]interface{}{
		"modified_capsule_hash": modifiedHash,
		"simulated_root_hash":   simulatedMEG.ExecutionRootHash,
		"simulation_mode":        simulationMode,
		"constraints": map[string]interface{}{
			"memory_limit":    req.MemoryLimit,
			"runtime_version": req.RuntimeVersion,
			"region":          req.Region,
		},
		"simulated_at": time.Now().UTC().Format(time.RFC3339),
		"re_executed":  reExecuted,
	}
	if reExecError != "" {
		simulationPayload["re_execution_skipped"] = true
		simulationPayload["re_execution_error"] = reExecError
	} else if reExecuted {
		simulationPayload["note"] = "Function was re-executed with the modified constraints; MEG hashes reflect actual execution."
	} else {
		simulationPayload["note"] = "Simulated divergence (no WASM binary or re-execution skipped)."
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":    fmt.Sprintf("fx://%s/%s", author, name),
		"function_id": fn.ID,
		"version":     fnVersion.Version,
		"simulation":  simulationPayload,
	})
}

// HandleListExecutions lists MEG records (executions) for a function with pagination and filters.
//
// GET /registry/{author}/{name}/executions
func (h *Handler) HandleListExecutions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	// Parse pagination params
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// Parse filters
	filters := registry.MEGRecordFilters{
		Version:      r.URL.Query().Get("version"),
		VerifiedOnly: r.URL.Query().Get("verified_only") == "true",
	}

	// Parse date filters
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filters.From = &from
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filters.To = &to
		}
	}

	// Get MEG records
	records, total, err := h.repo.GetMEGRecordsByFunctionID(fn.ID, limit, offset, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list executions: %v", err))
		return
	}

	// Build response
	executions := make([]map[string]interface{}, len(records))
	for i, rec := range records {
		rootsMatch := rec.ReplayRootHash != "" && timingSafeHashEqual(rec.ReplayRootHash, rec.ExecutionRootHash)
		executions[i] = map[string]interface{}{
			"execution_id":        rec.ExecutionID.String(),
			"execution_root_hash": rec.ExecutionRootHash,
			"version":             rec.Version,
			"created_at":          rec.CreatedAt,
			"replay_verified":     rec.ReplayVerifiedAt != nil,
			"roots_match":         rootsMatch,
			"determinism_tier":    rec.DeterminismTier,
			"protocol_version":    rec.ProtocolVersion,
			"component_hashes": map[string]string{
				"input":       rec.InputHash,
				"output":      rec.OutputHash,
				"environment": rec.EnvironmentHash,
				"dependency":  rec.DependencyHash,
				"trace":       rec.TraceHash,
				"resource":    rec.ResourceHash,
				"metadata":    rec.MetadataHash,
			},
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":   fmt.Sprintf("fx://%s/%s", author, name),
		"executions": executions,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// HandleGetExecution returns detailed information about a specific execution.
//
// GET /registry/{author}/{name}/executions/{execution_id}
func (h *Handler) HandleGetExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	executionIDStr := vars["execution_id"]

	// Get function
	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	// Parse execution ID
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid execution_id")
		return
	}

	// Get MEG record
	rec, err := h.repo.GetMEGByExecutionID(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get execution: %v", err))
		return
	}
	if rec == nil {
		writeError(w, http.StatusNotFound, "execution not found")
		return
	}

	// Verify this execution belongs to the requested function
	if rec.FunctionID != fn.ID {
		writeError(w, http.StatusNotFound, "execution not found for this function")
		return
	}

	// Get associated certificate if available (by execution ID for correct lookup)
	cert, certErr := h.repo.GetCertificateByExecutionID(executionID)
	var certInfo map[string]interface{}
	var trustInfo map[string]interface{}
	if certErr == nil && cert != nil {
		certInfo = map[string]interface{}{
			"certificate_id":   cert.CertificateID,
			"cert_level":       cert.CertLevel,
			"certificate_hash": cert.CertificateHash,
			"created_at":       cert.CreatedAt,
			"anchored":         cert.Anchored,
		}
		// Include trust section from FXCERT when present
		var fxcert drecert.FXCert
		if err := json.Unmarshal(cert.CertJSON, &fxcert); err == nil {
			trustInfo = map[string]interface{}{
				"trust_score":               fxcert.Trust.TrustScore,
				"determinism_score":         fxcert.Trust.DeterminismScore,
				"replay_consistency_score":  fxcert.Trust.ReplayConsistencyScore,
				"drift_incidents_total":     fxcert.Trust.DriftIncidentsTotal,
				"verified_executions_total": fxcert.Trust.VerifiedExecutionsTotal,
			}
		}
	}

	execPayload := map[string]interface{}{
		"execution_id":        rec.ExecutionID.String(),
		"execution_root_hash": rec.ExecutionRootHash,
		"version":             rec.Version,
		"created_at":          rec.CreatedAt,
		"determinism_tier":    rec.DeterminismTier,
		"protocol_version":    rec.ProtocolVersion,
		"replay_verified_at":  rec.ReplayVerifiedAt,
		"replay_root_hash":    rec.ReplayRootHash,
		"replay_node_id":      rec.ReplayNodeID,
		"roots_match":         rec.ReplayRootHash != "" && timingSafeHashEqual(rec.ReplayRootHash, rec.ExecutionRootHash),
		"component_hashes": map[string]string{
			"input":       rec.InputHash,
			"output":      rec.OutputHash,
			"environment": rec.EnvironmentHash,
			"dependency":  rec.DependencyHash,
			"trace":       rec.TraceHash,
			"resource":    rec.ResourceHash,
			"metadata":    rec.MetadataHash,
		},
		"certificate": certInfo,
	}
	if trustInfo != nil {
		execPayload["trust"] = trustInfo
	}

	response := map[string]interface{}{
		"execution": execPayload,
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleGetExecutionByHash returns execution details by execution_root_hash (for conversation execution refs).
//
// GET /registry/{author}/{name}/executions/by-hash?execution_root_hash=0x...
func (h *Handler) HandleGetExecutionByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	hash := r.URL.Query().Get("execution_root_hash")
	if hash == "" {
		writeError(w, http.StatusBadRequest, "execution_root_hash query parameter is required")
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	rec, err := h.repo.GetMEGByExecutionRootHash(hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get execution: %v", err))
		return
	}
	if rec == nil {
		writeError(w, http.StatusNotFound, "execution not found")
		return
	}
	if rec.FunctionID != fn.ID {
		writeError(w, http.StatusNotFound, "execution not found for this function")
		return
	}

	executionID := rec.ExecutionID
	cert, certErr := h.repo.GetCertificateByExecutionID(executionID)
	var certInfo map[string]interface{}
	var trustInfo map[string]interface{}
	if certErr == nil && cert != nil {
		certInfo = map[string]interface{}{
			"certificate_id":   cert.CertificateID,
			"cert_level":       cert.CertLevel,
			"certificate_hash": cert.CertificateHash,
			"created_at":       cert.CreatedAt,
			"anchored":         cert.Anchored,
		}
		var fxcert drecert.FXCert
		if err := json.Unmarshal(cert.CertJSON, &fxcert); err == nil {
			trustInfo = map[string]interface{}{
				"trust_score":               fxcert.Trust.TrustScore,
				"determinism_score":         fxcert.Trust.DeterminismScore,
				"replay_consistency_score":  fxcert.Trust.ReplayConsistencyScore,
				"drift_incidents_total":     fxcert.Trust.DriftIncidentsTotal,
				"verified_executions_total": fxcert.Trust.VerifiedExecutionsTotal,
			}
		}
	}

	execPayload := map[string]interface{}{
		"execution_id":        rec.ExecutionID.String(),
		"execution_root_hash": rec.ExecutionRootHash,
		"version":             rec.Version,
		"created_at":          rec.CreatedAt,
		"determinism_tier":    rec.DeterminismTier,
		"protocol_version":    rec.ProtocolVersion,
		"replay_verified_at":  rec.ReplayVerifiedAt,
		"replay_root_hash":    rec.ReplayRootHash,
		"replay_node_id":      rec.ReplayNodeID,
		"roots_match":         rec.ReplayRootHash != "" && timingSafeHashEqual(rec.ReplayRootHash, rec.ExecutionRootHash),
		"component_hashes": map[string]string{
			"input":       rec.InputHash,
			"output":      rec.OutputHash,
			"environment": rec.EnvironmentHash,
			"dependency":  rec.DependencyHash,
			"trace":       rec.TraceHash,
			"resource":    rec.ResourceHash,
			"metadata":    rec.MetadataHash,
		},
		"certificate": certInfo,
	}
	if trustInfo != nil {
		execPayload["trust"] = trustInfo
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"execution": execPayload})
}

// HandleGetExecutionTimeline returns time-bucketed metrics for executions (for conversation overlay).
// GET /registry/{author}/{name}/executions/timeline?from=&to=&metric=latency|errors|trust
func (h *Handler) HandleGetExecutionTimeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "latency"
	}
	if metric != "latency" && metric != "errors" && metric != "trust" {
		metric = "latency"
	}

	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -7)
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t.UTC()
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.UTC()
		}
	}
	if from.After(to) {
		from, to = to, from
	}

	buckets, err := h.repo.GetExecutionTimelineBuckets(fn.ID, from, to, metric)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("timeline: %v", err))
		return
	}

	bucketMaps := make([]map[string]interface{}, 0, len(buckets))
	for _, b := range buckets {
		bucketMaps = append(bucketMaps, map[string]interface{}{
			"bucket":       b.Bucket,
			"value":        b.Value,
			"sample_count": b.SampleCount,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":    fmt.Sprintf("fx://%s/%s", author, name),
		"function_id": fn.ID,
		"metric":      metric,
		"from":        from.Format("2006-01-02"),
		"to":          to.Format("2006-01-02"),
		"buckets":     bucketMaps,
		"insight":     "",
	})
}

// HandleListDriftReports lists drift reports for a function (paginated).
//
// GET /registry/{author}/{name}/drift-reports
func (h *Handler) HandleListDriftReports(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	reports, err := h.repo.GetDriftReportsByFunctionID(fn.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list drift reports: %v", err))
		return
	}

	items := make([]map[string]interface{}, len(reports))
	for i, report := range reports {
		items[i] = map[string]interface{}{
			"id":                 report.ID.String(),
			"execution_id":       report.ExecutionID.String(),
			"version":            report.Version,
			"drift_category":     report.DriftCategory,
			"original_root_hash": report.OriginalRootHash,
			"replay_root_hash":   report.ReplayRootHash,
			"trust_penalty":      report.TrustPenalty,
			"detected_at":        report.DetectedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function": fmt.Sprintf("fx://%s/%s", author, name),
		"reports":  items,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleGetDRESummary returns a high-level DRE summary for a function.
//
// GET /registry/{author}/{name}/dre-stats
func (h *Handler) HandleGetDRESummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	passport, err := h.repo.GetPassportByFunctionID(fn.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get passport: %v", err))
		return
	}

	if passport == nil {
		writeJSON(w, http.StatusOK, DREStatsResponse{
			FunctionID: fn.ID,
			Summary: DREStatsSummary{
				DeterminismScore:        0,
				ReplayIntegrityScore:    0,
				VerifiedExecutionsTotal: 0,
				TotalExecutions:         0,
				ReplayDriftIncidents:    0,
				DriftScore:              1.0,
				DeterminismTier:         "unknown",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, DREStatsResponse{
		FunctionID: fn.ID,
		Summary: DREStatsSummary{
			DeterminismScore:        passport.DeterminismScore,
			ReplayIntegrityScore:    passport.ReplayIntegrityScore,
			VerifiedExecutionsTotal: passport.VerifiedExecutionsTotal,
			TotalExecutions:         passport.TotalExecutions,
			ReplayDriftIncidents:    passport.ReplayDriftIncidents,
			DriftScore:              passport.DriftScore,
			DeterminismTier:         "full",
		},
	})
}

// HandleGetPassportByFunctionID returns the Execution Passport for a function by function ID.
// This is an internal endpoint for platform services.
//
// GET /internal/functions/{function_id}/passport
func (h *Handler) HandleGetPassportByFunctionID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid function_id")
		return
	}

	passport, err := h.repo.GetPassportByFunctionID(functionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get passport: %v", err))
		return
	}

	if passport == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"function_id": functionIDStr,
			"passport": map[string]interface{}{
				"deterministic_reliability":   0,
				"replay_drift_incidents":      0,
				"verified_executions_total":   0,
				"total_executions":            0,
				"determinism_score":           0,
				"replay_integrity_score":      0,
				"performance_stability_score": 0,
				"drift_score":                 1.0,
				"capsule_version":             "dcc/1.0",
				"determinism_tier":            "full",
				"last_verified_at":            nil,
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function_id": functionIDStr,
		"passport": map[string]interface{}{
			"deterministic_reliability":   passport.DeterministicReliability,
			"replay_drift_incidents":      passport.ReplayDriftIncidents,
			"verified_executions_total":   passport.VerifiedExecutionsTotal,
			"total_executions":            passport.TotalExecutions,
			"determinism_score":           passport.DeterminismScore,
			"replay_integrity_score":      passport.ReplayIntegrityScore,
			"performance_stability_score": passport.PerformanceStabilityScore,
			"drift_score":                 passport.DriftScore,
			"capsule_version":             "dcc/1.0",
			"determinism_tier":            "full",
			"last_verified_at":            passport.LastVerifiedAt,
		},
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":  message,
		"status": status,
	})
}

// timingSafeHashEqual compares two hash strings in a timing-safe manner to prevent
// timing attacks on sensitive comparisons. Returns true if hashes are equal.
func timingSafeHashEqual(a, b string) bool {
	if len(a) != len(b) {
		// Still perform comparison to maintain constant time
		_ = subtle.ConstantTimeCompare([]byte(a), []byte(b))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
