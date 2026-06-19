package vault

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
)

// HandleExportAudit handles GET /v1/vault/audit/export.
//
// Query params:
//
//	from       RFC3339 timestamp (default: 24h ago)
//	to         RFC3339 timestamp (default: now)
//	format     json | csv | cef    (default: json)
//	secret_id  optional UUID filter
//	action     optional action filter
//
// Response headers:
//
//	X-Audit-Signature: base64 HMAC-SHA-256 of the body keyed by the
//	                    audit-export key (see VAULT_AUDIT_SIGNING_KEY)
//	Content-Disposition: attachment; filename=...
//	X-Audit-Generated-At: ISO timestamp
//	X-Audit-Row-Count:    <integer>
func (h *Handler) HandleExportAudit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	q := vault.AuditExportQuery{
		TenantID: claims.TenantID,
		Format:   vault.AuditExportFormat(r.URL.Query().Get("format")),
	}
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("invalid 'from' timestamp"))
			return
		}
		q.From = t
	} else {
		q.From = time.Now().Add(-24 * time.Hour)
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("invalid 'to' timestamp"))
			return
		}
		q.To = t
	} else {
		q.To = time.Now()
	}
	if v := r.URL.Query().Get("action"); v != "" {
		q.Action = v
	}
	if v := r.URL.Query().Get("secret_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("invalid 'secret_id'"))
			return
		}
		q.SecretID = &id
	}

	// Use the configured audit signing key if present; otherwise fall
	// back to a deterministic per-tenant derivation so dev installs
	// still produce verifiable exports.
	signingKey := h.auditSigningKey(claims.TenantID)
	result, err := h.repo.ExportAudit(r.Context(), q, signingKey)
	if err != nil {
		apierror.LogAndInternal(w, r, err, "vault export audit log")
		return
	}
	filename := fmt.Sprintf("vault-audit-%s.%s",
		claims.TenantID.String(), string(result.Format))
	w.Header().Set("Content-Type", contentTypeForExport(string(result.Format)))
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Audit-Signature", result.HMAC)
	w.Header().Set("X-Audit-Generated-At", result.Generated.Format(time.RFC3339Nano))
	w.Header().Set("X-Audit-Row-Count", strconv.Itoa(result.RowCount))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Body)
}

// auditSigningKey returns the per-tenant HMAC key used to sign audit
// exports. If the handler has been configured with an explicit
// AuditKey, that key is used; otherwise a deterministic
// per-tenant derivation is used (suitable for development).
func (h *Handler) auditSigningKey(tenantID uuid.UUID) []byte {
	if h.AuditKey != "" {
		return vault.AuditSigningKey(h.AuditKey, tenantID)
	}
	// For dev/test, derive from a static fallback. Production must
	// set VAULT_AUDIT_SIGNING_KEY explicitly.
	return vault.AuditSigningKey("dev-fallback-audit-signing-key", tenantID)
}

func contentTypeForExport(format string) string {
	switch format {
	case "csv":
		return "text/csv; charset=utf-8"
	case "cef":
		return "text/plain; charset=utf-8"
	default:
		return "application/json"
	}
}
