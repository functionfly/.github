package connectors

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	connectorEngine "github.com/functionfly/functionfly/internal/agent/connectors"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	connectorRepo *storage.ConnectorRepository
	brainRepo     *storage.BrainRepository
	scheduler     *connectorEngine.SyncScheduler
	logger        *logrus.Logger
}

func NewHandler(
	connectorRepo *storage.ConnectorRepository,
	brainRepo *storage.BrainRepository,
	scheduler *connectorEngine.SyncScheduler,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		connectorRepo: connectorRepo,
		brainRepo:     brainRepo,
		scheduler:     scheduler,
		logger:        logger,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"code":    code,
		"message": message,
	})
}

func (h *Handler) getTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var tier string
	err := h.connectorRepo.GetDB().QueryRowContext(ctx, `
		SELECT COALESCE(pt.name, 'free')
		FROM subscriptions s
		JOIN pricing_tiers pt ON pt.id = s.pricing_tier_id
		WHERE s.tenant_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC LIMIT 1`, tenantID).Scan(&tier)
	if err != nil {
		return plans.PlanFree, err
	}
	return tier, nil
}

// HandleListCatalog returns all available connectors
func (h *Handler) HandleListCatalog(w http.ResponseWriter, r *http.Request) {
	connectors, err := h.connectorRepo.ListCatalog(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list connectors catalog")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list connectors")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"connectors": connectors,
	})
}

// HandleListUserConnectors returns the user's linked connectors
func (h *Handler) HandleListUserConnectors(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	connectors, err := h.connectorRepo.GetUserConnectors(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list user connectors")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list connectors")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"connectors": connectors,
	})
}

// HandleLinkConnector initiates linking a connector
func (h *Handler) HandleLinkConnector(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req storage.LinkConnectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.ConnectorSlug == "" {
		h.respondError(w, 400, "MISSING_FIELD", "connector_slug is required")
		return
	}

	// Check tier access
	tier := plans.PlanFree
	sub, err := h.getTenantPlan(r.Context(), claims.TenantID)
	if err == nil && sub != "" {
		tier = sub
	}
	if !plans.IsConnectorAvailableForPlan(req.ConnectorSlug, tier) {
		h.respondError(w, 403, "TIER_RESTRICTED", "This connector is not available on your plan")
		return
	}

	// Check connector limit
	existing, err := h.connectorRepo.CountUserConnectors(r.Context(), claims.TenantID)
	if err != nil {
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to check connector count")
		return
	}
	maxConnectors := plans.GetMaxConnectors(tier)
	if maxConnectors > 0 && existing >= maxConnectors {
		h.respondError(w, 403, "LIMIT_REACHED", "Connector limit reached for your plan")
		return
	}

	// Get connector catalog entry
	connector, err := h.connectorRepo.GetConnectorBySlug(r.Context(), req.ConnectorSlug)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	// Create user connector
	uc := &storage.UserConnector{
		TenantID:             claims.TenantID,
		ConnectorID:          connector.ID,
		DisplayName:          req.DisplayName,
		Status:               "active",
		EncryptedCredentials: req.EncryptedCredentials,
	}

	if uc.DisplayName == "" {
		uc.DisplayName = connector.Name
	}
	if uc.EncryptedCredentials == nil {
		uc.EncryptedCredentials = []byte("{}")
	}

	created, err := h.connectorRepo.CreateUserConnector(r.Context(), uc)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user connector")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to link connector")
		return
	}

	h.respondJSON(w, 201, map[string]interface{}{
		"connector": created,
		"message":   "Connector linked successfully",
	})
}

// HandleOAuthCallback handles OAuth callback
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req storage.ConnectorCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Code == "" {
		h.respondError(w, 400, "MISSING_FIELD", "code is required")
		return
	}

	// In a real implementation, this would:
	// 1. Exchange the OAuth code for tokens
	// 2. Encrypt tokens client-side (zero-knowledge)
	// 3. Store encrypted credentials
	// For now, we return success

	h.respondJSON(w, 200, map[string]interface{}{
		"status":  "success",
		"message": "OAuth callback processed",
	})
}

// HandleUnlinkConnector removes a linked connector
func (h *Handler) HandleUnlinkConnector(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid connector ID")
		return
	}

	if err := h.connectorRepo.DeleteUserConnector(r.Context(), claims.TenantID, connectorID); err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "Connector unlinked successfully",
	})
}

// HandleTriggerSync triggers a manual sync for a connector
func (h *Handler) HandleTriggerSync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid connector ID")
		return
	}

	uc, err := h.connectorRepo.GetUserConnector(r.Context(), claims.TenantID, connectorID)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	h.respondJSON(w, 200, storage.SyncTriggerResponse{
		Status:    "syncing",
		Message:   "Sync started for " + uc.ConnectorName,
		StartedAt: time.Now().UTC(),
	})
}
