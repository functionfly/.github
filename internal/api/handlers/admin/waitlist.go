package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListWaitlist returns all waitlist entries with optional status filter
func (h *Handler) HandleListWaitlist(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "all"
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	entries, total, err := h.repo.ListWaitlistEntries(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list waitlist entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to list waitlist entries"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  entries,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// HandleGetWaitlistStats returns waitlist statistics
func (h *Handler) HandleGetWaitlistStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetWaitlistStats(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get waitlist stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get waitlist stats"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// HandleUpdateWaitlistStatus updates the status of a waitlist entry
func (h *Handler) HandleUpdateWaitlistStatus(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid id"))
		return
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	validStatuses := map[string]bool{"pending": true, "approved": true, "rejected": true, "invited": true}
	if !validStatuses[req.Status] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid status"))
		return
	}

	if err := h.repo.UpdateWaitlistEntryStatus(r.Context(), id, req.Status, req.Notes); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			apierror.WriteError(w, apierror.NewNotFound("Entry not found"))
			return
		}
		logrus.WithError(err).Error("Failed to update waitlist entry status")
		apierror.WriteError(w, apierror.NewInternal("Failed to update status"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

// HandleIssueInviteToWaitlistEntry creates an invite code and assigns it to a waitlist entry
func (h *Handler) HandleIssueInviteToWaitlistEntry(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	entryID, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid id"))
		return
	}

	// Get the waitlist entry to check email
	var req struct {
		Label   string `json:"label,omitempty"`
		MaxUses *int   `json:"maxUses,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional - we can proceed with defaults
		req.Label = ""
	}

	// Create the invite code
	inviteID, plainCode, err := h.repo.CreateSignupInvite(r.Context(), req.Label, req.MaxUses, nil, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create signup invite")
		apierror.WriteError(w, apierror.NewInternal("Failed to create invite"))
		return
	}

	// Update the waitlist entry with the invite
	if err := h.repo.IssueInviteToWaitlistEntry(r.Context(), entryID, inviteID); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			apierror.WriteError(w, apierror.NewNotFound("Waitlist entry not found"))
			return
		}
		logrus.WithError(err).Error("Failed to issue invite to waitlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to issue invite"))
		return
	}

	// Return the plain invite code so admin can share it
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Invite issued successfully",
		"inviteCode": plainCode,
		"inviteId":   inviteID,
	})
}

// HandleDeleteWaitlistEntry removes a waitlist entry
func (h *Handler) HandleDeleteWaitlistEntry(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid id"))
		return
	}

	if err := h.repo.DeleteWaitlistEntry(r.Context(), id); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			apierror.WriteError(w, apierror.NewNotFound("Entry not found"))
			return
		}
		logrus.WithError(err).Error("Failed to delete waitlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete entry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Entry deleted successfully"})
}