package trustapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
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
	auditLogRepo   *trustapi.AuditLogRepository
	merkleRepo     *trustapi.MerkleRepository
}

// NewExtendedHandler creates a new extended handler
func NewExtendedHandler(apikeyRepo *apikey.Repository, repo *trustapi.Repository, registryRepo *registry.RegistryRepository, revocationRepo *trustapi.RevocationRepository, webhookService *trustapi.WebhookService) *ExtendedHandler {
	return &ExtendedHandler{
		Handler:        NewHandler(apikeyRepo, repo, registryRepo),
		revocationRepo: revocationRepo,
		webhookService: webhookService,
	}
}

// SetAuditLogRepository sets the audit log repository for audit logging
func (h *ExtendedHandler) SetAuditLogRepository(repo *trustapi.AuditLogRepository) {
	h.auditLogRepo = repo
}

// SetMerkleRepository sets the Merkle audit trail repository
func (h *ExtendedHandler) SetMerkleRepository(repo *trustapi.MerkleRepository) {
	h.merkleRepo = repo
}

// StartRevocationExpirationJob starts a background goroutine that periodically
// checks for expired revocations and marks them as expired, restoring trust scores.
func (h *ExtendedHandler) StartRevocationExpirationJob(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				expired, err := h.revocationRepo.GetExpiredRevocations()
				if err != nil {
					h.logger.WithError(err).Error("Failed to get expired revocations")
					continue
				}
				for _, rev := range expired {
					if err := h.revocationRepo.ExpireRevocation(rev.ID); err != nil {
						h.logger.WithError(err).WithField("revocation_id", rev.RevocationID).Error("Failed to expire revocation")
						continue
					}
					// Restore original trust state
					originalTier := registry.TrustTier(rev.OriginalTrustTier)
					if originalTier == "" {
						originalTier = registry.TrustTierTrusted
					}
					if err := h.registryRepo.UpdateFunctionTrustScore(ctx, rev.FunctionID, rev.OriginalTrustScore, originalTier); err != nil {
						h.logger.WithError(err).Warn("Failed to restore trust score after expiration")
					}
					if err := h.revocationRepo.InvalidateCacheForFunction(rev.FunctionID); err != nil {
						h.logger.WithError(err).Warn("Failed to invalidate cache after expiration")
					}
					// Write audit log
					if h.auditLogRepo != nil {
						if err := h.auditLogRepo.LogRevocationLifted(&rev, uuid.Nil, "system", "expired", "", "", ""); err != nil {
							h.logger.WithError(err).Warn("Failed to write audit log for revocation expiration")
						}
					}
					h.logger.WithFields(logrus.Fields{
						"revocation_id": rev.RevocationID,
						"function_id":   rev.FunctionID,
					}).Info("Revocation expired and trust restored")
				}
			}
		}
	}()
}

// StartAttestationExpirationJob starts a background goroutine that periodically
// checks for attestations past their ValidUntil date and marks them as expired.
func (h *ExtendedHandler) StartAttestationExpirationJob(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := h.revocationRepo.ExpireStaleAttestations()
				if err != nil {
					h.logger.WithError(err).Error("Failed to expire stale attestations")
					continue
				}
				if count > 0 {
					h.logger.WithField("count", count).Info("Expired stale attestations")
				}
			}
		}
	}()
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
	fn, err := h.registryRepo.GetFunctionByID(r.Context(), req.FunctionID)
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
	trustState, err := h.registryRepo.GetLatestTrustHistory(r.Context(), req.FunctionID)
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
	if err := h.registryRepo.UpdateFunctionTrustScore(r.Context(), req.FunctionID, 0, registry.TrustTierUntrusted); err != nil {
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

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogRevocationCreated(revocation, claims.UserID, "admin", r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for revocation")
		}
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

	if err := h.registryRepo.UpdateFunctionTrustScore(r.Context(), revocation.FunctionID, revocation.OriginalTrustScore, originalTier); err != nil {
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

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogRevocationLifted(revocation, claims.UserID, "admin", req.Reason, r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for revocation lift")
		}
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

// HandleCreateAttestation handles POST /v1/trust/attestations
// Creates a new attestation for a function
func (h *ExtendedHandler) HandleCreateAttestation(w http.ResponseWriter, r *http.Request) {
	var req trustapi.AttestationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate required fields
	if req.FunctionID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "function_id is required", "missing_function_id")
		return
	}
	if req.Type == "" {
		h.writeError(w, http.StatusBadRequest, "type is required", "missing_type")
		return
	}
	if req.Title == "" {
		h.writeError(w, http.StatusBadRequest, "title is required", "missing_title")
		return
	}

	// Validate attestation type
	validTypes := map[trustapi.AttestationType]bool{
		trustapi.AttestationTypeVerification: true,
		trustapi.AttestationTypeSecurityScan: true,
		trustapi.AttestationTypeCodeReview:   true,
		trustapi.AttestationTypeExecution:    true,
		trustapi.AttestationTypeCompliance:   true,
		trustapi.AttestationTypeSignature:    true,
		trustapi.AttestationTypeDelegation:   true,
	}
	if !validTypes[req.Type] {
		h.writeError(w, http.StatusBadRequest, "Invalid attestation type. Must be one of: verification, security_scan, code_review, execution, compliance, signature", "invalid_type")
		return
	}

	// Validate function exists
	fn, err := h.registryRepo.GetFunctionByID(r.Context(), req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get user/partner identity from context
	claims := middleware.GetUserFromContext(r)
	partner := getPartnerFromContext(r)

	var attesterID uuid.UUID
	var attesterType, attesterName string

	if partner != nil {
		attesterID = partner.ID
		attesterType = "partner"
		attesterName = partner.Name
	} else if claims != nil {
		attesterID = claims.UserID
		attesterType = "user"
		if req.AttesterName != "" {
			attesterName = req.AttesterName
		} else {
			attesterName = claims.Email
		}
	} else {
		attesterID = uuid.Nil
		attesterType = "system"
		attesterName = "system"
	}

	if req.AttesterType != "" {
		attesterType = req.AttesterType
	}

	// Marshal results
	var resultsJSON json.RawMessage
	if req.Results != nil {
		var err error
		resultsJSON, err = json.Marshal(req.Results)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid results format", "invalid_results")
			return
		}
	} else {
		resultsJSON = json.RawMessage("{}")
	}

	// Build attestation
	attestation := &trustapi.TrustAttestation{
		FunctionID:        req.FunctionID,
		FunctionVersion:   req.FunctionVersion,
		FunctionAuthor:    fn.Author,
		FunctionName:      fn.Name,
		Type:              string(req.Type),
		Title:             req.Title,
		Description:       req.Description,
		Results:           resultsJSON,
		AttesterID:        attesterID,
		AttesterType:      attesterType,
		AttesterName:      attesterName,
		VerificationLevel: req.VerificationLevel,
		CodeHash:          req.CodeHash,
		InputHash:         req.InputHash,
		OutputHash:        req.OutputHash,
		SourceDataHash:    req.SourceDataHash,
		ValidUntil:        req.ValidUntil,
	}

	if partner != nil && partner.ID != uuid.Nil {
		attestation.AttesterPartnerID = &partner.ID
	}

	// Create the attestation (repository handles signing, proof hash, previous hash)
	if err := h.revocationRepo.CreateAttestation(attestation); err != nil {
		h.logger.WithError(err).Error("Failed to create attestation")
		h.writeError(w, http.StatusInternalServerError, "Failed to create attestation", "internal_error")
		return
	}

	// Append to Merkle audit trail
	if h.merkleRepo != nil {
		signer := trustapi.GetSigner()
		if _, err := h.merkleRepo.AppendLeaf(attestation, signer); err != nil {
			h.logger.WithError(err).Warn("Failed to append attestation to Merkle audit trail")
			// Don't fail — attestation was created successfully
		}
	}

	// Trigger webhooks
	if h.webhookService != nil {
		webhookData := map[string]interface{}{
			"attestation_id": attestation.AttestationID,
			"function_id":    req.FunctionID.String(),
			"function_name":  fn.Name,
			"type":           string(req.Type),
			"title":          req.Title,
			"attester_type":  attesterType,
			"attester_name":  attesterName,
			"proof_hash":     attestation.ProofHash,
			"attested_at":    attestation.AttestedAt,
		}
		go func() { _ = h.webhookService.TriggerEvent(trustapi.WebhookEventAttestationCreated, &req.FunctionID, webhookData) }()
	}

	// Write audit log
	if h.auditLogRepo != nil {
		actorID := attesterID
		if err := h.auditLogRepo.LogAttestationCreated(attestation, actorID, attesterType, r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for attestation creation")
		}
	}

	h.logger.WithFields(logrus.Fields{
		"attestation_id": attestation.AttestationID,
		"function_id":    req.FunctionID,
		"type":           string(req.Type),
		"attester_type":  attesterType,
	}).Info("Attestation created")

	response := trustapi.AttestationResponse{
		ID:                attestation.ID,
		AttestationID:     attestation.AttestationID,
		FunctionID:        attestation.FunctionID,
		FunctionVersion:   attestation.FunctionVersion,
		FunctionAuthor:    attestation.FunctionAuthor,
		FunctionName:      attestation.FunctionName,
		Type:              attestation.Type,
		Status:            attestation.Status,
		Title:             attestation.Title,
		Description:       attestation.Description,
		Results:           attestation.Results,
		AttesterID:        attestation.AttesterID,
		AttesterType:      attestation.AttesterType,
		AttesterName:      attestation.AttesterName,
		VerificationLevel: attestation.VerificationLevel,
		ProofHash:         attestation.ProofHash,
		Signature:         attestation.Signature,
		PublicKeyID:       attestation.PublicKeyID,
		CodeHash:          attestation.CodeHash,
		InputHash:         attestation.InputHash,
		OutputHash:        attestation.OutputHash,
		SourceDataHash:    attestation.SourceDataHash,
		PreviousHash:      attestation.PreviousHash,
		AttestedAt:        attestation.AttestedAt,
		ValidUntil:        attestation.ValidUntil,
		IsValid:           true,
		SignatureValid:    attestation.Signature != "",
		ChainValid:        true,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleRevokeAttestation handles POST /v1/trust/attestations/{attestation_id}/revoke
// Revokes an existing attestation
func (h *ExtendedHandler) HandleRevokeAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attestationID := vars["attestation_id"]

	// Get existing attestation
	attestation, err := h.revocationRepo.GetAttestationByAttestationID(attestationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Attestation not found", "attestation_not_found")
		return
	}

	// Check if already revoked
	if attestation.Status == string(trustapi.AttestationStatusRevoked) {
		h.writeError(w, http.StatusConflict, "Attestation is already revoked", "already_revoked")
		return
	}

	// Parse request
	var req trustapi.AttestationRevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	if req.Reason == "" {
		h.writeError(w, http.StatusBadRequest, "reason is required", "missing_reason")
		return
	}

	// Get user from context
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	// Revoke the attestation
	if err := h.revocationRepo.RevokeAttestation(attestation.ID, claims.UserID, req.Reason, req.RevocationID); err != nil {
		h.logger.WithError(err).Error("Failed to revoke attestation")
		h.writeError(w, http.StatusInternalServerError, "Failed to revoke attestation", "internal_error")
		return
	}

	// Trigger webhooks
	if h.webhookService != nil {
		webhookData := map[string]interface{}{
			"attestation_id": attestation.AttestationID,
			"function_id":    attestation.FunctionID.String(),
			"function_name":  attestation.FunctionName,
			"type":           attestation.Type,
			"title":          attestation.Title,
			"revoke_reason":  req.Reason,
			"revoked_by":     claims.UserID.String(),
		}
		go func() { _ = h.webhookService.TriggerEvent(trustapi.WebhookEventAttestationRevoked, &attestation.FunctionID, webhookData) }()
	}

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogAttestationRevoked(attestation, claims.UserID, "admin", req.Reason, r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for attestation revocation")
		}
	}

	h.logger.WithFields(logrus.Fields{
		"attestation_id": attestation.AttestationID,
		"function_id":    attestation.FunctionID,
		"reason":         req.Reason,
		"revoked_by":     claims.UserID,
	}).Info("Attestation revoked")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"attestation_id": attestation.AttestationID,
		"status":         string(trustapi.AttestationStatusRevoked),
		"revoked_at":     time.Now(),
		"revoke_reason":  req.Reason,
	})
}

// HandleGetPublicKey handles GET /v1/trust/attestations/public-key
// Returns the signing public key for external attestation verification
func (h *ExtendedHandler) HandleGetPublicKey(w http.ResponseWriter, r *http.Request) {
	signer := trustapi.GetSigner()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"public_key":   signer.PublicKeyHex(),
		"key_id":       signer.KeyID(),
		"algorithm":    string(signer.Algorithm()),
		"key_encoding": "hex",
	})
}

// HandleVerifyChain handles GET /v1/trust/attestations/chain/{function_id}/verify
// Verifies the full cryptographic chain of attestations for a function
func (h *ExtendedHandler) HandleVerifyChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	chainValid, chainLength, err := h.revocationRepo.VerifyAttestationChain(functionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify attestation chain")
		h.writeError(w, http.StatusInternalServerError, "Failed to verify chain", "verification_error")
		return
	}

	signer := trustapi.GetSigner()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"function_id":   functionID,
		"chain_valid":   chainValid,
		"chain_length":  chainLength,
		"verified_at":   time.Now(),
		"signing_key_id": signer.KeyID(),
		"algorithm":     string(signer.Algorithm()),
	})
}
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
			Results:           att.Results,
			AttesterID:        att.AttesterID,
			AttesterType:      att.AttesterType,
			AttesterName:      att.AttesterName,
			VerificationLevel: att.VerificationLevel,
			ProofHash:         att.ProofHash,
			Signature:         att.Signature,
			PublicKeyID:       att.PublicKeyID,
			CodeHash:          att.CodeHash,
			InputHash:         att.InputHash,
			OutputHash:        att.OutputHash,
			SourceDataHash:    att.SourceDataHash,
			PreviousHash:      att.PreviousHash,
			AttestedAt:        att.AttestedAt,
			ValidUntil:        att.ValidUntil,
			RevokedAt:         att.RevokedAt,
			RevokeReason:      att.RevokeReason,
			IsValid:           att.VerifyIntegrity(),
			SignatureValid:    att.Signature != "",
			ChainValid:        att.VerifyIntegrity(),
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

	// Verify integrity and signature
	isValid := attestation.VerifyIntegrity()
	signatureValid := attestation.Signature != ""

	if signatureValid {
		signer := trustapi.GetSigner()
		if verified, err := signer.VerifyAttestationSignature(attestation); err == nil {
			signatureValid = verified
		} else {
			signatureValid = false
		}
	}

	response := trustapi.AttestationResponse{
		ID:                attestation.ID,
		AttestationID:     attestation.AttestationID,
		FunctionID:        attestation.FunctionID,
		FunctionVersion:   attestation.FunctionVersion,
		FunctionAuthor:    attestation.FunctionAuthor,
		FunctionName:      attestation.FunctionName,
		Type:              attestation.Type,
		Status:            attestation.Status,
		Title:             attestation.Title,
		Description:       attestation.Description,
		Results:           attestation.Results,
		AttesterID:        attestation.AttesterID,
		AttesterType:      attestation.AttesterType,
		AttesterName:      attestation.AttesterName,
		VerificationLevel: attestation.VerificationLevel,
		ProofHash:         attestation.ProofHash,
		Signature:         attestation.Signature,
		PublicKeyID:       attestation.PublicKeyID,
		CodeHash:          attestation.CodeHash,
		InputHash:         attestation.InputHash,
		OutputHash:        attestation.OutputHash,
		SourceDataHash:    attestation.SourceDataHash,
		PreviousHash:      attestation.PreviousHash,
		AttestedAt:        attestation.AttestedAt,
		ValidUntil:        attestation.ValidUntil,
		RevokedAt:         attestation.RevokedAt,
		RevokeReason:      attestation.RevokeReason,
		IsValid:           isValid,
		SignatureValid:    signatureValid,
		ChainValid:        isValid,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleVerifyAttestation handles GET /v1/trust/attestations/{attestation_id}/verify
// Verifies the cryptographic integrity of an attestation
func (h *ExtendedHandler) HandleVerifyAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attestationID := vars["attestation_id"]

	attestation, err := h.revocationRepo.GetAttestationByAttestationID(attestationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Attestation not found", "attestation_not_found")
		return
	}

	integrityValid := attestation.VerifyIntegrity()

	// Verify signature
	signatureValid := false
	if attestation.Signature != "" {
		signer := trustapi.GetSigner()
		if verified, err := signer.VerifyAttestationSignature(attestation); err == nil && verified {
			signatureValid = true
		} else if attestation.PublicKeyID != "" && attestation.PublicKeyID != signer.KeyID() {
			// Try historical keys if current key doesn't match
			if historicalKey, err := trustapi.GetKeyByID(attestation.PublicKeyID); err == nil && historicalKey != nil {
				if verified, err := trustapi.VerifyAttestationSignatureWithKey(attestation, historicalKey.PublicKeyHex, historicalKey.Algorithm); err == nil {
					signatureValid = verified
				}
			}
		}
	}

	response := map[string]interface{}{
		"attestation_id":     attestationID,
		"integrity_verified": integrityValid,
		"signature_verified": signatureValid,
		"proof_hash":         attestation.ProofHash,
		"signature":          attestation.Signature,
		"public_key_id":      attestation.PublicKeyID,
		"algorithm":          string(trustapi.GetSigner().Algorithm()),
		"verified_at":        time.Now(),
	}

	if !integrityValid {
		response["warning"] = "Attestation integrity check failed - data may have been tampered with"
	} else if attestation.Signature != "" && !signatureValid {
		response["warning"] = "Attestation signature verification failed"
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

	// Verify full chain integrity
	chainIntegrityValid := true
	chain := make([]map[string]interface{}, len(attestations))
	for i, att := range attestations {
		individualValid := att.VerifyIntegrity()
		chainLinkValid := true

		if i > 0 && attestations[i-1].ProofHash != att.PreviousHash {
			chainLinkValid = false
			chainIntegrityValid = false
		}
		if i == 0 && att.PreviousHash != "" {
			chainLinkValid = false
			chainIntegrityValid = false
		}
		if !individualValid {
			chainIntegrityValid = false
		}

		chain[i] = map[string]interface{}{
			"attestation_id":     att.AttestationID,
			"type":               att.Type,
			"title":              att.Title,
			"status":             att.Status,
			"attester_type":      att.AttesterType,
			"attested_at":        att.AttestedAt,
			"proof_hash":         att.ProofHash,
			"previous_hash":      att.PreviousHash,
			"signature":          att.Signature,
			"integrity_verified": individualValid,
			"chain_link_valid":   chainLinkValid,
		}
	}

	response := map[string]interface{}{
		"function_id":       functionID,
		"chain_length":      len(chain),
		"chain_valid":       chainIntegrityValid,
		"signing_algorithm": string(trustapi.GetSigner().Algorithm()),
		"attestations":      chain,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleCompleteVerification handles POST /v1/trust/verify/{verification_id}/complete
// Completes a verification request and automatically creates a verification attestation
func (h *ExtendedHandler) HandleCompleteVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	verificationID := vars["verification_id"]

	// Get existing verification
	verification, err := h.trustRepo.GetVerificationByVerificationID(verificationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Verification not found", "verification_not_found")
		return
	}

	// Check if already completed
	if verification.Status == string(trustapi.VerificationStatusCompleted) {
		h.writeError(w, http.StatusConflict, "Verification is already completed", "already_completed")
		return
	}

	// Parse request
	var req struct {
		TrustScore  *float64 `json:"trust_score"`
		TrustTier   string   `json:"trust_tier"`
		BadgeURL    string   `json:"badge_url"`
		Notes       string   `json:"notes"`
	}
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

	// Complete the verification
	if err := h.trustRepo.UpdateVerificationResult(verification.ID, req.TrustScore, req.TrustTier, req.BadgeURL, req.Notes, &claims.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to complete verification")
		h.writeError(w, http.StatusInternalServerError, "Failed to complete verification", "internal_error")
		return
	}

	// Auto-create a verification attestation
	attestation := &trustapi.TrustAttestation{
		FunctionID:        verification.FunctionID,
		FunctionVersion:   verification.FunctionVersion,
		FunctionAuthor:    verification.FunctionAuthor,
		FunctionName:      verification.FunctionName,
		Type:              string(trustapi.AttestationTypeVerification),
		Title:             fmt.Sprintf("Function verified (%s)", verification.VerificationLevel),
		Description:       fmt.Sprintf("Function verified at %s level", verification.VerificationLevel),
		Results:           json.RawMessage(fmt.Sprintf(`{"verification_id":"%s","trust_score":%f,"trust_tier":"%s","verified_by":"%s"}`,
			verification.VerificationID,
			func() float64 { if req.TrustScore != nil { return *req.TrustScore }; return 0 }(),
			req.TrustTier,
			claims.UserID.String(),
		)),
		AttesterID:        claims.UserID,
		AttesterType:      "user",
		AttesterName:      claims.Email,
		VerificationLevel: verification.VerificationLevel,
	}

	if err := h.revocationRepo.CreateAttestation(attestation); err != nil {
		h.logger.WithError(err).Warn("Failed to auto-create verification attestation")
		// Don't fail the request - verification completion succeeded
	} else {
		h.logger.WithFields(logrus.Fields{
			"attestation_id":  attestation.AttestationID,
			"verification_id": verification.VerificationID,
			"function_id":     verification.FunctionID,
		}).Info("Auto-created verification attestation")

		// Trigger webhook for attestation creation
		if h.webhookService != nil {
			webhookData := map[string]interface{}{
				"attestation_id": attestation.AttestationID,
				"function_id":    verification.FunctionID.String(),
				"function_name":  verification.FunctionName,
				"type":           string(trustapi.AttestationTypeVerification),
				"title":          attestation.Title,
				"proof_hash":     attestation.ProofHash,
			}
			go func() { _ = h.webhookService.TriggerEvent(trustapi.WebhookEventAttestationCreated, &verification.FunctionID, webhookData) }()
		}

		// Write audit log
		if h.auditLogRepo != nil {
			if err := h.auditLogRepo.LogAttestationCreated(attestation, claims.UserID, "admin", r.RemoteAddr, r.UserAgent(), ""); err != nil {
				h.logger.WithError(err).Warn("Failed to write audit log for auto-created attestation")
			}
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"verification_id": verification.VerificationID,
		"status":          string(trustapi.VerificationStatusCompleted),
		"trust_score":     req.TrustScore,
		"trust_tier":      req.TrustTier,
		"attestation_id":  attestation.AttestationID,
		"completed_at":    time.Now(),
	})
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

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogPolicyCreated(policy, claims.UserID, r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for policy creation")
		}
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

	// Ownership check: only the policy owner can view it
	claims := middleware.GetUserFromContext(r)
	if claims != nil && policy.OwnerID != claims.UserID {
		h.writeError(w, http.StatusForbidden, "Not authorized to view this policy", "forbidden")
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

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogPolicyUpdated(policy, claims.UserID, "policy updated", r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for policy update")
		}
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

	// Write audit log
	if h.auditLogRepo != nil {
		if err := h.auditLogRepo.LogPolicyDeleted(policy, claims.UserID, r.RemoteAddr, r.UserAgent(), ""); err != nil {
			h.logger.WithError(err).Warn("Failed to write audit log for policy deletion")
		}
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
	fn, err := h.registryRepo.GetFunctionByID(r.Context(), req.FunctionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Function not found", "function_not_found")
		return
	}

	// Get trust score
	trustState, err := h.registryRepo.GetLatestTrustHistory(r.Context(), req.FunctionID)
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
		passed, ruleDecision, ruleReason := h.evaluateRule(r.Context(), rule, trustState, isRevoked, revocationStatus, fn)
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
		fn, err := h.registryRepo.GetFunctionByID(r.Context(), functionID)
		if err != nil {
			errors = append(errors, trustapi.BatchPolicyEvaluationError{
				FunctionID: functionID,
				Error:      "Function not found",
			})
			continue
		}

		// Get trust state
		trustState, _ := h.registryRepo.GetLatestTrustHistory(r.Context(), functionID)
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

		// Full rule evaluation (consistent with single evaluate endpoint)
		rules, err := policy.GetRules()
		if err != nil {
			errors = append(errors, trustapi.BatchPolicyEvaluationError{
				FunctionID: functionID,
				Error:      "Failed to parse policy rules",
			})
			continue
		}

		overallResult := policy.DefaultAction
		if overallResult == "" {
			overallResult = "deny"
		}
		decision := "policy_default"
		reason := "Default policy action applied"

		for _, rule := range rules {
			passed, ruleDecision, ruleReason := h.evaluateRule(r.Context(), rule, trustState, isRevoked, revocationStatus, fn)
			if !passed {
				if ruleDecision == "deny" {
					overallResult = "denied"
					decision = ruleDecision
					reason = ruleReason
					break
				} else if ruleDecision == "warn" && overallResult != "denied" {
					overallResult = "warned"
					decision = ruleDecision
					reason = ruleReason
				}
			} else if overallResult == "deny" || overallResult == "denied" {
				overallResult = "allowed"
				decision = "policy_rule_passed"
				reason = fmt.Sprintf("Rule '%s' passed", rule.ID)
			}
		}

		results = append(results, trustapi.PolicyEvaluateResponse{
			FunctionID:       functionID,
			FunctionAuthor:   fn.Author,
			FunctionName:     fn.Name,
			PolicyID:         policy.PolicyID,
			Result:           overallResult,
			Decision:         decision,
			Reason:           reason,
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
	ctx context.Context,
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
		successRate, err := h.getFunctionSuccessRate(ctx, fn.ID)
		if err != nil {
			h.logger.WithError(err).WithField("function_id", fn.ID).Warn("Failed to get function success rate for rule evaluation")
			// Fail closed: if we can't get metrics, deny the request
			return false, "deny", "Unable to verify success rate metric"
		}
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

// getFunctionSuccessRate fetches the actual success rate for a function from recent metrics
func (h *ExtendedHandler) getFunctionSuccessRate(ctx context.Context, functionID uuid.UUID) (float64, error) {
	totalCalls, successRate, _, _, err := h.registryRepo.GetFunctionStats(functionID, time.Now().Add(-30*24*time.Hour))
	if err != nil {
		return 0, err
	}
	if totalCalls == 0 {
		// No calls recorded — treat as neutral rather than failing
		return 100.0, nil
	}
	return successRate, nil
}

// ============================================
// Merkle Audit Trail Handlers
// ============================================

// HandleGetMerkleTreeHead handles GET /v1/trust/merkle/head
// Returns the latest signed Merkle tree head.
func (h *ExtendedHandler) HandleGetMerkleTreeHead(w http.ResponseWriter, r *http.Request) {
	if h.merkleRepo == nil {
		h.writeError(w, http.StatusServiceUnavailable, "Merkle audit trail not configured", "merkle_unavailable")
		return
	}

	head, err := h.merkleRepo.LatestTreeHead()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get Merkle tree head")
		h.writeError(w, http.StatusInternalServerError, "Failed to get tree head", "internal_error")
		return
	}
	if head == nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"tree_size": 0,
			"root_hash": "",
			"message":   "No attestations in the log yet",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tree_size":     head.TreeSize,
		"root_hash":     head.RootHash,
		"previous_hash": head.PreviousHash,
		"timestamp":     head.Timestamp,
		"signature":     head.Signature,
		"public_key_id": head.PublicKeyID,
		"metadata":      head.Metadata,
	})
}

// HandleGetMerkleInclusionProof handles GET /v1/trust/merkle/inclusion?leaf_index=N
// Returns an inclusion proof for a specific leaf in the tree.
func (h *ExtendedHandler) HandleGetMerkleInclusionProof(w http.ResponseWriter, r *http.Request) {
	if h.merkleRepo == nil {
		h.writeError(w, http.StatusServiceUnavailable, "Merkle audit trail not configured", "merkle_unavailable")
		return
	}

	indexStr := r.URL.Query().Get("leaf_index")
	if indexStr == "" {
		h.writeError(w, http.StatusBadRequest, "leaf_index is required", "missing_leaf_index")
		return
	}

	var leafIndex int64
	if _, err := fmt.Sscanf(indexStr, "%d", &leafIndex); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid leaf_index", "invalid_leaf_index")
		return
	}

	proof, err := h.merkleRepo.GetInclusionProof(leafIndex)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get inclusion proof")
		h.writeError(w, http.StatusInternalServerError, "Failed to get inclusion proof", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, proof)
}

// HandleGetMerkleConsistencyProof handles GET /v1/trust/merkle/consistency?old_size=N
// Returns a consistency proof between an old tree size and the current size.
func (h *ExtendedHandler) HandleGetMerkleConsistencyProof(w http.ResponseWriter, r *http.Request) {
	if h.merkleRepo == nil {
		h.writeError(w, http.StatusServiceUnavailable, "Merkle audit trail not configured", "merkle_unavailable")
		return
	}

	oldSizeStr := r.URL.Query().Get("old_size")
	if oldSizeStr == "" {
		h.writeError(w, http.StatusBadRequest, "old_size is required", "missing_old_size")
		return
	}

	var oldSize int64
	if _, err := fmt.Sscanf(oldSizeStr, "%d", &oldSize); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid old_size", "invalid_old_size")
		return
	}

	proof, err := h.merkleRepo.GetConsistencyProof(oldSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get consistency proof")
		h.writeError(w, http.StatusInternalServerError, "Failed to get consistency proof", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, proof)
}

// HandleVerifyMerkleInclusion handles POST /v1/trust/merkle/verify/inclusion
// Verifies an inclusion proof provided by the client.
func (h *ExtendedHandler) HandleVerifyMerkleInclusion(w http.ResponseWriter, r *http.Request) {
	var req trustapi.MerkleInclusionProof
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	valid := trustapi.VerifyInclusion(req.LeafHash, req.LeafIndex, req.TreeSize, req.Path, req.RootHash)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      valid,
		"leaf_index": req.LeafIndex,
		"tree_size":  req.TreeSize,
		"root_hash":  req.RootHash,
	})
}

// HandleGetMerkleRoot handles GET /v1/trust/merkle/root
// Returns just the current root hash (lightweight check).
func (h *ExtendedHandler) HandleGetMerkleRoot(w http.ResponseWriter, r *http.Request) {
	if h.merkleRepo == nil {
		h.writeError(w, http.StatusServiceUnavailable, "Merkle audit trail not configured", "merkle_unavailable")
		return
	}

	root, err := h.merkleRepo.ComputeRoot()
	if err != nil {
		h.logger.WithError(err).Error("Failed to compute Merkle root")
		h.writeError(w, http.StatusInternalServerError, "Failed to compute root", "internal_error")
		return
	}

	size, _ := h.merkleRepo.TreeSize()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"root_hash": root,
		"tree_size": size,
		"algorithm": "SHA-256",
		"format":    "RFC-6962",
	})
}

// HandleGetMerkleProofForAttestation handles GET /v1/trust/merkle/proof/{attestation_id}
// Looks up the leaf index by attestation ID and returns the inclusion proof.
func (h *ExtendedHandler) HandleGetMerkleProofForAttestation(w http.ResponseWriter, r *http.Request) {
	if h.merkleRepo == nil {
		h.writeError(w, http.StatusServiceUnavailable, "Merkle audit trail not configured", "merkle_unavailable")
		return
	}

	vars := mux.Vars(r)
	attestationID := vars["attestation_id"]

	leafIndex, err := h.merkleRepo.GetLeafIndexByAttestationID(attestationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Attestation not found in Merkle tree", "not_found")
		return
	}

	proof, err := h.merkleRepo.GetInclusionProof(leafIndex)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get Merkle proof")
		h.writeError(w, http.StatusInternalServerError, "Failed to get proof", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, proof)
}

// ============================================
// Chain of Custody Handlers
// ============================================

// HandleGetDelegationChain handles GET /v1/trust/delegation/chain/{chain_id}
// Returns the full chain of custody for a delegation sequence.
func (h *ExtendedHandler) HandleGetDelegationChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := vars["chain_id"]

	attestations, err := h.revocationRepo.GetDelegationChain(chainID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get delegation chain")
		h.writeError(w, http.StatusInternalServerError, "Failed to get delegation chain", "internal_error")
		return
	}

	chain := make([]map[string]interface{}, len(attestations))
	for i, att := range attestations {
		chain[i] = map[string]interface{}{
			"attestation_id":        att.AttestationID,
			"depth":                 att.DelegationDepth,
			"function_id":           att.FunctionID,
			"function_name":         att.FunctionName,
			"function_author":       att.FunctionAuthor,
			"delegator_function_id": att.DelegatorFunctionID,
			"delegator_agent_id":    att.DelegatorAgentID,
			"delegator_trust_score": att.DelegatorTrustScore,
			"delegation_input_hash": att.DelegationInputHash,
			"delegation_output_hash": att.DelegationOutputHash,
			"proof_hash":            att.ProofHash,
			"signature":             att.Signature,
			"parent_attestation_id": att.ParentAttestationID,
			"attested_at":           att.AttestedAt,
			"integrity_verified":    att.VerifyIntegrity(),
		}
	}

	// Verify full chain
	chainValid, chainLength, _ := h.revocationRepo.VerifyDelegationChain(chainID)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"chain_id":      chainID,
		"chain_valid":   chainValid,
		"chain_length":  chainLength,
		"attestations":  chain,
	})
}

// HandleVerifyDelegationChain handles GET /v1/trust/delegation/chain/{chain_id}/verify
// Verifies the cryptographic integrity of a delegation chain.
func (h *ExtendedHandler) HandleVerifyDelegationChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := vars["chain_id"]

	chainValid, chainLength, err := h.revocationRepo.VerifyDelegationChain(chainID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify delegation chain")
		h.writeError(w, http.StatusInternalServerError, "Failed to verify chain", "verification_error")
		return
	}

	signer := trustapi.GetSigner()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"chain_id":       chainID,
		"chain_valid":    chainValid,
		"chain_length":   chainLength,
		"signing_key_id": signer.KeyID(),
		"algorithm":      signer.Algorithm(),
		"verified_at":    time.Now(),
	})
}

// HandleGetFunctionDelegationChains handles GET /v1/trust/delegation/function/{function_id}
// Returns all delegation chains a function participated in.
func (h *ExtendedHandler) HandleGetFunctionDelegationChains(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid function ID", "invalid_function_id")
		return
	}

	chainIDs, err := h.revocationRepo.GetDelegationChainsForFunction(functionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get delegation chains")
		h.writeError(w, http.StatusInternalServerError, "Failed to get delegation chains", "internal_error")
		return
	}

	chains := make([]map[string]interface{}, len(chainIDs))
	for i, chainID := range chainIDs {
		chainValid, chainLength, _ := h.revocationRepo.VerifyDelegationChain(chainID)
		chains[i] = map[string]interface{}{
			"chain_id":     chainID,
			"chain_valid":  chainValid,
			"chain_length": chainLength,
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"function_id": functionID,
		"chains":      chains,
		"total":       len(chains),
	})
}

// HandleGetSignerStatus handles GET /v1/trust/signer/status
// Returns the active signer backend health and key information
func (h *ExtendedHandler) HandleGetSignerStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	isAdmin := claims.Role == "super_admin" || claims.Role == "admin"

	signer := trustapi.GetSigner()

	testPayload := []byte("signer-status-health-check")
	start := time.Now()
	sig, err := signer.Sign(testPayload)
	signLatency := time.Since(start).Milliseconds()

	healthy := false
	var verifyLatency int64 = 0
	if err == nil && sig != "" {
		start = time.Now()
		_, verifyErr := signer.Verify(testPayload, sig)
		verifyLatency = time.Since(start).Milliseconds()
		healthy = verifyErr == nil
	}

	backend := "unknown"
	switch signer.(type) {
	case *trustapi.SoftwareSigner:
		backend = "software"
	case *trustapi.PKCS11Signer:
		backend = "pkcs11"
	case *trustapi.AWSSigner:
		backend = "awskms"
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"healthy":         healthy,
		"backend":         backend,
		"algorithm":       string(signer.Algorithm()),
		"key_id":          signer.KeyID(),
		"public_key_hex":  signer.PublicKeyHex(),
		"sign_latency_ms": signLatency,
		"verify_latency_ms": verifyLatency,
		"is_admin":        isAdmin,
	})
}

// HandleTestSigner handles POST /v1/trust/signer/test
// Signs a test payload, verifies the signature, and returns the result
func (h *ExtendedHandler) HandleTestSigner(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	if claims.Role != "super_admin" && claims.Role != "admin" {
		h.writeError(w, http.StatusForbidden, "Admin access required", "admin_required")
		return
	}

	signer := trustapi.GetSigner()

	testPayload := []byte("signer-test-payload-12345")
	start := time.Now()
	sig, err := signer.Sign(testPayload)
	signLatency := time.Since(start).Milliseconds()

	testPass := false
	var verifyLatency int64 = 0
	if err == nil && sig != "" {
		start = time.Now()
		valid, verifyErr := signer.Verify(testPayload, sig)
		verifyLatency = time.Since(start).Milliseconds()
		testPass = verifyErr == nil && valid
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"pass":             testPass,
		"algorithm":        string(signer.Algorithm()),
		"sign_latency_ms":  signLatency,
		"verify_latency_ms": verifyLatency,
		"key_id":           signer.KeyID(),
		"error":            func() string { if err != nil { return err.Error() }; return "" }(),
	})
}

// HandleListSigningKeys handles GET /v1/trust/signer/keys
// Returns all historical signing keys for audit and rotation tracking
func (h *ExtendedHandler) HandleListSigningKeys(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	keys, err := trustapi.GetAllKeys()
	if err != nil {
		h.logger.WithError(err).Error("Failed to list signing keys")
		h.writeError(w, http.StatusInternalServerError, "Failed to list keys", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"keys":  keys,
		"total": len(keys),
	})
}

// HandleRotateSigningKey handles POST /v1/trust/signer/rotate
// Records the current key as historical and prepares for key rotation.
// Note: actual key generation depends on the backend (software generates new PEM,
// PKCS#11/AWS KMS require external key rotation).
func (h *ExtendedHandler) HandleRotateSigningKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	if claims.Role != "super_admin" && claims.Role != "admin" {
		h.writeError(w, http.StatusForbidden, "Admin access required", "admin_required")
		return
	}

	currentSigner := trustapi.GetSigner()

	// Deactivate current key in history
	if err := trustapi.DeactivateKeyByID(currentSigner.KeyID()); err != nil {
		h.logger.WithError(err).Warn("Failed to deactivate old key in history (may not exist)")
	}

	// For software backend, generate a new key by resetting the singleton
	backend := "unknown"
	switch currentSigner.(type) {
	case *trustapi.SoftwareSigner:
		backend = "software"
	case *trustapi.PKCS11Signer:
		backend = "pkcs11"
	case *trustapi.AWSSigner:
		backend = "awskms"
	}

	if backend == "software" {
		trustapi.ResetSigner()
		newSigner := trustapi.GetSigner()
		if err := trustapi.RecordKey(newSigner, backend); err != nil {
			h.logger.WithError(err).Error("Failed to record new key")
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":       true,
			"backend":       backend,
			"old_key_id":    currentSigner.KeyID(),
			"new_key_id":    newSigner.KeyID(),
			"new_algorithm": string(newSigner.Algorithm()),
			"message":       "Software key rotated successfully",
		})
		return
	}

	// For HSM backends, rotation must happen externally
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"backend":    backend,
		"old_key_id": currentSigner.KeyID(),
		"message":    fmt.Sprintf("Key rotation for %s backend must be performed externally. After rotating, restart the service to pick up the new key.", backend),
	})
}
