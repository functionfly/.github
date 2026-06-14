package trustapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles Trust API requests
type Handler struct {
	apikeyRepo   *apikey.Repository  // unified platform API key repository (for key CRUD)
	trustRepo    *trustapi.Repository // Trust API-specific repository (partners, rate limits, usage, etc.)
	registryRepo *registry.RegistryRepository
	logger       *logrus.Logger
}

// NewHandler creates a new Trust API handler
func NewHandler(apikeyRepo *apikey.Repository, trustRepo *trustapi.Repository, registryRepo *registry.RegistryRepository) *Handler {
	return &Handler{
		apikeyRepo:   apikeyRepo,
		trustRepo:    trustRepo,
		registryRepo: registryRepo,
		logger:       logrus.New(),
	}
}

// RegisterRoutes registers Trust API routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Trust endpoints (partner authenticated)
	trust := r.PathPrefix("/v1/trust").Subrouter()
	trust.HandleFunc("/score/{function_id}", h.HandleGetTrustScore).Methods("GET")
	trust.HandleFunc("/batch", h.HandleBatchTrustScore).Methods("POST")
	trust.HandleFunc("/history/{function_id}", h.HandleGetTrustHistory).Methods("GET")
	trust.HandleFunc("/verify", h.HandleSubmitVerification).Methods("POST")
	trust.HandleFunc("/verify/{verification_id}", h.HandleGetVerification).Methods("GET")
	trust.HandleFunc("/report", h.HandleSubmitReport).Methods("POST")
	trust.HandleFunc("/report/{report_id}", h.HandleGetReport).Methods("GET")

	// Partner management endpoints (authenticated)
	partners := r.PathPrefix("/v1/partners").Subrouter()
	partners.HandleFunc("", h.HandleCreatePartner).Methods("POST")
	partners.HandleFunc("", h.HandleListPartners).Methods("GET")
	partners.HandleFunc("/{partner_id}", h.HandleGetPartner).Methods("GET")
	partners.HandleFunc("/{partner_id}", h.HandleUpdatePartner).Methods("PATCH")
	partners.HandleFunc("/{partner_id}/usage", h.HandleGetPartnerUsage).Methods("GET")
	partners.HandleFunc("/{partner_id}/api-keys", h.HandleCreateAPIKey).Methods("POST")
	partners.HandleFunc("/{partner_id}/api-keys", h.HandleListAPIKeys).Methods("GET")
	partners.HandleFunc("/{partner_id}/api-keys/{key_id}", h.HandleRevokeAPIKey).Methods("DELETE")
}

// RegisterExtendedRoutes registers extended Trust API routes (requires ExtendedHandler)
func (h *ExtendedHandler) RegisterExtendedRoutes(r *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Trust revocation endpoints
	trust := r.PathPrefix("/v1/trust").Subrouter()

	// Revocation endpoints (admin/authorized only)
	revokeRouter := trust.PathPrefix("/revoke").Subrouter()
	if authMiddleware != nil {
		revokeRouter = revokeRouter.Methods("POST", "DELETE", "GET").Subrouter()
	}
	revokeRouter.HandleFunc("", h.HandleRevokeTrust).Methods("POST")
	revokeRouter.HandleFunc("/revoked", h.HandleListRevocations).Methods("GET")
	revokeRouter.HandleFunc("/revoked/{function_id}", h.HandleCheckFunctionRevoked).Methods("GET")
	revokeRouter.HandleFunc("/{revocation_id}", h.HandleGetRevocation).Methods("GET")
	revokeRouter.HandleFunc("/{revocation_id}/lift", h.HandleUnrevokeTrust).Methods("POST")

	// Attestation endpoints
	trust.HandleFunc("/attestations", h.HandleGetAttestations).Methods("GET")
	trust.HandleFunc("/attestations/{attestation_id}", h.HandleGetAttestation).Methods("GET")
	trust.HandleFunc("/attestations/{attestation_id}/verify", h.HandleVerifyAttestation).Methods("GET")
	trust.HandleFunc("/attestations/{function_id}/chain", h.HandleGetAttestationChain).Methods("GET")

	// Policy endpoints
	policies := trust.PathPrefix("/policies").Subrouter()
	policies.HandleFunc("", h.HandleCreatePolicy).Methods("POST")
	policies.HandleFunc("", h.HandleListPolicies).Methods("GET")
	policies.HandleFunc("/evaluate", h.HandleEvaluatePolicy).Methods("POST")
	policies.HandleFunc("/evaluate/batch", h.HandleBatchEvaluatePolicy).Methods("POST")
	policies.HandleFunc("/{policy_id}", h.HandleGetPolicy).Methods("GET")
	policies.HandleFunc("/{policy_id}", h.HandleUpdatePolicy).Methods("PUT")
	policies.HandleFunc("/{policy_id}", h.HandleDeletePolicy).Methods("DELETE")
}

// getFunctionInfo retrieves function information from the registry
func (h *Handler) getFunctionInfo(ctx context.Context, functionID string) (*registry.RegistryFunction, error) {
	id, err := uuid.Parse(functionID)
	if err != nil {
		return nil, fmt.Errorf("invalid function ID: %w", err)
	}
	return h.registryRepo.GetFunctionByID(ctx, id)
}

// parseFunctionID parses and validates a function ID from URL params
func (h *Handler) parseFunctionID(functionIDStr string) (string, error) {
	if functionIDStr == "" {
		return "", nil
	}
	return functionIDStr, nil
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, err string, code string) {
	h.writeJSON(w, status, trustapi.ErrorResponse{
		Error: err,
		Code:  code,
	})
}
