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

func (h *Handler) HandleProvisionEmail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req struct {
		EmployeeID  string  `json:"employee_id"`
		Email       string  `json:"email"`
		DisplayName *string `json:"display_name,omitempty"`
		Provider    string  `json:"provider,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.EmployeeID == "" || req.Email == "" {
		apierror.WriteError(w, apierror.NewBadRequest("employee_id and email are required"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	provider := "spacemail"
	if req.Provider != "" {
		provider = req.Provider
	}

	now := time.Now()
	ea := &storage.EmailAccount{
		EmployeeID:    employeeID,
		TenantID:      claims.TenantID,
		Email:         req.Email,
		DisplayName:   req.DisplayName,
		Provider:      provider,
		Status:        "active",
		ProvisionedAt: &now,
	}

	created, err := h.repo.CreateEmailAccount(r.Context(), ea)
	if err != nil {
		h.log.WithError(err).Error("Failed to provision email")
		apierror.WriteError(w, apierror.NewInternal("Failed to provision email"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email_account": created,
	})
}

func (h *Handler) HandleListEmails(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListEmailAccountsOpts{
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
	if eid := q.Get("employee_id"); eid != "" {
		if id, err := uuid.Parse(eid); err == nil {
			opts.EmployeeID = &id
		}
	}

	accounts, total, err := h.repo.ListEmailAccounts(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list email accounts")
		apierror.WriteError(w, apierror.NewInternal("Failed to list email accounts"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email_accounts": accounts,
		"total":          total,
		"limit":          opts.Limit,
		"offset":         opts.Offset,
	})
}

func (h *Handler) HandleUpdateEmailStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid email account ID"))
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Status == "" {
		apierror.WriteError(w, apierror.NewBadRequest("status is required"))
		return
	}

	if err := h.repo.UpdateEmailAccount(r.Context(), id, map[string]interface{}{"status": req.Status}); err != nil {
		h.log.WithError(err).Error("Failed to update email account")
		apierror.WriteError(w, apierror.NewInternal("Failed to update email account"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
