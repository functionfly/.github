package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// HandleListSignupInvites returns invite metadata (no plaintext codes).
func (h *Handler) HandleListSignupInvites(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListSignupInvitesAdmin(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list signup invites")
		http.Error(w, "Failed to list invites", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": rows})
}

type createSignupInviteRequest struct {
	Label     string `json:"label"`
	MaxUses   *int   `json:"maxUses"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

type createSignupInviteResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	MaxUses   *int      `json:"maxUses,omitempty"`
	ExpiresAt *string   `json:"expiresAt,omitempty"`
}

// HandleCreateSignupInvite mints a new invite; plaintext code is returned once.
func (h *Handler) HandleCreateSignupInvite(w http.ResponseWriter, r *http.Request) {
	var req createSignupInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			http.Error(w, "expiresAt must be RFC3339", http.StatusBadRequest)
			return
		}
		u := t.UTC()
		expiresAt = &u
	}
	var createdBy *uuid.UUID
	if claims := middleware.GetClaimsFromContext(r.Context()); claims != nil {
		createdBy = &claims.UserID
	}
	id, plain, err := h.repo.CreateSignupInvite(r.Context(), req.Label, req.MaxUses, expiresAt, createdBy)
	if err != nil {
		logrus.WithError(err).Error("Failed to create signup invite")
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}
	resp := createSignupInviteResponse{
		ID:      id,
		Code:    plain,
		Label:   req.Label,
		MaxUses: req.MaxUses,
	}
	if expiresAt != nil {
		s := expiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleRevokeSignupInvite soft-revokes an invite.
func (h *Handler) HandleRevokeSignupInvite(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := h.repo.RevokeSignupInvite(r.Context(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Invite not found or already revoked", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to revoke signup invite")
		http.Error(w, "Failed to revoke invite", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invite revoked"})
}
