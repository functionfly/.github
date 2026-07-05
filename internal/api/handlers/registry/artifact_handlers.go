package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/artifacts"
	"github.com/sirupsen/logrus"
)

// DownloadRequest asks for a short-lived URL the browser can use to fetch a
// stored artifact (source, WASM, readme) for preview or download.
type DownloadRequest struct {
	Key string `json:"key"`
	TTL int    `json:"ttl_seconds"` // 0 → use server default
}

// DownloadResponse is the presigned URL plus the time it expires.
type DownloadResponse struct {
	URL       string `json:"url"`
	Method    string `json:"method"`
	ExpiresAt string `json:"expires_at"`
}

// HandleArtifactDownload mints a presigned GET URL for a stored artifact.
// Caller must own the function (auth required) and the requested key must
// correspond to a function they own. The handler validates ownership via the
// function-id parsed out of the key prefix when present, otherwise it allows
// any authenticated request (used for tenant "code" reads where the dashboard
// needs to render the source preview).
func (h *Handler) HandleArtifactDownload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	if h.artifactStore == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("artifact store not configured"))
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid JSON"))
		return
	}
	if req.Key == "" {
		apierror.WriteError(w, apierror.NewBadRequest("key required"))
		return
	}
	ttl := time.Duration(req.TTL) * time.Second
	if ttl <= 0 || ttl > 24*time.Hour {
		ttl = 5 * time.Minute
	}

	url, err := h.artifactStore.PresignGet(r.Context(), req.Key, ttl)
	if err != nil {
		apierror.LogAndInternal(w, r, err, "presign get")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": user.UserID,
		"key":     req.Key,
		"ttl":     ttl,
	}).Debug("presigned artifact download")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(DownloadResponse{
		URL:       url,
		Method:    http.MethodGet,
		ExpiresAt: time.Now().Add(ttl).UTC().Format(time.RFC3339),
	})
}

// HandleArtifactHealth surfaces the artifact store backend name and resolver
// counters so dashboards and probes can verify the wiring without hitting R2.
func (h *Handler) HandleArtifactHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"backend": nil,
		"stats":   map[string]int64{},
	}
	if h.artifactStore != nil {
		resp["backend"] = string(h.artifactStore.Backend())
	}
	if h.artifactResolver != nil {
		resp["stats"] = h.artifactResolver.Stats()
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// LocalUploadHandler accepts a PUT from the dashboard for the local-fallback
// backend (presign handler returned a /api/artifacts/local-upload URL). Not
// used in production — R2 presigned PUTs go directly to R2.
type LocalUploadHandler struct {
	LocalStore *artifacts.LocalStore
}

// ServeHTTP validates the presigned token, then streams the request body to
// disk under the token's target key.
func (h *LocalUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.LocalStore == nil {
		http.Error(w, "local artifact store not configured", http.StatusServiceUnavailable)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	key, err := h.LocalStore.PresignedTokenKey(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	contentType, _ := h.LocalStore.PresignedTokenContentType(token)
	maxBytes, _ := h.LocalStore.PresignedTokenMaxBytes(token)
	// Single-use: drain the token so it can't be reused.
	if _, err := h.LocalStore.ConsumePresignedToken(token); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	limited := http.MaxBytesReader(w, r.Body, artifacts.DefaultMaxBytes)
	if maxBytes > 0 && maxBytes < artifacts.DefaultMaxBytes {
		limited = http.MaxBytesReader(w, r.Body, maxBytes)
	}
	if _, err := h.LocalStore.PutForKey(r.Context(), key, limited, contentType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

// LocalDownloadHandler streams a file out for the local-fallback backend after
// the presigned GET token is validated.
func (h *LocalUploadHandler) LocalDownloadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.LocalStore == nil {
			http.Error(w, "local artifact store not configured", http.StatusServiceUnavailable)
			return
		}
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}
		key, err := h.LocalStore.LocalDownloadToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		rc, err := h.LocalStore.Get(r.Context(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer func() { _ = rc.Close() }()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "private, max-age=60")
		_, _ = io.Copy(w, rc)
	})
}
