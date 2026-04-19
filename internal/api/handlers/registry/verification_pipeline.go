package registry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
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

	// Verify function exists
	fn, err := h.repo.GetFunctionByID(functionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", "Function not found")
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

	// Get the latest version for this function
	var versionID uuid.UUID
	if fn.LatestVersion.Valid && fn.LatestVersion.String != "" {
		latest, err := h.repo.GetFunctionVersion(functionID, fn.LatestVersion.String)
		if err == nil && latest != nil {
			versionID = latest.ID
		}
	}
	if versionID == uuid.Nil {
		// Fall back to latest version by published_at
		latest, err := h.repo.GetLatestFunctionVersion(functionID)
		if err == nil && latest != nil {
			versionID = latest.ID
		}
	}
	if versionID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "NO_VERSION", "Function has no published versions")
		return
	}

	// Persist the job record before running (pipeline runs async in production; here we run synchronously)
	job := &registry.VerificationJob{
		ID:                 uuid.New(),
		FunctionID:         functionID,
		FunctionVersionID:  versionID,
		Level:             level.String(),
		Status:            "pending",
		Priority:          "normal",
		RequestedAt:        time.Now(),
		ResultStatus:      "pending",
		IsAutoVerify:      false,
	}
	if err := h.repo.CreateVerificationJob(job); err != nil {
		h.writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create verification job")
		return
	}

	// Create pipeline and run verification
	pipeline := verification.NewPipeline(verification.PipelineConfig{
		EnableManualReview: level == verification.Level3Full,
	})

	result, err := pipeline.Run(r.Context(), functionID, versionID, level)
	if err != nil {
		// Update job as failed
		_ = h.repo.UpdateVerificationJobStatus(job.ID, "failed", "failed", nil, err.Error())
		h.writeError(w, http.StatusInternalServerError, "VERIFICATION_FAILED", "Verification failed: "+err.Error())
		return
	}

	// Persist pipeline result back to the job record
	jobStatus := "completed"
	if result.Status == "fail" {
		jobStatus = "failed"
	}
	stagesJSON, _ := json.Marshal(result.Stages)
	_ = h.repo.UpdateVerificationJobStatus(job.ID, jobStatus, result.Status, stagesJSON, result.Error)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":      job.ID,
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

	// Get the latest version to look up its verification status
	latest, err := h.repo.GetLatestFunctionVersion(functionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", "Function not found")
		return
	}

	// Fetch verification status for the latest version
	status, err := h.repo.GetVerificationStatus(latest.ID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to get verification status")
		return
	}

	if status == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"function_id":          functionID,
			"status":              "unverified",
			"verification_level":  "unverified",
			"latest_version":      latest.Version,
			"content_hash_verified": false,
			"signature_verified":   false,
			"malware_scanned":      false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"function_id":           functionID,
		"version_id":           latest.ID.String(),
		"latest_version":       latest.Version,
		"status":               status.OverallStatus,
		"verification_level":   status.ApprovalStatus,
		"content_hash_verified": status.ContentHashVerified,
		"signature_verified":   status.SignatureVerified,
		"malware_scanned":      status.MalwareScanned,
		"malware_status":       status.MalwareStatus,
		"last_verified_at":    status.LastVerifiedAt,
		"next_verification_at": status.NextVerificationAt,
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

	job, err := h.repo.GetVerificationJob(jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "JOB_NOT_FOUND", "Verification job not found")
		return
	}

	var stages map[string]interface{}
	if len(job.ResultData) > 0 {
		_ = json.Unmarshal(job.ResultData, &stages)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":         job.ID,
		"function_id":   job.FunctionID,
		"version_id":    job.FunctionVersionID,
		"level":        job.Level,
		"status":       job.Status,
		"priority":     job.Priority,
		"result_status": job.ResultStatus,
		"result_data":  stages,
		"error":       job.Error,
		"requested_at": job.RequestedAt,
		"started_at":  job.StartedAt,
		"completed_at": job.CompletedAt,
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

	if err := h.repo.CancelVerificationJob(jobID); err != nil {
		h.writeError(w, http.StatusConflict, "CANCEL_FAILED", "Job not found or already finished")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "cancelled",
	})
}
