package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VaultMFAContext is the key under which the MFA-verified state is
// stored in the request context. Handlers can read this to detect
// whether the request is operating under an MFA session.
type vaultMFAContextKey struct{}

// VaultMFABypass indicates the request is operating under an active
// break-glass grant; the audit log must include the grant ID.
type VaultMFABypass struct {
	BreakGlassID uuid.UUID
}

// RequireVaultMFA returns a middleware that enforces the tenant's
// vault MFA policy. If the tenant has VaultMFAConfig.EnforceForAPI
// the middleware checks the audit log for a recent
// AuditActionMFAVerify row for the (user, tenant) pair within
// MFASessionTTL. Active break-glass grants bypass the check.
//
// The middleware is intentionally lenient on missing MFA config: if
// the tenant has no VaultMFAConfig row the request passes through.
func RequireVaultMFA(repo *vault.Repository, logger *logrus.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}
			if repo == nil {
				next.ServeHTTP(w, r)
				return
			}
			cfg, err := repo.GetMFAConfig(r.Context(), claims.TenantID)
			if err != nil {
				// Best-effort: do not block on a transient MFA
				// config read failure.
				next.ServeHTTP(w, r)
				return
			}
			if cfg == nil || !cfg.EnforceForAPI {
				next.ServeHTTP(w, r)
				return
			}

			// Break-glass bypass: any active grant covers MFA.
			bg, bgErr := repo.ListActiveBreakGlassForUser(r.Context(), claims.TenantID, claims.UserID)
			if bgErr == nil && bg != nil {
				ctx := context.WithValue(r.Context(), vaultMFAContextKey{}, &VaultMFABypass{BreakGlassID: bg.ID})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Look for a recent AuditActionMFAVerify row.
			verified, err := repo.HasRecentMFAVerify(r.Context(), claims.TenantID, claims.UserID,
				time.Duration(cfg.MFASessionTTLSeconds)*time.Second)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if verified {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Vault-MFA-Required", "true")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":           "MFA verification required",
				"code":            "VAULT_MFA_SESSION_EXPIRED",
				"action_required": "verify_mfa",
			})
		})
	}
}

// GetVaultMFABypass extracts the break-glass bypass info from the
// request context, if present. Handlers use this to include the
// break_glass_id in their audit metadata.
func GetVaultMFABypass(ctx context.Context) *VaultMFABypass {
	v, ok := ctx.Value(vaultMFAContextKey{}).(*VaultMFABypass)
	if !ok {
		return nil
	}
	return v
}
