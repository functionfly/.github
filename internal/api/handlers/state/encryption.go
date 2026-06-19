package state

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleMigrateEncryption handles POST /v1/state/encrypt - Encrypt existing state values at rest
func (h *Handler) HandleMigrateEncryption(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check if encryption is enabled server-side
	if !h.stateRepo.IsEncryptionEnabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("server-side encryption not configured"))
		return
	}

	var req EncryptionMigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	if req.BatchSize == 0 {
		req.BatchSize = 100
	}
	if req.BatchSize > 1000 {
		req.BatchSize = 1000 // Max batch size for safety
	}

	tenantID := claims.TenantID

	// Track migration results
	response := &EncryptionMigrationResponse{
		StatesProcessed: 0,
		ValuesEncrypted: 0,
		ValuesSkipped:   0,
		Errors:          []string{},
		Completed:       true,
	}

	// If specific state ID provided, migrate only that state
	if req.StateID != "" {
		stateUUID, err := uuid.Parse(req.StateID)
		if err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("invalid state ID"))
			return
		}

		state, err := h.stateRepo.GetStateByID(r.Context(), stateUUID)
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("state not found"))
			return
		}

		// Verify tenant ownership
		if state.TenantID != tenantID {
			apierror.WriteError(w, apierror.NewForbidden("forbidden"))
			return
		}

		errs := h.migrateStateEncryption(r.Context(), state, req.DryRun, req.ForceRotate)
		response.StatesProcessed = 1
		response.Errors = append(response.Errors, errs...)
	} else {
		// Migrate all states for tenant
		states, _, err := h.stateRepo.ListStatesByTenant(r.Context(), tenantID, 1000, 0)
		if err != nil {
			logrus.Errorf("failed to list states for encryption migration: %v", err)
			apierror.WriteError(w, apierror.NewInternal("failed to list states"))
			return
		}

		for _, state := range states {
			errs := h.migrateStateEncryption(r.Context(), state, req.DryRun, req.ForceRotate)
			response.StatesProcessed++
			response.Errors = append(response.Errors, errs...)

			// Stop if too many errors
			if len(response.Errors) >= 50 {
				response.Completed = false
				response.Errors = append(response.Errors, "migration stopped due to too many errors")
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// migrateStateEncryption encrypts all unencrypted values in a state
func (h *Handler) migrateStateEncryption(ctx context.Context, state *staterepo.State, dryRun, forceRotate bool) []string {
	errors := []string{}

	// Get all values for this state
	values, err := h.stateRepo.GetAllStateValues(ctx, state.ID)
	if err != nil {
		return []string{fmt.Sprintf("failed to get values for state %s: %v", state.ID, err)}
	}

	for _, value := range values {
		// Skip if already encrypted and not forcing rotation
		if value.IsEncrypted && !forceRotate {
			continue
		}

		// Dry run - just count what would be encrypted
		if dryRun {
			continue
		}

		// Encrypt the value
		if _, err := h.stateRepo.SetStateValue(ctx, state.ID, value.Key, value.Value, "encryption_migration", "system"); err != nil {
			errors = append(errors, fmt.Sprintf("failed to encrypt value %s in state %s: %v", value.Key, state.ID, err))
		}
	}

	// Update state to mark as encrypted
	if !dryRun && !state.IsEncrypted {
		state.IsEncrypted = true
		if _, err := h.stateRepo.UpdateState(ctx, state); err != nil {
			errors = append(errors, fmt.Sprintf("failed to update state %s encryption flag: %v", state.ID, err))
		}
	}

	return errors
}

// HandleGetEncryptionStats handles GET /v1/state/encryption-stats - Get encryption statistics
func (h *Handler) HandleGetEncryptionStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	// Get all states for tenant
	states, _, err := h.stateRepo.ListStatesByTenant(r.Context(), tenantID, 1000, 0)
	if err != nil {
		logrus.Errorf("failed to list states for encryption stats: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to get states"))
		return
	}

	stats := &EncryptionStatsResponse{
		TotalStates:       len(states),
		EncryptedStates:   0,
		UnencryptedStates: 0,
		TotalValues:       0,
		EncryptedValues:   0,
		UnencryptedValues: 0,
		EncryptionEnabled: h.stateRepo.IsEncryptionEnabled(),
	}

	for _, state := range states {
		if state.IsEncrypted {
			stats.EncryptedStates++
		} else {
			stats.UnencryptedStates++
		}

		// Count values - this is a simplified count, in production you'd want a more efficient query
		values, err := h.stateRepo.GetAllStateValues(r.Context(), state.ID)
		if err != nil {
			logrus.WithError(err).Warnf("failed to get values for state %s", state.ID)
			continue
		}

		for _, value := range values {
			stats.TotalValues++
			if value.IsEncrypted {
				stats.EncryptedValues++
			} else {
				stats.UnencryptedValues++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleEnableEncryption handles POST /v1/state/{path}/enable-encryption - Enable encryption for a specific state
func (h *Handler) HandleEnableEncryption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check if encryption is enabled server-side
	if !h.stateRepo.IsEncryptionEnabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("server-side encryption not configured"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	// Check admin permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_admin") {
		return
	}

	// Already encrypted
	if state.IsEncrypted {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"state_id":  state.ID,
			"encrypted": true,
			"message":   "state encryption already enabled",
		})
		return
	}

	// Enable encryption and migrate values
	state.IsEncrypted = true
	updated, err := h.stateRepo.UpdateState(r.Context(), state)
	if err != nil {
		logrus.Errorf("failed to enable encryption for state: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to enable encryption"))
		return
	}

	// Migrate existing values in the background (or sync for now)
	go func() {
		ctx := context.Background()
		values, _ := h.stateRepo.GetAllStateValues(ctx, state.ID)
		for _, value := range values {
			if !value.IsEncrypted {
				h.stateRepo.SetStateValue(ctx, state.ID, value.Key, value.Value, "encryption_enable", "system")
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"state_id":   updated.ID,
		"encrypted":  true,
		"message":    "state encryption enabled and values are being migrated",
		"migrating":  true,
		"started_at": time.Now().Format(time.RFC3339),
	})
}

// HandleRotateEncryptionKey handles POST /v1/state/rotate-key - Rotate encryption key
func (h *Handler) HandleRotateEncryptionKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check if encryption is enabled
	if !h.stateRepo.IsEncryptionEnabled() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("server-side encryption not configured"))
		return
	}

	// Read new key from request (base64 encoded)
	var req struct {
		NewKeyBase64 string `json:"new_key_base64"`
		BatchSize    int    `json:"batch_size,omitempty"`
		DryRun       bool   `json:"dry_run,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	if req.BatchSize == 0 {
		req.BatchSize = 100
	}

	// This is a placeholder - actual key rotation would require:
	// 1. Decrypt all values with old key
	// 2. Re-encrypt with new key
	// 3. Update state values atomically

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "key rotation endpoint - set ENCRYPTED_STATE_ENABLED=true and STATE_ENCRYPTION_KEY env var, then call /state/encrypt with force_rotate=true",
		"note":    "Key rotation is currently done by updating STATE_ENCRYPTION_KEY and re-encrypting all values",
		"dry_run": req.DryRun,
	})
}

// boolPtr safely converts a bool to a *bool
func boolPtr(b bool) *bool {
	return &b
}

// intPtr safely converts an int to a *int
func intPtr(i int) *int {
	return &i
}
