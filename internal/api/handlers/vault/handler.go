// Package vault provides HTTP handlers for the Secrets Vault API.
//
// The Secrets Vault provides secure storage for encrypted secrets with:
// - Client-side encryption (server never sees plaintext)
// - Scoped access tokens for runtime access
// - Comprehensive audit logging
// - Soft-delete with recovery support
//
// This package is organized into several focused files:
// - handler.go: Main Handler struct and constructor
// - types.go: Request/response types and DTOs
// - secrets.go: Secret CRUD operations
// - tokens.go: Token management and validation
// - audit.go: Audit log querying
//
// Security considerations:
// - All endpoints require authentication
// - Tenant isolation is enforced on all operations
// - Audit logs are created for all mutations
// - Tokens are hashed before storage (plaintext shown once at creation)
package vault

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/functionfly/functionfly/internal/storage/vault/quota"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler handles vault API requests
type Handler struct {
	repo           *vault.Repository
	logger         *logrus.Logger
	DynamicService *vault.DynamicSecretService
	RBAC           *vault.RBACEngine
	SIEM           *vault.SIEMDispatcher
	Cache          *vault.SecretCache
	AuditKey       string
	quotaEnforcer  *quota.Enforcer
}

// NewHandler creates a new vault handler
func NewHandler(repo *vault.Repository, logger *logrus.Logger, quotaEnforcer *quota.Enforcer) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		repo:           repo,
		logger:         logger,
		DynamicService: vault.NewDynamicSecretService(repo),
		RBAC:           vault.NewRBACEngine(repo),
		SIEM:           vault.NewSIEMDispatcher(repo),
		Cache:          vault.NewSecretCache(nil, vault.CacheConfig{}),
		quotaEnforcer:  quotaEnforcer,
	}
}

// respondJSON sends a JSON response with the given status code.
// Cache-Control and Pragma prevent caching of secret metadata and encrypted payloads.
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// respondError sends an error response with appropriate status code
func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(status),
		Code:    code,
		Message: message,
	}
	h.respondJSON(w, status, response)
}

// respondErrorStandard sends a standardized error response using apierror
func (h *Handler) respondErrorStandard(w http.ResponseWriter, err *apierror.APIError) {
	requestID := ""
	if r := getCurrentRequest(); r != nil {
		requestID = r.Header.Get("X-Request-ID")
	}
	if requestID != "" {
		err = err.WithRequestID(requestID)
	}
	apierror.WriteError(w, err)
}

// respondErrorCode sends a standardized error response using error code and message
func (h *Handler) respondErrorCode(w http.ResponseWriter, status int, code apierror.ErrorCode, message string) {
	h.respondErrorStandard(w, &apierror.APIError{
		Status:  status,
		Code:    code,
		Message: message,
	})
}

// getCurrentRequest returns the current request from context (for error handling)
func getCurrentRequest() *http.Request {
	return nil
}

// parseUUID parses a UUID string, returning nil if invalid
func parseUUID(s string) *uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// hashToken creates a SHA-256 hash of a token for storage
// This is a one-way hash - the original token cannot be recovered
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// constantTimeCompare performs a constant-time comparison of two strings
// to prevent timing attacks when comparing token hashes
func constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// scopesToJSONMap converts a string slice to vault.JSONMap for storage
func scopesToJSONMap(scopes []string) vault.JSONMap {
	if scopes == nil {
		return vault.JSONMap{"scopes": []string{}}
	}
	return vault.JSONMap{"scopes": scopes}
}

// jsonMapToScopes extracts scopes from a vault.JSONMap
func jsonMapToScopes(m vault.JSONMap) []string {
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

// secretToResponse converts a vault.Secret to SecretResponse
func secretToResponse(s *vault.Secret) SecretResponse {
	resp := SecretResponse{
		ID:          s.ID,
		TenantID:    s.TenantID,
		Name:        s.Name,
		Description: s.Description,
		SecretType:  s.SecretType,
		EncryptedData: EncryptedDataPayload{
			Ciphertext: base64.StdEncoding.EncodeToString(s.EncryptedValue),
			Salt:       base64.StdEncoding.EncodeToString(s.EncryptionSalt),
			IV:         base64.StdEncoding.EncodeToString(s.IV),
			Tag:        base64.StdEncoding.EncodeToString(s.EncryptionAuthTag),
			KeyVersion: s.KeyVersion,
		},
		Scopes:         jsonMapToScopes(s.Scopes),
		Metadata:       s.Metadata,
		LastAccessedAt: s.LastAccessedAt,
		AccessCount:    s.AccessCount,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}

	// Include version info if available
	if s.CurrentVersion != nil {
		resp.CurrentVersion = *s.CurrentVersion
	}
	if s.LastModifiedAt != nil {
		resp.LastModifiedAt = s.LastModifiedAt
	}

	return resp
}

// secretToMetadataResponse converts a vault.Secret to SecretMetadataResponse (no encrypted data)
func secretToMetadataResponse(s *vault.Secret) SecretMetadataResponse {
	resp := SecretMetadataResponse{
		ID:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		SecretType:     s.SecretType,
		Scopes:         jsonMapToScopes(s.Scopes),
		Metadata:       s.Metadata,
		LastAccessedAt: s.LastAccessedAt,
		AccessCount:    s.AccessCount,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}

	// Include version info if available
	if s.CurrentVersion != nil {
		resp.CurrentVersion = *s.CurrentVersion
	}
	if s.LastModifiedAt != nil {
		resp.LastModifiedAt = s.LastModifiedAt
	}

	return resp
}

// tokenToResponse converts a vault.AccessToken to TokenResponse
func tokenToResponse(t *vault.AccessToken) TokenResponse {
	return TokenResponse{
		ID:            t.ID,
		SecretID:      t.SecretID,
		Name:          t.Name,
		Scopes:        jsonMapToScopes(t.Scopes),
		ExpiresAt:     t.ExpiresAt,
		IsRevoked:     t.IsRevoked,
		RevokedAt:     t.RevokedAt,
		RevokedReason: t.RevokedReason,
		LastUsedAt:    t.LastUsedAt,
		UseCount:      t.UseCount,
		CreatedAt:     t.CreatedAt,
	}
}

// auditLogToResponse converts a vault.AuditLog to AuditLogEntryResponse
func auditLogToResponse(l *vault.AuditLog) AuditLogEntryResponse {
	resp := AuditLogEntryResponse{
		ID:           l.ID,
		Action:       l.Action,
		ActorID:      l.ActorID,
		ActorType:    l.ActorType,
		RequestID:    l.RequestID,
		IPAddress:    l.IPAddress,
		UserAgent:    l.UserAgent,
		Metadata:     l.Metadata,
		Success:      l.Success,
		ErrorMessage: l.ErrorMessage,
		CreatedAt:    l.CreatedAt,
	}
	if l.SecretID != nil {
		resp.SecretID = l.SecretID
	}
	return resp
}

// secretVersionToResponse converts a vault.SecretVersion to SecretVersionResponse
func secretVersionToResponse(v *vault.SecretVersion, includeEncrypted bool) SecretVersionResponse {
	resp := SecretVersionResponse{
		ID:            v.ID,
		SecretID:      v.SecretID,
		VersionNumber: v.VersionNumber,
		Name:          v.Name,
		Description:   v.Description,
		SecretType:    v.SecretType,
		Scopes:        jsonMapToScopes(v.Scopes),
		Metadata:      v.Metadata,
		ChangeType:    v.ChangeType,
		ChangeSummary: v.ChangeSummary,
		ActorID:       v.ActorID,
		ActorType:     v.ActorType,
		CreatedAt:     v.CreatedAt,
	}
	if includeEncrypted {
		resp.EncryptedData = EncryptedDataPayload{
			Ciphertext: base64.StdEncoding.EncodeToString(v.EncryptedValue),
			Salt:       base64.StdEncoding.EncodeToString(v.EncryptionSalt),
			IV:         base64.StdEncoding.EncodeToString(v.IV),
			Tag:        base64.StdEncoding.EncodeToString(v.EncryptionAuthTag),
			KeyVersion: v.KeyVersion,
		}
	}
	return resp
}

// secretVersionToMetadataResponse converts a vault.SecretVersion to SecretVersionMetadataResponse
func secretVersionToMetadataResponse(v *vault.SecretVersion) SecretVersionMetadataResponse {
	return SecretVersionMetadataResponse{
		ID:            v.ID,
		SecretID:      v.SecretID,
		VersionNumber: v.VersionNumber,
		Name:          v.Name,
		Description:   v.Description,
		SecretType:    v.SecretType,
		Scopes:        jsonMapToScopes(v.Scopes),
		Metadata:      v.Metadata,
		ChangeType:    v.ChangeType,
		ChangeSummary: v.ChangeSummary,
		ActorID:       v.ActorID,
		ActorType:     v.ActorType,
		CreatedAt:     v.CreatedAt,
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		for _, ip := range splitAndTrim(xff, ",") {
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// splitAndTrim splits a string by separator and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, p := range splitString(s, sep) {
		parts = append(parts, trimSpace(p))
	}
	return parts
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i < len(s)-len(sep)+1 && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
