package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		http.Error(w, "Failed to list waitlist entries", http.StatusInternalServerError)
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
		http.Error(w, "Failed to get waitlist stats", http.StatusInternalServerError)
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
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"pending": true, "approved": true, "rejected": true, "invited": true}
	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateWaitlistEntryStatus(r.Context(), id, req.Status, req.Notes); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			http.Error(w, "Entry not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to update waitlist entry status")
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
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
		http.Error(w, "Invalid id", http.StatusBadRequest)
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
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}

	// Update the waitlist entry with the invite
	if err := h.repo.IssueInviteToWaitlistEntry(r.Context(), entryID, inviteID); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			http.Error(w, "Waitlist entry not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to issue invite to waitlist entry")
		http.Error(w, "Failed to issue invite", http.StatusInternalServerError)
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
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteWaitlistEntry(r.Context(), id); err != nil {
		if err == storage.ErrWaitlistEntryNotFound {
			http.Error(w, "Entry not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to delete waitlist entry")
		http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Entry deleted successfully"})
}