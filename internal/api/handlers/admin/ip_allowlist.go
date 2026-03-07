package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// IPAllowlistHandler contains IP allowlist handlers
type IPAllowlistHandler struct {
	repo *storage.IPAllowlistRepository
}

// NewIPAllowlistHandler creates a new IP allowlist handler
func NewIPAllowlistHandler(repo *storage.IPAllowlistRepository) *IPAllowlistHandler {
	return &IPAllowlistHandler{
		repo: repo,
	}
}

// HandleListTenantIPAllowlists lists all IP allowlists for a tenant
// GET /tenants/{tenantId}/ip-allowlists
func (h *IPAllowlistHandler) HandleListTenantIPAllowlists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	allowlists, err := h.repo.ListAllowlistsByTenantID(tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list IP allowlists")
		http.Error(w, "Failed to list IP allowlists", http.StatusInternalServerError)
		return
	}

	if allowlists == nil {
		allowlists = []*storage.IPAllowlist{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ip_allowlists": allowlists,
		"tenant_id":     tenantID,
	})
}

// HandleCreateIPAllowlist creates a new IP allowlist for a tenant
// POST /tenants/{tenantId}/ip-allowlists
func (h *IPAllowlistHandler) HandleCreateIPAllowlist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name                    string `json:"name"`
		Description             string `json:"description"`
		DefaultPolicy           string `json:"default_policy"`
		MFARequiredForUnknownIP bool   `json:"mfa_required_for_unknown_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Validate default_policy
	if req.DefaultPolicy == "" {
		req.DefaultPolicy = "deny"
	}
	if req.DefaultPolicy != "allow" && req.DefaultPolicy != "deny" {
		http.Error(w, "default_policy must be 'allow' or 'deny'", http.StatusBadRequest)
		return
	}

	allowlist := &storage.IPAllowlist{
		ID:                      uuid.New(),
		TenantID:                tenantID,
		Name:                    req.Name,
		DefaultPolicy:           req.DefaultPolicy,
		MFARequiredForUnknownIP: req.MFARequiredForUnknownIP,
	}

	if req.Description != "" {
		allowlist.Description = &req.Description
	}

	err = h.repo.CreateAllowlist(r.Context(), allowlist)
	if err != nil {
		logrus.WithError(err).Error("Failed to create IP allowlist")
		http.Error(w, "Failed to create IP allowlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(allowlist)
}

// HandleGetIPAllowlist gets a specific IP allowlist
// GET /ip-allowlists/{id}
func (h *IPAllowlistHandler) HandleGetIPAllowlist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		http.Error(w, "Failed to get IP allowlist", http.StatusInternalServerError)
		return
	}
	if allowlist == nil {
		http.Error(w, "IP allowlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allowlist)
}

// HandleUpdateIPAllowlist updates an IP allowlist
// PUT /ip-allowlists/{id}
func (h *IPAllowlistHandler) HandleUpdateIPAllowlist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	// Get existing allowlist
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		http.Error(w, "Failed to get IP allowlist", http.StatusInternalServerError)
		return
	}
	if allowlist == nil {
		http.Error(w, "IP allowlist not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name                    string `json:"name"`
		Description             string `json:"description"`
		DefaultPolicy           string `json:"default_policy"`
		MFARequiredForUnknownIP *bool  `json:"mfa_required_for_unknown_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		allowlist.Name = req.Name
	}
	if req.Description != "" {
		allowlist.Description = &req.Description
	}
	if req.DefaultPolicy != "" {
		if req.DefaultPolicy != "allow" && req.DefaultPolicy != "deny" {
			http.Error(w, "default_policy must be 'allow' or 'deny'", http.StatusBadRequest)
			return
		}
		allowlist.DefaultPolicy = req.DefaultPolicy
	}
	if req.MFARequiredForUnknownIP != nil {
		allowlist.MFARequiredForUnknownIP = *req.MFARequiredForUnknownIP
	}

	err = h.repo.UpdateAllowlist(r.Context(), allowlist)
	if err != nil {
		logrus.WithError(err).Error("Failed to update IP allowlist")
		http.Error(w, "Failed to update IP allowlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allowlist)
}

// HandleDeleteIPAllowlist deletes an IP allowlist
// DELETE /ip-allowlists/{id}
func (h *IPAllowlistHandler) HandleDeleteIPAllowlist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteAllowlist(r.Context(), allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete IP allowlist")
		http.Error(w, "Failed to delete IP allowlist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListIPAllowlistEntries lists all entries for an IP allowlist
// GET /ip-allowlists/{id}/entries
func (h *IPAllowlistHandler) HandleListIPAllowlistEntries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	// Verify the allowlist exists
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		http.Error(w, "Failed to get IP allowlist", http.StatusInternalServerError)
		return
	}
	if allowlist == nil {
		http.Error(w, "IP allowlist not found", http.StatusNotFound)
		return
	}

	entries, err := h.repo.GetEntriesByAllowlistID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entries")
		http.Error(w, "Failed to get IP allowlist entries", http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []*storage.IPAllowlistEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":      entries,
		"allowlist_id": allowlistID,
	})
}

// HandleCreateIPAllowlistEntry creates a new entry in an IP allowlist
// POST /ip-allowlists/{id}/entries
func (h *IPAllowlistHandler) HandleCreateIPAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	// Verify the allowlist exists
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		http.Error(w, "Failed to get IP allowlist", http.StatusInternalServerError)
		return
	}
	if allowlist == nil {
		http.Error(w, "IP allowlist not found", http.StatusNotFound)
		return
	}

	var req struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.Value == "" {
		http.Error(w, "type and value are required", http.StatusBadRequest)
		return
	}

	// Validate type
	if req.Type != "ip" && req.Type != "cidr" {
		http.Error(w, "type must be 'ip' or 'cidr'", http.StatusBadRequest)
		return
	}

	entry := &storage.IPAllowlistEntry{
		ID:          uuid.New(),
		AllowlistID: allowlistID,
		Type:        req.Type,
		Value:       req.Value,
	}

	if req.Description != "" {
		entry.Description = &req.Description
	}

	err = h.repo.CreateEntry(r.Context(), entry)
	if err != nil {
		logrus.WithError(err).Error("Failed to create IP allowlist entry")
		http.Error(w, "Failed to create IP allowlist entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// HandleUpdateIPAllowlistEntry updates an IP allowlist entry
// PUT /ip-allowlists/{id}/entries/{entryId}
func (h *IPAllowlistHandler) HandleUpdateIPAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	entryIDStr := vars["entryId"]
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	// Get existing entry
	entry, err := h.repo.GetEntryByID(entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entry")
		http.Error(w, "Failed to get IP allowlist entry", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "IP allowlist entry not found", http.StatusNotFound)
		return
	}

	// Verify entry belongs to the allowlist
	if entry.AllowlistID != allowlistID {
		http.Error(w, "Entry does not belong to this allowlist", http.StatusBadRequest)
		return
	}

	var req struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Type != "" {
		if req.Type != "ip" && req.Type != "cidr" {
			http.Error(w, "type must be 'ip' or 'cidr'", http.StatusBadRequest)
			return
		}
		entry.Type = req.Type
	}
	if req.Value != "" {
		entry.Value = req.Value
	}
	if req.Description != "" {
		entry.Description = &req.Description
	}

	err = h.repo.UpdateEntry(r.Context(), entry)
	if err != nil {
		logrus.WithError(err).Error("Failed to update IP allowlist entry")
		http.Error(w, "Failed to update IP allowlist entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleDeleteIPAllowlistEntry deletes an IP allowlist entry
// DELETE /ip-allowlists/{id}/entries/{entryId}
func (h *IPAllowlistHandler) HandleDeleteIPAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allowlistIDStr := vars["id"]
	allowlistID, err := uuid.Parse(allowlistIDStr)
	if err != nil {
		http.Error(w, "Invalid allowlist ID", http.StatusBadRequest)
		return
	}

	entryIDStr := vars["entryId"]
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	// Get existing entry
	entry, err := h.repo.GetEntryByID(entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entry")
		http.Error(w, "Failed to get IP allowlist entry", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "IP allowlist entry not found", http.StatusNotFound)
		return
	}

	// Verify entry belongs to the allowlist
	if entry.AllowlistID != allowlistID {
		http.Error(w, "Entry does not belong to this allowlist", http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteEntry(r.Context(), entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete IP allowlist entry")
		http.Error(w, "Failed to delete IP allowlist entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
