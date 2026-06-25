package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type generateWalletPassRequest struct {
	EmployeeID string `json:"employee_id"`
	PassType   string `json:"pass_type,omitempty"`
	Platform   string `json:"platform"`
}

func (h *Handler) HandleGenerateWalletPass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req generateWalletPassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.EmployeeID == "" || req.Platform == "" {
		apierror.WriteError(w, apierror.NewBadRequest("employee_id and platform are required"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	passType := "employee_badge"
	if req.PassType != "" {
		passType = req.PassType
	}

	wp := &storage.WalletPass{
		EmployeeID:  employeeID,
		TenantID:    claims.TenantID,
		PassType:    passType,
		Platform:    req.Platform,
		PassID:      uuid.New().String(),
		QRToken:     uuid.New().String(),
		QRExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		Status:      "active",
	}

	created, err := h.repo.CreateWalletPass(r.Context(), wp)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate wallet pass")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate wallet pass"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallet_pass": created,
	})
}

func (h *Handler) HandleVerifyBadgeQR(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	qrToken := mux.Vars(r)["token"]
	if qrToken == "" {
		apierror.WriteError(w, apierror.NewBadRequest("QR token is required"))
		return
	}

	wp, err := h.repo.GetWalletPassByQRToken(r.Context(), qrToken)
	if err != nil {
		h.log.WithError(err).Error("Failed to verify QR token")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify QR token"))
		return
	}
	if wp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Invalid or expired QR token"))
		return
	}

	if wp.QRExpiresAt.Before(time.Now()) {
		apierror.WriteError(w, apierror.NewBadRequest("QR token has expired"))
		return
	}

	now := time.Now()
	h.repo.UpdateWalletPass(r.Context(), wp.ID, map[string]interface{}{
		"last_presented_at": now,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":        true,
		"pass":         wp,
		"verified_at":  now,
	})
}

func (h *Handler) HandleRevokeWalletPass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid wallet pass ID"))
		return
	}

	if err := h.repo.UpdateWalletPass(r.Context(), id, map[string]interface{}{
		"status": "revoked",
	}); err != nil {
		h.log.WithError(err).Error("Failed to revoke wallet pass")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke wallet pass"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleListWalletPasses(w http.ResponseWriter, r *http.Request) {
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
	opts := storage.ListWalletPassesOpts{
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

	passes, total, err := h.repo.ListWalletPasses(r.Context(), employeeID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list wallet passes")
		apierror.WriteError(w, apierror.NewInternal("Failed to list wallet passes"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallet_passes": passes,
		"total":         total,
		"limit":         opts.Limit,
		"offset":        opts.Offset,
	})
}
