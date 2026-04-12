package trustapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ExtendedHandler extends the base Handler with revocation and attestation handlers
type ExtendedHandler struct {
	*Handler
	revocationRepo *trustapi.RevocationRepository
	webhookService *trustapi.WebhookService
}

// NewExtendedHandler creates a new extended handler
func NewExtendedHandler(repo *trustapi.Repository, registryRepo *registry.RegistryRepository, revocationRepo *trustapi.RevocationRepository, webhookService *trustapi.WebhookService) *ExtendedHandler {
	return &ExtendedHandler{
		Handler:        NewHandler(repo, registryRepo),
		revocationRepo: revocationRepo,
		webhookService: webhookService,
	}
}

// ============================================
// Trust Revocation Handlers
// ============================================

// HandleRevokeTrust handles POST /v1/trust/revoke
// Revokes trust for a function, marking it as untrusted
func (h *ExtendedHandler) HandleRevokeTrust(w http.ResponseWriter, r *http.Request) {
	var req trustapi.RevocationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate function exists
	fn, err := h.registryRepo.GetFunctionByID(req.FunctionID)
	if err != nil {
		h.logger.WithError(err).WithField("function_id", req.FunctionID).Error("Function not found")
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get user from context
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Get current trust state for restoration later
	trustState, err := h.registryRepo.GetLatestTrustHistory(req.FunctionID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get current trust state for revocation")
		// Continue anyway, will use defaults
	}

	// Check if there's already an active revocation
	existingRevocation, err := h.revocationRepo.GetActiveRevocationForFunction(req.FunctionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check existing revocations")
		h.writeError(w, http.StatusInternalServerError, "Failed to check existing revocations", "internal_error")
		return
	}
	if existingRevocation != nil {
		h.writeError(w, http.StatusConflict, "Function already has an active revocation", "already_revoked")
		return
	}

	// Create revocation record
	revocation := &trustapi.TrustRevocation{
		FunctionID:         req.FunctionID,
		FunctionAuthor:     fn.Author,
		FunctionName:       fn.Name,
		Reason:             req.Reason,
		ReasonDetails:      req.ReasonDetails,
		Severity:           req.Severity,
		RevocationType:     req.RevocationType,
		ImpactDescription:  req.ImpactDescription,
		RevokedBy:          claims.UserID,
		RevokedByType:      "admin",
		ReportID:           req.ReportID,
		OriginalTrustScore: trustState.TrustScore,
		OriginalTrustTier:  string(trustState.TrustTier),
		OriginalIsVerified: trustState.IsVerified,
	}

	// Set expiration if provided
	if req.ExpiresAt != nil {
		revocation.ExpiresAt = req.ExpiresAt
	}

	// Set evidence URLs
	if len(req.EvidenceURLs) > 0 {
		evidenceJSON, _ := json.Marshal(req.EvidenceURLs)
		revocation.EvidenceURLs = evidenceJSON
	}
	if req.DocumentationURL != "" {
		revocation.DocumentationURL = req.DocumentationURL
	}

	if err := h.revocationRepo.CreateRevocation(revocation); err != nil {
		h.logger.WithError(err).Error("Failed to create revocation")
		h.writeError(w, http.StatusInternalServerError, "Failed to revoke trust", "internal_error")
		return
	}

	// Update function's trust score to untrusted
	if err := h.registryRepo.UpdateFunctionTrustScore(req.FunctionID, 0, registry.TrustTierUntrusted); err != nil {
		h.logger.WithError(err).Warn("Failed to update function trust score after revocation")
		// Don't fail the request, but log it
	}

	// Invalidate cached evaluations
	if err := h.revocationRepo.InvalidateCacheForFunction(req.FunctionID); err != nil {
		h.logger.WithError(err).Warn("Failed to invalidate cache for revoked function")
	}

	// Log the action
	h.logger.WithFields(logrus.Fields{
		"revocation_id": revocation.RevocationID,
		"function_id":   req.FunctionID,
		"function_name": fn.Name,
		"reason":        req.Reason,
		"revoked_by":    claims.UserID,
	}).Info("Trust revoked for function")

	// Trigger webhooks
	if h.webhookService != nil {
		webhookData := map[string]interface{}{
			"revocation_id":        revocation.RevocationID,
			"function_id":          req.FunctionID.String(),
			"function_author":      fn.Author,
			"function_name":        fn.Name,
			"reason":               revocation.Reason,
			"severity":             revocation.Severity,
			"revocation_type":      revocation.RevocationType,
			"original_trust_score": revocation.OriginalTrustScore,
			"original_trust_tier":  revocation.OriginalTrustTier,
			"revoked_by":           claims.UserID.String(),
			"revoked_at":           revocation.RevokedAt,
		}
		go h.webhookService.TriggerEvent(trustapi.WebhookEventRevocationCreated, &req.FunctionID, webhookData)
	}

	// Send response
	response := trustapi.RevocationResponse{
		ID:                 revocation.ID,
		RevocationID:       revocation.RevocationID,
		FunctionID:         revocation.FunctionID,
		FunctionAuthor:     revocation.FunctionAuthor,
		FunctionName:       revocation.FunctionName,
		Reason:             revocation.Reason,
		ReasonDetails:      revocation.ReasonDetails,
		Severity:           revocation.Severity,
		Status:             revocation.Status,
		RevocationType:     revocation.RevocationType,
		RevokedAt:          revocation.RevokedAt,
		RevokedBy:          revocation.RevokedBy,
		RevokedByType:      revocation.RevokedByType,
		OriginalTrustScore: revocation.OriginalTrustScore,
		OriginalTrustTier:  revocation.OriginalTrustTier,
		ExpiresAt:          revocation.ExpiresAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleUnrevokeTrust handles POST /v1/trust/revoke/{revocation_id}/lift
// Lifts a trust revocation, restoring the function's original trust state
func (h *ExtendedHandler) HandleUnrevokeTrust(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	revocationIDStr := vars["revocation_id"]

	// Get the revocation record
	revocation, err := h.revocationRepo.GetRevocationByRevocationID(revocationIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Revocation not found", "revocation_not_found")
		return
	}

	// Check if already lifted
	if revocation.Status != string(trustapi.RevocationStatusActive) {
		h.writeError(w, http.StatusConflict, "Revocation is not active", "not_active")
		return
	}

	// Parse request
	var req trustapi.RevocationLiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Get user from context
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Lift the revocation
	if err := h.revocationRepo.LiftRevocation(revocation.ID, claims.UserID, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to lift revocation")
		h.writeError(w, http.StatusInternalServerError, "Failed to lift revocation", "internal_error")
		return
	}

	// Restore original trust state
	originalTier := registry.TrustTier(revocation.OriginalTrustTier)
	if originalTier == "" {
		originalTier = registry.TrustTierTrusted
	}

	if err := h.registryRepo.UpdateFunctionTrustScore(revocation.FunctionID, revocation.OriginalTrustScore, originalTier); err != nil {
		h.logger.WithError(err).Warn("Failed to restore function trust score after lifting revocation")
	}

	// Invalidate cached evaluations
	if err := h.revocationRepo.InvalidateCacheForFunction(revocation.FunctionID); err != nil {
		h.logger.WithError(err).Warn("Failed to invalidate cache after lifting revocation")
	}

	// Log the action
	h.logger.WithFields(logrus.Fields{
		"revocation_id": revocation.RevocationID,
		"function_id":   revocation.FunctionID,
		"lifted_by":     claims.UserID,
		"reason":        req.Reason,
	}).Info("Trust revocation lifted")

	// Trigger webhooks
	if h.webhookService != nil {
		webhookData := map[string]interface{}{
			"revocation_id":        revocation.RevocationID,
			"function_id":          revocation.FunctionID.String(),
			"function_author":      revocation.FunctionAuthor,
			"function_name":        revocation.FunctionName,
			"original_reason":      revocation.Reason,
			"lift_reason":          req.Reason,
			"lifted_by":            claims.UserID.String(),
			"restored_trust_score": revocation.OriginalTrustScore,
			"restored_trust_tier":  revocation.OriginalTrustTier,
		}
		go h.webhookService.TriggerEvent(trustapi.WebhookEventRevocationLifted, &revocation.FunctionID, webhookData)
	}

	// Send response
	response := trustapi.RevocationResponse{
		ID:                 revocation.ID,
		RevocationID:       revocation.RevocationID,
		FunctionID:         revocation.FunctionID,
		FunctionAuthor:     revocation.FunctionAuthor,
		FunctionName:       revocation.FunctionName,
		Reason:             revocation.Reason,
		Severity:           revocation.Severity,
		Status:             string(trustapi.RevocationStatusLifted),
		RevokedAt:          revocation.RevokedAt,
		RevokedBy:          revocation.RevokedBy,
		RevokedByType:      revocation.RevokedByType,
		LiftReason:         req.Reason,
		OriginalTrustScore: revocation.OriginalTrustScore,
		OriginalTrustTier:  revocation.OriginalTrustTier,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetRevocation handles GET /v1/trust/revoke/{revocation_id}
// Gets details of a specific revocation
func (h *ExtendedHandler) HandleGetRevocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	revocationIDStr := vars["revocation_id"]

	revocation, err := h.revocationRepo.GetRevocationByRevocationID(revocationIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Revocation not found", "revocation_not_found")
		return
	}

	response := trustapi.RevocationResponse{
		ID:                 revocation.ID,
		RevocationID:       revocation.RevocationID,
		FunctionID:         revocation.FunctionID,
		FunctionAuthor:     revocation.FunctionAuthor,
		FunctionName:       revocation.FunctionName,
		Reason:             revocation.Reason,
		ReasonDetails:      revocation.ReasonDetails,
		Severity:           revocation.Severity,
		Status:             revocation.Status,
		RevocationType:     revocation.RevocationType,
		RevokedAt:          revocation.RevokedAt,
		RevokedBy:          revocation.RevokedBy,
		RevokedByType:      revocation.RevokedByType,
		LiftedAt:           revocation.LiftedAt,
		LiftReason:         revocation.LiftReason,
		OriginalTrustScore: revocation.OriginalTrustScore,
		OriginalTrustTier:  revocation.OriginalTrustTier,
		ExpiresAt:          revocation.ExpiresAt,
		AppealStatus:       revocation.AppealStatus,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleListRevocations handles GET /v1/trust/revoked
// Lists all trust revocations with filtering
func (h *ExtendedHandler) HandleListRevocations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	functionIDStr := r.URL.Query().Get("function_id")
	status := r.URL.Query().Get("status")
	reason := r.URL.Query().Get("reason")
	severity := r.URL.Query().Get("severity")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Parse function_id if provided
	var functionID *uuid.UUID
	if functionIDStr != "" {
		fid, err := uuid.Parse(functionIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid function_id", "invalid_function_id")
			return
		}
		functionID = &fid
	}

	// List revocations
	revocations, total, err := h.revocationRepo.ListRevocations(functionID, status, reason, severity, nil, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list revocations")
		h.writeError(w, http.StatusInternalServerError, "Failed to list revocations", "internal_error")
		return
	}

	// Convert to response format
	responseList := make([]trustapi.RevocationResponse, len(revocations))
	for i, revocation := range revocations {
		responseList[i] = trustapi.RevocationResponse{
			ID:                 revocation.ID,
			RevocationID:       revocation.RevocationID,
			FunctionID:         revocation.FunctionID,
			FunctionAuthor:     revocation.FunctionAuthor,
			FunctionName:       revocation.FunctionName,
			Reason:             revocation.Reason,
			ReasonDetails:      revocation.ReasonDetails,
			Severity:           revocation.Severity,
			Status:             revocation.Status,
			RevocationType:     revocation.RevocationType,
			RevokedAt:          revocation.RevokedAt,
			RevokedBy:          revocation.RevokedBy,
			RevokedByType:      revocation.RevokedByType,
			LiftedAt:           revocation.LiftedAt,
			LiftReason:         revocation.LiftReason,
			OriginalTrustScore: revocation.OriginalTrustScore,
			OriginalTrustTier:  revocation.OriginalTrustTier,
			ExpiresAt:          revocation.ExpiresAt,
			AppealStatus:       revocation.AppealStatus,
		}
	}

	response := trustapi.RevocationListResponse{
		Revocations: responseList,
		TotalCount:  total,
		Page:        page,
		PageSize:    pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleCheckFunctionRevoked handles GET /v1/trust/revoked/{function_id}
// Checks if a specific function is currently revoked
func (h *ExtendedHandler) HandleCheckFunctionRevoked(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	revocation, err := h.revocationRepo.GetActiveRevocationForFunction(functionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check revocation status")
		h.writeError(w, http.StatusInternalServerError, "Failed to check status", "internal_error")
		return
	}

	if revocation == nil {
		// Function is not revoked
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"function_id": functionID,
			"is_revoked":  false,
		})
		return
	}

	// Function is revoked
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"function_id":        functionID,
		"is_revoked":         true,
		"revocation_id":      revocation.RevocationID,
		"reason":             revocation.Reason,
		"severity":           revocation.Severity,
		"revoked_at":         revocation.RevokedAt,
		"revocation_type":    revocation.RevocationType,
		"impact_description": revocation.ImpactDescription,
	})
}

// ============================================
// Attestation Handlers
// ============================================

// HandleGetAttestations handles GET /v1/trust/attestations
// Lists attestations for a function
func (h *ExtendedHandler) HandleGetAttestations(w http.ResponseWriter, r *http.Request) {
	functionIDStr := r.URL.Query().Get("function_id")
	attestationType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if functionIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "function_id is required", "missing_function_id")
		return
	}

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function_id", "invalid_function_id")
		return
	}

	includeRevoked := r.URL.Query().Get("include_revoked") == "true"

	attestations, total, err := h.revocationRepo.ListAttestationsForFunction(functionID, attestationType, status, includeRevoked, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list attestations")
		h.writeError(w, http.StatusInternalServerError, "Failed to list attestations", "internal_error")
		return
	}

	// Convert to response format
	responseList := make([]trustapi.AttestationResponse, len(attestations))
	for i, att := range attestations {
		responseList[i] = trustapi.AttestationResponse{
			ID:                att.ID,
			AttestationID:     att.AttestationID,
			FunctionID:        att.FunctionID,
			FunctionVersion:   att.FunctionVersion,
			FunctionAuthor:    att.FunctionAuthor,
			FunctionName:      att.FunctionName,
			Type:              att.Type,
			Status:            att.Status,
			Title:             att.Title,
			Description:       att.Description,
			AttesterID:        att.AttesterID,
			AttesterType:      att.AttesterType,
			AttesterName:      att.AttesterName,
			VerificationLevel: att.VerificationLevel,
			ProofHash:         att.ProofHash,
			AttestedAt:        att.AttestedAt,
			ValidUntil:        att.ValidUntil,
			RevokedAt:         att.RevokedAt,
			RevokeReason:      att.RevokeReason,
		}
	}

	response := trustapi.AttestationListResponse{
		Attestations: responseList,
		TotalCount:   total,
		Page:         page,
		PageSize:     pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetAttestation handles GET /v1/trust/attestations/{attestation_id}
// Gets a specific attestation
func (h *ExtendedHandler) HandleGetAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attestationID := vars["attestation_id"]

	attestation, err := h.revocationRepo.GetAttestationByAttestationID(attestationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Attestation not found", "attestation_not_found")
		return
	}

	// Verify integrity
	isValid := attestation.VerifyIntegrity()

	response := map[string]interface{}{
		"id":                 attestation.ID,
		"attestation_id":     attestation.AttestationID,
		"function_id":        attestation.FunctionID,
		"function_version":   attestation.FunctionVersion,
		"function_author":    attestation.FunctionAuthor,
		"function_name":      attestation.FunctionName,
		"type":               attestation.Type,
		"status":             attestation.Status,
		"title":              attestation.Title,
		"description":        attestation.Description,
		"attester_id":        attestation.AttesterID,
		"attester_type":      attestation.AttesterType,
		"attester_name":      attestation.AttesterName,
		"verification_level": attestation.VerificationLevel,
		"proof_hash":         attestation.ProofHash,
		"attested_at":        attestation.AttestedAt,
		"valid_until":        attestation.ValidUntil,
		"revoked_at":         attestation.RevokedAt,
		"revoke_reason":      attestation.RevokeReason,
		"integrity_verified": isValid,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleVerifyAttestation handles GET /v1/trust/attestations/{attestation_id}/verify
// Verifies the cryptographic integrity of an attestation
func (h *ExtendedHandler) HandleVerifyAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attestationID := vars["attestation_id"]

	isValid, err := h.revocationRepo.VerifyAttestationIntegrity(attestationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify attestation")
		h.writeError(w, http.StatusInternalServerError, "Failed to verify attestation", "verification_error")
		return
	}

	response := map[string]interface{}{
		"attestation_id":     attestationID,
		"integrity_verified": isValid,
		"verified_at":        time.Now(),
	}

	if !isValid {
		response["warning"] = "Attestation integrity check failed - data may have been tampered with"
		w.WriteHeader(http.StatusOK) // Still return 200 with warning
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetAttestationChain handles GET /v1/trust/attestations/{function_id}/chain
// Gets the full chain of attestations for a function (for audit)
func (h *ExtendedHandler) HandleGetAttestationChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	attestations, err := h.revocationRepo.GetAttestationChain(functionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get attestation chain")
		h.writeError(w, http.StatusInternalServerError, "Failed to get chain", "internal_error")
		return
	}

	// Build chain with integrity verification
	chain := make([]map[string]interface{}, len(attestations))
	for i, att := range attestations {
		chain[i] = map[string]interface{}{
			"attestation_id":     att.AttestationID,
			"type":               att.Type,
			"title":              att.Title,
			"status":             att.Status,
			"attester_type":      att.AttesterType,
			"attested_at":        att.AttestedAt,
			"proof_hash":         att.ProofHash,
			"previous_hash":      att.PreviousHash,
			"integrity_verified": att.VerifyIntegrity(),
		}
	}

	response := map[string]interface{}{
		"function_id":  functionID,
		"chain_length": len(chain),
		"attestations": chain,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ============================================
// Trust Policy Handlers
// ============================================

// HandleCreatePolicy handles POST /v1/trust/policies
// Creates a new trust policy
func (h *ExtendedHandler) HandleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req trustapi.PolicyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Get user from context
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Marshal rules to JSON
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid rules format", "invalid_rules")
		return
	}

	policy := &trustapi.TrustPolicy{
		Name:          req.Name,
		Description:   req.Description,
		Rules:         rulesJSON,
		DefaultAction: req.DefaultAction,
		OwnerID:       claims.UserID,
		OwnerType:     "user",
		CreatedBy:     claims.UserID,
		ValidFrom:     time.Now(),
		ValidUntil:    req.ValidUntil,
	}

	if err := h.revocationRepo.CreatePolicy(policy); err != nil {
		h.logger.WithError(err).Error("Failed to create policy")
		h.writeError(w, http.StatusInternalServerError, "Failed to create policy", "internal_error")
		return
	}

	// Parse rules for response
	var rules []trustapi.TrustPolicyRule
	json.Unmarshal(policy.Rules, &rules)

	response := trustapi.PolicyResponse{
		ID:            policy.ID,
		PolicyID:      policy.PolicyID,
		Name:          policy.Name,
		Description:   policy.Description,
		Version:       policy.Version,
		OwnerID:       policy.OwnerID,
		OwnerType:     policy.OwnerType,
		Rules:         rules,
		DefaultAction: policy.DefaultAction,
		Status:        policy.Status,
		IsDefault:     policy.IsDefault,
		ValidFrom:     policy.ValidFrom,
		ValidUntil:    policy.ValidUntil,
		CreatedAt:     policy.CreatedAt,
		UpdatedAt:     policy.UpdatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleGetPolicy handles GET /v1/trust/policies/{policy_id}
// Gets a specific policy
func (h *ExtendedHandler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policy, err := h.revocationRepo.GetPolicyByPolicyID(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Policy not found", "policy_not_found")
		return
	}

	rules, _ := policy.GetRules()

	response := trustapi.PolicyResponse{
		ID:            policy.ID,
		PolicyID:      policy.PolicyID,
		Name:          policy.Name,
		Description:   policy.Description,
		Version:       policy.Version,
		OwnerID:       policy.OwnerID,
		OwnerType:     policy.OwnerType,
		Rules:         rules,
		DefaultAction: policy.DefaultAction,
		Status:        policy.Status,
		IsDefault:     policy.IsDefault,
		UseCount:      policy.UseCount,
		LastUsedAt:    policy.LastUsedAt,
		ValidFrom:     policy.ValidFrom,
		ValidUntil:    policy.ValidUntil,
		CreatedAt:     policy.CreatedAt,
		UpdatedAt:     policy.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleListPolicies handles GET /v1/trust/policies
// Lists trust policies for the authenticated user
func (h *ExtendedHandler) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	status := r.URL.Query().Get("status")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	policies, total, err := h.revocationRepo.ListPoliciesForOwner(claims.UserID, "user", status, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list policies")
		h.writeError(w, http.StatusInternalServerError, "Failed to list policies", "internal_error")
		return
	}

	// Convert to response format
	responseList := make([]trustapi.PolicyResponse, len(policies))
	for i, policy := range policies {
		rules, _ := policy.GetRules()
		responseList[i] = trustapi.PolicyResponse{
			ID:            policy.ID,
			PolicyID:      policy.PolicyID,
			Name:          policy.Name,
			Description:   policy.Description,
			Version:       policy.Version,
			OwnerID:       policy.OwnerID,
			OwnerType:     policy.OwnerType,
			Rules:         rules,
			DefaultAction: policy.DefaultAction,
			Status:        policy.Status,
			IsDefault:     policy.IsDefault,
			UseCount:      policy.UseCount,
			LastUsedAt:    policy.LastUsedAt,
			ValidFrom:     policy.ValidFrom,
			ValidUntil:    policy.ValidUntil,
			CreatedAt:     policy.CreatedAt,
			UpdatedAt:     policy.UpdatedAt,
		}
	}

	response := trustapi.PolicyListResponse{
		Policies:   responseList,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleUpdatePolicy handles PUT /v1/trust/policies/{policy_id}
// Updates a trust policy
func (h *ExtendedHandler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policy, err := h.revocationRepo.GetPolicyByPolicyID(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Policy not found", "policy_not_found")
		return
	}

	// Verify ownership
	claims := middleware.GetUserFromContext(r)
	if claims == nil || policy.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to update this policy", "forbidden")
		return
	}

	var req trustapi.PolicyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Update fields
	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Description != "" {
		policy.Description = req.Description
	}
	if len(req.Rules) > 0 {
		rulesJSON, err := json.Marshal(req.Rules)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid rules format", "invalid_rules")
			return
		}
		policy.Rules = rulesJSON
		policy.Version++ // Increment version on rule change
	}
	if req.DefaultAction != "" {
		policy.DefaultAction = req.DefaultAction
	}
	if req.Status != "" {
		policy.Status = req.Status
	}
	if req.ValidUntil != nil {
		policy.ValidUntil = req.ValidUntil
	}

	if err := h.revocationRepo.UpdatePolicy(policy); err != nil {
		h.logger.WithError(err).Error("Failed to update policy")
		h.writeError(w, http.StatusInternalServerError, "Failed to update policy", "internal_error")
		return
	}

	rules, _ := policy.GetRules()

	response := trustapi.PolicyResponse{
		ID:            policy.ID,
		PolicyID:      policy.PolicyID,
		Name:          policy.Name,
		Description:   policy.Description,
		Version:       policy.Version,
		OwnerID:       policy.OwnerID,
		OwnerType:     policy.OwnerType,
		Rules:         rules,
		DefaultAction: policy.DefaultAction,
		Status:        policy.Status,
		IsDefault:     policy.IsDefault,
		UseCount:      policy.UseCount,
		LastUsedAt:    policy.LastUsedAt,
		ValidFrom:     policy.ValidFrom,
		ValidUntil:    policy.ValidUntil,
		CreatedAt:     policy.CreatedAt,
		UpdatedAt:     policy.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleDeletePolicy handles DELETE /v1/trust/policies/{policy_id}
// Deprecates a trust policy (soft delete)
func (h *ExtendedHandler) HandleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policy, err := h.revocationRepo.GetPolicyByPolicyID(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Policy not found", "policy_not_found")
		return
	}

	// Verify ownership
	claims := middleware.GetUserFromContext(r)
	if claims == nil || policy.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to delete this policy", "forbidden")
		return
	}

	if err := h.revocationRepo.DeprecatePolicy(policy.ID, claims.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to deprecate policy")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete policy", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Policy deprecated successfully",
		"policy_id": policyIDStr,
	})
}

// HandleEvaluatePolicy handles POST /v1/trust/policies/evaluate
// Evaluates a function against a trust policy
func (h *ExtendedHandler) HandleEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var req trustapi.PolicyEvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Get the policy to use
	var policy *trustapi.TrustPolicy
	var err error
	if req.PolicyID != nil {
		policy, err = h.revocationRepo.GetPolicyByPolicyID(*req.PolicyID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, "Policy not found", "policy_not_found")
			return
		}
		// Verify ownership
		if policy.OwnerID != claims.UserID {
			h.writeError(w, http.StatusForbidden, "Not authorized to use this policy", "forbidden")
			return
		}
	} else {
		// Use default policy
		policy, err = h.revocationRepo.GetDefaultPolicyForOwner(claims.UserID, "user")
		if err != nil {
			h.logger.WithError(err).Error("Failed to get default policy")
			h.writeError(w, http.StatusInternalServerError, "Failed to get default policy", "internal_error")
			return
		}
		if policy == nil {
			h.writeError(w, http.StatusNotFound, "No default policy found", "no_default_policy")
			return
		}
	}

	// Check for cached evaluation
	cachedEval, err := h.revocationRepo.GetCachedEvaluation(policy.ID, req.FunctionID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to check cached evaluation")
	}
	if cachedEval != nil {
		// Return cached result
		ruleResults := []trustapi.PolicyRuleResult{}
		json.Unmarshal(cachedEval.RuleResults, &ruleResults)

		response := trustapi.PolicyEvaluateResponse{
			EvaluationID:     cachedEval.EvaluationID,
			PolicyID:         policy.PolicyID,
			FunctionID:       cachedEval.FunctionID,
			FunctionAuthor:   cachedEval.FunctionAuthor,
			FunctionName:     cachedEval.FunctionName,
			Result:           cachedEval.Result,
			Decision:         cachedEval.Decision,
			Reason:           cachedEval.Reason,
			TrustScore:       cachedEval.TrustScore,
			TrustTier:        cachedEval.TrustTier,
			IsVerified:       cachedEval.IsVerified,
			IsRevoked:        cachedEval.IsRevoked,
			RevocationStatus: cachedEval.RevocationStatus,
			RuleResults:      ruleResults,
			EvaluatedAt:      cachedEval.EvaluatedAt,
			CacheValidUntil:  cachedEval.CacheValidUntil,
		}

		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"result": response,
			"cached": true,
		})
		return
	}

	// Get function details
	fn, err := h.registryRepo.GetFunctionByID(req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get trust score
	trustState, err := h.registryRepo.GetLatestTrustHistory(req.FunctionID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get trust history for evaluation")
		// Continue with zero trust score
		trustState = &registry.TrustHistory{
			TrustScore: 0,
			TrustTier:  registry.TrustTierUntrusted,
			IsVerified: false,
		}
	}

	// Check for active revocation
	isRevoked := false
	revocationStatus := ""
	revocation, err := h.revocationRepo.GetActiveRevocationForFunction(req.FunctionID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to check revocation status")
	}
	if revocation != nil {
		isRevoked = true
		revocationStatus = revocation.Status
	}

	// Get policy rules
	rules, err := policy.GetRules()
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse policy rules")
		h.writeError(w, http.StatusInternalServerError, "Failed to evaluate policy", "internal_error")
		return
	}

	// Evaluate each rule
	ruleResults := make([]trustapi.PolicyRuleResult, len(rules))
	overallResult := policy.DefaultAction
	if overallResult == "" {
		overallResult = "deny"
	}
	decision := "policy_default"
	reason := "Default policy action applied"

	for i, rule := range rules {
		passed, ruleDecision, ruleReason := h.evaluateRule(rule, trustState, isRevoked, revocationStatus, fn)
		ruleResults[i] = trustapi.PolicyRuleResult{
			RuleID:        rule.ID,
			Type:          rule.Type,
			Passed:        passed,
			Reason:        ruleReason,
			ActualValue:   getRuleActualValue(rule.Type, trustState, isRevoked),
			ExpectedValue: rule.Value,
		}

		if !passed {
			// Rule failed - apply action
			if ruleDecision == "deny" {
				overallResult = "denied"
				decision = ruleDecision
				reason = ruleReason
				break // Short circuit on deny
			} else if ruleDecision == "warn" && overallResult != "denied" {
				overallResult = "warned"
				decision = ruleDecision
				reason = ruleReason
			}
		} else if overallResult == "deny" || overallResult == "denied" {
			// Rule passed and we're currently denying - allow it
			overallResult = "allowed"
			decision = "policy_rule_passed"
			reason = fmt.Sprintf("Rule '%s' passed", rule.ID)
		}
	}

	// Create evaluation record
	ruleResultsJSON, _ := json.Marshal(ruleResults)
	eval := &trustapi.TrustPolicyEvaluation{
		PolicyID:         policy.ID,
		FunctionID:       req.FunctionID,
		FunctionAuthor:   fn.Author,
		FunctionName:     fn.Name,
		TrustScore:       trustState.TrustScore,
		TrustTier:        string(trustState.TrustTier),
		IsVerified:       trustState.IsVerified,
		IsRevoked:        isRevoked,
		RevocationStatus: revocationStatus,
		Result:           overallResult,
		Decision:         decision,
		Reason:           reason,
		RuleResults:      ruleResultsJSON,
		EvaluatedBy:      claims.UserID,
		EvaluatedByType:  "api",
		IsCached:         true,
		CacheValidUntil:  &[]time.Time{time.Now().Add(5 * time.Minute)}[0], // 5 minute cache
	}

	if err := h.revocationRepo.CreateEvaluation(eval); err != nil {
		h.logger.WithError(err).Warn("Failed to save evaluation")
	}

	// Increment policy use count
	if err := h.revocationRepo.IncrementPolicyUseCount(policy.ID); err != nil {
		h.logger.WithError(err).Warn("Failed to increment policy use count")
	}

	response := trustapi.PolicyEvaluateResponse{
		EvaluationID:     eval.EvaluationID,
		PolicyID:         policy.PolicyID,
		FunctionID:       req.FunctionID,
		FunctionAuthor:   fn.Author,
		FunctionName:     fn.Name,
		Result:           overallResult,
		Decision:         decision,
		Reason:           reason,
		TrustScore:       trustState.TrustScore,
		TrustTier:        string(trustState.TrustTier),
		IsVerified:       trustState.IsVerified,
		IsRevoked:        isRevoked,
		RevocationStatus: revocationStatus,
		RuleResults:      ruleResults,
		EvaluatedAt:      eval.EvaluatedAt,
		CacheValidUntil:  eval.CacheValidUntil,
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"result": response,
		"cached": false,
	})
}

// HandleBatchEvaluatePolicy handles POST /v1/trust/policies/evaluate/batch
// Batch evaluates multiple functions against a policy
func (h *ExtendedHandler) HandleBatchEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var req trustapi.BatchPolicyEvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Get the policy (reuse same policy for all evaluations)
	var policy *trustapi.TrustPolicy
	var err error
	if req.PolicyID != nil {
		policy, err = h.revocationRepo.GetPolicyByPolicyID(*req.PolicyID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, "Policy not found", "policy_not_found")
			return
		}
		if policy.OwnerID != claims.UserID {
			h.writeError(w, http.StatusForbidden, "Not authorized to use this policy", "forbidden")
			return
		}
	} else {
		policy, err = h.revocationRepo.GetDefaultPolicyForOwner(claims.UserID, "user")
		if err != nil || policy == nil {
			h.writeError(w, http.StatusNotFound, "No default policy found", "no_default_policy")
			return
		}
	}

	results := make([]trustapi.PolicyEvaluateResponse, 0, len(req.FunctionIDs))
	errors := make([]trustapi.BatchPolicyEvaluationError, 0)

	for _, functionID := range req.FunctionIDs {
		fn, err := h.registryRepo.GetFunctionByID(functionID)
		if err != nil {
			errors = append(errors, trustapi.BatchPolicyEvaluationError{
				FunctionID: functionID,
				Error:      "Function not found",
			})
			continue
		}

		// Get trust state
		trustState, _ := h.registryRepo.GetLatestTrustHistory(functionID)
		if trustState == nil {
			trustState = &registry.TrustHistory{
				TrustScore: 0,
				TrustTier:  registry.TrustTierUntrusted,
			}
		}

		// Check revocation
		isRevoked := false
		revocationStatus := ""
		revocation, _ := h.revocationRepo.GetActiveRevocationForFunction(functionID)
		if revocation != nil {
			isRevoked = true
			revocationStatus = revocation.Status
		}

		// Quick evaluation (simplified for batch)
		result := "allowed"
		if isRevoked {
			result = "denied"
		} else if trustState.TrustScore < 50 {
			result = "warned"
		}

		results = append(results, trustapi.PolicyEvaluateResponse{
			FunctionID:       functionID,
			FunctionAuthor:   fn.Author,
			FunctionName:     fn.Name,
			PolicyID:         policy.PolicyID,
			Result:           result,
			Decision:         "batch_quick_eval",
			TrustScore:       trustState.TrustScore,
			TrustTier:        string(trustState.TrustTier),
			IsVerified:       trustState.IsVerified,
			IsRevoked:        isRevoked,
			RevocationStatus: revocationStatus,
			EvaluatedAt:      time.Now(),
		})
	}

	response := trustapi.BatchPolicyEvaluateResponse{
		Results:     results,
		Errors:      errors,
		EvaluatedAt: time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// evaluateRule evaluates a single policy rule
func (h *ExtendedHandler) evaluateRule(
	rule trustapi.TrustPolicyRule,
	trustState *registry.TrustHistory,
	isRevoked bool,
	revocationStatus string,
	fn *storage.RegistryFunction,
) (passed bool, decision string, reason string) {
	switch rule.Type {
	case "min_trust_score":
		minScore, ok := rule.Value.(float64)
		if !ok {
			return false, "deny", "Invalid rule value"
		}
		if trustState.TrustScore >= minScore {
			return true, "", ""
		}
		return false, "deny", fmt.Sprintf("Trust score %.2f below minimum %.2f", trustState.TrustScore, minScore)

	case "verification_required":
		if trustState.IsVerified {
			return true, "", ""
		}
		return false, "deny", "Function is not verified"

	case "tier_minimum":
		minTier, ok := rule.Value.(string)
		if !ok {
			return false, "deny", "Invalid rule value"
		}
		tierRank := map[string]int{
			"untrusted":      0,
			"trusted":        1,
			"verified":       2,
			"highly_trusted": 3,
		}
		currentRank := tierRank[string(trustState.TrustTier)]
		requiredRank := tierRank[minTier]
		if currentRank >= requiredRank {
			return true, "", ""
		}
		return false, "deny", fmt.Sprintf("Trust tier %s below minimum %s", trustState.TrustTier, minTier)

	case "no_revocation":
		if !isRevoked {
			return true, "", ""
		}
		return false, "deny", fmt.Sprintf("Function is revoked: %s", revocationStatus)

	case "min_success_rate":
		minRate, ok := rule.Value.(float64)
		if !ok {
			return false, "deny", "Invalid rule value"
		}
		// Would need to get success rate from metrics
		successRate := 100.0 // Placeholder
		if successRate >= minRate {
			return true, "", ""
		}
		return false, "warn", fmt.Sprintf("Success rate %.2f below minimum %.2f", successRate, minRate)

	default:
		return false, "deny", fmt.Sprintf("Unknown rule type: %s", rule.Type)
	}
}

// getRuleActualValue gets the actual value for a rule type
func getRuleActualValue(ruleType string, trustState *registry.TrustHistory, isRevoked bool) interface{} {
	switch ruleType {
	case "min_trust_score":
		return trustState.TrustScore
	case "verification_required":
		return trustState.IsVerified
	case "tier_minimum":
		return string(trustState.TrustTier)
	case "no_revocation":
		return !isRevoked
	default:
		return nil
	}
}
