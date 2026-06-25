package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type generateCertificateKeyRequest struct {
	CertificateID string `json:"certificate_id"`
	PublicKeyPEM  string `json:"public_key_pem"`
	PrivateKeyPEM string `json:"private_key_pem,omitempty"`
	KeyType       string `json:"key_type,omitempty"`
	KeySize       int    `json:"key_size,omitempty"`
}

func (h *Handler) HandleGenerateCertificateKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req generateCertificateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.CertificateID == "" || req.PublicKeyPEM == "" {
		apierror.WriteError(w, apierror.NewBadRequest("certificate_id and public_key_pem are required"))
		return
	}

	certificateID, err := uuid.Parse(req.CertificateID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid certificate_id"))
		return
	}

	keyType := "RSA"
	if req.KeyType != "" {
		keyType = req.KeyType
	}
	keySize := 2048
	if req.KeySize > 0 {
		keySize = req.KeySize
	}

	ck := &storage.CertificateKey{
		CertificateID: certificateID,
		PublicKeyPEM:  req.PublicKeyPEM,
		KeyType:       keyType,
		KeySize:       keySize,
	}
	if req.PrivateKeyPEM != "" {
		ck.PrivateKeyPEM = &req.PrivateKeyPEM
	}

	created, err := h.repo.CreateCertificateKey(r.Context(), ck)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate certificate key")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate certificate key"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificate_key": created,
	})
}

func (h *Handler) HandleGetCertificateChain(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	certID, err := uuid.Parse(mux.Vars(r)["certificateId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid certificate ID"))
		return
	}

	keys, err := h.repo.GetCertificateKeysByCertID(r.Context(), certID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get certificate chain")
		apierror.WriteError(w, apierror.NewInternal("Failed to get certificate chain"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys": keys,
		"total": len(keys),
	})
}
