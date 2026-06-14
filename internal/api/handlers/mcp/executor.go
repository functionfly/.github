package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	registrystorage "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// RegistryExecutor is the production executor used by the MCP layer. It
// delegates to the existing SecureExecution middleware + HandleExecute flow
// so MCP tool calls share 100% of the platform's rate-limit, verification,
// tenant-isolation, and observability stack.
type RegistryExecutor struct {
	Handler          *registry.Handler
	ExecutionSecurity *middleware.ExecutionCoordinatorMiddleware
	Repo             *registrystorage.RegistryRepository
}

// NewRegistryExecutor wires a RegistryExecutor.
func NewRegistryExecutor(h *registry.Handler, es *middleware.ExecutionCoordinatorMiddleware, repo *registrystorage.RegistryRepository) *RegistryExecutor {
	return &RegistryExecutor{Handler: h, ExecutionSecurity: es, Repo: repo}
}

// ExecuteFunction runs the named function and returns the raw HTTP response.
// It mirrors the inner-REST contract used by SecureExecution: the inner
// request body is `{"input": <json>}` and the inner response is the standard
// {"ok": bool, "data": ...} envelope (or an error envelope).
func (e *RegistryExecutor) ExecuteFunction(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	author, name, version string,
	input json.RawMessage,
) (int, []byte) {
	if e == nil || e.Repo == nil || e.Handler == nil {
		return http.StatusInternalServerError, []byte(`{"ok":false,"error":{"code":"INTERNAL","message":"executor not configured"}}`)
	}
	if e.ExecutionSecurity == nil {
		return http.StatusInternalServerError, []byte(`{"ok":false,"error":{"code":"INTERNAL","message":"execution security not configured"}}`)
	}

	// Resolve function via the public repo.
	fn, err := e.Repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil || fn == nil {
		return http.StatusNotFound, []byte(`{"ok":false,"error":{"code":"NOT_FOUND","message":"function not found"}}`)
	}

	// Build a request that mux.Vars can read. We synthesise a path that
	// matches the routes_registry.go pattern: /v1/fx/{author}/{name}.
	path := fmt.Sprintf("/v1/fx/%s/%s", author, name)
	if version != "" {
		path = fmt.Sprintf("%s@%s", path, version)
	}
	r2 := r.Clone(ctx)
	r2.Method = http.MethodPost
	if r2.URL == nil {
		r2.URL = &url.URL{Path: path}
	} else {
		r2.URL.Path = path
	}
	r2 = mux.SetURLVars(r2, map[string]string{
		"author":  author,
		"name":    name,
		"version": version,
	})
	// Inject the body the executor expects: {"input": <json>}.
	if len(input) == 0 {
		input = json.RawMessage("{}")
	}
	innerBody := struct {
		Input json.RawMessage `json:"input"`
	}{Input: input}
	rawBody, _ := json.Marshal(innerBody)
	r2.Body = io.NopCloser(bytes.NewReader(rawBody))
	r2.ContentLength = int64(len(rawBody))
	r2.Header.Set("Content-Type", "application/json")

	final := e.ExecutionSecurity.SecureExecution(fn.ID, version)(e.Handler.HandleExecute)
	rec := newResponseRecorder()
	final.ServeHTTP(rec, r2)
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	return rec.status, rec.body
}
