// Package docs FunctionFly
//
// OpenAPI 3.0 specification for the FunctionFly platform API
//
//	Schemes: https
//	Host: api.functionfly.com
//	BasePath: /
//	Version: 1.0.0
//	License: Proprietary - contact@functionfly.com
//
//	Security:
//	  - BearerAuth: JWT token obtained from /auth/login or /auth/signup
//	  - ApiKeyAuth: API key for partner integrations
//
//	OpenAI Compatibility:
//	  - Tools/Function Calling: /registry/functions/{author}/{name}/ai-schema exposes
//	    JSON Schema compatible with OpenAI function calling format
//	  - Streaming: /ai/composer/generate/stream for streaming AI responses
//
// swagger:meta
package docs

import (
	"embed"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// SpecFS embeds the OpenAPI spec files
//
//go:embed swagger.yaml swagger.json
var SpecFS embed.FS

// SwaggerUI embeds Swagger UI static assets
//
//go:embed ui/*
var SwaggerUI embed.FS

// ServerInfo holds OpenAPI server metadata
type ServerInfo struct {
	Version     string `json:"version"`
	Description string `json:"description"`
}

// APIInfo returns current API metadata for OpenAPI spec
func APIInfo() ServerInfo {
	return ServerInfo{
		Version:     "1.0.0",
		Description: "FunctionFly Platform API - Function registry, execution, AI composition, and partner trust APIs",
	}
}

// ServeYAMLSpec serves the raw OpenAPI YAML spec at /swagger/doc.yaml
func ServeYAMLSpec(w http.ResponseWriter, r *http.Request) {
	data, err := SpecFS.ReadFile("swagger.yaml")
	if err != nil {
		logrus.WithError(err).Warn("OpenAPI spec not found")
		apierror.WriteError(w, apierror.NewNotFound("OpenAPI spec not found"))
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeJSONSpec serves the OpenAPI spec as JSON at /swagger/doc.json
// For OpenAI compatibility and tools that prefer JSON
func ServeJSONSpec(w http.ResponseWriter, r *http.Request) {
	data, err := SpecFS.ReadFile("swagger.json")
	if err != nil {
		// Fallback: serve YAML with JSON content-type for consumers expecting JSON
		data, err = SpecFS.ReadFile("swagger.yaml")
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("OpenAPI spec not found"))
			return
		}
		// Some clients require JSON; for full conversion, use a YAML-to-JSON tool
		// or add swag annotations to handlers and run `swag init`
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeSwaggerUI serves the Swagger UI HTML
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "index.html") && !strings.HasSuffix(path, "/") {
		http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
		return
	}

	specURL := getSpecURL(r)
	html := strings.ReplaceAll(swaggerUIHTML, "{{.SpecURL}}", specURL)
	html = strings.ReplaceAll(html, "{{.Title}}", "FunctionFly API Documentation")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// ServeSwaggerStatic serves individual Swagger UI asset files
func ServeSwaggerStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/swagger/static/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	data, err := SwaggerUI.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := mimeType(path)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func mimeType(path string) string {
	switch {
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(path, ".woff"):
		return "font/woff"
	default:
		return "application/octet-stream"
	}
}

func getSpecURL(r *http.Request) string {
	scheme := "https"
	if r.Header.Get("X-Forwarded-Proto") == "http" || strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}
	return scheme + "://" + r.Host + "/swagger/doc.yaml"
}

// GetSwaggerUI returns the embedded Swagger UI filesystem for routing
func GetSwaggerUI() embed.FS {
	return SwaggerUI
}

// swaggerUIHTML is the Swagger UI HTML entry point
// It uses local embedded assets via /swagger/static/ path
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="/swagger/static/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow: hidden; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; padding: 0; background: #fafafa; }
    #swagger-ui { height: 100vh; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="/swagger/static/swagger-ui-bundle.js"></script>
  <script src="/swagger/static/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "{{.SpecURL}}",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrlPlugin
        ],
        layout: "StandaloneLayout",
        validatorUrl: "",
      });
    };
  </script>
</body>
</html>`

// Timestamp is a date-time string wrapper for schema definitions
type Timestamp string

// Time parses the timestamp into a time.Time value
func (t Timestamp) Time() time.Time {
	result, _ := time.Parse(time.RFC3339, string(t))
	return result
}
