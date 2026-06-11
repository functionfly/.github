package registry

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/apierror"
)

// serveSDKLocally serves SDK files from local storage
func (h *Handler) serveSDKLocally(w http.ResponseWriter, r *http.Request, sdkType, version, filename string) {
	// Set content type based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".py":
		w.Header().Set("Content-Type", "text/x-python")
	case ".go":
		w.Header().Set("Content-Type", "text/x-go")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	// Generate SDK content based on type
	var content string
	switch sdkType {
	case "javascript":
		// Generate a generic SDK file
		content = h.generateJavaScriptSDKFile(version)
	case "python":
		content = h.generatePythonSDKFile(version)
	case "go":
		content = h.generateGoSDKFile(version)
	default:
		apierror.WriteError(w, apierror.NewBadRequest("Unsupported SDK type"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

// serveDocsLocally serves documentation from local storage
func (h *Handler) serveDocsLocally(w http.ResponseWriter, r *http.Request, docType, version, path string) {
	// Set content type based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		w.Header().Set("Content-Type", "text/markdown")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	// Generate documentation content
	content := h.generateDocumentation(docType, version, path)
	if content == "" {
		apierror.WriteError(w, apierror.NewNotFound("Documentation not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

// serveStaticLocally serves static assets from local storage
func (h *Handler) serveStaticLocally(w http.ResponseWriter, r *http.Request, category, path string) {
	// Set content type based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Generate or serve static content
	content := h.generateStaticAsset(category, path)
	if content == "" {
		apierror.WriteError(w, apierror.NewNotFound("Asset not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}
