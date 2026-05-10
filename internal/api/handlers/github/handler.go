package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo        storage.Repository
	githubRepo  *storage.GitHubRepository
	registryRepo *storageregistry.RegistryRepository
	logger      *logrus.Logger
	vaultKey    string
	baseURL     string
	authSvc     *auth.AuthService
	progressCh  sync.Map
	importSem   sync.Map // tenant UUID -> *int32 (atomic concurrent import count)
}

const maxConcurrentImportsPerTenant = 10

func NewHandler(repo storage.Repository, githubRepo *storage.GitHubRepository, registryRepo *storageregistry.RegistryRepository, logger *logrus.Logger, vaultKey, baseURL string) *Handler {
	return &Handler{
		repo:        repo,
		githubRepo:  githubRepo,
		registryRepo: registryRepo,
		logger:      logger,
		vaultKey:    vaultKey,
		baseURL:     baseURL,
	}
}

// SetAuthService sets the auth service for SSE token validation.
func (h *Handler) SetAuthService(svc *auth.AuthService) {
	h.authSvc = svc
}

// SetRegistryRepo sets the registry repository for function creation during import.
func (h *Handler) SetRegistryRepo(repo *storageregistry.RegistryRepository) {
	h.registryRepo = repo
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func (h *Handler) parseUUID(r *http.Request, key string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	return uuid.Parse(vars[key])
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return nil
	}
	return claims
}

// requireAuthOrToken checks Authorization header first, then falls back to ?token= query param.
// This is needed for SSE (EventSource) which cannot set custom headers.
func (h *Handler) requireAuthOrToken(w http.ResponseWriter, r *http.Request) *auth.Claims {
	// Try standard auth first (from middleware)
	claims := middleware.GetUserFromContext(r)
	if claims != nil {
		return claims
	}

	// Fall back to token query param for SSE
	token := r.URL.Query().Get("token")
	if token == "" {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return nil
	}

	if h.authSvc == nil {
		h.respondError(w, http.StatusInternalServerError, "config_error", "Auth service not available")
		return nil
	}

	parsedClaims, err := h.authSvc.ValidateToken(token)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
		return nil
	}

	return parsedClaims
}

// resolveAuthorUsername fetches the username for a user ID, required for proper registry authorship.
// Returns the username string or an error if the user cannot be found.
func (h *Handler) resolveAuthorUsername(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user by ID: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("user not found: %s", userID)
	}
	if user.Username != nil && *user.Username != "" {
		return *user.Username, nil
	}
	// Fallback to email local part if username is not set
	atIdx := strings.Index(user.Email, "@")
	if atIdx > 0 {
		return user.Email[:atIdx], nil
	}
	return user.Email, nil
}

func (h *Handler) getGitHubClient(ctx context.Context, userID uuid.UUID) (*githubsvc.Client, error) {
	conn, err := h.githubRepo.GetConnectionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	if conn == nil {
		return nil, fmt.Errorf("no GitHub connection found")
	}

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		return nil, fmt.Errorf("create vault: %w", err)
	}

	token, err := vault.Decrypt(conn.EncryptedToken, conn.TokenIV, conn.TokenTag)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"encrypted_len": len(conn.EncryptedToken),
			"iv_len":        len(conn.TokenIV),
			"tag_len":       len(conn.TokenTag),
		}).Error("getGitHubClient: token decrypt failed")
		return nil, fmt.Errorf("decrypt token: %w", err)
	}

	h.logger.WithField("token_prefix", token[:min(10, len(token))]).Info("getGitHubClient: token decrypted")
	client := githubsvc.NewClient(token, githubsvc.WithLogger(h.logger))
	return client, nil
}

// getProgressChan returns or creates the progress event channel for an import.
func (h *Handler) getProgressChan(importID uuid.UUID) chan ProgressEvent {
	ch := make(chan ProgressEvent, 64)
	actual, _ := h.progressCh.LoadOrStore(importID, ch)
	return actual.(chan ProgressEvent)
}

// sendProgress sends a progress event and updates the import status in DB.
func (h *Handler) sendProgress(ctx context.Context, importID uuid.UUID, stage string, progress int, message string) {
	event := ProgressEvent{
		Stage:    stage,
		Progress: progress,
		Message:  message,
	}

	if v, ok := h.progressCh.Load(importID); ok {
		ch := v.(chan ProgressEvent)
		select {
		case ch <- event:
		default:
		}
	}

	if err := h.githubRepo.UpdateImportStatus(ctx, importID, stage, progress, ""); err != nil {
		h.logger.WithError(err).WithField("import_id", importID).Warn("Failed to update import progress")
	}
}

// completeProgress closes the progress channel and removes it from the map.
func (h *Handler) completeProgress(importID uuid.UUID) {
	if v, ok := h.progressCh.LoadAndDelete(importID); ok {
		ch := v.(chan ProgressEvent)
		close(ch)
	}
}

// checkImportRateLimit enforces a maximum of maxConcurrentImportsPerTenant concurrent imports per tenant.
// Returns true if the import can proceed, false if the limit is reached.
func (h *Handler) checkImportRateLimit(_ context.Context, tenantID uuid.UUID) bool {
	val, _ := h.importSem.LoadOrStore(tenantID, new(atomic.Int32))
	count := val.(*atomic.Int32)
	for {
		old := count.Load()
		if old >= maxConcurrentImportsPerTenant {
			return false
		}
		if count.CompareAndSwap(old, old+1) {
			return true
		}
	}
}

// releaseImportSlot decrements the concurrent import count for a tenant.
func (h *Handler) releaseImportSlot(tenantID uuid.UUID) {
	if val, ok := h.importSem.Load(tenantID); ok {
		count := val.(*atomic.Int32)
		for {
			old := count.Load()
			if old <= 0 {
				return
			}
			if count.CompareAndSwap(old, old-1) {
				return
			}
		}
	}
}
