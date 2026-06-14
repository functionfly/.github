package providers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleConnectProvider validates an API token for the given provider and, if valid, saves
// it encrypted in the database. Returns a ConnectedProvider JSON object on success.
func (h *Handler) HandleConnectProvider(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	ipAddress := middleware.ExtractClientIP(r)
	userAgent := r.UserAgent()
	requestID := ""
	if id := ctx.Value("request_id"); id != nil {
		requestID = id.(string)
	}

	var req struct {
		ProviderID string `json:"providerId"`
		APIKey     string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ProviderID == "" {
		http.Error(w, "providerId is required", http.StatusBadRequest)
		return
	}


	if req.ProviderID == "functionfly-edge" {
		if existing, _ := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, "functionfly-edge"); existing != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"provider": providerResponse(existing),
			})
			return
		}

		var edgeAPIKey string
		var edgeAPIKeyID string

		if h.apikeyRepo != nil {
			generator := apikey.NewGenerator()
			plaintext, _ := generator.Generate(apikey.KeyTypeEdge)
			edgeAPIKey = plaintext

			hasher := apikey.NewHasher()
			keyHash := hasher.Hash(plaintext)

			now := time.Now()
			keyID := uuid.New()
			apiKeyRecord := &apikey.APIKey{
				ID:                    keyID,
				TenantID:              claims.TenantID,
				UserID:                claims.UserID,
				Name:                  "FunctionFly Edge",
				Description:           "Auto-generated Edge API key for FunctionFly Edge provider",
				KeyType:               apikey.KeyTypeEdge,
				KeyID:                 plaintext,
				KeyPrefix:             apikey.PrefixEdge,
				KeyHash:               keyHash,
				KeyVersion:            1,
				LastRotatedAt:         now,
				RotationFrequencyDays: 90,
				RateLimitRPM:          1000,
				RateLimitRPH:          60000,
				RateLimitRPD:          1000000,
				IsActive:              true,
				Metadata:              apikey.JSONBMap{"provider_id": "functionfly-edge"},
				CreatedAt:             now,
				UpdatedAt:             now,
				CreatedBy:             "functionfly-edge-connect",
			}
			if err := h.apikeyRepo.CreatePreGenerated(ctx, apiKeyRecord); err != nil {
				logrus.WithError(err).Error("Failed to store Edge API key")
				http.Error(w, "Failed to store API key", http.StatusInternalServerError)
				return
			}
			edgeAPIKeyID = keyID.String()
		}

		idBytes := make([]byte, 16)
		rand.Read(idBytes)
		provider := &storage.Provider{
			ID:       hex.EncodeToString(idBytes),
			UserID:   claims.UserID,
			Provider: "functionfly-edge",
			Token:    "managed",
			Status:   "active",
		}
		if err := h.repo.CreateProvider(r.Context(), provider); err != nil {
			logrus.WithError(err).Error("Failed to store functionfly-edge provider")
			http.Error(w, "Failed to enable provider", http.StatusInternalServerError)
			return
		}

		providerUUID, _ := uuid.Parse(provider.ID)
		utils.LogAuditEvent(ctx, h.repo, r, "provider.connect", "provider", &providerUUID, nil, map[string]interface{}{
			"provider_id":   provider.ID,
			"provider_type": "functionfly-edge",
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
			"request_id":    requestID,
		}, true)

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"provider": providerResponse(provider),
		}
		if edgeAPIKey != "" {
			response["apiKey"] = edgeAPIKey
			response["apiKeyId"] = edgeAPIKeyID
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.APIKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}

	providerName := mapProviderIDForValidation(req.ProviderID)

	validation := h.validateProviderToken(providerName, req.APIKey)
	if !validation.IsValid {
		if h.notify != nil {
			_ = h.notify.SendProviderDegraded(r.Context(), claims.UserID, req.ProviderID, providerName, validation.Message)
		}
		http.Error(w, validation.Message, http.StatusUnprocessableEntity)
		return
	}

	existing, _ := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, req.ProviderID)
	if existing != nil {
		if _, err := h.repo.UpdateProvider(r.Context(), existing.ID, map[string]interface{}{
			"token":  req.APIKey,
			"status": "active",
		}); err != nil {
			logrus.WithError(err).Error("Failed to update existing provider")
			http.Error(w, "Failed to update provider", http.StatusInternalServerError)
			return
		}
		existing.Status = "active"

		if h.notify != nil {
			_ = h.notify.SendProviderOnline(r.Context(), claims.UserID, existing.ID, existing.Provider)
		}

		existingUUID, _ := uuid.Parse(existing.ID)
		utils.LogAuditEvent(ctx, h.repo, r, "provider.update", "provider", &existingUUID, map[string]interface{}{
			"provider_id":   existing.ID,
			"provider_type": req.ProviderID,
			"status":        "active",
		}, map[string]interface{}{
			"provider_id":   existing.ID,
			"provider_type": req.ProviderID,
			"status":        "active",
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
			"request_id":    requestID,
		}, true)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider": providerResponse(existing),
		})
		return
	}

	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	provider := &storage.Provider{
		ID:       hex.EncodeToString(idBytes),
		UserID:   claims.UserID,
		Provider: req.ProviderID,
		Token:    req.APIKey,
		Status:   "active",
	}
	if err := h.repo.CreateProvider(r.Context(), provider); err != nil {
		logrus.WithError(err).Error("Failed to store provider")
		http.Error(w, "Failed to save provider", http.StatusInternalServerError)
		return
	}

	providerUUID, _ := uuid.Parse(provider.ID)
	utils.LogAuditEvent(ctx, h.repo, r, "provider.connect", "provider", &providerUUID, nil, map[string]interface{}{
		"provider_id":   provider.ID,
		"provider_type": req.ProviderID,
		"ip_address":    ipAddress,
		"user_agent":    userAgent,
		"request_id":    requestID,
	}, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider": providerResponse(provider),
	})
}

// HandleDisconnectProvider deletes a connected provider for the authenticated user.
func (h *Handler) HandleDisconnectProvider(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	ipAddress := middleware.ExtractClientIP(r)
	userAgent := r.UserAgent()
	requestID := ""
	if id := ctx.Value("request_id"); id != nil {
		requestID = id.(string)
	}

	vars := mux.Vars(r)
	providerID := vars["providerId"]
	if providerID == "" {
		http.Error(w, "providerId is required", http.StatusBadRequest)
		return
	}

	provider, _ := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, providerID)

	if err := h.repo.DeleteProvider(r.Context(), providerID, claims.UserID); err != nil {
		logrus.WithError(err).WithField("providerID", providerID).Error("Failed to delete provider")

		providerUUID, _ := uuid.Parse(providerID)
		utils.LogAuditEvent(ctx, h.repo, r, "provider.disconnect", "provider", &providerUUID, map[string]interface{}{
			"provider_id": providerID,
			"status":      "failed",
			"error":       err.Error(),
		}, nil, false)

		http.Error(w, "Failed to disconnect provider", http.StatusInternalServerError)
		return
	}

	var providerUUID uuid.UUID
	if provider != nil {
		providerUUID, _ = uuid.Parse(provider.ID)
		utils.LogAuditEvent(ctx, h.repo, r, "provider.disconnect", "provider", &providerUUID, map[string]interface{}{
			"provider_id":   provider.ID,
			"provider_type": provider.Provider,
		}, map[string]interface{}{
			"provider_id": providerID,
			"status":      "disconnected",
			"ip_address":  ipAddress,
			"user_agent":  userAgent,
			"request_id":  requestID,
		}, true)
	} else {
		providerUUID, _ = uuid.Parse(providerID)
		utils.LogAuditEvent(ctx, h.repo, r, "provider.disconnect", "provider", &providerUUID, map[string]interface{}{
			"provider_id": providerID,
		}, map[string]interface{}{
			"provider_id": providerID,
			"status":      "disconnected",
			"ip_address":  ipAddress,
			"user_agent":  userAgent,
			"request_id":  requestID,
		}, true)
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleTestConnection tests whether the stored credentials for a provider still work.
func (h *Handler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	providerID := vars["providerId"]

	provider, err := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, providerID)
	if err != nil || provider == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Provider not found"})
		return
	}

	if err := h.repo.UpdateProviderLastUsed(ctx, provider.ID); err != nil {
		logrus.WithError(err).WithField("provider_id", provider.ID).Warn("Failed to update provider last_used_at")
	}

	if providerID == "functionfly-edge" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "FunctionFly Edge is operational"})
		return
	}

	result := h.validateProviderToken(mapProviderIDForValidation(providerID), provider.Token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": result.IsValid,
		"message": result.Message,
	})
}

// HandleRotateProvider rotates an API key for an existing provider.
func (h *Handler) HandleRotateProvider(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	ipAddress := middleware.ExtractClientIP(r)
	userAgent := r.UserAgent()
	requestID := ""
	if id := ctx.Value("request_id"); id != nil {
		requestID = id.(string)
	}

	vars := mux.Vars(r)
	providerID := vars["providerId"]
	if providerID == "" {
		http.Error(w, "providerId is required", http.StatusBadRequest)
		return
	}

	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.APIKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}

	validation := h.validateProviderToken(mapProviderIDForValidation(providerID), req.APIKey)
	if !validation.IsValid {
		http.Error(w, validation.Message, http.StatusUnprocessableEntity)
		return
	}

	existing, err := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, providerID)
	if err != nil || existing == nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	updated, err := h.repo.UpdateProvider(ctx, existing.ID, map[string]interface{}{
		"token":  req.APIKey,
		"status": "active",
	})

	providerUUID, _ := uuid.Parse(existing.ID)
	utils.LogAuditEvent(ctx, h.repo, r, "provider.rotate", "provider", &providerUUID, map[string]interface{}{
		"provider_id":   existing.ID,
		"provider_type": providerID,
	}, map[string]interface{}{
		"provider_id":   existing.ID,
		"provider_type": providerID,
		"ip_address":    ipAddress,
		"user_agent":    userAgent,
		"request_id":    requestID,
	}, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider": providerResponse(updated),
	})
}
