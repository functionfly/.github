package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/artifacts"
	"github.com/sirupsen/logrus"
)

// PresignRequest is the JSON the dashboard sends to ask for a direct-upload
// URL. The dashboard already knows the content hash from prior validation.
type PresignRequest struct {
	Kind        string `json:"kind"`         // "wasm" | "source" | "readme"
	ContentHash string `json:"content_hash"` // sha256 hex of the bytes the client is about to PUT
	ContentType string `json:"content_type"` // optional, default by kind
	Ext         string `json:"ext"`          // optional, default by kind
}

// PresignResponse contains the URL the dashboard should PUT to plus the
// storage keys it must echo back in /registry/publish.
type PresignResponse struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Method      string `json:"method"`
	ContentType string `json:"content_type"`
	MaxBytes    int64  `json:"max_bytes"`
	ExpiresAt   string `json:"expires_at"`
}

// HandlePublishPresign returns a presigned PUT URL for direct browser → R2
// upload of a single artifact, plus the storage key the dashboard must echo
// back in the publish request.
func (h *Handler) HandlePublishPresign(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	if h.artifactStore == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("artifact store not configured"))
		return
	}

	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid JSON"))
		return
	}
	if len(req.ContentHash) != 64 {
		apierror.WriteError(w, apierror.NewBadRequest("content_hash must be 64 hex chars"))
		return
	}
	kind := artifacts.Kind(req.Kind)
	switch kind {
	case artifacts.KindWASM, artifacts.KindSource, artifacts.KindReadme, artifacts.KindCode:
	default:
		apierror.WriteError(w, apierror.NewBadRequest("invalid kind"))
		return
	}

	ct := req.ContentType
	if ct == "" {
		ct = defaultContentTypeForKind(kind)
	}
	key := artifacts.KeyFor(kind, req.ContentHash, req.Ext)

	url, err := h.artifactStore.PresignPut(r.Context(), key, ct, artifacts.DefaultMaxBytes)
	if err != nil {
		apierror.LogAndInternal(w, r, err, "presign put")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": user.UserID,
		"kind":    kind,
		"key":     key,
	}).Debug("presigned artifact upload")

	resp := PresignResponse{
		Key:         key,
		URL:         url,
		Method:      http.MethodPut,
		ContentType: ct,
		MaxBytes:    artifacts.DefaultMaxBytes,
		ExpiresAt:   time.Now().Add(artifacts.PresignTTL).UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
	_ = fmt.Sprintf // keep fmt referenced
}

func defaultContentTypeForKind(k artifacts.Kind) string {
	switch k {
	case artifacts.KindWASM:
		return "application/wasm"
	case artifacts.KindSource:
		return "text/plain; charset=utf-8"
	case artifacts.KindReadme:
		return "text/markdown; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}