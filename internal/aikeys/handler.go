package aikeys

import (
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler manages BYOK AI provider key endpoints.
type Handler struct {
	repo      *Repository
	auditRepo storage.Repository
}

// NewHandler creates a new BYOK handler.
func NewHandler(repo *Repository, auditRepo storage.Repository) *Handler {
	return &Handler{
		repo:      repo,
		auditRepo: auditRepo,
	}
}

// HandleConnectKey validates and stores a BYOK key.
func (h *Handler) HandleConnectKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req ConnectRequest
	if err := DecodeJSON(r, &req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Provider == "" {
		apierror.WriteError(w, apierror.NewBadRequest("provider is required"))
		return
	}
	if req.APIKey == "" {
		apierror.WriteError(w, apierror.NewBadRequest("apiKey is required"))
		return
	}

	// Validate region for token plan keys
	if req.Provider == "mimo-token-plan" {
		if !ValidTokenPlanRegions[req.Region] {
			apierror.WriteError(w, apierror.NewBadRequest("region is required for MiMo Token Plan (cn, sgp, eu)"))
			return
		}
	}

	// Validate format + test API call
	// Token plan keys: format-only validation (skip live test to avoid rate limit friction)
	var validation ValidationResponse
	if req.Provider == "mimo-token-plan" || req.Provider == "minimax-token-plan" {
		if err := validateFormat(req.Provider, req.APIKey); err != nil {
			validation = ValidationResponse{IsValid: false, Message: err.Error()}
		} else {
			validation = ValidationResponse{IsValid: true, Message: "token plan key format validated"}
		}
	} else {
		validation = ValidateProviderKey(req.Provider, req.APIKey)
	}
	if !validation.IsValid {
		apierror.WriteError(w, apierror.NewValidation("Validation failed: "+validation.Message))
		return
	}

	// Encrypt the key (tag appended to ciphertext, matching dynamic_encryption pattern)
	ciphertext, nonce, version, err := EncryptKey([]byte(req.APIKey), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to encrypt BYOK key")
		apierror.WriteError(w, apierror.NewInternal("Failed to encrypt key"))
		return
	}

	// Check for existing key
	existing, _ := h.repo.GetByTenantAndProvider(r.Context(), claims.TenantID, req.Provider)
	if existing != nil {
		now := time.Now()
		healthMsg := ""
		if req.Provider == "mimo-token-plan" {
			healthMsg = "region:" + req.Region
		}
		err = h.repo.Update(r.Context(), existing.ID, map[string]interface{}{
			"encrypted_key":   ciphertext,
			"key_nonce":       nonce,
			"key_version":     version,
			"key_last4":       KeyLast4(req.APIKey),
			"status":          "active",
			"health_message":  healthMsg,
			"last_health_check": nil,
			"updated_at":      now,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to update BYOK key")
			apierror.WriteError(w, apierror.NewInternal("Failed to update key"))
			return
		}

		utils.LogAuditEvent(r.Context(), h.auditRepo, r, "ai_key.update", "ai_provider_key", nil, map[string]interface{}{
			"provider":   req.Provider,
			"key_last4":  existing.KeyLast4,
			"key_id":     existing.ID,
		}, map[string]interface{}{
			"provider":   req.Provider,
			"key_last4":  KeyLast4(req.APIKey),
			"key_id":     existing.ID,
		}, true)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"key": KeyResponse{
				ID:              existing.ID,
				Provider:        req.Provider,
				KeyLast4:        KeyLast4(req.APIKey),
				Status:          "active",
				TokenPlanRegion: req.Region,
			},
		})
		return
	}

	// Create new key
	id, err := GenerateID()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate BYOK key ID")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate ID"))
		return
	}

	healthMsg := ""
	if req.Provider == "mimo-token-plan" {
		healthMsg = "region:" + req.Region
	}

	key := &types.AIProviderKey{
		ID:            id,
		TenantID:      claims.TenantID,
		Provider:      req.Provider,
		EncryptedKey:  ciphertext,
		KeyNonce:      nonce,
		KeyTag:        []byte{},
		KeyVersion:    version,
		KeyLast4:      KeyLast4(req.APIKey),
		Status:        "active",
		HealthMessage: healthMsg,
		ConnectedBy:   claims.UserID,
	}

	if err := h.repo.Create(r.Context(), key); err != nil {
		logrus.WithError(err).Error("Failed to store BYOK key")
		apierror.WriteError(w, apierror.NewInternal("Failed to store key"))
		return
	}

	utils.LogAuditEvent(r.Context(), h.auditRepo, r, "ai_key.connect", "ai_provider_key", nil, nil, map[string]interface{}{
		"provider":  req.Provider,
		"key_last4": KeyLast4(req.APIKey),
		"key_id":    id,
	}, true)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"key": KeyResponse{
			ID:              key.ID,
			Provider:        key.Provider,
			KeyLast4:        key.KeyLast4,
			Status:          key.Status,
			CreatedAt:       key.CreatedAt.Format(time.RFC3339),
			TokenPlanRegion: req.Region,
		},
	})
}

// HandleListKeys lists all BYOK keys for the tenant.
func (h *Handler) HandleListKeys(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	keys, err := h.repo.ListByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list BYOK keys")
		apierror.WriteError(w, apierror.NewInternal("Failed to list keys"))
		return
	}

	responses := make([]KeyResponse, len(keys))
	for i := range keys {
		responses[i] = toKeyResponse(&keys[i])
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"keys": responses,
	})
}

// HandleDisconnectKey removes a BYOK key.
func (h *Handler) HandleDisconnectKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	provider := mux.Vars(r)["provider"]
	if provider == "" {
		apierror.WriteError(w, apierror.NewBadRequest("provider is required"))
		return
	}

	existing, _ := h.repo.GetByTenantAndProvider(r.Context(), claims.TenantID, provider)
	if existing == nil {
		apierror.WriteError(w, apierror.NewNotFound("No BYOK key found for provider: "+provider))
		return
	}

	if err := h.repo.DeleteByTenantAndProvider(r.Context(), claims.TenantID, provider); err != nil {
		logrus.WithError(err).Error("Failed to delete BYOK key")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete key"))
		return
	}

	utils.LogAuditEvent(r.Context(), h.auditRepo, r, "ai_key.disconnect", "ai_provider_key", nil, map[string]interface{}{
		"provider":   provider,
		"key_last4":  existing.KeyLast4,
		"key_id":     existing.ID,
	}, nil, true)

	w.WriteHeader(http.StatusNoContent)
}

// HandleTestKey tests if a BYOK key still works.
func (h *Handler) HandleTestKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	provider := mux.Vars(r)["provider"]

	existing, err := h.repo.GetByTenantAndProvider(r.Context(), claims.TenantID, provider)
	if err != nil || existing == nil {
		writeJSON(w, http.StatusOK, TestResponse{Success: false, Message: "No BYOK key found for this provider"})
		return
	}

	// Decrypt the key
	plaintext, err := DecryptKey(existing.EncryptedKey, existing.KeyNonce, claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to decrypt BYOK key for test")
		writeJSON(w, http.StatusOK, TestResponse{Success: false, Message: "Failed to decrypt key"})
		return
	}

	// Use region-specific endpoint for token plan keys
	var result ValidationResponse
	if provider == "mimo-token-plan" || provider == "minimax-token-plan" {
		region := extractRegion(provider, existing.HealthMessage)
		result = TestProviderAPIWithRegion(provider, string(plaintext), region)
	} else {
		result = testProviderAPI(provider, string(plaintext))
	}

	newStatus := "active"
	if !result.IsValid {
		newStatus = "degraded"
	}
	_ = h.repo.UpdateHealthStatus(r.Context(), existing.ID, newStatus, result.Message)
	_ = h.repo.UpdateLastUsed(r.Context(), existing.ID)

	writeJSON(w, http.StatusOK, TestResponse{
		Success: result.IsValid,
		Message: result.Message,
	})
}

// HandleRotateKey replaces a BYOK key with a new one.
func (h *Handler) HandleRotateKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	provider := mux.Vars(r)["provider"]

	var req RotateRequest
	if err := DecodeJSON(r, &req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.APIKey == "" {
		apierror.WriteError(w, apierror.NewBadRequest("apiKey is required"))
		return
	}

	existing, _ := h.repo.GetByTenantAndProvider(r.Context(), claims.TenantID, provider)
	if existing == nil {
		apierror.WriteError(w, apierror.NewNotFound("No BYOK key found for provider: "+provider))
		return
	}

	validation := ValidateProviderKey(provider, req.APIKey)
	if !validation.IsValid {
		apierror.WriteError(w, apierror.NewValidation("Validation failed: "+validation.Message))
		return
	}

	ciphertext, nonce, version, err := EncryptKey([]byte(req.APIKey), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to encrypt rotated BYOK key")
		apierror.WriteError(w, apierror.NewInternal("Failed to encrypt key"))
		return
	}

	now := time.Now()
	err = h.repo.Update(r.Context(), existing.ID, map[string]interface{}{
		"encrypted_key":   ciphertext,
		"key_nonce":       nonce,
		"key_version":     version,
		"key_last4":       KeyLast4(req.APIKey),
		"status":          "active",
		"health_message":  "",
		"last_health_check": nil,
		"updated_at":      now,
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to update rotated BYOK key")
		apierror.WriteError(w, apierror.NewInternal("Failed to update key"))
		return
	}

	utils.LogAuditEvent(r.Context(), h.auditRepo, r, "ai_key.rotate", "ai_provider_key", nil, map[string]interface{}{
		"key_last4":  existing.KeyLast4,
		"key_id":     existing.ID,
	}, map[string]interface{}{
		"key_last4":  KeyLast4(req.APIKey),
		"key_id":     existing.ID,
	}, true)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key": KeyResponse{
			ID:       existing.ID,
			Provider: provider,
			KeyLast4: KeyLast4(req.APIKey),
			Status:   "active",
		},
	})
}

// HandleListSupportedProviders returns all providers available for BYOK.
func (h *Handler) HandleListSupportedProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": SupportedProviders(),
	})
}

func toKeyResponse(k *types.AIProviderKey) KeyResponse {
	var lastCheck, lastUsed *string
	if k.LastHealthCheck != nil {
		s := k.LastHealthCheck.Format(time.RFC3339)
		lastCheck = &s
	}
	if k.LastUsedAt != nil {
		s := k.LastUsedAt.Format(time.RFC3339)
		lastUsed = &s
	}
	kr := KeyResponse{
		ID:              k.ID,
		Provider:        k.Provider,
		KeyLast4:        k.KeyLast4,
		Status:          k.Status,
		HealthMessage:   k.HealthMessage,
		LastHealthCheck: lastCheck,
		LastUsedAt:      lastUsed,
		CreatedAt:       k.CreatedAt.Format(time.RFC3339),
		TokenPlanRegion: extractRegion(k.Provider, k.HealthMessage),
	}
	return kr
}

// extractRegion extracts the region code from health_message for token plan keys.
func extractRegion(provider, healthMessage string) string {
	if provider != "mimo-token-plan" {
		return ""
	}
	if strings.HasPrefix(healthMessage, "region:") {
		return strings.TrimPrefix(healthMessage, "region:")
	}
	return ""
}


