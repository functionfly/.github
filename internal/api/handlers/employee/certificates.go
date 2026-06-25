package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListCertificates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeID, err := uuid.Parse(mux.Vars(r)["employeeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListCertificatesOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}

	certs, total, err := h.repo.ListCertificates(r.Context(), employeeID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list certificates")
		apierror.WriteError(w, apierror.NewInternal("Failed to list certificates"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificates": certs,
		"total":        total,
		"limit":        opts.Limit,
		"offset":       opts.Offset,
	})
}

type issueCertificateRequest struct {
	EmployeeID       string `json:"employee_id"`
	CertificateType  string `json:"certificate_type,omitempty"`
	Subject          string `json:"subject"`
	Issuer           string `json:"issuer"`
	PublicKeyPEM     string `json:"public_key_pem"`
	Fingerprint      string `json:"fingerprint"`
	DeviceID         *string `json:"device_id,omitempty"`
	DeviceName       *string `json:"device_name,omitempty"`
	ValidityDays     int    `json:"validity_days,omitempty"`
}

func (h *Handler) HandleIssueCertificate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req issueCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.EmployeeID == "" || req.Subject == "" || req.Issuer == "" || req.PublicKeyPEM == "" || req.Fingerprint == "" {
		apierror.WriteError(w, apierror.NewBadRequest("employee_id, subject, issuer, public_key_pem, and fingerprint are required"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	cert := &storage.EmployeeCertificate{
		EmployeeID:        employeeID,
		TenantID:          claims.TenantID,
		CertificateSerial: uuid.New().String(),
		CertificateType:   "employee",
		Subject:           req.Subject,
		Issuer:            req.Issuer,
		PublicKeyPEM:      req.PublicKeyPEM,
		Fingerprint:       req.Fingerprint,
		DeviceID:          req.DeviceID,
		DeviceName:        req.DeviceName,
		Status:            "active",
	}
	if req.CertificateType != "" {
		cert.CertificateType = req.CertificateType
	}

	created, err := h.repo.IssueCertificate(r.Context(), cert)
	if err != nil {
		h.log.WithError(err).Error("Failed to issue certificate")
		apierror.WriteError(w, apierror.NewInternal("Failed to issue certificate"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificate": created,
	})
}

type revokeCertificateRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) HandleRevokeCertificate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid certificate ID"))
		return
	}

	var req revokeCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Reason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("reason is required"))
		return
	}

	if err := h.repo.RevokeCertificate(r.Context(), id, req.Reason); err != nil {
		h.log.WithError(err).Error("Failed to revoke certificate")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke certificate"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleGetCertificate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	serial := mux.Vars(r)["serial"]
	if serial == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Certificate serial is required"))
		return
	}

	cert, err := h.repo.GetCertificateBySerial(r.Context(), serial)
	if err != nil {
		h.log.WithError(err).Error("Failed to get certificate")
		apierror.WriteError(w, apierror.NewInternal("Failed to get certificate"))
		return
	}
	if cert == nil {
		apierror.WriteError(w, apierror.NewNotFound("Certificate not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificate": cert,
	})
}
