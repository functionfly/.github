package registry

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetVerificationStatus gets the verification status for a function version
func (h *Handler) HandleGetVerificationStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionVersionIDStr := vars["functionVersionId"]

	functionVersionID, err := uuid.Parse(functionVersionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid function version ID")
		return
	}

	verificationSvc := verification.NewVerificationService(h.repo, "", "")
	status, err := verificationSvc.GetVerificationStatus(r.Context(), functionVersionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get verification status")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleSignFunction signs a function version
func (h *Handler) HandleSignFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionVersionIDStr := vars["functionVersionId"]

	functionVersionID, err := uuid.Parse(functionVersionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid function version ID")
		return
	}

	var req struct {
		PrivateKeyPEM string `json:"private_key_pem"`
		Algorithm     string `json:"algorithm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if req.PrivateKeyPEM == "" || req.Algorithm == "" {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "private_key_pem and algorithm are required")
		return
	}

	verificationSvc := verification.NewVerificationService(h.repo, "", "")
	signature, err := verificationSvc.SignFunction(functionVersionID, user.Email, req.PrivateKeyPEM, req.Algorithm)
	if err != nil {
		logrus.WithError(err).Error("Failed to sign function")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to sign function")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signature)
}

// HandleVerifySignature verifies a function signature
func (h *Handler) HandleVerifySignature(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	signatureIDStr := vars["signatureId"]

	signatureID, err := uuid.Parse(signatureIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid signature ID")
		return
	}

	var req struct {
		PublicKeyPEM string `json:"public_key_pem"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if req.PublicKeyPEM == "" {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "public_key_pem is required")
		return
	}

	verificationSvc := verification.NewVerificationService(h.repo, "", "")
	err = verificationSvc.VerifySignature(signatureID, req.PublicKeyPEM)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify signature")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to verify signature")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "verified"})
}

// HandleRequestApproval requests approval for a function version
func (h *Handler) HandleRequestApproval(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionVersionIDStr := vars["functionVersionId"]

	functionVersionID, err := uuid.Parse(functionVersionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid function version ID")
		return
	}

	var req struct {
		ApprovalType string `json:"approval_type"`
		TrustLevel   string `json:"trust_level"`
		Priority     string `json:"priority,omitempty"`
		Comments     string `json:"comments,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if req.ApprovalType == "" || req.TrustLevel == "" {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "approval_type and trust_level are required")
		return
	}

	approvalReq := verification.ApprovalRequest{
		FunctionVersionID: functionVersionID,
		ApprovalType:      req.ApprovalType,
		TrustLevel:        req.TrustLevel,
		Priority:          req.Priority,
		RequestedBy:       user.UserID,
		Comments:          req.Comments,
	}

	verificationSvc := verification.NewVerificationService(h.repo, "", "")
	approval, err := verificationSvc.RequestApproval(approvalReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to request approval")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to request approval")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(approval)
}

// HandleMakeApprovalDecision processes an approval decision
func (h *Handler) HandleMakeApprovalDecision(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	approvalIDStr := vars["approvalId"]

	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid approval ID")
		return
	}

	var req struct {
		Decision        string                        `json:"decision"`
		Comments        string                        `json:"comments,omitempty"`
		RequiredActions []verification.ApprovalAction `json:"required_actions,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if req.Decision == "" {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "decision is required")
		return
	}

	decision := verification.ApprovalDecision{
		ApprovalID:      approvalID,
		Decision:        req.Decision,
		ReviewerID:      user.UserID,
		Comments:        req.Comments,
		RequiredActions: req.RequiredActions,
	}

	verificationSvc := verification.NewVerificationService(h.repo, "", "")
	err = verificationSvc.MakeApprovalDecision(decision)
	if err != nil {
		logrus.WithError(err).Error("Failed to process approval decision")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to process approval decision")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleGetApprovals gets approvals for a function version
func (h *Handler) HandleGetApprovals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionVersionIDStr := vars["functionVersionId"]

	functionVersionID, err := uuid.Parse(functionVersionIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid function version ID")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	approvals, err := h.repo.GetFunctionApprovals(functionVersionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get approvals")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"approvals": approvals,
		"limit":     limit,
		"offset":    offset,
	})
}

// HandleGetPendingApprovals gets pending approvals for review
func (h *Handler) HandleGetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	trustLevel := r.URL.Query().Get("trust_level")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var approvals []storage.RegistryFunctionApproval
	var err error

	if trustLevel != "" {
		// Get pending approvals filtered by trust level
		approvals, err = h.repo.GetApprovalsByTrustLevel(trustLevel, limit, offset)
	} else {
		// Get all pending approvals
		approvals, err = h.repo.GetPendingApprovals(limit, offset)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to get pending approvals")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"approvals": approvals,
		"limit":     limit,
		"offset":    offset,
	})
}

// HandleAddApprovalComment adds a comment to an approval
func (h *Handler) HandleAddApprovalComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	approvalIDStr := vars["approvalId"]

	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid approval ID")
		return
	}

	var req struct {
		Comment    string `json:"comment"`
		IsInternal bool   `json:"is_internal,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if req.Comment == "" {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "comment is required")
		return
	}

	approvalSvc := verification.NewApprovalService(h.repo)
	err = approvalSvc.AddComment(approvalID, user.UserID, req.Comment, req.IsInternal)
	if err != nil {
		logrus.WithError(err).Error("Failed to add approval comment")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Failed to add comment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleGetApprovalComments gets comments for an approval
func (h *Handler) HandleGetApprovalComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalIDStr := vars["approvalId"]

	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, functionregistry.ErrCodeInvalidInput, "Invalid approval ID")
		return
	}

	comments, err := h.repo.GetApprovalComments(approvalID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get approval comments")
		h.writeError(w, http.StatusInternalServerError, functionregistry.ErrCodeInternalError, "Internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comments": comments,
	})
}
