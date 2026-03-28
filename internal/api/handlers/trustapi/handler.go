package trustapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles Trust API requests
type Handler struct {
	repo         *trustapi.Repository
	registryRepo *registry.RegistryRepository
	logger       *logrus.Logger
}

// NewHandler creates a new Trust API handler
func NewHandler(repo *trustapi.Repository, registryRepo *registry.RegistryRepository) *Handler {
	return &Handler{
		repo:         repo,
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

// getFunctionInfo retrieves function information from the registry
func (h *Handler) getFunctionInfo(functionID string) (*registry.RegistryFunction, error) {
	id, err := uuid.Parse(functionID)
	if err != nil {
		return nil, fmt.Errorf("invalid function ID: %w", err)
	}
	return h.registryRepo.GetFunctionByID(id)
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
