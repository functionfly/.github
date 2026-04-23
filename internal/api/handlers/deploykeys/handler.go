package deploykeys

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo   *storage.DeployKeysRepository
	logger *logrus.Logger
}

func NewHandler(repo *storage.DeployKeysRepository) *Handler {
	return &Handler{
		repo:   repo,
		logger: logrus.New(),
	}
}

func (h *Handler) SetLogger(logger *logrus.Logger) {
	h.logger = logger
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errMsg string, code string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": errMsg,
		"code":  code,
	})
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	var req storage.DeployKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	key, err := h.repo.Create(r.Context(), claims.TenantID, req.Name, req.PublicKey, &claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create deploy key")
		h.writeError(w, http.StatusInternalServerError, "Failed to create deploy key: "+err.Error(), "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"key_id":    key.ID,
		"tenant_id": claims.TenantID,
	}).Info("Deploy key created")

	h.writeJSON(w, http.StatusCreated, storage.DeployKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		PublicKey:   key.PublicKey,
		Fingerprint: key.Fingerprint,
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
		CreatedBy:   key.CreatedBy,
	})
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	keys, total, err := h.repo.ListByTenant(r.Context(), claims.TenantID, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list deploy keys")
		h.writeError(w, http.StatusInternalServerError, "Failed to list deploy keys", "internal_error")
		return
	}

	responseList := make([]storage.DeployKeyResponse, len(keys))
	for i, key := range keys {
		responseList[i] = storage.DeployKeyResponse{
			ID:          key.ID,
			Name:        key.Name,
			PublicKey:   key.PublicKey,
			Fingerprint: key.Fingerprint,
			CreatedAt:   key.CreatedAt,
			ExpiresAt:   key.ExpiresAt,
			LastUsedAt:  key.LastUsedAt,
			CreatedBy:   key.CreatedBy,
		}
	}

	h.writeJSON(w, http.StatusOK, storage.DeployKeyListResponse{
		DeployKeys: responseList,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid key ID", "invalid_id")
		return
	}

	key, err := h.repo.GetByID(r.Context(), keyID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Deploy key not found", "not_found")
		return
	}

	h.writeJSON(w, http.StatusOK, storage.DeployKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		PublicKey:   key.PublicKey,
		Fingerprint: key.Fingerprint,
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
		LastUsedAt:  key.LastUsedAt,
		CreatedBy:   key.CreatedBy,
	})
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid key ID", "invalid_id")
		return
	}

	if err := h.repo.Delete(r.Context(), keyID, claims.TenantID); err != nil {
		h.logger.WithError(err).Error("Failed to delete deploy key")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete deploy key", "internal_error")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"key_id":    keyID,
		"tenant_id": claims.TenantID,
	}).Info("Deploy key deleted")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Deploy key deleted successfully",
	})
}

func (h *Handler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid key ID", "invalid_id")
		return
	}

	key, err := h.repo.VerifyKey(r.Context(), keyID, claims.TenantID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Deploy key verification failed: "+err.Error(), "verification_failed")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":       true,
		"fingerprint": key.Fingerprint,
	})
}
