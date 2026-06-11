package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	allowlists, err := h.repo.ListAllowlistsByTenantID(tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list IP allowlists")
		apierror.WriteError(w, apierror.NewInternal("Failed to list IP allowlists"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	var req struct {
		Name                    string `json:"name"`
		Description             string `json:"description"`
		DefaultPolicy           string `json:"default_policy"`
		MFARequiredForUnknownIP bool   `json:"mfa_required_for_unknown_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Name is required"))
		return
	}

	// Validate default_policy
	if req.DefaultPolicy == "" {
		req.DefaultPolicy = "deny"
	}
	if req.DefaultPolicy != "allow" && req.DefaultPolicy != "deny" {
		apierror.WriteError(w, apierror.NewBadRequest("default_policy must be 'allow' or 'deny'"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to create IP allowlist"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist"))
		return
	}
	if allowlist == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist not found"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	// Get existing allowlist
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist"))
		return
	}
	if allowlist == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist not found"))
		return
	}

	var req struct {
		Name                    string `json:"name"`
		Description             string `json:"description"`
		DefaultPolicy           string `json:"default_policy"`
		MFARequiredForUnknownIP *bool  `json:"mfa_required_for_unknown_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
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
			apierror.WriteError(w, apierror.NewBadRequest("default_policy must be 'allow' or 'deny'"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to update IP allowlist"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	err = h.repo.DeleteAllowlist(r.Context(), allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete IP allowlist")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete IP allowlist"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	// Verify the allowlist exists
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist"))
		return
	}
	if allowlist == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist not found"))
		return
	}

	entries, err := h.repo.GetEntriesByAllowlistID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist entries"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	// Verify the allowlist exists
	allowlist, err := h.repo.GetAllowlistByID(allowlistID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist"))
		return
	}
	if allowlist == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist not found"))
		return
	}

	var req struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Type == "" || req.Value == "" {
		apierror.WriteError(w, apierror.NewBadRequest("type and value are required"))
		return
	}

	// Validate type
	if req.Type != "ip" && req.Type != "cidr" {
		apierror.WriteError(w, apierror.NewBadRequest("type must be 'ip' or 'cidr'"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to create IP allowlist entry"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	entryIDStr := vars["entryId"]
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid entry ID"))
		return
	}

	// Get existing entry
	entry, err := h.repo.GetEntryByID(entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist entry"))
		return
	}
	if entry == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
		return
	}

	// Verify entry belongs to the allowlist
	if entry.AllowlistID != allowlistID {
		apierror.WriteError(w, apierror.NewBadRequest("Entry does not belong to this allowlist"))
		return
	}

	var req struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Update fields
	if req.Type != "" {
		if req.Type != "ip" && req.Type != "cidr" {
			apierror.WriteError(w, apierror.NewBadRequest("type must be 'ip' or 'cidr'"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to update IP allowlist entry"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid allowlist ID"))
		return
	}

	entryIDStr := vars["entryId"]
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid entry ID"))
		return
	}

	// Get existing entry
	entry, err := h.repo.GetEntryByID(entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to get IP allowlist entry"))
		return
	}
	if entry == nil {
		apierror.WriteError(w, apierror.NewNotFound("IP allowlist entry not found"))
		return
	}

	// Verify entry belongs to the allowlist
	if entry.AllowlistID != allowlistID {
		apierror.WriteError(w, apierror.NewBadRequest("Entry does not belong to this allowlist"))
		return
	}

	err = h.repo.DeleteEntry(r.Context(), entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete IP allowlist entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete IP allowlist entry"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
