package vault

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DynamicTokenPrefix is the bearer prefix for CI/CD agent tokens.
// The agent sends `Authorization: Bearer ff_dyn_<token>`.
const DynamicTokenPrefix = "ff_dyn_"

// DynamicTokenSecretBytes is the number of random bytes in the raw
// token. 32 bytes = 256 bits of entropy, base64-url encoded → 43 chars.
const DynamicTokenSecretBytes = 32

// ============================================================================
// Dynamic access tokens (2026-06-16 design) — CI/CD agent auth
// ============================================================================

// HandleCreateDynamicToken mints a new dynamic-credential access
// token. The raw token is returned exactly once in the response.
func (h *Handler) HandleCreateDynamicToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if h.RBAC == nil {
		apierror.WriteError(w, apierror.NewInternal("RBAC not configured"))
		return
	}
	dec, err := h.RBAC.Check(r.Context(), claims.TenantID, claims.UserID,
		"dynamic_credentials:token_mint", "/")
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to check permission"))
		return
	}
	if !dec.Allowed {
		apierror.WriteError(w, apierror.NewForbidden("Permission denied"))
		return
	}
	var req CreateDynamicTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.CredentialID == uuid.Nil {
		apierror.WriteError(w, apierror.NewBadRequest("credential_id is required"))
		return
	}
	if req.ExpiresInHours < 1 || req.ExpiresInHours > 8760 {
		apierror.WriteError(w, apierror.NewBadRequest("expires_in_hours must be between 1 and 8760"))
		return
	}
	cred, err := h.repo.GetCredential(r.Context(), req.CredentialID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load credential"))
		return
	}
	if cred == nil {
		apierror.WriteError(w, apierror.NewNotFound("Credential not found"))
		return
	}

	rawSecret, err := generateDynamicTokenSecret()
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to generate token"))
		return
	}
	raw := DynamicTokenPrefix + rawSecret
	hash := hashDynamicToken(raw)

	tok := &vault.DynamicWrappedAccessToken{
		TenantID:     claims.TenantID,
		CredentialID: req.CredentialID,
		TokenHash:    hash,
		Name:         req.Name,
		Scopes:       vault.JSONMap{"scopes": req.Scopes},
		ExpiresAt:    time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour),
		CreatedBy:    claims.UserID,
	}
	if req.IPPolicy != nil {
		tok.AllowedIPs = vault.StringArray(req.IPPolicy.AllowedIPs)
		tok.DeniedIPs = vault.StringArray(req.IPPolicy.DeniedIPs)
		tok.IPRestrictionEnabled = req.IPPolicy.IPRestrictionEnabled
	}
	if err := h.repo.CreateDynamicToken(r.Context(), tok); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create token"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDYNTokenCreate, true, "", vault.JSONMap{
			"operation":      "dynamic_token_create",
			"token_id":       tok.ID.String(),
			"credential_id":  req.CredentialID.String(),
			"expires_at":     tok.ExpiresAt,
		})
	resp := dynamicTokenResponse(tok)
	resp.Token = raw
	h.respondJSON(w, http.StatusCreated, resp)
}

// HandleListDynamicTokens lists all tokens for the tenant. Hash-only.
func (h *Handler) HandleListDynamicTokens(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	limit := parseLimit(r, 50)
	offset := parseOffset(r)
	tokens, err := h.repo.ListDynamicTokens(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list tokens"))
		return
	}
	resp := make([]DynamicTokenResponse, len(tokens))
	for i := range tokens {
		resp[i] = dynamicTokenResponse(&tokens[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"tokens": resp, "total": len(resp)})
}

// HandleRevokeDynamicToken marks a token revoked.
func (h *Handler) HandleRevokeDynamicToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid token ID"))
		return
	}
	reason := ""
	if r.Body != nil {
		var body RevokeDynamicTokenRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		reason = body.Reason
	}
	if err := h.repo.RevokeDynamicToken(r.Context(), *id, claims.TenantID, reason); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Token not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDYNTokenRevoke, true, "", vault.JSONMap{
			"operation": "dynamic_token_revoke",
			"token_id":  id.String(),
			"reason":    reason,
		})
	w.WriteHeader(http.StatusNoContent)
}

// HandleGetDynamicTokenUsage returns last_used_at + use_count.
func (h *Handler) HandleGetDynamicTokenUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if h.RBAC != nil {
		dec, err := h.RBAC.Check(r.Context(), claims.TenantID, claims.UserID, "audit:read", "/")
		if err == nil && !dec.Allowed {
			apierror.WriteError(w, apierror.NewForbidden("Permission denied"))
			return
		}
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid token ID"))
		return
	}
	tok, err := h.repo.GetDynamicTokenByID(r.Context(), *id, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load token"))
		return
	}
	if tok == nil {
		apierror.WriteError(w, apierror.NewNotFound("Token not found"))
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":           tok.ID,
		"last_used_at": tok.LastUsedAt,
		"use_count":    tok.UseCount,
		"is_revoked":   tok.IsRevoked,
		"expires_at":   tok.ExpiresAt,
	})
}

// --- helpers ---

func generateDynamicTokenSecret() (string, error) {
	b := make([]byte, DynamicTokenSecretBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashDynamicToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func dynamicTokenResponse(t *vault.DynamicWrappedAccessToken) DynamicTokenResponse {
	scopes := []string{}
	if arr, ok := t.Scopes["scopes"].([]interface{}); ok {
		for _, s := range arr {
			if str, ok := s.(string); ok {
				scopes = append(scopes, str)
			}
		}
	}
	return DynamicTokenResponse{
		ID:           t.ID,
		TenantID:     t.TenantID,
		CredentialID: t.CredentialID,
		Name:         t.Name,
		Scopes:       scopes,
		ExpiresAt:    t.ExpiresAt,
		IsRevoked:    t.IsRevoked,
		RevokedAt:    t.RevokedAt,
		RevokedReason: t.RevokedReason,
		LastUsedAt:   t.LastUsedAt,
		UseCount:     t.UseCount,
		CreatedBy:    t.CreatedBy,
		CreatedAt:    t.CreatedAt,
		IPPolicy: IPPolicyDTO{
			AllowedIPs:           []string(t.AllowedIPs),
			DeniedIPs:            []string(t.DeniedIPs),
			IPRestrictionEnabled: t.IPRestrictionEnabled,
		},
	}
}

// parseDynamicTokenBearing extracts an ff_dyn_<token> bearer from the
// Authorization header. Returns the raw token (without prefix) or "".
func parseDynamicTokenBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) < len(DynamicTokenPrefix)+4 {
		return ""
	}
	if !startsWith(auth, "Bearer ") {
		return ""
	}
	token := auth[len("Bearer "):]
	if len(token) < len(DynamicTokenPrefix) {
		return ""
	}
	if !startsWith(token, DynamicTokenPrefix) {
		return ""
	}
	return token
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// parseIntDefault is a small helper used in tests.
func parseIntDefault(s string, dflt int) int {
	if s == "" {
		return dflt
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return dflt
	}
	return n
}
