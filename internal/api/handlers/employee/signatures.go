package employee

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type requestSignatureRequest struct {
	DocumentID  string  `json:"document_id"`
	SignerID    string  `json:"signer_id"`
	SignerName  string  `json:"signer_name"`
	SignerEmail *string `json:"signer_email,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

func (h *Handler) HandleRequestSignature(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req requestSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.DocumentID == "" || req.SignerID == "" || req.SignerName == "" {
		apierror.WriteError(w, apierror.NewBadRequest("document_id, signer_id, and signer_name are required"))
		return
	}

	documentID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid document_id"))
		return
	}
	signerID, err := uuid.Parse(req.SignerID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid signer_id"))
		return
	}

	ds := &storage.DocumentSignature{
		DocumentID:  documentID,
		SignerID:    signerID,
		SignerName:  req.SignerName,
		SignerEmail: req.SignerEmail,
		Status:      "pending",
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			ds.ExpiresAt = &t
		}
	}

	created, err := h.repo.CreateDocumentSignature(r.Context(), ds)
	if err != nil {
		h.log.WithError(err).Error("Failed to request signature")
		apierror.WriteError(w, apierror.NewInternal("Failed to request signature"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"signature": created,
	})
}

type signDocumentRequest struct {
	SignatureData string `json:"signature_data"`
}

func (h *Handler) HandleSignDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid signature ID"))
		return
	}

	var req signDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.SignatureData == "" {
		apierror.WriteError(w, apierror.NewBadRequest("signature_data is required"))
		return
	}

	now := time.Now()
	if err := h.repo.UpdateDocumentSignature(r.Context(), id, map[string]interface{}{
		"status":         "signed",
		"signature_data": req.SignatureData,
		"signed_at":      &now,
	}); err != nil {
		h.log.WithError(err).Error("Failed to sign document")
		apierror.WriteError(w, apierror.NewInternal("Failed to sign document"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type declineSignatureRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) HandleDeclineSignature(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid signature ID"))
		return
	}

	var req declineSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Reason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("reason is required"))
		return
	}

	if err := h.repo.UpdateDocumentSignature(r.Context(), id, map[string]interface{}{
		"status":         "declined",
		"decline_reason": req.Reason,
	}); err != nil {
		h.log.WithError(err).Error("Failed to decline signature")
		apierror.WriteError(w, apierror.NewInternal("Failed to decline signature"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleGetSignatureStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid signature ID"))
		return
	}

	sig, err := h.repo.GetDocumentSignatureByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get signature status")
		apierror.WriteError(w, apierror.NewInternal("Failed to get signature status"))
		return
	}
	if sig == nil {
		apierror.WriteError(w, apierror.NewNotFound("Signature not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"signature": sig,
	})
}
