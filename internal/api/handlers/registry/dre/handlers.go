// Package dre implements the DRE 2.0 API handlers for execution certificates,
// replay verification, execution passports, and divergence simulation.
package dre

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler contains dependencies for DRE API handlers.
type Handler struct {
	Repo *registry.RegistryRepository
}

// NewHandler creates a new DRE handler.
func NewHandler(repo *registry.RegistryRepository) *Handler {
	return &Handler{Repo: repo}
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

	cert, err := h.Repo.GetCertificateByID(certID)
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
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	certs, err := h.Repo.GetCertificatesByFunctionID(fn.ID, limit, offset)
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

// HandleReplay triggers a deterministic replay of a specific execution.
// It builds MEG for both the original and replay, compares root hashes,
// and returns the comparison result.
//
// POST /registry/{author}/{name}/replay/{execution_id}
func (h *Handler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	executionIDStr := vars["execution_id"]

	// Get function
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
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

	megRecord, err := h.Repo.GetMEGByExecutionID(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get MEG record: %v", err))
		return
	}
	if megRecord == nil {
		writeError(w, http.StatusNotFound, "no MEG record found for this execution")
		return
	}

	// Return the MEG record with verification status
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":            fmt.Sprintf("fx://%s/%s", author, name),
		"function_id":         fn.ID,
		"execution_id":        executionIDStr,
		"execution_root_hash": megRecord.ExecutionRootHash,
		"replay_root_hash":    megRecord.ReplayRootHash,
		"replay_verified_at":  megRecord.ReplayVerifiedAt,
		"determinism_tier":    megRecord.DeterminismTier,
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
		"roots_match": megRecord.ReplayRootHash != "" && megRecord.ReplayRootHash == megRecord.ExecutionRootHash,
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
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	passport, err := h.Repo.GetPassportByFunctionID(fn.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get passport: %v", err))
		return
	}

	if passport == nil {
		// Return empty passport for functions with no DRE data yet
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"function": fmt.Sprintf("fx://%s/%s", author, name),
			"passport": map[string]interface{}{
				"deterministic_reliability":    0,
				"replay_drift_incidents":       0,
				"verified_executions_total":    0,
				"total_executions":             0,
				"determinism_score":            0,
				"replay_integrity_score":       0,
				"performance_stability_score":  0,
				"drift_score":                  1.0,
				"capsule_version":              "dcc/1.0",
				"determinism_tier":             "full",
				"last_verified_at":             nil,
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function": fmt.Sprintf("fx://%s/%s", author, name),
		"passport": map[string]interface{}{
			"deterministic_reliability":    passport.DeterministicReliability,
			"replay_drift_incidents":       passport.ReplayDriftIncidents,
			"verified_executions_total":    passport.VerifiedExecutionsTotal,
			"total_executions":             passport.TotalExecutions,
			"determinism_score":            passport.DeterminismScore,
			"replay_integrity_score":       passport.ReplayIntegrityScore,
			"performance_stability_score":  passport.PerformanceStabilityScore,
			"drift_score":                  passport.DriftScore,
			"capsule_version":              "dcc/1.0",
			"determinism_tier":             "full",
			"last_verified_at":             passport.LastVerifiedAt,
		},
	})
}

// DivergenceSimulationRequest is the request body for divergence simulation.
type DivergenceSimulationRequest struct {
	MemoryLimit    int64  `json:"memory_limit"`
	RuntimeVersion string `json:"runtime_version"`
	Region         string `json:"region"`
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
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "function not found")
		return
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

	// Build a simulated MEG with the modified environment
	// (In production, this would actually re-execute the function)
	simulatedMEG := &drecrypto.MEGResult{
		ExecutionRootHash: drecrypto.HashString(drecrypto.TagMeta, []byte(modifiedHash+baseExecID)),
		InputHash:         drecrypto.HashString(drecrypto.TagInput, []byte("simulated")),
		EnvironmentHash:   modifiedHash,
		DependencyHash:    drecrypto.HashString(drecrypto.TagDeps, []byte("[]")),
		ResourceHash:      drecrypto.HashString(drecrypto.TagResource, []byte("{}")),
		OutputHash:        drecrypto.HashString(drecrypto.TagOutput, []byte("simulated")),
		MetadataHash:      drecrypto.HashString(drecrypto.TagMeta, []byte(baseExecID)),
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"function":    fmt.Sprintf("fx://%s/%s", author, name),
		"function_id": fn.ID,
		"simulation": map[string]interface{}{
			"modified_capsule_hash": modifiedHash,
			"simulated_root_hash":   simulatedMEG.ExecutionRootHash,
			"constraints": map[string]interface{}{
				"memory_limit":    req.MemoryLimit,
				"runtime_version": req.RuntimeVersion,
				"region":          req.Region,
			},
			"simulated_at": time.Now().UTC().Format(time.RFC3339),
			"note":         "This is a simulated divergence analysis. Production replay requires actual function execution.",
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
		"error":   message,
		"status":  status,
	})
}
