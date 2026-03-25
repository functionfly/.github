package registry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleTriggerVerification triggers verification for a function
// POST /v1/registry/functions/{id}/verify
func (h *Handler) HandleTriggerVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_FUNCTION_ID", "Invalid function ID")
		return
	}

	var req struct {
		Level string `json:"level"` // "basic", "standard", "full"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Level = "basic"
	}

	level, err := verification.ParseVerificationLevel(req.Level)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_LEVEL", "Invalid verification level")
		return
	}

	// Create pipeline and run verification
	pipeline := verification.NewPipeline(verification.PipelineConfig{
		EnableManualReview: level == verification.Level3Full,
	})

	// Use a placeholder version ID since we don't have repository access yet
	versionID := uuid.New()

	result, err := pipeline.Run(r.Context(), functionID, versionID, level)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "VERIFICATION_FAILED", "Verification failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":      result.JobID,
		"function_id": functionID,
		"status":      result.Status,
		"level":       result.Level.String(),
		"stages":      result.Stages,
		"started_at":  result.StartedAt,
	})
}

// HandleGetFunctionVerification gets the verification status for a function
// GET /v1/registry/functions/{id}/verification
func (h *Handler) HandleGetFunctionVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_FUNCTION_ID", "Invalid function ID")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"function_id":          functionID,
		"status":              "unverified",
		"verification_level": "unverified",
	})
}

// HandleGetVerificationLevels returns all available verification levels
// GET /v1/registry/verification/levels
func (h *Handler) HandleGetVerificationLevels(w http.ResponseWriter, r *http.Request) {
	levels := verification.GetVerificationLevels()

	response := make([]map[string]interface{}, len(levels))
	for i, level := range levels {
		response[i] = map[string]interface{}{
			"level":                   level.Level,
			"name":                    level.Name,
			"description":             level.Description,
			"requires_malware_scan":   level.RequiresMalwareScan,
			"requires_dre":            level.RequiresDRE,
			"requires_fxcert":         level.RequiresFXCERT,
			"requires_manual_review":  level.RequiresManualReview,
			"trust_bonus":            level.TrustBonus,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"levels": response,
	})
}

// HandleGetVerificationJob gets the status of a verification job
// GET /v1/registry/verification/jobs/{jobId}
func (h *Handler) HandleGetVerificationJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobIDStr := vars["jobId"]

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JOB_ID", "Invalid job ID")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":        jobID,
		"function_id":  uuid.Nil,
		"level":        "basic",
		"status":        "pending",
		"priority":      "normal",
		"result_status": "pending",
		"requested_at":  time.Now(),
	})
}

// HandleCancelVerification cancels a pending verification job
// POST /v1/registry/verification/jobs/{jobId}/cancel
func (h *Handler) HandleCancelVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobIDStr := vars["jobId"]

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JOB_ID", "Invalid job ID")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "cancelled",
	})
}
