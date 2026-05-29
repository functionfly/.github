package admin

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// VaultHandler handles admin vault API requests (cross-tenant secret management)
type VaultHandler struct {
	vaultRepo *vault.Repository
	logger    *logrus.Logger
}

// NewVaultHandler creates a new admin vault handler
func NewVaultHandler(vaultRepo *vault.Repository, logger *logrus.Logger) *VaultHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &VaultHandler{
		vaultRepo: vaultRepo,
		logger:    logger,
	}
}

// respondJSON sends a JSON response
func (h *VaultHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// respondError sends an error response
func (h *VaultHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	response := map[string]string{
		"error":   http.StatusText(status),
		"code":    code,
		"message": message,
	}
	h.respondJSON(w, status, response)
}

// parseUUID parses a UUID string, returning nil if invalid
func parseAdminVaultUUID(s string) *uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// HandleListSecrets handles GET /v1/admin/vault/secrets
// Lists all secrets across all tenants (admin view)
func (h *VaultHandler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Optional tenant filter
	tenantFilter := r.URL.Query().Get("tenant_id")
	var tenantID *uuid.UUID
	if tenantFilter != "" {
		tid := parseAdminVaultUUID(tenantFilter)
		if tid == nil {
			h.respondError(w, http.StatusBadRequest, "INVALID_TENANT", "Invalid tenant_id format")
			return
		}
		tenantID = tid
	}

	var secrets []vault.Secret
	var total int64
	var err error

	if tenantID != nil {
		secrets, err = h.vaultRepo.GetSecretsByTenantPaginated(r.Context(), *tenantID, limit, offset)
		if err != nil {
			h.logger.WithError(err).Error("Failed to list secrets by tenant")
			h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list secrets")
			return
		}
		total, err = h.vaultRepo.CountSecretsByTenant(r.Context(), *tenantID)
	} else {
		secrets, err = h.vaultRepo.GetAllSecretsPaginated(r.Context(), limit, offset)
		if err != nil {
			h.logger.WithError(err).Error("Failed to list all secrets")
			h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list secrets")
			return
		}
		total, err = h.vaultRepo.CountAllSecrets(r.Context())
	}

	// Build response
	type SecretMetadata struct {
		ID             uuid.UUID      `json:"id"`
		TenantID       uuid.UUID      `json:"tenant_id"`
		Name           string         `json:"name"`
		Description    string         `json:"description,omitempty"`
		SecretType     string         `json:"secret_type"`
		Scopes         []string       `json:"scopes,omitempty"`
		Metadata       map[string]interface{} `json:"metadata,omitempty"`
		LastAccessedAt *interface{}   `json:"last_accessed_at,omitempty"`
		AccessCount    int            `json:"access_count"`
		CreatedAt      interface{}    `json:"created_at"`
		UpdatedAt      interface{}    `json:"updated_at"`
		CurrentVersion *int           `json:"current_version,omitempty"`
		LastModifiedAt *interface{}   `json:"last_modified_at,omitempty"`
	}

	items := make([]SecretMetadata, len(secrets))
	for i, s := range secrets {
		items[i] = SecretMetadata{
			ID:          s.ID,
			TenantID:    s.TenantID,
			Name:        s.Name,
			Description: s.Description,
			SecretType:  s.SecretType.String(),
			Scopes:      parseScopes(s.Scopes),
			Metadata:    parseMetadata(s.Metadata),
			AccessCount: s.AccessCount,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		}
		if s.CurrentVersion != nil {
			items[i].CurrentVersion = s.CurrentVersion
		}
		if s.LastAccessedAt != nil {
			items[i].LastAccessedAt = toInterfacePtr(s.LastAccessedAt)
		}
		if s.LastModifiedAt != nil {
			items[i].LastModifiedAt = toInterfacePtr(s.LastModifiedAt)
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"secrets": items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetSecret handles GET /v1/admin/vault/secrets/{secretId}
// Gets a single secret by ID (cross-tenant, admin sees all)
func (h *VaultHandler) HandleGetSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	secretID := parseAdminVaultUUID(vars["secretId"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	secret, err := h.vaultRepo.GetSecretByIDAdmin(r.Context(), *secretID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	type SecretResponse struct {
		ID             uuid.UUID      `json:"id"`
		TenantID       uuid.UUID      `json:"tenant_id"`
		UserID         uuid.UUID      `json:"user_id"`
		Name           string         `json:"name"`
		Description    string         `json:"description,omitempty"`
		SecretType     string         `json:"secret_type"`
		EncryptedData  EncryptedDataPayload `json:"encrypted_data"`
		Scopes         []string       `json:"scopes,omitempty"`
		Metadata       map[string]interface{} `json:"metadata,omitempty"`
		LastAccessedAt *interface{}   `json:"last_accessed_at,omitempty"`
		AccessCount    int            `json:"access_count"`
		CreatedAt      interface{}    `json:"created_at"`
		UpdatedAt      interface{}    `json:"updated_at"`
		CurrentVersion *int           `json:"current_version,omitempty"`
		LastModifiedAt *interface{}   `json:"last_modified_at,omitempty"`
	}

	// Note: We return the encrypted data since this is admin view
	// Client-side encryption means admins don't see plaintext
	resp := SecretResponse{
		ID:          secret.ID,
		TenantID:    secret.TenantID,
		UserID:      secret.UserID,
		Name:        secret.Name,
		Description: secret.Description,
		SecretType:  secret.SecretType.String(),
		EncryptedData: EncryptedDataPayload{
			Ciphertext: base64Encode(secret.EncryptedValue),
			Salt:       base64Encode(secret.EncryptionSalt),
			IV:         base64Encode(secret.IV),
			Tag:        base64Encode(secret.EncryptionAuthTag),
			KeyVersion: secret.KeyVersion,
		},
		Scopes:      parseScopes(secret.Scopes),
		Metadata:    parseMetadata(secret.Metadata),
		AccessCount: secret.AccessCount,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
	}
	if secret.CurrentVersion != nil {
		resp.CurrentVersion = secret.CurrentVersion
	}
	if secret.LastAccessedAt != nil {
		resp.LastAccessedAt = toInterfacePtr(secret.LastAccessedAt)
	}
	if secret.LastModifiedAt != nil {
		resp.LastModifiedAt = toInterfacePtr(secret.LastModifiedAt)
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleGetSecretAuditLogs handles GET /v1/admin/vault/secrets/{secretId}/audit
// Gets audit logs for a specific secret
func (h *VaultHandler) HandleGetSecretAuditLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	secretID := parseAdminVaultUUID(vars["secretId"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	// Verify secret exists
	secret, err := h.vaultRepo.GetSecretByIDAdmin(r.Context(), *secretID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	logs, err := h.vaultRepo.GetAuditLogsBySecret(r.Context(), *secretID, secret.TenantID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get audit logs")
		return
	}

	type AuditLogEntry struct {
		ID           uuid.UUID  `json:"id"`
		SecretID     *uuid.UUID `json:"secret_id,omitempty"`
		Action       string     `json:"action"`
		ActorID      string     `json:"actor_id"`
		ActorType    string     `json:"actor_type"`
		RequestID    string     `json:"request_id,omitempty"`
		IPAddress    string     `json:"ip_address,omitempty"`
		UserAgent    string     `json:"user_agent,omitempty"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
		Success      bool       `json:"success"`
		ErrorMessage string     `json:"error_message,omitempty"`
		CreatedAt    interface{} `json:"created_at"`
	}

	entries := make([]AuditLogEntry, len(logs))
	for i, l := range logs {
		entries[i] = AuditLogEntry{
			ID:           l.ID,
			SecretID:     l.SecretID,
			Action:       l.Action.String(),
			ActorID:      l.ActorID,
			ActorType:    l.ActorType.String(),
			RequestID:    l.RequestID,
			IPAddress:    l.IPAddress,
			UserAgent:    l.UserAgent,
			Metadata:     parseMetadata(l.Metadata),
			Success:      l.Success,
			ErrorMessage: l.ErrorMessage,
			CreatedAt:    l.CreatedAt,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"total":   len(entries),
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetVaultStats handles GET /v1/admin/vault/stats
// Returns aggregate vault statistics across all tenants
func (h *VaultHandler) HandleGetVaultStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.vaultRepo.GetVaultStats(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get vault stats")
		h.respondError(w, http.StatusInternalServerError, "STATS_FAILED", "Failed to get vault statistics")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// HandleListTenantsWithSecrets handles GET /v1/admin/vault/tenants
// Lists all tenants that have at least one secret in the vault
func (h *VaultHandler) HandleListTenantsWithSecrets(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.vaultRepo.GetTenantsWithSecrets(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tenants with secrets")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list tenants")
		return
	}

	type TenantInfo struct {
		TenantID       uuid.UUID `json:"tenant_id"`
		SecretCount    int64     `json:"secret_count"`
		OldestSecretAt interface{} `json:"oldest_secret_at,omitempty"`
		NewestSecretAt interface{} `json:"newest_secret_at,omitempty"`
	}

	items := make([]TenantInfo, len(tenants))
	for i, t := range tenants {
		items[i] = TenantInfo{
			TenantID:       t.TenantID,
			SecretCount:    t.SecretCount,
			OldestSecretAt: t.OldestSecretAt,
			NewestSecretAt: t.NewestSecretAt,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"tenants": items,
		"total":   len(items),
	})
}

// EncryptedDataPayload for admin vault responses
type EncryptedDataPayload struct {
	Ciphertext string `json:"ciphertext"`
	Salt       string `json:"salt"`
	IV         string `json:"iv"`
	Tag        string `json:"tag"`
	KeyVersion int    `json:"key_version"`
}

// Helper functions

func base64Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func parseScopes(m vault.JSONMap) []string {
	if m == nil {
		return []string{}
	}
	if scopes, ok := m["scopes"].([]interface{}); ok {
		result := make([]string, len(scopes))
		for i, s := range scopes {
			if str, ok := s.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	if scopes, ok := m["scopes"].([]string); ok {
		return scopes
	}
	return []string{}
}

func parseMetadata(m vault.JSONMap) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func toInterfacePtr(t interface{}) *interface{} {
	v := t
	return &v
}

// TenantVaultStats represents stats for a tenant with secrets
type TenantVaultStats struct {
	TenantID       uuid.UUID
	SecretCount    int64
	OldestSecretAt interface{}
	NewestSecretAt interface{}
}
