package apputil

import (
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HTTPError carries a status and message for writing http.Error from handlers.
type HTTPError struct {
	Status  int
	Message string
}

// appSegmentFromRequest returns the app UUID or slug from mux vars, or parses /v1/apps/<segment>/... from the path.
func appSegmentFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if vars := mux.Vars(r); vars != nil {
		if v := strings.TrimSpace(vars["appId"]); v != "" {
			return v
		}
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	const prefix = "/v1/apps/"
	if strings.HasPrefix(path, prefix) {
		rest := strings.TrimPrefix(path, prefix)
		seg, _, _ := strings.Cut(rest, "/")
		return strings.TrimSpace(seg)
	}
	return ""
}

// ResolveAppForRequest resolves the app from the request path: UUID (existing access rules) or tenant-scoped slug.
func ResolveAppForRequest(repo storage.Repository, user *auth.Claims, r *http.Request) (*storage.App, *HTTPError) {
	if user == nil {
		return nil, &HTTPError{http.StatusUnauthorized, "Unauthorized"}
	}
	idOrSlug := appSegmentFromRequest(r)
	if idOrSlug == "" {
		return nil, &HTTPError{http.StatusBadRequest, "Invalid app ID"}
	}
	if id, err := uuid.Parse(idOrSlug); err == nil {
		app, dberr := repo.GetAppByID(id)
		if dberr != nil {
			return nil, &HTTPError{http.StatusInternalServerError, "Failed to get app"}
		}
		if app == nil {
			return nil, &HTTPError{http.StatusNotFound, "App not found"}
		}
		if app.TenantID != user.TenantID {
			return nil, &HTTPError{http.StatusForbidden, "Access denied"}
		}
		return app, nil
	}
	app, dberr := repo.GetAppBySlugAndTenant(idOrSlug, user.TenantID)
	if dberr != nil {
		return nil, &HTTPError{http.StatusInternalServerError, "Failed to get app"}
	}
	if app == nil {
		return nil, &HTTPError{http.StatusNotFound, "App not found"}
	}
	return app, nil
}
