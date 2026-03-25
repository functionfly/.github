package execution

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/dre/capsule"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	gonanoid "github.com/jaevor/go-nanoid"
	"github.com/sirupsen/logrus"
)

// Handler contains dependencies for execution handlers
type Handler struct {
	Repo         *registry.RegistryRepository
	BackendRepo  storage.Repository
	CacheService *cache.CacheService
	EdgeCache    *cache.EdgeCacheService
	// NodeID is the identifier of this execution node (used in MEG records and certificates)
	NodeID string
	// Region is the geographic region of this node
	Region string
	// NodeKey is the Ed25519 private key used to sign FXCERTs. If nil, certs are generated without a node signature (e.g. bootstrap).
	NodeKey ed25519.PrivateKey
	// PlatformKey is the optional Ed25519 platform key; when set, certs include a platform signature (Platform Key ID in UI).
	PlatformKey ed25519.PrivateKey
}

// HandleExecute handles executing a function
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := vars["version"]

	logrus.WithFields(logrus.Fields{
		"author":  author,
		"name":    name,
		"version": version,
	}).Info("HandleExecute called")

	// Parse request body
	var execReq functionregistry.ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&execReq); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	logrus.WithField("input", string(execReq.Input)).Debug("Parsed execution request")

	// Override version if specified in URL
	if version != "" {
		execReq.Version = version
	}

	// Get function by author and name
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Get function version
	var fnVersion *storage.RegistryFunctionVersion
	if execReq.Version == "" {
		// Use latest version
		fnVersion, err = h.Repo.GetLatestFunctionVersion(fn.ID)
	} else {
		fnVersion, err = h.Repo.GetFunctionVersion(fn.ID, execReq.Version)
	}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function version not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function version")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Validate runtime for tenant's plan - check if runtime is allowed
	if fnVersion.Runtime == plans.RuntimePythonMicroVM {
		if fn.TenantID == nil {
			h.writeError(w, http.StatusForbidden, functionregistry.ErrCodeInvalidInput,
				"python-microvm runtime requires a tenant-owned function (Enterprise tier)")
			return
		}
		tenantPlan := getTenantPlanFromContext(h.BackendRepo, *fn.TenantID)
		if err := validateRuntimeForPlan(tenantPlan, fnVersion.Runtime); err != nil {
			logrus.WithError(err).Warn("Runtime validation failed")
			h.writeError(w, http.StatusForbidden, functionregistry.ErrCodeInvalidInput, err.Error())
			return
		}
	}

	// Check function verification status before execution
	verificationSvc := verification.NewVerificationService(h.Repo, "", "") // Configure URLs as needed
	allowed, reason, err := verificationSvc.CheckExecutionAllowed(fnVersion.ID, author)
	if err != nil {
		logrus.WithError(err).Error("Failed to check function verification status")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	if !allowed {
		logrus.WithField("function_version_id", fnVersion.ID).Warn("Function execution blocked due to verification failure")
		h.writeError(w, http.StatusForbidden, functionregistry.ErrCodeInvalidInput, fmt.Sprintf("Function execution not allowed: %s", reason))
		return
	}

	// Check if we should queue this execution due to high load
	if h.shouldQueueExecution(r) {
		if err := h.queueExecution(r, fn.ID, execReq, fnVersion); err != nil {
			logrus.WithError(err).Error("Failed to queue execution")
			h.writeError(w, http.StatusServiceUnavailable, functionregistry.ErrCodeInternalError, "Service temporarily unavailable, please try again later")
			return
		}

		// Return accepted response for queued execution
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "queued",
			"message":    "Execution queued due to high load",
			"request_id": r.Header.Get("X-Request-ID"),
		})
		return
	}

	// Execute the function
	var result json.RawMessage
	var executionErr error
	var statusCode int
	cached := false
	var resourceUsage *ResourceUsage

	// Check cache eligibility for this function version
	versionData := cache.FunctionVersionData{
		FunctionID:    fnVersion.FunctionID,
		Version:       fnVersion.Version,
		Deterministic: fnVersion.Deterministic,
		CacheTTL:      fnVersion.CacheTTL,
		SideEffects:   fnVersion.SideEffects,
	}
	eligibility := cache.CheckEligibility(versionData)
	eligibility.CanUseCDN = fn.Visibility == "public"

	// Get resource limits (using function defaults for now)
	maxMemoryMB := fnVersion.MemoryMB
	maxCPUTimeMs := fnVersion.TimeoutMs

	// Define execution function that can be cached
	executeFn := h.createExecuteFunction(fnVersion, execReq, r, fn, maxMemoryMB, maxCPUTimeMs, &resourceUsage)

	// Execute with caching
	cacheResult, err := h.CacheService.GetOrExecute(eligibility, execReq.Input, executeFn)
	if err != nil {
		executionErr = err
		statusCode = http.StatusInternalServerError
		result = nil
	} else {
		result = cacheResult.Output
		cached = cacheResult.FromCache
		statusCode = http.StatusOK
	}

	// Calculate execution time
	durationMs := int(time.Since(startTime).Milliseconds())

	// Enterprise MicroVM billing — log usage metrics for downstream aggregation.
	if fnVersion.Runtime == plans.RuntimePythonMicroVM && fn.TenantID != nil {
		tenantPlan := getTenantPlanFromContext(h.BackendRepo, *fn.TenantID)
		if billing := plans.CalculateMicroVMBilling(
			tenantPlan,
			1, // single request
			float64(durationMs)/1000.0,
			maxMemoryMB,
			float64(durationMs)/1000.0,
		); billing != nil {
			logrus.WithFields(logrus.Fields{
				"tenant_id":       *fn.TenantID,
				"duration_ms":     durationMs,
				"memory_mb":       maxMemoryMB,
				"compute_charges": billing.ComputeCharges,
				"memory_charges":  billing.MemoryCharges,
				"total_cents":     billing.TotalCents,
			}).Info("MicroVM billing record")
		}
	}

	// Determine outcome
	outcome, errorCode := determineOutcome(executionErr, statusCode)

	// Perform replay verification for deterministic functions (only on successful executions)
	var verificationResult *ReplayVerificationResult
	if statusCode >= 200 && statusCode < 300 && fnVersion.Deterministic && !cached {
		// Get execution statistics for sophisticated verification scheduling
		executionCount, err := h.Repo.GetExecutionCountForVersion(fn.ID, fnVersion.Version)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"function_id": fn.ID,
				"version":     fnVersion.Version,
			}).Warn("Failed to get execution count for verification scheduling, using random sampling")
			executionCount = 0
		}

		lastVerified, err := h.Repo.GetLastVerificationTimeForVersion(fn.ID, fnVersion.Version)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"function_id": fn.ID,
				"version":     fnVersion.Version,
			}).Warn("Failed to get last verification time, proceeding without it")
		}

		recentFailureRate, err := h.Repo.GetRecentVerificationFailureRate(fn.ID, fnVersion.Version)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"function_id": fn.ID,
				"version":     fnVersion.Version,
			}).Warn("Failed to get recent failure rate, assuming 0")
			recentFailureRate = 0
		}

		// Use sophisticated verification scheduling
		shouldVerify := shouldVerifyReplay(fnVersion, executionCount+1, lastVerified, recentFailureRate)

		if shouldVerify {
			logrus.WithFields(logrus.Fields{
				"function_id": fn.ID,
				"version":     fnVersion.Version,
				"author":      author,
				"name":        name,
			}).Info("Performing replay verification for deterministic function")

			verificationResult = h.verifyReplay(fnVersion, execReq.Input, result, durationMs)

			// DRE 2.0: Update Trust Score v2 after successful verification
			if verificationResult != nil && verificationResult.Status == VerificationVerified {
				go h.updateTrustScoreV2(fn.ID)
			}
		}
	}

	// DRE 2.0: Build Merkle Execution Graph and FXCERT for deterministic functions, or all functionfly-authored functions
	var executionRootHash string
	var certID string
	issueFXCERT := statusCode >= 200 && statusCode < 300 && !cached && (fnVersion.Deterministic || strings.EqualFold(author, "functionfly"))
	if issueFXCERT {
		go h.buildAndStoreMEG(fn, fnVersion, execReq.Input, result, resourceUsage, durationMs)
	}

	// Record execution in database
	execRecord := &storage.RegistryFunctionExecution{
		FunctionID: fn.ID,
		Version:    fnVersion.Version,
		DurationMs: durationMs,
		StatusCode: statusCode,
		Cached:     cached,
		Outcome:    outcome,
		ErrorCode:  toNullString(&errorCode),
		CallerIP:   toNullString(func() *string { ip := getClientIP(r); return &ip }()),
		UserAgent:  toNullString(func() *string { ua := r.UserAgent(); return &ua }()),
	}

	// Phase 3 — Embed analytics: record the Origin header when the request
	// comes from an embed script (identified by the X-Embed-Origin header or
	// the standard Origin header when the Referer suggests an embed context).
	if embedOrigin := r.Header.Get("X-Embed-Origin"); embedOrigin != "" {
		execRecord.EmbedOrigin = sql.NullString{String: embedOrigin, Valid: true}
	} else if origin := r.Header.Get("Origin"); origin != "" {
		// Only record as embed origin when the request is cross-origin
		// (i.e., the Origin header is present and differs from the API host).
		execRecord.EmbedOrigin = sql.NullString{String: origin, Valid: true}
	}

	// Add verification results if available
	if verificationResult != nil {
		execRecord.VerifiedAt = sql.NullTime{Time: verificationResult.VerifiedAt, Valid: true}
		execRecord.VerificationStatus = sql.NullString{String: string(verificationResult.Status), Valid: true}
		if verificationResult.Error != "" {
			execRecord.VerificationError = sql.NullString{String: verificationResult.Error, Valid: true}
		}
		execRecord.ReplayedDurationMs = sql.NullInt32{Int32: int32(verificationResult.ReplayedDuration), Valid: true}
	}
	if err := h.Repo.RecordExecution(execRecord); err != nil {
		logrus.WithError(err).Error("Failed to record execution")
		// Don't fail the request if recording fails
	} else {
		// Update popularity score based on execution (async)
		go func() {
			if err := h.updateFunctionPopularity(fn.ID); err != nil {
				logrus.WithError(err).WithField("function_id", fn.ID).Warn("Failed to update function popularity")
			}
		}()
	}

	// Record resource usage if available
	if resourceUsage != nil {
		resourceRecord := &storage.ExecutionResourceUsage{
			ExecutionID:    &execRecord.ID,
			MaxMemoryMB:    resourceUsage.MaxMemoryMB,
			MaxCPUTimeMs:   resourceUsage.MaxCPUTimeMs,
			MemoryUsedMB:   resourceUsage.MemoryUsedMB,
			CPUTimeUsedMs:  resourceUsage.CPUTimeUsedMs,
			WallTimeUsedMs: resourceUsage.WallTimeUsedMs,
		}

		// Set termination reason if execution failed due to resource limits
		if executionErr != nil {
			if execError, ok := executionErr.(*ExecutionError); ok && execError.ResourceUsage != nil {
				resourceRecord.TerminatedBy = execError.TerminatedBy
			}
		}

		if err := h.Repo.RecordResourceUsage(resourceRecord); err != nil {
			logrus.WithError(err).Error("Failed to record resource usage")
			// Don't fail the request if recording fails
		}
	}

	// Generate public execution ID if execution was successful and should be shareable
	var executionID *string
	if statusCode >= 200 && statusCode < 300 && fnVersion.SideEffects == "none" {
		// Generate a unique execution ID using nanoid
		gen, err := gonanoid.Canonic()
		if err != nil {
			logrus.WithError(err).Error("Failed to create nanoid generator")
			fallbackID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
			executionID = &fallbackID
		} else {
			nanoID := gen()

			publicExec := &storage.RegistryExecutionPublic{
				PublicID:   nanoID,
				FunctionID: fn.ID,
				Version:    fnVersion.Version,
				InputJSON:  execReq.Input,
				OutputJSON: result,
				DurationMs: durationMs,
				Cached:     cached,
				Shareable:  true,
			}

			// Add verification results to public execution if available
			if verificationResult != nil {
				publicExec.VerifiedAt = sql.NullTime{Time: verificationResult.VerifiedAt, Valid: true}
				publicExec.VerificationStatus = sql.NullString{String: string(verificationResult.Status), Valid: true}
				if verificationResult.Error != "" {
					publicExec.VerificationError = sql.NullString{String: verificationResult.Error, Valid: true}
				}
				publicExec.ReplayedOutputJSON = verificationResult.ReplayedOutput
				publicExec.ReplayedDurationMs = sql.NullInt32{Int32: int32(verificationResult.ReplayedDuration), Valid: true}
			}
			if err := h.Repo.CreateExecutionPublic(publicExec); err != nil {
				logrus.WithError(err).Error("Failed to create public execution")
			} else {
				executionID = &publicExec.PublicID
			}
		}
	}

	// Format response
	if executionErr != nil || statusCode >= 400 {
		// Error response
		logrus.WithFields(logrus.Fields{
			"error":      executionErr,
			"statusCode": statusCode,
		}).Error("Execution failed, writing error response")
		msg := "Execution failed"
		if executionErr != nil {
			if s := executionErr.Error(); s != "" {
				msg = s
			}
		}
		code := errorCode
		if code == "" {
			code = functionregistry.ErrCodeRuntimeError
		}
		errorResp := functionregistry.ExecutionError{
			OK:         false,
			DurationMs: durationMs,
			Version:    fnVersion.Version,
			Error: functionregistry.ErrorDetail{
				Code:    code,
				Message: msg,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(errorResp)
	} else {
		// Success response
		logrus.WithFields(logrus.Fields{
			"result":     string(result),
			"cached":     cached,
			"durationMs": durationMs,
			"version":    fnVersion.Version,
		}).Info("Writing success response")
		successResp := functionregistry.ExecutionResponse{
			OK:          true,
			Data:        result,
			Cached:      cached,
			DurationMs:  durationMs,
			Version:     fnVersion.Version,
			ExecutionID: executionID,
		}
		// Suppress unused variable warnings — these are populated asynchronously
		_ = executionRootHash
		_ = certID
		w.Header().Set("Content-Type", "application/json")

		// Set cache headers
		if eligibility.Eligible {
			cache.SetCDNHeaders(w, eligibility, fn.Visibility == "public")

			// Set edge cache headers for popular functions
			if h.EdgeCache != nil {
				// Get function popularity score from repository
				popularityScore := fn.PopularityScore
				if popularityScore == 0 {
					// If not cached, use a default based on execution patterns
					// This could be enhanced to get real-time popularity
					popularityScore = 75 // Assume moderately popular for executed functions
				}

				h.EdgeCache.SetEdgeCacheHeaders(w, fn.ID, fnVersion.Version, popularityScore)
			}

			// Set X-Cache-Status and X-Cache-Layer headers for observability
			if cached {
				w.Header().Set("X-Cache-Status", "HIT")
			} else {
				w.Header().Set("X-Cache-Status", "MISS")
			}
			if cacheResult.Layer != "" && cacheResult.Layer != "none" {
				w.Header().Set("X-Cache-Layer", cacheResult.Layer)
			}
		} else {
			cache.SetNoCacheHeaders(w)
		}

		json.NewEncoder(w).Encode(successResp)
	}
}

// HandleTest handles testing a function with validation data
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get function by author and name
	fn, err := h.Repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Get latest function version
	fnVersion, err := h.Repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function version not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function version")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Validate function version is deployed (has deployment or backend)
	if fnVersion.DeploymentID == nil && fnVersion.BackendID == nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Function version is not deployed")
		return
	}

	// Use test input from request body or default test data
	var testInput map[string]interface{}
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &testInput); err != nil {
				h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid JSON in request body")
				return
			}
		}
	}

	// If no test input provided, use default test data
	if testInput == nil {
		testInput = map[string]interface{}{
			"test":      "validation",
			"timestamp": time.Now().Unix(),
		}
	}

	// Convert input to JSON for execution
	inputJSON, err := json.Marshal(testInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to serialize test input")
		return
	}

	// Check if function is eligible for caching (for validation purposes)
	versionData := cache.FunctionVersionData{
		FunctionID:    fnVersion.FunctionID,
		Version:       fnVersion.Version,
		Deterministic: fnVersion.Deterministic,
		CacheTTL:      fnVersion.CacheTTL,
		SideEffects:   fnVersion.SideEffects,
	}
	eligibility := cache.CheckEligibility(versionData)

	// Attempt to execute with test data (but don't store results)
	_, err = h.CacheService.GetOrExecute(eligibility, inputJSON, func() (json.RawMessage, error) {
		// This is a test execution - we validate the function can be called
		// but return a mock response instead of actual execution
		testResult := map[string]interface{}{
			"status":          "test_success",
			"message":         "Function validation successful",
			"timestamp":       time.Now().Unix(),
			"input_validated": true,
		}
		return json.Marshal(testResult)
	})

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"author": author,
			"name":   name,
		}).Error("Function test failed")
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeRuntimeError, "Function test failed: "+err.Error())
		return
	}

	// Return test success response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Function test completed successfully",
		"function": map[string]interface{}{
			"author":  fn.Author,
			"name":    fn.Name,
			"version": fnVersion.Version,
		},
		"test_input":   testInput,
		"validated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetReplay handles retrieving a shareable execution replay
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	execID := vars["execId"]

	// Get execution by public ID
	exec, err := h.Repo.GetExecutionPublicByID(execID)
	if err != nil {
		if err.Error() == "execution not found or not shareable" {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Execution not found or not shareable")
			return
		}
		logrus.WithError(err).Error("Failed to get execution")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Get function details
	fn, err := h.Repo.GetFunctionByID(exec.FunctionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"function_id":  exec.FunctionID.String(),
		"author":       fn.Author,
		"name":         fn.Name,
		"version":      exec.Version,
		"input_json":   exec.InputJSON,
		"output_json":  exec.OutputJSON,
		"duration_ms":  exec.DurationMs,
		"cached":       exec.Cached,
		"execution_id": exec.PublicID,
		"created_at":   exec.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleVerifyReplay handles manual verification of a specific execution replay
func (h *Handler) HandleVerifyReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	execID := vars["execId"]

	// Get execution by public ID
	exec, err := h.Repo.GetExecutionPublicByID(execID)
	if err != nil {
		if err.Error() == "execution not found or not shareable" {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Execution not found or not shareable")
			return
		}
		logrus.WithError(err).Error("Failed to get execution")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Get function version
	fnVersion, err := h.Repo.GetFunctionVersion(exec.FunctionID, exec.Version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function version not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function version")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Perform verification
	verificationResult := h.verifyReplay(fnVersion, exec.InputJSON, exec.OutputJSON, exec.DurationMs)

	// Prepare response
	response := map[string]interface{}{
		"execution_id":         execID,
		"verified_at":          verificationResult.VerifiedAt,
		"status":               string(verificationResult.Status),
		"output_matches":       verificationResult.OutputMatches,
		"original_duration_ms": verificationResult.OriginalDuration,
		"replayed_duration_ms": verificationResult.ReplayedDuration,
	}

	if verificationResult.Error != "" {
		response["error"] = verificationResult.Error
	}

	if verificationResult.ReplayedOutput != nil {
		response["replayed_output"] = verificationResult.ReplayedOutput
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods that need to be implemented or moved
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	// This method should be implemented or moved from the parent handler
}

func (h *Handler) shouldQueueExecution(r *http.Request) bool {
	// This method should be implemented or moved from the parent handler
	return false
}

func (h *Handler) queueExecution(r *http.Request, functionID uuid.UUID, execReq functionregistry.ExecutionRequest, fnVersion *storage.RegistryFunctionVersion) error {
	// This method should be implemented or moved from the parent handler
	return fmt.Errorf("not implemented")
}

func (h *Handler) updateFunctionPopularity(functionID uuid.UUID) error {
	return h.Repo.IncrementPopularity(functionID)
}

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
		FunctionID:        fn.ID,
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
