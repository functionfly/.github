package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/atlas"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/privacy"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	gonanoid "github.com/jaevor/go-nanoid"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

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

	// Limit request body to 10 MB to prevent memory exhaustion
	const maxBodyBytes = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	// Parse request body
	var execReq functionregistry.ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&execReq); err != nil {
		if err.Error() == "http: request body too large" {
			h.writeError(w, http.StatusRequestEntityTooLarge, functionregistry.ErrCodeInvalidInput, "Request body too large (max 10MB)")
			return
		}
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	logrus.WithField("input", string(execReq.Input)).Debug("Parsed execution request")

	// Override version if specified in URL
	if version != "" {
		execReq.Version = version
	}

	// Get function by author and name
	fn, err := h.Repo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, functionregistry.ErrCodeNotFound, "Function not found")
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	// Check if this is a paid function and process payment
	if fn.PricePerCall > 0 && h.BillingController != nil {
		// Require authentication for paid functions
		userID := getUserIDFromRequest(r)
		if userID == "" {
			h.writeError(w, http.StatusUnauthorized, functionregistry.ErrCodeUnauthorized,
				"Authentication required for paid function execution")
			return
		}

		// Check credit balance and charge
		controls, err := h.BillingController.GetOrCreateControls(r.Context(), userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to get billing controls")
			h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError,
				"Failed to process payment")
			return
		}

		if controls.CreditBalanceUSD < fn.PricePerCall {
			h.writeError(w, http.StatusPaymentRequired, functionregistry.ErrCodePaymentRequired,
				fmt.Sprintf("Insufficient balance: need $%.4f, have $%.4f", fn.PricePerCall, controls.CreditBalanceUSD))
			return
		}

		// Deduct payment
		update, err := h.BillingController.ConsumeCredits(r.Context(), userID, fn.PricePerCall)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to consume credits")
			h.writeError(w, http.StatusPaymentRequired, functionregistry.ErrCodePaymentRequired,
				fmt.Sprintf("Payment failed: %s", err.Error()))
			return
		}

		logrus.WithFields(logrus.Fields{
			"user_id":          userID,
			"function":         fmt.Sprintf("%s/%s", author, name),
			"amount_usd":       fn.PricePerCall,
			"previous_balance": update.PreviousUSD,
			"new_balance":      update.CurrentUSD,
		}).Info("Charged for paid function execution")
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

	// Validate runtime for tenant's plan
	if fnVersion.Runtime == plans.RuntimePythonMicroVM {
		if fn.TenantID == nil {
			h.writeError(w, http.StatusForbidden, functionregistry.ErrCodeInvalidInput,
				"python-microvm runtime requires a tenant-owned function (Enterprise tier)")
			return
		}
		tenantPlan := getTenantPlanFromContext(h.BackendRepo, *fn.TenantID)
		if err := validateRuntimeForPlan(tenantPlan, fnVersion.Runtime); err != nil {
			logrus.WithError(err).Warn("Runtime validation failed")
			h.writeError(w, http.StatusForbidden, functionregistry.ErrCodeInvalidInput, "Runtime not allowed for this plan")
			return
		}
	}

	// Check function verification status before execution
	verificationSvc := verification.NewVerificationService(h.Repo, "", "")
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

	// Real-time quota enforcement
	if fn.TenantID != nil && h.UsageTracker != nil && h.UsageTracker.IsEnabled() {
		quotaResult, err := h.UsageTracker.RecordExecution(r.Context(), *fn.TenantID, "")
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", *fn.TenantID).Warn("Quota check failed, allowing execution")
		} else if !quotaResult.Allowed {
			logrus.WithFields(logrus.Fields{
				"tenant_id": *fn.TenantID,
				"reason":    quotaResult.Reason,
				"status":    quotaResult.Status,
			}).Warn("Quota exceeded, blocking execution")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired) // 402
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "QUOTA_EXCEEDED",
					"message": quotaResult.Reason,
					"type":    "quota_exceeded",
				},
				"quota_status": quotaResult.Status,
				"upgrade_url":  "/settings/billing",
			})
			return
		}

		// Add quota headers to response
		if quotaResult.Status != nil {
			w.Header().Set("X-Quota-Executions-Percent", fmt.Sprintf("%.1f", quotaResult.Status.ExecutionsPercent))
			w.Header().Set("X-Quota-Compute-Percent", fmt.Sprintf("%.1f", quotaResult.Status.ComputeMsPercent))
			w.Header().Set("X-Quota-Status", quotaResult.Status.Status)
		}
	}

	// Embed origin rate limiting — enforce at execution time, not just script serve time.
	// A malicious actor could cache the embed script and call ff.run() without limit.
	if embedOrigin := r.Header.Get("X-Embed-Origin"); embedOrigin != "" {
		embedCfg, err := h.Repo.GetFunctionEmbedConfig(fn.ID)
		if err == nil && embedCfg != nil && embedCfg.Enabled && embedCfg.RateLimitPerHour > 0 {
			since := time.Now().Add(-time.Hour)
			count, rlErr := h.Repo.GetEmbedExecutionCountByOrigin(fn.ID, embedOrigin, since)
			if rlErr == nil && count >= int64(embedCfg.RateLimitPerHour) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "3600")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "EMBED_RATE_LIMITED",
						"message": "Embed rate limit exceeded for this origin",
						"type":    "rate_limit",
					},
					"retry_after": 3600,
					"origin":      embedOrigin,
				})
				return
			}
		}
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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

	// Atlas tracing — start a trace run for this execution
	var atlasRunID string
	if h.AtlasTracer != nil && h.AtlasTracer.Enabled() {
		atlasRunID, _ = h.AtlasTracer.StartExecutionTrace(r.Context(), &atlas.ExecutionTrace{
			FunctionID:   fn.ID.String(),
			FunctionName: fn.Name,
			Author:       author,
			Version:      fnVersion.Version,
			Runtime:      fnVersion.Runtime,
			Tier:         resolveTierFromRequest(fn, h.BackendRepo),
			StartTime:    startTime,
			InputPayload: execReq.Input,
		})
		if atlasRunID != "" {
			logrus.WithField("atlas_run_id", atlasRunID).Debug("atlas: started execution trace")
		}
	}

	// MicroVM execution tracking
	var microvmExecutionID uuid.UUID
	var microvmMemoryMB int
	var microvmVCPUs int
	if fnVersion.Runtime == plans.RuntimePythonMicroVM && fn.TenantID != nil && h.MicroVMRepo != nil {
		microvmMemoryMB = fnVersion.MemoryMB
		if microvmMemoryMB == 0 {
			microvmMemoryMB = plans.EnterpriseDefaultMemoryMB
		}
		microvmVCPUs = plans.EnterpriseDefaultVCPU
		if limits := plans.GetMicroVMLimits(plans.PlanEnterprise); limits != nil && microvmMemoryMB > limits.MaxMemoryMB {
			microvmMemoryMB = limits.MaxMemoryMB
		}

		microvmExec := &storage.MicroVMExecution{
			ID:              uuid.New(),
			TenantID:        *fn.TenantID,
			FunctionID:      fnVersion.FunctionID,
			FunctionVersion: fnVersion.Version,
			ExecutionID:    uuid.New(),
			StartedAt:       time.Now(),
			MemoryMB:        microvmMemoryMB,
			VCPUs:           microvmVCPUs,
			Status:          "running",
			CreatedAt:       time.Now(),
		}
		if err := h.MicroVMRepo.CreateExecution(r.Context(), microvmExec); err != nil {
			logrus.WithError(err).Error("Failed to create MicroVM execution record")
		} else {
			microvmExecutionID = microvmExec.ID
		}
	}

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

	// Get resource limits
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

	// Enterprise MicroVM billing - log usage metrics for downstream aggregation
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

	// Atlas tracing — finish the trace with result or error
	if h.AtlasTracer != nil && h.AtlasTracer.Enabled() && atlasRunID != "" {
		atlasResult := &atlas.ExecutionResult{
			Output:     result,
			DurationMs: durationMs,
			Cached:     cached,
			StatusCode: statusCode,
		}
		if executionErr != nil {
			atlasResult.Error = executionErr.Error()
		}
		if resourceUsage != nil {
			atlasResult.ResourceUsage = &atlas.ResourceUsageResult{
				MaxMemoryMB:    resourceUsage.MaxMemoryMB,
				MemoryUsedMB:   resourceUsage.MemoryUsedMB,
				CPUTimeUsedMs:  resourceUsage.CPUTimeUsedMs,
				WallTimeUsedMs: resourceUsage.WallTimeUsedMs,
			}
		}
		go h.AtlasTracer.FinishExecutionTrace(context.Background(), atlasRunID, atlasResult)
	}

	// Update MicroVM execution record if we created one
	if microvmExecutionID != uuid.Nil {
		execStatus := "completed"
		execOutcome := string(outcome)
		var execErrMsg *string
		if executionErr != nil {
			errStr := executionErr.Error()
			execErrMsg = &errStr
		}
		if statusCode >= 400 {
			execStatus = "failed"
		} else if statusCode == -1 || strings.Contains(execOutcome, "timeout") {
			execStatus = "timeout"
		}
		go func() {
			if err := h.MicroVMRepo.UpdateExecutionStatus(context.Background(), microvmExecutionID, execStatus, &execOutcome, execErrMsg, time.Now(), durationMs); err != nil {
				logrus.WithError(err).Error("Failed to update MicroVM execution record")
			}
		}()
	}

	// Perform replay verification for deterministic functions (only on successful executions)
	var verificationResult *ReplayVerificationResult
	if statusCode >= 200 && statusCode < 300 && fnVersion.Deterministic && !cached {
		verificationResult = h.performReplayVerification(fn, fnVersion, author, name, execReq.Input, result, durationMs)
	}

	// Build and store MEG for deterministic functions, or all functionfly-authored functions
	var executionRootHash string
	var certID string
	if shouldIssueFXCERT(author, fnVersion) && statusCode >= 200 && statusCode < 300 && !cached {
		go h.buildAndStoreMEG(fn, fnVersion, execReq.Input, result, resourceUsage, durationMs)
	}

	// Record execution in database
	execRecord := h.createExecutionRecord(fn, fnVersion, durationMs, statusCode, cached, outcome, errorCode, r, verificationResult)

	// Save execution record and trigger async updates
	if err := h.Repo.RecordExecution(r.Context(), execRecord); err != nil {
		logrus.WithError(err).Error("Failed to record execution")
	} else {
		// Async updates
		go func() { _ = h.updateFunctionPopularity(fn.ID) }()
		go func() { _ = h.recordBillingUsageEvent(fn, execRecord, resourceUsage) }()
		go h.syncRealtimeUsage(fn, resourceUsage)

		// Record execution for DNA analysis pipeline (fire-and-forget)
		if h.DNARecorder != nil {
			go h.DNARecorder.RecordExecutionFromPipeline(
				context.Background(),
				fn.ID.String(),
				"registry",
				durationMs,
				statusCode,
				false, // coldStart detection would require runtime metadata
				h.Region,
			)
		}
	}

	// Record resource usage if available
	if resourceUsage != nil {
		h.recordResourceUsage(execRecord.ID, resourceUsage, executionErr)
	}

	// Generate public execution ID if successful and shareable
	executionID := h.generateExecutionID(statusCode, fnVersion, fn, execReq.Input, result, durationMs, cached, verificationResult, r)

	// Fire receipt milestone hook if receipt was created (executionID != nil)
	if executionID != nil && h.ReceiptMilestoneHook != nil {
		go h.ReceiptMilestoneHook(r.Context(), fn.ID, fn.TenantID, *executionID)
	}

	// Format response
	if executionErr != nil || statusCode >= 400 {
		h.writeErrorResponse(w, executionErr, statusCode, errorCode, durationMs, fnVersion.Version)
	} else {
		h.writeSuccessResponse(w, result, cached, durationMs, fnVersion.Version, executionID, eligibility, cacheResult, fn)
	}

	// Suppress unused variable warnings
	_ = executionRootHash
	_ = certID
}

// performReplayVerification handles replay verification for deterministic functions
func (h *Handler) performReplayVerification(
	fn *storage.RegistryFunction,
	fnVersion *storage.RegistryFunctionVersion,
	author, name string,
	input json.RawMessage,
	result json.RawMessage,
	durationMs int,
) *ReplayVerificationResult {
	// Get execution statistics for verification scheduling
	executionCount, err := h.Repo.GetExecutionCountForVersion(context.Background(), fn.ID, fnVersion.Version)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fn.ID,
			"version":     fnVersion.Version,
		}).Warn("Failed to get execution count for verification scheduling")
		executionCount = 0
	}

	lastVerified, err := h.Repo.GetLastVerificationTimeForVersion(context.Background(), fn.ID, fnVersion.Version)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fn.ID,
			"version":     fnVersion.Version,
		}).Warn("Failed to get last verification time")
	}

	recentFailureRate, err := h.Repo.GetRecentVerificationFailureRate(context.Background(), fn.ID, fnVersion.Version)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fn.ID,
			"version":     fnVersion.Version,
		}).Warn("Failed to get recent failure rate")
		recentFailureRate = 0
	}

	// Use sophisticated verification scheduling
	shouldVerify := shouldVerifyReplay(fnVersion, executionCount+1, lastVerified, recentFailureRate)

	if !shouldVerify {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"function_id": fn.ID,
		"version":     fnVersion.Version,
		"author":      author,
		"name":        name,
	}).Info("Performing replay verification for deterministic function")

	verificationResult := h.verifyReplay(fnVersion, input, result, durationMs)

	// Update Trust Score v2 after successful verification
	if verificationResult != nil && verificationResult.Status == VerificationVerified {
		go h.updateTrustScoreV2(fn.ID)
	}

	return verificationResult
}

// createExecutionRecord creates the execution record for database storage
// with privacy-aware logging
func (h *Handler) createExecutionRecord(
	fn *storage.RegistryFunction,
	fnVersion *storage.RegistryFunctionVersion,
	durationMs, statusCode int,
	cached bool,
	outcome ExecutionOutcome,
	errorCode string,
	r *http.Request,
	verificationResult *ReplayVerificationResult,
) *storage.RegistryFunctionExecution {
	// Extract raw values
	rawIP := getClientIP(r)
	rawUA := r.UserAgent()
	var embedOrigin string
	if embedOrigin = r.Header.Get("X-Embed-Origin"); embedOrigin == "" {
		embedOrigin = r.Header.Get("Origin")
	}

	// Check for privacy context from middleware
	var privacySettings *privacy.PrivacySettings
	if ctxSettings := privacy.GetPrivacySettingsFromContext(r.Context()); ctxSettings != nil {
		privacySettings = ctxSettings
	}

	// Apply privacy controls if privacy service is available
	var callerIP, userAgent string
	if h.PrivacyService != nil {
		// Anonymize data based on privacy settings
		callerIP, userAgent, embedOrigin = h.PrivacyService.AnonymizeExecutionData(rawIP, rawUA, embedOrigin, privacySettings)
	} else {
		// Default: use raw values (backward compatible)
		callerIP = rawIP
		userAgent = rawUA
	}

	// Check if we should log geo data
	logGeo := true
	if h.PrivacyService != nil && privacySettings != nil {
		logGeo = h.PrivacyService.ShouldLogGeoData(privacySettings)
	}

	// Check if we should log embed origin
	logEmbed := true
	if h.PrivacyService != nil && privacySettings != nil {
		logEmbed = h.PrivacyService.ShouldLogEmbedOrigin(privacySettings)
	}

	execRecord := &storage.RegistryFunctionExecution{
		FunctionID: fn.ID,
		Version:    fnVersion.Version,
		DurationMs: durationMs,
		StatusCode: statusCode,
		Cached:     cached,
		Outcome:    string(outcome),
		ErrorCode:  toNullString(&errorCode),
		CallerIP:   toNullString(&callerIP),
		UserAgent:  toNullString(&userAgent),
	}

	// Add geo country if logging is enabled
	if logGeo && callerIP != "" && callerIP != "[REDACTED]" {
		// Get privacy-preserving region instead of exact geo
		region := privacy.GetRegionFromIP(callerIP)
		execRecord.GeoCountry = sql.NullString{String: region, Valid: true}
	}

	// Record embed analytics origin if logging is enabled
	if logEmbed && embedOrigin != "" {
		execRecord.EmbedOrigin = sql.NullString{String: embedOrigin, Valid: true}
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

	return execRecord
}

// recordResourceUsage records resource usage for an execution
func (h *Handler) recordResourceUsage(executionID uuid.UUID, resourceUsage *ResourceUsage, executionErr error) {
	if resourceUsage == nil {
		return
	}

	resourceRecord := &storage.ExecutionResourceUsage{
		ExecutionID:    &executionID,
		MaxMemoryMB:    resourceUsage.MaxMemoryMB,
		MaxCPUTimeMs:   resourceUsage.MaxCPUTimeMs,
		MemoryUsedMB:   float64(resourceUsage.MemoryUsedMB),
		CPUTimeUsedMs:  resourceUsage.CPUTimeUsedMs,
		WallTimeUsedMs: resourceUsage.WallTimeUsedMs,
	}

	// Set termination reason if execution failed due to resource limits
	if executionErr != nil {
		if execError, ok := executionErr.(*ExecutionError); ok && execError.ResourceUsage != nil {
			resourceRecord.TerminatedBy = execError.TerminatedBy
		}
	}

	if err := h.Repo.RecordResourceUsage(context.Background(), resourceRecord); err != nil {
		logrus.WithError(err).Error("Failed to record resource usage")
	}

	// Update resourceUsage region from handler for cost tracking
	if resourceUsage.Region == "" && h.Region != "" {
		resourceUsage.Region = h.Region
	}
}

// syncRealtimeUsage syncs usage to the realtime tracker for quota enforcement
func (h *Handler) syncRealtimeUsage(fn *storage.RegistryFunction, resourceUsage *ResourceUsage) {
	if fn.TenantID != nil && h.UsageTracker != nil && h.UsageTracker.IsEnabled() {
		if resourceUsage != nil && resourceUsage.CPUTimeUsedMs > 0 {
			if err := h.UsageTracker.RecordComputeUsage(context.Background(), *fn.TenantID, resourceUsage.CPUTimeUsedMs); err != nil {
				logrus.WithError(err).WithField("tenant_id", *fn.TenantID).Warn("Failed to record realtime compute usage")
			}
		}
	}
}

// generateExecutionID creates a public execution ID for successful, shareable executions
// with privacy-aware input/output handling
func (h *Handler) generateExecutionID(
	statusCode int,
	fnVersion *storage.RegistryFunctionVersion,
	fn *storage.RegistryFunction,
	input json.RawMessage,
	result json.RawMessage,
	durationMs int,
	cached bool,
	verificationResult *ReplayVerificationResult,
	r *http.Request,
) *string {
	if statusCode < 200 || statusCode >= 300 || fnVersion.SideEffects != "none" {
		return nil
	}

	// Check privacy settings for input/output storage
	var storeInputOutput = true
	var sanitizedInput, sanitizedOutput = input, result

	if h.PrivacyService != nil {
		var privacySettings *privacy.PrivacySettings
		if ctxSettings := privacy.GetPrivacySettingsFromContext(r.Context()); ctxSettings != nil {
			privacySettings = ctxSettings
		}
		storeInputOutput = h.PrivacyService.ShouldStoreInputOutput(privacySettings)

		// If storing, sanitize for PII first
		if storeInputOutput {
			sanitizedInput, sanitizedOutput, _ = h.PrivacyService.SanitizeInputOutput(input, result)
		}
	}

	// If input/output storage is disabled, don't create public execution
	if !storeInputOutput {
		logrus.Debug("Input/output storage disabled for privacy, skipping public execution creation")
		return nil
	}

	gen, err := gonanoid.Canonic()
	if err != nil {
		logrus.WithError(err).Error("Failed to create nanoid generator")
		fallbackID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
		return &fallbackID
	}

	nanoID := gen()

	// Ensure input/output are never nil (which would become NULL in DB)
	inputJSON := sanitizedInput
	if len(inputJSON) == 0 {
		inputJSON = []byte("null")
	}
	outputJSON := sanitizedOutput
	if len(outputJSON) == 0 {
		outputJSON = []byte("null")
	}

	publicExec := &storage.RegistryExecutionPublic{
		PublicID:   nanoID,
		FunctionID: fn.ID,
		Version:    fnVersion.Version,
		InputJSON:  inputJSON,
		OutputJSON: outputJSON,
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
		if verificationResult.ReplayedDuration > 0 {
			replayMs := int32(verificationResult.ReplayedDuration)
			publicExec.ReplayedDurationMs = sql.NullInt32{Int32: replayMs, Valid: true}
		}
	}

	if err := h.Repo.CreateExecutionPublic(r.Context(), publicExec); err != nil {
		logrus.WithError(err).Error("Failed to create public execution")
		return nil
	}

	return &publicExec.PublicID
}

// writeErrorResponse writes an error response for failed executions
func (h *Handler) writeErrorResponse(w http.ResponseWriter, executionErr error, statusCode int, errorCode string, durationMs int, version string) {
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
		Version:    version,
		Error: functionregistry.ErrorDetail{
			Code:    code,
			Message: msg,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResp)
}

// writeSuccessResponse writes a success response for successful executions
func (h *Handler) writeSuccessResponse(
	w http.ResponseWriter,
	result json.RawMessage,
	cached bool,
	durationMs int,
	version string,
	executionID *string,
	eligibility cache.EligibilityResult,
	cacheResult *cache.CacheResult,
	fn *storage.RegistryFunction,
) {
	logrus.WithFields(logrus.Fields{
		"result":     string(result),
		"cached":     cached,
		"durationMs": durationMs,
		"version":    version,
	}).Info("Writing success response")

	// Ensure Data always serializes to valid JSON (null when empty, not an omission).
	data := result
	if len(data) == 0 {
		data = json.RawMessage("null")
	}

	successResp := functionregistry.ExecutionResponse{
		OK:          true,
		Data:        data,
		Cached:      cached,
		DurationMs:  durationMs,
		Version:     version,
		ExecutionID: executionID,
	}

	w.Header().Set("Content-Type", "application/json")

	// Set cache headers
	if eligibility.Eligible {
		cache.SetCDNHeaders(w, eligibility, fn.Visibility == "public")

		// Set edge cache headers for popular functions
		if h.EdgeCache != nil {
			popularityScore := fn.PopularityScore
			if popularityScore == 0 {
				popularityScore = 75
			}
			h.EdgeCache.SetEdgeCacheHeaders(w, fn.ID, version, popularityScore)
		}

		// Set X-Cache-Status and X-Cache-Layer headers
		if cached {
			w.Header().Set("X-Cache-Status", "HIT")
		} else {
			w.Header().Set("X-Cache-Status", "MISS")
		}
		if cacheResult != nil && cacheResult.Layer != "" && cacheResult.Layer != "none" {
			w.Header().Set("X-Cache-Layer", cacheResult.Layer)
		}
	} else {
		cache.SetNoCacheHeaders(w)
	}

	_ = json.NewEncoder(w).Encode(successResp)
}

// HandleTest handles testing a function with validation data
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get function by author and name
	fn, err := h.Repo.GetFunctionByAuthorName(r.Context(), author, name)
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

	// Validate function version has code to execute (deployed or source code for lazy bundling)
	if fnVersion.DeploymentID == nil && fnVersion.BackendID == nil && (!fnVersion.SourceCode.Valid || fnVersion.SourceCode.String == "") {
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

	// Check if function is eligible for caching
	versionData := cache.FunctionVersionData{
		FunctionID:    fnVersion.FunctionID,
		Version:       fnVersion.Version,
		Deterministic: fnVersion.Deterministic,
		CacheTTL:      fnVersion.CacheTTL,
		SideEffects:   fnVersion.SideEffects,
	}
	eligibility := cache.CheckEligibility(versionData)

	// Attempt to execute with test data
	_, err = h.CacheService.GetOrExecute(eligibility, inputJSON, func() (json.RawMessage, error) {
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
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeRuntimeError, "Function test failed. Check server logs for details.")
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
	_ = json.NewEncoder(w).Encode(response)
}

// HandleGetReplay handles retrieving a shareable execution replay
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	execID := vars["execId"]

	// Get execution by public ID
	exec, err := h.Repo.GetExecutionPublicByID(r.Context(), execID)
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
	fn, err := h.Repo.GetFunctionByID(r.Context(), exec.FunctionID)
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
	_ = json.NewEncoder(w).Encode(response)
}

// HandleVerifyReplay handles manual verification of a specific execution replay
func (h *Handler) HandleVerifyReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	execID := vars["execId"]

	// Get execution by public ID
	exec, err := h.Repo.GetExecutionPublicByID(r.Context(), execID)
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
	_ = json.NewEncoder(w).Encode(response)
}
