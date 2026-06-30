package trustapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ============================================
// Trust Score Endpoints
// ============================================

// HandleGetTrustScore handles GET /v1/trust/score/{function_id}
// Returns the trust score for a specific function
func (h *Handler) HandleGetTrustScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	// Get trust score from registry
	history, err := h.registryRepo.GetLatestTrustHistory(r.Context(), functionID)
	if err != nil {
		h.logger.WithError(err).WithField("function_id", functionID).Error("Failed to get trust history")
		h.writeError(w, http.StatusNotFound, "Trust score not found for function", "trust_not_found")
		return
	}

	// If no trust history exists, calculate it on-demand
	if history == nil {
		windowStart := time.Now().Add(-24 * time.Hour)
		history, err = h.registryRepo.CalculateTrustScore(r.Context(), functionID, windowStart, time.Now())
		if err != nil {
			h.logger.WithError(err).Error("Failed to calculate trust score")
			h.writeError(w, http.StatusInternalServerError, "Failed to calculate trust score", "calculation_error")
			return
		}
	}

	// Build response - with explicit type conversions
	response := trustapi.TrustScoreResponse{
		FunctionID:         functionID,
		TrustScore:        history.TrustScore,
		TrustTier:         string(history.TrustTier),
		IsVerified:        history.IsVerified,
		VerificationLevel: history.VerificationLevel,
		LastUpdated:       history.CalculatedAt,
	}

	// Set component scores
	response.Components.Reliability = history.ReliabilityScore
	response.Components.Latency = history.LatencyScore
	response.Components.ErrorRate = history.ErrorRateScore
	response.Components.UserRating = history.UserRatingScore
	response.Components.Verification = history.VerificationBonus

	// Set metrics - with type conversions
	response.Metrics.TotalCalls = int64(history.TotalCalls)
	response.Metrics.SuccessRate = history.SuccessRate
	response.Metrics.P50LatencyMs = float64(history.P50LatencyMs)
	response.Metrics.P95LatencyMs = float64(history.P95LatencyMs)
	response.Metrics.P99LatencyMs = float64(history.P99LatencyMs)
	response.Metrics.ErrorRate = history.ErrorRate
	response.Metrics.TimeoutRate = history.TimeoutRate

	h.writeJSON(w, http.StatusOK, response)
}

// HandleBatchTrustScore handles POST /v1/trust/batch
// Returns trust scores for multiple functions
func (h *Handler) HandleBatchTrustScore(w http.ResponseWriter, r *http.Request) {
	var req trustapi.BatchTrustScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	if len(req.FunctionIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "Function IDs are required", "missing_function_ids")
		return
	}

	if len(req.FunctionIDs) > 100 {
		h.writeError(w, http.StatusBadRequest, "Maximum 100 function IDs per batch", "batch_too_large")
		return
	}

	scores := make([]trustapi.TrustScoreResponse, 0, len(req.FunctionIDs))
	errors := make([]trustapi.BatchTrustScoreError, 0)

	for _, functionID := range req.FunctionIDs {
		history, err := h.registryRepo.GetLatestTrustHistory(r.Context(), functionID)
		if err != nil {
			errors = append(errors, trustapi.BatchTrustScoreError{
				FunctionID: functionID,
				Error:      "Trust score not found",
			})
			continue
		}

		// If no trust history exists, calculate it on-demand
		if history == nil {
			windowStart := time.Now().Add(-24 * time.Hour)
			history, err = h.registryRepo.CalculateTrustScore(r.Context(), functionID, windowStart, time.Now())
			if err != nil {
				errors = append(errors, trustapi.BatchTrustScoreError{
					FunctionID: functionID,
					Error:      "Failed to calculate trust score",
				})
				continue
			}
		}

		score := trustapi.TrustScoreResponse{
			FunctionID:         functionID,
			TrustScore:        history.TrustScore,
			TrustTier:         string(history.TrustTier),
			IsVerified:        history.IsVerified,
			VerificationLevel: history.VerificationLevel,
			LastUpdated:       history.CalculatedAt,
		}

		score.Components.Reliability = history.ReliabilityScore
		score.Components.Latency = history.LatencyScore
		score.Components.ErrorRate = history.ErrorRateScore
		score.Components.UserRating = history.UserRatingScore
		score.Components.Verification = history.VerificationBonus

		score.Metrics.TotalCalls = int64(history.TotalCalls)
		score.Metrics.SuccessRate = history.SuccessRate
		score.Metrics.P50LatencyMs = float64(history.P50LatencyMs)
		score.Metrics.P95LatencyMs = float64(history.P95LatencyMs)
		score.Metrics.P99LatencyMs = float64(history.P99LatencyMs)
		score.Metrics.ErrorRate = history.ErrorRate
		score.Metrics.TimeoutRate = history.TimeoutRate

		scores = append(scores, score)
	}

	response := trustapi.BatchTrustScoreResponse{
		Scores: scores,
		Errors: errors,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetTrustHistory handles GET /v1/trust/history/{function_id}
// Returns the trust score history for a function
func (h *Handler) HandleGetTrustHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Get trust history
	history, total, err := h.registryRepo.GetTrustHistory(r.Context(), functionID, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trust history")
		h.writeError(w, http.StatusInternalServerError, "Failed to get trust history", "internal_error")
		return
	}

	// Convert to response format
	historyItems := make([]trustapi.TrustHistoryItem, len(history))
	for i, hist := range history {
		historyItems[i] = trustapi.TrustHistoryItem{
			TrustScore:   hist.TrustScore,
			TrustTier:    string(hist.TrustTier),
			IsVerified:   hist.IsVerified,
			CalculatedAt: hist.CalculatedAt,
			WindowStart:  hist.WindowStart,
			WindowEnd:    hist.WindowEnd,
		}
	}

	response := trustapi.TrustHistoryResponse{
		FunctionID: functionID,
		History:   historyItems,
		TotalCount: int64(total),
		Page:      page,
		PageSize:  pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ============================================
// Verification Endpoints
// ============================================

// HandleSubmitVerification handles POST /v1/trust/verify
// Submits a function for verification
func (h *Handler) HandleSubmitVerification(w http.ResponseWriter, r *http.Request) {
	var req trustapi.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate function exists
	fn, err := h.registryRepo.GetFunctionByID(r.Context(), req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get partner from context (set by auth middleware)
	partner := getPartnerFromContext(r)
	if partner == nil {
		h.writeError(w, http.StatusUnauthorized, "Partner not authenticated", "unauthorized")
		return
	}

	// Create verification request
	metadataJSON, err := mustMarshalJSON(req.Metadata)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid metadata format", "invalid_metadata")
		return
	}
	verification := &trustapi.TrustAPIVerification{
		PartnerID:          partner.ID,
		FunctionID:         req.FunctionID,
		FunctionAuthor:     fn.Author,
		FunctionName:       fn.Name,
		FunctionVersion:    req.FunctionVersion,
		VerificationLevel:  string(req.VerificationLevel),
		Metadata:           metadataJSON,
		Status:            string(trustapi.VerificationStatusPending),
	}

	if err := h.trustRepo.CreateVerification(verification); err != nil {
		h.logger.WithError(err).Error("Failed to create verification request")
		h.writeError(w, http.StatusInternalServerError, "Failed to create verification request", "internal_error")
		return
	}

	response := trustapi.VerificationResponse{
		ID:                verification.ID,
		VerificationID:    verification.VerificationID,
		FunctionID:        verification.FunctionID,
		FunctionAuthor:    verification.FunctionAuthor,
		FunctionName:      verification.FunctionName,
		FunctionVersion:   verification.FunctionVersion,
		VerificationLevel: verification.VerificationLevel,
		Status:            verification.Status,
		CreatedAt:         verification.CreatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleGetVerification handles GET /v1/trust/verify/{verification_id}
// Returns the status of a verification request
func (h *Handler) HandleGetVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	verificationID := vars["verification_id"]

	verification, err := h.trustRepo.GetVerificationByVerificationID(verificationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Verification not found", "verification_not_found")
		return
	}

	// Ownership check: API key partner must own this verification
	partner := getPartnerFromContext(r)
	if partner != nil && verification.PartnerID != partner.ID {
		h.writeError(w, http.StatusForbidden, "Not authorized to view this verification", "forbidden")
		return
	}

	response := trustapi.VerificationResponse{
		ID:                    verification.ID,
		VerificationID:        verification.VerificationID,
		FunctionID:            verification.FunctionID,
		FunctionAuthor:        verification.FunctionAuthor,
		FunctionName:          verification.FunctionName,
		FunctionVersion:       verification.FunctionVersion,
		VerificationLevel:     verification.VerificationLevel,
		Status:                verification.Status,
		TrustScore:            verification.TrustScore,
		TrustTier:             verification.TrustTier,
		VerificationBadgeURL: verification.VerificationBadgeURL,
		CreatedAt:             verification.CreatedAt,
		CompletedAt:           verification.CompletedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ============================================
// Report Endpoints
// ============================================

// HandleSubmitReport handles POST /v1/trust/report
// Submits a trust issue report
func (h *Handler) HandleSubmitReport(w http.ResponseWriter, r *http.Request) {
	var req trustapi.ReportCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate function exists
	fn, err := h.registryRepo.GetFunctionByID(r.Context(), req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get partner from context (set by auth middleware)
	partner := getPartnerFromContext(r)
	if partner == nil {
		h.writeError(w, http.StatusUnauthorized, "Partner not authenticated", "unauthorized")
		return
	}

	// Create report
	evidenceJSON, err := mustMarshalJSON(req.Evidence)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid evidence format", "invalid_evidence")
		return
	}
	report := &trustapi.TrustAPIReport{
		PartnerID:     partner.ID,
		FunctionID:    req.FunctionID,
		FunctionAuthor: fn.Author,
		FunctionName:  fn.Name,
		ReportType:    req.ReportType,
		Severity:     req.Severity,
		Title:        req.Title,
		Description:  req.Description,
		Evidence:     evidenceJSON,
		Status:       string(trustapi.ReportStatusPending),
	}

	if err := h.trustRepo.CreateReport(report); err != nil {
		h.logger.WithError(err).Error("Failed to create trust report")
		h.writeError(w, http.StatusInternalServerError, "Failed to create report", "internal_error")
		return
	}

	response := trustapi.ReportResponse{
		ID:             report.ID,
		ReportID:       report.ReportID,
		FunctionID:     report.FunctionID,
		FunctionAuthor: report.FunctionAuthor,
		FunctionName:   report.FunctionName,
		ReportType:     report.ReportType,
		Severity:       report.Severity,
		Title:          report.Title,
		Description:    report.Description,
		Status:         report.Status,
		CreatedAt:      report.CreatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleGetReport handles GET /v1/trust/report/{report_id}
// Returns the status of a trust report
func (h *Handler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["report_id"]

	report, err := h.trustRepo.GetReportByReportID(reportID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Report not found", "report_not_found")
		return
	}

	// Ownership check: API key partner must own this report
	partner := getPartnerFromContext(r)
	if partner != nil && report.PartnerID != partner.ID {
		h.writeError(w, http.StatusForbidden, "Not authorized to view this report", "forbidden")
		return
	}

	response := trustapi.ReportResponse{
		ID:             report.ID,
		ReportID:       report.ReportID,
		FunctionID:     report.FunctionID,
		FunctionAuthor: report.FunctionAuthor,
		FunctionName:   report.FunctionName,
		ReportType:     report.ReportType,
		Severity:       report.Severity,
		Title:          report.Title,
		Description:    report.Description,
		Status:         report.Status,
		CreatedAt:      report.CreatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// mustMarshalJSON marshals data to JSON, panics on error
func mustMarshalJSON(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}
